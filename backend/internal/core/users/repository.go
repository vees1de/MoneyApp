package users

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"moneyapp/backend/internal/platform/db"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) base(exec ...db.DBTX) db.DBTX {
	if len(exec) > 0 && exec[0] != nil {
		return exec[0]
	}

	return r.db
}

func (r *Repository) Create(ctx context.Context, user User, exec ...db.DBTX) error {
	query := `
		insert into users (
			id, email, display_name, avatar_url, timezone, base_currency,
			onboarding_completed, weekly_review_weekday, weekly_review_hour, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.base(exec...).ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.DisplayName,
		user.AvatarURL,
		user.Timezone,
		user.BaseCurrency,
		user.OnboardingCompleted,
		user.WeeklyReviewWeekday,
		user.WeeklyReviewHour,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *Repository) GetByID(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) (User, error) {
	query := `
		select id, email, display_name, avatar_url, timezone, base_currency,
		       onboarding_completed, weekly_review_weekday, weekly_review_hour, created_at, updated_at
		from users
		where id = $1
	`

	var user User
	err := r.base(exec...).QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Timezone,
		&user.BaseCurrency,
		&user.OnboardingCompleted,
		&user.WeeklyReviewWeekday,
		&user.WeeklyReviewHour,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}

func (r *Repository) GetByEmail(ctx context.Context, email string, exec ...db.DBTX) (User, error) {
	query := `
		select id, email, display_name, avatar_url, timezone, base_currency,
		       onboarding_completed, weekly_review_weekday, weekly_review_hour, created_at, updated_at
		from users
		where lower(email) = lower($1)
		limit 1
	`

	var user User
	err := r.base(exec...).QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Timezone,
		&user.BaseCurrency,
		&user.OnboardingCompleted,
		&user.WeeklyReviewWeekday,
		&user.WeeklyReviewHour,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}

func (r *Repository) UpdatePreferences(ctx context.Context, userID uuid.UUID, in UpdatePreferencesRequest, exec ...db.DBTX) (User, error) {
	current, err := r.GetByID(ctx, userID, exec...)
	if err != nil {
		return User{}, err
	}

	if in.Timezone != nil {
		current.Timezone = *in.Timezone
	}
	if in.BaseCurrency != nil {
		current.BaseCurrency = *in.BaseCurrency
	}
	if in.OnboardingCompleted != nil {
		current.OnboardingCompleted = *in.OnboardingCompleted
	}
	if in.WeeklyReviewWeekday != nil {
		current.WeeklyReviewWeekday = *in.WeeklyReviewWeekday
	}
	if in.WeeklyReviewHour != nil {
		current.WeeklyReviewHour = *in.WeeklyReviewHour
	}
	current.UpdatedAt = time.Now().UTC()

	query := `
		update users
		set timezone = $2,
		    base_currency = $3,
		    onboarding_completed = $4,
		    weekly_review_weekday = $5,
		    weekly_review_hour = $6,
		    updated_at = $7
		where id = $1
	`

	_, err = r.base(exec...).ExecContext(ctx, query,
		userID,
		current.Timezone,
		current.BaseCurrency,
		current.OnboardingCompleted,
		current.WeeklyReviewWeekday,
		current.WeeklyReviewHour,
		current.UpdatedAt,
	)
	if err != nil {
		return User{}, err
	}

	return current, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
