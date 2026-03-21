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
