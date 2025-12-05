package route

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"uas-backend/app/model"
	"uas-backend/app/repository"
	"uas-backend/app/service"
	"uas-backend/database"
	"uas-backend/middleware"
)

func AuthRoutes(r fiber.Router) {

	userRepo := repository.NewUserRepository(database.PG)
	authService := service.NewAuthService(userRepo)

	// ===========================
	// PUBLIC ROUTES (NO JWT)
	// ===========================
	r.Post("/login", func(c *fiber.Ctx) error {

		var req model.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    400,
				"message": "Bad Request",
				"error":   "Invalid input data",
			})
		}

		access, refresh, user, err := authService.Login(context.Background(), req)

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    401,
				"message": "Unauthorized",
				"error":   err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"code":    200,
			"message": "Login successful",
			"data": fiber.Map{
				"token":        access,
				"refreshToken": refresh,
				"user":         user,
			},
		})
	})

	// ===========================
	// PROTECTED ROUTES (JWT REQUIRED)
	// ===========================

	protected := r.Group("/", middleware.JWTAuth())

	// FR-003 (nanti) — profile endpoint
	protected.Get("/profile", func(c *fiber.Ctx) error {

		claims := c.Locals("user").(jwt.MapClaims)

		return c.JSON(fiber.Map{
			"code":    200,
			"message": "Profile fetched",
			"data": fiber.Map{
				"id":          claims["user_id"],
				"username":    claims["username"],
				"full_name":   claims["full_name"],
				"role":        claims["role"],
				"role_id":     claims["role_id"],
				"permissions": claims["permissions"],
			},
		})
	})

}
