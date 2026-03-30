package org

import (
	"context"
	"database/sql"
	"time"

	"moneyapp/backend/internal/platform/db"

	"github.com/google/uuid"
)

type Department struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Code       *string    `json:"code,omitempty"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	HeadUserID *uuid.UUID `json:"head_user_id,omitempty"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type EmployeeProfile struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	EmployeeNo       *string    `json:"employee_no,omitempty"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	MiddleName       *string    `json:"middle_name,omitempty"`
	PositionTitle    *string    `json:"position_title,omitempty"`
	DepartmentID     *uuid.UUID `json:"department_id,omitempty"`
	HireDate         *time.Time `json:"hire_date,omitempty"`
	EmploymentStatus string     `json:"employment_status"`
	Timezone         *string    `json:"timezone,omitempty"`
	OutlookEmail     *string    `json:"outlook_email,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CreateProfileInput struct {
	UserID        uuid.UUID
	FirstName     string
	LastName      string
	MiddleName    *string
	PositionTitle *string
	DepartmentID  *uuid.UUID
	Timezone      *string
	OutlookEmail  *string
}

type UpdateProfileInput struct {
	FirstName     *string
	LastName      *string
	MiddleName    *string
	PositionTitle *string
	DepartmentID  *uuid.UUID
	Timezone      *string
	OutlookEmail  *string
}

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

func (r *Repository) CreateProfile(ctx context.Context, profile EmployeeProfile, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into employee_profiles (
			id, user_id, employee_no, first_name, last_name, middle_name, position_title,
			department_id, hire_date, employment_status, timezone, outlook_email, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, profile.ID, profile.UserID, profile.EmployeeNo, profile.FirstName, profile.LastName, profile.MiddleName,
		profile.PositionTitle, profile.DepartmentID, profile.HireDate, profile.EmploymentStatus,
		profile.Timezone, profile.OutlookEmail, profile.CreatedAt, profile.UpdatedAt)
	return err
}

func (r *Repository) UpdateProfile(ctx context.Context, profile EmployeeProfile, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update employee_profiles
		set first_name = $2,
		    last_name = $3,
		    middle_name = $4,
		    position_title = $5,
		    department_id = $6,
		    timezone = $7,
		    outlook_email = $8,
		    updated_at = $9
		where user_id = $1
	`, profile.UserID, profile.FirstName, profile.LastName, profile.MiddleName, profile.PositionTitle,
		profile.DepartmentID, profile.Timezone, profile.OutlookEmail, profile.UpdatedAt)
	return err
}

func (r *Repository) GetProfileByUserID(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) (EmployeeProfile, error) {
	var item EmployeeProfile
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, user_id, employee_no, first_name, last_name, middle_name, position_title,
		       department_id, hire_date, employment_status, timezone, outlook_email, created_at, updated_at
		from employee_profiles
		where user_id = $1
	`, userID).Scan(
		&item.ID, &item.UserID, &item.EmployeeNo, &item.FirstName, &item.LastName, &item.MiddleName,
		&item.PositionTitle, &item.DepartmentID, &item.HireDate, &item.EmploymentStatus, &item.Timezone,
		&item.OutlookEmail, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

func (r *Repository) ListUserIDsByDepartment(ctx context.Context, departmentID uuid.UUID, exec ...db.DBTX) ([]uuid.UUID, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select user_id
		from employee_profiles
		where department_id = $1 and employment_status = 'active'
	`, departmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		result = append(result, userID)
	}

	return result, rows.Err()
}

func (r *Repository) ListUserIDsByGroup(ctx context.Context, groupID uuid.UUID, exec ...db.DBTX) ([]uuid.UUID, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select user_id
		from org_group_members
		where group_id = $1
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		result = append(result, userID)
	}

	return result, rows.Err()
}

func (r *Repository) GetPrimaryManager(ctx context.Context, employeeUserID uuid.UUID, exec ...db.DBTX) (*uuid.UUID, error) {
	var managerID uuid.UUID
	err := r.base(exec...).QueryRowContext(ctx, `
		select manager_user_id
		from manager_relations
		where employee_user_id = $1
		order by is_primary desc, created_at asc
		limit 1
	`, employeeUserID).Scan(&managerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &managerID, nil
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateDefaultProfile(ctx context.Context, input CreateProfileInput, exec ...db.DBTX) (EmployeeProfile, error) {
	now := time.Now().UTC()
	profile := EmployeeProfile{
		ID:               uuid.New(),
		UserID:           input.UserID,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		MiddleName:       input.MiddleName,
		PositionTitle:    input.PositionTitle,
		DepartmentID:     input.DepartmentID,
		EmploymentStatus: "active",
		Timezone:         input.Timezone,
		OutlookEmail:     input.OutlookEmail,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	return profile, s.repo.CreateProfile(ctx, profile, exec...)
}

func (s *Service) GetProfileByUserID(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) (EmployeeProfile, error) {
	return s.repo.GetProfileByUserID(ctx, userID, exec...)
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput, exec ...db.DBTX) (EmployeeProfile, error) {
	profile, err := s.repo.GetProfileByUserID(ctx, userID, exec...)
	if err != nil {
		return EmployeeProfile{}, err
	}

	if input.FirstName != nil {
		profile.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		profile.LastName = *input.LastName
	}
	if input.MiddleName != nil {
		profile.MiddleName = input.MiddleName
	}
	if input.PositionTitle != nil {
		profile.PositionTitle = input.PositionTitle
	}
	if input.DepartmentID != nil {
		profile.DepartmentID = input.DepartmentID
	}
	if input.Timezone != nil {
		profile.Timezone = input.Timezone
	}
	if input.OutlookEmail != nil {
		profile.OutlookEmail = input.OutlookEmail
	}
	profile.UpdatedAt = time.Now().UTC()

	return profile, s.repo.UpdateProfile(ctx, profile, exec...)
}

func (s *Service) ResolveTargetUsers(ctx context.Context, targetType string, targetID uuid.UUID) ([]uuid.UUID, error) {
	switch targetType {
	case "user":
		return []uuid.UUID{targetID}, nil
	case "department":
		return s.repo.ListUserIDsByDepartment(ctx, targetID)
	case "group":
		return s.repo.ListUserIDsByGroup(ctx, targetID)
	default:
		return nil, sql.ErrNoRows
	}
}

func (s *Service) GetPrimaryManager(ctx context.Context, employeeUserID uuid.UUID, exec ...db.DBTX) (*uuid.UUID, error) {
	return s.repo.GetPrimaryManager(ctx, employeeUserID, exec...)
}
