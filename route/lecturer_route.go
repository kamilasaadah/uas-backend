package route

import (
	"uas-backend/app/repository"
	"uas-backend/app/service"
	"uas-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func LecturerRoutes(
	r fiber.Router,
	lecturerSvc *service.LecturerService,
	userRepo repository.UserRepository,
) {

	api := r.Group("/lecturers", middleware.JWTAuth(userRepo))
	api.Get("/:id/advisees", middleware.RequireAnyPermission(
		"achievement:read", "achievement:verify", "user:manage",
	), lecturerSvc.GetAdvisees)

	admin := api.Group("/",
		middleware.RequirePermission("user:manage"),
	)
	admin.Get("/", lecturerSvc.GetAllLecturers)
}
