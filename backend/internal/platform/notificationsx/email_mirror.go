package notificationsx

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"moneyapp/backend/internal/platform/db"

	"github.com/google/uuid"
)

const outlookEmailJobType = "outlook_send_notification_email"

func CreateInAppWithLinkedEmailMirror(
	ctx context.Context,
	exec db.DBTX,
	userID uuid.UUID,
	typ string,
	title string,
	body string,
	relatedEntityType string,
	relatedEntityID uuid.UUID,
	createdAt time.Time,
) error {
	if _, err := exec.ExecContext(ctx, `
		insert into notifications (
			id, user_id, channel, type, title, body, status, related_entity_type, related_entity_id, created_at
		)
		values ($1, $2, 'in_app', $3, $4, $5, 'pending', $6, $7, $8)
	`, uuid.New(), userID, typ, title, body, relatedEntityType, relatedEntityID, createdAt); err != nil {
		return err
	}

	var accountID uuid.UUID
	err := exec.QueryRowContext(ctx, `
		select id
		from outlook_accounts
		where user_id = $1
		  and status = 'active'
		  and system_email_enabled = true
		order by updated_at desc
		limit 1
	`, userID).Scan(&accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	emailNotificationID := uuid.New()
	if _, err := exec.ExecContext(ctx, `
		insert into notifications (
			id, user_id, channel, type, title, body, status, related_entity_type, related_entity_id, created_at
		)
		values ($1, $2, 'email', $3, $4, $5, 'pending', $6, $7, $8)
	`, emailNotificationID, userID, typ, title, body, relatedEntityType, relatedEntityID, createdAt); err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]any{
		"notification_id": emailNotificationID,
		"account_id":      accountID,
	})
	if err != nil {
		return err
	}

	_, err = exec.ExecContext(ctx, `
		insert into background_jobs (
			id, queue, job_type, payload, idempotency_key, status, run_at,
			attempt, max_attempts, created_at, updated_at
		)
		values ($1, 'integrations', $2, $3::jsonb, null, 'pending', $4, 0, 5, $4, $4)
	`, uuid.New(), outlookEmailJobType, string(payload), createdAt)
	return err
}
