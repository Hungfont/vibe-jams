package model

import (
	"errors"
	"fmt"
	"strings"
)

const (
	CommandPlay  = "play"
	CommandPause = "pause"
	CommandNext  = "next"
	CommandPrev  = "prev"
	CommandSeek  = "seek"
)

var (
	// ErrInvalidCommand indicates unsupported playback command values.
	ErrInvalidCommand = errors.New("invalid command")
)

// PlaybackCommandRequest defines the payload for command endpoint.
type PlaybackCommandRequest struct {
	Command              string `json:"command"`
	TrackID              string `json:"trackId,omitempty"`
	ClientEventID        string `json:"clientEventId"`
	ExpectedQueueVersion int64  `json:"expectedQueueVersion"`
	PositionMS           int64  `json:"positionMs,omitempty"`
}

// Validate checks request fields for command execution.
func (r PlaybackCommandRequest) Validate() error {
	command := strings.ToLower(strings.TrimSpace(r.Command))
	if command == "" {
		return fmt.Errorf("%w: command is required", ErrInvalidCommand)
	}
	switch command {
	case CommandPlay, CommandPause, CommandNext, CommandPrev, CommandSeek:
	default:
		return fmt.Errorf("%w: unsupported value %q", ErrInvalidCommand, r.Command)
	}
	if strings.TrimSpace(r.ClientEventID) == "" {
		return errors.New("clientEventId is required")
	}
	if r.ExpectedQueueVersion <= 0 {
		return errors.New("expectedQueueVersion must be positive")
	}
	if command == CommandSeek && r.PositionMS < 0 {
		return errors.New("positionMs must be non-negative for seek command")
	}
	return nil
}

// CommandAcceptedResponse is returned when command is accepted.
type CommandAcceptedResponse struct {
	Accepted      bool  `json:"accepted"`
	QueueVersion  int64 `json:"queueVersion"`
	PlaybackEpoch int64 `json:"playbackEpoch"`
}

// PlaybackTransition stores computed transition metadata after command execution.
type PlaybackTransition struct {
	SessionID        string `json:"sessionId"`
	Command          string `json:"command"`
	State            string `json:"state"`
	PositionMS       int64  `json:"positionMs"`
	QueueVersion     int64  `json:"queueVersion"`
	PlaybackEpoch    int64  `json:"playbackEpoch"`
	AggregateVersion int64  `json:"aggregateVersion"`
	ActorUserID      string `json:"actorUserId"`
	ClientEventID    string `json:"clientEventId"`
}
