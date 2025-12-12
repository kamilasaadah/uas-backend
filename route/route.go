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
	studentRepo := repository.NewStudentRepository(database.PG)
	lecturerRepo := repository.NewLecturerRepository(database.PG)

	// === INIT SERVICE ===
	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo, studentRepo, lecturerRepo)
	studentSvc := service.NewStudentService(studentRepo, lecturerRepo)
	lecturerSvc := service.NewLecturerService(lecturerRepo, studentRepo)


	// ROUTES
	AuthRoutes(api.Group("/auth"), authService)
	AdminRoutes(api, userService)
	StudentRoutes(api, studentSvc)
	LecturerRoutes(api, lecturerSvc)

}
