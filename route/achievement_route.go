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
	)

	api.Get("/", achievementSvc.GetAchievements)
	api.Get("/:id", achievementSvc.GetAchievementByID)

	api.Post(
		"/",
		middleware.RequirePermission("achievement:create"),
		achievementSvc.CreateAchievement,
	)

	api.Post(
		"/:id/attachments",
		middleware.RequirePermission("achievement:update"),
		achievementSvc.UploadAttachment,
	)

	api.Put(
		"/:id",
		middleware.RequirePermission("achievement:update"),
		achievementSvc.UpdateAchievement,
	)

	api.Delete(
		"/:id",
		// mahasiswa & admin boleh, permission bebas
		achievementSvc.DeleteAchievement,
	)

	api.Post(
		"/:id/submit",
		achievementSvc.SubmitAchievement,
	)

}
