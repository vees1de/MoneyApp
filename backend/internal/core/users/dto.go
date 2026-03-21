package users

type MeResponse struct {
	User User `json:"user"`
}

type UpdatePreferencesRequest struct {
	Timezone            *string `json:"timezone"`
	BaseCurrency        *string `json:"base_currency"`
	OnboardingCompleted *bool   `json:"onboarding_completed"`
	WeeklyReviewWeekday *int    `json:"weekly_review_weekday"`
	WeeklyReviewHour    *int    `json:"weekly_review_hour"`
}
