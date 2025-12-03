package model

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	FullName     string `json:"full_name"`
	RoleID       string `json:"role_id"`
	RoleName     string `json:"role_name"`
	IsActive     bool   `json:"is_active"`
}

// ======================= LOGIN REQUEST =======================

type LoginRequest struct {
	Username string `json:"username" validate:"required"` // boleh username/email
	Password string `json:"password" validate:"required"`
}

// ======================= LOGIN RESPONSE =======================

type AuthUserResponse struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	FullName    string   `json:"full_name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

// ======================= USER REQUEST =======================
type CreateUserRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required"`
	Password string `json:"password" validate:"required"`
	RoleID   string `json:"role_id" validate:"required"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`

	StudentProfile  *StudentProfileRequest  `json:"student_profile"`
	LecturerProfile *LecturerProfileRequest `json:"lecturer_profile"`
}

type UpdateUserRoleRequest struct {
	RoleID string `json:"role_id" validate:"required"`
}
