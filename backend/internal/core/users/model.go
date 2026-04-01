package users

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                  uuid.UUID `json:"id"`
	Email               *string   `json:"email,omitempty"`
	DisplayName         *string   `json:"display_name,omitempty"`
	AvatarURL           *string   `json:"avatar_url,omitempty"`
	Timezone            string    `json:"timezone"`
	BaseCurrency        string    `json:"base_currency"`
	OnboardingCompleted bool      `json:"onboarding_completed"`
	WeeklyReviewWeekday int       `json:"weekly_review_weekday"`
	WeeklyReviewHour    int       `json:"weekly_review_hour"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type Profile struct {
	ID           uuid.UUID         `json:"id"`
	Email        string            `json:"email"`
	DisplayName  *string           `json:"display_name,omitempty"`
	AvatarURL    *string           `json:"avatar_url,omitempty"`
	ProfileRoles []ProfileRole     `json:"profile_roles"`
	Teams        []DevelopmentTeam `json:"teams"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type ProfileRole struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	SortOrder   int       `json:"sort_order"`
}

type DevelopmentTeam struct {
	ID              uuid.UUID               `json:"id"`
	Name            string                  `json:"name"`
	Description     *string                 `json:"description,omitempty"`
	LeadUserID      *uuid.UUID              `json:"lead_user_id,omitempty"`
	CreatedByUserID *uuid.UUID              `json:"created_by_user_id,omitempty"`
	Members         []DevelopmentTeamMember `json:"members"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

type DevelopmentTeamMember struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	IsLead      bool      `json:"is_lead"`
}

type AvatarUpload struct {
	OriginalName string
	ContentType  string
	Content      []byte
}
