package repository

import (
	"context"
	"errors"

	"uas-backend/app/model"
	"uas-backend/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {

	// AUTH / LOGIN
	FindByUsernameOrEmail(ctx context.Context, username string) (*model.User, error)
	GetUserPermissions(userID string) ([]string, error)

	// ADMIN CRUD
	CheckDuplicate(username, email string) error
	CreateUser(ctx context.Context, user *model.User) error

	UpdateUser(ctx context.Context, id string, req *model.UpdateUserRequest) error
	AssignRole(ctx context.Context, id string, roleID string) error

	UpsertStudentProfile(ctx context.Context, userID string, s *model.StudentProfileRequest) error
	UpsertLecturerProfile(ctx context.Context, userID string, l *model.LecturerProfileRequest) error

	GetAllUsers(ctx context.Context) ([]*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db}
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

///////////////////////////////////////////////////////////////////////////////
// ======================= ADMIN: CHECK DUPLICATE =======================
///////////////////////////////////////////////////////////////////////////////

func (r *userRepository) CheckDuplicate(username, email string) error {
	var exists bool

	err := database.PG.QueryRow(
		context.Background(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE username=$1 OR email=$2)`,
		username, email,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if exists {
		return errors.New("username or email already exists")
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// ======================= ADMIN: CREATE USER =======================
///////////////////////////////////////////////////////////////////////////////

func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {

	query := `
		INSERT INTO users (username, email, password_hash, full_name, role_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	return database.PG.QueryRow(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.RoleID,
	).Scan(&user.ID)
}

///////////////////////////////////////////////////////////////////////////////
// ======================= ADMIN: UPDATE USER =======================
///////////////////////////////////////////////////////////////////////////////

func (r *userRepository) UpdateUser(ctx context.Context, id string, req *model.UpdateUserRequest) error {

	_, err := database.PG.Exec(ctx,
		`UPDATE users 
		 SET username = COALESCE(NULLIF($1, ''), username),
			 email = COALESCE(NULLIF($2, ''), email),
			 full_name = COALESCE(NULLIF($3, ''), full_name),
			 updated_at = NOW()
		 WHERE id = $4`,
		req.Username, req.Email, req.FullName, id,
	)

	return err
}

///////////////////////////////////////////////////////////////////////////////
// ======================= ADMIN: ASSIGN ROLE =======================
///////////////////////////////////////////////////////////////////////////////

func (r *userRepository) AssignRole(ctx context.Context, id string, roleID string) error {

	_, err := database.PG.Exec(ctx,
		`UPDATE users SET role_id = $1 WHERE id = $2`,
		roleID, id,
	)

	return err
}

///////////////////////////////////////////////////////////////////////////////
// ======================= ADMIN: UPSERT STUDENT PROFILE =======================
///////////////////////////////////////////////////////////////////////////////

func (r *userRepository) UpsertStudentProfile(ctx context.Context, userID string, s *model.StudentProfileRequest) error {

	// 1. cek apakah student profile sudah ada
	var exists bool
	err := database.PG.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM students WHERE user_id = $1)`,
		userID,
	).Scan(&exists)
	if err != nil {
		return err
	}

	// 2. UPDATE jika sudah ada
	if exists {
		_, err := database.PG.Exec(ctx,
			`UPDATE students 
			 SET student_id=$1, program_study=$2, academic_year=$3, advisor_id=$4
			 WHERE user_id=$5`,
			s.StudentID, s.ProgramStudy, s.AcademicYear, s.AdvisorID, userID,
		)
		return err
	}

	// 3. INSERT jika belum ada
	_, err = database.PG.Exec(ctx,
		`INSERT INTO students (user_id, student_id, program_study, academic_year, advisor_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, s.StudentID, s.ProgramStudy, s.AcademicYear, s.AdvisorID,
	)

	return err
}

///////////////////////////////////////////////////////////////////////////////
// ======================= ADMIN: UPSERT LECTURER PROFILE =======================
///////////////////////////////////////////////////////////////////////////////

func (r *userRepository) UpsertLecturerProfile(ctx context.Context, userID string, l *model.LecturerProfileRequest) error {

	// cek apakah lecturer profile sudah ada
	var exists bool
	err := database.PG.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM lecturers WHERE user_id=$1)`,
		userID,
	).Scan(&exists)
	if err != nil {
		return err
	}

	// UPDATE jika sudah ada
	if exists {
		_, err := database.PG.Exec(ctx,
			`UPDATE lecturers
			 SET lecturer_id=$1, department=$2
			 WHERE user_id=$3`,
			l.LecturerID, l.Department, userID,
		)
		return err
	}

	// INSERT jika belum ada
	_, err = database.PG.Exec(ctx,
		`INSERT INTO lecturers (user_id, lecturer_id, department)
		 VALUES ($1, $2, $3)`,
		userID, l.LecturerID, l.Department,
	)

	return err
}

// /////////////////////////////////////////////////////////////////////////////
// ======================= ADMIN: GET ALL USERS =======================
// /////////////////////////////////////////////////////////////////////////////
func (r *userRepository) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	sql := `
        SELECT u.id, u.username, u.email, u.full_name,
               u.role_id, r.name AS role_name,
               u.is_active
        FROM users u
        LEFT JOIN roles r ON r.id = u.role_id
        ORDER BY u.created_at DESC
    `
	rows, err := r.db.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User

	for rows.Next() {
		u := &model.User{}
		err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.FullName,
			&u.RoleID,
			&u.RoleName,
			&u.IsActive,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, nil
}

///////////////////////////////////////////////////////////////////////////////
// ======================= ADMIN: GET USER BY ID =======================
///////////////////////////////////////////////////////////////////////////////

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	sql := `
        SELECT u.id, u.username, u.email, u.full_name,
               u.role_id, r.name AS role_name,
               u.is_active
        FROM users u
        LEFT JOIN roles r ON r.id = u.role_id
        WHERE u.id = $1
    `
	u := &model.User{}

	err := r.db.QueryRow(ctx, sql, id).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.FullName,
		&u.RoleID,
		&u.RoleName,
		&u.IsActive,
	)
	if err != nil {
		return nil, err
	}

	return u, nil
}
