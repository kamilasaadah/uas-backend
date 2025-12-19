package service_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"uas-backend/app/model"
	"uas-backend/app/service"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

/*
=================================================
MOCK REPOSITORIES
=================================================
*/

type mockLecturerRepository struct {
	mock.Mock
}

func (m *mockLecturerRepository) GetAllLecturers(ctx context.Context) ([]*model.Lecturer, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Lecturer), args.Error(1)
}

func (m *mockLecturerRepository) GetLecturerProfile(ctx context.Context, userID string) (*model.Lecturer, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Lecturer), args.Error(1)
}

func (m *mockLecturerRepository) GetLecturerByID(ctx context.Context, lecturerID string) (*model.Lecturer, error) {
	args := m.Called(ctx, lecturerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Lecturer), args.Error(1)
}

type mockStudentRepository struct {
	mock.Mock
}

func (m *mockStudentRepository) GetAllStudents(ctx context.Context) ([]*model.Student, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Student), args.Error(1)
}

func (m *mockStudentRepository) GetStudentsByAdvisor(ctx context.Context, advisorID string) ([]*model.Student, error) {
	args := m.Called(ctx, advisorID)
	return args.Get(0).([]*model.Student), args.Error(1)
}

func (m *mockStudentRepository) GetStudentProfile(ctx context.Context, userID string) (*model.Student, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *mockStudentRepository) UpdateAdvisor(ctx context.Context, studentID string, advisorID string) error {
	args := m.Called(ctx, studentID, advisorID)
	return args.Error(0)
}

func (m *mockStudentRepository) GetStudentByID(ctx context.Context, studentID string) (*model.Student, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Student), args.Error(1)
}

/*
=================================================
TEST: GET ALL LECTURERS
=================================================
*/

func TestLecturerService_GetAllLecturers(t *testing.T) {
	mockLecturerRepo := new(mockLecturerRepository)
	mockStudentRepo := new(mockStudentRepository)

	svc := service.NewLecturerService(mockLecturerRepo, mockStudentRepo)

	tests := []struct {
		name           string
		role           string
		setupMock      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success - Admin",
			role: "Admin",
			setupMock: func() {
				mockLecturerRepo.On("GetAllLecturers", mock.Anything).Return([]*model.Lecturer{
					{ID: "1", LecturerID: "L001"},
					{ID: "2", LecturerID: "L002"},
				}, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedBody: []model.Lecturer{
				{ID: "1", LecturerID: "L001"},
				{ID: "2", LecturerID: "L002"},
			},
		},
		{
			name:           "Forbidden - Non Admin",
			role:           "Dosen Wali",
			setupMock:      func() {},
			expectedStatus: fiber.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Get("/lecturers", func(c *fiber.Ctx) error {
				c.Locals("user", &model.JWTClaims{
					UserID: "u-1",
					Role:   tt.role,
				})
				return svc.GetAllLecturers(c)
			})

			tt.setupMock()

			req := httptest.NewRequest("GET", "/lecturers", nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == fiber.StatusOK {
				var body []model.Lecturer
				_ = json.NewDecoder(resp.Body).Decode(&body)
				assert.Equal(t, tt.expectedBody, body)
			}

			mockLecturerRepo.AssertExpectations(t)
			mockStudentRepo.AssertExpectations(t)
		})
	}
}

/*
=================================================
TEST: GET ADVISEES
=================================================
*/

func TestLecturerService_GetAdvisees(t *testing.T) {
	mockLecturerRepo := new(mockLecturerRepository)
	mockStudentRepo := new(mockStudentRepository)

	svc := service.NewLecturerService(mockLecturerRepo, mockStudentRepo)

	tests := []struct {
		name           string
		role           string
		userID         string
		lecturerIDPath string
		setupMock      func()
		expectedStatus int
	}{
		{
			name:           "Admin gets all students",
			role:           "Admin",
			userID:         "admin-1",
			lecturerIDPath: "123",
			setupMock: func() {
				mockStudentRepo.On("GetAllStudents", mock.Anything).Return([]*model.Student{
					{ID: "s1"},
					{ID: "s2"},
				}, nil)
			},
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "Dosen Wali gets own advisees",
			role:           "Dosen Wali",
			userID:         "lecturer-user-1",
			lecturerIDPath: "lecturer-123",
			setupMock: func() {
				mockLecturerRepo.
					On("GetLecturerProfile", mock.Anything, "lecturer-user-1").
					Return(&model.Lecturer{ID: "lecturer-123"}, nil)

				mockStudentRepo.
					On("GetStudentsByAdvisor", mock.Anything, "lecturer-123").
					Return([]*model.Student{{ID: "s3"}}, nil)
			},
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "Forbidden - wrong lecturer",
			role:           "Dosen Wali",
			userID:         "lecturer-user-1",
			lecturerIDPath: "lecturer-999",
			setupMock: func() {
				mockLecturerRepo.
					On("GetLecturerProfile", mock.Anything, "lecturer-user-1").
					Return(&model.Lecturer{ID: "lecturer-123"}, nil)
			},
			expectedStatus: fiber.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Get("/lecturers/:id/advisees", func(c *fiber.Ctx) error {
				c.Locals("user", &model.JWTClaims{
					UserID: tt.userID,
					Role:   tt.role,
				})
				return svc.GetAdvisees(c)
			})

			tt.setupMock()

			req := httptest.NewRequest(
				"GET",
				"/lecturers/"+tt.lecturerIDPath+"/advisees",
				nil,
			)

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			mockLecturerRepo.AssertExpectations(t)
			mockStudentRepo.AssertExpectations(t)
		})
	}
}
