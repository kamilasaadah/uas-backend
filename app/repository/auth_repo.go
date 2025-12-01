package repository

import (
	"context"
	"errors"

	"uas-backend/database"

	"github.com/jackc/pgx/v5"
)

type AuthRepository interface {
	FindUserByUsernameOrEmail(ctx context.Context, identifier string) (*AuthUserRaw, error)
	GetPermissionsByRole(ctx context.Context, roleID string) ([]string, error)
}

type authRepository struct{}

func NewAuthRepository() AuthRepository {
	return &authRepository{}
}

type AuthUserRaw struct {
	ID       string
	Username string
	Email    string
	Password string
	FullName string
	RoleID   string
	RoleName string
	IsActive bool
}

func (r *authRepository) FindUserByUsernameOrEmail(ctx context.Context, identifier string) (*AuthUserRaw, error) {

	query := `
		SELECT 
			u.id, u.username, u.email, u.password_hash,
			u.full_name, u.role_id, r.name, u.is_active
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.username = $1 OR u.email = $1
		LIMIT 1;
	`

	row := database.PG.QueryRow(ctx, query, identifier)

	var user AuthUserRaw

	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.FullName, &user.RoleID, &user.RoleName, &user.IsActive,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("invalid credentials")
	}

	return &user, err
}

func (r *authRepository) GetPermissionsByRole(ctx context.Context, roleID string) ([]string, error) {

	query := `
		SELECT p.name
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		WHERE rp.role_id = $1;
	`

	rows, err := database.PG.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := []string{}
	var perm string

	for rows.Next() {
		if err := rows.Scan(&perm); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}
