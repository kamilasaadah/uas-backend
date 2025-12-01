package route

import "github.com/gofiber/fiber/v2"

func SetupRoutes(app *fiber.App) {

	api := app.Group("/api/v1")

	// Auth
	AuthRoutes(api.Group("/auth"))

}
