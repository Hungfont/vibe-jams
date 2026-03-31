# Kafka Event Envelope Contract

This package defines the shared envelope used by all Phase 1 producers.

## Required Fields

- `eventId`: unique event identifier for deduplication
- `eventType`: event name such as `jam.queue.item.added`
- `occurredAt`: RFC3339 timestamp at producer commit point
- `payload`: event payload encoded as JSON object

## Session-Scoped Rules

For topics keyed by `sessionId` (`jam.session.events`, `jam.queue.events`, `jam.playback.events`):

- `sessionId` is required and MUST be used as Kafka message key
- `aggregateVersion` is required and MUST increase monotonically per `sessionId`

## Analytics Rules

For `analytics.user.actions`:

- Kafka key MUST be `userId`
- `sessionId` and `aggregateVersion` are optional (include when available)

## Producer Validation

Producer adapters fail fast before publish when:

- required key fields are missing (`sessionId` or `userId` based on topic)
- envelope required metadata is missing
- session-scoped event has non-positive `aggregateVersion`
