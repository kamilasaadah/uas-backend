package repository

import (
	"context"
	"uas-backend/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentRepository interface {
	GetStudentProfile(ctx context.Context, userID string) (*model.Student, error)
	UpdateAdvisor(ctx context.Context, studentID string, advisorID string) error

	GetAllStudents(ctx context.Context) ([]*model.Student, error)
	GetStudentsByAdvisor(ctx context.Context, advisorID string) ([]*model.Student, error)
	GetStudentByID(ctx context.Context, studentID string) (*model.Student, error)
}

type studentRepository struct {
	db *pgxpool.Pool
}

func NewStudentRepository(db *pgxpool.Pool) StudentRepository {
	return &studentRepository{db: db}
}

func (r *studentRepository) GetStudentProfile(ctx context.Context, userID string) (*model.Student, error) {
	sql := `
        SELECT id, user_id, student_id, program_study, academic_year, advisor_id
        FROM students
        WHERE user_id = $1
    `
	s := &model.Student{}
	err := r.db.QueryRow(ctx, sql, userID).Scan(
		&s.ID, &s.UserID, &s.StudentID,
		&s.ProgramStudy, &s.AcademicYear, &s.AdvisorID,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *studentRepository) UpdateAdvisor(ctx context.Context, studentID string, advisorID string) error {
	query := `
        UPDATE students 
        SET advisor_id = $1 
        WHERE id = $2
    `
	_, err := r.db.Exec(ctx, query, advisorID, studentID)
	return err
}

func (r *studentRepository) GetAllStudents(ctx context.Context) ([]*model.Student, error) {
	query := `
        SELECT id, user_id, student_id, program_study, academic_year, advisor_id
        FROM students
    `
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []*model.Student

	for rows.Next() {
		s := &model.Student{}
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.StudentID,
			&s.ProgramStudy, &s.AcademicYear, &s.AdvisorID,
		); err != nil {
			return nil, err
		}
		students = append(students, s)
	}

	return students, nil
}

func (r *studentRepository) GetStudentsByAdvisor(ctx context.Context, advisorID string) ([]*model.Student, error) {
	query := `
        SELECT id, user_id, student_id, program_study, academic_year, advisor_id
        FROM students
        WHERE advisor_id = $1
    `
	rows, err := r.db.Query(ctx, query, advisorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []*model.Student

	for rows.Next() {
		s := &model.Student{}
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.StudentID,
			&s.ProgramStudy, &s.AcademicYear, &s.AdvisorID,
		); err != nil {
			return nil, err
		}
		students = append(students, s)
	}

	return students, nil
}

func (r *studentRepository) GetStudentByID(ctx context.Context, studentID string) (*model.Student, error) {
	query := `
        SELECT id, user_id, student_id, program_study, academic_year, advisor_id
        FROM students
        WHERE id = $1
    `
	s := &model.Student{}

	err := r.db.QueryRow(ctx, query, studentID).Scan(
		&s.ID, &s.UserID, &s.StudentID,
		&s.ProgramStudy, &s.AcademicYear, &s.AdvisorID,
	)
	if err != nil {
		return nil, err
	}

	return s, nil
}
