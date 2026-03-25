package telegram

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"moneyapp/backend/internal/core/auth"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/golang-jwt/jwt/v5"
)

const (
	issuer   = "https://oauth.telegram.org"
	jwksURL  = issuer + "/.well-known/jwks.json"
	cacheTTL = 5 * time.Minute
)

type Verifier struct {
	clientID   string
	httpClient *http.Client
	clock      func() time.Time
	jwksURL    string
	issuer     string

	mu         sync.RWMutex
	cachedAt   time.Time
	cachedKeys map[string]*rsa.PublicKey
}

type jwkSet struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Alg string `json:"alg"`
	E   string `json:"e"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	Use string `json:"use"`
}

type idTokenClaims struct {
	jwt.RegisteredClaims
	TelegramID        int64  `json:"id"`
	Name              string `json:"name"`
	Nonce             string `json:"nonce"`
	PhoneNumber       string `json:"phone_number"`
	Picture           string `json:"picture"`
	PreferredUsername string `json:"preferred_username"`
}

func NewVerifier(clientID string) *Verifier {
	return newVerifier(
		clientID,
		issuer,
		jwksURL,
		&http.Client{Timeout: 5 * time.Second},
		time.Now,
	)
}

func newVerifier(
	clientID string,
	tokenIssuer string,
	keysURL string,
	httpClient *http.Client,
	now func() time.Time,
) *Verifier {
	return &Verifier{
		clientID:   clientID,
		httpClient: httpClient,
		clock:      now,
		jwksURL:    keysURL,
		issuer:     tokenIssuer,
	}
}

func (v *Verifier) Verify(ctx context.Context, input auth.TelegramVerificationInput) (auth.VerifiedIdentity, error) {
	if strings.TrimSpace(v.clientID) == "" {
		return auth.VerifiedIdentity{}, httpx.Unauthorized("telegram_verification_unavailable", "telegram oidc is not configured")
	}

	if input.IDToken == nil || strings.TrimSpace(*input.IDToken) == "" {
		return auth.VerifiedIdentity{}, httpx.BadRequest("id_token_required", "id_token is required")
	}

	keys, err := v.getKeys(ctx)
	if err != nil {
		return auth.VerifiedIdentity{}, httpx.NewError(
			http.StatusBadGateway,
			"telegram_verification_upstream_unavailable",
			"telegram verification is temporarily unavailable",
		)
	}

	var claims idTokenClaims
	_, err = jwt.ParseWithClaims(
		strings.TrimSpace(*input.IDToken),
		&claims,
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				return nil, jwt.ErrTokenSignatureInvalid
			}

			kid, _ := token.Header["kid"].(string)
			if kid == "" {
				return nil, jwt.ErrTokenUnverifiable
			}

			key, ok := keys[kid]
			if !ok {
				return nil, jwt.ErrTokenUnverifiable
			}

			return key, nil
		},
		jwt.WithAudience(v.clientID),
		jwt.WithIssuer(v.issuer),
		jwt.WithLeeway(30*time.Second),
		jwt.WithTimeFunc(v.clock),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
	)
	if err != nil {
		return auth.VerifiedIdentity{}, httpx.Unauthorized("telegram_invalid_token", "telegram id_token is invalid")
	}

	if input.Nonce != nil {
		expectedNonce := strings.TrimSpace(*input.Nonce)
		if expectedNonce != "" && expectedNonce != claims.Nonce {
			return auth.VerifiedIdentity{}, httpx.Unauthorized("telegram_invalid_nonce", "telegram nonce is invalid")
		}
	}

	providerUserID := strings.TrimSpace(claims.Subject)
	if claims.TelegramID > 0 {
		providerUserID = strconv.FormatInt(claims.TelegramID, 10)
	}
	if providerUserID == "" {
		return auth.VerifiedIdentity{}, httpx.Unauthorized("telegram_invalid_token", "telegram id_token is invalid")
	}

	displayName := firstNonEmptyPointer(claims.Name, claims.PreferredUsername)
	avatarURL := firstNonEmptyPointer(claims.Picture)

	accessMeta := map[string]any{
		"issuer":  claims.Issuer,
		"mode":    "oidc",
		"subject": claims.Subject,
	}
	if claims.IssuedAt != nil {
		accessMeta["auth_date"] = claims.IssuedAt.Unix()
	}
	if claims.TelegramID > 0 {
		accessMeta["telegram_id"] = claims.TelegramID
	}
	if claims.PreferredUsername != "" {
		accessMeta["username"] = claims.PreferredUsername
	}
	if claims.PhoneNumber != "" {
		accessMeta["phone_number"] = claims.PhoneNumber
	}
	if claims.Nonce != "" {
		accessMeta["nonce"] = claims.Nonce
	}

	return auth.VerifiedIdentity{
		Provider:       auth.ProviderTelegram,
		ProviderUserID: providerUserID,
		DisplayName:    displayName,
		AvatarURL:      avatarURL,
		AccessMeta:     accessMeta,
	}, nil
}

func (v *Verifier) getKeys(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	v.mu.RLock()
	if len(v.cachedKeys) > 0 && v.clock().Sub(v.cachedAt) < cacheTTL {
		keys := cloneKeys(v.cachedKeys)
		v.mu.RUnlock()
		return keys, nil
	}
	v.mu.RUnlock()

	v.mu.Lock()
	defer v.mu.Unlock()

	if len(v.cachedKeys) > 0 && v.clock().Sub(v.cachedAt) < cacheTTL {
		return cloneKeys(v.cachedKeys), nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return nil, err
	}

	response, err := v.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, httpx.NewError(http.StatusBadGateway, "telegram_jwks_bad_response", "telegram jwks endpoint returned an unexpected status")
	}

	var payload jwkSet
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	keys := make(map[string]*rsa.PublicKey, len(payload.Keys))
	for _, key := range payload.Keys {
		publicKey, err := decodeRSAPublicKey(key)
		if err != nil {
			continue
		}
		keys[key.Kid] = publicKey
	}

	if len(keys) == 0 {
		return nil, httpx.NewError(http.StatusBadGateway, "telegram_jwks_empty", "telegram jwks endpoint returned no usable keys")
	}

	v.cachedKeys = keys
	v.cachedAt = v.clock()

	return cloneKeys(keys), nil
}

func decodeRSAPublicKey(key jwk) (*rsa.PublicKey, error) {
	if key.Kid == "" || key.Kty != "RSA" || key.N == "" || key.E == "" {
		return nil, jwt.ErrTokenUnverifiable
	}

	if key.Use != "" && key.Use != "sig" {
		return nil, jwt.ErrTokenUnverifiable
	}

	if key.Alg != "" && key.Alg != jwt.SigningMethodRS256.Alg() {
		return nil, jwt.ErrTokenUnverifiable
	}

	modulusBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, err
	}

	exponentBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, err
	}

	exponent := int(new(big.Int).SetBytes(exponentBytes).Int64())
	if exponent <= 0 {
		return nil, jwt.ErrTokenUnverifiable
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(modulusBytes),
		E: exponent,
	}, nil
}

func cloneKeys(keys map[string]*rsa.PublicKey) map[string]*rsa.PublicKey {
	cloned := make(map[string]*rsa.PublicKey, len(keys))
	for kid, key := range keys {
		cloned[kid] = key
	}
	return cloned
}

func firstNonEmptyPointer(values ...string) *string {
	for _, value := range values {
		nextValue := strings.TrimSpace(value)
		if nextValue == "" {
			continue
		}

		return &nextValue
	}

	return nil
}
