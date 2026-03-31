## 1. Service Scaffold and Contracts

- [x] 1.1 Create `jams` microservice module structure (API handlers, domain service, Redis repository, tests).
- [x] 1.2 Define queue command/request-response contracts for add, remove, reorder, and snapshot endpoints.
- [x] 1.3 Wire shared error model integration, including `409 version_conflict` response mapping.

## 2. Redis Data Model and Atomic Mutations

- [x] 2.1 Implement Redis key schema for session queue items, queue metadata, and `queueVersion`.
- [x] 2.2 Implement atomic add/remove/reorder mutation logic (Lua script or `MULTI/EXEC`) with single-step `queueVersion` increment.
- [x] 2.3 Add repository-level guards and validation for stale versions and missing queue items.

## 3. Idempotency and Concurrency Control

- [x] 3.1 Implement idempotency key persistence for queue-add commands scoped by jam session with TTL cleanup.
- [x] 3.2 Return prior committed result for duplicate add requests with the same idempotency key.
- [x] 3.3 Enforce optimistic concurrency for reorder via `expectedQueueVersion` and reject stale requests.

## 4. API Endpoints and Snapshot Read Model

- [x] 4.1 Implement add/remove/reorder endpoints that call the domain service and return updated queue snapshots.
- [x] 4.2 Implement snapshot endpoint returning ordered queue items with current `queueVersion`.
- [x] 4.3 Add request validation and consistent error responses for invalid input and backend failures.
- [x] 4.4 Add/Update swagger with new APIs.
## 5. Verification and Rollout Readiness

- [x] 5.1 Add unit tests for command handlers and Redis repository behavior (including idempotency paths).
- [x] 5.2 Add concurrent write integration tests proving deterministic queue state and monotonic versioning.
- [x] 5.3 Add API contract tests for `409 version_conflict` and snapshot consistency after mutations.
- [x] 5.4 Document feature-flagged rollout and rollback steps for production deployment.

- [x] 5.4 Document feature-flagged rollout and rollback steps for production deployment.
