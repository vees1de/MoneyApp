package yandex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"moneyapp/backend/internal/core/auth"
	"moneyapp/backend/internal/platform/httpx"
)

const (
	defaultTokenURL    = "https://oauth.yandex.com/token"
	defaultUserInfoURL = "https://login.yandex.ru/info"
	avatarSize         = "islands-200"
)

type Verifier struct {
	allowInsecure bool
	clientID      string
	clientSecret  string
	redirectURI   string
	httpClient    *http.Client
	tokenURL      string
	userInfoURL   string
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

type userInfoResponse struct {
	ID              string            `json:"id"`
	Login           string            `json:"login"`
	DisplayName     string            `json:"display_name"`
	RealName        string            `json:"real_name"`
	DefaultEmail    string            `json:"default_email"`
	DefaultAvatarID string            `json:"default_avatar_id"`
	IsAvatarEmpty   bool              `json:"is_avatar_empty"`
	ClientID        string            `json:"client_id"`
	Psuid           string            `json:"psuid"`
	DefaultPhone    *defaultPhoneInfo `json:"default_phone"`
}

type defaultPhoneInfo struct {
	ID     int64  `json:"id"`
	Number string `json:"number"`
}

func NewVerifier(clientID, clientSecret, redirectURI string, allowInsecure bool) *Verifier {
	return newVerifier(
		clientID,
		clientSecret,
		redirectURI,
		allowInsecure,
		&http.Client{Timeout: 5 * time.Second},
		defaultTokenURL,
		defaultUserInfoURL,
	)
}

func newVerifier(
	clientID string,
	clientSecret string,
	redirectURI string,
	allowInsecure bool,
	httpClient *http.Client,
	tokenURL string,
	userInfoURL string,
) *Verifier {
	return &Verifier{
		allowInsecure: allowInsecure,
		clientID:      strings.TrimSpace(clientID),
		clientSecret:  strings.TrimSpace(clientSecret),
		redirectURI:   strings.TrimSpace(redirectURI),
		httpClient:    httpClient,
		tokenURL:      tokenURL,
		userInfoURL:   userInfoURL,
	}
}

func (v *Verifier) Verify(ctx context.Context, input auth.YandexVerificationInput) (auth.VerifiedIdentity, error) {
	if code := trimPointer(input.Code); code != "" {
		return v.verifyWithAuthorizationCode(ctx, code)
	}

	if !v.allowInsecure {
		return auth.VerifiedIdentity{}, httpx.BadRequest("code_required", "code is required")
	}

	providerUserID := trimPointer(input.ProviderUserID)
	if providerUserID == "" {
		return auth.VerifiedIdentity{}, httpx.BadRequest("provider_user_id_required", "provider_user_id is required")
	}

	return auth.VerifiedIdentity{
		Provider:       auth.ProviderYandex,
		ProviderUserID: providerUserID,
		Email:          input.Email,
		DisplayName:    input.DisplayName,
		AvatarURL:      input.AvatarURL,
		AccessMeta: map[string]any{
			"mode": verificationMode(v.allowInsecure),
		},
	}, nil
}

func (v *Verifier) verifyWithAuthorizationCode(ctx context.Context, code string) (auth.VerifiedIdentity, error) {
	if strings.TrimSpace(v.clientID) == "" || strings.TrimSpace(v.clientSecret) == "" || strings.TrimSpace(v.redirectURI) == "" {
		return auth.VerifiedIdentity{}, httpx.Unauthorized("yandex_verification_unavailable", "yandex oauth is not configured")
	}

	token, err := v.exchangeCode(ctx, code)
	if err != nil {
		return auth.VerifiedIdentity{}, err
	}

	userInfo, err := v.fetchUserInfo(ctx, token.AccessToken)
	if err != nil {
		return auth.VerifiedIdentity{}, err
	}

	providerUserID := strings.TrimSpace(userInfo.ID)
	if providerUserID == "" {
		return auth.VerifiedIdentity{}, httpx.Unauthorized("yandex_invalid_userinfo", "yandex user info is invalid")
	}

	email := firstNonEmptyPointer(userInfo.DefaultEmail)
	displayName := firstNonEmptyPointer(userInfo.RealName, userInfo.DisplayName, userInfo.Login)
	avatarURL := buildAvatarURL(userInfo.DefaultAvatarID, userInfo.IsAvatarEmpty)

	accessMeta := map[string]any{
		"mode":       verificationMode(v.allowInsecure),
		"token_type": token.TokenType,
	}
	if token.ExpiresIn > 0 {
		accessMeta["expires_in"] = token.ExpiresIn
	}
	if token.Scope != "" {
		accessMeta["scope"] = token.Scope
	}
	if userInfo.Login != "" {
		accessMeta["login"] = userInfo.Login
	}
	if userInfo.ClientID != "" {
		accessMeta["client_id"] = userInfo.ClientID
	}
	if userInfo.Psuid != "" {
		accessMeta["psuid"] = userInfo.Psuid
	}
	if userInfo.DefaultPhone != nil && strings.TrimSpace(userInfo.DefaultPhone.Number) != "" {
		accessMeta["phone_number"] = userInfo.DefaultPhone.Number
		accessMeta["phone_id"] = userInfo.DefaultPhone.ID
	}

	return auth.VerifiedIdentity{
		Provider:       auth.ProviderYandex,
		ProviderUserID: providerUserID,
		Email:          email,
		DisplayName:    displayName,
		AvatarURL:      avatarURL,
		AccessMeta:     accessMeta,
	}, nil
}

func (v *Verifier) exchangeCode(ctx context.Context, code string) (tokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", v.clientID)
	form.Set("client_secret", v.clientSecret)
	form.Set("redirect_uri", v.redirectURI)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		v.tokenURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return tokenResponse{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := v.httpClient.Do(request)
	if err != nil {
		return tokenResponse{}, err
	}
	defer response.Body.Close()

	var payload tokenResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return tokenResponse{}, err
	}

	if response.StatusCode != http.StatusOK {
		code := "yandex_token_exchange_failed"
		if response.StatusCode >= http.StatusInternalServerError {
			code = "yandex_token_exchange_unavailable"
		}
		return tokenResponse{}, httpx.NewError(http.StatusBadGateway, code, formatYandexError("yandex token exchange failed", payload.Error, payload.Description))
	}

	if strings.TrimSpace(payload.AccessToken) == "" {
		return tokenResponse{}, httpx.NewError(http.StatusBadGateway, "yandex_token_exchange_invalid", "yandex token exchange returned an empty access token")
	}

	return payload, nil
}

func (v *Verifier) fetchUserInfo(ctx context.Context, accessToken string) (userInfoResponse, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, v.userInfoURL, nil)
	if err != nil {
		return userInfoResponse{}, err
	}
	request.Header.Set("Authorization", "OAuth "+accessToken)

	response, err := v.httpClient.Do(request)
	if err != nil {
		return userInfoResponse{}, err
	}
	defer response.Body.Close()

	var payload userInfoResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return userInfoResponse{}, err
	}

	if response.StatusCode != http.StatusOK {
		return userInfoResponse{}, httpx.NewError(http.StatusBadGateway, "yandex_userinfo_failed", "yandex user info request failed")
	}

	return payload, nil
}

func buildAvatarURL(avatarID string, isAvatarEmpty bool) *string {
	trimmedID := strings.TrimSpace(avatarID)
	if trimmedID == "" || isAvatarEmpty {
		return nil
	}

	url := fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/%s", trimmedID, avatarSize)
	return &url
}

func firstNonEmptyPointer(values ...string) *string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return &trimmed
		}
	}

	return nil
}

func trimPointer(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
}

func formatYandexError(prefix, upstreamCode, upstreamDescription string) string {
	upstreamCode = strings.TrimSpace(upstreamCode)
	upstreamDescription = strings.TrimSpace(upstreamDescription)
	switch {
	case upstreamCode != "" && upstreamDescription != "":
		return fmt.Sprintf("%s: %s (%s)", prefix, upstreamCode, upstreamDescription)
	case upstreamCode != "":
		return fmt.Sprintf("%s: %s", prefix, upstreamCode)
	case upstreamDescription != "":
		return fmt.Sprintf("%s: %s", prefix, upstreamDescription)
	default:
		return prefix
	}
}

func verificationMode(allowInsecure bool) string {
	if allowInsecure {
		return "development-pass-through"
	}
	return "oauth"
}
