package route

import (
	"uas-backend/app/service"
	"uas-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func StudentRoutes(r fiber.Router, studentSvc *service.StudentService) {

	admin := r.Group("/students",
		middleware.JWTAuth(),
		middleware.RequirePermission("user:manage"),
	)

	// PUT /api/v1/students/:id/advisor  (Admin only)
	admin.Get("/", studentSvc.GetAllStudents)
	admin.Get("/:id", studentSvc.GetStudentByID)
	admin.Put("/:id/advisor", studentSvc.SetAdvisor)

}
