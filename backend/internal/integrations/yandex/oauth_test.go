package yandex

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"moneyapp/backend/internal/core/auth"
)

func TestVerifierVerifyWithAuthorizationCode(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.Form.Get("grant_type"); got != "authorization_code" {
			t.Fatalf("unexpected grant_type: %s", got)
		}
		if got := r.Form.Get("code"); got != "test-code" {
			t.Fatalf("unexpected code: %s", got)
		}
		if got := r.Form.Get("client_id"); got != "client-123" {
			t.Fatalf("unexpected client_id: %s", got)
		}
		if got := r.Form.Get("client_secret"); got != "secret-456" {
			t.Fatalf("unexpected client_secret: %s", got)
		}
		if got := r.Form.Get("redirect_uri"); got != "https://bims.su/auth/yandex/callback" {
			t.Fatalf("unexpected redirect_uri: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"access_token":"access-123",
			"token_type":"bearer",
			"expires_in":3600,
			"scope":"login:info login:email login:avatar"
		}`))
	})
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "OAuth access-123" {
			t.Fatalf("unexpected authorization header: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"1000034426",
			"login":"ivan",
			"real_name":"Ivan Ivanov",
			"display_name":"ivan",
			"default_email":"ivan@example.com",
			"default_avatar_id":"131652443",
			"is_avatar_empty":false,
			"client_id":"client-123",
			"psuid":"1.test",
			"default_phone":{"id":12345678,"number":"+79037659418"}
		}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	verifier := newVerifier(
		"client-123",
		"secret-456",
		"https://bims.su/auth/yandex/callback",
		false,
		server.Client(),
		server.URL+"/token",
		server.URL+"/info",
	)

	identity, err := verifier.Verify(context.Background(), auth.YandexVerificationInput{
		Code: stringPtr("test-code"),
	})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if identity.Provider != auth.ProviderYandex {
		t.Fatalf("unexpected provider: %s", identity.Provider)
	}
	if identity.ProviderUserID != "1000034426" {
		t.Fatalf("unexpected provider user id: %s", identity.ProviderUserID)
	}
	if identity.Email == nil || *identity.Email != "ivan@example.com" {
		t.Fatalf("unexpected email: %#v", identity.Email)
	}
	if identity.DisplayName == nil || *identity.DisplayName != "Ivan Ivanov" {
		t.Fatalf("unexpected display name: %#v", identity.DisplayName)
	}
	if identity.AvatarURL == nil || *identity.AvatarURL != "https://avatars.yandex.net/get-yapic/131652443/islands-200" {
		t.Fatalf("unexpected avatar url: %#v", identity.AvatarURL)
	}
	if got := identity.AccessMeta["mode"]; got != "oauth" {
		t.Fatalf("unexpected mode: %#v", got)
	}
	if got := identity.AccessMeta["scope"]; got != "login:info login:email login:avatar" {
		t.Fatalf("unexpected scope: %#v", got)
	}
	if got := identity.AccessMeta["phone_number"]; got != "+79037659418" {
		t.Fatalf("unexpected phone number: %#v", got)
	}
}

func TestVerifierVerifyAllowsDevelopmentPassThrough(t *testing.T) {
	verifier := newVerifier("", "", "", true, http.DefaultClient, "", "")

	identity, err := verifier.Verify(context.Background(), auth.YandexVerificationInput{
		ProviderUserID: stringPtr("yandex_10001"),
		Email:          stringPtr("dev@example.com"),
		DisplayName:    stringPtr("Dev User"),
	})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if identity.ProviderUserID != "yandex_10001" {
		t.Fatalf("unexpected provider user id: %s", identity.ProviderUserID)
	}
	if got := identity.AccessMeta["mode"]; got != "development-pass-through" {
		t.Fatalf("unexpected mode: %#v", got)
	}
}

func stringPtr(value string) *string {
	return &value
}
