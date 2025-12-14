package service

import (
	"time"

	"uas-backend/app/model"
	"uas-backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

/* =======================
   REQUEST DTO
======================= */

type CreateAchievementRequest struct {
	StudentID       string         `json:"studentId"`
	AchievementType string         `json:"achievementType"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	Details         map[string]any `json:"details"`
	Tags            []string       `json:"tags"`
	Points          int            `json:"points"`
}

type AchievementService struct {
	achievementRepo repository.AchievementRepository
	referenceRepo   repository.AchievementReferenceRepository
	studentRepo     repository.StudentRepository
}

func NewAchievementService(
	achievementRepo repository.AchievementRepository,
	referenceRepo repository.AchievementReferenceRepository,
	studentRepo repository.StudentRepository, // ✅ TAMBAH
) *AchievementService {
	return &AchievementService{
		achievementRepo: achievementRepo,
		referenceRepo:   referenceRepo,
		studentRepo:     studentRepo,
	}
}

func (s *AchievementService) CreateAchievement(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)

	var req CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	var studentID string

	switch claims.Role {

	case "Mahasiswa":
		// mahasiswa → ambil student_id dari user_id
		student, err := s.studentRepo.GetStudentProfile(
			c.Context(),
			claims.UserID,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusForbidden, "student profile not found")
		}
		studentID = student.ID

	case "Admin":
		// admin → HARUS eksplisit target mahasiswa
		if req.StudentID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "studentId is required for admin")
		}
		studentID = req.StudentID

	default:
		return fiber.NewError(fiber.StatusForbidden, "access denied")
	}

	now := time.Now()

	achievement := &model.Achievement{
		StudentID:       studentID,
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Tags:            req.Tags,
		Points:          req.Points,
		Attachments:     []model.Attachment{},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// 1️⃣ MongoDB
	oid, err := s.achievementRepo.Create(c.Context(), achievement)
	if err != nil {
		return fiber.NewError(500, "failed to create achievement")
	}

	// 2️⃣ PostgreSQL reference
	if err := s.referenceRepo.CreateDraft(
		c.Context(),
		studentID, // ✅ FIX
		oid.Hex(),
	); err != nil {
		return fiber.NewError(500, "failed to create achievement reference")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "achievement created",
		"data":    achievement,
	})
}
