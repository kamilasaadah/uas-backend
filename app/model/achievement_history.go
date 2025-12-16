package model

import "time"

type AchievementStatusHistory struct {
	Status    string    `json:"status"`
	Note      *string   `json:"note,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy *string   `json:"updated_by,omitempty"` // verified_by / admin
}

type AchievementHistoryResponse struct {
	Achievement Achievement                `json:"achievement"`
	History     []AchievementStatusHistory `json:"history"`
}
