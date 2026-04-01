package bff

import sharedauth "video-streaming/backend/shared/auth"

// OrchestrateRequest is the BFF request payload for MVP orchestration.
type OrchestrateRequest struct {
	TrackID         string                  `json:"trackId,omitempty"`
	PlaybackCommand *PlaybackCommandRequest `json:"playbackCommand,omitempty"`
}

// PlaybackCommandRequest mirrors playback-service command payload.
type PlaybackCommandRequest struct {
	Command              string `json:"command"`
	TrackID              string `json:"trackId,omitempty"`
	ClientEventID        string `json:"clientEventId"`
	ExpectedQueueVersion int64  `json:"expectedQueueVersion"`
	PositionMS           int64  `json:"positionMs,omitempty"`
}

// PlaybackCommandAccepted mirrors playback-service accepted response.
type PlaybackCommandAccepted struct {
	Accepted bool `json:"accepted"`
}

// SessionStateSnapshot mirrors jam-service /state response schema.
type SessionStateSnapshot struct {
	Session          SessionSnapshot `json:"session"`
	Queue            QueueSnapshot   `json:"queue"`
	AggregateVersion int64           `json:"aggregateVersion"`
}

// SessionSnapshot mirrors jam session metadata.
type SessionSnapshot struct {
	JamID          string               `json:"jamId"`
	Status         string               `json:"status"`
	HostUserID     string               `json:"hostUserId"`
	Participants   []SessionParticipant `json:"participants"`
	SessionVersion int64                `json:"sessionVersion"`
	EndCause       string               `json:"endCause,omitempty"`
	EndedBy        string               `json:"endedBy,omitempty"`
}

// SessionParticipant mirrors participant schema.
type SessionParticipant struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// QueueSnapshot mirrors jam queue state.
type QueueSnapshot struct {
	JamID        string      `json:"jamId"`
	QueueVersion int64       `json:"queueVersion"`
	Items        []QueueItem `json:"items"`
}

// QueueItem mirrors queue item schema.
type QueueItem struct {
	ItemID  string `json:"itemId"`
	TrackID string `json:"trackId"`
	AddedBy string `json:"addedBy"`
}

// LookupResponse mirrors shared catalog response schema.
type LookupResponse struct {
	TrackID    string `json:"trackId"`
	IsPlayable bool   `json:"isPlayable"`
	ReasonCode string `json:"reasonCode,omitempty"`
	Title      string `json:"title,omitempty"`
	Artist     string `json:"artist,omitempty"`
}

// OrchestrateData is aggregated BFF response payload.
type OrchestrateData struct {
	Claims             sharedauth.Claims        `json:"claims"`
	SessionState       SessionStateSnapshot     `json:"sessionState"`
	Track              *LookupResponse          `json:"track,omitempty"`
	Playback           *PlaybackCommandAccepted `json:"playback,omitempty"`
	Partial            bool                     `json:"partial"`
	DependencyStatuses map[string]string        `json:"dependencyStatuses"`
	Issues             []DependencyIssue        `json:"issues,omitempty"`
}

// DependencyIssue captures optional dependency degradation details.
type DependencyIssue struct {
	Dependency string `json:"dependency"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

// Envelope is the standardized BFF response envelope.
type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

// ErrorBody is deterministic client-facing error detail.
type ErrorBody struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Dependency string `json:"dependency,omitempty"`
}
