package repository

import (
	"context"
	"errors"

	"uas-backend/app/model"
	"uas-backend/database"
)

type UserRepository interface {
	FindByUsernameOrEmail(ctx context.Context, username string) (*model.User, error)
	GetUserPermissions(userID string) ([]string, error)
}

type userRepository struct{}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

// ======================= GET USER BY USERNAME OR EMAIL =======================

func (r *userRepository) FindByUsernameOrEmail(ctx context.Context, username string) (*model.User, error) {
	query := `
	SELECT u.id, u.username, u.email, u.password_hash, 
       u.full_name, u.role_id, r.name AS role_name, u.is_active
	FROM users u
	JOIN roles r ON r.id = u.role_id
	WHERE u.username = $1 OR u.email = $1`

	row := database.PG.QueryRow(ctx, query, username)

	var user model.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.RoleID,
		&user.RoleName,
		&user.IsActive,
	)

	if err != nil {
		return nil, errors.New("user not found")
	}

	return &user, nil
}

// ======================= GET USER PERMISSIONS =======================

func (r *userRepository) GetUserPermissions(userID string) ([]string, error) {

	query := `
	SELECT p.name
	FROM users u
	JOIN roles r ON r.id = u.role_id
	JOIN role_permissions rp ON r.id = rp.role_id
	JOIN permissions p ON p.id = rp.permission_id
	WHERE u.id = $1
	`

	rows, err := database.PG.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var perm string
		rows.Scan(&perm)
		perms = append(perms, perm)
	}

	return perms, nil
}
