package fanout

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"video-streaming/backend/rt-gateway/internal/metrics"
	"video-streaming/backend/rt-gateway/internal/model"
	sharedevent "video-streaming/backend/shared/event"
)

// SnapshotFetcher obtains authoritative jam state for gap recovery.
type SnapshotFetcher interface {
	FetchSessionState(ctx context.Context, sessionID string) (model.SessionStateSnapshot, error)
}

// Processor applies ordering/gap-recovery rules before fanout.
type Processor struct {
	hub      *Hub
	metrics  *metrics.Registry
	snapshot SnapshotFetcher

	recoveryMaxRetries int
	recoveryBackoff    time.Duration

	recoveryLocks sync.Map
}

// NewProcessor creates a fanout event processor.
func NewProcessor(hub *Hub, registry *metrics.Registry, fetcher SnapshotFetcher, recoveryMaxRetries int, recoveryBackoff time.Duration) *Processor {
	if recoveryMaxRetries < 0 {
		recoveryMaxRetries = 0
	}
	if recoveryBackoff <= 0 {
		recoveryBackoff = 100 * time.Millisecond
	}
	return &Processor{
		hub:                hub,
		metrics:            registry,
		snapshot:           fetcher,
		recoveryMaxRetries: recoveryMaxRetries,
		recoveryBackoff:    recoveryBackoff,
	}
}

// HandleEnvelope processes one session-scoped event envelope.
func (p *Processor) HandleEnvelope(ctx context.Context, envelope sharedevent.Envelope) error {
	if envelope.SessionID == "" {
		return fmt.Errorf("sessionId is required")
	}
	if envelope.AggregateVersion <= 0 {
		return fmt.Errorf("aggregateVersion must be positive")
	}

	start := time.Now()
	if !envelope.OccurredAt.IsZero() && p.metrics != nil {
		p.metrics.ObserveConsumerLag(time.Since(envelope.OccurredAt))
	}

	sessionID := envelope.SessionID
	lastVersion := p.hub.LastVersion(sessionID)

	switch {
	case envelope.AggregateVersion <= lastVersion:
		if p.metrics != nil {
			p.metrics.IncDuplicate()
		}
		return nil
	case envelope.AggregateVersion == lastVersion+1:
		if err := p.broadcastEnvelope(envelope, false); err != nil {
			return err
		}
		p.hub.SetLastVersion(sessionID, envelope.AggregateVersion)
		if p.metrics != nil {
			p.metrics.IncFanout()
			p.metrics.ObserveFanoutLatency(time.Since(start))
		}
		return nil
	default:
		if p.metrics != nil {
			p.metrics.IncGapDetected()
		}
		if err := p.recoverGap(ctx, sessionID); err != nil {
			if p.metrics != nil {
				p.metrics.IncRecoveryFailure()
			}
			return err
		}

		recoveredVersion := p.hub.LastVersion(sessionID)
		if envelope.AggregateVersion <= recoveredVersion {
			if p.metrics != nil {
				p.metrics.IncDuplicate()
			}
			return nil
		}
		if envelope.AggregateVersion != recoveredVersion+1 {
			return fmt.Errorf("recovery incomplete for session %s: recovered=%d incoming=%d", sessionID, recoveredVersion, envelope.AggregateVersion)
		}

		if err := p.broadcastEnvelope(envelope, false); err != nil {
			return err
		}
		p.hub.SetLastVersion(sessionID, envelope.AggregateVersion)
		if p.metrics != nil {
			p.metrics.IncFanout()
			p.metrics.ObserveFanoutLatency(time.Since(start))
		}
		return nil
	}
}

// HandleReconnect emits snapshot fallback when cursor is stale.
func (p *Processor) HandleReconnect(ctx context.Context, sessionID string, lastSeenVersion int64) ([]byte, bool, error) {
	currentVersion := p.hub.LastVersion(sessionID)
	if lastSeenVersion >= currentVersion {
		return nil, false, nil
	}
	if p.snapshot == nil {
		return nil, false, fmt.Errorf("snapshot fetcher is not configured")
	}

	start := time.Now()
	snapshot, err := p.snapshot.FetchSessionState(ctx, sessionID)
	if err != nil {
		if p.metrics != nil {
			p.metrics.IncRecoveryFailure()
		}
		return nil, false, fmt.Errorf("fetch snapshot for reconnect: %w", err)
	}
	if p.metrics != nil {
		p.metrics.ObserveSnapshotLatency(time.Since(start))
	}
	if snapshot.AggregateVersion > currentVersion {
		p.hub.SetLastVersion(sessionID, snapshot.AggregateVersion)
	}

	payload, err := marshalOutbound(model.OutboundEvent{
		EventType:        "jam.session.snapshot",
		SessionID:        sessionID,
		AggregateVersion: snapshot.AggregateVersion,
		OccurredAt:       time.Now().UTC(),
		Payload:          snapshot,
		Recovery:         true,
	})
	if err != nil {
		return nil, false, fmt.Errorf("marshal reconnect snapshot: %w", err)
	}
	return payload, true, nil
}

func (p *Processor) recoverGap(ctx context.Context, sessionID string) error {
	if p.snapshot == nil {
		return fmt.Errorf("snapshot fetcher is not configured")
	}

	lock := p.sessionLock(sessionID)
	lock.Lock()
	defer lock.Unlock()

	var (
		snapshot model.SessionStateSnapshot
		err      error
	)

	for attempt := 0; attempt <= p.recoveryMaxRetries; attempt++ {
		start := time.Now()
		snapshot, err = p.snapshot.FetchSessionState(ctx, sessionID)
		if p.metrics != nil {
			p.metrics.ObserveSnapshotLatency(time.Since(start))
		}
		if err == nil {
			break
		}
		if attempt == p.recoveryMaxRetries {
			return fmt.Errorf("fetch snapshot after %d attempts: %w", attempt+1, err)
		}

		timer := time.NewTimer(p.recoveryBackoff * time.Duration(attempt+1))
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}

	payload, err := marshalOutbound(model.OutboundEvent{
		EventType:        "jam.session.snapshot",
		SessionID:        sessionID,
		AggregateVersion: snapshot.AggregateVersion,
		OccurredAt:       time.Now().UTC(),
		Payload:          snapshot,
		Recovery:         true,
	})
	if err != nil {
		return fmt.Errorf("marshal snapshot outbound: %w", err)
	}

	p.hub.Broadcast(sessionID, payload)
	p.hub.SetLastVersion(sessionID, snapshot.AggregateVersion)
	if p.metrics != nil {
		p.metrics.IncRecoverySuccess()
	}
	return nil
}

func (p *Processor) sessionLock(sessionID string) *sync.Mutex {
	lock, _ := p.recoveryLocks.LoadOrStore(sessionID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (p *Processor) broadcastEnvelope(envelope sharedevent.Envelope, recovery bool) error {
	outbound := model.OutboundEvent{
		EventType:        envelope.EventType,
		SessionID:        envelope.SessionID,
		AggregateVersion: envelope.AggregateVersion,
		OccurredAt:       envelope.OccurredAt,
		Payload:          json.RawMessage(envelope.Payload),
		Recovery:         recovery,
	}
	payload, err := marshalOutbound(outbound)
	if err != nil {
		return fmt.Errorf("marshal outbound event: %w", err)
	}
	p.hub.Broadcast(envelope.SessionID, payload)
	return nil
}

func marshalOutbound(outbound model.OutboundEvent) ([]byte, error) {
	return json.Marshal(outbound)
}
