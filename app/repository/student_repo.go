package repository

import (
	"context"
	"uas-backend/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentRepository interface {
	GetStudentProfile(ctx context.Context, userID string) (*model.Student, error)
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
