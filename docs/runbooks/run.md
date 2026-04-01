# Runbook

## Purpose

This file is the execution source of truth for validated backend flows.
It captures only behaviors already implemented and testable in the repository.

## Usage Rules

1. Check this file before implementing new logic.
2. Reuse an existing flow pattern when possible.
3. If behavior changes, update the matching flow in this file in the same change.
4. Keep instructions deterministic and step-based.

## Handbook / Test Flows

### Flow 1: Jam Auth Entitlement Guard (Create and End)

Steps:
1. Send `POST /api/v1/jams/create` or `POST /api/v1/jams/{jamId}/end` with bearer token.
2. Validate claims contract: `userId`, `plan`, `sessionState`.
3. Reject invalid auth context before business logic.
4. Enforce premium entitlement for protected operations.

Expected outcome:
1. Missing/invalid token or invalid session state returns `401 unauthorized`.
2. Non-premium valid user returns `403 premium_required`.
3. Premium valid user proceeds to jam business handler.

Edge cases:
1. Unknown validator error is mapped to `401 unauthorized`.

Validation evidence:
1. `go test ./backend/jam-service/internal/handler -run TestCreateAndEndAuthorizationMatrix -count=1`
2. `go test ./backend/shared/auth -count=1`

### Flow 2: Jam Session Lifecycle and Ended-State Write Blocking

Steps:
1. Create session with host identity.
2. Join and leave participants via lifecycle endpoints.
3. End session explicitly by host or implicitly by host leave.
4. Attempt queue writes after session ended.

Expected outcome:
1. Session transitions are deterministic: create -> active, host leave/end -> ended.
2. Non-host explicit end is rejected with host-only error.
3. Queue/playback writes after ended state are rejected with `session_ended` behavior.

Edge cases:
1. Concurrent join/leave operations preserve consistent active session state for remaining host.

Validation evidence:
1. `go test ./backend/jam-service/internal/repository -run TestSessionLifecycle -count=1`
2. `go test ./backend/jam-service/internal/handler -run TestSessionLifecycleJoinLeaveAndHostLeaveEnds -count=1`

### Flow 3: Queue Command Handling (Atomic, Idempotent, Versioned)

Steps:
1. Execute queue add/remove/reorder for active jam.
2. Enforce idempotency on add using `idempotencyKey`.
3. Enforce optimistic concurrency for reorder using `expectedQueueVersion`.
4. Publish queue/session events for accepted mutations.

Expected outcome:
1. Successful mutation increments `queueVersion` exactly once.
2. Duplicate add with same key does not duplicate queue item.
3. Stale reorder returns deterministic `409 version_conflict`.

Edge cases:
1. Ended session rejects writes without state mutation.

Validation evidence:
1. `go test ./backend/jam-service/internal/repository -count=1`
2. `go test ./backend/jam-service/internal/kafka -count=1`

### Flow 4: Playback Command Pipeline (Host-only, Version-safe, Event-emitting)

Steps:
1. Receive playback command for session endpoint.
2. Validate auth and host ownership.
3. Validate session active state and queue version staleness.
4. Apply command and publish `jam.playback.events` envelope on success.

Expected outcome:
1. Accepted host command returns `202/accepted` behavior and emits one playback event keyed by `sessionId`.
2. Non-host, stale version, unauthorized, and ended session are rejected deterministically.
3. Rejected commands do not publish playback events.

Edge cases:
1. Missing session ID for producer publish is rejected pre-publish.

Validation evidence:
1. `go test ./backend/playback-service/internal/handler -run TestPlaybackCommand -count=1`
2. `go test ./backend/playback-service/internal/kafka -count=1`

### Flow 5: Kafka Event Foundation Baseline

Steps:
1. Provision topics and ACLs for Phase 1.
2. Validate partition and retention settings.
3. Validate producer/consumer ACL expectations.
4. Validate shared envelope serialization and parsing contract.

Expected outcome:
1. Required topics exist with expected policy values.
2. Producer and consumer keying follows contract.
3. Invalid envelopes are rejected before publish.

Edge cases:
1. Missing topic or ACL mismatch fails baseline validation deterministically.

Validation evidence:
1. `go test ./backend/shared/kafka -count=1`
2. `go test ./backend/shared/event -count=1`
3. See `docs/runbooks/kafka-foundation-rollout-checklist.md`.

### Flow 6: Realtime Fanout (Ordered Broadcast and Gap Recovery)

Steps:
1. Client subscribes websocket room by `sessionId`.
2. Fanout consumer reads queue/playback streams for session.
3. Processor enforces monotonic `aggregateVersion`.
4. On version gap, fetch snapshot from jam state endpoint and broadcast recovery snapshot.
5. On reconnect with stale cursor, send snapshot fallback before incremental updates.

Expected outcome:
1. In-order events are broadcast once.
2. Duplicate/stale events are suppressed.
3. Gap detection triggers snapshot recovery and resumes stream.
4. Metrics capture fanout latency, consumer lag, gap count, and recovery outcomes.

Edge cases:
1. Snapshot fetch failures trigger retry/backoff and recovery-failure telemetry.
2. Slow websocket consumers are evicted by bounded buffer policy.

Validation evidence:
1. `go test ./backend/rt-gateway/internal/fanout -count=1`
2. `go test ./backend/rt-gateway/internal/server -count=1`
3. `go test ./backend/rt-gateway/internal/fanout -run TestLoadScenario_FanoutP95LatencyUnderTarget -count=1`
4. See `docs/runbooks/rt-gateway-realtime-fanout-rollout-checklist.md`.
5. See `docs/runbooks/rt-gateway-realtime-fanout-load-evidence.md`.

### Flow 7: Catalog Track Validation for Queue and Playback Commands

Steps:
1. Keep catalog validation disabled by default with `ENABLE_CATALOG_VALIDATION=false` in jam-service and playback-service.
2. For queue add commands, call catalog lookup by `trackId` before any queue mutation.
3. For playback commands that include `trackId`, call catalog lookup before session/host/version transition checks.
4. Map catalog not-found/unavailable outcomes to deterministic API errors and short-circuit command execution.

Expected outcome:
1. Playable tracks continue through command pipeline and mutate state normally.
2. Missing tracks return `track_not_found` and do not mutate queue/playback state.
3. Unavailable tracks return `track_unavailable` and do not mutate queue/playback state.
4. Rejected playback commands publish zero Kafka playback events.

Edge cases:
1. If `trackId` is omitted in playback request, catalog pre-check is skipped and existing behavior is preserved.
2. Catalog timeout or upstream failure fails command fast and does not mutate state.

Rollout checklist:
1. Deploy catalog-service endpoint first, keep toggles off in jam-service and playback-service.
2. Enable `ENABLE_CATALOG_VALIDATION=true` on jam-service canary and monitor add-command p95 latency plus `track_not_found`/`track_unavailable` error rates.
3. Enable `ENABLE_CATALOG_VALIDATION=true` on playback-service canary and monitor command p95 latency plus `track_not_found`/`track_unavailable` rejection rates.
4. Roll back by setting `ENABLE_CATALOG_VALIDATION=false` on affected service and restart pods.

Validation evidence:
1. `cd backend/catalog-service && go test ./...`
2. `cd backend/shared && go test ./...`
3. `cd backend/jam-service && go test ./...`
4. `cd backend/playback-service && go test ./...`
5. `cd backend/jam-service && go test ./internal/handler -run TestAddQueueTrack -count=1`
6. `cd backend/playback-service && go test ./internal/handler -run TestPlaybackCommand_Track -count=1`

## Assumptions Logged

1. This baseline excludes proposed but not-yet-implemented flows.
2. Service contracts are based on current tests/specs in repository at initialization time.

## Update Protocol

1. When `/opsx:apply` introduces a new or modified execution flow, add or refine one flow section above.
2. Update expected outcomes and edge cases based on real tests, not planned behavior.
3. If two flows overlap, merge into one generalized pattern and keep old wording out.
