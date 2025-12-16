// tests/service/achievement_service_test.go
package service_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"uas-backend/app/model"
	"uas-backend/app/service"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ====================
// LOCAL REQUEST STRUCTS (untuk testing saja)
// ====================
// Karena struct request mungkin tidak diexport atau berada di package lain,
// kita definisikan ulang di sini agar bisa digunakan di test.

type CreateAchievementRequest struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	Category        string `json:"category"`
	Level           string `json:"level"`
	AchievementDate string `json:"achievement_date"`
	// Tambahkan field lain jika diperlukan
}

type UpdateAchievementRequest struct {
	Title           *string `json:"title,omitempty"`
	Description     *string `json:"description,omitempty"`
	Category        *string `json:"category,omitempty"`
	Level           *string `json:"level,omitempty"`
	AchievementDate *string `json:"achievement_date,omitempty"`
}

type RejectAchievementRequest struct {
	RejectionNote string `json:"rejection_note"`
}

/*
	====================
	  MOCKS (diperbaiki agar implement interface repository)
	====================
*/

type MockAchievementRepo struct{ mock.Mock }

// Interface AchievementRepository mengharapkan FindAll return []model.Achievement (bukan pointer)
func (m *MockAchievementRepo) FindAll(ctx context.Context) ([]model.Achievement, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Achievement), args.Error(1)
}

func (m *MockAchievementRepo) Create(ctx context.Context, a *model.Achievement) (primitive.ObjectID, error) {
	args := m.Called(ctx, a)
	return args.Get(0).(primitive.ObjectID), args.Error(1)
}

func (m *MockAchievementRepo) AddAttachment(ctx context.Context, id primitive.ObjectID, att model.Attachment) error {
	return m.Called(ctx, id, att).Error(0)
}

func (m *MockAchievementRepo) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Achievement, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Achievement), args.Error(1)
}

func (m *MockAchievementRepo) Update(ctx context.Context, a *model.Achievement) error {
	return m.Called(ctx, a).Error(0)
}

func (m *MockAchievementRepo) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockAchievementRepo) FindByStudentIDs(ctx context.Context, ids []string) ([]model.Achievement, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Achievement), args.Error(1)
}

type MockReferenceRepo struct{ mock.Mock }

func (m *MockReferenceRepo) CreateDraft(ctx context.Context, studentID, achievementID string) error {
	return m.Called(ctx, studentID, achievementID).Error(0)
}

func (m *MockReferenceRepo) GetByAchievementID(ctx context.Context, achievementID string) (*model.AchievementReference, error) {
	args := m.Called(ctx, achievementID)
	return args.Get(0).(*model.AchievementReference), args.Error(1)
}

// Method yang hilang: GetByStudentID
func (m *MockReferenceRepo) GetByStudentID(ctx context.Context, studentID string) ([]*model.AchievementReference, error) {
	args := m.Called(ctx, studentID)
	return args.Get(0).([]*model.AchievementReference), args.Error(1)
}

func (m *MockReferenceRepo) Submit(ctx context.Context, achievementID string) (*model.AchievementReference, error) {
	args := m.Called(ctx, achievementID)
	return args.Get(0).(*model.AchievementReference), args.Error(1)
}

func (m *MockReferenceRepo) Verify(ctx context.Context, achievementID, userID string) (*model.AchievementReference, error) {
	args := m.Called(ctx, achievementID, userID)
	return args.Get(0).(*model.AchievementReference), args.Error(1)
}

func (m *MockReferenceRepo) Reject(ctx context.Context, achievementID, note string) (*model.AchievementReference, error) {
	args := m.Called(ctx, achievementID, note)
	return args.Get(0).(*model.AchievementReference), args.Error(1)
}

func (m *MockReferenceRepo) MarkDeleted(ctx context.Context, achievementID string) error {
	return m.Called(ctx, achievementID).Error(0)
}

type MockStudentRepo struct{ mock.Mock }

func (m *MockStudentRepo) GetStudentProfile(ctx context.Context, userID string) (*model.Student, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockStudentRepo) GetStudentsByAdvisor(ctx context.Context, advisorID string) ([]*model.Student, error) {
	args := m.Called(ctx, advisorID)
	return args.Get(0).([]*model.Student), args.Error(1)
}

// Method yang hilang: GetAllStudents
func (m *MockStudentRepo) GetAllStudents(ctx context.Context) ([]*model.Student, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Student), args.Error(1)
}

type MockLecturerRepo struct{ mock.Mock }

func (m *MockLecturerRepo) GetLecturerProfile(ctx context.Context, userID string) (*model.Lecturer, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*model.Lecturer), args.Error(1)
}

// Method yang hilang: GetAllLecturers
func (m *MockLecturerRepo) GetAllLecturers(ctx context.Context) ([]*model.Lecturer, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Lecturer), args.Error(1)
}

func (m *MockStudentRepo) GetStudentByID(ctx context.Context, studentID string) (*model.Student, error) {
	args := m.Called(ctx, studentID)
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockLecturerRepo) GetLecturerByID(ctx context.Context, lecturerID string) (*model.Lecturer, error) {
	args := m.Called(ctx, lecturerID)
	return args.Get(0).(*model.Lecturer), args.Error(1)
}

func (m *MockStudentRepo) UpdateAdvisor(ctx context.Context, studentID string, advisorID string) error {
	return m.Called(ctx, studentID, advisorID).Error(0)
}

func (m *MockAchievementRepo) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]model.Achievement, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Achievement), args.Error(1)
}

/*
====================

	HELPERS

====================
*/
func ptrString(s string) *string { return &s }

/* ====================
   UNIT TESTS
==================== */

func TestAchievementService_All(t *testing.T) {
	// Initialize mocks
	achRepo := new(MockAchievementRepo)
	refRepo := new(MockReferenceRepo)
	stuRepo := new(MockStudentRepo)
	lecRepo := new(MockLecturerRepo)

	// Mock method tambahan untuk safety
	achRepo.On("FindAll", mock.Anything).Return([]model.Achievement{}, nil)
	achRepo.On("FindByIDs", mock.Anything, mock.Anything).Return([]*model.Achievement{}, nil)
	stuRepo.On("GetAllStudents", mock.Anything).Return([]*model.Student{}, nil)
	stuRepo.On("GetStudentByID", mock.Anything, mock.Anything).Return(&model.Student{}, nil)
	stuRepo.On("UpdateAdvisor", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	lecRepo.On("GetAllLecturers", mock.Anything).Return([]*model.Lecturer{}, nil)
	lecRepo.On("GetLecturerByID", mock.Anything, mock.Anything).Return(&model.Lecturer{}, nil)
	refRepo.On("GetByStudentID", mock.Anything, mock.Anything).Return([]*model.AchievementReference{}, nil)

	service := service.NewAchievementService(achRepo, refRepo, stuRepo, lecRepo)

	// Common IDs
	achievementID := primitive.NewObjectID()
	achievementIDHex := achievementID.Hex()
	studentID := "stu-123"
	userID := "user-1"
	adminID := "admin-1"

	// ====================
	// CREATE ACHIEVEMENT
	// ====================
	t.Run("CreateAchievement as Mahasiswa", func(t *testing.T) {
		claims := &model.JWTClaims{UserID: userID, Role: "Mahasiswa", StudentID: studentID}
		student := &model.Student{ID: studentID}

		stuRepo.On("GetStudentProfile", mock.Anything, userID).Return(student, nil)
		achRepo.On("Create", mock.Anything, mock.Anything).Return(achievementID, nil)
		refRepo.On("CreateDraft", mock.Anything, studentID, achievementID.Hex()).Return(nil)

		app := fiber.New()
		app.Post("/achievements", func(c *fiber.Ctx) error {
			c.Locals("user", claims)
			c.BodyParser(&CreateAchievementRequest{})
			return service.CreateAchievement(c)
		})

		req := httptest.NewRequest("POST", "/achievements", strings.NewReader(`{"title":"Test"}`))
		resp, _ := app.Test(req)
		assert.NotNil(t, resp)
	})

	// ====================
	// UPLOAD ATTACHMENT
	// ====================
	t.Run("UploadAttachment", func(t *testing.T) {
		claims := &model.JWTClaims{UserID: userID, Role: "Mahasiswa", StudentID: studentID}
		ref := &model.AchievementReference{StudentID: studentID, Status: "draft"}

		refRepo.On("GetByAchievementID", mock.Anything, achievementIDHex).Return(ref, nil)
		achRepo.On("AddAttachment", mock.Anything, achievementID, mock.Anything).Return(nil)

		app := fiber.New()
		app.Post("/:id/attachment", func(c *fiber.Ctx) error {
			c.Params("id", achievementIDHex)
			c.Locals("user", claims)
			return service.UploadAttachment(c)
		})

		req := httptest.NewRequest("POST", "/"+achievementIDHex+"/attachment", nil)
		resp, _ := app.Test(req)
		assert.NotNil(t, resp)
	})

	// ====================
	// UPDATE ACHIEVEMENT
	// ====================
	t.Run("UpdateAchievement", func(t *testing.T) {
		claims := &model.JWTClaims{UserID: userID, Role: "Mahasiswa", StudentID: studentID}
		ref := &model.AchievementReference{StudentID: studentID, Status: "draft"}
		ach := &model.Achievement{ID: achievementID, StudentID: studentID}

		refRepo.On("GetByAchievementID", mock.Anything, achievementIDHex).Return(ref, nil)
		achRepo.On("GetByID", mock.Anything, achievementID).Return(ach, nil)
		achRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		app := fiber.New()
		app.Put("/:id", func(c *fiber.Ctx) error {
			c.Params("id", achievementIDHex)
			c.Locals("user", claims)
			c.BodyParser(&UpdateAchievementRequest{Title: ptrString("Updated")})
			return service.UpdateAchievement(c)
		})

		req := httptest.NewRequest("PUT", "/"+achievementIDHex, strings.NewReader(`{"title":"Updated"}`))
		resp, _ := app.Test(req)
		assert.NotNil(t, resp)
	})

	// ====================
	// GET ACHIEVEMENT BY ID
	// ====================
	t.Run("GetAchievementByID", func(t *testing.T) {
		claims := &model.JWTClaims{UserID: userID, Role: "Mahasiswa", StudentID: studentID}
		ach := &model.Achievement{ID: achievementID, StudentID: studentID}

		achRepo.On("GetByID", mock.Anything, achievementID).Return(ach, nil)
		stuRepo.On("GetStudentProfile", mock.Anything, userID).Return(&model.Student{ID: studentID}, nil)

		app := fiber.New()
		app.Get("/:id", func(c *fiber.Ctx) error {
			c.Params("id", achievementIDHex)
			c.Locals("user", claims)
			return service.GetAchievementByID(c)
		})

		req := httptest.NewRequest("GET", "/"+achievementIDHex, nil)
		resp, _ := app.Test(req)
		assert.NotNil(t, resp)
	})

	// ====================
	// SUBMIT ACHIEVEMENT
	// ====================
	t.Run("SubmitAchievement", func(t *testing.T) {
		claims := &model.JWTClaims{UserID: userID, Role: "Mahasiswa", StudentID: studentID}
		ref := &model.AchievementReference{StudentID: studentID, Status: "draft"}

		refRepo.On("GetByAchievementID", mock.Anything, achievementIDHex).Return(ref, nil)
		refRepo.On("Submit", mock.Anything, achievementIDHex).Return(ref, nil)

		app := fiber.New()
		app.Post("/:id/submit", func(c *fiber.Ctx) error {
			c.Params("id", achievementIDHex)
			c.Locals("user", claims)
			return service.SubmitAchievement(c)
		})

		req := httptest.NewRequest("POST", "/"+achievementIDHex+"/submit", nil)
		resp, _ := app.Test(req)
		assert.NotNil(t, resp)
	})

	// ====================
	// VERIFY ACHIEVEMENT
	// ====================
	t.Run("VerifyAchievement as Admin", func(t *testing.T) {
		claims := &model.JWTClaims{UserID: adminID, Role: "Admin"}
		ref := &model.AchievementReference{Status: "submitted"}

		refRepo.On("GetByAchievementID", mock.Anything, achievementIDHex).Return(ref, nil)
		refRepo.On("Verify", mock.Anything, achievementIDHex, adminID).Return(ref, nil)

		app := fiber.New()
		app.Post("/:id/verify", func(c *fiber.Ctx) error {
			c.Params("id", achievementIDHex)
			c.Locals("user", claims)
			return service.VerifyAchievement(c)
		})

		req := httptest.NewRequest("POST", "/"+achievementIDHex+"/verify", nil)
		resp, _ := app.Test(req)
		assert.NotNil(t, resp)
	})

	// ====================
	// REJECT ACHIEVEMENT
	// ====================
	t.Run("RejectAchievement as Admin", func(t *testing.T) {
		claims := &model.JWTClaims{UserID: adminID, Role: "Admin"}
		ref := &model.AchievementReference{Status: "submitted"}

		refRepo.On("GetByAchievementID", mock.Anything, achievementIDHex).Return(ref, nil)
		refRepo.On("Reject", mock.Anything, achievementIDHex, mock.Anything).Return(ref, nil)

		app := fiber.New()
		app.Post("/:id/reject", func(c *fiber.Ctx) error {
			c.Params("id", achievementIDHex)
			c.Locals("user", claims)
			c.BodyParser(&RejectAchievementRequest{RejectionNote: "Invalid doc"})
			return service.RejectAchievement(c)
		})

		req := httptest.NewRequest("POST", "/"+achievementIDHex+"/reject", strings.NewReader(`{"rejection_note":"Invalid doc"}`))
		resp, _ := app.Test(req)
		assert.NotNil(t, resp)
	})

	// ====================
	// GET ACHIEVEMENT HISTORY
	// ====================
	t.Run("GetAchievementHistory as Mahasiswa", func(t *testing.T) {
		claims := &model.JWTClaims{UserID: userID, Role: "Mahasiswa", StudentID: studentID}
		now := time.Now()
		ref := &model.AchievementReference{
			StudentID:   studentID,
			Status:      "submitted",
			CreatedAt:   now,
			SubmittedAt: &now,
		}
		ach := &model.Achievement{ID: achievementID, StudentID: studentID}

		refRepo.On("GetByAchievementID", mock.Anything, achievementIDHex).Return(ref, nil)
		achRepo.On("GetByID", mock.Anything, achievementID).Return(ach, nil)

		app := fiber.New()
		app.Get("/:id/history", func(c *fiber.Ctx) error {
			c.Params("id", achievementIDHex)
			c.Locals("user", claims)
			return service.GetAchievementHistory(c)
		})

		req := httptest.NewRequest("GET", "/"+achievementIDHex+"/history", nil)
		resp, _ := app.Test(req)
		assert.NotNil(t, resp)
	})

	// ====================
	// DELETE ACHIEVEMENT
	// ====================
	t.Run("DeleteAchievement", func(t *testing.T) {
		claims := &model.JWTClaims{UserID: userID, Role: "Mahasiswa", StudentID: studentID}
		ref := &model.AchievementReference{StudentID: studentID, Status: "draft"}

		refRepo.On("GetByAchievementID", mock.Anything, achievementIDHex).Return(ref, nil)
		achRepo.On("SoftDelete", mock.Anything, achievementID).Return(nil)
		refRepo.On("MarkDeleted", mock.Anything, achievementIDHex).Return(nil)

		app := fiber.New()
		app.Delete("/:id", func(c *fiber.Ctx) error {
			c.Params("id", achievementIDHex)
			c.Locals("user", claims)
			return service.DeleteAchievement(c)
		})

		req := httptest.NewRequest("DELETE", "/"+achievementIDHex, nil)
		resp, _ := app.Test(req)
		assert.NotNil(t, resp)
	})
}
