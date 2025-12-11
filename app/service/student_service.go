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
