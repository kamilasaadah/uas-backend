package repository

import (
	"context"

	"uas-backend/app/model"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AchievementHistoryRepository interface {
	FindByAchievementReferenceID(
		ctx context.Context,
		refID string,
	) ([]model.AchievementStatusHistory, error)
}

type achievementHistoryRepository struct {
	db *pgxpool.Pool
}

func NewAchievementHistoryRepository(
	db *pgxpool.Pool,
) AchievementHistoryRepository {
	return &achievementHistoryRepository{
		db: db,
	}
}

func (r *achievementHistoryRepository) FindByAchievementReferenceID(
	ctx context.Context,
	refID string,
) ([]model.AchievementStatusHistory, error) {

	query := `
		SELECT status, rejection_note, submitted_at, verified_at, verified_by, updated_at
		FROM achievement_references
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, refID)

	var status string
	var rejectionNote pgtype.Text
	var submittedAt, verifiedAt pgtype.Timestamptz
	var verifiedBy pgtype.UUID
	var updatedAt pgtype.Timestamptz

	if err := row.Scan(&status, &rejectionNote, &submittedAt, &verifiedAt, &verifiedBy, &updatedAt); err != nil {
		return nil, err
	}

	history := []model.AchievementStatusHistory{}

	// draft
	if status == "draft" {
		history = append(history, model.AchievementStatusHistory{
			Status:    "draft",
			UpdatedAt: updatedAt.Time,
		})
	}

	// submitted
	if !submittedAt.Valid {
		history = append(history, model.AchievementStatusHistory{
			Status:    "submitted",
			Note:      nil,
			UpdatedAt: submittedAt.Time,
		})
	}

	// verified
	if !verifiedAt.Valid {
		verifiedByStr := ""
		if verifiedBy.Valid {
			verifiedByStr = verifiedBy.String()
		}
		history = append(history, model.AchievementStatusHistory{
			Status:    "verified",
			UpdatedAt: verifiedAt.Time,
			UpdatedBy: &verifiedByStr,
		})
	}

	// rejected
	if status == "rejected" && rejectionNote.Valid {
		history = append(history, model.AchievementStatusHistory{
			Status:    "rejected",
			Note:      &rejectionNote.String,
			UpdatedAt: updatedAt.Time,
		})
	}

	return history, nil
}
