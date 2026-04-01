package ai_recommendations

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode"

	"moneyapp/backend/internal/config"
	"moneyapp/backend/internal/core/audit"
	"moneyapp/backend/internal/core/common"
	catalogmodule "moneyapp/backend/internal/modules/catalog"
	courseintakesmodule "moneyapp/backend/internal/modules/course_intakes"
	yougilemodule "moneyapp/backend/internal/modules/yougile"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"
	platformworker "moneyapp/backend/internal/platform/worker"

	"github.com/google/uuid"
)

const aiPoolLimit = 50
const yandexAIKeyIssueMessage = "YANDEX_AI_API_KEY не найден в env или не подходит для Yandex AI API"

type AIRecommendation struct {
	CourseID         string `json:"course_id"`
	Title            string `json:"title"`
	ShortDescription string `json:"short_description,omitempty"`
	Reason           string `json:"reason"`
}

type AIIntakeRecommendation struct {
	IntakeID            string `json:"intake_id"`
	CourseID            string `json:"course_id,omitempty"`
	Title               string `json:"title"`
	Description         string `json:"description,omitempty"`
	Reason              string `json:"reason"`
	StartDate           string `json:"start_date,omitempty"`
	ApplicationDeadline string `json:"application_deadline,omitempty"`
}

type DebugLog struct {
	PromptSentToAI      string `json:"prompt_sent_to_ai"`
	AIRawResponse       string `json:"ai_raw_response"`
	AIModelURI          string `json:"ai_model_uri"`
	AIRequestDurationMs int64  `json:"ai_request_duration_ms,omitempty"`
	AIResponseID        string `json:"ai_response_id,omitempty"`
	AIResponseStatus    string `json:"ai_response_status,omitempty"`
	AIResponseErrorCode string `json:"ai_response_error_code,omitempty"`
	AIResponseErrorMsg  string `json:"ai_response_error_msg,omitempty"`
	AIIncompleteReason  string `json:"ai_incomplete_reason,omitempty"`
	AIInputTokens       int    `json:"ai_input_tokens,omitempty"`
	AIOutputTokens      int    `json:"ai_output_tokens,omitempty"`
	AIReasoningTokens   int    `json:"ai_reasoning_tokens,omitempty"`
	AITotalTokens       int    `json:"ai_total_tokens,omitempty"`
	AIOutputTextLength  int    `json:"ai_output_text_length,omitempty"`
	TasksSummary        string `json:"tasks_summary"`
	CoursesSummary      string `json:"courses_summary"`
	IntakesSummary      string `json:"intakes_summary"`
	CoursesSource       string `json:"courses_source,omitempty"`
	IntakesSource       string `json:"intakes_source,omitempty"`
	CoursesError        string `json:"courses_error,omitempty"`
	IntakesError        string `json:"intakes_error,omitempty"`
}

type RecommendResponse struct {
	Tasks                 int                      `json:"tasks_analyzed"`
	CoursesInPool         int                      `json:"courses_in_pool"`
	IntakesInPool         int                      `json:"intakes_in_pool"`
	Recommendations       []AIRecommendation       `json:"recommendations"`
	IntakeRecommendations []AIIntakeRecommendation `json:"intake_recommendations"`
	Debug                 *DebugLog                `json:"debug,omitempty"`
}

type RecommendOptions struct {
	IP        *string
	UserAgent *string
}

type yandexRequest struct {
	Model           string             `json:"model"`
	Temperature     float64            `json:"temperature"`
	Instructions    string             `json:"instructions"`
	Input           string             `json:"input"`
	MaxOutputTokens int                `json:"max_output_tokens"`
	Text            *yandexRequestText `json:"text,omitempty"`
}

type yandexRequestText struct {
	Format yandexTextResponseFormat `json:"format"`
}

type yandexTextResponseFormat struct {
	Type        string         `json:"type"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema,omitempty"`
}

type yandexResponse struct {
	ID               string                  `json:"id"`
	Status           string                  `json:"status"`
	OutputText       string                  `json:"output_text"`
	Error            *yandexResponseError    `json:"error,omitempty"`
	IncompleteDetail *yandexIncompleteDetail `json:"incomplete_details,omitempty"`
	Usage            *yandexUsage            `json:"usage,omitempty"`
	Output           []yandexOutputItem      `json:"output,omitempty"`
}

type yandexResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type yandexIncompleteDetail struct {
	Reason string `json:"reason"`
}

type yandexUsage struct {
	InputTokens   int                  `json:"input_tokens"`
	OutputTokens  int                  `json:"output_tokens"`
	TotalTokens   int                  `json:"total_tokens"`
	InputDetails  *yandexInputDetails  `json:"input_tokens_details,omitempty"`
	OutputDetails *yandexOutputDetails `json:"output_tokens_details,omitempty"`
}

type yandexInputDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type yandexOutputDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

type yandexOutputItem struct {
	Type    string                `json:"type,omitempty"`
	Content []yandexOutputContent `json:"content,omitempty"`
}

type yandexOutputContent struct {
	Type    string `json:"type,omitempty"`
	Text    string `json:"text,omitempty"`
	Refusal string `json:"refusal,omitempty"`
}

func (r yandexResponse) getText() string {
	if text := strings.TrimSpace(r.OutputText); text != "" {
		return text
	}

	for _, output := range r.Output {
		if len(output.Content) > 0 {
			for _, content := range output.Content {
				if content.Type == "output_text" && strings.TrimSpace(content.Text) != "" {
					return content.Text
				}
			}
			for _, content := range output.Content {
				if strings.TrimSpace(content.Text) != "" && content.Type == "" {
					return content.Text
				}
			}
		}
	}
	return ""
}

func (r yandexResponse) incompleteReason() string {
	if r.IncompleteDetail == nil {
		return ""
	}
	return r.IncompleteDetail.Reason
}

func (r yandexResponse) usageTokens() (int, int, int, int) {
	if r.Usage == nil {
		return 0, 0, 0, 0
	}

	reasoningTokens := 0
	if r.Usage.OutputDetails != nil {
		reasoningTokens = r.Usage.OutputDetails.ReasoningTokens
	}

	return r.Usage.InputTokens, r.Usage.OutputTokens, r.Usage.TotalTokens, reasoningTokens
}

type aiParsedResponse struct {
	CourseRecommendations []struct {
		CourseID string `json:"course_id"`
		Reason   string `json:"reason"`
	} `json:"course_recommendations"`
	IntakeRecommendations []struct {
		IntakeID string `json:"intake_id"`
		Reason   string `json:"reason"`
	} `json:"intake_recommendations"`
}

type aiResult struct {
	courseRecommendations []AIRecommendation
	intakeRecommendations []AIIntakeRecommendation
	debug                 DebugLog
}

type Service struct {
	db                   *sql.DB
	queue                *platformworker.Queue
	yougileService       *yougilemodule.Service
	catalogService       *catalogmodule.Service
	courseIntakesService *courseintakesmodule.Service
	auditService         *audit.Service
	aiConfig             config.YandexAIConfig
	logger               *slog.Logger
}

func NewService(
	db *sql.DB,
	queue *platformworker.Queue,
	yougileService *yougilemodule.Service,
	catalogService *catalogmodule.Service,
	courseIntakesService *courseintakesmodule.Service,
	auditService *audit.Service,
	aiConfig config.YandexAIConfig,
	logger *slog.Logger,
) *Service {
	return &Service{
		db:                   db,
		queue:                queue,
		yougileService:       yougileService,
		catalogService:       catalogService,
		courseIntakesService: courseIntakesService,
		auditService:         auditService,
		aiConfig:             aiConfig,
		logger:               logger,
	}
}

func (s *Service) logInfo(msg string, args ...any) {
	if s != nil && s.logger != nil {
		s.logger.Info(msg, args...)
	}
}

func (s *Service) logWarn(msg string, args ...any) {
	if s != nil && s.logger != nil {
		s.logger.Warn(msg, args...)
	}
}

func (s *Service) logError(msg string, args ...any) {
	if s != nil && s.logger != nil {
		s.logger.Error(msg, args...)
	}
}

func (s *Service) Recommend(ctx context.Context, principal platformauth.Principal, options RecommendOptions) (RecommendResponse, error) {
	connectionID, err := s.getActiveYougileConnection(ctx, principal.UserID)
	if err != nil {
		return RecommendResponse{}, err
	}

	tasks, err := s.yougileService.ListTasks(ctx, connectionID, yougilemodule.ListTasksQuery{Limit: aiPoolLimit})
	if err != nil {
		return RecommendResponse{}, fmt.Errorf("fetch yougile tasks: %w", err)
	}

	activeTasks := make([]yougilemodule.TaskItem, 0, len(tasks.Content))
	for _, task := range tasks.Content {
		if !task.Completed && !task.Archived && !task.Deleted {
			activeTasks = append(activeTasks, task)
		}
	}

	if len(activeTasks) == 0 {
		s.logInfo("ai recommendations skipped: no active tasks", "user_id", principal.UserID.String())
		response := RecommendResponse{
			Tasks:                 0,
			CoursesInPool:         0,
			IntakesInPool:         0,
			Recommendations:       []AIRecommendation{},
			IntakeRecommendations: []AIIntakeRecommendation{},
			Debug: &DebugLog{
				TasksSummary:   "Нет активных задач в YouGile",
				CoursesSummary: "Пропущено — нет задач",
				IntakesSummary: "Пропущено — нет задач",
			},
		}
		s.tryRecordAudit(ctx, principal.UserID, "ai.recommendations.no_tasks", response, options, nil, nil, nil)
		return response, nil
	}

	courses, coursesErr := s.listCoursePool(ctx, principal)
	intakes, intakesErr := s.courseIntakesService.ListIntakes(ctx, "open")

	if coursesErr != nil && intakesErr != nil {
		s.logError(
			"ai recommendations pool fetch failed",
			"user_id", principal.UserID.String(),
			"courses_error", coursesErr.Error(),
			"intakes_error", intakesErr.Error(),
		)
		debug := &DebugLog{
			TasksSummary:   buildTasksSummary(activeTasks),
			CoursesSummary: "Не удалось получить опубликованные курсы",
			IntakesSummary: "Не удалось получить открытые наборы",
			CoursesSource:  "/api/v1/courses?limit=50&offset=0",
			IntakesSource:  "/api/v1/intakes?status=open",
			CoursesError:   coursesErr.Error(),
			IntakesError:   intakesErr.Error(),
		}
		response := RecommendResponse{
			Tasks:                 len(activeTasks),
			CoursesInPool:         0,
			IntakesInPool:         0,
			Recommendations:       []AIRecommendation{},
			IntakeRecommendations: []AIIntakeRecommendation{},
			Debug:                 debug,
		}
		s.tryRecordAudit(ctx, principal.UserID, "ai.recommendations.failed", response, options, coursesErr, intakesErr, fmt.Errorf("fetch pools failed"))
		return RecommendResponse{}, fmt.Errorf("fetch recommendation pools: courses: %w; intakes: %w", coursesErr, intakesErr)
	}

	if len(courses) == 0 && len(intakes) == 0 {
		s.logWarn(
			"ai recommendations skipped: empty course and intake pools",
			"user_id", principal.UserID.String(),
			"tasks_count", len(activeTasks),
		)
		debug := &DebugLog{
			TasksSummary:   buildTasksSummary(activeTasks),
			CoursesSummary: "Нет опубликованных курсов в пуле /api/v1/courses?limit=50&offset=0",
			IntakesSummary: "Нет открытых наборов в пуле /api/v1/intakes?status=open",
			CoursesSource:  "/api/v1/courses?limit=50&offset=0",
			IntakesSource:  "/api/v1/intakes?status=open",
		}
		if coursesErr != nil {
			debug.CoursesError = coursesErr.Error()
		}
		if intakesErr != nil {
			debug.IntakesError = intakesErr.Error()
		}
		response := RecommendResponse{
			Tasks:                 len(activeTasks),
			CoursesInPool:         0,
			IntakesInPool:         0,
			Recommendations:       []AIRecommendation{},
			IntakeRecommendations: []AIIntakeRecommendation{},
			Debug:                 debug,
		}
		s.tryRecordAudit(ctx, principal.UserID, "ai.recommendations.no_pool", response, options, coursesErr, intakesErr, nil)
		return response, nil
	}

	if strings.TrimSpace(s.aiConfig.APIKey) == "" {
		s.logWarn(
			"ai recommendations using heuristic fallback: yandex key missing or invalid",
			"user_id", principal.UserID.String(),
			"tasks_count", len(activeTasks),
			"courses_count", len(courses),
			"intakes_count", len(intakes),
		)
		result := recommendHeuristically(activeTasks, courses, intakes, yandexAIKeyIssueMessage)
		result.debug.CoursesSource = "/api/v1/courses?limit=50&offset=0"
		result.debug.IntakesSource = "/api/v1/intakes?status=open"
		if coursesErr != nil {
			result.debug.CoursesError = coursesErr.Error()
		}
		if intakesErr != nil {
			result.debug.IntakesError = intakesErr.Error()
		}

		response := RecommendResponse{
			Tasks:                 len(activeTasks),
			CoursesInPool:         len(courses),
			IntakesInPool:         len(intakes),
			Recommendations:       result.courseRecommendations,
			IntakeRecommendations: result.intakeRecommendations,
			Debug:                 &result.debug,
		}
		s.tryRecordAudit(ctx, principal.UserID, "ai.recommendations.completed_fallback", response, options, coursesErr, intakesErr, nil)
		return response, nil
	}

	s.logInfo(
		"ai recommendations ai request starting",
		"user_id", principal.UserID.String(),
		"tasks_count", len(activeTasks),
		"courses_count", len(courses),
		"intakes_count", len(intakes),
		"api_key_configured", true,
	)
	result, err := s.callYandexAI(ctx, activeTasks, courses, intakes)
	result.debug.CoursesSource = "/api/v1/courses?limit=50&offset=0"
	result.debug.IntakesSource = "/api/v1/intakes?status=open"
	if coursesErr != nil {
		result.debug.CoursesError = coursesErr.Error()
	}
	if intakesErr != nil {
		result.debug.IntakesError = intakesErr.Error()
	}

	response := RecommendResponse{
		Tasks:                 len(activeTasks),
		CoursesInPool:         len(courses),
		IntakesInPool:         len(intakes),
		Recommendations:       result.courseRecommendations,
		IntakeRecommendations: result.intakeRecommendations,
		Debug:                 &result.debug,
	}

	if err != nil {
		s.logWarn(
			"ai recommendations using heuristic fallback after yandex error",
			"user_id", principal.UserID.String(),
			"error", err.Error(),
			"request_duration_ms", result.debug.AIRequestDurationMs,
			"response_status", result.debug.AIResponseStatus,
			"incomplete_reason", result.debug.AIIncompleteReason,
		)
		fallback := recommendHeuristically(activeTasks, courses, intakes, err.Error())
		fallback.debug.CoursesSource = "/api/v1/courses?limit=50&offset=0"
		fallback.debug.IntakesSource = "/api/v1/intakes?status=open"
		fallback.debug.AIModelURI = result.debug.AIModelURI
		fallback.debug.AIRequestDurationMs = result.debug.AIRequestDurationMs
		fallback.debug.PromptSentToAI = result.debug.PromptSentToAI
		fallback.debug.AIRawResponse = result.debug.AIRawResponse
		fallback.debug.AIResponseID = result.debug.AIResponseID
		fallback.debug.AIResponseStatus = result.debug.AIResponseStatus
		fallback.debug.AIResponseErrorCode = result.debug.AIResponseErrorCode
		fallback.debug.AIResponseErrorMsg = result.debug.AIResponseErrorMsg
		fallback.debug.AIIncompleteReason = result.debug.AIIncompleteReason
		fallback.debug.AIInputTokens = result.debug.AIInputTokens
		fallback.debug.AIOutputTokens = result.debug.AIOutputTokens
		fallback.debug.AIReasoningTokens = result.debug.AIReasoningTokens
		fallback.debug.AITotalTokens = result.debug.AITotalTokens
		fallback.debug.AIOutputTextLength = result.debug.AIOutputTextLength
		if coursesErr != nil {
			fallback.debug.CoursesError = coursesErr.Error()
		}
		if intakesErr != nil {
			fallback.debug.IntakesError = intakesErr.Error()
		}

		response = RecommendResponse{
			Tasks:                 len(activeTasks),
			CoursesInPool:         len(courses),
			IntakesInPool:         len(intakes),
			Recommendations:       fallback.courseRecommendations,
			IntakeRecommendations: fallback.intakeRecommendations,
			Debug:                 &fallback.debug,
		}
		s.tryRecordAudit(ctx, principal.UserID, "ai.recommendations.completed_fallback", response, options, coursesErr, intakesErr, err)
		return response, nil
	}

	s.logInfo(
		"ai recommendations completed",
		"user_id", principal.UserID.String(),
		"tasks_count", len(activeTasks),
		"courses_count", len(courses),
		"intakes_count", len(intakes),
		"course_recommendations", len(response.Recommendations),
		"intake_recommendations", len(response.IntakeRecommendations),
		"request_duration_ms", result.debug.AIRequestDurationMs,
		"response_status", result.debug.AIResponseStatus,
		"incomplete_reason", result.debug.AIIncompleteReason,
	)
	s.tryRecordAudit(ctx, principal.UserID, "ai.recommendations.completed", response, options, coursesErr, intakesErr, nil)
	return response, nil
}

func (s *Service) listCoursePool(ctx context.Context, principal platformauth.Principal) ([]catalogmodule.Course, error) {
	filters := catalogmodule.CourseListFilters{
		Statuses: []string{"published"},
		Sort:     "newest",
		Pagination: common.Pagination{
			Limit:  aiPoolLimit,
			Offset: 0,
		},
	}
	if principal.HasPermission("courses.write") {
		filters.Statuses = nil
	}
	return s.catalogService.ListCourses(ctx, principal, filters)
}

type scoredMatch struct {
	ID      string
	Score   int
	Matches []string
}

func buildTasksSummary(tasks []yougilemodule.TaskItem) string {
	var taskLines []string
	for index, task := range tasks {
		line := fmt.Sprintf("%d. %s", index+1, task.Title)
		if task.Description != "" {
			line += " — " + task.Description
		}
		if task.BoardTitle != "" {
			line += " [доска: " + task.BoardTitle + "]"
		}
		if task.ColumnTitle != "" {
			line += " [колонка: " + task.ColumnTitle + "]"
		}
		taskLines = append(taskLines, line)
	}
	return strings.Join(taskLines, "\n")
}

func recommendHeuristically(tasks []yougilemodule.TaskItem, courses []catalogmodule.Course, intakes []courseintakesmodule.Intake, reason string) aiResult {
	taskText := buildTasksSummary(tasks)
	taskTokens := tokenizeForMatch(taskText)

	courseMatches := make([]scoredMatch, 0, len(courses))
	for _, course := range courses {
		score, matches := scoreTextAgainstTaskTokens(taskTokens, course.Title, nullableString(course.ShortDescription), nullableString(course.Description))
		if score == 0 && len(courses) > 5 {
			continue
		}
		courseMatches = append(courseMatches, scoredMatch{
			ID:      course.ID.String(),
			Score:   score,
			Matches: matches,
		})
	}

	intakeMatches := make([]scoredMatch, 0, len(intakes))
	for _, intake := range intakes {
		score, matches := scoreTextAgainstTaskTokens(taskTokens, intake.Title, nullableString(intake.Description))
		if score == 0 && len(intakes) > 5 {
			continue
		}
		intakeMatches = append(intakeMatches, scoredMatch{
			ID:      intake.ID.String(),
			Score:   score,
			Matches: matches,
		})
	}

	sort.SliceStable(courseMatches, func(i, j int) bool {
		if courseMatches[i].Score == courseMatches[j].Score {
			return courseMatches[i].ID < courseMatches[j].ID
		}
		return courseMatches[i].Score > courseMatches[j].Score
	})
	sort.SliceStable(intakeMatches, func(i, j int) bool {
		if intakeMatches[i].Score == intakeMatches[j].Score {
			return intakeMatches[i].ID < intakeMatches[j].ID
		}
		return intakeMatches[i].Score > intakeMatches[j].Score
	})

	courseIndex := make(map[string]catalogmodule.Course, len(courses))
	for _, course := range courses {
		courseIndex[course.ID.String()] = course
	}
	intakeIndex := make(map[string]courseintakesmodule.Intake, len(intakes))
	for _, intake := range intakes {
		intakeIndex[intake.ID.String()] = intake
	}

	courseRecommendations := make([]AIRecommendation, 0, minInt(5, len(courseMatches)))
	for _, match := range courseMatches[:minInt(5, len(courseMatches))] {
		course := courseIndex[match.ID]
		recommendation := AIRecommendation{
			CourseID: course.ID.String(),
			Title:    course.Title,
			Reason:   buildCourseFallbackReason(match),
		}
		if course.ShortDescription != nil {
			recommendation.ShortDescription = *course.ShortDescription
		}
		courseRecommendations = append(courseRecommendations, recommendation)
	}

	intakeRecommendations := make([]AIIntakeRecommendation, 0, minInt(5, len(intakeMatches)))
	for _, match := range intakeMatches[:minInt(5, len(intakeMatches))] {
		intake := intakeIndex[match.ID]
		recommendation := AIIntakeRecommendation{
			IntakeID: intake.ID.String(),
			Title:    intake.Title,
			Reason:   buildIntakeFallbackReason(match, intake),
		}
		if intake.CourseID != nil {
			recommendation.CourseID = intake.CourseID.String()
		}
		if intake.Description != nil {
			recommendation.Description = *intake.Description
		}
		if intake.StartDate != nil {
			recommendation.StartDate = *intake.StartDate
		}
		if intake.ApplicationDeadline != nil {
			recommendation.ApplicationDeadline = intake.ApplicationDeadline.Format(time.RFC3339)
		}
		intakeRecommendations = append(intakeRecommendations, recommendation)
	}

	return aiResult{
		courseRecommendations: courseRecommendations,
		intakeRecommendations: intakeRecommendations,
		debug: DebugLog{
			PromptSentToAI: "AI request skipped; heuristic fallback used.",
			AIRawResponse:  "fallback_reason: " + reason,
			AIModelURI:     "fallback://heuristic",
			TasksSummary:   taskText,
			CoursesSummary: buildCoursesSummary(courses),
			IntakesSummary: buildIntakesSummary(intakes),
		},
	}
}

func tokenizeForMatch(values ...string) map[string]struct{} {
	tokens := make(map[string]struct{})
	for _, value := range values {
		normalized := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				return unicode.ToLower(r)
			}
			return ' '
		}, value)

		for _, part := range strings.Fields(normalized) {
			if len([]rune(part)) < 3 {
				continue
			}
			if _, skip := ignoredMatchTokens[part]; skip {
				continue
			}
			tokens[part] = struct{}{}
		}
	}
	return tokens
}

var ignoredMatchTokens = map[string]struct{}{
	"для": {}, "что": {}, "как": {}, "или": {}, "при": {}, "без": {}, "под": {}, "над": {},
	"это": {}, "the": {}, "and": {}, "with": {}, "from": {}, "into": {}, "task": {},
}

func scoreTextAgainstTaskTokens(taskTokens map[string]struct{}, values ...string) (int, []string) {
	itemTokens := tokenizeForMatch(values...)
	matches := make([]string, 0, len(itemTokens))
	for token := range itemTokens {
		if _, ok := taskTokens[token]; ok {
			matches = append(matches, token)
		}
	}
	sort.Strings(matches)
	score := len(matches)
	if score > 0 {
		score += minInt(2, len(matches))
	}
	return score, matches[:minInt(5, len(matches))]
}

func buildCourseFallbackReason(match scoredMatch) string {
	if len(match.Matches) == 0 {
		return "Подобрано эвристически из опубликованного пула курсов по общему профилю текущих задач."
	}
	return fmt.Sprintf("Подобрано эвристически по совпадениям с задачами: %s.", strings.Join(match.Matches, ", "))
}

func buildIntakeFallbackReason(match scoredMatch, intake courseintakesmodule.Intake) string {
	base := "Подобрано эвристически из открытого пула наборов."
	if len(match.Matches) > 0 {
		base = fmt.Sprintf("Подобрано эвристически по совпадениям с задачами: %s.", strings.Join(match.Matches, ", "))
	}
	if intake.ApplicationDeadline != nil {
		base += " Набор сейчас открыт для подачи заявки."
	}
	return base
}

func nullableString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func buildCoursesSummary(courses []catalogmodule.Course) string {
	type courseRef struct {
		ID               string `json:"id"`
		Title            string `json:"title"`
		ShortDescription string `json:"short_description,omitempty"`
	}

	items := make([]courseRef, 0, len(courses))
	for _, course := range courses {
		item := courseRef{ID: course.ID.String(), Title: course.Title}
		if course.ShortDescription != nil {
			item.ShortDescription = *course.ShortDescription
		}
		items = append(items, item)
	}

	payload, _ := json.MarshalIndent(items, "", "  ")
	return string(payload)
}

func buildIntakesSummary(intakes []courseintakesmodule.Intake) string {
	type intakeRef struct {
		ID                  string `json:"id"`
		CourseID            string `json:"course_id,omitempty"`
		Title               string `json:"title"`
		Description         string `json:"description,omitempty"`
		StartDate           string `json:"start_date,omitempty"`
		EndDate             string `json:"end_date,omitempty"`
		ApplicationDeadline string `json:"application_deadline,omitempty"`
		Status              string `json:"status"`
	}

	items := make([]intakeRef, 0, len(intakes))
	for _, intake := range intakes {
		item := intakeRef{
			ID:     intake.ID.String(),
			Title:  intake.Title,
			Status: intake.Status,
		}
		if intake.CourseID != nil {
			item.CourseID = intake.CourseID.String()
		}
		if intake.Description != nil {
			item.Description = *intake.Description
		}
		if intake.StartDate != nil {
			item.StartDate = *intake.StartDate
		}
		if intake.EndDate != nil {
			item.EndDate = *intake.EndDate
		}
		if intake.ApplicationDeadline != nil {
			item.ApplicationDeadline = intake.ApplicationDeadline.Format(time.RFC3339)
		}
		items = append(items, item)
	}

	payload, _ := json.MarshalIndent(items, "", "  ")
	return string(payload)
}

func yandexRecommendationResponseText() yandexRequestText {
	return yandexRequestText{
		Format: yandexTextResponseFormat{
			Type:        "json_schema",
			Name:        "ai_recommendations",
			Description: "Structured recommendations for courses and intakes.",
			Schema:      yandexRecommendationSchema(),
		},
	}
}

func yandexRecommendationSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"course_recommendations", "intake_recommendations"},
		"properties": map[string]any{
			"course_recommendations": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"course_id", "reason"},
					"properties": map[string]any{
						"course_id": map[string]any{"type": "string"},
						"reason":    map[string]any{"type": "string"},
					},
				},
			},
			"intake_recommendations": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"intake_id", "reason"},
					"properties": map[string]any{
						"intake_id": map[string]any{"type": "string"},
						"reason":    map[string]any{"type": "string"},
					},
				},
			},
		},
	}
}

func truncateForLog(value string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(value) <= max {
		return value
	}
	return value[:max] + "..."
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

func (s *Service) callYandexAI(ctx context.Context, tasks []yougilemodule.TaskItem, courses []catalogmodule.Course, intakes []courseintakesmodule.Intake) (aiResult, error) {
	if s.aiConfig.APIKey == "" {
		return aiResult{}, httpx.BadRequest("yandex_ai_key_missing_or_invalid", yandexAIKeyIssueMessage)
	}

	requestStartedAt := time.Now()
	tasksSummary := buildTasksSummary(tasks)
	coursesSummary := buildCoursesSummary(courses)
	intakesSummary := buildIntakesSummary(intakes)

	instructions := `Ты — AI-ассистент корпоративной системы обучения. На основе текущих рабочих задач сотрудника нужно рекомендовать:
1. Подходящие опубликованные курсы из пула course_recommendations.
2. Подходящие открытые наборы обучения из пула intake_recommendations.

Правила:
1. Анализируй только переданные задачи, курсы и наборы.
2. Возвращай от 0 до 5 курсов и от 0 до 5 наборов.
3. Если подходящих вариантов нет, возвращай пустой массив для соответствующего раздела.
4. Для каждого элемента укажи только id и причину.
5. Отвечай ТОЛЬКО валидным JSON-объектом без markdown-обёрток.

Формат ответа:
{
  "course_recommendations": [{"course_id": "uuid", "reason": "почему курс полезен"}],
  "intake_recommendations": [{"intake_id": "uuid", "reason": "почему набор полезен прямо сейчас"}]
}`

	input := fmt.Sprintf(
		"Задачи сотрудника:\n%s\n\nОпубликованные курсы (JSON):\n%s\n\nОткрытые наборы (JSON):\n%s",
		tasksSummary,
		coursesSummary,
		intakesSummary,
	)

	modelURI := fmt.Sprintf("gpt://%s/%s", s.aiConfig.FolderID, s.aiConfig.Model)
	responseTextFormat := yandexRecommendationResponseText()
	reqData := yandexRequest{
		Model:           modelURI,
		Temperature:     0.2,
		Instructions:    instructions,
		Input:           input,
		MaxOutputTokens: 256,
		Text:            &responseTextFormat,
	}

	fullPrompt := fmt.Sprintf("=== INSTRUCTIONS ===\n%s\n\n=== INPUT ===\n%s", instructions, input)
	debug := DebugLog{
		PromptSentToAI: fullPrompt,
		AIModelURI:     modelURI,
		TasksSummary:   tasksSummary,
		CoursesSummary: coursesSummary,
		IntakesSummary: intakesSummary,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return aiResult{debug: debug}, err
	}

	s.logInfo(
		"yandex ai request prepared",
		"model_uri", modelURI,
		"tasks_count", len(tasks),
		"courses_count", len(courses),
		"intakes_count", len(intakes),
		"request_bytes", len(jsonData),
		"input_bytes", len(input),
		"max_output_tokens", reqData.MaxOutputTokens,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://ai.api.cloud.yandex.net/v1/responses", bytes.NewBuffer(jsonData))
	if err != nil {
		return aiResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+s.aiConfig.APIKey)
	req.Header.Set("OpenAI-Project", s.aiConfig.FolderID)

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.logError(
			"yandex ai request failed",
			"model_uri", modelURI,
			"error", err.Error(),
			"elapsed_ms", time.Since(requestStartedAt).Milliseconds(),
		)
		debug.AIRequestDurationMs = time.Since(requestStartedAt).Milliseconds()
		return aiResult{debug: debug}, fmt.Errorf("yandex ai request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	debug.AIRequestDurationMs = time.Since(requestStartedAt).Milliseconds()
	if err != nil {
		s.logError(
			"read yandex ai response failed",
			"model_uri", modelURI,
			"status_code", resp.StatusCode,
			"elapsed_ms", debug.AIRequestDurationMs,
			"error", err.Error(),
		)
		return aiResult{debug: debug}, fmt.Errorf("read yandex ai response: %w", err)
	}

	rawResponse := string(body)
	debug.AIRawResponse = rawResponse

	var aiResp yandexResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		s.logError(
			"parse yandex ai response failed",
			"model_uri", modelURI,
			"status_code", resp.StatusCode,
			"elapsed_ms", debug.AIRequestDurationMs,
			"response_bytes", len(body),
			"error", err.Error(),
			"response_snippet", truncateForLog(rawResponse, 600),
		)
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return aiResult{debug: debug}, httpx.BadRequest("yandex_ai_key_missing_or_invalid", yandexAIKeyIssueMessage)
		}
		return aiResult{debug: debug}, fmt.Errorf("parse yandex ai response: %w", err)
	}

	debug.AIResponseID = aiResp.ID
	debug.AIResponseStatus = aiResp.Status
	debug.AIIncompleteReason = aiResp.incompleteReason()
	debug.AIInputTokens, debug.AIOutputTokens, debug.AITotalTokens, debug.AIReasoningTokens = aiResp.usageTokens()
	if aiResp.Error != nil {
		debug.AIResponseErrorCode = aiResp.Error.Code
		debug.AIResponseErrorMsg = aiResp.Error.Message
	}

	text := strings.TrimSpace(aiResp.getText())
	debug.AIOutputTextLength = len(text)

	s.logInfo(
		"yandex ai response received",
		"model_uri", modelURI,
		"status_code", resp.StatusCode,
		"response_status", aiResp.Status,
		"response_id", aiResp.ID,
		"elapsed_ms", debug.AIRequestDurationMs,
		"response_bytes", len(body),
		"output_text_length", debug.AIOutputTextLength,
		"incomplete_reason", debug.AIIncompleteReason,
		"input_tokens", debug.AIInputTokens,
		"output_tokens", debug.AIOutputTokens,
		"reasoning_tokens", debug.AIReasoningTokens,
		"total_tokens", debug.AITotalTokens,
		"error_code", debug.AIResponseErrorCode,
		"error_message", debug.AIResponseErrorMsg,
	)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return aiResult{debug: debug}, httpx.BadRequest("yandex_ai_key_missing_or_invalid", yandexAIKeyIssueMessage)
		}
		if aiResp.Error != nil {
			debug.AIResponseErrorCode = aiResp.Error.Code
			debug.AIResponseErrorMsg = aiResp.Error.Message
		}
		return aiResult{debug: debug}, fmt.Errorf("yandex ai returned status %d: %s", resp.StatusCode, truncateForLog(rawResponse, 1200))
	}

	if aiResp.Error != nil {
		s.logWarn(
			"yandex ai response reported error",
			"model_uri", modelURI,
			"response_id", aiResp.ID,
			"response_status", aiResp.Status,
			"error_code", aiResp.Error.Code,
			"error_message", aiResp.Error.Message,
		)
	}

	if text == "" {
		s.logWarn(
			"yandex ai response had empty output text",
			"model_uri", modelURI,
			"response_id", aiResp.ID,
			"response_status", aiResp.Status,
			"incomplete_reason", debug.AIIncompleteReason,
			"input_tokens", debug.AIInputTokens,
			"output_tokens", debug.AIOutputTokens,
			"reasoning_tokens", debug.AIReasoningTokens,
			"total_tokens", debug.AITotalTokens,
			"response_snippet", truncateForLog(rawResponse, 600),
		)
		return aiResult{
			courseRecommendations: []AIRecommendation{},
			intakeRecommendations: []AIIntakeRecommendation{},
			debug:                 debug,
		}, fmt.Errorf("yandex ai returned empty response: status=%s incomplete_reason=%s", aiResp.Status, aiResp.incompleteReason())
	}

	if aiResp.Status != "" && aiResp.Status != "completed" {
		s.logWarn(
			"yandex ai response finished with non-completed status",
			"model_uri", modelURI,
			"response_id", aiResp.ID,
			"response_status", aiResp.Status,
			"incomplete_reason", debug.AIIncompleteReason,
		)
	}

	if start := strings.Index(text, "{"); start >= 0 {
		if end := strings.LastIndex(text, "}"); end > start {
			text = text[start : end+1]
		}
	}

	var parsed aiParsedResponse
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		s.logError(
			"parse ai recommendations json failed",
			"model_uri", modelURI,
			"response_id", aiResp.ID,
			"response_status", aiResp.Status,
			"error", err.Error(),
			"response_snippet", truncateForLog(text, 800),
		)
		return aiResult{debug: debug}, fmt.Errorf("parse ai recommendations json: %w (raw: %s)", err, text)
	}

	courseMap := make(map[string]catalogmodule.Course, len(courses))
	for _, course := range courses {
		courseMap[course.ID.String()] = course
	}

	intakeMap := make(map[string]courseintakesmodule.Intake, len(intakes))
	for _, intake := range intakes {
		intakeMap[intake.ID.String()] = intake
	}

	courseRecommendations := make([]AIRecommendation, 0, len(parsed.CourseRecommendations))
	for _, item := range parsed.CourseRecommendations {
		course, ok := courseMap[item.CourseID]
		if !ok {
			continue
		}

		recommendation := AIRecommendation{
			CourseID: course.ID.String(),
			Title:    course.Title,
			Reason:   item.Reason,
		}
		if course.ShortDescription != nil {
			recommendation.ShortDescription = *course.ShortDescription
		}
		courseRecommendations = append(courseRecommendations, recommendation)
	}

	intakeRecommendations := make([]AIIntakeRecommendation, 0, len(parsed.IntakeRecommendations))
	for _, item := range parsed.IntakeRecommendations {
		intake, ok := intakeMap[item.IntakeID]
		if !ok {
			continue
		}

		recommendation := AIIntakeRecommendation{
			IntakeID: intake.ID.String(),
			Title:    intake.Title,
			Reason:   item.Reason,
		}
		if intake.CourseID != nil {
			recommendation.CourseID = intake.CourseID.String()
		}
		if intake.Description != nil {
			recommendation.Description = *intake.Description
		}
		if intake.StartDate != nil {
			recommendation.StartDate = *intake.StartDate
		}
		if intake.ApplicationDeadline != nil {
			recommendation.ApplicationDeadline = intake.ApplicationDeadline.Format(time.RFC3339)
		}
		intakeRecommendations = append(intakeRecommendations, recommendation)
	}

	s.logInfo(
		"yandex ai response parsed",
		"model_uri", modelURI,
		"response_id", aiResp.ID,
		"response_status", aiResp.Status,
		"course_recommendations", len(courseRecommendations),
		"intake_recommendations", len(intakeRecommendations),
		"duration_ms", debug.AIRequestDurationMs,
	)

	return aiResult{
		courseRecommendations: courseRecommendations,
		intakeRecommendations: intakeRecommendations,
		debug:                 debug,
	}, nil
}

func (s *Service) tryRecordAudit(
	ctx context.Context,
	userID uuid.UUID,
	action string,
	response RecommendResponse,
	options RecommendOptions,
	coursesErr error,
	intakesErr error,
	callErr error,
) {
	if s.auditService == nil || response.Debug == nil {
		return
	}

	meta := map[string]any{
		"tasks_analyzed":               response.Tasks,
		"courses_in_pool":              response.CoursesInPool,
		"intakes_in_pool":              response.IntakesInPool,
		"course_recommendations_count": len(response.Recommendations),
		"intake_recommendations_count": len(response.IntakeRecommendations),
		"courses_source":               response.Debug.CoursesSource,
		"intakes_source":               response.Debug.IntakesSource,
		"ai_response_id":               response.Debug.AIResponseID,
		"ai_response_status":           response.Debug.AIResponseStatus,
		"ai_incomplete_reason":         response.Debug.AIIncompleteReason,
		"ai_input_tokens":              response.Debug.AIInputTokens,
		"ai_output_tokens":             response.Debug.AIOutputTokens,
		"ai_reasoning_tokens":          response.Debug.AIReasoningTokens,
		"ai_total_tokens":              response.Debug.AITotalTokens,
		"ai_output_text_length":        response.Debug.AIOutputTextLength,
	}
	if coursesErr != nil {
		meta["courses_error"] = coursesErr.Error()
	}
	if intakesErr != nil {
		meta["intakes_error"] = intakesErr.Error()
	}
	if callErr != nil {
		meta["error"] = callErr.Error()
	}
	if response.Debug.AIModelURI != "" {
		meta["ai_model_uri"] = response.Debug.AIModelURI
	}

	changeSet := map[string]any{
		"after": map[string]any{
			"tasks_summary":          response.Debug.TasksSummary,
			"courses_summary":        response.Debug.CoursesSummary,
			"intakes_summary":        response.Debug.IntakesSummary,
			"prompt_sent_to_ai":      response.Debug.PromptSentToAI,
			"ai_raw_response":        response.Debug.AIRawResponse,
			"ai_response_id":         response.Debug.AIResponseID,
			"ai_response_status":     response.Debug.AIResponseStatus,
			"ai_incomplete_reason":   response.Debug.AIIncompleteReason,
			"ai_input_tokens":        response.Debug.AIInputTokens,
			"ai_output_tokens":       response.Debug.AIOutputTokens,
			"ai_reasoning_tokens":    response.Debug.AIReasoningTokens,
			"ai_total_tokens":        response.Debug.AITotalTokens,
			"ai_output_text_length":  response.Debug.AIOutputTextLength,
			"recommendations":        response.Recommendations,
			"intake_recommendations": response.IntakeRecommendations,
		},
	}

	_ = s.auditService.RecordChange(ctx, audit.RecordInput{
		UserID:     userID,
		Action:     action,
		EntityType: "ai_recommendations",
		Meta:       meta,
		ChangeSet:  changeSet,
		Source:     audit.SourceSystem,
		IP:         options.IP,
		UserAgent:  options.UserAgent,
	})
}

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

	result, err := h.service.Recommend(r.Context(), principal, RecommendOptions{
		IP:        requestIP(r),
		UserAgent: optionalString(strings.TrimSpace(r.UserAgent())),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func requestIP(r *http.Request) *string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		ip := strings.TrimSpace(strings.Split(forwarded, ",")[0])
		return optionalString(ip)
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return optionalString(realIP)
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return optionalString(host)
	}
	return optionalString(strings.TrimSpace(r.RemoteAddr))
}
