package outlook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"moneyapp/backend/internal/config"
	"moneyapp/backend/internal/platform/httpx"
)

const (
	graphBaseURL       = "https://graph.microsoft.com/v1.0"
	defaultTenantID    = "common"
	defaultMessagesTop = 15
	defaultEventsTop   = 15
)

var defaultScopes = []string{
	"openid",
	"offline_access",
	"User.Read",
	"Mail.Read",
	"Mail.Send",
	"Calendars.Read",
}

type GraphClient struct {
	httpClient     *http.Client
	tenantID       string
	clientID       string
	clientSecret   string
	redirectURI    string
	postConnectURL string
}

type tokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Scope        string
}

type graphTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type graphAPIError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type graphUser struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	Mail              string `json:"mail"`
	UserPrincipalName string `json:"userPrincipalName"`
}

type graphMessagesResponse struct {
	Value []graphMessage `json:"value"`
}

type graphMessage struct {
	ID               string `json:"id"`
	ConversationID   string `json:"conversationId"`
	Subject          string `json:"subject"`
	ReceivedDateTime string `json:"receivedDateTime"`
	IsRead           bool   `json:"isRead"`
	BodyPreview      string `json:"bodyPreview"`
	WebLink          string `json:"webLink"`
	From             struct {
		EmailAddress struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"emailAddress"`
	} `json:"from"`
}

type graphEventsResponse struct {
	Value []graphEvent `json:"value"`
}

type graphEvent struct {
	ID       string `json:"id"`
	Subject  string `json:"subject"`
	WebLink  string `json:"webLink"`
	Location struct {
		DisplayName string `json:"displayName"`
	} `json:"location"`
	Organizer struct {
		EmailAddress struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"emailAddress"`
	} `json:"organizer"`
	Start struct {
		DateTime string `json:"dateTime"`
		TimeZone string `json:"timeZone"`
	} `json:"start"`
	End struct {
		DateTime string `json:"dateTime"`
		TimeZone string `json:"timeZone"`
	} `json:"end"`
}

func NewGraphClient(cfg config.OutlookConfig) *GraphClient {
	tenantID := strings.TrimSpace(cfg.TenantID)
	if tenantID == "" {
		tenantID = defaultTenantID
	}

	return &GraphClient{
		httpClient:     &http.Client{Timeout: 15 * time.Second},
		tenantID:       tenantID,
		clientID:       strings.TrimSpace(cfg.ClientID),
		clientSecret:   strings.TrimSpace(cfg.ClientSecret),
		redirectURI:    strings.TrimSpace(cfg.RedirectURI),
		postConnectURL: strings.TrimSpace(cfg.PostConnectURL),
	}
}

func (c *GraphClient) IsConfigured() bool {
	return c.clientID != "" && c.clientSecret != "" && c.redirectURI != ""
}

func (c *GraphClient) PostConnectURL() string {
	if c.postConnectURL != "" {
		return c.postConnectURL
	}
	return "/calendar/overview"
}

func (c *GraphClient) BuildAuthorizeURL(state string) (string, error) {
	if !c.IsConfigured() {
		return "", httpx.BadRequest("outlook_oauth_not_configured", "Microsoft OAuth is not configured")
	}

	query := url.Values{}
	query.Set("client_id", c.clientID)
	query.Set("response_type", "code")
	query.Set("redirect_uri", c.redirectURI)
	query.Set("response_mode", "query")
	query.Set("scope", strings.Join(defaultScopes, " "))
	query.Set("state", state)
	query.Set("prompt", "select_account")

	return fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/authorize?%s",
		url.PathEscape(c.tenantID),
		query.Encode(),
	), nil
}

func (c *GraphClient) ExchangeCode(ctx context.Context, code string) (tokenResult, error) {
	if !c.IsConfigured() {
		return tokenResult{}, httpx.BadRequest("outlook_oauth_not_configured", "Microsoft OAuth is not configured")
	}
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", c.redirectURI)
	form.Set("scope", strings.Join(defaultScopes, " "))
	return c.exchangeToken(ctx, form)
}

func (c *GraphClient) RefreshToken(ctx context.Context, refreshToken string) (tokenResult, error) {
	if !c.IsConfigured() {
		return tokenResult{}, httpx.BadRequest("outlook_refresh_not_configured", "Microsoft OAuth refresh is not configured")
	}
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("redirect_uri", c.redirectURI)
	form.Set("scope", strings.Join(defaultScopes, " "))
	return c.exchangeToken(ctx, form)
}

func (c *GraphClient) exchangeToken(ctx context.Context, form url.Values) (tokenResult, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", url.PathEscape(c.tenantID)),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return tokenResult{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return tokenResult{}, err
	}
	defer response.Body.Close()

	var payload graphTokenResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return tokenResult{}, err
	}
	if response.StatusCode != http.StatusOK {
		message := "Microsoft token exchange failed"
		if payload.ErrorDescription != "" {
			message = payload.ErrorDescription
		}
		return tokenResult{}, httpx.NewError(http.StatusBadGateway, "outlook_token_exchange_failed", message)
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return tokenResult{}, httpx.NewError(http.StatusBadGateway, "outlook_token_exchange_invalid", "Microsoft token exchange returned an empty access token")
	}

	expiresAt := time.Now().UTC().Add(time.Duration(payload.ExpiresIn) * time.Second)
	return tokenResult{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		ExpiresAt:    expiresAt,
		Scope:        strings.TrimSpace(payload.Scope),
	}, nil
}

func (c *GraphClient) FetchMe(ctx context.Context, accessToken string) (graphUser, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, graphBaseURL+"/me?$select=id,displayName,mail,userPrincipalName", nil)
	if err != nil {
		return graphUser{}, err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)

	var payload graphUser
	if err := c.doJSON(request, &payload); err != nil {
		return graphUser{}, err
	}
	return payload, nil
}

func (c *GraphClient) ListMessages(ctx context.Context, accessToken string, limit int) ([]graphMessage, error) {
	if limit <= 0 {
		limit = defaultMessagesTop
	}
	query := url.Values{}
	query.Set("$top", strconv.Itoa(limit))
	query.Set("$select", "id,conversationId,subject,receivedDateTime,isRead,bodyPreview,webLink,from")
	query.Set("$orderby", "receivedDateTime DESC")

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, graphBaseURL+"/me/messages?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)

	var payload graphMessagesResponse
	if err := c.doJSON(request, &payload); err != nil {
		return nil, err
	}
	return payload.Value, nil
}

func (c *GraphClient) ListEvents(ctx context.Context, accessToken string, limit int) ([]graphEvent, error) {
	if limit <= 0 {
		limit = defaultEventsTop
	}
	query := url.Values{}
	query.Set("$top", strconv.Itoa(limit))
	query.Set("$select", "id,subject,start,end,webLink,location,organizer")

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, graphBaseURL+"/me/calendar/events?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)

	var payload graphEventsResponse
	if err := c.doJSON(request, &payload); err != nil {
		return nil, err
	}
	return payload.Value, nil
}

func (c *GraphClient) SendMail(ctx context.Context, accessToken string, recipientEmail string, subject string, body string) error {
	payload := map[string]any{
		"message": map[string]any{
			"subject": subject,
			"body": map[string]any{
				"contentType": "Text",
				"content":     body,
			},
			"toRecipients": []map[string]any{
				{
					"emailAddress": map[string]any{
						"address": recipientEmail,
					},
				},
			},
		},
		"saveToSentItems": true,
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, graphBaseURL+"/me/sendMail", bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted && response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("microsoft sendMail failed: %s", strings.TrimSpace(string(bodyBytes)))
	}

	return nil
}

func (c *GraphClient) doJSON(request *http.Request, target any) error {
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		rawBody, _ := io.ReadAll(io.LimitReader(response.Body, 8192))
		var payload graphAPIError
		if err := json.Unmarshal(rawBody, &payload); err == nil && strings.TrimSpace(payload.Error.Message) != "" {
			return httpx.NewError(http.StatusBadGateway, "outlook_graph_request_failed", payload.Error.Message)
		}
		return httpx.NewError(http.StatusBadGateway, "outlook_graph_request_failed", strings.TrimSpace(string(rawBody)))
	}

	if target == nil {
		return nil
	}
	return json.NewDecoder(response.Body).Decode(target)
}

func graphPrimaryEmail(user graphUser) string {
	if strings.TrimSpace(user.Mail) != "" {
		return strings.TrimSpace(user.Mail)
	}
	return strings.TrimSpace(user.UserPrincipalName)
}

func parseGraphTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty graph time")
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.9999999",
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if strings.Contains(layout, "Z07:00") {
			if parsed, err := time.Parse(layout, value); err == nil {
				return parsed.UTC(), nil
			}
			continue
		}
		if parsed, err := time.ParseInLocation(layout, value, time.UTC); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported graph time format: %s", value)
}
