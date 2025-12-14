package service

import (
	"fmt"
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
	lecturerIDParam := c.Params("id")
	if lecturerIDParam == "" {
		return fiber.NewError(fiber.StatusBadRequest, "lecturer id is required")
	}

	fmt.Println("=== DEBUG ADVISEES ===")
	fmt.Println("JWT UserID:", claims.UserID)
	fmt.Println("JWT Role:", claims.Role)
	fmt.Println("Path lecturerID:", lecturerIDParam)
	fmt.Println("Permissions:", c.Locals("permissions"))

	switch claims.Role {
	case "Admin":
		students, err := s.studentRepo.GetAllStudents(c.Context())
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch students")
		}
		return c.JSON(students)

	case "Dosen Wali":
		lecturer, err := s.lecturerRepo.GetLecturerProfile(c.Context(), claims.UserID)
		if err != nil {
			return fiber.NewError(fiber.StatusForbidden, "lecturer profile not found")
		}

		if lecturer.ID != lecturerIDParam {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}

		students, err := s.studentRepo.GetStudentsByAdvisor(c.Context(), lecturer.ID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch advisees")
		}
		return c.JSON(students)

	default:
		return fiber.NewError(fiber.StatusForbidden, "forbidden")
	}
}
