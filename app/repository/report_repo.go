package repository

import "context"

type ReportRepository interface {
	GetVerifiedAchievementIDs(ctx context.Context) ([]string, error)
	GetVerifiedAchievementIDsByStudent(ctx context.Context, studentID string) ([]string, error)
}

type reportRepository struct {
	ref AchievementReferenceRepository
}

func NewReportRepository(
	refRepo AchievementReferenceRepository,
) ReportRepository {
	return &reportRepository{ref: refRepo}
}

func (r *reportRepository) GetVerifiedAchievementIDs(
	ctx context.Context,
) ([]string, error) {

	rows, err := r.ref.(*achievementReferenceRepository).db.Query(
		ctx,
		`
		SELECT mongo_achievement_id
		FROM achievement_references
		WHERE status = 'verified'
		`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *reportRepository) GetVerifiedAchievementIDsByStudent(
	ctx context.Context,
	studentID string,
) ([]string, error) {

	rows, err := r.ref.(*achievementReferenceRepository).db.Query(
		ctx,
		`
		SELECT mongo_achievement_id
		FROM achievement_references
		WHERE status = 'verified'
		  AND student_id = $1
		`,
		studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
