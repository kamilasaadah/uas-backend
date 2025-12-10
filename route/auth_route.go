package route

import (
	"uas-backend/app/service"
	"uas-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(r fiber.Router, authService service.AuthHttpHandler) {

	// PUBLIC
	r.Post("/login", authService.Login)

	// PROTECTED
	protected := r.Group("/", middleware.JWTAuth())
	protected.Get("/profile", authService.Profile)
}
