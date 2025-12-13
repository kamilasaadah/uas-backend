package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AchievementReferenceRepository interface {
	CreateDraft(ctx context.Context, studentID, mongoID string) error
}

type achievementReferenceRepository struct {
	db *pgxpool.Pool
}

func NewAchievementReferenceRepository(db *pgxpool.Pool) AchievementReferenceRepository {
	return &achievementReferenceRepository{db: db}
}

func (r *achievementReferenceRepository) CreateDraft(
	ctx context.Context,
	studentID string,
	mongoID string,
) error {

	query := `
		INSERT INTO achievement_references (
			id, student_id, mongo_achievement_id, status
		) VALUES (
			gen_random_uuid(), $1, $2, 'draft'
		)
	`

	_, err := r.db.Exec(ctx, query, studentID, mongoID)
	return err
}
