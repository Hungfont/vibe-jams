## Why

The platform needs a dedicated `jams` microservice to own queue command handling for jam sessions and provide deterministic behavior under concurrent writes. Building this now unblocks MVP delivery for queue add/remove/reorder operations with clear conflict and idempotency semantics.

## What Changes

- Add a new `jams` microservice with API endpoints for queue add, remove, reorder, and snapshot retrieval.
- Implement atomic queue mutations backed by Redis list/hash primitives with a monotonically increasing `queueVersion`.
- Add idempotency key handling for queue-add commands to prevent duplicate queue items on retries.
- Return `409 version_conflict` for stale reorder requests where the client version is behind current `queueVersion`.
- Add concurrency and contract tests validating deterministic queue state and latest-version snapshots.

## Capabilities

### New Capabilities
- `jams-queue-command-handling`: Handles add/remove/reorder queue commands with atomic writes, idempotency for add, version conflict detection, and queue snapshot reads.

### Modified Capabilities
- None.

## Impact

- Affected code: new service module(s) under backend microservices for `jams`, queue command handlers, Redis repository layer, and API error mapping.
- APIs: new `jams` queue command/read endpoints and standardized error responses including `version_conflict`.
- Dependencies: Redis for queue state/version storage and shared API error model package.
- Systems: session/jam flow will call `jams` service for queue mutations instead of handling queue writes elsewhere.
