package testing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Test struct {
	ID               uuid.UUID  `json:"id"`
	CourseID         *uuid.UUID `json:"course_id,omitempty"`
	Title            string     `json:"title"`
	Description      *string    `json:"description,omitempty"`
	AttemptsLimit    *int       `json:"attempts_limit,omitempty"`
	PassingScore     string     `json:"passing_score"`
	ShuffleQuestions bool       `json:"shuffle_questions"`
	ShuffleAnswers   bool       `json:"shuffle_answers"`
	Status           string     `json:"status"`
	CreatedBy        uuid.UUID  `json:"created_by"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	Questions        []Question `json:"questions,omitempty"`
}

type Question struct {
	ID          uuid.UUID      `json:"id"`
	TestID      uuid.UUID      `json:"test_id"`
	Type        string         `json:"type"`
	Text        string         `json:"text"`
	Explanation *string        `json:"explanation,omitempty"`
	SortOrder   int            `json:"sort_order"`
	Points      string         `json:"points"`
	IsRequired  bool           `json:"is_required"`
	Options     []AnswerOption `json:"options,omitempty"`
}

type AnswerOption struct {
	ID         uuid.UUID `json:"id"`
	QuestionID uuid.UUID `json:"question_id"`
	Text       string    `json:"text"`
	IsCorrect  bool      `json:"is_correct"`
	SortOrder  int       `json:"sort_order"`
}

type TestAttempt struct {
	ID           uuid.UUID  `json:"id"`
	TestID       uuid.UUID  `json:"test_id"`
	UserID       uuid.UUID  `json:"user_id"`
	EnrollmentID *uuid.UUID `json:"enrollment_id,omitempty"`
	AttemptNo    int        `json:"attempt_no"`
	Status       string     `json:"status"`
	StartedAt    time.Time  `json:"started_at"`
	SubmittedAt  *time.Time `json:"submitted_at,omitempty"`
	CheckedAt    *time.Time `json:"checked_at,omitempty"`
	Score        *string    `json:"score,omitempty"`
	Passed       *bool      `json:"passed,omitempty"`
}

type TestResult struct {
	ID            uuid.UUID `json:"id"`
	TestID        uuid.UUID `json:"test_id"`
	UserID        uuid.UUID `json:"user_id"`
	BestAttemptID uuid.UUID `json:"best_attempt_id"`
	BestScore     string    `json:"best_score"`
	Passed        bool      `json:"passed"`
	CompletedAt   time.Time `json:"completed_at"`
}

type CreateAnswerOptionRequest struct {
	Text      string `json:"text" validate:"required"`
	IsCorrect bool   `json:"is_correct"`
	SortOrder int    `json:"sort_order"`
}

type CreateQuestionRequest struct {
	Type        string                      `json:"type" validate:"required,oneof=single_choice multiple_choice text number true_false"`
	Text        string                      `json:"text" validate:"required"`
	Explanation *string                     `json:"explanation,omitempty"`
	SortOrder   int                         `json:"sort_order"`
	Points      string                      `json:"points" validate:"required"`
	IsRequired  bool                        `json:"is_required"`
	Options     []CreateAnswerOptionRequest `json:"options,omitempty"`
}

type CreateTestRequest struct {
	CourseID         *uuid.UUID              `json:"course_id,omitempty"`
	Title            string                  `json:"title" validate:"required"`
	Description      *string                 `json:"description,omitempty"`
	AttemptsLimit    *int                    `json:"attempts_limit,omitempty"`
	PassingScore     string                  `json:"passing_score" validate:"required"`
	ShuffleQuestions bool                    `json:"shuffle_questions"`
	ShuffleAnswers   bool                    `json:"shuffle_answers"`
	Status           string                  `json:"status" validate:"omitempty,oneof=draft published archived"`
	Questions        []CreateQuestionRequest `json:"questions,omitempty"`
}

type StartAttemptRequest struct {
	EnrollmentID *uuid.UUID `json:"enrollment_id,omitempty"`
}

type AnswerInput struct {
	QuestionID        uuid.UUID   `json:"question_id" validate:"required"`
	AnswerText        *string     `json:"answer_text,omitempty"`
	SelectedOptionID  *uuid.UUID  `json:"selected_option_id,omitempty"`
	SelectedOptionIDs []uuid.UUID `json:"selected_option_ids,omitempty"`
}

type SubmitAnswersRequest struct {
	Answers []AnswerInput `json:"answers" validate:"required,min=1,dive"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) base(exec ...db.DBTX) db.DBTX {
	if len(exec) > 0 && exec[0] != nil {
		return exec[0]
	}
	return r.db
}

func (r *Repository) CreateTest(ctx context.Context, item Test, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into tests (
			id, course_id, title, description, attempts_limit, passing_score, shuffle_questions,
			shuffle_answers, status, created_by, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, nullif($6, '')::numeric, $7, $8, $9, $10, $11, $12)
	`, item.ID, item.CourseID, item.Title, item.Description, item.AttemptsLimit, item.PassingScore,
		item.ShuffleQuestions, item.ShuffleAnswers, item.Status, item.CreatedBy, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) CreateQuestion(ctx context.Context, item Question, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into questions (id, test_id, type, text, explanation, sort_order, points, is_required)
		values ($1, $2, $3, $4, $5, $6, nullif($7, '')::numeric, $8)
	`, item.ID, item.TestID, item.Type, item.Text, item.Explanation, item.SortOrder, item.Points, item.IsRequired)
	return err
}

func (r *Repository) CreateOption(ctx context.Context, item AnswerOption, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into answer_options (id, question_id, text, is_correct, sort_order)
		values ($1, $2, $3, $4, $5)
	`, item.ID, item.QuestionID, item.Text, item.IsCorrect, item.SortOrder)
	return err
}

func (r *Repository) GetTest(ctx context.Context, id uuid.UUID, exec ...db.DBTX) (Test, error) {
	var item Test
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, course_id, title, description, attempts_limit, passing_score::text, shuffle_questions,
		       shuffle_answers, status, created_by, created_at, updated_at
		from tests
		where id = $1
	`, id).Scan(&item.ID, &item.CourseID, &item.Title, &item.Description, &item.AttemptsLimit, &item.PassingScore,
		&item.ShuffleQuestions, &item.ShuffleAnswers, &item.Status, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) ListQuestions(ctx context.Context, testID uuid.UUID, exec ...db.DBTX) ([]Question, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, test_id, type, text, explanation, sort_order, points::text, is_required
		from questions
		where test_id = $1
		order by sort_order asc, id asc
	`, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Question
	for rows.Next() {
		var item Question
		if err := rows.Scan(&item.ID, &item.TestID, &item.Type, &item.Text, &item.Explanation, &item.SortOrder, &item.Points, &item.IsRequired); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListOptions(ctx context.Context, questionID uuid.UUID, exec ...db.DBTX) ([]AnswerOption, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, question_id, text, is_correct, sort_order
		from answer_options
		where question_id = $1
		order by sort_order asc, id asc
	`, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AnswerOption
	for rows.Next() {
		var item AnswerOption
		if err := rows.Scan(&item.ID, &item.QuestionID, &item.Text, &item.IsCorrect, &item.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) NextAttemptNo(ctx context.Context, testID, userID uuid.UUID, exec ...db.DBTX) (int, error) {
	var value int
	err := r.base(exec...).QueryRowContext(ctx, `
		select coalesce(max(attempt_no), 0) + 1
		from test_attempts
		where test_id = $1 and user_id = $2
	`, testID, userID).Scan(&value)
	return value, err
}

func (r *Repository) CreateAttempt(ctx context.Context, item TestAttempt, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into test_attempts (
			id, test_id, user_id, enrollment_id, attempt_no, status, started_at, submitted_at, checked_at, score, passed
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, nullif($10, '')::numeric, $11)
	`, item.ID, item.TestID, item.UserID, item.EnrollmentID, item.AttemptNo, item.Status, item.StartedAt, item.SubmittedAt, item.CheckedAt, item.Score, item.Passed)
	return err
}

func (r *Repository) GetAttempt(ctx context.Context, attemptID uuid.UUID, exec ...db.DBTX) (TestAttempt, error) {
	var item TestAttempt
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, test_id, user_id, enrollment_id, attempt_no, status, started_at, submitted_at, checked_at, score::text, passed
		from test_attempts
		where id = $1
	`, attemptID).Scan(&item.ID, &item.TestID, &item.UserID, &item.EnrollmentID, &item.AttemptNo, &item.Status, &item.StartedAt, &item.SubmittedAt, &item.CheckedAt, &item.Score, &item.Passed)
	return item, err
}

func (r *Repository) SaveAnswer(ctx context.Context, attemptID uuid.UUID, answer AnswerInput, exec ...db.DBTX) error {
	var selectedOptionIDs any
	if len(answer.SelectedOptionIDs) > 0 {
		encoded, err := json.Marshal(answer.SelectedOptionIDs)
		if err != nil {
			return err
		}
		selectedOptionIDs = encoded
	}

	_, err := r.base(exec...).ExecContext(ctx, `
		insert into test_answers (
			id, attempt_id, question_id, answer_text, selected_option_id, selected_option_ids, is_correct, awarded_points
		)
		values ($1, $2, $3, $4, $5, $6, null, null)
		on conflict (attempt_id, question_id) do update
		set answer_text = excluded.answer_text,
		    selected_option_id = excluded.selected_option_id,
		    selected_option_ids = excluded.selected_option_ids
	`, uuid.New(), attemptID, answer.QuestionID, answer.AnswerText, answer.SelectedOptionID, selectedOptionIDs)
	return err
}

func (r *Repository) ListAnswers(ctx context.Context, attemptID uuid.UUID, exec ...db.DBTX) ([]AnswerInput, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select question_id, answer_text, selected_option_id, selected_option_ids
		from test_answers
		where attempt_id = $1
	`, attemptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AnswerInput
	for rows.Next() {
		var item AnswerInput
		var raw []byte
		if err := rows.Scan(&item.QuestionID, &item.AnswerText, &item.SelectedOptionID, &raw); err != nil {
			return nil, err
		}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &item.SelectedOptionIDs); err != nil {
				return nil, err
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateAnswerGrade(ctx context.Context, attemptID, questionID uuid.UUID, isCorrect bool, awardedPoints string, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update test_answers
		set is_correct = $3,
		    awarded_points = nullif($4, '')::numeric
		where attempt_id = $1 and question_id = $2
	`, attemptID, questionID, isCorrect, awardedPoints)
	return err
}

func (r *Repository) FinalizeAttempt(ctx context.Context, attemptID uuid.UUID, status string, score string, passed bool, submittedAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update test_attempts
		set status = $2,
		    submitted_at = $3,
		    checked_at = $3,
		    score = nullif($4, '')::numeric,
		    passed = $5
		where id = $1
	`, attemptID, status, submittedAt, score, passed)
	return err
}

func (r *Repository) UpsertResult(ctx context.Context, item TestResult, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into test_results (id, test_id, user_id, best_attempt_id, best_score, passed, completed_at)
		values ($1, $2, $3, $4, nullif($5, '')::numeric, $6, $7)
		on conflict (test_id, user_id) do update
		set best_attempt_id = excluded.best_attempt_id,
		    best_score = excluded.best_score,
		    passed = excluded.passed,
		    completed_at = excluded.completed_at
		where test_results.best_score::numeric <= excluded.best_score::numeric
	`, item.ID, item.TestID, item.UserID, item.BestAttemptID, item.BestScore, item.Passed, item.CompletedAt)
	return err
}

func (r *Repository) ListResults(ctx context.Context, testID uuid.UUID, exec ...db.DBTX) ([]TestResult, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, test_id, user_id, best_attempt_id, best_score::text, passed, completed_at
		from test_results
		where test_id = $1
		order by completed_at desc
	`, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TestResult
	for rows.Next() {
		var item TestResult
		if err := rows.Scan(&item.ID, &item.TestID, &item.UserID, &item.BestAttemptID, &item.BestScore, &item.Passed, &item.CompletedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type Service struct {
	db    *sql.DB
	repo  *Repository
	clock clock.Clock
}

func NewService(database *sql.DB, repo *Repository, appClock clock.Clock) *Service {
	return &Service{db: database, repo: repo, clock: appClock}
}

func (s *Service) CreateTest(ctx context.Context, principal platformauth.Principal, req CreateTestRequest) (Test, error) {
	now := s.clock.Now()
	status := req.Status
	if status == "" {
		status = "draft"
	}
	item := Test{
		ID:               uuid.New(),
		CourseID:         req.CourseID,
		Title:            req.Title,
		Description:      req.Description,
		AttemptsLimit:    req.AttemptsLimit,
		PassingScore:     req.PassingScore,
		ShuffleQuestions: req.ShuffleQuestions,
		ShuffleAnswers:   req.ShuffleAnswers,
		Status:           status,
		CreatedBy:        principal.UserID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.CreateTest(ctx, item, tx); err != nil {
			return err
		}
		for _, questionReq := range req.Questions {
			question := Question{
				ID:          uuid.New(),
				TestID:      item.ID,
				Type:        questionReq.Type,
				Text:        questionReq.Text,
				Explanation: questionReq.Explanation,
				SortOrder:   questionReq.SortOrder,
				Points:      questionReq.Points,
				IsRequired:  questionReq.IsRequired,
			}
			if err := s.repo.CreateQuestion(ctx, question, tx); err != nil {
				return err
			}
			for _, optionReq := range questionReq.Options {
				if err := s.repo.CreateOption(ctx, AnswerOption{
					ID:         uuid.New(),
					QuestionID: question.ID,
					Text:       optionReq.Text,
					IsCorrect:  optionReq.IsCorrect,
					SortOrder:  optionReq.SortOrder,
				}, tx); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return Test{}, err
	}

	return s.GetTest(ctx, item.ID)
}

func (s *Service) GetTest(ctx context.Context, id uuid.UUID) (Test, error) {
	item, err := s.repo.GetTest(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Test{}, httpx.NotFound("test_not_found", "test not found")
		}
		return Test{}, err
	}
	questions, err := s.repo.ListQuestions(ctx, item.ID)
	if err != nil {
		return Test{}, err
	}
	for i := range questions {
		options, err := s.repo.ListOptions(ctx, questions[i].ID)
		if err != nil {
			return Test{}, err
		}
		questions[i].Options = options
	}
	item.Questions = questions
	return item, nil
}

func (s *Service) StartAttempt(ctx context.Context, principal platformauth.Principal, testID uuid.UUID, req StartAttemptRequest) (TestAttempt, error) {
	test, err := s.GetTest(ctx, testID)
	if err != nil {
		return TestAttempt{}, err
	}
	nextAttempt, err := s.repo.NextAttemptNo(ctx, test.ID, principal.UserID)
	if err != nil {
		return TestAttempt{}, err
	}
	if test.AttemptsLimit != nil && nextAttempt > *test.AttemptsLimit {
		return TestAttempt{}, httpx.Conflict("attempts_limit_reached", "attempts limit reached")
	}

	item := TestAttempt{
		ID:           uuid.New(),
		TestID:       test.ID,
		UserID:       principal.UserID,
		EnrollmentID: req.EnrollmentID,
		AttemptNo:    nextAttempt,
		Status:       "started",
		StartedAt:    s.clock.Now(),
	}
	return item, s.repo.CreateAttempt(ctx, item)
}

func (s *Service) SaveAnswers(ctx context.Context, principal platformauth.Principal, attemptID uuid.UUID, req SubmitAnswersRequest) error {
	attempt, err := s.repo.GetAttempt(ctx, attemptID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return httpx.NotFound("attempt_not_found", "attempt not found")
		}
		return err
	}
	if attempt.UserID != principal.UserID && !principal.HasPermission("enrollments.manage") {
		return httpx.Forbidden("forbidden", "permission denied")
	}
	for _, answer := range req.Answers {
		if err := s.repo.SaveAnswer(ctx, attemptID, answer); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) SubmitAttempt(ctx context.Context, principal platformauth.Principal, attemptID uuid.UUID) (TestAttempt, error) {
	attempt, err := s.repo.GetAttempt(ctx, attemptID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TestAttempt{}, httpx.NotFound("attempt_not_found", "attempt not found")
		}
		return TestAttempt{}, err
	}
	if attempt.UserID != principal.UserID && !principal.HasPermission("enrollments.manage") {
		return TestAttempt{}, httpx.Forbidden("forbidden", "permission denied")
	}

	test, err := s.GetTest(ctx, attempt.TestID)
	if err != nil {
		return TestAttempt{}, err
	}
	answers, err := s.repo.ListAnswers(ctx, attempt.ID)
	if err != nil {
		return TestAttempt{}, err
	}
	answerMap := make(map[uuid.UUID]AnswerInput, len(answers))
	for _, answer := range answers {
		answerMap[answer.QuestionID] = answer
	}

	scoreValue := 0.0
	for _, question := range test.Questions {
		answer, ok := answerMap[question.ID]
		if !ok {
			continue
		}
		correct := evaluate(question, answer)
		awarded := "0"
		if correct {
			awarded = question.Points
			scoreValue += parseFloat(question.Points)
		}
		if err := s.repo.UpdateAnswerGrade(ctx, attempt.ID, question.ID, correct, awarded); err != nil {
			return TestAttempt{}, err
		}
	}

	score := formatScore(scoreValue)
	passed := scoreValue >= parseFloat(test.PassingScore)
	now := s.clock.Now()
	if err := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.FinalizeAttempt(ctx, attempt.ID, "checked", score, passed, now, tx); err != nil {
			return err
		}
		return s.repo.UpsertResult(ctx, TestResult{
			ID:            uuid.New(),
			TestID:        attempt.TestID,
			UserID:        attempt.UserID,
			BestAttemptID: attempt.ID,
			BestScore:     score,
			Passed:        passed,
			CompletedAt:   now,
		}, tx)
	}); err != nil {
		return TestAttempt{}, err
	}

	return s.repo.GetAttempt(ctx, attempt.ID)
}

func (s *Service) ListResults(ctx context.Context, testID uuid.UUID) ([]TestResult, error) {
	return s.repo.ListResults(ctx, testID)
}

func evaluate(question Question, answer AnswerInput) bool {
	switch question.Type {
	case "single_choice", "true_false":
		for _, option := range question.Options {
			if option.IsCorrect && answer.SelectedOptionID != nil && option.ID == *answer.SelectedOptionID {
				return true
			}
		}
	case "multiple_choice":
		correctIDs := make([]uuid.UUID, 0)
		for _, option := range question.Options {
			if option.IsCorrect {
				correctIDs = append(correctIDs, option.ID)
			}
		}
		if len(correctIDs) != len(answer.SelectedOptionIDs) {
			return false
		}
		for _, id := range correctIDs {
			if !slices.Contains(answer.SelectedOptionIDs, id) {
				return false
			}
		}
		return true
	}
	return false
}

func parseFloat(value string) float64 {
	var parsed float64
	_, _ = fmt.Sscanf(value, "%f", &parsed)
	return parsed
}

func formatScore(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service, validator *validator.Validate) *Handler {
	return &Handler{service: service, validator: validator}
}

func testingPrincipal(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func (h *Handler) CreateTest(w http.ResponseWriter, r *http.Request) {
	principal, err := testingPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req CreateTestRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.CreateTest(r.Context(), principal, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) GetTest(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_test_id", "invalid test id"))
		return
	}
	item, err := h.service.GetTest(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) StartAttempt(w http.ResponseWriter, r *http.Request) {
	principal, err := testingPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	testID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_test_id", "invalid test id"))
		return
	}
	var req StartAttemptRequest
	if err := httpx.DecodeJSON(r, &req); err != nil && !errors.Is(err, io.EOF) {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.StartAttempt(r.Context(), principal, testID, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) SaveAnswers(w http.ResponseWriter, r *http.Request) {
	principal, err := testingPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	attemptID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_attempt_id", "invalid attempt id"))
		return
	}
	var req SubmitAnswersRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	if err := h.service.SaveAnswers(r.Context(), principal, attemptID, req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) SubmitAttempt(w http.ResponseWriter, r *http.Request) {
	principal, err := testingPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	attemptID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_attempt_id", "invalid attempt id"))
		return
	}
	item, err := h.service.SubmitAttempt(r.Context(), principal, attemptID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) ListResults(w http.ResponseWriter, r *http.Request) {
	testID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_test_id", "invalid test id"))
		return
	}
	items, err := h.service.ListResults(r.Context(), testID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
