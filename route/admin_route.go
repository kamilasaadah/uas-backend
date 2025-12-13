package route

import (
	"uas-backend/app/repository"
	"uas-backend/app/service"
	"uas-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func AdminRoutes(
	r fiber.Router, 
	userService service.UserHttpHandler,
	userRepo repository.UserRepository,
	) {

	admin := r.Group("/users",
		middleware.JWTAuth(userRepo),
		middleware.RequirePermission("user:manage"),
	)

	admin.Post("/", userService.Create)
	admin.Put("/:id", userService.Update)
	admin.Put("/:id/role", userService.AssignRole)

	admin.Get("/", userService.GetAll)
	admin.Get("/:id", userService.GetByID)

	admin.Delete("/:id", userService.Delete)
}
