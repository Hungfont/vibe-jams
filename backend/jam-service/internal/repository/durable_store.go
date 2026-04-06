package repository

import (
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"os"
	"path/filepath"

	"video-streaming/backend/jams/internal/model"
)

type persistedQueueRepository struct {
	NextSessionSequence int64                             `json:"nextSessionSequence"`
	Jams                map[string]persistedJamQueueState `json:"jams"`
}

type persistedJamQueueState struct {
	Items             []model.QueueItem              `json:"items"`
	QueueVersion      int64                          `json:"queueVersion"`
	NextItemSequence  int64                          `json:"nextItemSequence"`
	IdempotencyResult map[string]model.QueueSnapshot `json:"idempotencyResult"`
	SessionVersion    int64                          `json:"sessionVersion"`
	Status            model.SessionStatus            `json:"status"`
	HostUserID        string                         `json:"hostUserId"`
	Participants      map[string]model.SessionRole   `json:"participants"`
	MutedUsers        map[string]bool                `json:"mutedUsers"`
	KickedUsers       map[string]bool                `json:"kickedUsers"`
	EndCause          string                         `json:"endCause"`
	EndedBy           string                         `json:"endedBy"`
}

func (r *RedisQueueRepository) loadDurableStateLocked() error {
	if r.storagePath == "" {
		return nil
	}

	data, err := os.ReadFile(r.storagePath)
	if err != nil {
		if stdErrors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read durable queue state: %w", err)
	}

	var persisted persistedQueueRepository
	if err := json.Unmarshal(data, &persisted); err != nil {
		return fmt.Errorf("unmarshal durable queue state: %w", err)
	}

	r.nextSessionSequence = persisted.NextSessionSequence
	r.jams = make(map[string]*jamQueueState, len(persisted.Jams))
	for jamID, state := range persisted.Jams {
		idempotency := make(map[string]addResult, len(state.IdempotencyResult))
		for key, snapshot := range state.IdempotencyResult {
			idempotency[key] = addResult{snapshot: cloneSnapshot(snapshot)}
		}

		participants := make(map[string]model.SessionRole, len(state.Participants))
		for userID, role := range state.Participants {
			participants[userID] = role
		}
		mutedUsers := make(map[string]bool, len(state.MutedUsers))
		for userID, muted := range state.MutedUsers {
			mutedUsers[userID] = muted
		}
		kickedUsers := make(map[string]bool, len(state.KickedUsers))
		for userID, kicked := range state.KickedUsers {
			kickedUsers[userID] = kicked
		}

		r.jams[jamID] = &jamQueueState{
			items:             cloneItems(state.Items),
			queueVersion:      state.QueueVersion,
			nextItemSequence:  state.NextItemSequence,
			idempotencyResult: idempotency,
			sessionVersion:    state.SessionVersion,
			status:            state.Status,
			hostUserID:        state.HostUserID,
			participants:      participants,
			mutedUsers:        mutedUsers,
			kickedUsers:       kickedUsers,
			endCause:          state.EndCause,
			endedBy:           state.EndedBy,
		}
	}

	return nil
}

func (r *RedisQueueRepository) saveDurableStateLocked() error {
	if r.storagePath == "" {
		return nil
	}

	persisted := persistedQueueRepository{
		NextSessionSequence: r.nextSessionSequence,
		Jams:                make(map[string]persistedJamQueueState, len(r.jams)),
	}

	for jamID, state := range r.jams {
		idempotency := make(map[string]model.QueueSnapshot, len(state.idempotencyResult))
		for key, result := range state.idempotencyResult {
			idempotency[key] = cloneSnapshot(result.snapshot)
		}

		participants := make(map[string]model.SessionRole, len(state.participants))
		for userID, role := range state.participants {
			participants[userID] = role
		}

		mutedUsers := make(map[string]bool, len(state.mutedUsers))
		for userID, muted := range state.mutedUsers {
			mutedUsers[userID] = muted
		}

		kickedUsers := make(map[string]bool, len(state.kickedUsers))
		for userID, kicked := range state.kickedUsers {
			kickedUsers[userID] = kicked
		}

		persisted.Jams[jamID] = persistedJamQueueState{
			Items:             cloneItems(state.items),
			QueueVersion:      state.queueVersion,
			NextItemSequence:  state.nextItemSequence,
			IdempotencyResult: idempotency,
			SessionVersion:    state.sessionVersion,
			Status:            state.status,
			HostUserID:        state.hostUserID,
			Participants:      participants,
			MutedUsers:        mutedUsers,
			KickedUsers:       kickedUsers,
			EndCause:          state.endCause,
			EndedBy:           state.endedBy,
		}
	}

	body, err := json.Marshal(persisted)
	if err != nil {
		return fmt.Errorf("marshal durable queue state: %w", err)
	}

	dir := filepath.Dir(r.storagePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create durable queue state dir: %w", err)
		}
	}

	if err := os.WriteFile(r.storagePath, body, 0o644); err != nil {
		return fmt.Errorf("write durable queue state: %w", err)
	}

	return nil
}
