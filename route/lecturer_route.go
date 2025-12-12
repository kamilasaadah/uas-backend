package route

import (
	"uas-backend/app/service"
	"uas-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func LecturerRoutes(r fiber.Router, lecturerSvc *service.LecturerService) {

	admin := r.Group("/lecturers",
		middleware.JWTAuth(),
		middleware.RequirePermission("user:manage"), // admin only
	)

	admin.Get("/", lecturerSvc.GetAllLecturers)
	admin.Get("/:id/advisees", lecturerSvc.GetAdvisees)
}
