package repository

import (
	"context"
	"uas-backend/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AchievementReferenceRepository interface {
	CreateDraft(ctx context.Context, studentID, mongoID string) error
	GetByAchievementID(ctx context.Context, achievementID string) (*model.AchievementReference, error)
	MarkDeleted(ctx context.Context, achievementID string) error
	Submit(ctx context.Context, achievementID string) (*model.AchievementReference, error)
	Verify(
		ctx context.Context,
		achievementID string,
		verifiedBy string,
	) (*model.AchievementReference, error)
	Reject(
		ctx context.Context,
		achievementID string,
		rejectionNote string,
	) (*model.AchievementReference, error)
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

func (r *achievementReferenceRepository) MarkDeleted(ctx context.Context, achievementID string) error {
	query := `
		UPDATE achievement_references
		SET status = 'deleted', updated_at = NOW()
		WHERE mongo_achievement_id = $1
	`
	_, err := r.db.Exec(ctx, query, achievementID)
	return err
}

func (r *achievementReferenceRepository) Submit(
	ctx context.Context,
	achievementID string,
) (*model.AchievementReference, error) {

	query := `
		UPDATE achievement_references
		SET
			status = 'submitted',
			submitted_at = NOW(),
			updated_at = NOW()
		WHERE mongo_achievement_id = $1
		  AND status = 'draft'
		RETURNING
			id,
			student_id,
			mongo_achievement_id,
			status,
			submitted_at,
			verified_at,
			verified_by,
			rejection_note,
			created_at,
			updated_at
	`

	var ref model.AchievementReference
	err := r.db.QueryRow(ctx, query, achievementID).Scan(
		&ref.ID,
		&ref.StudentID,
		&ref.MongoAchievementID,
		&ref.Status,
		&ref.SubmittedAt,
		&ref.VerifiedAt,
		&ref.VerifiedBy,
		&ref.RejectionNote,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &ref, nil
}

func (r *achievementReferenceRepository) Verify(
	ctx context.Context,
	achievementID string,
	verifiedBy string,
) (*model.AchievementReference, error) {

	query := `
		UPDATE achievement_references
		SET
			status = 'verified',
			verified_at = NOW(),
			verified_by = $2,
			updated_at = NOW()
		WHERE mongo_achievement_id = $1
		  AND status = 'submitted'
		RETURNING
			id,
			student_id,
			mongo_achievement_id,
			status,
			submitted_at,
			verified_at,
			verified_by,
			rejection_note,
			created_at,
			updated_at
	`

	var ref model.AchievementReference
	err := r.db.QueryRow(
		ctx,
		query,
		achievementID,
		verifiedBy,
	).Scan(
		&ref.ID,
		&ref.StudentID,
		&ref.MongoAchievementID,
		&ref.Status,
		&ref.SubmittedAt,
		&ref.VerifiedAt,
		&ref.VerifiedBy,
		&ref.RejectionNote,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &ref, nil
}

func (r *achievementReferenceRepository) Reject(
	ctx context.Context,
	achievementID string,
	rejectionNote string,
) (*model.AchievementReference, error) {

	query := `
		UPDATE achievement_references
		SET
			status = 'rejected',
			rejection_note = $2,
			updated_at = NOW()
		WHERE mongo_achievement_id = $1
		  AND status = 'submitted'
		RETURNING
			id,
			student_id,
			mongo_achievement_id,
			status,
			submitted_at,
			verified_at,
			verified_by,
			rejection_note,
			created_at,
			updated_at
	`

	var ref model.AchievementReference
	err := r.db.QueryRow(
		ctx,
		query,
		achievementID,
		rejectionNote,
	).Scan(
		&ref.ID,
		&ref.StudentID,
		&ref.MongoAchievementID,
		&ref.Status,
		&ref.SubmittedAt,
		&ref.VerifiedAt,
		&ref.VerifiedBy,
		&ref.RejectionNote,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &ref, nil
}
