package model

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	StudentID   string   `json:"student_id,omitempty"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`

	jwt.RegisteredClaims
}
