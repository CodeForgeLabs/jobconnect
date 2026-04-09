package grpcadapter

import (
	"context"
	"testing"
	"time"

	userv1 "jobconnect/user/gen/user"
	"jobconnect/user/internal/application"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakePortfolioPresigner struct {
	calls []string
}

func (f *fakePortfolioPresigner) PresignGetObject(_ context.Context, storageKey string, _ time.Duration) (string, error) {
	f.calls = append(f.calls, storageKey)
	return "https://example.invalid/presigned/" + storageKey, nil
}

func (f *fakePortfolioPresigner) PresignPutObject(_ context.Context, storageKey string, _ string, _ time.Duration) (string, error) {
	return "https://example.invalid/upload/" + storageKey, nil
}

type fakePortfolioUploadStore struct{}

func (f *fakePortfolioUploadStore) PresignPutObject(_ context.Context, storageKey string, _ string, _ time.Duration) (string, error) {
	return "https://example.invalid/upload/" + storageKey, nil
}

func TestToProtoPortfolioItemPresignsStoredMedia(t *testing.T) {
	presigner := &fakePortfolioPresigner{}
	srv := &UserServer{PortfolioStore: presigner}
	item := application.PortfolioItem{
		ID:     7,
		UserID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Title:  "Portfolio",
		Media: []application.PortfolioMedia{{
			ID:          11,
			MediaType:   "IMAGE",
			StorageKey:  "portfolio/11111111-1111-1111-1111-111111111111/7/11.png",
			FileName:    "shot.png",
			ContentType: "image/png",
		}},
		CreatedAt: time.Unix(100, 0).UTC(),
		UpdatedAt: time.Unix(200, 0).UTC(),
	}

	resp, err := srv.toProtoPortfolioItem(context.Background(), item)
	if err != nil {
		t.Fatalf("toProtoPortfolioItem returned error: %v", err)
	}
	if len(presigner.calls) != 1 || presigner.calls[0] != item.Media[0].StorageKey {
		t.Fatalf("expected one presign call for %q, got %#v", item.Media[0].StorageKey, presigner.calls)
	}
	if got := resp.GetMedia()[0].GetStorageKey(); got != "" {
		t.Fatalf("expected storage_key to be hidden on response, got %q", got)
	}
	if got := resp.GetMedia()[0].GetExternalUrl(); got != "https://example.invalid/presigned/"+item.Media[0].StorageKey {
		t.Fatalf("expected presigned external_url, got %q", got)
	}
}

func TestToProtoPortfolioItemLeavesLinkMediaUntouched(t *testing.T) {
	srv := &UserServer{}
	item := application.PortfolioItem{
		ID:     8,
		UserID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Title:  "Portfolio",
		Media: []application.PortfolioMedia{{
			ID:          12,
			MediaType:   "LINK",
			ExternalURL: "https://example.com/work",
			FileName:    "link",
		}},
		CreatedAt: time.Unix(100, 0).UTC(),
		UpdatedAt: time.Unix(200, 0).UTC(),
	}

	resp, err := srv.toProtoPortfolioItem(context.Background(), item)
	if err != nil {
		t.Fatalf("toProtoPortfolioItem returned error: %v", err)
	}
	if got := resp.GetMedia()[0].GetStorageKey(); got != "" {
		t.Fatalf("expected no storage_key for link media, got %q", got)
	}
	if got := resp.GetMedia()[0].GetExternalUrl(); got != item.Media[0].ExternalURL {
		t.Fatalf("expected external_url passthrough, got %q", got)
	}
}

func TestMapPortfolioMediaTypeToProto(t *testing.T) {
	if got := mapPortfolioMediaTypeToProto("IMAGE"); got != userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_IMAGE {
		t.Fatalf("expected image type, got %v", got)
	}
	if got := mapPortfolioMediaTypeToProto("LINK"); got != userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_LINK {
		t.Fatalf("expected link type, got %v", got)
	}
}

func TestGetMyPortfolioMediaUploadUrl(t *testing.T) {
	uc := &application.GetPortfolioMediaUploadURL{Store: &fakePortfolioUploadStore{}}
	srv := &UserServer{GetPortfolioMediaUploadURLUC: uc}
	userID := uuid.New().String()

	resp, err := srv.GetMyPortfolioMediaUploadUrl(context.Background(), &userv1.GetMyPortfolioMediaUploadUrlRequest{
		UserId:      userID,
		FileName:    "sample.png",
		ContentType: "image/png",
	})
	if err != nil {
		t.Fatalf("GetMyPortfolioMediaUploadUrl() error = %v", err)
	}
	if resp.GetStorageKey() == "" {
		t.Fatalf("expected non-empty storage key")
	}
	if resp.GetUploadUrl() == "" {
		t.Fatalf("expected non-empty upload url")
	}
}

func TestGetMyPortfolioMediaUploadUrlRequiresConfiguredUseCase(t *testing.T) {
	srv := &UserServer{}
	_, err := srv.GetMyPortfolioMediaUploadUrl(context.Background(), &userv1.GetMyPortfolioMediaUploadUrlRequest{UserId: uuid.New().String()})
	if err == nil {
		t.Fatalf("expected error when use-case is not configured")
	}
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal, got %v", status.Code(err))
	}
}
