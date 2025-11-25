package model

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
