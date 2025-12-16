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

// PUT /api/v1/students/:id/advisor
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
