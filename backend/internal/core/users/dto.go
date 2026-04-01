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

// EmployeePublicProfile is a public view of an employee profile.
type EmployeePublicProfile struct {
	UserID         uuid.UUID         `json:"user_id"`
	Email          string            `json:"email"`
	DisplayName    *string           `json:"display_name,omitempty"`
	AvatarURL      *string           `json:"avatar_url,omitempty"`
	FirstName      string            `json:"first_name"`
	LastName       string            `json:"last_name"`
	MiddleName     *string           `json:"middle_name,omitempty"`
	PositionTitle  *string           `json:"position_title,omitempty"`
	DepartmentName *string           `json:"department_name,omitempty"`
	HireDate       *string           `json:"hire_date,omitempty"`
	ProfileRoles   []ProfileRole     `json:"profile_roles"`
	Teams          []DevelopmentTeam `json:"teams"`
}

// EmployeeEnrollmentItem is a single enrollment row with embedded course info.
type EmployeeEnrollmentItem struct {
	ID                uuid.UUID `json:"id"`
	CourseID          uuid.UUID `json:"course_id"`
	CourseTitle       string    `json:"course_title"`
	CourseProvider    *string   `json:"course_provider,omitempty"`
	CourseLevel       *string   `json:"course_level,omitempty"`
	Status            string    `json:"status"`
	CompletionPercent string    `json:"completion_percent"`
	IsMandatory       bool      `json:"is_mandatory"`
	EnrolledAt        string    `json:"enrolled_at"`
	StartedAt         *string   `json:"started_at,omitempty"`
	CompletedAt       *string   `json:"completed_at,omitempty"`
	DeadlineAt        *string   `json:"deadline_at,omitempty"`
}

// EmployeeProfileResponse is the response for GET /employees/:userId.
type EmployeeProfileResponse struct {
	Profile     EmployeePublicProfile    `json:"profile"`
	Enrollments []EmployeeEnrollmentItem `json:"enrollments"`
}
