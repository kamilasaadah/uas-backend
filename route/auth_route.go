package route

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"uas-backend/app/model"
	"uas-backend/app/repository"
	"uas-backend/app/service"
)

func AuthRoutes(r fiber.Router) {

	authRepo := repository.NewAuthRepository()
	authService := service.NewAuthService(authRepo)

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
}
