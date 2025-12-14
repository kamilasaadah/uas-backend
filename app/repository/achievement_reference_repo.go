package repository

import (
	"context"
	"uas-backend/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AchievementReferenceRepository interface {
	CreateDraft(ctx context.Context, studentID, mongoID string) error
	GetByAchievementID(ctx context.Context, achievementID string) (*model.AchievementReference, error)
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

func (r *achievementReferenceRepository) GetByAchievementID(
	ctx context.Context,
	achievementID string,
) (*model.AchievementReference, error) {

	var ref model.AchievementReference

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			student_id,
			mongo_achievement_id,
			status
		FROM achievement_references
		WHERE mongo_achievement_id = $1
		`,
		achievementID,
	).Scan(
		&ref.ID,
		&ref.StudentID,
		&ref.MongoAchievementID,
		&ref.Status,
	)

	if err != nil {
		return nil, err
	}

	return &ref, nil
}
