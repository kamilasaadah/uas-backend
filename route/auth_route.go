package route

import (
	"uas-backend/app/repository"
	"uas-backend/app/service"
	"uas-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(
	r fiber.Router,
	authService service.AuthHttpHandler,
	userRepo repository.UserRepository) {

	// PUBLIC
	r.Post("/login", authService.Login)
	r.Post("/refresh", authService.Refresh)

	// PROTECTED
	protected := r.Group("/", middleware.JWTAuth(userRepo))
	protected.Get("/profile", authService.Profile)
	protected.Post("/logout", authService.Logout)
}
