// tests/service/user_service_test.go
package service_test

import (
	"context"
	"errors"
	"testing"

	"uas-backend/app/model"
	"uas-backend/app/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

////////////////////////////////////////////////////////////////////////////////
// MOCKS (KHUSUS USER SERVICE — NAMA UNIK)
////////////////////////////////////////////////////////////////////////////////

type MockUserRepoUserSvc struct {
	mock.Mock
}

func (m *MockUserRepoUserSvc) FindByUsernameOrEmail(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepoUserSvc) GetUserPermissions(userID string) ([]string, error) {
	args := m.Called(userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockUserRepoUserSvc) CheckDuplicate(username, email string) error {
	args := m.Called(username, email)
	return args.Error(0)
}

func (m *MockUserRepoUserSvc) CreateUser(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepoUserSvc) UpdateUser(ctx context.Context, id string, req *model.UpdateUserRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockUserRepoUserSvc) AssignRole(ctx context.Context, id string, roleID string) error {
	args := m.Called(ctx, id, roleID)
	return args.Error(0)
}

func (m *MockUserRepoUserSvc) UpsertStudentProfile(ctx context.Context, userID string, s *model.StudentProfileRequest) error {
	args := m.Called(ctx, userID, s)
	return args.Error(0)
}

func (m *MockUserRepoUserSvc) UpsertLecturerProfile(ctx context.Context, userID string, l *model.LecturerProfileRequest) error {
	args := m.Called(ctx, userID, l)
	return args.Error(0)
}

func (m *MockUserRepoUserSvc) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *MockUserRepoUserSvc) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepoUserSvc) SoftDeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ---------------------------------------------------------------------------

type MockStudentRepoUserSvc struct {
	mock.Mock
}

func (m *MockStudentRepoUserSvc) GetStudentProfile(ctx context.Context, userID string) (*model.Student, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockStudentRepoUserSvc) UpdateAdvisor(ctx context.Context, studentID string, advisorID string) error {
	args := m.Called(ctx, studentID, advisorID)
	return args.Error(0)
}

func (m *MockStudentRepoUserSvc) GetAllStudents(ctx context.Context) ([]*model.Student, error) {
	panic("not used")
}

func (m *MockStudentRepoUserSvc) GetStudentsByAdvisor(ctx context.Context, advisorID string) ([]*model.Student, error) {
	panic("not used")
}

func (m *MockStudentRepoUserSvc) GetStudentByID(ctx context.Context, studentID string) (*model.Student, error) {
	panic("not used")
}

// ---------------------------------------------------------------------------

type MockLecturerRepoUserSvc struct {
	mock.Mock
}

func (m *MockLecturerRepoUserSvc) GetLecturerProfile(ctx context.Context, userID string) (*model.Lecturer, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Lecturer), args.Error(1)
}

func (m *MockLecturerRepoUserSvc) GetAllLecturers(ctx context.Context) ([]*model.Lecturer, error) {
	panic("not used")
}

func (m *MockLecturerRepoUserSvc) GetLecturerByID(ctx context.Context, lecturerID string) (*model.Lecturer, error) {
	panic("not used")
}

////////////////////////////////////////////////////////////////////////////////
// TESTS
////////////////////////////////////////////////////////////////////////////////

func TestCreateUser_Success(t *testing.T) {
	userRepo := new(MockUserRepoUserSvc)
	studentRepo := new(MockStudentRepoUserSvc)
	lecturerRepo := new(MockLecturerRepoUserSvc)

	svc := service.NewUserService(userRepo, studentRepo, lecturerRepo)

	req := model.CreateUserRequest{
		Username: "john",
		Email:    "john@mail.com",
		Password: "secret",
		FullName: "John Doe",
	}

	userRepo.On("CheckDuplicate", req.Username, req.Email).Return(nil)
	userRepo.
		On("CreateUser", mock.Anything, mock.AnythingOfType("*model.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			u := args.Get(1).(*model.User)
			u.ID = "user-1"
		})

	user, err := svc.(*service.UserService).CreateUser(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "john", user.Username)
	assert.True(t, user.IsActive)
	userRepo.AssertExpectations(t)
}

func TestCreateUser_Duplicate(t *testing.T) {
	userRepo := new(MockUserRepoUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	req := model.CreateUserRequest{
		Username: "john",
		Email:    "john@mail.com",
	}

	userRepo.On("CheckDuplicate", req.Username, req.Email).
		Return(errors.New("duplicate"))

	user, err := svc.(*service.UserService).CreateUser(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestAssignRole_EmptyRole(t *testing.T) {
	svc := service.NewUserService(nil, nil, nil)

	err := svc.(*service.UserService).AssignRoleLogic(context.Background(), "1", "")

	assert.Error(t, err)
}

func TestAssignRole_Success(t *testing.T) {
	userRepo := new(MockUserRepoUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	userRepo.On("AssignRole", mock.Anything, "1", "role-1").Return(nil)

	err := svc.(*service.UserService).AssignRoleLogic(context.Background(), "1", "role-1")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestGetUserByID_WithProfiles(t *testing.T) {
	userRepo := new(MockUserRepoUserSvc)
	studentRepo := new(MockStudentRepoUserSvc)
	lecturerRepo := new(MockLecturerRepoUserSvc)

	svc := service.NewUserService(userRepo, studentRepo, lecturerRepo)

	userRepo.On("GetUserByID", mock.Anything, "1").
		Return(&model.User{
			ID:       "1",
			Username: "john",
			Email:    "john@mail.com",
			FullName: "John",
		}, nil)

	studentRepo.On("GetStudentProfile", mock.Anything, "1").
		Return(&model.Student{StudentID: "S001"}, nil)

	lecturerRepo.On("GetLecturerProfile", mock.Anything, "1").
		Return(nil, errors.New("not lecturer"))

	resp, err := svc.(*service.UserService).GetUserByID(context.Background(), "1")

	assert.NoError(t, err)
	assert.Equal(t, "john", resp.Username)
	assert.NotNil(t, resp.Student)
}

func TestDeleteUser_Success(t *testing.T) {
	userRepo := new(MockUserRepoUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	userRepo.On("SoftDeleteUser", mock.Anything, "1").Return(nil)

	err := svc.(*service.UserService).DeleteUser(context.Background(), "1")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}
