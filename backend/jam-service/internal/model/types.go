package model

// SessionStatus represents lifecycle state for one jam session.
type SessionStatus string

const (
	// SessionStatusActive means session accepts write commands.
	SessionStatusActive SessionStatus = "active"
	// SessionStatusEnded means session is terminated and write commands are blocked.
	SessionStatusEnded SessionStatus = "ended"
)

// SessionRole defines participant authorization role.
type SessionRole string

const (
	// SessionRoleHost marks the session owner.
	SessionRoleHost SessionRole = "host"
	// SessionRoleMember marks a non-host participant.
	SessionRoleMember SessionRole = "member"
)

// SessionParticipant represents one member plus role.
type SessionParticipant struct {
	UserID string      `json:"userId"`
	Role   SessionRole `json:"role"`
}

// SessionSnapshot describes persisted jam session metadata and participants.
type SessionSnapshot struct {
	JamID          string               `json:"jamId"`
	Status         SessionStatus        `json:"status"`
	HostUserID     string               `json:"hostUserId"`
	Participants   []SessionParticipant `json:"participants"`
	SessionVersion int64                `json:"sessionVersion"`
	EndCause       string               `json:"endCause,omitempty"`
	EndedBy        string               `json:"endedBy,omitempty"`
}

// QueueItem represents one track entry inside a jam queue.
type QueueItem struct {
	ItemID  string `json:"itemId"`
	TrackID string `json:"trackId"`
	AddedBy string `json:"addedBy"`
}

// QueueSnapshot is the read model returned after each successful mutation.
type QueueSnapshot struct {
	JamID        string      `json:"jamId"`
	QueueVersion int64       `json:"queueVersion"`
	Items        []QueueItem `json:"items"`
}

// SessionStateSnapshot combines lifecycle and queue read models for recovery.
type SessionStateSnapshot struct {
	Session          SessionSnapshot `json:"session"`
	Queue            QueueSnapshot   `json:"queue"`
	AggregateVersion int64           `json:"aggregateVersion"`
}

// AddQueueItemRequest defines request payload for queue-add endpoint.
type AddQueueItemRequest struct {
	TrackID        string `json:"trackId"`
	AddedBy        string `json:"addedBy"`
	IdempotencyKey string `json:"idempotencyKey"`
}

// RemoveQueueItemRequest defines request payload for queue-remove endpoint.
type RemoveQueueItemRequest struct {
	ItemID string `json:"itemId"`
}

// ReorderQueueRequest defines request payload for queue-reorder endpoint.
type ReorderQueueRequest struct {
	ItemIDs              []string `json:"itemIds"`
	ExpectedQueueVersion int64    `json:"expectedQueueVersion"`
}
