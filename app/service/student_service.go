package service

import (
	"context"
	"uas-backend/app/model"
	"uas-backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type StudentService struct {
	studentRepo     repository.StudentRepository
	lecturerRepo    repository.LecturerRepository
	achievementRepo repository.AchievementRepository
	refRepo         repository.AchievementReferenceRepository
}

func NewStudentService(
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	achievementRepo repository.AchievementRepository,
	refRepo repository.AchievementReferenceRepository,
) *StudentService {
	return &StudentService{
		studentRepo:     studentRepo,
		lecturerRepo:    lecturerRepo,
		achievementRepo: achievementRepo,
		refRepo:         refRepo,
	}
}

// SetAdvisor godoc
// @Summary Set student advisor
// @Description Admin only. Assign or update advisor (lecturer) for a student
// @Tags Students
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Student ID"
// @Param request body model.SetAdvisorRequest true "Advisor payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /students/{id}/advisor [put]
func (s *StudentService) SetAdvisor(c *fiber.Ctx) error {
	studentID := c.Params("id")

	req := new(model.SetAdvisorRequest)
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// cek apakah advisor lecturer exist
	lecturer, err := s.lecturerRepo.GetLecturerByID(context.Background(), req.AdvisorID)
	if err != nil || lecturer == nil {
		return fiber.NewError(fiber.StatusNotFound, "advisor not found")
	}

	// update advisor
	err = s.studentRepo.UpdateAdvisor(context.Background(), studentID, req.AdvisorID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update advisor")
	}

	return c.JSON(fiber.Map{
		"message": "advisor updated successfully",
	})
}

// GetAllStudents godoc
// @Summary Get all students
// @Description Admin only. Get list of all students
// @Tags Students
// @Security BearerAuth
// @Produce json
// @Success 200 {array} model.Student
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /students [get]
func (s *StudentService) GetAllStudents(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)

	if claims.Role != "Admin" {
		return fiber.NewError(403, "forbidden")
	}

	students, err := s.studentRepo.GetAllStudents(c.Context())
	if err != nil {
		return fiber.NewError(500, "failed to fetch students")
	}
	return c.JSON(students)
}

// GetStudentByID godoc
// @Summary Get student by ID
// @Description Admin only. Get detail of a student
// @Tags Students
// @Security BearerAuth
// @Produce json
// @Param id path string true "Student ID"
// @Success 200 {object} model.Student
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /students/{id} [get]
func (s *StudentService) GetStudentByID(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)
	studentID := c.Params("id")

	if claims.Role != "Admin" {
		return fiber.NewError(403, "forbidden")
	}

	student, err := s.studentRepo.GetStudentByID(c.Context(), studentID)
	if err != nil {
		return fiber.NewError(404, "student not found")
	}
	return c.JSON(student)
}

// GetStudentAchievements godoc
// @Summary Get student achievements
// @Description
//
//	Admin: access any student achievements
//	Dosen Wali: only achievements of own advisee
//
// @Tags Students
// @Security BearerAuth
// @Produce json
// @Param id path string true "Student ID"
// @Success 200 {array} model.StudentAchievementResponse
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /students/{id}/achievements [get]
func (s *StudentService) GetStudentAchievements(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)
	studentID := c.Params("id")

	student, err := s.studentRepo.GetStudentByID(c.Context(), studentID)
	if err != nil {
		return fiber.NewError(404, "student not found")
	}

	// 🔐 akses role
	switch claims.Role {
	case "Admin":
		// ok

	case "Dosen Wali":
		// ambil lecturer berdasarkan user_id dari JWT
		lecturer, err := s.lecturerRepo.GetLecturerProfile(
			c.Context(),
			claims.UserID, // users.id
		)
		if err != nil || lecturer == nil {
			return fiber.NewError(403, "lecturer profile not found")
		}

		// cocokkan advisor_id (lecturers.id)
		if student.AdvisorID != lecturer.ID {
			return fiber.NewError(403, "forbidden")
		}

	default:
		return fiber.NewError(403, "forbidden")
	}

	// =========================
	// Mongo
	// =========================
	achievements, err := s.achievementRepo.FindByStudentIDs(
		c.Context(),
		[]string{student.ID},
	)
	if err != nil {
		return fiber.NewError(500, "failed to fetch achievements")
	}

	// =========================
	// Postgres
	// =========================
	refs, err := s.refRepo.GetByStudentID(c.Context(), student.ID)
	if err != nil {
		return fiber.NewError(500, "failed to fetch achievement references")
	}

	refMap := map[string]*model.AchievementReference{}
	for _, r := range refs {
		refMap[r.MongoAchievementID] = r
	}

	var result []model.StudentAchievementResponse
	for _, a := range achievements {
		result = append(result, model.StudentAchievementResponse{
			Achievement: a,
			Reference:   refMap[a.ID.Hex()],
		})
	}

	return c.JSON(result)
}
