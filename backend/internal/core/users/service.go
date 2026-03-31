package users

import (
	"context"
	"database/sql"
	"slices"
	"strings"
	"time"

	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type Service struct {
	db   *sql.DB
	repo *Repository
}

func NewService(database *sql.DB, repo *Repository) *Service {
	return &Service{
		db:   database,
		repo: repo,
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

func (s *Service) CreateDevelopmentTeam(ctx context.Context, userID uuid.UUID, request CreateDevelopmentTeamRequest) (DevelopmentTeam, error) {
	name := strings.TrimSpace(request.Name)
	if name == "" {
		return DevelopmentTeam{}, httpx.BadRequest("validation_error", "name is required")
	}

	leadUserID := userID
	if request.LeadUserID != nil && *request.LeadUserID != uuid.Nil {
		leadUserID = *request.LeadUserID
	}

	memberUserIDs := uniqueUserIDs(append(append([]uuid.UUID{}, request.MemberUserIDs...), leadUserID, userID))
	existingUserIDs, err := s.repo.ListExistingUserIDs(ctx, memberUserIDs)
	if err != nil {
		return DevelopmentTeam{}, err
	}

	missingUserIDs := make([]string, 0, len(memberUserIDs))
	for _, memberUserID := range memberUserIDs {
		if _, ok := existingUserIDs[memberUserID]; !ok {
			missingUserIDs = append(missingUserIDs, memberUserID.String())
		}
	}
	if len(missingUserIDs) > 0 {
		slices.Sort(missingUserIDs)
		return DevelopmentTeam{}, httpx.BadRequest("users_not_found", "one or more team members were not found: "+strings.Join(missingUserIDs, ", "))
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
		return DevelopmentTeam{}, err
	}

	return s.repo.GetDevelopmentTeamByID(ctx, team.ID)
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
