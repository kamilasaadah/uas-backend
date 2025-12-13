package route

import (
	"uas-backend/app/repository"
	"uas-backend/app/service"
	"uas-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func AchievementRoutes(
	r fiber.Router,
	achievementSvc *service.AchievementService,
	userRepo repository.UserRepository,
) {

	api := r.Group(
		"/achievements",
		middleware.JWTAuth(userRepo),
		middleware.RequirePermission("achievement:create"),
	)

	api.Post("/", achievementSvc.CreateAchievement)
}
