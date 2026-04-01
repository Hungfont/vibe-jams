package fanout

import (
	"sync"

	"video-streaming/backend/rt-gateway/internal/metrics"
)

// Subscriber receives fanout events for one websocket connection.
type Subscriber struct {
	Send chan []byte
}

// Hub manages room membership and broadcast behavior.
type Hub struct {
	mu         sync.RWMutex
	rooms      map[string]map[*Subscriber]struct{}
	lastByRoom map[string]int64
	maxBuffer  int
	metrics    *metrics.Registry
}

// NewHub creates a room fanout hub.
func NewHub(maxBuffer int, registry *metrics.Registry) *Hub {
	if maxBuffer <= 0 {
		maxBuffer = 1
	}
	return &Hub{
		rooms:      make(map[string]map[*Subscriber]struct{}),
		lastByRoom: make(map[string]int64),
		maxBuffer:  maxBuffer,
		metrics:    registry,
	}
}

// AddSubscriber registers a new subscriber for one session room.
func (h *Hub) AddSubscriber(sessionID string) *Subscriber {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[sessionID]
	if !ok {
		room = make(map[*Subscriber]struct{})
		h.rooms[sessionID] = room
	}

	subscriber := &Subscriber{Send: make(chan []byte, h.maxBuffer)}
	room[subscriber] = struct{}{}
	return subscriber
}

// RemoveSubscriber unregisters a subscriber and closes its send channel.
func (h *Hub) RemoveSubscriber(sessionID string, subscriber *Subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[sessionID]
	if !ok {
		return
	}
	if _, exists := room[subscriber]; !exists {
		return
	}
	delete(room, subscriber)
	close(subscriber.Send)

	if len(room) == 0 {
		delete(h.rooms, sessionID)
	}
}

// Broadcast publishes one payload to all subscribers in a session room.
func (h *Hub) Broadcast(sessionID string, payload []byte) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[sessionID]
	if !ok {
		return 0
	}

	delivered := 0
	for subscriber := range room {
		select {
		case subscriber.Send <- payload:
			delivered++
		default:
			delete(room, subscriber)
			close(subscriber.Send)
			if h.metrics != nil {
				h.metrics.IncSlowConsumer()
			}
		}
	}

	if len(room) == 0 {
		delete(h.rooms, sessionID)
	}
	return delivered
}

// LastVersion returns last broadcast aggregateVersion for the session room.
func (h *Hub) LastVersion(sessionID string) int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastByRoom[sessionID]
}

// SetLastVersion stores last broadcast aggregateVersion for the session room.
func (h *Hub) SetLastVersion(sessionID string, version int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastByRoom[sessionID] = version
}
