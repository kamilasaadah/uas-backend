// tests/service/auth_service_test.go
package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"uas-backend/app/model"
	"uas-backend/app/service"
	"uas-backend/config"
	"uas-backend/middleware" // import untuk ClearBlocklistForTest dan IsJWTBlocked

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// ====================
// MOCKS
// ====================

type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) FindByUsernameOrEmail(ctx context.Context, identifier string) (*model.User, error) {
	args := m.Called(ctx, identifier)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetUserPermissions(userID string) ([]string, error) {
	args := m.Called(userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.User), args.Error(1)
}

// Implementasi minimal untuk interface lengkap
func (m *MockUserRepository) CheckDuplicate(username, email string) error            { return nil }
func (m *MockUserRepository) CreateUser(ctx context.Context, user *model.User) error { return nil }
func (m *MockUserRepository) UpdateUser(ctx context.Context, id string, req *model.UpdateUserRequest) error {
	return nil
}
func (m *MockUserRepository) AssignRole(ctx context.Context, id string, roleID string) error {
	return nil
}
func (m *MockUserRepository) UpsertStudentProfile(ctx context.Context, userID string, s *model.StudentProfileRequest) error {
	return nil
}
func (m *MockUserRepository) UpsertLecturerProfile(ctx context.Context, userID string, l *model.LecturerProfileRequest) error {
	return nil
}
func (m *MockUserRepository) GetAllUsers(ctx context.Context) ([]*model.User, error) { return nil, nil }
func (m *MockUserRepository) SoftDeleteUser(ctx context.Context, id string) error    { return nil }

type MockStudentRepository struct{ mock.Mock }

func (m *MockStudentRepository) GetStudentProfile(ctx context.Context, userID string) (*model.Student, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockStudentRepository) UpdateAdvisor(ctx context.Context, studentID, advisorID string) error {
	return nil
}
func (m *MockStudentRepository) GetAllStudents(ctx context.Context) ([]*model.Student, error) {
	return nil, nil
}
func (m *MockStudentRepository) GetStudentsByAdvisor(ctx context.Context, advisorID string) ([]*model.Student, error) {
	return nil, nil
}
func (m *MockStudentRepository) GetStudentByID(ctx context.Context, studentID string) (*model.Student, error) {
	return nil, nil
}

// ====================
// HELPER: hash password
// ====================
func hashPassword(password string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed)
}

// ====================
// UNIT TESTS
// ====================

func TestAuthService_All(t *testing.T) {
	// Bersihkan blocklist di awal dan akhir setiap test
	t.Cleanup(middleware.ClearBlocklistForTest)
	middleware.ClearBlocklistForTest()

	userRepo := new(MockUserRepository)
	studentRepo := new(MockStudentRepository)

	authService := service.NewAuthService(userRepo, studentRepo)

	commonUserID := "usr-123"
	commonUsername := "johndoe"
	commonPassword := "secret123"
	hashedPass := hashPassword(commonPassword)

	// ====================
	// LOGIN - Success (Admin)
	// ====================
	t.Run("Login Success - Admin", func(t *testing.T) {
		userRepo.On("FindByUsernameOrEmail", mock.Anything, commonUsername).
			Return(&model.User{
				ID:           commonUserID,
				Username:     commonUsername,
				Email:        "john@example.com",
				PasswordHash: hashedPass,
				FullName:     "John Doe",
				RoleID:       "role-admin",
				RoleName:     "Admin",
				IsActive:     true,
			}, nil)

		userRepo.On("GetUserPermissions", commonUserID).
			Return([]string{"manage_users", "view_reports"}, nil)

		app := fiber.New()
		app.Post("/login", authService.Login)

		body := map[string]string{"username": commonUsername, "password": commonPassword}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		assert.Equal(t, 200, resp.StatusCode)

		var respBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respBody)

		data := respBody["data"].(map[string]interface{})
		assert.Contains(t, data, "token")
		assert.Contains(t, data, "refreshToken")
		assert.Equal(t, "Admin", data["user"].(map[string]interface{})["role"])
	})

	// ====================
	// LOGIN - Success (Mahasiswa)
	// ====================
	t.Run("Login Success - Mahasiswa", func(t *testing.T) {
		userRepo.On("FindByUsernameOrEmail", mock.Anything, "student123").
			Return(&model.User{
				ID:           "usr-student",
				Username:     "student123",
				PasswordHash: hashedPass,
				FullName:     "Budi Student",
				RoleName:     "Mahasiswa",
				IsActive:     true,
			}, nil)

		userRepo.On("GetUserPermissions", "usr-student").
			Return([]string{"submit_achievement"}, nil)

		studentRepo.On("GetStudentProfile", mock.Anything, "usr-student").
			Return(&model.Student{ID: "stu-001"}, nil)

		app := fiber.New()
		app.Post("/login", authService.Login)

		body := map[string]string{"username": "student123", "password": commonPassword}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		assert.Equal(t, 200, resp.StatusCode)

		var respBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respBody)
		tokenStr := respBody["data"].(map[string]interface{})["token"].(string)

		token, _ := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.JWTSecret()), nil
		})

		claims := token.Claims.(jwt.MapClaims)
		assert.Equal(t, "stu-001", claims["student_id"])
	})

	// ====================
	// PROFILE - Success
	// ====================
	t.Run("Profile Success", func(t *testing.T) {
		claims := jwt.MapClaims{
			"user_id":     commonUserID,
			"username":    commonUsername,
			"full_name":   "John Doe",
			"role":        "Admin",
			"role_id":     "role-admin",
			"permissions": []string{"all"},
			"exp":         time.Now().Add(time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ := token.SignedString([]byte(config.JWTSecret()))

		app := fiber.New()
		app.Get("/profile", middleware.JWTAuth(userRepo), authService.Profile)

		req := httptest.NewRequest("GET", "/profile", nil)
		req.Header.Set("Authorization", "Bearer "+tokenStr)

		resp, _ := app.Test(req)
		assert.Equal(t, 200, resp.StatusCode)

		var respBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respBody)
		assert.Equal(t, "Profile fetched", respBody["message"])
	})

	// ====================
	// REFRESH TOKEN - Success
	// ====================
	t.Run("Refresh Token Success", func(t *testing.T) {
		refreshClaims := jwt.MapClaims{
			"user_id": commonUserID,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		}
		refreshToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
			SignedString([]byte(config.JWTSecret()))

		userRepo.On("GetUserByID", mock.Anything, commonUserID).
			Return(&model.User{
				ID:       commonUserID,
				Username: commonUsername,
				FullName: "John Doe",
				RoleName: "Admin",
				IsActive: true,
			}, nil)

		userRepo.On("GetUserPermissions", commonUserID).
			Return([]string{"manage_users"}, nil)

		app := fiber.New()
		app.Post("/refresh", authService.Refresh)

		body := map[string]string{"refresh_token": refreshToken}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/refresh", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		assert.Equal(t, 200, resp.StatusCode)
	})

	// ====================
	// LOGOUT - Success
	// ====================
	t.Run("Logout Success", func(t *testing.T) {
		accessClaims := jwt.MapClaims{
			"user_id": commonUserID,
			"exp":     time.Now().Add(1 * time.Hour).Unix(),
		}
		accessToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
			SignedString([]byte(config.JWTSecret()))

		app := fiber.New()
		app.Post("/logout", authService.Logout)

		req := httptest.NewRequest("POST", "/logout", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, _ := app.Test(req)
		assert.Equal(t, 200, resp.StatusCode)

		// Verifikasi token masuk blocklist
		assert.True(t, middleware.IsJWTBlocked(accessToken))
	})

	// ====================
	// LOGOUT - Failed (No Token)
	// ====================
	t.Run("Logout Failed - No Token", func(t *testing.T) {
		app := fiber.New()
		app.Post("/logout", authService.Logout)

		req := httptest.NewRequest("POST", "/logout", nil)
		resp, _ := app.Test(req)
		assert.Equal(t, 401, resp.StatusCode)
	})
}
