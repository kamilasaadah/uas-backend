package repository

import (
	"context"
	"uas-backend/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LecturerRepository interface {
	GetLecturerProfile(ctx context.Context, userID string) (*model.Lecturer, error)
}

type lecturerRepository struct {
	db *pgxpool.Pool
}

func NewLecturerRepository(db *pgxpool.Pool) LecturerRepository {
	return &lecturerRepository{db: db}
}

func (r *lecturerRepository) GetLecturerProfile(ctx context.Context, userID string) (*model.Lecturer, error) {
	sql := `
        SELECT id, user_id, lecturer_id, department
        FROM lecturers
        WHERE user_id = $1
    `
	l := &model.Lecturer{}

	err := r.db.QueryRow(ctx, sql, userID).Scan(
		&l.ID, &l.UserID, &l.LecturerID, &l.Department,
	)
	if err != nil {
		return nil, err
	}

	return l, nil
}
