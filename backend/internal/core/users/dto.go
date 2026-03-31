package users

import "github.com/google/uuid"

type MeResponse struct {
	Profile               Profile       `json:"profile"`
	AvailableProfileRoles []ProfileRole `json:"available_profile_roles"`
}

type ProfileRolesResponse struct {
	Items []ProfileRole `json:"items"`
}

type DevelopmentTeamsResponse struct {
	Items []DevelopmentTeam `json:"items"`
}

type UpdateProfileRequest struct {
	DisplayName *string  `json:"display_name" validate:"omitempty,max=150"`
	AvatarURL   *string  `json:"avatar_url" validate:"omitempty,url,max=2048"`
	RoleCodes   []string `json:"role_codes" validate:"omitempty,max=20,dive,required,max=60"`
}

type CreateDevelopmentTeamRequest struct {
	Name          string      `json:"name" validate:"required,max=120"`
	Description   *string     `json:"description" validate:"omitempty,max=1000"`
	LeadUserID    *uuid.UUID  `json:"lead_user_id,omitempty"`
	MemberUserIDs []uuid.UUID `json:"member_user_ids" validate:"omitempty,max=50"`
}
