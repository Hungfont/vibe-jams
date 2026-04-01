package repository

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"video-streaming/backend/playback-service/internal/model"
)

var (
	// ErrSessionNotFound indicates a command references unknown session.
	ErrSessionNotFound = errors.New("session not found")
	// ErrSessionEnded indicates command references ended session.
	ErrSessionEnded = errors.New("session ended")
	// ErrInvalidSessionState indicates state mutation attempted on invalid data.
	ErrInvalidSessionState = errors.New("invalid session state")
)

type sessionState struct {
	hostUserID       string
	status           string
	queueVersion     int64
	aggregateVersion int64
	playbackState    string
	positionMS       int64
}

// RedisPlaybackRepository stores session metadata using in-memory maps with Redis-like keys.
type RedisPlaybackRepository struct {
	mu          sync.Mutex
	sessions    map[string]*sessionState
	storagePath string
}

// NewRedisPlaybackRepository builds repository used by playback command pipeline.
func NewRedisPlaybackRepository() *RedisPlaybackRepository {
	return &RedisPlaybackRepository{
		sessions: make(map[string]*sessionState),
	}
}

// NewDurablePlaybackRepository builds a playback repository with file-backed persistence.
func NewDurablePlaybackRepository(storagePath string) (*RedisPlaybackRepository, error) {
	repo := &RedisPlaybackRepository{
		sessions:    make(map[string]*sessionState),
		storagePath: storagePath,
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if err := repo.loadDurableStateLocked(); err != nil {
		return nil, err
	}

	return repo, nil
}

// QueueMetadataKey returns Redis hash key format for queue metadata/version.
func QueueMetadataKey(sessionID string) string {
	return fmt.Sprintf("jams:%s:queue:meta", sessionID)
}

// PlaybackMetadataKey returns Redis hash key format for playback metadata.
func PlaybackMetadataKey(sessionID string) string {
	return fmt.Sprintf("jams:%s:playback:meta", sessionID)
}

// SessionMetadataKey returns Redis hash key format for session metadata.
func SessionMetadataKey(sessionID string) string {
	return fmt.Sprintf("jams:%s:session:meta", sessionID)
}

// SeedSession creates or updates one in-memory session for tests/local bootstrapping.
func (r *RedisPlaybackRepository) SeedSession(sessionID string, hostUserID string, queueVersion int64) error {
	if strings.TrimSpace(sessionID) == "" || strings.TrimSpace(hostUserID) == "" || queueVersion <= 0 {
		return ErrInvalidSessionState
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.sessions[sessionID]
	if !ok {
		current = &sessionState{
			playbackState: "paused",
			status:        "active",
		}
		r.sessions[sessionID] = current
	}
	current.hostUserID = hostUserID
	current.queueVersion = queueVersion
	if err := r.saveDurableStateLocked(); err != nil {
		return err
	}
	return nil
}

// EndSession marks one session as ended for command rejection checks.
func (r *RedisPlaybackRepository) EndSession(sessionID string, endedBy string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}
	if strings.TrimSpace(endedBy) == "" {
		return ErrInvalidSessionState
	}
	current.status = "ended"
	if err := r.saveDurableStateLocked(); err != nil {
		return err
	}
	return nil
}

// HostUserID returns session host user identifier.
func (r *RedisPlaybackRepository) HostUserID(sessionID string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.sessions[sessionID]
	if !ok {
		return "", ErrSessionNotFound
	}
	return current.hostUserID, nil
}

// QueueVersion returns current queue version for one session.
func (r *RedisPlaybackRepository) QueueVersion(sessionID string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.sessions[sessionID]
	if !ok {
		return 0, ErrSessionNotFound
	}
	return current.queueVersion, nil
}

// SessionStatus returns current lifecycle status for one session.
func (r *RedisPlaybackRepository) SessionStatus(sessionID string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.sessions[sessionID]
	if !ok {
		return "", ErrSessionNotFound
	}
	return current.status, nil
}

// ApplyCommand mutates playback state and returns transition metadata.
func (r *RedisPlaybackRepository) ApplyCommand(sessionID string, command string, positionMS int64, actorUserID string, clientEventID string) (model.PlaybackTransition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.sessions[sessionID]
	if !ok {
		return model.PlaybackTransition{}, ErrSessionNotFound
	}
	if current.status == "ended" {
		return model.PlaybackTransition{}, ErrSessionEnded
	}

	switch command {
	case model.CommandPlay:
		current.playbackState = "playing"
	case model.CommandPause:
		current.playbackState = "paused"
	case model.CommandNext, model.CommandPrev:
		current.playbackState = "playing"
		current.positionMS = 0
	case model.CommandSeek:
		current.positionMS = positionMS
	default:
		return model.PlaybackTransition{}, model.ErrInvalidCommand
	}

	current.aggregateVersion++
	if err := r.saveDurableStateLocked(); err != nil {
		return model.PlaybackTransition{}, err
	}

	return model.PlaybackTransition{
		SessionID:        sessionID,
		Command:          command,
		State:            current.playbackState,
		PositionMS:       current.positionMS,
		QueueVersion:     current.queueVersion,
		AggregateVersion: current.aggregateVersion,
		ActorUserID:      actorUserID,
		ClientEventID:    clientEventID,
	}, nil
}
