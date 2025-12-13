package model

import "time"

type AchievementReference struct {
	ID                 string     `db:"id"`
	StudentID          string     `db:"student_id"`
	MongoAchievementID string     `db:"mongo_achievement_id"`
	Status             string     `db:"status"`
	SubmittedAt        *time.Time `db:"submitted_at"`
	VerifiedAt         *time.Time `db:"verified_at"`
	VerifiedBy         *string    `db:"verified_by"`
	RejectionNote      *string    `db:"rejection_note"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
}
