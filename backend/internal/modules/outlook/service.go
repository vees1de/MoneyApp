package outlook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"moneyapp/backend/internal/modules/org"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/httpx"
	"moneyapp/backend/internal/platform/worker"

	"github.com/google/uuid"
)

const (
	defaultListLimit = 15
	maxListLimit     = 50
	stateTTL         = 15 * time.Minute
)

type Service struct {
	db          *sql.DB
	repo        *Repository
	queue       *worker.Queue
	clock       clock.Clock
	orgService  *org.Service
	graph       *GraphClient
	stateSecret string
}

type signedState struct {
	UserID   string `json:"user_id"`
	IssuedAt int64  `json:"issued_at"`
}

type syncJobPayload struct {
	UserID uuid.UUID `json:"user_id"`
}

type notificationJobPayload struct {
	NotificationID uuid.UUID `json:"notification_id"`
	AccountID      uuid.UUID `json:"account_id"`
}

type accessTokenClaims struct {
	Exp int64  `json:"exp"`
	Scp string `json:"scp"`
}

func NewService(
	database *sql.DB,
	repo *Repository,
	queue *worker.Queue,
	appClock clock.Clock,
	orgService *org.Service,
	graph *GraphClient,
	stateSecret string,
) *Service {
	return &Service{
		db:          database,
		repo:        repo,
		queue:       queue,
		clock:       appClock,
		orgService:  orgService,
		graph:       graph,
		stateSecret: strings.TrimSpace(stateSecret),
	}
}

func (s *Service) Connect(principal platformauth.Principal) (ConnectResponse, error) {
	state, err := s.signState(principal.UserID)
	if err != nil {
		return ConnectResponse{}, err
	}
	authURL, err := s.graph.BuildAuthorizeURL(state)
	if err != nil {
		return ConnectResponse{}, err
	}
	return ConnectResponse{
		AuthURL: authURL,
		State:   state,
	}, nil
}

func (s *Service) Callback(ctx context.Context, state, code, oauthErr, oauthErrDescription string) (string, error) {
	if strings.TrimSpace(oauthErr) != "" {
		message := strings.TrimSpace(oauthErrDescription)
		if message == "" {
			message = strings.TrimSpace(oauthErr)
		}
		return s.redirectURL("error", message), httpx.BadRequest("outlook_oauth_failed", message)
	}
	if strings.TrimSpace(code) == "" {
		return s.redirectURL("error", "Microsoft did not return an authorization code"), httpx.BadRequest("outlook_code_missing", "Microsoft did not return an authorization code")
	}

	userID, err := s.verifyState(state)
	if err != nil {
		return s.redirectURL("error", err.Error()), err
	}

	token, err := s.graph.ExchangeCode(ctx, strings.TrimSpace(code))
	if err != nil {
		return s.redirectURL("error", err.Error()), err
	}

	me, err := s.graph.FetchMe(ctx, token.AccessToken)
	if err != nil {
		return s.redirectURL("error", err.Error()), err
	}

	email := graphPrimaryEmail(me)
	if email == "" {
		return s.redirectURL("error", "Microsoft account did not provide a mailbox address"), httpx.BadRequest("outlook_email_missing", "Microsoft account did not provide a mailbox address")
	}

	account, err := s.buildAccount(ctx, userID, me.ID, email, token.AccessToken, token.RefreshToken, token.ExpiresAt, token.Scope, "oauth", nil)
	if err != nil {
		return s.redirectURL("error", err.Error()), err
	}
	if err := s.repo.UpsertAccount(ctx, account); err != nil {
		return s.redirectURL("error", "Failed to save Outlook connection"), err
	}
	if err := s.syncLinkedEmail(ctx, userID, &account.Email); err != nil {
		return s.redirectURL("warning", "Outlook connected, but linked mailbox was not saved to employee profile"), nil
	}

	return s.redirectURL("connected", "Outlook account linked"), nil
}

func (s *Service) ManualConnect(ctx context.Context, principal platformauth.Principal, req ManualConnectRequest) (IntegrationStatus, error) {
	accessToken := strings.TrimSpace(req.AccessToken)
	if accessToken == "" {
		return IntegrationStatus{}, httpx.BadRequest("outlook_access_token_required", "Microsoft access token is required")
	}

	me, err := s.graph.FetchMe(ctx, accessToken)
	if err != nil {
		return IntegrationStatus{}, err
	}
	email := graphPrimaryEmail(me)
	if email == "" {
		return IntegrationStatus{}, httpx.BadRequest("outlook_email_missing", "Microsoft account did not provide a mailbox address")
	}

	expiresAt := s.clock.Now().Add(45 * time.Minute)
	scope := ""
	if claims, err := decodeAccessTokenClaims(accessToken); err == nil {
		if claims.Exp > 0 {
			expiresAt = time.Unix(claims.Exp, 0).UTC()
		}
		scope = strings.TrimSpace(claims.Scp)
	}

	refreshToken := ""
	if req.RefreshToken != nil {
		refreshToken = strings.TrimSpace(*req.RefreshToken)
	}

	account, err := s.buildAccount(ctx, principal.UserID, me.ID, email, accessToken, refreshToken, expiresAt, scope, "access_token", req.SystemEmailEnabled)
	if err != nil {
		return IntegrationStatus{}, err
	}
	if err := s.repo.UpsertAccount(ctx, account); err != nil {
		return IntegrationStatus{}, err
	}
	if err := s.syncLinkedEmail(ctx, principal.UserID, &account.Email); err != nil {
		return IntegrationStatus{}, err
	}

	return s.Status(ctx, principal)
}

func (s *Service) Disconnect(ctx context.Context, principal platformauth.Principal) error {
	if err := s.repo.Disconnect(ctx, principal.UserID, s.clock.Now()); err != nil {
		return err
	}
	return s.syncLinkedEmail(ctx, principal.UserID, nil)
}

func (s *Service) Status(ctx context.Context, principal platformauth.Principal) (IntegrationStatus, error) {
	account, err := s.repo.GetByUserID(ctx, principal.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return IntegrationStatus{
				GraphConfigured: s.graph.IsConfigured(),
				Connected:       false,
			}, nil
		}
		return IntegrationStatus{}, err
	}

	return IntegrationStatus{
		GraphConfigured: s.graph.IsConfigured(),
		Connected:       account.Status == "active",
		Account:         &account,
	}, nil
}

func (s *Service) Sync(ctx context.Context, principal platformauth.Principal) (SyncResponse, error) {
	account, err := s.requireAccount(ctx, principal.UserID)
	if err != nil {
		return SyncResponse{}, err
	}

	messagesSynced, eventsSynced, err := s.syncAccount(ctx, &account)
	if err != nil {
		return SyncResponse{}, err
	}

	status, err := s.Status(ctx, principal)
	if err != nil {
		return SyncResponse{}, err
	}

	return SyncResponse{
		Status:         status,
		MessagesSynced: messagesSynced,
		EventsSynced:   eventsSynced,
	}, nil
}

func (s *Service) ListMessages(ctx context.Context, principal platformauth.Principal, limit int) ([]OutlookMessage, error) {
	return s.repo.ListMessages(ctx, principal.UserID, clampLimit(limit))
}

func (s *Service) ListEvents(ctx context.Context, principal platformauth.Principal, limit int) ([]OutlookEvent, error) {
	return s.repo.ListEvents(ctx, principal.UserID, clampLimit(limit))
}

func (s *Service) UpdateSettings(ctx context.Context, principal platformauth.Principal, req UpdateSettingsRequest) (IntegrationStatus, error) {
	account, err := s.requireAccount(ctx, principal.UserID)
	if err != nil {
		return IntegrationStatus{}, err
	}
	account.SystemEmailEnabled = req.SystemEmailEnabled
	account.UpdatedAt = s.clock.Now()
	if err := s.repo.UpsertAccount(ctx, account); err != nil {
		return IntegrationStatus{}, err
	}
	return s.Status(ctx, principal)
}

func (s *Service) SendTestEmail(ctx context.Context, principal platformauth.Principal, req TestEmailRequest) (TestEmailResponse, error) {
	account, err := s.requireAccount(ctx, principal.UserID)
	if err != nil {
		return TestEmailResponse{}, err
	}

	accessToken, err := s.ensureAccessToken(ctx, &account)
	if err != nil {
		return TestEmailResponse{}, err
	}

	subject := "MoneyApp system email test"
	if req.Subject != nil && strings.TrimSpace(*req.Subject) != "" {
		subject = strings.TrimSpace(*req.Subject)
	}
	body := "Microsoft Outlook integration is active. MoneyApp can deliver system email notifications to this linked mailbox."
	if req.Body != nil && strings.TrimSpace(*req.Body) != "" {
		body = strings.TrimSpace(*req.Body)
	}

	if err := s.graph.SendMail(ctx, accessToken, account.Email, subject, body); err != nil {
		return TestEmailResponse{}, err
	}

	return TestEmailResponse{
		Recipient: account.Email,
		Queued:    false,
	}, nil
}

func (s *Service) ProcessSyncJob(ctx context.Context, job worker.Job) error {
	var payload syncJobPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}
	if payload.UserID == uuid.Nil {
		return nil
	}
	account, err := s.requireAccount(ctx, payload.UserID)
	if err != nil {
		return err
	}
	_, _, err = s.syncAccount(ctx, &account)
	return err
}

func (s *Service) ProcessNotificationEmailJob(ctx context.Context, job worker.Job) error {
	var payload notificationJobPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}
	if payload.NotificationID == uuid.Nil {
		return nil
	}

	notification, err := s.repo.GetEmailNotification(ctx, payload.NotificationID)
	if err != nil {
		return err
	}

	account, err := s.requireAccount(ctx, notification.UserID)
	if err != nil {
		if job.Attempt+1 >= job.MaxAttempts {
			_ = s.repo.MarkNotificationFailed(ctx, notification.ID)
		}
		return err
	}
	if !account.SystemEmailEnabled || account.Status != "active" {
		_ = s.repo.MarkNotificationFailed(ctx, notification.ID)
		return nil
	}

	accessToken, err := s.ensureAccessToken(ctx, &account)
	if err != nil {
		if job.Attempt+1 >= job.MaxAttempts {
			_ = s.repo.MarkNotificationFailed(ctx, notification.ID)
			errorText := err.Error()
			_ = s.repo.CreateNotificationLog(ctx, notification.ID, "failed", nil, &errorText, s.clock.Now())
		}
		return err
	}

	if err := s.graph.SendMail(ctx, accessToken, account.Email, notification.Title, notification.Body); err != nil {
		if job.Attempt+1 >= job.MaxAttempts {
			_ = s.repo.MarkNotificationFailed(ctx, notification.ID)
			errorText := err.Error()
			_ = s.repo.CreateNotificationLog(ctx, notification.ID, "failed", nil, &errorText, s.clock.Now())
		}
		return err
	}

	now := s.clock.Now()
	if err := s.repo.MarkNotificationSent(ctx, notification.ID, now); err != nil {
		return err
	}
	responsePayload := fmt.Sprintf(`{"recipient":%q}`, account.Email)
	return s.repo.CreateNotificationLog(ctx, notification.ID, "sent", &responsePayload, nil, now)
}

func (s *Service) ProcessCreateEventJob(context.Context, worker.Job) error {
	return nil
}

func (s *Service) requireAccount(ctx context.Context, userID uuid.UUID) (Account, error) {
	account, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Account{}, httpx.Conflict("outlook_not_connected", "Outlook account is not connected")
		}
		return Account{}, err
	}
	return account, nil
}

func (s *Service) syncAccount(ctx context.Context, account *Account) (int, int, error) {
	accessToken, err := s.ensureAccessToken(ctx, account)
	if err != nil {
		return 0, 0, err
	}

	graphMessages, err := s.graph.ListMessages(ctx, accessToken, defaultListLimit)
	if err != nil {
		s.markAccountFailure(ctx, account, "error", err.Error())
		return 0, 0, err
	}
	graphEvents, err := s.graph.ListEvents(ctx, accessToken, defaultListLimit)
	if err != nil {
		s.markAccountFailure(ctx, account, "error", err.Error())
		return 0, 0, err
	}

	now := s.clock.Now()
	messages := make([]OutlookMessage, 0, len(graphMessages))
	for _, raw := range graphMessages {
		receivedAt, err := parseGraphTime(raw.ReceivedDateTime)
		if err != nil || strings.TrimSpace(raw.ID) == "" {
			continue
		}
		item := OutlookMessage{
			ID:                uuid.NewSHA1(uuid.NameSpaceURL, []byte(account.UserID.String()+":message:"+raw.ID)),
			ExternalMessageID: raw.ID,
			Subject:           strings.TrimSpace(raw.Subject),
			ReceivedAt:        receivedAt,
			IsRead:            raw.IsRead,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if value := strings.TrimSpace(raw.ConversationID); value != "" {
			item.ConversationID = &value
		}
		if value := strings.TrimSpace(raw.BodyPreview); value != "" {
			item.BodyPreview = &value
		}
		if value := strings.TrimSpace(raw.WebLink); value != "" {
			item.WebLink = &value
		}
		if value := strings.TrimSpace(raw.From.EmailAddress.Address); value != "" {
			item.SenderEmail = &value
		}
		if value := strings.TrimSpace(raw.From.EmailAddress.Name); value != "" {
			item.SenderName = &value
		}
		messages = append(messages, item)
	}

	events := make([]OutlookEvent, 0, len(graphEvents))
	payloads := make(map[uuid.UUID]string, len(graphEvents))
	for _, raw := range graphEvents {
		startAt, err := parseGraphTime(raw.Start.DateTime)
		if err != nil || strings.TrimSpace(raw.ID) == "" {
			continue
		}
		endAt, err := parseGraphTime(raw.End.DateTime)
		if err != nil {
			continue
		}
		item := OutlookEvent{
			ID:        uuid.NewSHA1(uuid.NameSpaceURL, []byte(account.UserID.String()+":event:"+raw.ID)),
			Title:     strings.TrimSpace(raw.Subject),
			StartAt:   startAt,
			EndAt:     endAt,
			Status:    "scheduled",
			CreatedAt: now,
			UpdatedAt: now,
		}
		externalEventID := strings.TrimSpace(raw.ID)
		item.ExternalEventID = &externalEventID
		if value := strings.TrimSpace(raw.Start.TimeZone); value != "" {
			item.Timezone = &value
		}
		if value := strings.TrimSpace(raw.Location.DisplayName); value != "" {
			item.Location = &value
		}
		if value := strings.TrimSpace(raw.WebLink); value != "" {
			item.WebLink = &value
		}
		if value := strings.TrimSpace(raw.Organizer.EmailAddress.Address); value != "" {
			item.OrganizerEmail = &value
		}
		if value := strings.TrimSpace(raw.Organizer.EmailAddress.Name); value != "" {
			item.OrganizerName = &value
		}
		payload, _ := json.Marshal(map[string]any{
			"organizer_email": optionalString(item.OrganizerEmail),
			"organizer_name":  optionalString(item.OrganizerName),
		})
		payloads[item.ID] = string(payload)
		events = append(events, item)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].StartAt.Before(events[j].StartAt)
	})

	if err := s.repo.UpsertMessages(ctx, account.UserID, account.ID, messages); err != nil {
		return 0, 0, err
	}
	if err := s.repo.UpsertEvents(ctx, account.UserID, events, payloads); err != nil {
		return len(messages), 0, err
	}

	account.Status = "active"
	account.LastSyncAt = &now
	account.LastMailSyncAt = &now
	account.LastCalendarSyncAt = &now
	account.LastError = nil
	account.UpdatedAt = now
	if err := s.repo.UpsertAccount(ctx, *account); err != nil {
		return len(messages), len(events), err
	}

	return len(messages), len(events), nil
}

func (s *Service) ensureAccessToken(ctx context.Context, account *Account) (string, error) {
	now := s.clock.Now()
	if strings.TrimSpace(account.AccessToken) == "" {
		s.markAccountFailure(ctx, account, "error", "Microsoft access token is missing")
		return "", httpx.Conflict("outlook_token_missing", "Microsoft access token is missing")
	}
	if account.TokenExpiresAt.After(now.Add(2 * time.Minute)) {
		return account.AccessToken, nil
	}
	if strings.TrimSpace(account.RefreshToken) == "" {
		s.markAccountFailure(ctx, account, "expired", "Microsoft access token has expired. Reconnect Outlook.")
		return "", httpx.Conflict("outlook_token_expired", "Microsoft access token has expired. Reconnect Outlook.")
	}

	token, err := s.graph.RefreshToken(ctx, account.RefreshToken)
	if err != nil {
		s.markAccountFailure(ctx, account, "expired", err.Error())
		return "", err
	}

	account.AccessToken = token.AccessToken
	if strings.TrimSpace(token.RefreshToken) != "" {
		account.RefreshToken = token.RefreshToken
	}
	account.TokenExpiresAt = token.ExpiresAt
	if scope := strings.TrimSpace(token.Scope); scope != "" {
		account.Scope = &scope
	}
	account.Status = "active"
	account.LastError = nil
	account.UpdatedAt = now
	if err := s.repo.UpsertAccount(ctx, *account); err != nil {
		return "", err
	}

	return account.AccessToken, nil
}

func (s *Service) buildAccount(
	ctx context.Context,
	userID uuid.UUID,
	externalAccountID string,
	email string,
	accessToken string,
	refreshToken string,
	expiresAt time.Time,
	scope string,
	authMode string,
	overrideSystemEmailEnabled *bool,
) (Account, error) {
	now := s.clock.Now()
	account := Account{
		ID:                 uuid.New(),
		UserID:             userID,
		ExternalAccountID:  strings.TrimSpace(externalAccountID),
		Email:              strings.TrimSpace(email),
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		TokenExpiresAt:     expiresAt,
		Status:             "active",
		AuthMode:           authMode,
		SystemEmailEnabled: true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if strings.TrimSpace(scope) != "" {
		account.Scope = &scope
	}

	existing, err := s.repo.GetByUserID(ctx, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Account{}, err
	}
	if err == nil {
		account.ID = existing.ID
		account.CreatedAt = existing.CreatedAt
		account.LastSyncAt = existing.LastSyncAt
		account.LastMailSyncAt = existing.LastMailSyncAt
		account.LastCalendarSyncAt = existing.LastCalendarSyncAt
		if overrideSystemEmailEnabled == nil {
			account.SystemEmailEnabled = existing.SystemEmailEnabled
		}
	}
	if overrideSystemEmailEnabled != nil {
		account.SystemEmailEnabled = *overrideSystemEmailEnabled
	}
	return account, nil
}

func (s *Service) syncLinkedEmail(ctx context.Context, userID uuid.UUID, email *string) error {
	_, err := s.orgService.UpdateProfile(ctx, userID, org.UpdateProfileInput{OutlookEmail: email})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	return nil
}

func (s *Service) signState(userID uuid.UUID) (string, error) {
	if strings.TrimSpace(s.stateSecret) == "" {
		return "", httpx.Internal("outlook_state_secret_missing")
	}
	payload, err := json.Marshal(signedState{
		UserID:   userID.String(),
		IssuedAt: s.clock.Now().Unix(),
	})
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(s.stateSecret))
	_, _ = mac.Write([]byte(encodedPayload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encodedPayload + "." + signature, nil
}

func (s *Service) verifyState(raw string) (uuid.UUID, error) {
	parts := strings.Split(strings.TrimSpace(raw), ".")
	if len(parts) != 2 {
		return uuid.Nil, httpx.BadRequest("invalid_state", "invalid outlook state")
	}

	mac := hmac.New(sha256.New, []byte(s.stateSecret))
	_, _ = mac.Write([]byte(parts[0]))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return uuid.Nil, httpx.BadRequest("invalid_state", "invalid outlook state signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return uuid.Nil, httpx.BadRequest("invalid_state", "invalid outlook state payload")
	}

	var payload signedState
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return uuid.Nil, httpx.BadRequest("invalid_state", "invalid outlook state payload")
	}
	if payload.UserID == "" {
		return uuid.Nil, httpx.BadRequest("invalid_state", "outlook state is missing user id")
	}
	if issuedAt := time.Unix(payload.IssuedAt, 0); payload.IssuedAt <= 0 || s.clock.Now().After(issuedAt.Add(stateTTL)) {
		return uuid.Nil, httpx.BadRequest("expired_state", "outlook state has expired")
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return uuid.Nil, httpx.BadRequest("invalid_state", "outlook state user id is invalid")
	}
	return userID, nil
}

func (s *Service) redirectURL(status string, message string) string {
	target := s.graph.PostConnectURL()
	parsed, err := url.Parse(target)
	if err != nil {
		return "/calendar/overview"
	}
	query := parsed.Query()
	if strings.TrimSpace(status) != "" {
		query.Set("outlook", status)
	}
	if strings.TrimSpace(message) != "" {
		query.Set("outlook_message", message)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func (s *Service) markAccountFailure(ctx context.Context, account *Account, status string, message string) {
	text := strings.TrimSpace(message)
	if text != "" {
		account.LastError = &text
	} else {
		account.LastError = nil
	}
	account.Status = status
	account.UpdatedAt = s.clock.Now()
	_ = s.repo.UpsertAccount(ctx, *account)
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func decodeAccessTokenClaims(token string) (accessTokenClaims, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) < 2 {
		return accessTokenClaims{}, fmt.Errorf("token is not a JWT")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return accessTokenClaims{}, err
	}

	var claims accessTokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return accessTokenClaims{}, err
	}
	return claims, nil
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
