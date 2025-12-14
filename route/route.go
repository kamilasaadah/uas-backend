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
	achievementRepo := repository.NewAchievementRepository(database.MongoDB)
	achievementRefRepo := repository.NewAchievementReferenceRepository(database.PG)

	// === INIT SERVICE ===
	authService := service.NewAuthService(userRepo, studentRepo)
	userService := service.NewUserService(userRepo, studentRepo, lecturerRepo)
	studentSvc := service.NewStudentService(studentRepo, lecturerRepo)
	lecturerSvc := service.NewLecturerService(lecturerRepo, studentRepo)
	achievementSvc := service.NewAchievementService(
		achievementRepo,
		achievementRefRepo,
		studentRepo,
	)

	// ROUTES
	AuthRoutes(api.Group("/auth"), authService, userRepo)
	AdminRoutes(api, userService, userRepo)
	StudentRoutes(api, studentSvc, userRepo)
	LecturerRoutes(api, lecturerSvc, userRepo)
	AchievementRoutes(api, achievementSvc, userRepo)

}
