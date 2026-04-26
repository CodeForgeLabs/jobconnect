package grpcadapter

import (
	"context"
	"errors"
	"testing"
	"time"

	userv1 "jobconnect/user/gen/user"
	"jobconnect/user/internal/application"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeDiscoverableRepo struct {
	application.ProfileDetailsRepository
	lastFilter   application.DiscoverableFreelancerFilter
	lastPageSize uint32
	lastPageTok  string
	result       application.ListResult[application.DiscoverableFreelancer]
	err          error
}

func (f *fakeDiscoverableRepo) ListDiscoverableFreelancers(_ context.Context, filter application.DiscoverableFreelancerFilter, pageSize uint32, pageToken string) (application.ListResult[application.DiscoverableFreelancer], error) {
	f.lastFilter = filter
	f.lastPageSize = pageSize
	f.lastPageTok = pageToken
	if f.err != nil {
		return application.ListResult[application.DiscoverableFreelancer]{}, f.err
	}
	return f.result, nil
}

func TestListDiscoverableFreelancersAppliesCapabilityPolicy(t *testing.T) {
	repo := &fakeDiscoverableRepo{
		result: application.ListResult[application.DiscoverableFreelancer]{
			Items: []application.DiscoverableFreelancer{{
				UserID:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				Headline:     "Go engineer",
				Skills:       []string{"Go", "gRPC"},
				HourlyRate:   85,
				Availability: "FULL_TIME",
				Rating:       4.7,
				TotalReviews: 12,
			}},
			NextPageToken: "20",
		},
	}
	srv := &UserServer{
		ProfileDetailsRepo: repo,
		CapabilityPolicy: CapabilityPolicy{
			MinSkillsForDiscovery:        2,
			RequireHeadlineForFreelancer: true,
		},
	}

	resp, err := srv.ListDiscoverableFreelancers(context.Background(), &userv1.ListDiscoverableFreelancersRequest{
		PageSize: 5,
		Skills:   []string{"Go"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := repo.lastFilter.MinSkills; got != 2 {
		t.Fatalf("expected MinSkills 2 propagated from policy, got %d", got)
	}
	if !repo.lastFilter.RequireHeadline {
		t.Fatalf("expected RequireHeadline forwarded to filter")
	}
	if !repo.lastFilter.RequireActiveAccount {
		t.Fatalf("expected RequireActiveAccount always true")
	}
	if got := repo.lastPageSize; got != 5 {
		t.Fatalf("expected page size 5, got %d", got)
	}
	if len(resp.GetFreelancers()) != 1 {
		t.Fatalf("expected 1 freelancer, got %d", len(resp.GetFreelancers()))
	}
	card := resp.GetFreelancers()[0]
	if !card.GetCanBeDiscovered() {
		t.Fatalf("expected card.can_be_discovered=true")
	}
	if card.GetAvailability() != userv1.Availability_AVAILABILITY_FULL_TIME {
		t.Fatalf("unexpected availability: %v", card.GetAvailability())
	}
	if resp.GetPage().GetNextPageToken() != "20" {
		t.Fatalf("expected next_page_token propagated, got %q", resp.GetPage().GetNextPageToken())
	}
}

func TestListDiscoverableFreelancersRejectsNilRequest(t *testing.T) {
	srv := &UserServer{ProfileDetailsRepo: &fakeDiscoverableRepo{}}
	_, err := srv.ListDiscoverableFreelancers(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", got)
	}
}

func TestListDiscoverableFreelancersPropagatesRepoError(t *testing.T) {
	repo := &fakeDiscoverableRepo{err: errors.New("boom: not found")}
	srv := &UserServer{ProfileDetailsRepo: repo}
	_, err := srv.ListDiscoverableFreelancers(context.Background(), &userv1.ListDiscoverableFreelancersRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := status.Code(err); got != codes.NotFound {
		t.Fatalf("expected NotFound mapped from 'not found' error, got %v", got)
	}
}

func TestListDiscoverableFreelancersNormalizesLastActive(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	repo := &fakeDiscoverableRepo{
		result: application.ListResult[application.DiscoverableFreelancer]{
			Items: []application.DiscoverableFreelancer{{
				UserID:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				LastActiveAt: &now,
			}},
		},
	}
	srv := &UserServer{ProfileDetailsRepo: repo}

	resp, err := srv.ListDiscoverableFreelancers(context.Background(), &userv1.ListDiscoverableFreelancersRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	card := resp.GetFreelancers()[0]
	if card.LastActiveAtUnix == nil || *card.LastActiveAtUnix != now.Unix() {
		t.Fatalf("expected last_active_at_unix=%d, got %+v", now.Unix(), card.LastActiveAtUnix)
	}
}
