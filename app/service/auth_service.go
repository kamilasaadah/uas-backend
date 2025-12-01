package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"uas-backend/app/model"
	"uas-backend/app/repository"
	"uas-backend/config"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
	Login(ctx context.Context, req model.LoginRequest) (string, string, *model.AuthUserResponse, error)
}

type authService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Login(ctx context.Context, req model.LoginRequest) (string, string, *model.AuthUserResponse, error) {

	user, err := s.repo.FindUserByUsernameOrEmail(ctx, req.Username)
	if err != nil {
		return "", "", nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return "", "", nil, errors.New("account is not active")
	}

	// check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", "", nil, errors.New("invalid credentials")
	}

	// load permissions
	perms, err := s.repo.GetPermissionsByRole(ctx, user.RoleID)
	if err != nil {
		return "", "", nil, errors.New("failed to load permissions")
	}

	// JWT payload
	claims := jwt.MapClaims{
		"user_id":     user.ID,
		"username":    user.Username,
		"role":        user.RoleName,
		"role_id":     user.RoleID,
		"permissions": perms,
		"exp":         time.Now().Add(time.Hour * 2).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := []byte(config.JWTSecret())

	accessToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", "", nil, errors.New("failed to sign token")
	}

	// refresh token (optional)
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refresh.SignedString(jwtSecret)

	if err != nil {
		return "", "", nil, errors.New("failed to sign refresh token")
	}

	resp := &model.AuthUserResponse{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Role:        user.RoleName,
		Permissions: perms,
	}

	return accessToken, refreshToken, resp, nil
}
