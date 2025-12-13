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

	api := r.Group("/lecturers",
		middleware.JWTAuth(userRepo),
	)

	read := api.Group("/",
		middleware.RequireAnyPermission(
			"student:read", // dosen
			"user:manage",  // admin
		),
	)
	read.Get("/:id/advisees", lecturerSvc.GetAdvisees)

	admin := api.Group("/",
		middleware.RequirePermission("user:manage"),
	)
	admin.Get("/", lecturerSvc.GetAllLecturers)
}
