package service

import (
	"uas-backend/app/model"
	"uas-backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type LecturerService struct {
	lecturerRepo repository.LecturerRepository
	studentRepo  repository.StudentRepository
}

func NewLecturerService(lecturerRepo repository.LecturerRepository, studentRepo repository.StudentRepository) *LecturerService {
	return &LecturerService{
		lecturerRepo: lecturerRepo,
		studentRepo:  studentRepo,
	}
}

// =====================================
// GET /lecturers  (Admin Only)
// =====================================
func (s *LecturerService) GetAllLecturers(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)

	if claims.Role != "Admin" {
		return fiber.NewError(fiber.StatusForbidden, "forbidden")
	}

	list, err := s.lecturerRepo.GetAllLecturers(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch lecturers")
	}

	return c.JSON(list)
}

// =====================================
// GET /lecturers/:id/advisees (Admin Only)
// =====================================
func (s *LecturerService) GetAdvisees(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)

	if claims.Role != "Admin" {
		return fiber.NewError(fiber.StatusForbidden, "forbidden")
	}

	lecturerID := c.Params("id")
	if lecturerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "lecturer id is required")
	}

	// Validate lecturer exists
	_, err := s.lecturerRepo.GetLecturerByID(c.Context(), lecturerID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "lecturer not found")
	}

	// Fetch advisee students
	students, err := s.studentRepo.GetStudentsByAdvisor(c.Context(), lecturerID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch advisees")
	}

	return c.JSON(students)
}
