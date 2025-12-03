package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"uas-backend/app/model"
	"uas-backend/app/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo}
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
