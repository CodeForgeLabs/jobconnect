package domain

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

// Data models fetched from other services (Job and User)
type JobData struct {
	ID             int64
	RequiredSkills []string
}

type UserData struct {
	ID     string
	Skills []string
	Rating float32
}
