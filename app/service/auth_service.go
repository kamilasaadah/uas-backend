package service

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"uas-backend/app/model"
	"uas-backend/app/repository"
	"uas-backend/config"
	"uas-backend/middleware"
)

type AuthHttpHandler interface {
	Login(c *fiber.Ctx) error
	Profile(c *fiber.Ctx) error
	Refresh(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
}

type authService struct {
	userRepo    repository.UserRepository
	studentRepo repository.StudentRepository
}

func NewAuthService(
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
) AuthHttpHandler {
	return &authService{
		userRepo:    userRepo,
		studentRepo: studentRepo,
	}
}

///////////////////////////////////////////////////////////////////////////////
// LOGIN HANDLER (BISNIS LOGIC + ERROR CODE DI DALAM SINI)
///////////////////////////////////////////////////////////////////////////////

// Login godoc
// @Summary Login user
// @Description Login menggunakan username/email dan password
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "Login payload"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 403 {object} map[string]interface{} "Account not active"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/login [post]
func (s *authService) Login(c *fiber.Ctx) error {

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return s.error(c, 400, "Invalid input data")
	}

	// 1. FIND USER
	user, err := s.userRepo.FindByUsernameOrEmail(context.Background(), req.Username)
	if err != nil {
		return s.error(c, 401, "invalid credentials")
	}

	// 2. CHECK ACTIVE
	if !user.IsActive {
		return s.error(c, 403, "account is not active")
	}

	// 3. VERIFY PASSWORD
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return s.error(c, 401, "invalid credentials")
	}

	// 4. LOAD PERMISSIONS
	perms, err := s.userRepo.GetUserPermissions(user.ID)
	if err != nil {
		return s.error(c, 500, "failed to load permissions")
	}

	// 5. ACCESS TOKEN
	claims := jwt.MapClaims{
		"user_id":     user.ID,
		"username":    user.Username,
		"full_name":   user.FullName,
		"role":        user.RoleName,
		"role_id":     user.RoleID,
		"permissions": perms,
		"exp":         time.Now().Add(2 * time.Hour).Unix(),
	}

	if user.RoleName == "Mahasiswa" {
		student, err := s.studentRepo.GetStudentProfile(
			context.Background(),
			user.ID,
		)
		if err != nil {
			return s.error(c, 403, "student profile not found")
		}
		claims["student_id"] = student.ID
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(config.JWTSecret()))
	if err != nil {
		return s.error(c, 500, "failed to sign token")
	}

	// 6. REFRESH TOKEN
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
		SignedString([]byte(config.JWTSecret()))
	if err != nil {
		return s.error(c, 500, "failed to sign refresh token")
	}

	// 7. RESPONSE
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Login successful",
		"data": fiber.Map{
			"token":        accessToken,
			"refreshToken": refreshToken,
			"user": fiber.Map{
				"id":          user.ID,
				"username":    user.Username,
				"full_name":   user.FullName,
				"role":        user.RoleName,
				"permissions": perms,
			},
		},
	})
}

///////////////////////////////////////////////////////////////////////////////
// PROFILE HANDLER
///////////////////////////////////////////////////////////////////////////////

// Profile godoc
// @Summary Get current user profile
// @Description Mengambil profile user dari JWT
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Profile fetched"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /auth/profile [get]
func (s *authService) Profile(c *fiber.Ctx) error {

	claims := c.Locals("user").(*model.JWTClaims)

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Profile fetched",
		"data":    claims,
	})
}

///////////////////////////////////////////////////////////////////////////////
// ERROR HELPER (DIPANGGIL DI SERVICE)
///////////////////////////////////////////////////////////////////////////////

func (s *authService) error(c *fiber.Ctx, code int, msg string) error {
	return c.Status(code).JSON(fiber.Map{
		"code":    code,
		"message": msg,
		"error":   msg,
	})
}

// Refresh godoc
// @Summary Refresh access token
// @Description Generate access token baru menggunakan refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.RefreshTokenRequest true "Refresh token payload"
// @Success 200 {object} map[string]interface{} "Token refreshed"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Invalid refresh token"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/refresh [post]
func (s *authService) Refresh(c *fiber.Ctx) error {

	var req model.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return s.error(c, 400, "invalid input")
	}

	// 1️⃣ parse refresh token
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		req.RefreshToken,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(config.JWTSecret()), nil
		},
	)

	if err != nil || !token.Valid {
		return s.error(c, 401, "invalid refresh token")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return s.error(c, 401, "invalid refresh token")
	}

	// 2️⃣ ambil user
	user, err := s.userRepo.GetUserByID(
		c.Context(),
		userID,
	)
	if err != nil || !user.IsActive {
		return s.error(c, 401, "user not found")
	}

	// 3️⃣ permissions
	perms, err := s.userRepo.GetUserPermissions(user.ID)
	if err != nil {
		return s.error(c, 500, "failed to load permissions")
	}

	// 4️⃣ new access token
	newClaims := jwt.MapClaims{
		"user_id":     user.ID,
		"username":    user.Username,
		"full_name":   user.FullName,
		"role":        user.RoleName,
		"role_id":     user.RoleID,
		"permissions": perms,
		"exp":         time.Now().Add(2 * time.Hour).Unix(),
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims).
		SignedString([]byte(config.JWTSecret()))
	if err != nil {
		return s.error(c, 500, "failed to sign token")
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Token refreshed",
		"data": fiber.Map{
			"token": accessToken,
		},
	})
}

// Logout godoc
// @Summary Logout user
// @Description Logout user dan memblokir JWT sampai expired
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 400 {object} map[string]interface{} "Invalid token"
// @Router /auth/logout [post]
func (s *authService) Logout(c *fiber.Ctx) error {

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return s.error(c, 401, "missing token")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(config.JWTSecret()), nil
		},
	)
	if err != nil || !token.Valid {
		return s.error(c, 401, "invalid token")
	}

	expUnix, ok := claims["exp"].(float64)
	if !ok {
		return s.error(c, 400, "invalid token payload")
	}

	expiredAt := time.Unix(int64(expUnix), 0)

	// 🔥 BLOCK TOKEN DI MEMORY
	middleware.BlockJWT(tokenString, expiredAt)

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Logout successful",
	})
}
