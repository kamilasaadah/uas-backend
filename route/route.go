package route

import (
	"github.com/gofiber/fiber/v2"

	"uas-backend/app/repository"
	"uas-backend/app/service"
	"uas-backend/database"
)

func SetupRoutes(app *fiber.App) {

	api := app.Group("/api/v1")

	// === INIT REPO ===
	userRepo := repository.NewUserRepository(database.PG)

	// === INIT SERVICE ===
	authService := service.NewAuthService(userRepo)

	// ROUTES
	AuthRoutes(api.Group("/auth"), authService)
	// AuthRoutes(api.Group("/auth"))
	AdminRoutes(api.Group("/"))

}
