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

	// ======================
	// ADMIN + DOSEN
	// HARUS DI ATAS
	// ======================
	api.Get("/:id/achievements", studentSvc.GetStudentAchievements)

	// ======================
	// ADMIN ONLY
	// ======================
	admin := api.Group("",
		middleware.RequirePermission("user:manage"),
	)
	admin.Get("/", studentSvc.GetAllStudents)
	admin.Get("/:id", studentSvc.GetStudentByID)
	admin.Put("/:id/advisor", studentSvc.SetAdvisor)
}
