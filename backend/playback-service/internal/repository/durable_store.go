package repository

import (
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"os"
	"path/filepath"
)

type persistedPlaybackRepository struct {
	Sessions map[string]persistedSessionState `json:"sessions"`
}

type persistedSessionState struct {
	HostUserID       string `json:"hostUserId"`
	Status           string `json:"status"`
	QueueVersion     int64  `json:"queueVersion"`
	AggregateVersion int64  `json:"aggregateVersion"`
	PlaybackState    string `json:"playbackState"`
	PositionMS       int64  `json:"positionMs"`
}

func (r *RedisPlaybackRepository) loadDurableStateLocked() error {
	if r.storagePath == "" {
		return nil
	}

	data, err := os.ReadFile(r.storagePath)
	if err != nil {
		if stdErrors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read durable playback state: %w", err)
	}

	var persisted persistedPlaybackRepository
	if err := json.Unmarshal(data, &persisted); err != nil {
		return fmt.Errorf("unmarshal durable playback state: %w", err)
	}

	r.sessions = make(map[string]*sessionState, len(persisted.Sessions))
	for sessionID, state := range persisted.Sessions {
		r.sessions[sessionID] = &sessionState{
			hostUserID:       state.HostUserID,
			status:           state.Status,
			queueVersion:     state.QueueVersion,
			aggregateVersion: state.AggregateVersion,
			playbackState:    state.PlaybackState,
			positionMS:       state.PositionMS,
		}
	}

	return nil
}

func (r *RedisPlaybackRepository) saveDurableStateLocked() error {
	if r.storagePath == "" {
		return nil
	}

	persisted := persistedPlaybackRepository{
		Sessions: make(map[string]persistedSessionState, len(r.sessions)),
	}
	for sessionID, state := range r.sessions {
		persisted.Sessions[sessionID] = persistedSessionState{
			HostUserID:       state.hostUserID,
			Status:           state.status,
			QueueVersion:     state.queueVersion,
			AggregateVersion: state.aggregateVersion,
			PlaybackState:    state.playbackState,
			PositionMS:       state.positionMS,
		}
	}

	body, err := json.Marshal(persisted)
	if err != nil {
		return fmt.Errorf("marshal durable playback state: %w", err)
	}

	dir := filepath.Dir(r.storagePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create durable playback state dir: %w", err)
		}
	}

	if err := os.WriteFile(r.storagePath, body, 0o644); err != nil {
		return fmt.Errorf("write durable playback state: %w", err)
	}

	return nil
}
