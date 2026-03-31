package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into users (
			id, email, display_name, avatar_url, timezone, base_currency,
			onboarding_completed, weekly_review_weekday, weekly_review_hour, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, user.ID, user.Email, user.DisplayName, user.AvatarURL, user.Timezone, user.BaseCurrency,
		user.OnboardingCompleted, user.WeeklyReviewWeekday, user.WeeklyReviewHour, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *Repository) GetByID(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) (User, error) {
	var user User
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, email::text, display_name, avatar_url, timezone, base_currency,
		       onboarding_completed, weekly_review_weekday, weekly_review_hour, created_at, updated_at
		from users
		where id = $1
	`, userID).Scan(
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
	var user User
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, email::text, display_name, avatar_url, timezone, base_currency,
		       onboarding_completed, weekly_review_weekday, weekly_review_hour, created_at, updated_at
		from users
		where lower(email::text) = lower($1)
		limit 1
	`, email).Scan(
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

func (r *Repository) GetProfileBase(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) (Profile, error) {
	var profile Profile
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, email::text, display_name, avatar_url, created_at, updated_at
		from users
		where id = $1
	`, userID).Scan(
		&profile.ID,
		&profile.Email,
		&profile.DisplayName,
		&profile.AvatarURL,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		return Profile{}, err
	}

	profile.ProfileRoles = []ProfileRole{}
	profile.Teams = []DevelopmentTeam{}
	return profile, nil
}

func (r *Repository) UpdateProfileFields(ctx context.Context, profile Profile, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update users
		set display_name = $2,
		    avatar_url = $3,
		    updated_at = $4
		where id = $1
	`, profile.ID, profile.DisplayName, profile.AvatarURL, profile.UpdatedAt)
	return err
}

func (r *Repository) ListProfileRoles(ctx context.Context, exec ...db.DBTX) ([]ProfileRole, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, code, name, description, sort_order
		from profile_roles
		order by sort_order asc, name asc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ProfileRole, 0, 12)
	for rows.Next() {
		var item ProfileRole
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Description, &item.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) ListUserProfileRoles(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]ProfileRole, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select pr.id, pr.code, pr.name, pr.description, pr.sort_order
		from user_profile_roles upr
		join profile_roles pr on pr.id = upr.profile_role_id
		where upr.user_id = $1
		order by pr.sort_order asc, pr.name asc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ProfileRole, 0, 4)
	for rows.Next() {
		var item ProfileRole
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Description, &item.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) ReplaceUserProfileRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, createdAt time.Time, exec ...db.DBTX) error {
	if _, err := r.base(exec...).ExecContext(ctx, `
		delete from user_profile_roles
		where user_id = $1
	`, userID); err != nil {
		return err
	}

	for _, roleID := range roleIDs {
		if _, err := r.base(exec...).ExecContext(ctx, `
			insert into user_profile_roles (user_id, profile_role_id, created_at)
			values ($1, $2, $3)
		`, userID, roleID, createdAt); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) ListExistingUserIDs(ctx context.Context, userIDs []uuid.UUID, exec ...db.DBTX) (map[uuid.UUID]struct{}, error) {
	result := make(map[uuid.UUID]struct{}, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	args := make([]any, 0, len(userIDs))
	for _, id := range userIDs {
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		select id
		from users
		where id in (%s)
	`, placeholders(1, len(args)))

	rows, err := r.base(exec...).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = struct{}{}
	}

	return result, rows.Err()
}

func (r *Repository) CreateDevelopmentTeam(ctx context.Context, team DevelopmentTeam, memberUserIDs []uuid.UUID, exec ...db.DBTX) error {
	if _, err := r.base(exec...).ExecContext(ctx, `
		insert into org_groups (
			id, name, code, department_id, is_active, created_at, updated_at,
			description, group_type, lead_user_id, created_by_user_id
		)
		values ($1, $2, null, null, true, $3, $4, $5, 'development_team', $6, $7)
	`, team.ID, team.Name, team.CreatedAt, team.UpdatedAt, team.Description, team.LeadUserID, team.CreatedByUserID); err != nil {
		return err
	}

	for _, memberUserID := range memberUserIDs {
		if _, err := r.base(exec...).ExecContext(ctx, `
			insert into org_group_members (group_id, user_id, created_at)
			values ($1, $2, $3)
		`, team.ID, memberUserID, team.CreatedAt); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) ListDevelopmentTeamsByUser(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]DevelopmentTeam, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select g.id, g.name, g.description, g.lead_user_id, g.created_by_user_id, g.created_at, g.updated_at
		from org_groups g
		join org_group_members gm on gm.group_id = g.id
		where gm.user_id = $1
		  and g.group_type = 'development_team'
		  and g.is_active = true
		order by g.created_at desc, g.name asc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]DevelopmentTeam, 0, 4)
	teamIndex := make(map[uuid.UUID]int)
	for rows.Next() {
		var team DevelopmentTeam
		if err := rows.Scan(
			&team.ID,
			&team.Name,
			&team.Description,
			&team.LeadUserID,
			&team.CreatedByUserID,
			&team.CreatedAt,
			&team.UpdatedAt,
		); err != nil {
			return nil, err
		}
		team.Members = []DevelopmentTeamMember{}
		teamIndex[team.ID] = len(teams)
		teams = append(teams, team)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return teams, nil
	}

	memberRows, err := r.base(exec...).QueryContext(ctx, `
		select g.id,
		       u.id,
		       coalesce(
		         nullif(trim(u.display_name), ''),
		         nullif(trim(concat_ws(' ', ep.first_name, ep.last_name)), ''),
		         u.email::text
		       ) as display_name,
		       u.email::text,
		       u.avatar_url,
		       case when g.lead_user_id is not null and g.lead_user_id = u.id then true else false end as is_lead
		from org_groups g
		join org_group_members self_member on self_member.group_id = g.id and self_member.user_id = $1
		join org_group_members gm on gm.group_id = g.id
		join users u on u.id = gm.user_id
		left join employee_profiles ep on ep.user_id = u.id
		where g.group_type = 'development_team'
		  and g.is_active = true
		order by g.created_at desc, is_lead desc, display_name asc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer memberRows.Close()

	for memberRows.Next() {
		var teamID uuid.UUID
		var member DevelopmentTeamMember
		if err := memberRows.Scan(
			&teamID,
			&member.UserID,
			&member.DisplayName,
			&member.Email,
			&member.AvatarURL,
			&member.IsLead,
		); err != nil {
			return nil, err
		}

		index, ok := teamIndex[teamID]
		if !ok {
			continue
		}
		teams[index].Members = append(teams[index].Members, member)
	}

	return teams, memberRows.Err()
}

func (r *Repository) GetDevelopmentTeamByID(ctx context.Context, teamID uuid.UUID, exec ...db.DBTX) (DevelopmentTeam, error) {
	var team DevelopmentTeam
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, name, description, lead_user_id, created_by_user_id, created_at, updated_at
		from org_groups
		where id = $1
		  and group_type = 'development_team'
		  and is_active = true
	`, teamID).Scan(
		&team.ID,
		&team.Name,
		&team.Description,
		&team.LeadUserID,
		&team.CreatedByUserID,
		&team.CreatedAt,
		&team.UpdatedAt,
	)
	if err != nil {
		return DevelopmentTeam{}, err
	}

	memberRows, err := r.base(exec...).QueryContext(ctx, `
		select u.id,
		       coalesce(
		         nullif(trim(u.display_name), ''),
		         nullif(trim(concat_ws(' ', ep.first_name, ep.last_name)), ''),
		         u.email::text
		       ) as display_name,
		       u.email::text,
		       u.avatar_url,
		       case when g.lead_user_id is not null and g.lead_user_id = u.id then true else false end as is_lead
		from org_groups g
		join org_group_members gm on gm.group_id = g.id
		join users u on u.id = gm.user_id
		left join employee_profiles ep on ep.user_id = u.id
		where g.id = $1
		order by is_lead desc, display_name asc
	`, teamID)
	if err != nil {
		return DevelopmentTeam{}, err
	}
	defer memberRows.Close()

	team.Members = make([]DevelopmentTeamMember, 0, 4)
	for memberRows.Next() {
		var member DevelopmentTeamMember
		if err := memberRows.Scan(
			&member.UserID,
			&member.DisplayName,
			&member.Email,
			&member.AvatarURL,
			&member.IsLead,
		); err != nil {
			return DevelopmentTeam{}, err
		}
		team.Members = append(team.Members, member)
	}

	return team, memberRows.Err()
}

func (r *Repository) FindCurrentDevelopmentTeamIDByUser(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) (*uuid.UUID, error) {
	var teamID uuid.UUID
	err := r.base(exec...).QueryRowContext(ctx, `
		select g.id
		from org_groups g
		join org_group_members gm on gm.group_id = g.id
		where gm.user_id = $1
		  and g.group_type = 'development_team'
		  and g.is_active = true
		order by g.created_at desc, g.id asc
		limit 1
	`, userID).Scan(&teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &teamID, nil
}

func (r *Repository) ListAvailableDevelopmentTeams(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]DevelopmentTeam, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select g.id, g.name, g.description, g.lead_user_id, g.created_by_user_id, g.created_at, g.updated_at
		from org_groups g
		where g.group_type = 'development_team'
		  and g.is_active = true
		  and not exists (
		    select 1
		    from org_group_members gm
		    where gm.group_id = g.id and gm.user_id = $1
		  )
		order by g.name asc, g.created_at desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]DevelopmentTeam, 0, 8)
	for rows.Next() {
		var team DevelopmentTeam
		if err := rows.Scan(
			&team.ID,
			&team.Name,
			&team.Description,
			&team.LeadUserID,
			&team.CreatedByUserID,
			&team.CreatedAt,
			&team.UpdatedAt,
		); err != nil {
			return nil, err
		}
		team.Members = []DevelopmentTeamMember{}
		teams = append(teams, team)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return teams, nil
	}

	for index, team := range teams {
		fullTeam, err := r.GetDevelopmentTeamByID(ctx, team.ID, exec...)
		if err != nil {
			return nil, err
		}
		teams[index] = fullTeam
	}

	return teams, nil
}

func (r *Repository) IsDevelopmentTeamMember(ctx context.Context, teamID, userID uuid.UUID, exec ...db.DBTX) (bool, error) {
	var exists bool
	err := r.base(exec...).QueryRowContext(ctx, `
		select exists(
			select 1
			from org_group_members gm
			join org_groups g on g.id = gm.group_id
			where gm.group_id = $1
			  and gm.user_id = $2
			  and g.group_type = 'development_team'
			  and g.is_active = true
		)
	`, teamID, userID).Scan(&exists)
	return exists, err
}

func (r *Repository) AddDevelopmentTeamMember(ctx context.Context, teamID, userID uuid.UUID, createdAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into org_group_members (group_id, user_id, created_at)
		values ($1, $2, $3)
		on conflict (group_id, user_id) do nothing
	`, teamID, userID, createdAt)
	return err
}

func (r *Repository) RemoveDevelopmentTeamMember(ctx context.Context, teamID, userID uuid.UUID, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		delete from org_group_members
		where group_id = $1 and user_id = $2
	`, teamID, userID)
	return err
}

func (r *Repository) ListDevelopmentTeamMemberUserIDs(ctx context.Context, teamID uuid.UUID, exec ...db.DBTX) ([]uuid.UUID, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select user_id
		from org_group_members
		where group_id = $1
		order by created_at asc, user_id asc
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDs := make([]uuid.UUID, 0, 4)
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}

func (r *Repository) UpdateDevelopmentTeamLead(ctx context.Context, teamID uuid.UUID, leadUserID *uuid.UUID, updatedAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update org_groups
		set lead_user_id = $2,
		    updated_at = $3
		where id = $1
		  and group_type = 'development_team'
	`, teamID, leadUserID, updatedAt)
	return err
}

func (r *Repository) DeactivateDevelopmentTeam(ctx context.Context, teamID uuid.UUID, updatedAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update org_groups
		set is_active = false,
		    lead_user_id = null,
		    updated_at = $2
		where id = $1
		  and group_type = 'development_team'
	`, teamID, updatedAt)
	return err
}

func (r *Repository) TouchDevelopmentTeam(ctx context.Context, teamID uuid.UUID, updatedAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update org_groups
		set updated_at = $2
		where id = $1
		  and group_type = 'development_team'
		  and is_active = true
	`, teamID, updatedAt)
	return err
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func placeholders(start, count int) string {
	parts := make([]string, 0, count)
	for i := 0; i < count; i++ {
		parts = append(parts, fmt.Sprintf("$%d", start+i))
	}

	return strings.Join(parts, ", ")
}
