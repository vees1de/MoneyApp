package telegram

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
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

	if input.AuthDate <= 0 {
		return auth.VerifiedIdentity{}, httpx.BadRequest("auth_date_required", "auth_date is required")
	}

	hash := ""
	if input.Hash != nil {
		hash = strings.TrimSpace(*input.Hash)
	}

	if v.allowInsecure && hash == "dev-mode" {
		return verifiedIdentity(input, "development-pass-through"), nil
	}

	if strings.TrimSpace(v.botToken) == "" {
		return auth.VerifiedIdentity{}, httpx.Unauthorized("telegram_verification_unavailable", "telegram verification is not configured")
	}

	if hash == "" {
		return auth.VerifiedIdentity{}, httpx.BadRequest("hash_required", "hash is required")
	}

	ok, err := verifyHash(input, v.botToken, hash)
	if err != nil {
		return auth.VerifiedIdentity{}, httpx.BadRequest("telegram_verification_failed", "telegram payload is malformed")
	}
	if !ok {
		return auth.VerifiedIdentity{}, httpx.Unauthorized("telegram_invalid_signature", "telegram payload signature is invalid")
	}

	return verifiedIdentity(input, "provider-signed"), nil
}

func verifyHash(input auth.TelegramVerificationInput, botToken, receivedHash string) (bool, error) {
	dataCheckString, err := buildDataCheckString(input)
	if err != nil {
		return false, err
	}

	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	_, _ = mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(expectedHash), []byte(strings.ToLower(receivedHash))) == 1, nil
}

func buildDataCheckString(input auth.TelegramVerificationInput) (string, error) {
	fields := map[string]string{
		"auth_date": strconv.FormatInt(input.AuthDate, 10),
		"id":        input.ProviderUserID,
	}

	if value := optionalValue(input.FirstName); value != "" {
		fields["first_name"] = value
	}
	if value := optionalValue(input.LastName); value != "" {
		fields["last_name"] = value
	}
	if value := optionalValue(input.PhotoURL); value != "" {
		fields["photo_url"] = value
	}
	if value := optionalValue(input.Username); value != "" {
		fields["username"] = value
	}

	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		if strings.ContainsRune(fields[key], '\n') {
			return "", fmt.Errorf("field %s contains newline", key)
		}
		lines = append(lines, key+"="+fields[key])
	}

	return strings.Join(lines, "\n"), nil
}

func verifiedIdentity(input auth.TelegramVerificationInput, mode string) auth.VerifiedIdentity {
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
			"mode":      mode,
		},
	}
}

func firstNonNil(values ...*string) *string {
	for _, value := range values {
		if value != nil && strings.TrimSpace(*value) != "" {
			return value
		}
	}

	return nil
}

func optionalValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
