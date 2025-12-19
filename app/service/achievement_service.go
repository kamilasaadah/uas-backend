package service

import (
	"fmt"
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

/*
	=======================
	  UPDATE REQUEST DTO

=======================
*/
type UpdateAchievementRequest struct {
	Title       *string        `json:"title"`
	Description *string        `json:"description"`
	Details     map[string]any `json:"details"`
	Tags        []string       `json:"tags"`
	Points      *int           `json:"points"`
}

// package service (atau dto khusus kalau kamu punya)
type RejectAchievementRequest struct {
	RejectionNote string `json:"rejection_note"`
}

type AchievementService struct {
	achievementRepo repository.AchievementRepository
	referenceRepo   repository.AchievementReferenceRepository
	studentRepo     repository.StudentRepository
	lecturerRepo    repository.LecturerRepository
}

func NewAchievementService(
	achievementRepo repository.AchievementRepository,
	referenceRepo repository.AchievementReferenceRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
) *AchievementService {
	return &AchievementService{
		achievementRepo: achievementRepo,
		referenceRepo:   referenceRepo,
		studentRepo:     studentRepo,
		lecturerRepo:    lecturerRepo,
	}
}

// CreateAchievement godoc
// @Summary Buat prestasi baru
// @Description Mahasiswa otomatis pakai student_id dari JWT, Admin wajib mengisi studentId
// @Tags Achievements
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body service.CreateAchievementRequest true "Create Achievement Payload"
// @Success 201 {object} map[string]interface{} "Achievement created"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 403 {object} map[string]interface{} "Access denied"
// @Failure 500 {object} map[string]interface{} "Failed to create achievement"
// @Router /achievements [post]
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

// UploadAttachment godoc
// @Summary Upload lampiran prestasi
// @Description
// Hanya boleh jika status prestasi masih draft
// @Tags Achievements
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Achievement ID"
// @Param file formData file true "Attachment file"
// @Success 200 {object} map[string]interface{} "Attachment uploaded"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 403 {object} map[string]interface{} "Access denied"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Failure 500 {object} map[string]interface{} "Failed to upload attachment"
// @Router /achievements/{id}/attachments [post]
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

// UpdateAchievement godoc
// @Summary Update prestasi
// @Description
// Hanya bisa update jika status masih draft
// @Tags Achievements
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID"
// @Param body body service.UpdateAchievementRequest true "Update Achievement Payload"
// @Success 200 {object} map[string]interface{} "Achievement updated"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 403 {object} map[string]interface{} "Access denied"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Failure 500 {object} map[string]interface{} "Failed to update achievement"
// @Router /achievements/{id} [put]
func (s *AchievementService) UpdateAchievement(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)
	achievementID := c.Params("id")

	// parse ObjectID
	objID, err := primitive.ObjectIDFromHex(achievementID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid achievement id")
	}

	// 1️⃣ cek reference di PostgreSQL
	ref, err := s.referenceRepo.GetByAchievementID(c.Context(), achievementID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "achievement reference not found")
	}

	if ref.Status != "draft" {
		return fiber.NewError(fiber.StatusBadRequest, "only draft achievements can be updated")
	}

	// 2️⃣ cek authorization
	if claims.Role == "Mahasiswa" {
		if claims.StudentID == "" || ref.StudentID != claims.StudentID {
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}
	}

	// 3️⃣ parse request
	var req UpdateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// 4️⃣ ambil achievement dari MongoDB
	achievement, err := s.achievementRepo.GetByID(c.Context(), objID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "achievement not found")
	}

	// 5️⃣ update fields
	if req.Title != nil {
		achievement.Title = *req.Title
	}
	if req.Description != nil {
		achievement.Description = *req.Description
	}
	if req.Details != nil {
		achievement.Details = req.Details
	}
	if req.Tags != nil {
		achievement.Tags = req.Tags
	}
	if req.Points != nil {
		achievement.Points = *req.Points
	}
	achievement.UpdatedAt = time.Now()

	// 6️⃣ simpan perubahan
	if err := s.achievementRepo.Update(c.Context(), achievement); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update achievement")
	}

	return c.JSON(fiber.Map{
		"message": "achievement updated",
		"data":    achievement,
	})
}

// DeleteAchievement godoc
// @Summary Hapus prestasi
// @Description
// Soft delete, hanya bisa jika status masih draft
// @Tags Achievements
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID"
// @Success 200 {object} map[string]interface{} "Achievement deleted"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 403 {object} map[string]interface{} "Access denied"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Failure 500 {object} map[string]interface{} "Failed to delete achievement"
// @Router /achievements/{id} [delete]
func (s *AchievementService) DeleteAchievement(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)
	achievementID := c.Params("id")

	// parse ObjectID
	objID, err := primitive.ObjectIDFromHex(achievementID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid achievement id")
	}

	// 1️⃣ cek reference di PostgreSQL
	ref, err := s.referenceRepo.GetByAchievementID(c.Context(), achievementID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "achievement reference not found")
	}

	// hanya draft yang bisa dihapus
	if ref.Status != "draft" {
		return fiber.NewError(fiber.StatusBadRequest, "only draft achievements can be deleted")
	}

	// 2️⃣ cek authorization
	switch claims.Role {
	case "Mahasiswa":
		if claims.StudentID == "" || ref.StudentID != claims.StudentID {
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}
	case "Admin":
		// admin bisa delete semua
	default:
		return fiber.NewError(fiber.StatusForbidden, "access denied")
	}

	// 3️⃣ soft delete MongoDB
	if err := s.achievementRepo.SoftDelete(c.Context(), objID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to soft delete achievement")
	}

	// 4️⃣ update PostgreSQL reference
	if err := s.referenceRepo.MarkDeleted(c.Context(), achievementID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update achievement reference")
	}

	return c.JSON(fiber.Map{
		"message": "achievement deleted",
	})
}

// GetAchievements godoc
// @Summary Ambil daftar prestasi
// @Description
// Mahasiswa: hanya prestasi miliknya
// Dosen Wali: prestasi mahasiswa bimbingan
// Admin: semua prestasi
// @Tags Achievements
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "List achievements"
// @Failure 403 {object} map[string]interface{} "Access denied"
// @Failure 500 {object} map[string]interface{} "Failed to fetch achievements"
// @Router /achievements/ [get]
func (s *AchievementService) GetAchievements(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)

	fmt.Println("\n================ ACHIEVEMENT DEBUG ================")
	fmt.Println("ROLE        :", claims.Role)
	fmt.Println("USER ID     :", claims.UserID)

	switch claims.Role {

	// =====================
	// MAHASISWA
	// =====================
	case "Mahasiswa":
		fmt.Println(">> FLOW: MAHASISWA")

		student, err := s.studentRepo.GetStudentProfile(
			c.Context(),
			claims.UserID,
		)
		if err != nil {
			fmt.Println("❌ STUDENT PROFILE NOT FOUND FOR USER ID:", claims.UserID)
			return fiber.NewError(fiber.StatusForbidden, "student profile not found")
		}

		fmt.Println("✅ STUDENT FOUND")
		fmt.Println("STUDENT.ID (UUID) :", student.ID)

		data, err := s.achievementRepo.FindByStudentIDs(
			c.Context(),
			[]string{student.ID},
		)
		if err != nil {
			fmt.Println("❌ MONGO QUERY ERROR:", err)
			return fiber.NewError(500, "failed to fetch achievements")
		}

		fmt.Println("🎯 RESULT COUNT :", len(data))
		return c.JSON(fiber.Map{"data": data})

	// =====================
	// DOSEN WALI (FIXED)
	// =====================
	case "Dosen Wali":
		fmt.Println(">> FLOW: DOSEN WALI")

		// 1️⃣ ambil lecturer dari user_id (JWT)
		lecturer, err := s.lecturerRepo.GetLecturerProfile(
			c.Context(),
			claims.UserID,
		)
		if err != nil {
			fmt.Println("❌ LECTURER PROFILE NOT FOUND FOR USER ID:", claims.UserID)
			return fiber.NewError(fiber.StatusForbidden, "lecturer profile not found")
		}

		fmt.Println("LECTURER.ID (advisor_id):", lecturer.ID)

		// 2️⃣ ambil mahasiswa bimbingan
		students, err := s.studentRepo.GetStudentsByAdvisor(
			c.Context(),
			lecturer.ID, // ✅ advisor_id
		)
		if err != nil {
			fmt.Println("❌ FAILED FETCH ADVISEES:", err)
			return fiber.NewError(500, "failed to fetch advisees")
		}

		fmt.Println("ADVISEE COUNT :", len(students))

		if len(students) == 0 {
			fmt.Println("⚠️ DOSEN TIDAK PUNYA ANAK WALI")
			return c.JSON(fiber.Map{"data": []any{}})
		}

		var studentIDs []string
		for _, st := range students {
			fmt.Println(" - ADVISEE STUDENT.ID :", st.ID)
			studentIDs = append(studentIDs, st.ID)
		}

		fmt.Println("STUDENT IDS FILTER :", studentIDs)

		data, err := s.achievementRepo.FindByStudentIDs(
			c.Context(),
			studentIDs,
		)
		if err != nil {
			fmt.Println("❌ MONGO QUERY ERROR:", err)
			return fiber.NewError(500, "failed to fetch achievements")
		}

		fmt.Println("🎯 RESULT COUNT :", len(data))
		return c.JSON(fiber.Map{"data": data})

	// =====================
	// ADMIN
	// =====================
	case "Admin":
		fmt.Println(">> FLOW: ADMIN (FIND ALL)")

		data, err := s.achievementRepo.FindAll(c.Context())
		if err != nil {
			fmt.Println("❌ MONGO FIND ALL ERROR:", err)
			return fiber.NewError(500, "failed to fetch achievements")
		}

		fmt.Println("🎯 RESULT COUNT :", len(data))
		return c.JSON(fiber.Map{"data": data})

	default:
		fmt.Println("❌ UNKNOWN ROLE:", claims.Role)
		return fiber.NewError(fiber.StatusForbidden, "access denied")
	}
}

// GetAchievementByID godoc
// @Summary Ambil detail prestasi
// @Description
// Mahasiswa hanya boleh lihat prestasi sendiri
// Dosen Wali hanya prestasi mahasiswa bimbingan
// Admin bebas
// @Tags Achievements
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (Mongo ObjectID)"
// @Success 200 {object} map[string]interface{} "Achievement detail"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 403 {object} map[string]interface{} "Access denied"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Router /achievements/{id} [get]
func (s *AchievementService) GetAchievementByID(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)
	id := c.Params("id")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid achievement id")
	}

	achievement, err := s.achievementRepo.GetByID(c.Context(), objID)
	if err != nil || achievement.IsDeleted {
		return fiber.NewError(fiber.StatusNotFound, "achievement not found")
	}

	switch claims.Role {

	case "Mahasiswa":
		student, err := s.studentRepo.GetStudentProfile(
			c.Context(),
			claims.UserID,
		)
		if err != nil || achievement.StudentID != student.ID {
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}

	case "Dosen Wali":
		// 1️⃣ ambil lecturer dari user_id
		lecturer, err := s.lecturerRepo.GetLecturerProfile(
			c.Context(),
			claims.UserID,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusForbidden, "lecturer profile not found")
		}

		// 2️⃣ ambil mahasiswa bimbingan
		students, err := s.studentRepo.GetStudentsByAdvisor(
			c.Context(),
			lecturer.ID, // ✅ advisor_id
		)
		if err != nil {
			return fiber.NewError(500, "failed to fetch advisees")
		}

		allowed := false
		for _, st := range students {
			if st.ID == achievement.StudentID {
				allowed = true
				break
			}
		}

		if !allowed {
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}

	case "Admin":
		// full access

	default:
		return fiber.NewError(fiber.StatusForbidden, "access denied")
	}

	return c.JSON(fiber.Map{"data": achievement})
}

// SubmitAchievement godoc
// @Summary Submit prestasi untuk verifikasi
// @Description
// Mengubah status dari draft ke submitted
// @Tags Achievements
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID"
// @Success 200 {object} map[string]interface{} "Achievement submitted"
// @Failure 400 {object} map[string]interface{} "Invalid status"
// @Failure 403 {object} map[string]interface{} "Access denied"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Router /achievements/{id}/submit [post]
func (s *AchievementService) SubmitAchievement(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)
	achievementID := c.Params("id")

	// 1️⃣ ambil reference
	ref, err := s.referenceRepo.GetByAchievementID(
		c.Context(),
		achievementID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "achievement reference not found")
	}

	// 2️⃣ hanya draft
	if ref.Status != "draft" {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"only draft achievement can be submitted",
		)
	}

	// 3️⃣ authorization
	switch claims.Role {
	case "Mahasiswa":
		if claims.StudentID == "" || ref.StudentID != claims.StudentID {
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}

	case "Admin":
		// allowed

	default:
		return fiber.NewError(fiber.StatusForbidden, "access denied")
	}

	// 4️⃣ update status → submitted
	updatedRef, err := s.referenceRepo.Submit(
		c.Context(),
		achievementID,
	)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			"failed to submit achievement",
		)
	}

	// 6️⃣ response
	return c.JSON(fiber.Map{
		"message": "achievement submitted for verification",
		"data":    updatedRef,
	})
}

// VerifyAchievement godoc
// @Summary Verifikasi prestasi
// @Description
// Hanya untuk Dosen Wali dan Admin
// @Tags Achievements
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID"
// @Success 200 {object} map[string]interface{} "Achievement verified"
// @Failure 400 {object} map[string]interface{} "Invalid status"
// @Failure 403 {object} map[string]interface{} "Access denied"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Router /achievements/{id}/verify [post]
func (s *AchievementService) VerifyAchievement(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)
	achievementID := c.Params("id")

	// 1️⃣ role check
	switch claims.Role {
	case "Dosen Wali", "Admin":
		// allowed
	default:
		return fiber.NewError(fiber.StatusForbidden, "access denied")
	}

	// 2️⃣ ambil reference
	ref, err := s.referenceRepo.GetByAchievementID(
		c.Context(),
		achievementID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "achievement reference not found")
	}

	// 3️⃣ hanya submitted yang bisa diverifikasi
	if ref.Status != "submitted" {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"only submitted achievement can be verified",
		)
	}

	// 4️⃣ khusus dosen wali → pastikan mahasiswa adalah anak walinya
	if claims.Role == "Dosen Wali" {
		lecturer, err := s.lecturerRepo.GetLecturerProfile(
			c.Context(),
			claims.UserID,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusForbidden, "lecturer profile not found")
		}

		students, err := s.studentRepo.GetStudentsByAdvisor(
			c.Context(),
			lecturer.ID,
		)
		if err != nil {
			return fiber.NewError(500, "failed to fetch advisees")
		}

		allowed := false
		for _, st := range students {
			if st.ID == ref.StudentID {
				allowed = true
				break
			}
		}

		if !allowed {
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}
	}

	// 5️⃣ verify (PostgreSQL)
	updatedRef, err := s.referenceRepo.Verify(
		c.Context(),
		achievementID,
		claims.UserID,
	)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			"failed to verify achievement",
		)
	}

	return c.JSON(fiber.Map{
		"message": "achievement verified",
		"data":    updatedRef,
	})
}

// / RejectAchievement godoc
// @Summary Tolak prestasi
// @Description Hanya untuk Dosen Wali dan Admin, status harus submitted
// @Tags Achievements
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID"
// @Param body body service.RejectAchievementRequest true "Reject Payload"
// @Success 200 {object} map[string]interface{} "Achievement rejected"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 403 {object} map[string]interface{} "Access denied"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Router /achievements/{id}/reject [post]
func (s *AchievementService) RejectAchievement(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)
	achievementID := c.Params("id")

	// 1️⃣ role check
	switch claims.Role {
	case "Dosen Wali", "Admin":
		// allowed
	default:
		return fiber.NewError(fiber.StatusForbidden, "access denied")
	}

	// 2️⃣ parse request
	var req RejectAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.RejectionNote == "" {
		return fiber.NewError(fiber.StatusBadRequest, "rejection_note is required")
	}

	// 3️⃣ ambil reference
	ref, err := s.referenceRepo.GetByAchievementID(
		c.Context(),
		achievementID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "achievement reference not found")
	}

	// 4️⃣ hanya submitted
	if ref.Status != "submitted" {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"only submitted achievement can be rejected",
		)
	}

	// 5️⃣ dosen wali → validasi anak wali
	if claims.Role == "Dosen Wali" {
		lecturer, err := s.lecturerRepo.GetLecturerProfile(
			c.Context(),
			claims.UserID,
		)
		if err != nil {
			return fiber.NewError(fiber.StatusForbidden, "lecturer profile not found")
		}

		students, err := s.studentRepo.GetStudentsByAdvisor(
			c.Context(),
			lecturer.ID,
		)
		if err != nil {
			return fiber.NewError(500, "failed to fetch advisees")
		}

		allowed := false
		for _, st := range students {
			if st.ID == ref.StudentID {
				allowed = true
				break
			}
		}

		if !allowed {
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}
	}

	// 6️⃣ reject (PostgreSQL)
	updatedRef, err := s.referenceRepo.Reject(
		c.Context(),
		achievementID,
		req.RejectionNote,
	)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			"failed to reject achievement",
		)
	}

	return c.JSON(fiber.Map{
		"message": "achievement rejected",
		"data":    updatedRef,
	})
}

// GetAchievementHistory godoc
// @Summary Ambil riwayat status prestasi
// @Description Menampilkan timeline status: draft, submitted, verified, rejected
// @Tags Achievements
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID"
// @Success 200 {object} map[string]interface{} "Achievement history"
// @Failure 403 {object} map[string]interface{} "Access denied"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Router /achievements/{id}/history [get]
func (s *AchievementService) GetAchievementHistory(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)
	achievementID := c.Params("id")

	// 1️⃣ ambil reference (SUMBER STATUS)
	ref, err := s.referenceRepo.GetByAchievementID(
		c.Context(),
		achievementID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "achievement reference not found")
	}

	// 2️⃣ ambil achievement MongoDB
	objID, err := primitive.ObjectIDFromHex(achievementID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid achievement id")
	}

	achievement, err := s.achievementRepo.GetByID(c.Context(), objID)
	if err != nil || achievement.IsDeleted {
		return fiber.NewError(fiber.StatusNotFound, "achievement not found")
	}

	// 3️⃣ AUTH (tetap seperti punyamu)
	switch claims.Role {
	case "Mahasiswa":
		if claims.StudentID == "" || ref.StudentID != claims.StudentID {
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}
	case "Admin":
		// ok
	case "Dosen Wali":
		lecturer, err := s.lecturerRepo.GetLecturerProfile(c.Context(), claims.UserID)
		if err != nil {
			return fiber.NewError(fiber.StatusForbidden, "lecturer profile not found")
		}
		students, _ := s.studentRepo.GetStudentsByAdvisor(c.Context(), lecturer.ID)

		allowed := false
		for _, st := range students {
			if st.ID == ref.StudentID {
				allowed = true
				break
			}
		}
		if !allowed {
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}
	default:
		return fiber.NewError(fiber.StatusForbidden, "access denied")
	}

	// 4️⃣ BENTUK HISTORY DARI reference
	history := []model.AchievementStatusHistory{}

	// draft (selalu ada)
	history = append(history, model.AchievementStatusHistory{
		Status:    "draft",
		UpdatedAt: ref.CreatedAt,
	})

	if ref.SubmittedAt != nil {
		history = append(history, model.AchievementStatusHistory{
			Status:    "submitted",
			UpdatedAt: *ref.SubmittedAt,
		})
	}

	if ref.VerifiedAt != nil {
		history = append(history, model.AchievementStatusHistory{
			Status:    "verified",
			UpdatedAt: *ref.VerifiedAt,
			UpdatedBy: ref.VerifiedBy,
		})
	}

	if ref.Status == "rejected" && ref.RejectionNote != nil {
		history = append(history, model.AchievementStatusHistory{
			Status:    "rejected",
			Note:      ref.RejectionNote,
			UpdatedAt: ref.UpdatedAt,
		})
	}

	// 5️⃣ RESPONSE
	return c.JSON(model.AchievementHistoryResponse{
		Achievement: *achievement,
		History:     history,
	})
}
