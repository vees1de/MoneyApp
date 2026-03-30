package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/events"

	"github.com/google/uuid"
)

type Job struct {
	ID          uuid.UUID
	Queue       string
	JobType     string
	Payload     []byte
	Attempt     int
	MaxAttempts int
}

type Handler func(context.Context, Job) error

type Queue struct {
	db       *sql.DB
	logger   *slog.Logger
	handlers map[string]Handler
}

func NewQueue(database *sql.DB, logger *slog.Logger) *Queue {
	return &Queue{
		db:       database,
		logger:   logger,
		handlers: make(map[string]Handler),
	}
}

func (q *Queue) Register(jobType string, handler Handler) {
	q.handlers[jobType] = handler
}

func (q *Queue) Enqueue(ctx context.Context, exec db.DBTX, queueName, jobType string, payload any, idempotencyKey *string, runAt time.Time) error {
	encoded, err := events.MarshalPayload(payload)
	if err != nil {
		return err
	}

	if runAt.IsZero() {
		runAt = time.Now().UTC()
	}

	_, err = exec.ExecContext(ctx, `
		insert into background_jobs (
			id, queue, job_type, payload, idempotency_key, status, run_at,
			attempt, max_attempts, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, 'pending', $6, 0, 5, $6, $6)
		on conflict (queue, idempotency_key) do nothing
	`, uuid.New(), queueName, jobType, encoded, idempotencyKey, runAt)
	return err
}

func (q *Queue) Run(ctx context.Context, workerID string, pollInterval time.Duration) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		if err := q.processOne(ctx, workerID); err != nil {
			q.logger.Error("process job", "worker_id", workerID, "error", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		default:
			time.Sleep(250 * time.Millisecond)
		}
	}
}

func (q *Queue) processOne(ctx context.Context, workerID string) error {
	claimed, err := q.claim(ctx, workerID)
	if err != nil {
		return err
	}
	if claimed == nil {
		return nil
	}

	handled, ok := q.handlers[claimed.JobType]
	if !ok {
		return q.fail(ctx, claimed.ID, fmt.Sprintf("handler for %s is not registered", claimed.JobType))
	}

	if err := handled(ctx, *claimed); err != nil {
		if claimed.Attempt+1 >= claimed.MaxAttempts {
			return q.fail(ctx, claimed.ID, err.Error())
		}
		return q.retry(ctx, claimed.ID, claimed.Attempt+1, err.Error())
	}

	return q.complete(ctx, claimed.ID)
}

func (q *Queue) claim(ctx context.Context, workerID string) (*Job, error) {
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	row := tx.QueryRowContext(ctx, `
		select id, queue, job_type, payload, attempt, max_attempts
		from background_jobs
		where status in ('pending', 'retry')
		  and run_at <= now()
		order by run_at asc, created_at asc
		for update skip locked
		limit 1
	`)

	var job Job
	if err := row.Scan(&job.ID, &job.Queue, &job.JobType, &job.Payload, &job.Attempt, &job.MaxAttempts); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		update background_jobs
		set status = 'processing',
		    locked_at = now(),
		    locked_by = $2,
		    attempt = attempt + 1,
		    updated_at = now()
		where id = $1
	`, job.ID, workerID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &job, nil
}

func (q *Queue) complete(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
		update background_jobs
		set status = 'completed',
		    completed_at = now(),
		    locked_at = null,
		    locked_by = null,
		    updated_at = now()
		where id = $1
	`, id)
	return err
}

func (q *Queue) retry(ctx context.Context, id uuid.UUID, attempt int, message string) error {
	backoff := time.Duration(attempt*attempt) * time.Minute
	_, err := q.db.ExecContext(ctx, `
		update background_jobs
		set status = 'retry',
		    locked_at = null,
		    locked_by = null,
		    run_at = now() + $2::interval,
		    last_error = $3,
		    updated_at = now()
		where id = $1
	`, id, fmt.Sprintf("%d minutes", int(backoff.Minutes())), message)
	return err
}

func (q *Queue) fail(ctx context.Context, id uuid.UUID, message string) error {
	_, err := q.db.ExecContext(ctx, `
		update background_jobs
		set status = 'failed',
		    locked_at = null,
		    locked_by = null,
		    last_error = $2,
		    updated_at = now()
		where id = $1
	`, id, message)
	return err
}
