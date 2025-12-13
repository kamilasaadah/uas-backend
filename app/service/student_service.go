package service

import (
	"context"
	"uas-backend/app/model"
	"uas-backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type StudentService struct {
	studentRepo  repository.StudentRepository
	lecturerRepo repository.LecturerRepository
}

func NewStudentService(studentRepo repository.StudentRepository, lecturerRepo repository.LecturerRepository) *StudentService {
	return &StudentService{
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
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

	// 🔐 ADMIN → semua mahasiswa
	if claims.Role == "Admin" {
		students, err := s.studentRepo.GetAllStudents(c.Context())
		if err != nil {
			return fiber.NewError(500, "failed to fetch students")
		}
		return c.JSON(students)
	}

	// 👨‍🏫 DOSEN → anak wali saja
	lecturer, err := s.lecturerRepo.GetLecturerProfile(
		c.Context(),
		claims.UserID,
	)
	if err != nil {
		return fiber.NewError(403, "lecturer profile not found")
	}

	students, err := s.studentRepo.GetStudentsByAdvisor(
		c.Context(),
		lecturer.ID, // 🔥 STRING UUID
	)
	if err != nil {
		return fiber.NewError(500, "failed to fetch students")
	}

	return c.JSON(students)
}

func (s *StudentService) GetStudentByID(c *fiber.Ctx) error {
	claims := c.Locals("user").(*model.JWTClaims)
	studentID := c.Params("id")

	// 🔐 ADMIN → bebas
	if claims.Role == "Admin" {
		student, err := s.studentRepo.GetStudentByID(c.Context(), studentID)
		if err != nil {
			return fiber.NewError(404, "student not found")
		}
		return c.JSON(student)
	}

	// 👨‍🏫 DOSEN → validasi anak wali
	lecturer, err := s.lecturerRepo.GetLecturerProfile(
		c.Context(),
		claims.UserID,
	)
	if err != nil {
		return fiber.NewError(403, "lecturer profile not found")
	}

	student, err := s.studentRepo.GetStudentByID(c.Context(), studentID)
	if err != nil {
		return fiber.NewError(404, "student not found")
	}

	if student.AdvisorID != lecturer.ID {
		return fiber.NewError(403, "not your advisee")
	}

	return c.JSON(student)
}
