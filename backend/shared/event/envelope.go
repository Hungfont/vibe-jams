package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrEventIDRequired          = errors.New("eventId is required")
	ErrEventTypeRequired        = errors.New("eventType is required")
	ErrOccurredAtRequired       = errors.New("occurredAt is required")
	ErrSessionIDRequired        = errors.New("sessionId is required for session-scoped events")
	ErrAggregateVersionRequired = errors.New("aggregateVersion must be positive for session-scoped events")
)

// Envelope is the standardized event wrapper for Kafka producers.
type Envelope struct {
	EventID          string          `json:"eventId"`
	EventType        string          `json:"eventType"`
	SessionID        string          `json:"sessionId,omitempty"`
	AggregateVersion int64           `json:"aggregateVersion,omitempty"`
	ActorUserID      string          `json:"actorUserId,omitempty"`
	OccurredAt       time.Time       `json:"occurredAt"`
	Payload          json.RawMessage `json:"payload"`
}

// Validate checks required metadata fields based on event scope.
func (e Envelope) Validate(sessionScoped bool) error {
	if e.EventID == "" {
		return ErrEventIDRequired
	}
	if e.EventType == "" {
		return ErrEventTypeRequired
	}
	if e.OccurredAt.IsZero() {
		return ErrOccurredAtRequired
	}
	if sessionScoped {
		if e.SessionID == "" {
			return ErrSessionIDRequired
		}
		if e.AggregateVersion <= 0 {
			return ErrAggregateVersionRequired
		}
	}
	return nil
}

// MustPayload marshals payload and panics only for programmer errors in static payloads.
func MustPayload(v any) json.RawMessage {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("marshal payload: %v", err))
	}
	return raw
}

// MarshalEnvelope validates then serializes the envelope payload.
func MarshalEnvelope(e Envelope, sessionScoped bool) ([]byte, error) {
	if err := e.Validate(sessionScoped); err != nil {
		return nil, err
	}
	return json.Marshal(e)
}

// UnmarshalEnvelope decodes bytes into Envelope.
func UnmarshalEnvelope(data []byte) (Envelope, error) {
	var out Envelope
	if err := json.Unmarshal(data, &out); err != nil {
		return Envelope{}, err
	}
	return out, nil
}
