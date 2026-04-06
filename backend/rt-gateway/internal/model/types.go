package model

import "time"

// SessionParticipant mirrors jam-service participant read model.
type SessionParticipant struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	Muted  bool   `json:"muted,omitempty"`
}

// SessionSnapshot is the lifecycle read model used for recovery snapshots.
type SessionSnapshot struct {
	JamID          string               `json:"jamId"`
	Status         string               `json:"status"`
	HostUserID     string               `json:"hostUserId"`
	Participants   []SessionParticipant `json:"participants"`
	SessionVersion int64                `json:"sessionVersion"`
	EndCause       string               `json:"endCause,omitempty"`
	EndedBy        string               `json:"endedBy,omitempty"`
}

// QueueItem mirrors jam queue read model.
type QueueItem struct {
	ItemID  string `json:"itemId"`
	TrackID string `json:"trackId"`
	AddedBy string `json:"addedBy"`
}

// QueueSnapshot is the queue read model used for recovery snapshots.
type QueueSnapshot struct {
	JamID        string      `json:"jamId"`
	QueueVersion int64       `json:"queueVersion"`
	Items        []QueueItem `json:"items"`
}

// SessionStateSnapshot is the authoritative state payload from jam-service.
type SessionStateSnapshot struct {
	Session          SessionSnapshot `json:"session"`
	Queue            QueueSnapshot   `json:"queue"`
	AggregateVersion int64           `json:"aggregateVersion"`
}

// OutboundEvent is the websocket event envelope for room subscribers.
type OutboundEvent struct {
	EventType        string    `json:"eventType"`
	SessionID        string    `json:"sessionId"`
	AggregateVersion int64     `json:"aggregateVersion"`
	OccurredAt       time.Time `json:"occurredAt"`
	Payload          any       `json:"payload"`
	Recovery         bool      `json:"recovery,omitempty"`
}
