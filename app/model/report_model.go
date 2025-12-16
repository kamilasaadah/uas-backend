package model

type AchievementTypeStat struct {
	Type  string `json:"type"`
	Total int    `json:"total"`
}

type PeriodStat struct {
	Period string `json:"period"` // contoh: 2025-12
	Total  int    `json:"total"`
}

type CompetitionLevelStat struct {
	Level string `json:"level"`
	Total int    `json:"total"`
}

type StudentAchievementStat struct {
	StudentID string `json:"student_id"`
	Total     int    `json:"total"`
}

type AchievementStatisticsResponse struct {
	TotalPerType         []AchievementTypeStat    `json:"total_per_type"`
	TotalPerPeriod       []PeriodStat             `json:"total_per_period"`
	CompetitionLevels    []CompetitionLevelStat   `json:"competition_level_distribution"`
	TotalPerStudent      []StudentAchievementStat `json:"total_per_student"`
}
