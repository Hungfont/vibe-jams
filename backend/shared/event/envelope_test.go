package event

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMarshalEnvelope_ValidSessionScoped(t *testing.T) {
	env := Envelope{
		EventID:          "evt-1",
		EventType:        "jam.queue.item.added",
		SessionID:        "jam_1",
		AggregateVersion: 2,
		OccurredAt:       time.Now().UTC(),
		Payload:          MustPayload(map[string]string{"trackId": "t1"}),
	}

	raw, err := MarshalEnvelope(env, true)
	if err != nil {
		t.Fatalf("MarshalEnvelope() error = %v", err)
	}

	decoded, err := UnmarshalEnvelope(raw)
	if err != nil {
		t.Fatalf("UnmarshalEnvelope() error = %v", err)
	}
	if decoded.SessionID != "jam_1" {
		t.Fatalf("unexpected sessionId: %s", decoded.SessionID)
	}
}

func TestMarshalEnvelope_MissingSessionMetadata(t *testing.T) {
	env := Envelope{
		EventID:    "evt-1",
		EventType:  "jam.queue.item.added",
		OccurredAt: time.Now().UTC(),
		Payload:    MustPayload(map[string]string{"trackId": "t1"}),
	}

	if _, err := MarshalEnvelope(env, true); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestConsumerContractSamples_ParseAcrossPhase1Topics(t *testing.T) {
	samples := []string{
		`{"eventId":"1","eventType":"jam.session.updated","sessionId":"jam_123","aggregateVersion":7,"occurredAt":"2026-03-30T10:00:00Z","payload":{"status":"active"}}`,
		`{"eventId":"2","eventType":"jam.queue.item.added","sessionId":"jam_123","aggregateVersion":8,"occurredAt":"2026-03-30T10:00:01Z","payload":{"trackId":"trk_1"}}`,
		`{"eventId":"3","eventType":"jam.playback.changed","sessionId":"jam_123","aggregateVersion":9,"occurredAt":"2026-03-30T10:00:02Z","payload":{"state":"play"}}`,
		`{"eventId":"4","eventType":"analytics.user.action","actorUserId":"usr_1","occurredAt":"2026-03-30T10:00:03Z","payload":{"action":"queue_add"}}`,
	}

	for _, sample := range samples {
		var env Envelope
		if err := json.Unmarshal([]byte(sample), &env); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if env.EventID == "" || env.EventType == "" {
			t.Fatalf("missing required envelope metadata: %+v", env)
		}
	}
}
