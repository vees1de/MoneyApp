package telegram

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"moneyapp/backend/internal/core/auth"
	"moneyapp/backend/internal/platform/httpx"
)

func TestVerifierVerifySignedPayload(t *testing.T) {
	const botToken = "telegram-bot-token"

	input := auth.TelegramVerificationInput{
		ProviderUserID: "987654321",
		Username:       stringPtr("localdev"),
		FirstName:      stringPtr("Local"),
		LastName:       stringPtr("User"),
		PhotoURL:       stringPtr("https://example.com/avatar.png"),
		AuthDate:       1774452027,
	}

	dataCheckString, err := buildDataCheckString(input)
	if err != nil {
		t.Fatalf("build data check string: %v", err)
	}

	hash := signHash(dataCheckString, botToken)
	input.Hash = &hash

	verified, err := NewVerifier(botToken, false).Verify(context.Background(), input)
	if err != nil {
		t.Fatalf("verify payload: %v", err)
	}

	if verified.Provider != auth.ProviderTelegram {
		t.Fatalf("provider = %q, want %q", verified.Provider, auth.ProviderTelegram)
	}

	if verified.ProviderUserID != "987654321" {
		t.Fatalf("provider user id = %q, want %q", verified.ProviderUserID, "987654321")
	}

	if verified.DisplayName == nil || *verified.DisplayName != "Local User" {
		t.Fatalf("display name = %v, want Local User", verified.DisplayName)
	}

	if verified.AvatarURL == nil || *verified.AvatarURL != "https://example.com/avatar.png" {
		t.Fatalf("avatar url = %v, want avatar url", verified.AvatarURL)
	}
}

func TestVerifierRejectsInvalidSignature(t *testing.T) {
	hash := "deadbeef"

	_, err := NewVerifier("telegram-bot-token", false).Verify(context.Background(), auth.TelegramVerificationInput{
		ProviderUserID: "987654321",
		AuthDate:       1774452027,
		Hash:           &hash,
	})
	if err == nil {
		t.Fatal("verify payload: expected error")
	}

	appError, ok := err.(*httpx.AppError)
	if !ok {
		t.Fatalf("error type = %T, want *httpx.AppError", err)
	}

	if appError.Code != "telegram_invalid_signature" {
		t.Fatalf("error code = %q, want telegram_invalid_signature", appError.Code)
	}
}

func TestVerifierAllowsDevMode(t *testing.T) {
	hash := "dev-mode"

	verified, err := NewVerifier("", true).Verify(context.Background(), auth.TelegramVerificationInput{
		ProviderUserID: "987654321",
		FirstName:      stringPtr("Local"),
		AuthDate:       1774452027,
		Hash:           &hash,
	})
	if err != nil {
		t.Fatalf("verify dev payload: %v", err)
	}

	if verified.ProviderUserID != "987654321" {
		t.Fatalf("provider user id = %q, want 987654321", verified.ProviderUserID)
	}
}

func signHash(dataCheckString, botToken string) string {
	secretKey := sha256Sum(botToken)
	return hmacHexSHA256(dataCheckString, secretKey)
}

func sha256Sum(input string) []byte {
	sum := sha256.Sum256([]byte(input))
	return sum[:]
}

func hmacHexSHA256(message string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func stringPtr(value string) *string {
	return &value
}
