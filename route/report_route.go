package route

import (
	"uas-backend/app/repository"
	"uas-backend/app/service"
	"uas-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func ReportRoutes(
	r fiber.Router,
	reportService *service.ReportService,
	userRepo repository.UserRepository,
) {

	report := r.Group(
		"/reports",
		middleware.JWTAuth(userRepo),
	)

	report.Get("/statistics", reportService.GetStatistics)
	report.Get("/student/:id", reportService.GetStudentStatistics)
}
