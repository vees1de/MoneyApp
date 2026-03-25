package auth

import (
	"moneyapp/backend/internal/core/sessions"
	"moneyapp/backend/internal/core/users"
)

type TelegramLoginRequest struct {
	ProviderUserID string  `json:"provider_user_id" validate:"required"`
	Username       *string `json:"username"`
	FirstName      *string `json:"first_name"`
	LastName       *string `json:"last_name"`
	PhotoURL       *string `json:"photo_url"`
	AuthDate       int64   `json:"auth_date"`
	Hash           *string `json:"hash"`
}

type YandexLoginRequest struct {
	Code           *string `json:"code"`
	IDToken        *string `json:"id_token"`
	ProviderUserID *string `json:"provider_user_id"`
	Email          *string `json:"email"`
	DisplayName    *string `json:"display_name"`
	AvatarURL      *string `json:"avatar_url"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type AuthResponse struct {
	User   users.User      `json:"user"`
	Tokens sessions.Tokens `json:"tokens"`
	Meta   map[string]any  `json:"meta,omitempty"`
}
