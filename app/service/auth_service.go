package service

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"uas-backend/app/model"
	"uas-backend/app/repository"
	"uas-backend/config"
)

type AuthHttpHandler interface {
	Login(c *fiber.Ctx) error
	Profile(c *fiber.Ctx) error
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthHttpHandler {
	return &authService{
		userRepo: userRepo,
	}
}

///////////////////////////////////////////////////////////////////////////////
// LOGIN HANDLER (BISNIS LOGIC + ERROR CODE DI DALAM SINI)
///////////////////////////////////////////////////////////////////////////////

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

func (s *authService) Profile(c *fiber.Ctx) error {

	claims := c.Locals("user").(jwt.MapClaims)

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
