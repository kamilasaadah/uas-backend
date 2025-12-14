package service

import (
	"os"
	"time"

	"uas-backend/app/model"
	"uas-backend/app/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)
	achievementID := c.Params("id")

	// 1️⃣ cek reference (PostgreSQL)
	ref, err := s.referenceRepo.GetByAchievementID(
		c.Context(),
		achievementID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "achievement reference not found")
	}

	if ref.Status != "draft" {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"attachments only allowed for draft achievement",
		)
	}

	// 2️⃣ auth
	if claims.Role == "Mahasiswa" {

		// pastikan student_id ada di JWT
		if claims.StudentID == "" {
			return fiber.NewError(fiber.StatusForbidden, "student id missing")
		}

		// bandingkan STUDENT ID vs STUDENT ID
		if ref.StudentID != claims.StudentID {
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}
	}

	// 3️⃣ ambil file
	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}

	// ✅ pastikan folder uploads ada
	_ = os.MkdirAll("./uploads", 0755)

	// 4️⃣ simpan file
	fileURL := "/uploads/" + file.Filename
	if err := c.SaveFile(file, "."+fileURL); err != nil {
		return fiber.NewError(500, "failed to save file")
	}

	// 5️⃣ parse ObjectID (PAKAI YANG ATAS)
	objID, err := primitive.ObjectIDFromHex(achievementID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid achievement id")
	}

	attachment := model.Attachment{
		FileName:   file.Filename,
		FileURL:    fileURL,
		FileType:   file.Header.Get("Content-Type"),
		UploadedAt: time.Now(),
	}

	if err := s.achievementRepo.AddAttachment(
		c.Context(),
		objID,
		attachment,
	); err != nil {
		return fiber.NewError(500, "failed to save attachment")
	}

	return c.JSON(fiber.Map{
		"message": "attachment uploaded",
		"data":    attachment,
	})
}
