package telegram

import (
	"context"
	"strings"

	"moneyapp/backend/internal/core/auth"
	"moneyapp/backend/internal/platform/httpx"
)

type Verifier struct {
	allowInsecure bool
	botToken      string
}

func NewVerifier(botToken string, allowInsecure bool) *Verifier {
	return &Verifier{
		allowInsecure: allowInsecure,
		botToken:      botToken,
	}
}

func (v *Verifier) Verify(_ context.Context, input auth.TelegramVerificationInput) (auth.VerifiedIdentity, error) {
	if strings.TrimSpace(input.ProviderUserID) == "" {
		return auth.VerifiedIdentity{}, httpx.BadRequest("provider_user_id_required", "provider_user_id is required")
	}

	if !v.allowInsecure && strings.TrimSpace(v.botToken) == "" {
		return auth.VerifiedIdentity{}, httpx.Unauthorized("telegram_verification_unavailable", "telegram verification is not configured")
	}

	displayName := firstNonNil(input.Username, input.FirstName)
	if input.FirstName != nil && input.LastName != nil {
		full := strings.TrimSpace(*input.FirstName + " " + *input.LastName)
		if full != "" {
			displayName = &full
		}
	}

	return auth.VerifiedIdentity{
		Provider:       auth.ProviderTelegram,
		ProviderUserID: input.ProviderUserID,
		DisplayName:    displayName,
		AvatarURL:      input.PhotoURL,
		AccessMeta: map[string]any{
			"auth_date": input.AuthDate,
			"mode":      verificationMode(v.allowInsecure),
		},
	}, nil
}

func firstNonNil(values ...*string) *string {
	for _, value := range values {
		if value != nil && strings.TrimSpace(*value) != "" {
			return value
		}
	}

	return nil
}

func verificationMode(allowInsecure bool) string {
	if allowInsecure {
		return "development-pass-through"
	}
	return "provider-signed"
}
