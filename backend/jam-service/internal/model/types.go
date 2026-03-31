package model

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
