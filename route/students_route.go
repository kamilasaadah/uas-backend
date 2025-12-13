package route

import (
	"uas-backend/app/repository"
	"uas-backend/app/service"
	"uas-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func StudentRoutes(
	r fiber.Router,
	studentSvc *service.StudentService,
	userRepo repository.UserRepository,
) {

	api := r.Group("/students", middleware.JWTAuth(userRepo))

	// Admin OR Dosen wali read access
	read := api.Group("/",
		middleware.RequireAnyPermission(
			"student:read",
			"user:manage",
		),
	)
	read.Get("/", studentSvc.GetAllStudents)
	read.Get("/:id", studentSvc.GetStudentByID)

	// Admin full access
	admin := api.Group("/",
		middleware.RequirePermission("user:manage"),
	)
	admin.Put("/:id/advisor", studentSvc.SetAdvisor)

}
