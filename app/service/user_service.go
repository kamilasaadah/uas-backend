package service

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"uas-backend/app/model"
	"uas-backend/app/repository"
)

type UserHttpHandler interface {
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	AssignRole(c *fiber.Ctx) error
	GetAll(c *fiber.Ctx) error
	GetByID(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type UserService struct {
	repo         repository.UserRepository
	studentRepo  repository.StudentRepository
	lecturerRepo repository.LecturerRepository
}

func NewUserService(
	repo repository.UserRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
) UserHttpHandler {
	return &UserService{
		repo:         repo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

////////////////////////////////////////////////////////////////////////////////
// ======================= BUSINESS LOGIC =====================================
////////////////////////////////////////////////////////////////////////////////

func (s *UserService) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {

	if err := s.repo.CheckDuplicate(req.Username, req.Email); err != nil {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		FullName:     req.FullName,
		PasswordHash: string(hashed),
		IsActive:     true,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, req model.UpdateUserRequest) error {

	if err := s.repo.UpdateUser(ctx, id, &req); err != nil {
		return err
	}

	if req.StudentProfile != nil {
		if err := s.repo.UpsertStudentProfile(ctx, id, req.StudentProfile); err != nil {
			return err
		}
	}

	if req.LecturerProfile != nil {
		if err := s.repo.UpsertLecturerProfile(ctx, id, req.LecturerProfile); err != nil {
			return err
		}
	}

	return nil
}

func (s *UserService) AssignRoleLogic(ctx context.Context, id string, roleID string) error {
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

		item.Student, _ = s.studentRepo.GetStudentProfile(ctx, u.ID)
		item.Lecturer, _ = s.lecturerRepo.GetLecturerProfile(ctx, u.ID)

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

	resp.Student, _ = s.studentRepo.GetStudentProfile(ctx, id)
	resp.Lecturer, _ = s.lecturerRepo.GetLecturerProfile(ctx, id)

	return resp, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.SoftDeleteUser(ctx, id)
}

////////////////////////////////////////////////////////////////////////////////
// ======================= FIBER HANDLER WRAPPER ===============================
////////////////////////////////////////////////////////////////////////////////

// Create godoc
// @Summary Create new user
// @Description Admin only. Create new system user (student / lecturer / admin)
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body model.CreateUserRequest true "Create user payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /users [post]
func (s *UserService) Create(c *fiber.Ctx) error {
	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "invalid request"})
	}

	user, err := s.CreateUser(context.Background(), req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "user created", "data": user})
}

// Update godoc
// @Summary Update user
// @Description Admin only. Update user data and profile
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body model.UpdateUserRequest true "Update user payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /users/{id} [put]
func (s *UserService) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "invalid request"})
	}

	if err := s.UpdateUser(context.Background(), id, req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "user updated"})
}

// AssignRole godoc
// @Summary Assign role to user
// @Description Admin only. Assign or change user role
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body model.UpdateUserRoleRequest true "Role payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /users/{id}/role [put]
func (s *UserService) AssignRole(c *fiber.Ctx) error {
	id := c.Params("id")
	var req model.UpdateUserRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "invalid request"})
	}

	if err := s.AssignRoleLogic(context.Background(), id, req.RoleID); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "role updated"})
}

// GetAll godoc
// @Summary Get all users
// @Description Admin only. Retrieve list of all users with profiles
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string][]model.UserWithProfileResponse
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [get]
func (s *UserService) GetAll(c *fiber.Ctx) error {
	users, err := s.GetAllUsers(context.Background())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "failed to get users"})
	}

	return c.JSON(fiber.Map{"data": users})
}

// GetByID godoc
// @Summary Get user by ID
// @Description Admin only. Get detail user with profile
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]model.UserWithProfileResponse
// @Failure 404 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /users/{id} [get]
func (s *UserService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	user, err := s.GetUserByID(context.Background(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "not found"})
	}

	return c.JSON(fiber.Map{"data": user})
}

// Delete godoc
// @Summary Delete user
// @Description Admin only. Soft delete user
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /users/{id} [delete]
func (s *UserService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.DeleteUser(context.Background(), id); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "user deleted"})
}
