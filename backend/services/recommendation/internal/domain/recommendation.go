package domain

import "time"

type JobRecommendation struct {
	JobID       int64
	MatchScore  float32
	MatchReason string
}

type FreelancerRecommendation struct {
	UserID      string
	MatchScore  float32
	MatchReason string
}

type JobData struct {
	ID             int64
	ClientID       string
	Title          string
	Description    string
	RequiredSkills []string
	BudgetMin      float64
	BudgetMax      float64
	HourlyRate     float64
	JobType        string
	Visibility     string
	CreatedAt      time.Time
}

type UserData struct {
	ID           string
	Headline     string
	Bio          string
	Skills       []string
	HourlyRate   float64
	Availability string
	Rating       float64
	CanApplyJobs bool
}

type WorkPreferences struct {
	PreferredProjectLength string
	MinBudgetUSD           float64
	MaxBudgetUSD           float64
	ContractTypes          []string
	WeeklyCapacityHours    uint32
}
