package ai_recommendations

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"moneyapp/backend/internal/config"
	catalogmodule "moneyapp/backend/internal/modules/catalog"
	yougilemodule "moneyapp/backend/internal/modules/yougile"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type AIRecommendation struct {
	CourseID         string `json:"course_id"`
	Title            string `json:"title"`
	ShortDescription string `json:"short_description,omitempty"`
	Reason           string `json:"reason"`
}

type DebugLog struct {
	PromptSentToAI string `json:"prompt_sent_to_ai"`
	AIRawResponse  string `json:"ai_raw_response"`
	AIModelURI     string `json:"ai_model_uri"`
	TasksSummary   string `json:"tasks_summary"`
	CoursesSummary string `json:"courses_summary"`
}

type RecommendResponse struct {
	Tasks           int                `json:"tasks_analyzed"`
	CoursesInPool   int                `json:"courses_in_pool"`
	Recommendations []AIRecommendation `json:"recommendations"`
	Debug           *DebugLog          `json:"debug,omitempty"`
}

// Yandex AI API types
type yandexRequest struct {
	Model           string  `json:"model"`
	Temperature     float64 `json:"temperature"`
	Instructions    string  `json:"instructions"`
	Input           string  `json:"input"`
	MaxOutputTokens int     `json:"max_output_tokens"`
}

type yandexResponse struct {
	Output []struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func (r yandexResponse) getText() string {
	for _, output := range r.Output {
		if len(output.Content) > 0 {
			return output.Content[0].Text
		}
	}
	return ""
}

type aiResult struct {
	recommendations []AIRecommendation
	debug           DebugLog
}

type Service struct {
	db             *sql.DB
	yougileService *yougilemodule.Service
	catalogService *catalogmodule.Service
	aiConfig       config.YandexAIConfig
}

func NewService(db *sql.DB, yougileService *yougilemodule.Service, catalogService *catalogmodule.Service, aiConfig config.YandexAIConfig) *Service {
	return &Service{
		db:             db,
		yougileService: yougileService,
		catalogService: catalogService,
		aiConfig:       aiConfig,
	}
}

func (s *Service) Recommend(ctx context.Context, principal platformauth.Principal) (RecommendResponse, error) {
	connectionID, err := s.getActiveYougileConnection(ctx, principal.UserID)
	if err != nil {
		return RecommendResponse{}, err
	}

	tasks, err := s.yougileService.ListTasks(ctx, connectionID, yougilemodule.ListTasksQuery{
		Limit: 50,
	})
	if err != nil {
		return RecommendResponse{}, fmt.Errorf("fetch yougile tasks: %w", err)
	}

	activeTasks := make([]yougilemodule.TaskItem, 0, len(tasks.Content))
	for _, t := range tasks.Content {
		if !t.Completed && !t.Archived && !t.Deleted {
			activeTasks = append(activeTasks, t)
		}
	}

	if len(activeTasks) == 0 {
		return RecommendResponse{
			Tasks:           0,
			CoursesInPool:   0,
			Recommendations: []AIRecommendation{},
			Debug: &DebugLog{
				TasksSummary:   "Нет активных задач в YouGile",
				CoursesSummary: "Пропущено — нет задач",
			},
		}, nil
	}

	courses, err := s.catalogService.ListCourses(ctx, principal, catalogmodule.CourseListFilters{})
	if err != nil {
		return RecommendResponse{}, fmt.Errorf("fetch courses: %w", err)
	}

	if len(courses) == 0 {
		return RecommendResponse{
			Tasks:         len(activeTasks),
			CoursesInPool: 0,
			Recommendations: []AIRecommendation{},
			Debug: &DebugLog{
				TasksSummary:   fmt.Sprintf("%d активных задач найдено", len(activeTasks)),
				CoursesSummary: "Нет курсов в каталоге",
			},
		}, nil
	}

	result, err := s.callYandexAI(ctx, activeTasks, courses)
	if err != nil {
		return RecommendResponse{}, fmt.Errorf("yandex ai: %w", err)
	}

	return RecommendResponse{
		Tasks:           len(activeTasks),
		CoursesInPool:   len(courses),
		Recommendations: result.recommendations,
		Debug:           &result.debug,
	}, nil
}

func (s *Service) getActiveYougileConnection(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var connectionID uuid.UUID
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM integration_yougile_connections
		WHERE created_by = $1 AND status <> 'revoked'
		ORDER BY updated_at DESC, created_at DESC
		LIMIT 1
	`, userID).Scan(&connectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, httpx.NotFound("no_yougile_connection", "no active YouGile connection found for this user")
		}
		return uuid.Nil, err
	}
	return connectionID, nil
}

func (s *Service) callYandexAI(ctx context.Context, tasks []yougilemodule.TaskItem, courses []catalogmodule.Course) (aiResult, error) {
	if s.aiConfig.APIKey == "" {
		return aiResult{}, httpx.BadRequest("ai_not_configured", "Yandex AI API key is not configured")
	}

	// Build task lines
	var taskLines []string
	for i, t := range tasks {
		line := fmt.Sprintf("%d. %s", i+1, t.Title)
		if t.Description != "" {
			line += " — " + t.Description
		}
		if t.BoardTitle != "" {
			line += " [доска: " + t.BoardTitle + "]"
		}
		if t.ColumnTitle != "" {
			line += " [колонка: " + t.ColumnTitle + "]"
		}
		taskLines = append(taskLines, line)
	}
	tasksSummary := strings.Join(taskLines, "\n")

	// Build courses JSON
	type courseRef struct {
		ID               string `json:"id"`
		Title            string `json:"title"`
		ShortDescription string `json:"short_description,omitempty"`
	}
	courseRefs := make([]courseRef, 0, len(courses))
	for _, c := range courses {
		ref := courseRef{ID: c.ID.String(), Title: c.Title}
		if c.ShortDescription != nil {
			ref.ShortDescription = *c.ShortDescription
		}
		courseRefs = append(courseRefs, ref)
	}
	coursesJSON, _ := json.MarshalIndent(courseRefs, "", "  ")

	instructions := `Ты — AI-ассистент корпоративной системы обучения. Твоя задача — на основе текущих рабочих задач сотрудника рекомендовать ему наиболее подходящие курсы из каталога.

Правила:
1. Анализируй задачи сотрудника и определи, какие навыки и знания ему нужны.
2. Подбери от 1 до 5 наиболее релевантных курсов из предоставленного каталога.
3. Для каждого курса объясни, почему он полезен для выполнения задач.
4. Отвечай ТОЛЬКО валидным JSON-массивом без markdown-обёрток.

Формат ответа (JSON массив):
[{"course_id": "uuid", "title": "название курса", "reason": "почему этот курс полезен"}]`

	input := fmt.Sprintf("Задачи сотрудника:\n%s\n\nДоступные курсы (JSON):\n%s", tasksSummary, string(coursesJSON))

	modelURI := fmt.Sprintf("gpt://%s/%s", s.aiConfig.FolderID, s.aiConfig.Model)
	reqData := yandexRequest{
		Model:           modelURI,
		Temperature:     0.3,
		Instructions:    instructions,
		Input:           input,
		MaxOutputTokens: 2000,
	}

	fullPrompt := fmt.Sprintf("=== INSTRUCTIONS ===\n%s\n\n=== INPUT ===\n%s", instructions, input)

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return aiResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://ai.api.cloud.yandex.net/v1/responses", bytes.NewBuffer(jsonData))
	if err != nil {
		return aiResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+s.aiConfig.APIKey)
	req.Header.Set("OpenAI-Project", s.aiConfig.FolderID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return aiResult{}, fmt.Errorf("yandex ai request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return aiResult{}, fmt.Errorf("read yandex ai response: %w", err)
	}

	rawResponse := string(body)

	debug := DebugLog{
		PromptSentToAI: fullPrompt,
		AIRawResponse:  rawResponse,
		AIModelURI:     modelURI,
		TasksSummary:   tasksSummary,
		CoursesSummary: string(coursesJSON),
	}

	if resp.StatusCode != http.StatusOK {
		return aiResult{debug: debug}, fmt.Errorf("yandex ai returned status %d: %s", resp.StatusCode, rawResponse)
	}

	var aiResp yandexResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return aiResult{debug: debug}, fmt.Errorf("parse yandex ai response: %w", err)
	}

	text := aiResp.getText()
	if text == "" {
		return aiResult{recommendations: []AIRecommendation{}, debug: debug}, nil
	}

	text = strings.TrimSpace(text)
	if start := strings.Index(text, "["); start >= 0 {
		if end := strings.LastIndex(text, "]"); end > start {
			text = text[start : end+1]
		}
	}

	var parsed []struct {
		CourseID string `json:"course_id"`
		Title    string `json:"title"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return aiResult{debug: debug}, fmt.Errorf("parse ai recommendations json: %w (raw: %s)", err, text)
	}

	courseMap := make(map[string]catalogmodule.Course, len(courses))
	for _, c := range courses {
		courseMap[c.ID.String()] = c
	}

	recommendations := make([]AIRecommendation, 0, len(parsed))
	for _, p := range parsed {
		course, exists := courseMap[p.CourseID]
		if !exists {
			// Курс из AI не найден в каталоге — всё равно добавим с данными от AI
			recommendations = append(recommendations, AIRecommendation{
				CourseID: p.CourseID,
				Title:    p.Title,
				Reason:   p.Reason,
			})
			continue
		}
		rec := AIRecommendation{
			CourseID: p.CourseID,
			Title:    course.Title,
			Reason:   p.Reason,
		}
		if course.ShortDescription != nil {
			rec.ShortDescription = *course.ShortDescription
		}
		recommendations = append(recommendations, rec)
	}

	return aiResult{recommendations: recommendations, debug: debug}, nil
}

// Handler

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Recommend(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authentication required"))
		return
	}

	result, err := h.service.Recommend(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}
