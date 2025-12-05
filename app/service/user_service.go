package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"uas-backend/app/model"
	"uas-backend/app/repository"
)

type UserService struct {
	repo         repository.UserRepository
	studentRepo  repository.StudentRepository
	lecturerRepo repository.LecturerRepository
}

func NewUserService(
	repo repository.UserRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
) *UserService {
	return &UserService{
		repo:         repo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo}
}

func (s *UserService) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {

	// ================================
	// 1. CHECK DUPLICATE USERNAME
	// ================================
	if err := s.repo.CheckDuplicate(req.Username, req.Email); err != nil {
		return nil, err
	}
	// 2. HASH PASSWORD
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 3. CREATE USER OBJECT
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		FullName:     req.FullName,
		PasswordHash: string(hashed),
		RoleID:       req.RoleID,
		IsActive:     true,
	}

	// 4. SAVE USER (WITHOUT PROFILE)
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

////////////////////////////////////////////////////////////////////////////////
// ======================= ADMIN: UPDATE USER (+ PROFILES) ====================
////////////////////////////////////////////////////////////////////////////////

func (s *UserService) UpdateUser(ctx context.Context, id string, req model.UpdateUserRequest) error {

	// 1. UPDATE USER BASIC DATA
	if err := s.repo.UpdateUser(ctx, id, &req); err != nil {
		return err
	}

	// 2. UPSERT STUDENT PROFILE (optional)
	if req.StudentProfile != nil {
		if err := s.repo.UpsertStudentProfile(ctx, id, req.StudentProfile); err != nil {
			return err
		}
	}

	// 3. UPSERT LECTURER PROFILE (optional)
	if req.LecturerProfile != nil {
		if err := s.repo.UpsertLecturerProfile(ctx, id, req.LecturerProfile); err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// ======================= ADMIN: ASSIGN ROLE ================================
////////////////////////////////////////////////////////////////////////////////

func (s *UserService) AssignRole(ctx context.Context, id string, roleID string) error {
	if roleID == "" {
		return errors.New("role_id required")
	}

	return s.repo.AssignRole(ctx, id, roleID)
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]*model.UserWithProfileResponse, error) {

	users, err := s.repo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	var resp []*model.UserWithProfileResponse

	for _, u := range users {

		item := &model.UserWithProfileResponse{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
			FullName: u.FullName,
			RoleID:   u.RoleID,
			RoleName: u.RoleName,
			IsActive: u.IsActive,
		}

		// attach student profile
		student, _ := s.studentRepo.GetStudentProfile(ctx, u.ID)
		item.Student = student

		// attach lecturer profile
		lecturer, _ := s.lecturerRepo.GetLecturerProfile(ctx, u.ID)
		item.Lecturer = lecturer

		resp = append(resp, item)
	}

	return resp, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*model.UserWithProfileResponse, error) {

	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := &model.UserWithProfileResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		FullName: u.FullName,
		RoleID:   u.RoleID,
		RoleName: u.RoleName,
		IsActive: u.IsActive,
	}

	student, _ := s.studentRepo.GetStudentProfile(ctx, id)
	resp.Student = student

	lecturer, _ := s.lecturerRepo.GetLecturerProfile(ctx, id)
	resp.Lecturer = lecturer

	return resp, nil
}
