package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	playbackauth "video-streaming/backend/playback-service/internal/auth"
	"video-streaming/backend/playback-service/internal/kafka"
	"video-streaming/backend/playback-service/internal/repository"
	"video-streaming/backend/playback-service/internal/service"
	sharedauth "video-streaming/backend/shared/auth"
	sharedevent "video-streaming/backend/shared/event"
	sharedkafka "video-streaming/backend/shared/kafka"
)

type stubValidator struct {
	claims sharedauth.Claims
	err    error
}

func (s stubValidator) ValidateBearerToken(_ context.Context, _ string) (sharedauth.Claims, error) {
	if s.err != nil {
		return sharedauth.Claims{}, s.err
	}
	return s.claims, nil
}

func TestPlaybackCommand_AcceptedPublishesKafkaEvent(t *testing.T) {
	t.Parallel()

	repo := repository.NewRedisPlaybackRepository()
	if err := repo.SeedSession("jam_1", "host_1", 7); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	pub := &kafka.InMemoryPublisher{}
	producer := kafka.NewProducer(pub)
	svc := service.New(repo, producer)
	h := NewHTTPHandler(svc, stubValidator{
		claims: sharedauth.Claims{
			UserID:       "host_1",
			Plan:         "premium",
			SessionState: sharedauth.SessionStateValid,
		},
	})

	body := []byte(`{"command":"pause","clientEventId":"evt_1","expectedQueueVersion":7}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/jam/sessions/jam_1/playback/commands", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer token-premium-valid")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusAccepted)
	}
	if len(pub.Records) != 1 {
		t.Fatalf("expected one published record, got %d", len(pub.Records))
	}
	if pub.Records[0].Topic != sharedkafka.TopicJamPlayback {
		t.Fatalf("topic mismatch: got %s want %s", pub.Records[0].Topic, sharedkafka.TopicJamPlayback)
	}
	if pub.Records[0].Key != "jam_1" {
		t.Fatalf("key mismatch: got %s want jam_1", pub.Records[0].Key)
	}

	envelope, err := sharedevent.UnmarshalEnvelope(pub.Records[0].Value)
	if err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if envelope.EventType != "jam.playback.updated" {
		t.Fatalf("eventType mismatch: got %q", envelope.EventType)
	}
	if envelope.AggregateVersion != 1 {
		t.Fatalf("aggregateVersion mismatch: got %d want 1", envelope.AggregateVersion)
	}
}

func TestPlaybackCommand_Unauthorized(t *testing.T) {
	t.Parallel()

	repo := repository.NewRedisPlaybackRepository()
	if err := repo.SeedSession("jam_1", "host_1", 7); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	svc := service.New(repo, kafka.NewProducer(&kafka.InMemoryPublisher{}))
	h := NewHTTPHandler(svc, stubValidator{
		err: playbackauth.ErrUnauthorized,
	})

	body := []byte(`{"command":"pause","clientEventId":"evt_1","expectedQueueVersion":7}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/jam/sessions/jam_1/playback/commands", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestPlaybackCommand_NonHostRejectedAndNoPublish(t *testing.T) {
	t.Parallel()

	repo := repository.NewRedisPlaybackRepository()
	if err := repo.SeedSession("jam_1", "host_1", 7); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	pub := &kafka.InMemoryPublisher{}
	svc := service.New(repo, kafka.NewProducer(pub))
	h := NewHTTPHandler(svc, stubValidator{
		claims: sharedauth.Claims{
			UserID:       "guest_2",
			Plan:         "premium",
			SessionState: sharedauth.SessionStateValid,
		},
	})

	body := []byte(`{"command":"next","clientEventId":"evt_2","expectedQueueVersion":7}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/jam/sessions/jam_1/playback/commands", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer token-guest-valid")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusForbidden)
	}
	if len(pub.Records) != 0 {
		t.Fatalf("expected no published record, got %d", len(pub.Records))
	}

	var errorBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&errorBody); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if errorBody.Error.Code != "host_only" {
		t.Fatalf("error code mismatch: got %q want host_only", errorBody.Error.Code)
	}
}

func TestPlaybackCommand_StaleVersionRejectedAndNoPublish(t *testing.T) {
	t.Parallel()

	repo := repository.NewRedisPlaybackRepository()
	if err := repo.SeedSession("jam_1", "host_1", 7); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	pub := &kafka.InMemoryPublisher{}
	svc := service.New(repo, kafka.NewProducer(pub))
	h := NewHTTPHandler(svc, stubValidator{
		claims: sharedauth.Claims{
			UserID:       "host_1",
			Plan:         "premium",
			SessionState: sharedauth.SessionStateValid,
		},
	})

	body := []byte(`{"command":"next","clientEventId":"evt_3","expectedQueueVersion":6}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/jam/sessions/jam_1/playback/commands", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer token-host-valid")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusConflict)
	}
	if len(pub.Records) != 0 {
		t.Fatalf("expected no published record, got %d", len(pub.Records))
	}

	var errorBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&errorBody); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if errorBody.Error.Code != "version_conflict" {
		t.Fatalf("error code mismatch: got %q want version_conflict", errorBody.Error.Code)
	}
}

func TestPlaybackCommand_EndedSessionRejectedAndNoPublish(t *testing.T) {
	t.Parallel()

	repo := repository.NewRedisPlaybackRepository()
	if err := repo.SeedSession("jam_1", "host_1", 7); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	if err := repo.EndSession("jam_1", "host_1"); err != nil {
		t.Fatalf("end session: %v", err)
	}

	pub := &kafka.InMemoryPublisher{}
	svc := service.New(repo, kafka.NewProducer(pub))
	h := NewHTTPHandler(svc, stubValidator{
		claims: sharedauth.Claims{
			UserID:       "host_1",
			Plan:         "premium",
			SessionState: sharedauth.SessionStateValid,
		},
	})

	body := []byte(`{"command":"pause","clientEventId":"evt_4","expectedQueueVersion":7}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/jam/sessions/jam_1/playback/commands", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer token-host-valid")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusConflict)
	}
	if len(pub.Records) != 0 {
		t.Fatalf("expected no published record, got %d", len(pub.Records))
	}

	var errorBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&errorBody); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if errorBody.Error.Code != "session_ended" {
		t.Fatalf("error code mismatch: got %q want session_ended", errorBody.Error.Code)
	}
}
