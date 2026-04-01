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
const aiPromptCourseLimit = 8
const aiPromptIntakeLimit = 5
const aiRetryPromptCourseLimit = 5
const aiRetryPromptIntakeLimit = 3
const aiMaxOutputTokens = 8000
const aiRetryMaxOutputTokens = 4000
const aiRequestTimeout = 60 * time.Second
const aiMaxReasonLength = 140
const yandexAIKeyIssueMessage = "YANDEX_AI_API_KEY не найден в env или не подходит для Yandex AI API"
const (
	aiResponseSourceAI        = "ai"
	aiResponseSourceHeuristic = "heuristic"
)

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
	ResponseSource      string `json:"response_source,omitempty"`
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
	// Enhanced quality metrics
	PromptCoursesCount int  `json:"prompt_courses_count,omitempty"`
	PromptIntakesCount int  `json:"prompt_intakes_count,omitempty"`
	FilteredTasksCount int  `json:"filtered_tasks_count,omitempty"`
	UsedCompactPrompt  bool `json:"used_compact_prompt,omitempty"`
	UsedRetry          bool `json:"used_retry,omitempty"`
	AIValidJSON        bool `json:"ai_valid_json,omitempty"`
	DiscardedAIItems   int  `json:"discarded_ai_items,omitempty"`
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
	Model           string                `json:"model"`
	Temperature     float64               `json:"temperature"`
	Instructions    string                `json:"instructions"`
	Input           string                `json:"input"`
	MaxOutputTokens int                   `json:"max_output_tokens"`
	Text            *yandexRequestText    `json:"text,omitempty"`
	Reasoning       *yandexReasoningParam `json:"reasoning,omitempty"`
}

type yandexReasoningParam struct {
	Effort string `json:"effort"`
}


type yandexRequestText struct {
	Format yandexTextResponseFormat `json:"format"`
}

type yandexTextResponseFormat struct {
	Type        string         `json:"type"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Strict      bool           `json:"strict,omitempty"`
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

type aiAttemptOptions struct {
	maxOutputTokens int
	compactInput    bool
	attemptLabel    string
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
	httpClient           *http.Client
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
		httpClient: &http.Client{
			Timeout: aiRequestTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		},
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
				ResponseSource: aiResponseSourceHeuristic,
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
			ResponseSource: aiResponseSourceHeuristic,
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
	coursesForAI, intakesForAI := selectAIPromptPools(activeTasks, courses, intakes)
	s.logInfo(
		"ai recommendations prompt pools selected",
		"user_id", principal.UserID.String(),
		"courses_prompt_count", len(coursesForAI),
		"intakes_prompt_count", len(intakesForAI),
		"courses_pool_count", len(courses),
		"intakes_pool_count", len(intakes),
	)
	result, err := s.callYandexAI(ctx, activeTasks, coursesForAI, intakesForAI, aiAttemptOptions{
		maxOutputTokens: aiMaxOutputTokens,
		attemptLabel:    "primary",
	})
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

	if err != nil && shouldRetryAI(result.debug, err) {
		retryCourses := coursesForAI[:minInt(aiRetryPromptCourseLimit, len(coursesForAI))]
		retryIntakes := intakesForAI[:minInt(aiRetryPromptIntakeLimit, len(intakesForAI))]
		s.logWarn(
			"ai recommendations retrying yandex request with compact prompt",
			"user_id", principal.UserID.String(),
			"initial_error", err.Error(),
			"initial_status", result.debug.AIResponseStatus,
			"initial_incomplete_reason", result.debug.AIIncompleteReason,
			"retry_courses_count", len(retryCourses),
			"retry_intakes_count", len(retryIntakes),
			"retry_max_output_tokens", aiRetryMaxOutputTokens,
		)

		retryResult, retryErr := s.callYandexAI(ctx, activeTasks, retryCourses, retryIntakes, aiAttemptOptions{
			maxOutputTokens: aiRetryMaxOutputTokens,
			compactInput:    true,
			attemptLabel:    "compact_retry",
		})
		retryResult.debug.CoursesSource = "/api/v1/courses?limit=50&offset=0"
		retryResult.debug.IntakesSource = "/api/v1/intakes?status=open"
		if coursesErr != nil {
			retryResult.debug.CoursesError = coursesErr.Error()
		}
		if intakesErr != nil {
			retryResult.debug.IntakesError = intakesErr.Error()
		}

		if retryErr == nil {
			response = RecommendResponse{
				Tasks:                 len(activeTasks),
				CoursesInPool:         len(courses),
				IntakesInPool:         len(intakes),
				Recommendations:       retryResult.courseRecommendations,
				IntakeRecommendations: retryResult.intakeRecommendations,
				Debug:                 &retryResult.debug,
			}
			s.logInfo(
				"ai recommendations completed after compact retry",
				"user_id", principal.UserID.String(),
				"tasks_count", len(activeTasks),
				"courses_count", len(courses),
				"intakes_count", len(intakes),
				"course_recommendations", len(response.Recommendations),
				"intake_recommendations", len(response.IntakeRecommendations),
				"request_duration_ms", retryResult.debug.AIRequestDurationMs,
				"response_status", retryResult.debug.AIResponseStatus,
				"incomplete_reason", retryResult.debug.AIIncompleteReason,
			)
			s.tryRecordAudit(ctx, principal.UserID, "ai.recommendations.completed_retry", response, options, coursesErr, intakesErr, nil)
			return response, nil
		}

		s.logWarn(
			"ai recommendations compact retry failed",
			"user_id", principal.UserID.String(),
			"error", retryErr.Error(),
			"response_status", retryResult.debug.AIResponseStatus,
			"incomplete_reason", retryResult.debug.AIIncompleteReason,
			"request_duration_ms", retryResult.debug.AIRequestDurationMs,
		)
		result = retryResult
		err = retryErr
		response = RecommendResponse{
			Tasks:                 len(activeTasks),
			CoursesInPool:         len(courses),
			IntakesInPool:         len(intakes),
			Recommendations:       result.courseRecommendations,
			IntakeRecommendations: result.intakeRecommendations,
			Debug:                 &result.debug,
		}
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

func selectAIPromptPools(
	tasks []yougilemodule.TaskItem,
	courses []catalogmodule.Course,
	intakes []courseintakesmodule.Intake,
) ([]catalogmodule.Course, []courseintakesmodule.Intake) {
	processed := preprocessTasks(tasks)
	taskSummary := buildTasksSummary(processed)
	taskTokens := tokenizeWeighted(taskSummary)

	type rankedCourse struct {
		item  catalogmodule.Course
		score float64
	}
	type rankedIntake struct {
		item  courseintakesmodule.Intake
		score float64
	}

	rankedCourses := make([]rankedCourse, 0, len(courses))
	for _, course := range courses {
		score, _ := scoreWeightedMatch(
			taskTokens,
			course.Title,
			nullableString(course.ShortDescription),
			nullableString(course.Description),
		)
		rankedCourses = append(rankedCourses, rankedCourse{item: course, score: score})
	}
	sort.SliceStable(rankedCourses, func(i, j int) bool {
		if rankedCourses[i].score == rankedCourses[j].score {
			return rankedCourses[i].item.ID.String() < rankedCourses[j].item.ID.String()
		}
		return rankedCourses[i].score > rankedCourses[j].score
	})

	rankedIntakes := make([]rankedIntake, 0, len(intakes))
	now := time.Now()
	for _, intake := range intakes {
		score, _ := scoreWeightedMatch(taskTokens, intake.Title, nullableString(intake.Description))
		// Urgency boost: intakes with deadline within 14 days get a bonus
		if intake.ApplicationDeadline != nil {
			daysUntil := intake.ApplicationDeadline.Sub(now).Hours() / 24
			if daysUntil > 0 && daysUntil <= 14 {
				score += 3.0
			} else if daysUntil > 14 && daysUntil <= 30 {
				score += 1.0
			}
		}
		rankedIntakes = append(rankedIntakes, rankedIntake{item: intake, score: score})
	}
	sort.SliceStable(rankedIntakes, func(i, j int) bool {
		if rankedIntakes[i].score == rankedIntakes[j].score {
			return rankedIntakes[i].item.ID.String() < rankedIntakes[j].item.ID.String()
		}
		return rankedIntakes[i].score > rankedIntakes[j].score
	})

	courseLimit := minInt(aiPromptCourseLimit, len(rankedCourses))
	intakeLimit := minInt(aiPromptIntakeLimit, len(rankedIntakes))

	selectedCourses := make([]catalogmodule.Course, 0, courseLimit)
	for _, ranked := range rankedCourses[:courseLimit] {
		selectedCourses = append(selectedCourses, ranked.item)
	}

	selectedIntakes := make([]courseintakesmodule.Intake, 0, intakeLimit)
	for _, ranked := range rankedIntakes[:intakeLimit] {
		selectedIntakes = append(selectedIntakes, ranked.item)
	}

	return selectedCourses, selectedIntakes
}

type scoredMatch struct {
	ID      string
	Score   int
	Matches []string
}

// noiseTaskPatterns are title substrings that indicate non-work tasks to deprioritize.
var noiseTaskPatterns = []string{
	"тренажерный", "тренажёрный", "спортзал", "фитнес",
	"обед", "завтрак", "ужин", "перерыв",
	"день рождения", "поздравить", "подарок",
	"уборка", "переезд", "ремонт",
}

// isNoiseTask checks if a task title looks like personal/noise activity.
func isNoiseTask(title string) bool {
	lower := strings.ToLower(title)
	for _, pattern := range noiseTaskPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	// Very short titles with no technical signal
	if len([]rune(strings.TrimSpace(title))) < 4 {
		return true
	}
	return false
}

// priorityColumns are column names that indicate active work (higher priority).
var priorityColumns = map[string]bool{
	"в работе":      true,
	"в процессе":    true,
	"in progress":   true,
	"тестирование":  true,
	"testing":       true,
	"review":        true,
	"на проверке":   true,
}

// preprocessTasks filters and prioritizes tasks: active work first, noise removed.
func preprocessTasks(tasks []yougilemodule.TaskItem) []yougilemodule.TaskItem {
	priority := make([]yougilemodule.TaskItem, 0, len(tasks))
	normal := make([]yougilemodule.TaskItem, 0, len(tasks))

	for _, task := range tasks {
		if isNoiseTask(task.Title) {
			continue
		}
		colLower := strings.ToLower(strings.TrimSpace(task.ColumnTitle))
		if priorityColumns[colLower] {
			priority = append(priority, task)
		} else {
			normal = append(normal, task)
		}
	}

	result := make([]yougilemodule.TaskItem, 0, len(priority)+len(normal))
	result = append(result, priority...)
	result = append(result, normal...)
	return result
}

func buildTasksSummary(tasks []yougilemodule.TaskItem) string {
	processed := preprocessTasks(tasks)
	limit := minInt(15, len(processed))
	taskLines := make([]string, 0, limit)
	for index := 0; index < limit; index++ {
		task := processed[index]
		line := fmt.Sprintf("%d. %s", index+1, trimForPrompt(task.Title, 120))
		if task.Description != "" {
			line += " — " + trimForPrompt(task.Description, 100)
		}
		taskLines = append(taskLines, line)
	}
	return strings.Join(taskLines, "\n")
}

func buildTasksSummaryCompact(tasks []yougilemodule.TaskItem) string {
	processed := preprocessTasks(tasks)
	limit := minInt(10, len(processed))
	taskLines := make([]string, 0, limit)
	for index := 0; index < limit; index++ {
		task := processed[index]
		line := fmt.Sprintf("%d. %s", index+1, trimForPrompt(task.Title, 100))
		taskLines = append(taskLines, line)
	}
	return strings.Join(taskLines, "\n")
}

func recommendHeuristically(tasks []yougilemodule.TaskItem, courses []catalogmodule.Course, intakes []courseintakesmodule.Intake, reason string) aiResult {
	processed := preprocessTasks(tasks)
	taskText := buildTasksSummary(processed)
	taskTokens := tokenizeWeighted(taskText)

	courseMatches := make([]scoredMatch, 0, len(courses))
	for _, course := range courses {
		score, matches := scoreWeightedMatch(taskTokens, course.Title, nullableString(course.ShortDescription), nullableString(course.Description))
		if score < 2.0 && len(courses) > 5 {
			continue
		}
		courseMatches = append(courseMatches, scoredMatch{
			ID:      course.ID.String(),
			Score:   int(score),
			Matches: matches,
		})
	}

	intakeMatches := make([]scoredMatch, 0, len(intakes))
	now := time.Now()
	for _, intake := range intakes {
		score, matches := scoreWeightedMatch(taskTokens, intake.Title, nullableString(intake.Description))
		// Urgency boost for fallback too
		if intake.ApplicationDeadline != nil {
			daysUntil := intake.ApplicationDeadline.Sub(now).Hours() / 24
			if daysUntil > 0 && daysUntil <= 14 {
				score += 3.0
			} else if daysUntil > 14 && daysUntil <= 30 {
				score += 1.0
			}
		}
		if score < 2.0 && len(intakes) > 5 {
			continue
		}
		intakeMatches = append(intakeMatches, scoredMatch{
			ID:      intake.ID.String(),
			Score:   int(score),
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
			ResponseSource: aiResponseSourceHeuristic,
			PromptSentToAI: "AI request skipped; heuristic fallback used.",
			AIRawResponse:  "fallback_reason: " + reason,
			AIModelURI:     "fallback://heuristic",
			TasksSummary:   taskText,
			CoursesSummary: buildCoursesSummaryCompact(courses),
			IntakesSummary: buildIntakesSummaryCompact(intakes),
		},
	}
}

// synonymMap maps tokens to their canonical form for matching.
var synonymMap = map[string]string{
	"golang":               "go",
	"бекенд":               "backend",
	"бэкенд":               "backend",
	"бекенда":              "backend",
	"бэкенда":              "backend",
	"фронтенд":             "frontend",
	"фронтэнд":             "frontend",
	"devops":               "devops",
	"cache":                "кеширование",
	"кеш":                  "кеширование",
	"кэш":                  "кеширование",
	"кэширование":          "кеширование",
	"презентация":          "выступление",
	"публичные":            "выступление",
	"domain-driven":        "ddd",
	"sql":                  "sql",
	"postgresql":           "sql",
	"postgres":             "sql",
	"docker":               "контейнеризация",
	"kubernetes":           "контейнеризация",
	"k8s":                  "контейнеризация",
	"облачных":             "cloud",
	"cloud":                "cloud",
	"aws":                  "cloud",
	"azure":                "cloud",
	"наставник":            "mentoring",
	"наставничество":       "mentoring",
	"менторинг":            "mentoring",
	"mentor":               "mentoring",
	"интеграция":           "integration",
	"интеграцию":           "integration",
	"api":                  "api",
	"rest":                 "api",
	"микросервисы":         "microservices",
	"микросервисов":        "microservices",
	"microservices":        "microservices",
	"тестирование":         "testing",
	"тестировании":         "testing",
	"тесты":                "testing",
	"test":                 "testing",
	"tests":                "testing",
	"архитектура":          "architecture",
	"архитектуру":          "architecture",
	"архитектор":           "architecture",
	"architecture":         "architecture",
}

// domainTokenWeights assigns higher scores to domain-specific tokens.
var domainTokenWeights = map[string]float64{
	// Programming languages & frameworks
	"go": 5, "python": 5, "java": 5, "javascript": 5, "typescript": 5,
	"react": 5, "angular": 5, "vue": 5, "django": 5, "spring": 5,
	// Infrastructure & tools
	"redis": 5, "kafka": 4, "rabbitmq": 4, "sql": 4,
	"docker": 4, "контейнеризация": 4, "devops": 5, "cloud": 4,
	"кеширование": 4,
	// Architecture & design
	"ddd": 5, "architecture": 4, "microservices": 4,
	"backend": 4, "frontend": 4, "api": 4, "integration": 3,
	// Skills
	"testing": 3, "mentoring": 2, "security": 4, "безопасность": 4,
	"ml": 4, "machine": 3, "learning": 3, "data": 3,
	// Medium-weight
	"выступление": 2, "команда": 1, "управление": 2, "аналитика": 3,
	"проект": 1, "scrum": 3, "agile": 3, "kanban": 3,
}

// noiseTokens are common words that produce false-positive matches.
var noiseTokens = map[string]struct{}{
	"для": {}, "что": {}, "как": {}, "или": {}, "при": {}, "без": {}, "под": {}, "над": {},
	"это": {}, "the": {}, "and": {}, "with": {}, "from": {}, "into": {}, "task": {},
	"работе": {}, "работа": {}, "работу": {}, "работы": {},
	"основы": {}, "основ": {}, "основам": {},
	"курс": {}, "курса": {}, "курсы": {},
	"сделать": {}, "написать": {}, "создать": {}, "реализовать": {},
	"нужно": {}, "надо": {}, "можно": {}, "будет": {},
	"специалист": {}, "новая": {}, "новый": {}, "новое": {},
	"система": {}, "системы": {}, "систему": {},
	"доска": {}, "колонка": {},
}

type weightedToken struct {
	Token  string
	Weight float64
}

func tokenizeWeighted(values ...string) []weightedToken {
	seen := make(map[string]float64)
	for _, value := range values {
		normalized := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				return unicode.ToLower(r)
			}
			return ' '
		}, value)

		for _, part := range strings.Fields(normalized) {
			if len([]rune(part)) < 2 {
				continue
			}
			if _, skip := noiseTokens[part]; skip {
				continue
			}
			// Apply synonym normalization
			canonical := part
			if syn, ok := synonymMap[part]; ok {
				canonical = syn
			}
			// Determine weight
			weight := 1.0
			if w, ok := domainTokenWeights[canonical]; ok {
				weight = w
			}
			if existing, ok := seen[canonical]; !ok || weight > existing {
				seen[canonical] = weight
			}
		}
	}

	tokens := make([]weightedToken, 0, len(seen))
	for token, weight := range seen {
		tokens = append(tokens, weightedToken{Token: token, Weight: weight})
	}
	return tokens
}

func tokenizeForMatch(values ...string) map[string]struct{} {
	weighted := tokenizeWeighted(values...)
	tokens := make(map[string]struct{}, len(weighted))
	for _, wt := range weighted {
		tokens[wt.Token] = struct{}{}
	}
	return tokens
}

func scoreTextAgainstTaskTokens(taskTokens map[string]struct{}, values ...string) (int, []string) {
	itemTokens := tokenizeWeighted(values...)
	matches := make([]string, 0)
	totalScore := 0.0
	for _, wt := range itemTokens {
		if _, ok := taskTokens[wt.Token]; ok {
			matches = append(matches, wt.Token)
			totalScore += wt.Weight
		}
	}
	sort.Strings(matches)
	score := int(totalScore)
	return score, matches[:minInt(5, len(matches))]
}

// scoreWeightedMatch scores item tokens against weighted task tokens for better ranking.
func scoreWeightedMatch(taskTokens []weightedToken, values ...string) (float64, []string) {
	taskIndex := make(map[string]float64, len(taskTokens))
	for _, wt := range taskTokens {
		if existing, ok := taskIndex[wt.Token]; !ok || wt.Weight > existing {
			taskIndex[wt.Token] = wt.Weight
		}
	}

	itemTokens := tokenizeWeighted(values...)
	matches := make([]string, 0)
	totalScore := 0.0
	for _, wt := range itemTokens {
		if taskWeight, ok := taskIndex[wt.Token]; ok {
			matches = append(matches, wt.Token)
			// Use the higher weight of the two sides for a strong signal
			matchWeight := wt.Weight
			if taskWeight > matchWeight {
				matchWeight = taskWeight
			}
			totalScore += matchWeight
		}
	}
	sort.Strings(matches)
	return totalScore, matches[:minInt(5, len(matches))]
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

func trimForPrompt(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return strings.TrimSpace(value[:max-3]) + "..."
}

// sanitizeReason trims and cleans up AI-generated reasons.
func sanitizeReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "Подобрано на основе анализа рабочих задач."
	}
	// Truncate to max length at word boundary
	runes := []rune(reason)
	if len(runes) > aiMaxReasonLength {
		truncated := string(runes[:aiMaxReasonLength])
		// Try to cut at last space for cleaner truncation
		if lastSpace := strings.LastIndex(truncated, " "); lastSpace > aiMaxReasonLength/2 {
			truncated = truncated[:lastSpace]
		}
		reason = strings.TrimRight(truncated, " .,;:—–-") + "."
	}
	// Filter out generic/empty reasons
	genericReasons := []string{
		"подходит под задачи сотрудника",
		"подходит для сотрудника",
		"рекомендуется к прохождению",
		"полезный курс",
	}
	lowerReason := strings.ToLower(reason)
	for _, generic := range genericReasons {
		if strings.TrimRight(lowerReason, ".!") == generic {
			return "Подобрано на основе анализа рабочих задач."
		}
	}
	return reason
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

func buildCoursesSummaryCompact(courses []catalogmodule.Course) string {
	type courseRef struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	items := make([]courseRef, 0, len(courses))
	for _, course := range courses {
		items = append(items, courseRef{
			ID:    course.ID.String(),
			Title: trimForPrompt(course.Title, 120),
		})
	}

	payload, _ := json.Marshal(items)
	return string(payload)
}

func buildIntakesSummary(intakes []courseintakesmodule.Intake) string {
	type intakeRef struct {
		ID                  string `json:"id"`
		CourseID            string `json:"course_id,omitempty"`
		Title               string `json:"title"`
		ApplicationDeadline string `json:"application_deadline,omitempty"`
		StartDate           string `json:"start_date,omitempty"`
	}

	items := make([]intakeRef, 0, len(intakes))
	for _, intake := range intakes {
		item := intakeRef{
			ID:    intake.ID.String(),
			Title: intake.Title,
		}
		if intake.CourseID != nil {
			item.CourseID = intake.CourseID.String()
		}
		if intake.StartDate != nil {
			item.StartDate = *intake.StartDate
		}
		if intake.ApplicationDeadline != nil {
			item.ApplicationDeadline = intake.ApplicationDeadline.Format(time.RFC3339)
		}
		items = append(items, item)
	}

	payload, _ := json.Marshal(items)
	return string(payload)
}

func buildIntakesSummaryCompact(intakes []courseintakesmodule.Intake) string {
	type intakeRef struct {
		ID                  string `json:"id"`
		CourseID            string `json:"course_id,omitempty"`
		Title               string `json:"title"`
		ApplicationDeadline string `json:"application_deadline,omitempty"`
	}

	items := make([]intakeRef, 0, len(intakes))
	for _, intake := range intakes {
		item := intakeRef{
			ID:    intake.ID.String(),
			Title: trimForPrompt(intake.Title, 120),
		}
		if intake.CourseID != nil {
			item.CourseID = intake.CourseID.String()
		}
		if intake.ApplicationDeadline != nil {
			item.ApplicationDeadline = intake.ApplicationDeadline.Format(time.RFC3339)
		}
		items = append(items, item)
	}

	payload, _ := json.Marshal(items)
	return string(payload)
}

func yandexRecommendationResponseText() yandexRequestText {
	return yandexRequestText{
		Format: yandexTextResponseFormat{
			Type:        "json_schema",
			Name:        "ai_recommendations",
			Description: "Structured recommendations for courses and intakes.",
			Strict:      true,
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

func shouldRetryAI(debug DebugLog, err error) bool {
	if err == nil {
		return false
	}
	if debug.AIResponseStatus == "incomplete" {
		return true
	}
	if debug.AIOutputTextLength == 0 && debug.AIResponseID != "" {
		return true
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "empty response") ||
		strings.Contains(errText, "max_output_tokens") ||
		strings.Contains(errText, "parse ai recommendations json")
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

func (s *Service) callYandexAI(
	ctx context.Context,
	tasks []yougilemodule.TaskItem,
	courses []catalogmodule.Course,
	intakes []courseintakesmodule.Intake,
	options aiAttemptOptions,
) (aiResult, error) {
	if s.aiConfig.APIKey == "" {
		return aiResult{}, httpx.BadRequest("yandex_ai_key_missing_or_invalid", yandexAIKeyIssueMessage)
	}
	if options.maxOutputTokens <= 0 {
		options.maxOutputTokens = aiMaxOutputTokens
	}
	if strings.TrimSpace(options.attemptLabel) == "" {
		options.attemptLabel = "primary"
	}

	requestStartedAt := time.Now()

	// Always use compact summaries — less tokens, better signal/noise ratio
	tasksSummary := buildTasksSummaryCompact(tasks)
	coursesSummary := buildCoursesSummaryCompact(courses)
	intakesSummary := buildIntakesSummaryCompact(intakes)

	instructions := `Ты подбираешь обучение под рабочие задачи сотрудника.

Верни JSON:
{"course_recommendations":[{"course_id":"uuid","reason":"..."}],"intake_recommendations":[{"intake_id":"uuid","reason":"..."}]}

Правила:
1. Используй только переданные задачи и кандидатов.
2. Верни не более 3 курсов и не более 3 наборов.
3. Если релевантных вариантов нет, верни пустой массив.
4. reason: максимум 140 символов.
5. В reason укажи конкретную связь с задачами — технологию, навык или тип работы.
6. Не рекомендуй варианты по слабым или случайным совпадениям слов.
7. Ответ только валидный JSON, без markdown и пояснений.`

	if options.compactInput {
		instructions = `Подбери обучение под задачи сотрудника. JSON-ответ:
{"course_recommendations":[{"course_id":"uuid","reason":"..."}],"intake_recommendations":[{"intake_id":"uuid","reason":"..."}]}
Макс 3 курса, 3 набора. reason до 120 символов. Только конкретные совпадения. Пустой массив если нет релевантных. Только JSON.`
	}

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
		MaxOutputTokens: options.maxOutputTokens,
		Text:            &responseTextFormat,
		Reasoning:       &yandexReasoningParam{Effort: "medium"},
	}

	fullPrompt := fmt.Sprintf("=== ATTEMPT ===\n%s\n\n=== INSTRUCTIONS ===\n%s\n\n=== INPUT ===\n%s", options.attemptLabel, instructions, input)
	debug := DebugLog{
		PromptSentToAI:     fullPrompt,
		AIModelURI:         modelURI,
		TasksSummary:       tasksSummary,
		CoursesSummary:     coursesSummary,
		IntakesSummary:     intakesSummary,
		PromptCoursesCount: len(courses),
		PromptIntakesCount: len(intakes),
		FilteredTasksCount: len(preprocessTasks(tasks)),
		UsedCompactPrompt:  true,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return aiResult{debug: debug}, err
	}

	s.logInfo(
		"yandex ai request prepared",
		"model_uri", modelURI,
		"attempt", options.attemptLabel,
		"compact_input", options.compactInput,
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

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logError(
			"yandex ai request failed",
			"model_uri", modelURI,
			"attempt", options.attemptLabel,
			"compact_input", options.compactInput,
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
		"attempt", options.attemptLabel,
		"compact_input", options.compactInput,
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
			"attempt", options.attemptLabel,
			"compact_input", options.compactInput,
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

	debug.AIValidJSON = true

	courseMap := make(map[string]catalogmodule.Course, len(courses))
	for _, course := range courses {
		courseMap[course.ID.String()] = course
	}

	intakeMap := make(map[string]courseintakesmodule.Intake, len(intakes))
	for _, intake := range intakes {
		intakeMap[intake.ID.String()] = intake
	}

	discarded := 0
	courseRecommendations := make([]AIRecommendation, 0, len(parsed.CourseRecommendations))
	for _, item := range parsed.CourseRecommendations {
		course, ok := courseMap[item.CourseID]
		if !ok {
			discarded++
			continue
		}

		recommendation := AIRecommendation{
			CourseID: course.ID.String(),
			Title:    course.Title,
			Reason:   sanitizeReason(item.Reason),
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
			discarded++
			continue
		}

		recommendation := AIIntakeRecommendation{
			IntakeID: intake.ID.String(),
			Title:    intake.Title,
			Reason:   sanitizeReason(item.Reason),
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

	debug.DiscardedAIItems = discarded

	s.logInfo(
		"yandex ai response parsed",
		"model_uri", modelURI,
		"response_id", aiResp.ID,
		"response_status", aiResp.Status,
		"course_recommendations", len(courseRecommendations),
		"intake_recommendations", len(intakeRecommendations),
		"discarded_items", discarded,
		"duration_ms", debug.AIRequestDurationMs,
	)
	debug.ResponseSource = aiResponseSourceAI

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
		"response_source":              response.Debug.ResponseSource,
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
			"response_source":        response.Debug.ResponseSource,
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
