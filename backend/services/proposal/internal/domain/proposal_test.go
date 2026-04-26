package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidateForSubmit(t *testing.T) {
	p := Proposal{
		JobID:         10,
		ClientID:      uuid.New(),
		FreelancerID:  uuid.New(),
		CoverLetter:   "I can deliver this project with clean architecture.",
		BidType:       BidTypeFixed,
		BidAmount:     500,
		EstimatedDays: 7,
		Status:        StatusSent,
	}
	if err := ValidateForSubmit(p); err != nil {
		t.Fatalf("expected valid proposal, got error: %v", err)
	}
}

func TestCanTransition(t *testing.T) {
	cases := []struct {
		name    string
		current string
		next    string
		want    bool
	}{
		{name: "sent to shortlisted", current: StatusSent, next: StatusShortlisted, want: true},
		{name: "sent to rejected", current: StatusSent, next: StatusRejected, want: true},
		{name: "shortlisted to offer_sent", current: StatusShortlisted, next: StatusOfferSent, want: true},
		{name: "offer_sent to hired", current: StatusOfferSent, next: StatusHired, want: true},
		{name: "shortlisted to hired", current: StatusShortlisted, next: StatusHired, want: true},
		{name: "hired to rejected not allowed", current: StatusHired, next: StatusRejected, want: false},
		{name: "withdrawn to shortlisted not allowed", current: StatusWithdrawn, next: StatusShortlisted, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanTransition(tc.current, tc.next); got != tc.want {
				t.Fatalf("CanTransition(%q,%q)=%v want %v", tc.current, tc.next, got, tc.want)
			}
		})
	}
}
