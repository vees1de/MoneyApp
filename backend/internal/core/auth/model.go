package auth

import (
	"time"

	"moneyapp/backend/internal/core/users"

	"github.com/google/uuid"
)

type Provider string

const (
	ProviderTelegram Provider = "telegram"
	ProviderYandex   Provider = "yandex"
)

type Identity struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Provider       Provider  `json:"provider"`
	ProviderUserID string    `json:"provider_user_id"`
	ProviderEmail  *string   `json:"provider_email,omitempty"`
	AccessMeta     []byte    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
}

type VerifiedIdentity struct {
	Provider       Provider
	ProviderUserID string
	Email          *string
	DisplayName    *string
	AvatarURL      *string
	AccessMeta     map[string]any
}

type TelegramVerificationInput struct {
	IDToken *string
	Nonce   *string
}

type YandexVerificationInput struct {
	Code           *string
	IDToken        *string
	ProviderUserID *string
	Email          *string
	DisplayName    *string
	AvatarURL      *string
}

type LoginResult struct {
	User   users.User     `json:"user"`
	Tokens map[string]any `json:"tokens"`
	Meta   map[string]any `json:"meta,omitempty"`
}
