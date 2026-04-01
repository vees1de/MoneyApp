package outlook

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             uuid.UUID  `json:"user_id"`
	ExternalAccountID  string     `json:"external_account_id"`
	Email              string     `json:"email"`
	AccessToken        string     `json:"-"`
	RefreshToken       string     `json:"-"`
	TokenExpiresAt     time.Time  `json:"token_expires_at"`
	Scope              *string    `json:"scope,omitempty"`
	Status             string     `json:"status"`
	AuthMode           string     `json:"auth_mode"`
	SystemEmailEnabled bool       `json:"system_email_enabled"`
	LastSyncAt         *time.Time `json:"last_sync_at,omitempty"`
	LastMailSyncAt     *time.Time `json:"last_mail_sync_at,omitempty"`
	LastCalendarSyncAt *time.Time `json:"last_calendar_sync_at,omitempty"`
	LastError          *string    `json:"last_error,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type IntegrationStatus struct {
	GraphConfigured bool     `json:"graph_configured"`
	Connected       bool     `json:"connected"`
	Account         *Account `json:"account,omitempty"`
}

type ConnectResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

type ManualConnectRequest struct {
	AccessToken        string  `json:"access_token"`
	RefreshToken       *string `json:"refresh_token,omitempty"`
	SystemEmailEnabled *bool   `json:"system_email_enabled,omitempty"`
}

type UpdateSettingsRequest struct {
	SystemEmailEnabled bool `json:"system_email_enabled"`
}

type SyncResponse struct {
	Status         IntegrationStatus `json:"status"`
	MessagesSynced int               `json:"messages_synced"`
	EventsSynced   int               `json:"events_synced"`
}

type TestEmailRequest struct {
	Subject *string `json:"subject,omitempty"`
	Body    *string `json:"body,omitempty"`
}

type TestEmailResponse struct {
	Recipient string `json:"recipient"`
	Queued    bool   `json:"queued"`
}

type OutlookMessage struct {
	ID                uuid.UUID `json:"id"`
	ExternalMessageID string    `json:"external_message_id"`
	ConversationID    *string   `json:"conversation_id,omitempty"`
	Subject           string    `json:"subject"`
	SenderEmail       *string   `json:"sender_email,omitempty"`
	SenderName        *string   `json:"sender_name,omitempty"`
	ReceivedAt        time.Time `json:"received_at"`
	IsRead            bool      `json:"is_read"`
	BodyPreview       *string   `json:"body_preview,omitempty"`
	WebLink           *string   `json:"web_link,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type OutlookEvent struct {
	ID              uuid.UUID `json:"id"`
	ExternalEventID *string   `json:"external_event_id,omitempty"`
	Title           string    `json:"title"`
	StartAt         time.Time `json:"start_at"`
	EndAt           time.Time `json:"end_at"`
	Timezone        *string   `json:"timezone,omitempty"`
	Status          string    `json:"status"`
	Location        *string   `json:"location,omitempty"`
	WebLink         *string   `json:"web_link,omitempty"`
	OrganizerEmail  *string   `json:"organizer_email,omitempty"`
	OrganizerName   *string   `json:"organizer_name,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type EmailNotification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      string
	Title     string
	Body      string
	Status    string
	CreatedAt time.Time
}
