package telegram

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"moneyapp/backend/internal/core/auth"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerifierVerify(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	const (
		clientID = "8521897198"
		nonce    = "nonce-123"
		kid      = "telegram-test-key"
	)

	now := time.Unix(1774452027, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		payload := jwkSet{
			Keys: []jwk{publicJWK(kid, &privateKey.PublicKey)},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	verifier := newVerifier(clientID, issuer, server.URL, server.Client(), func() time.Time {
		return now
	})

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, idTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{clientID},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    issuer,
			Subject:   "telegram-subject-42",
		},
		TelegramID:        777000,
		Name:              "Local Dev",
		Nonce:             nonce,
		Picture:           "https://cdn.example/avatar.png",
		PreferredUsername: "localdev",
	})
	token.Header["kid"] = kid

	idToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	verified, err := verifier.Verify(context.Background(), auth.TelegramVerificationInput{
		IDToken: &idToken,
		Nonce:   stringPtr(nonce),
	})
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}

	if verified.Provider != auth.ProviderTelegram {
		t.Fatalf("provider = %q, want %q", verified.Provider, auth.ProviderTelegram)
	}

	if verified.ProviderUserID != "777000" {
		t.Fatalf("provider user id = %q, want 777000", verified.ProviderUserID)
	}

	if verified.DisplayName == nil || *verified.DisplayName != "Local Dev" {
		t.Fatalf("display name = %v, want Local Dev", verified.DisplayName)
	}

	if verified.AvatarURL == nil || *verified.AvatarURL != "https://cdn.example/avatar.png" {
		t.Fatalf("avatar url = %v, want avatar url", verified.AvatarURL)
	}
}

func TestVerifierRejectsNonceMismatch(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	const (
		clientID = "8521897198"
		kid      = "telegram-test-key"
	)

	now := time.Unix(1774452027, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		payload := jwkSet{
			Keys: []jwk{publicJWK(kid, &privateKey.PublicKey)},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	verifier := newVerifier(clientID, issuer, server.URL, server.Client(), func() time.Time {
		return now
	})

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, idTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{clientID},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    issuer,
			Subject:   "telegram-subject-42",
		},
		Nonce: "nonce-from-telegram",
	})
	token.Header["kid"] = kid

	idToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	_, err = verifier.Verify(context.Background(), auth.TelegramVerificationInput{
		IDToken: &idToken,
		Nonce:   stringPtr("nonce-from-client"),
	})
	if err == nil {
		t.Fatal("verify token: expected error")
	}

	appError, ok := err.(*httpx.AppError)
	if !ok {
		t.Fatalf("error type = %T, want *httpx.AppError", err)
	}

	if appError.Code != "telegram_invalid_nonce" {
		t.Fatalf("error code = %q, want telegram_invalid_nonce", appError.Code)
	}
}

func publicJWK(kid string, publicKey *rsa.PublicKey) jwk {
	return jwk{
		Alg: "RS256",
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
		Kid: kid,
		Kty: "RSA",
		N:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
		Use: "sig",
	}
}

func stringPtr(value string) *string {
	return &value
}
