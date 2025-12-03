package model

type Student struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	StudentID    string `json:"student_id"`
	ProgramStudy string `json:"program_study"`
	AcademicYear string `json:"academic_year"`
	AdvisorID    string `json:"advisor_id"`
}

type StudentProfileRequest struct {
	StudentID    string `json:"student_id" validate:"required"`
	ProgramStudy string `json:"program_study"`
	AcademicYear string `json:"academic_year"`
	AdvisorID    string `json:"advisor_id"` // lecturer UUID
}
