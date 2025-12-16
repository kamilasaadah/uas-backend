package service

import (
	"uas-backend/app/model"
	"uas-backend/app/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// achievement type yang DIKENAL
var knownAchievementTypes = map[string]bool{
	"academic":      true,
	"competition":   true,
	"organization":  true,
	"publication":   true,
	"certification": true,
}

type ReportService struct {
	reportRepo      repository.ReportRepository
	achievementRepo repository.AchievementRepository
}

func NewReportService(
	reportRepo repository.ReportRepository,
	achievementRepo repository.AchievementRepository,
) *ReportService {
	return &ReportService{
		reportRepo:      reportRepo,
		achievementRepo: achievementRepo,
	}
}

func (s *ReportService) GetStatistics(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)

	// 1️⃣ Ambil mongo achievement IDs dari PostgreSQL
	var mongoIDs []string
	var err error

	if claims.Role == "Mahasiswa" {
		mongoIDs, err = s.reportRepo.GetVerifiedAchievementIDsByStudent(
			c.Context(),
			claims.StudentID,
		)
	} else {
		mongoIDs, err = s.reportRepo.GetVerifiedAchievementIDs(c.Context())
	}

	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			"failed to load verified achievements",
		)
	}

	// 2️⃣ Convert string → ObjectID
	var objectIDs []primitive.ObjectID
	for _, id := range mongoIDs {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue // skip ID rusak
		}
		objectIDs = append(objectIDs, oid)
	}

	// 3️⃣ Ambil data achievement dari MongoDB
	achievements, err := s.achievementRepo.FindByIDs(
		c.Context(),
		objectIDs,
	)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			"failed to fetch achievements",
		)
	}

	// 4️⃣ Hitung statistik
	typeCount := map[string]int{}
	periodCount := map[string]int{}
	levelCount := map[string]int{}
	studentCount := map[string]int{}

	for _, a := range achievements {

		// TYPE
		t := a.AchievementType
		if !knownAchievementTypes[t] {
			t = "other"
		}
		typeCount[t]++

		// PERIOD (YYYY-MM)
		period := a.CreatedAt.Format("2006-01")
		periodCount[period]++

		// STUDENT
		studentCount[a.StudentID]++

		// COMPETITION LEVEL (khusus competition)
		if lvl, ok := a.Details["competitionLevel"].(string); ok {
			levelCount[lvl]++
		}
	}

	// 5️⃣ Mapping ke response model
	var typeStats []model.AchievementTypeStat
	for k, v := range typeCount {
		typeStats = append(typeStats, model.AchievementTypeStat{
			Type:  k,
			Total: v,
		})
	}

	var periodStats []model.PeriodStat
	for k, v := range periodCount {
		periodStats = append(periodStats, model.PeriodStat{
			Period: k,
			Total:  v,
		})
	}

	var levelStats []model.CompetitionLevelStat
	for k, v := range levelCount {
		levelStats = append(levelStats, model.CompetitionLevelStat{
			Level: k,
			Total: v,
		})
	}

	var studentStats []model.StudentAchievementStat
	for k, v := range studentCount {
		studentStats = append(studentStats, model.StudentAchievementStat{
			StudentID: k,
			Total:     v,
		})
	}

	// 6️⃣ Response
	return c.JSON(model.AchievementStatisticsResponse{
		TotalPerType:      typeStats,
		TotalPerPeriod:    periodStats,
		CompetitionLevels: levelStats,
		TotalPerStudent:   studentStats,
	})
}
