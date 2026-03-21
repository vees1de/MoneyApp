package yandex

import (
	"context"
	"strings"

	"moneyapp/backend/internal/core/auth"
	"moneyapp/backend/internal/platform/httpx"
)

type Verifier struct {
	allowInsecure bool
	clientID      string
}

func NewVerifier(clientID string, allowInsecure bool) *Verifier {
	return &Verifier{
		allowInsecure: allowInsecure,
		clientID:      clientID,
	}
}

func (v *Verifier) Verify(_ context.Context, input auth.YandexVerificationInput) (auth.VerifiedIdentity, error) {
	if !v.allowInsecure && strings.TrimSpace(v.clientID) == "" {
		return auth.VerifiedIdentity{}, httpx.Unauthorized("yandex_verification_unavailable", "yandex oauth is not configured")
	}

	if input.ProviderUserID == nil || strings.TrimSpace(*input.ProviderUserID) == "" {
		return auth.VerifiedIdentity{}, httpx.BadRequest("provider_user_id_required", "provider_user_id is required")
	}

	return auth.VerifiedIdentity{
		Provider:       auth.ProviderYandex,
		ProviderUserID: *input.ProviderUserID,
		Email:          input.Email,
		DisplayName:    input.DisplayName,
		AvatarURL:      input.AvatarURL,
		AccessMeta: map[string]any{
			"mode": verificationMode(v.allowInsecure),
		},
	}, nil
}

func verificationMode(allowInsecure bool) string {
	if allowInsecure {
		return "development-pass-through"
	}
	return "oauth"
}
