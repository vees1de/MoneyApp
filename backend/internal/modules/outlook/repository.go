package outlook

import (
	"context"
	"database/sql"
	"time"

	"moneyapp/backend/internal/platform/db"

	"github.com/google/uuid"
)

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

func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) (Account, error) {
	var item Account
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, user_id, external_account_id, email, access_token_encrypted, refresh_token_encrypted,
		       token_expires_at, scope, status, auth_mode, system_email_enabled, last_sync_at,
		       last_mail_sync_at, last_calendar_sync_at, last_error, created_at, updated_at
		from outlook_accounts
		where user_id = $1
	`, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.ExternalAccountID,
		&item.Email,
		&item.AccessToken,
		&item.RefreshToken,
		&item.TokenExpiresAt,
		&item.Scope,
		&item.Status,
		&item.AuthMode,
		&item.SystemEmailEnabled,
		&item.LastSyncAt,
		&item.LastMailSyncAt,
		&item.LastCalendarSyncAt,
		&item.LastError,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}

func (r *Repository) UpsertAccount(ctx context.Context, item Account, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into outlook_accounts (
			id, user_id, external_account_id, email, access_token_encrypted, refresh_token_encrypted,
			token_expires_at, scope, status, auth_mode, system_email_enabled, last_sync_at,
			last_mail_sync_at, last_calendar_sync_at, last_error, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		on conflict (user_id) do update
		set external_account_id = excluded.external_account_id,
		    email = excluded.email,
		    access_token_encrypted = excluded.access_token_encrypted,
		    refresh_token_encrypted = excluded.refresh_token_encrypted,
		    token_expires_at = excluded.token_expires_at,
		    scope = excluded.scope,
		    status = excluded.status,
		    auth_mode = excluded.auth_mode,
		    system_email_enabled = excluded.system_email_enabled,
		    last_sync_at = excluded.last_sync_at,
		    last_mail_sync_at = excluded.last_mail_sync_at,
		    last_calendar_sync_at = excluded.last_calendar_sync_at,
		    last_error = excluded.last_error,
		    updated_at = excluded.updated_at
	`, item.ID, item.UserID, item.ExternalAccountID, item.Email, item.AccessToken, item.RefreshToken,
		item.TokenExpiresAt, item.Scope, item.Status, item.AuthMode, item.SystemEmailEnabled, item.LastSyncAt,
		item.LastMailSyncAt, item.LastCalendarSyncAt, item.LastError, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) Disconnect(ctx context.Context, userID uuid.UUID, updatedAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update outlook_accounts
		set status = 'revoked',
		    system_email_enabled = false,
		    last_error = null,
		    updated_at = $2
		where user_id = $1
	`, userID, updatedAt)
	return err
}

func (r *Repository) ListMessages(ctx context.Context, userID uuid.UUID, limit int, exec ...db.DBTX) ([]OutlookMessage, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, external_message_id, conversation_id, subject, sender_email, sender_name,
		       received_at, is_read, body_preview, web_link, created_at, updated_at
		from outlook_messages
		where user_id = $1
		order by received_at desc
		limit $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]OutlookMessage, 0, limit)
	for rows.Next() {
		var item OutlookMessage
		if err := rows.Scan(
			&item.ID,
			&item.ExternalMessageID,
			&item.ConversationID,
			&item.Subject,
			&item.SenderEmail,
			&item.SenderName,
			&item.ReceivedAt,
			&item.IsRead,
			&item.BodyPreview,
			&item.WebLink,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) UpsertMessages(ctx context.Context, userID, accountID uuid.UUID, items []OutlookMessage, exec ...db.DBTX) error {
	base := r.base(exec...)
	for _, item := range items {
		_, err := base.ExecContext(ctx, `
			insert into outlook_messages (
				id, user_id, account_id, external_message_id, conversation_id, subject,
				sender_email, sender_name, received_at, is_read, body_preview, web_link, created_at, updated_at
			)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			on conflict (user_id, external_message_id) do update
			set account_id = excluded.account_id,
			    conversation_id = excluded.conversation_id,
			    subject = excluded.subject,
			    sender_email = excluded.sender_email,
			    sender_name = excluded.sender_name,
			    received_at = excluded.received_at,
			    is_read = excluded.is_read,
			    body_preview = excluded.body_preview,
			    web_link = excluded.web_link,
			    updated_at = excluded.updated_at
		`, item.ID, userID, accountID, item.ExternalMessageID, item.ConversationID, item.Subject,
			item.SenderEmail, item.SenderName, item.ReceivedAt, item.IsRead, item.BodyPreview, item.WebLink, item.CreatedAt, item.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListEvents(ctx context.Context, userID uuid.UUID, limit int, exec ...db.DBTX) ([]OutlookEvent, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, external_event_id, title, start_at, end_at, timezone, status,
		       location, meeting_url, nullif(payload->>'organizer_email', ''), nullif(payload->>'organizer_name', '')
		from calendar_events
		where user_id = $1
		  and provider = 'outlook'
		  and status in ('scheduled', 'updated')
		order by start_at asc
		limit $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]OutlookEvent, 0, limit)
	for rows.Next() {
		var item OutlookEvent
		if err := rows.Scan(
			&item.ID,
			&item.ExternalEventID,
			&item.Title,
			&item.StartAt,
			&item.EndAt,
			&item.Timezone,
			&item.Status,
			&item.Location,
			&item.WebLink,
			&item.OrganizerEmail,
			&item.OrganizerName,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) UpsertEvents(
	ctx context.Context,
	userID uuid.UUID,
	items []OutlookEvent,
	payloads map[uuid.UUID]string,
	exec ...db.DBTX,
) error {
	base := r.base(exec...)
	for _, item := range items {
		payload := payloads[item.ID]
		_, err := base.ExecContext(ctx, `
			insert into calendar_events (
				id, user_id, source_type, source_id, provider, external_event_id, title, start_at, end_at,
				timezone, status, meeting_url, location, payload, created_at, updated_at
			)
			values ($1, $2, 'outlook_remote', $1, 'outlook', $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12, $13)
			on conflict (id) do update
			set external_event_id = excluded.external_event_id,
			    title = excluded.title,
			    start_at = excluded.start_at,
			    end_at = excluded.end_at,
			    timezone = excluded.timezone,
			    status = excluded.status,
			    meeting_url = excluded.meeting_url,
			    location = excluded.location,
			    payload = excluded.payload,
			    updated_at = excluded.updated_at
		`, item.ID, userID, item.ExternalEventID, item.Title, item.StartAt, item.EndAt,
			item.Timezone, item.Status, item.WebLink, item.Location, payload, item.CreatedAt, item.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) GetEmailNotification(ctx context.Context, id uuid.UUID, exec ...db.DBTX) (EmailNotification, error) {
	var item EmailNotification
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, user_id, type, title, body, status, created_at
		from notifications
		where id = $1
		  and channel = 'email'
	`, id).Scan(&item.ID, &item.UserID, &item.Type, &item.Title, &item.Body, &item.Status, &item.CreatedAt)
	return item, err
}

func (r *Repository) MarkNotificationSent(ctx context.Context, id uuid.UUID, sentAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update notifications
		set status = 'sent',
		    sent_at = $2
		where id = $1
	`, id, sentAt)
	return err
}

func (r *Repository) MarkNotificationFailed(ctx context.Context, id uuid.UUID, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update notifications
		set status = 'failed'
		where id = $1
	`, id)
	return err
}

func (r *Repository) CreateNotificationLog(
	ctx context.Context,
	notificationID uuid.UUID,
	status string,
	responsePayload *string,
	errorMessage *string,
	createdAt time.Time,
	exec ...db.DBTX,
) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into notification_logs (
			id, notification_id, provider, status, response_payload, error_message, created_at
		)
		values ($1, $2, 'outlook', $3, $4::jsonb, $5, $6)
	`, uuid.New(), notificationID, status, responsePayload, errorMessage, createdAt)
	return err
}
