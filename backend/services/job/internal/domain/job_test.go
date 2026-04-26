package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateCreateFixedJob(t *testing.T) {
	now := time.Now().UTC()
	err := ValidateCreate(Job{
		ClientID:       uuid.New(),
		Title:          "Need API integration",
		Description:    "Integrate our payment API",
		RequiredSkills: []string{"go", "grpc"},
		JobType:        JobTypeFixed,
		BudgetFixed:    500,
		Status:         JobStatusOpen,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, now)
	if err != nil {
		t.Fatalf("expected valid create, got err: %v", err)
	}
}

func TestValidateCreateRequiresHourlyRateForHourly(t *testing.T) {
	now := time.Now().UTC()
	err := ValidateCreate(Job{
		ClientID:    uuid.New(),
		Title:       "Need support",
		Description: "Support needed",
		JobType:     JobTypeHourly,
	}, now)
	if err == nil {
		t.Fatal("expected error for missing hourly_rate")
	}
}
