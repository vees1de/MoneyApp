package users

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

const (
	avatarPublicPrefix = "/api/uploads/"
)

type Service struct {
	db         *sql.DB
	repo       *Repository
	uploadsDir string
}

func NewService(database *sql.DB, repo *Repository, uploadsDir string) *Service {
	return &Service{
		db:         database,
		repo:       repo,
		uploadsDir: uploadsDir,
	}
}

func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (MeResponse, error) {
	return s.buildProfileResponse(ctx, userID)
}

func (s *Service) ListProfileRoles(ctx context.Context) ([]ProfileRole, error) {
	items, err := s.repo.ListProfileRoles(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []ProfileRole{}
	}

	return items, nil
}

func (s *Service) ListDevelopmentTeams(ctx context.Context, userID uuid.UUID) ([]DevelopmentTeam, error) {
	items, err := s.repo.ListDevelopmentTeamsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []DevelopmentTeam{}
	}

	return items, nil
}

func (s *Service) ListAvailableDevelopmentTeams(ctx context.Context, userID uuid.UUID) ([]DevelopmentTeam, error) {
	currentTeamID, err := s.repo.FindCurrentDevelopmentTeamIDByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if currentTeamID != nil {
		return []DevelopmentTeam{}, nil
	}

	items, err := s.repo.ListAvailableDevelopmentTeams(ctx, userID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []DevelopmentTeam{}
	}

	return items, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, request UpdateProfileRequest) (MeResponse, error) {
	availableRoles, err := s.repo.ListProfileRoles(ctx)
	if err != nil {
		return MeResponse{}, err
	}

	normalizedRoleCodes, err := normalizeRoleCodes(request.RoleCodes)
	if err != nil {
		return MeResponse{}, err
	}
	roleIDs, err := resolveProfileRoleIDs(availableRoles, normalizedRoleCodes)
	if err != nil {
		return MeResponse{}, err
	}

	now := time.Now().UTC()
	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		profile, err := s.repo.GetProfileBase(ctx, userID, tx)
		if err != nil {
			return err
		}

		if request.DisplayName != nil {
			profile.DisplayName = normalizeOptionalString(request.DisplayName)
		}
		if request.AvatarURL != nil {
			profile.AvatarURL = normalizeOptionalString(request.AvatarURL)
		}
		profile.UpdatedAt = now

		if err := s.repo.UpdateProfileFields(ctx, profile, tx); err != nil {
			return err
		}
		if request.RoleCodes != nil {
			if err := s.repo.ReplaceUserProfileRoles(ctx, userID, roleIDs, now, tx); err != nil {
				return err
			}
		}

		return nil
	})
	if IsNotFound(err) {
		return MeResponse{}, httpx.NotFound("user_not_found", "user not found")
	}
	if err != nil {
		return MeResponse{}, err
	}

	return s.buildProfileResponse(ctx, userID)
}

func (s *Service) UploadAvatar(ctx context.Context, userID uuid.UUID, upload AvatarUpload) (MeResponse, error) {
	if len(upload.Content) == 0 {
		return MeResponse{}, httpx.BadRequest("file_required", "avatar file is required")
	}
	if strings.TrimSpace(s.uploadsDir) == "" {
		return MeResponse{}, httpx.Internal("uploads_dir_not_configured")
	}
	if strings.TrimSpace(upload.BaseURL) == "" {
		return MeResponse{}, httpx.BadRequest("invalid_base_url", "base URL is required")
	}

	profile, err := s.repo.GetProfileBase(ctx, userID)
	if IsNotFound(err) {
		return MeResponse{}, httpx.NotFound("user_not_found", "user not found")
	}
	if err != nil {
		return MeResponse{}, err
	}
	previousAvatarURL := profile.AvatarURL

	ext, err := avatarExtension(upload.ContentType)
	if err != nil {
		return MeResponse{}, err
	}

	relativeKey := filepath.Join("profile-avatars", userID.String(), uuid.New().String()+ext)
	targetPath := filepath.Join(s.uploadsDir, relativeKey)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return MeResponse{}, err
	}
	if err := os.WriteFile(targetPath, upload.Content, 0o644); err != nil {
		return MeResponse{}, err
	}

	avatarURL := strings.TrimRight(upload.BaseURL, "/") + avatarPublicPrefix + path.Clean(filepath.ToSlash(relativeKey))
	profile.AvatarURL = &avatarURL
	profile.UpdatedAt = time.Now().UTC()

	updateErr := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		return s.repo.UpdateProfileFields(ctx, profile, tx)
	})
	if updateErr != nil {
		_ = os.Remove(targetPath)
		if IsNotFound(updateErr) {
			return MeResponse{}, httpx.NotFound("user_not_found", "user not found")
		}
		return MeResponse{}, updateErr
	}

	if oldAvatarPath := managedUploadFilePath(s.uploadsDir, upload.BaseURL, previousAvatarURL); oldAvatarPath != "" {
		_ = os.Remove(oldAvatarPath)
	}

	return s.buildProfileResponse(ctx, userID)
}

func (s *Service) CreateDevelopmentTeam(ctx context.Context, userID uuid.UUID, request CreateDevelopmentTeamRequest) (MeResponse, error) {
	name := strings.TrimSpace(request.Name)
	if name == "" {
		return MeResponse{}, httpx.BadRequest("validation_error", "name is required")
	}

	if err := s.ensureUserHasNoCurrentTeam(ctx, userID); err != nil {
		return MeResponse{}, err
	}

	leadUserID := userID
	if request.LeadUserID != nil && *request.LeadUserID != uuid.Nil {
		leadUserID = *request.LeadUserID
	}

	memberUserIDs := uniqueUserIDs(append(append([]uuid.UUID{}, request.MemberUserIDs...), leadUserID, userID))
	existingUserIDs, err := s.repo.ListExistingUserIDs(ctx, memberUserIDs)
	if err != nil {
		return MeResponse{}, err
	}

	missingUserIDs := make([]string, 0, len(memberUserIDs))
	for _, memberUserID := range memberUserIDs {
		if _, ok := existingUserIDs[memberUserID]; !ok {
			missingUserIDs = append(missingUserIDs, memberUserID.String())
		}
	}
	if len(missingUserIDs) > 0 {
		slices.Sort(missingUserIDs)
		return MeResponse{}, httpx.BadRequest("users_not_found", "one or more team members were not found: "+strings.Join(missingUserIDs, ", "))
	}
	if err := s.ensureUsersHaveNoCurrentTeam(ctx, memberUserIDs); err != nil {
		return MeResponse{}, err
	}

	now := time.Now().UTC()
	createdBy := userID
	team := DevelopmentTeam{
		ID:              uuid.New(),
		Name:            name,
		Description:     normalizeOptionalString(request.Description),
		LeadUserID:      &leadUserID,
		CreatedByUserID: &createdBy,
		CreatedAt:       now,
		UpdatedAt:       now,
		Members:         []DevelopmentTeamMember{},
	}

	if err := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		return s.repo.CreateDevelopmentTeam(ctx, team, memberUserIDs, tx)
	}); err != nil {
		return MeResponse{}, err
	}

	return s.buildProfileResponse(ctx, userID)
}

func (s *Service) JoinDevelopmentTeam(ctx context.Context, userID, teamID uuid.UUID) (MeResponse, error) {
	if err := s.ensureUserHasNoCurrentTeam(ctx, userID); err != nil {
		return MeResponse{}, err
	}

	team, err := s.repo.GetDevelopmentTeamByID(ctx, teamID)
	if IsNotFound(err) {
		return MeResponse{}, httpx.NotFound("development_team_not_found", "development team not found")
	}
	if err != nil {
		return MeResponse{}, err
	}

	memberExists, err := s.repo.IsDevelopmentTeamMember(ctx, teamID, userID)
	if err != nil {
		return MeResponse{}, err
	}
	if memberExists {
		return MeResponse{}, httpx.Conflict("development_team_member_exists", "user is already in this development team")
	}

	now := time.Now().UTC()
	if err := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.AddDevelopmentTeamMember(ctx, teamID, userID, now, tx); err != nil {
			return err
		}

		if team.LeadUserID == nil {
			return s.repo.UpdateDevelopmentTeamLead(ctx, teamID, &userID, now, tx)
		}

		return s.repo.TouchDevelopmentTeam(ctx, teamID, now, tx)
	}); err != nil {
		return MeResponse{}, err
	}

	return s.buildProfileResponse(ctx, userID)
}

func (s *Service) LeaveCurrentDevelopmentTeam(ctx context.Context, userID uuid.UUID) (MeResponse, error) {
	currentTeamID, err := s.repo.FindCurrentDevelopmentTeamIDByUser(ctx, userID)
	if err != nil {
		return MeResponse{}, err
	}
	if currentTeamID == nil {
		return MeResponse{}, httpx.Conflict("development_team_missing", "user is not in a development team")
	}

	team, err := s.repo.GetDevelopmentTeamByID(ctx, *currentTeamID)
	if IsNotFound(err) {
		return MeResponse{}, httpx.NotFound("development_team_not_found", "development team not found")
	}
	if err != nil {
		return MeResponse{}, err
	}

	memberUserIDs, err := s.repo.ListDevelopmentTeamMemberUserIDs(ctx, *currentTeamID)
	if err != nil {
		return MeResponse{}, err
	}

	remainingUserIDs := make([]uuid.UUID, 0, len(memberUserIDs))
	for _, memberUserID := range memberUserIDs {
		if memberUserID != userID {
			remainingUserIDs = append(remainingUserIDs, memberUserID)
		}
	}

	now := time.Now().UTC()
	if err := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.RemoveDevelopmentTeamMember(ctx, *currentTeamID, userID, tx); err != nil {
			return err
		}

		switch {
		case len(remainingUserIDs) == 0:
			return s.repo.DeactivateDevelopmentTeam(ctx, *currentTeamID, now, tx)
		case team.LeadUserID != nil && *team.LeadUserID == userID:
			newLeadUserID := remainingUserIDs[0]
			return s.repo.UpdateDevelopmentTeamLead(ctx, *currentTeamID, &newLeadUserID, now, tx)
		default:
			return s.repo.TouchDevelopmentTeam(ctx, *currentTeamID, now, tx)
		}
	}); err != nil {
		return MeResponse{}, err
	}

	return s.buildProfileResponse(ctx, userID)
}

func (s *Service) buildProfileResponse(ctx context.Context, userID uuid.UUID) (MeResponse, error) {
	profile, err := s.repo.GetProfileBase(ctx, userID)
	if IsNotFound(err) {
		return MeResponse{}, httpx.NotFound("user_not_found", "user not found")
	}
	if err != nil {
		return MeResponse{}, err
	}

	profileRoles, err := s.repo.ListUserProfileRoles(ctx, userID)
	if err != nil {
		return MeResponse{}, err
	}

	availableRoles, err := s.repo.ListProfileRoles(ctx)
	if err != nil {
		return MeResponse{}, err
	}

	teams, err := s.repo.ListDevelopmentTeamsByUser(ctx, userID)
	if err != nil {
		return MeResponse{}, err
	}

	profile.ProfileRoles = profileRoles
	profile.Teams = teams
	if profile.ProfileRoles == nil {
		profile.ProfileRoles = []ProfileRole{}
	}
	if profile.Teams == nil {
		profile.Teams = []DevelopmentTeam{}
	}
	if availableRoles == nil {
		availableRoles = []ProfileRole{}
	}

	return MeResponse{
		Profile:               profile,
		AvailableProfileRoles: availableRoles,
	}, nil
}

func (s *Service) ensureUserHasNoCurrentTeam(ctx context.Context, userID uuid.UUID) error {
	currentTeamID, err := s.repo.FindCurrentDevelopmentTeamIDByUser(ctx, userID)
	if err != nil {
		return err
	}
	if currentTeamID != nil {
		return httpx.Conflict("development_team_exists", "user already has a current development team")
	}

	return nil
}

func (s *Service) ensureUsersHaveNoCurrentTeam(ctx context.Context, userIDs []uuid.UUID) error {
	conflictedUserIDs := make([]string, 0)
	for _, userID := range userIDs {
		currentTeamID, err := s.repo.FindCurrentDevelopmentTeamIDByUser(ctx, userID)
		if err != nil {
			return err
		}
		if currentTeamID != nil {
			conflictedUserIDs = append(conflictedUserIDs, userID.String())
		}
	}
	if len(conflictedUserIDs) > 0 {
		slices.Sort(conflictedUserIDs)
		return httpx.Conflict("development_team_exists", "one or more users already participate in a development team: "+strings.Join(conflictedUserIDs, ", "))
	}

	return nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}

	return &normalized
}

func normalizeRoleCodes(roleCodes []string) ([]string, error) {
	if roleCodes == nil {
		return nil, nil
	}

	result := make([]string, 0, len(roleCodes))
	seen := make(map[string]struct{}, len(roleCodes))
	for _, code := range roleCodes {
		normalized := strings.ToLower(strings.TrimSpace(code))
		if normalized == "" {
			return nil, httpx.BadRequest("validation_error", "role_codes must not contain empty values")
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	return result, nil
}

func resolveProfileRoleIDs(availableRoles []ProfileRole, roleCodes []string) ([]uuid.UUID, error) {
	if roleCodes == nil {
		return nil, nil
	}

	roleMap := make(map[string]uuid.UUID, len(availableRoles))
	for _, role := range availableRoles {
		roleMap[role.Code] = role.ID
	}

	result := make([]uuid.UUID, 0, len(roleCodes))
	for _, code := range roleCodes {
		roleID, ok := roleMap[code]
		if !ok {
			return nil, httpx.BadRequest("invalid_profile_role", "unknown profile role: "+code)
		}
		result = append(result, roleID)
	}

	return result, nil
}

func uniqueUserIDs(items []uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(items))
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if item == uuid.Nil {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}

	return result
}

func avatarExtension(contentType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp":
		return ".webp", nil
	case "image/gif":
		return ".gif", nil
	default:
		return "", httpx.BadRequest("unsupported_file_type", fmt.Sprintf("unsupported avatar content type: %s", contentType))
	}
}

func managedUploadFilePath(uploadsDir string, baseURL string, assetURL *string) string {
	if assetURL == nil || strings.TrimSpace(*assetURL) == "" {
		return ""
	}

	trimmedBase := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	trimmedOld := strings.TrimSpace(*assetURL)
	prefix := trimmedBase + avatarPublicPrefix
	if !strings.HasPrefix(trimmedOld, prefix) {
		return ""
	}

	relative := strings.TrimPrefix(trimmedOld, prefix)
	if relative == "" {
		return ""
	}

	candidate := filepath.Clean(filepath.Join(uploadsDir, filepath.FromSlash(relative)))
	if !strings.HasPrefix(candidate, filepath.Clean(uploadsDir)+string(os.PathSeparator)) &&
		filepath.Clean(candidate) != filepath.Clean(uploadsDir) {
		return ""
	}

	return candidate
}
