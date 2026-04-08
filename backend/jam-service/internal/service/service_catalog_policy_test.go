package service

import (
	"context"
	"errors"
	"testing"

	"video-streaming/backend/jams/internal/model"
	"video-streaming/backend/jams/internal/repository"
	sharedcatalog "video-streaming/backend/shared/catalog"
)

type policyStubCatalogValidator struct {
	response sharedcatalog.LookupResponse
	err      error
}

func (v policyStubCatalogValidator) ValidateTrack(_ context.Context, _ string) (sharedcatalog.LookupResponse, error) {
	if v.err != nil {
		return sharedcatalog.LookupResponse{}, v.err
	}
	if v.response.TrackID != "" {
		return v.response, nil
	}
	return sharedcatalog.LookupResponse{TrackID: "trk_ok", IsPlayable: true, PolicyStatus: "allowed"}, nil
}

func TestAddRejectedWhenCatalogPolicyRestricted(t *testing.T) {
	t.Parallel()

	repo := repository.NewRedisQueueRepository()
	svc := NewWithCatalogValidator(repo, nil, policyStubCatalogValidator{
		response: sharedcatalog.LookupResponse{
			TrackID:      "trk_3",
			IsPlayable:   true,
			PolicyStatus: "restricted",
			PolicyReason: "region_blocked",
		},
	}, true)

	session, err := svc.CreateSession("host_1")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	_, _, err = svc.Add(session.JamID, model.AddQueueItemRequest{TrackID: "trk_3", AddedBy: "host_1", IdempotencyKey: "k_restricted"})
	if !errors.Is(err, ErrTrackRestricted) {
		t.Fatalf("expected ErrTrackRestricted, got %v", err)
	}

	snapshot, err := svc.Snapshot(session.JamID)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snapshot.QueueVersion != 0 || len(snapshot.Items) != 0 {
		t.Fatalf("queue state mutated on restriction: %+v", snapshot)
	}
}

func TestAddPolicyDisabledPreservesIdempotencyBehavior(t *testing.T) {
	t.Parallel()

	repo := repository.NewRedisQueueRepository()
	svc := NewWithCatalogValidator(repo, nil, policyStubCatalogValidator{
		response: sharedcatalog.LookupResponse{
			TrackID:      "trk_3",
			IsPlayable:   true,
			PolicyStatus: "restricted",
			PolicyReason: "region_blocked",
		},
	}, false)

	session, err := svc.CreateSession("host_1")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	first, fromCache, err := svc.Add(session.JamID, model.AddQueueItemRequest{TrackID: "trk_3", AddedBy: "host_1", IdempotencyKey: "idem_1"})
	if err != nil {
		t.Fatalf("first add: %v", err)
	}
	if fromCache {
		t.Fatal("expected first add to be cache miss")
	}

	second, fromCache, err := svc.Add(session.JamID, model.AddQueueItemRequest{TrackID: "trk_3", AddedBy: "host_1", IdempotencyKey: "idem_1"})
	if err != nil {
		t.Fatalf("second add: %v", err)
	}
	if !fromCache {
		t.Fatal("expected second add to be idempotent cache hit")
	}
	if first.QueueVersion != second.QueueVersion {
		t.Fatalf("queueVersion mismatch for idempotent retry: first=%d second=%d", first.QueueVersion, second.QueueVersion)
	}
	if len(second.Items) != 1 {
		t.Fatalf("expected single queue item after idempotent retry, got %d", len(second.Items))
	}
}
