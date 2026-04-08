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

### Flow 16: Stronger Concurrency Guards for Queue and Playback (ES-P2-003)

Steps:
1. Submit queue remove or reorder writes with `expectedQueueVersion` from latest known snapshot.
2. For stale queue writes, return `409 version_conflict` with `error.retry.currentQueueVersion` guidance.
3. Submit playback command writes with `expectedQueueVersion` and host authorization.
4. For stale playback writes, return `409 version_conflict` with `error.retry.currentQueueVersion` and `error.retry.playbackEpoch`.
5. On accepted playback updates, include `queueVersion` and `playbackEpoch` in command response and emitted playback event payload.
6. Frontend reconciliation applies retry guidance, refreshes authoritative snapshot, and retries versioned queue mutation with updated version.

Expected outcome:
1. Reorder/remove stale writes are deterministically rejected with conflict guidance and no queue mutation.
2. Remove/reorder writes with matching expected version commit atomically and increment queue version exactly once.
3. Playback stale writes are deterministically rejected with queue/playback retry metadata.
4. Accepted playback updates always include both `queueVersion` and `playbackEpoch`.

Edge cases:
1. Missing expected queue version in remove/reorder payload is treated as stale/default conflict and returns deterministic `version_conflict` response.
2. Repeated version conflicts keep reconciliation deterministic by applying latest retry metadata before next retry attempt.

Validation evidence:
1. `cd backend/jam-service && go test ./...`
2. `cd backend/playback-service && go test ./...`
3. `cd frontend && bun run test src/components/jam/jam-room-client.test.tsx`

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
3. Treat catalog policy-restricted lookup outcomes as deterministic `track_restricted` rejection when validation is enabled.
4. For playback commands that include `trackId`, call catalog lookup before session/host/version transition checks.
5. Map catalog not-found/unavailable/restricted outcomes to deterministic API errors and short-circuit command execution.

Expected outcome:
1. Playable tracks continue through command pipeline and mutate state normally.
2. Missing tracks return `track_not_found` and do not mutate queue/playback state.
3. Unavailable tracks return `track_unavailable` and do not mutate queue/playback state.
4. Restricted tracks return `track_restricted` and do not mutate queue state.
5. Rejected playback commands publish zero Kafka playback events.
6. When validation is disabled, baseline queue add idempotency/version behavior remains unchanged.

Edge cases:
1. If `trackId` is omitted in playback request, catalog pre-check is skipped and existing behavior is preserved.
2. Catalog timeout or upstream failure fails command fast and does not mutate state.
3. Catalog lookup payload keeps stable policy metadata fields (`policyStatus`, `policyReason`) for policy-on/off caller compatibility.

Rollout checklist:
1. Deploy catalog-service endpoint first, keep toggles off in jam-service and playback-service.
2. Enable `ENABLE_CATALOG_VALIDATION=true` on jam-service canary and monitor add-command p95 latency plus `track_not_found`/`track_unavailable` error rates.
3. Enable `ENABLE_CATALOG_VALIDATION=true` on playback-service canary and monitor command p95 latency plus `track_not_found`/`track_unavailable` rejection rates.
4. Roll back by setting `ENABLE_CATALOG_VALIDATION=false` on affected service and restart pods.

Validation evidence:
1. `cd backend/catalog-service && go test ./...`
2. `cd backend/shared && go test ./catalog -count=1`
3. `cd backend/jam-service && go test ./...`
4. `cd backend/playback-service && go test ./...`
5. `cd backend/jam-service && go test ./internal/handler -run TestAddQueueTrack -count=1`
6. `cd backend/playback-service && go test ./internal/handler -run TestPlaybackCommand_Track -count=1`

### Flow 8: Runtime Integration Policy and Origin Hardening (ES-P1-008)

Steps:
1. Load service configs with runtime profile and adapter policy fields (`APP_ENV`, Kafka backend settings, catalog/auth backend settings, websocket allowed origins).
2. In non-test profiles, reject `STATE_STORE_BACKEND=inmemory` and require durable state path configuration.
3. Build jam/playback repositories from durable backends (`redis`/`postgres` runtime modes) and persist state before command completion.
4. Build jam/playback producers and rt-gateway consumer using Kafka transport backend.
5. In profiles where Kafka transport is not selected as default, use shared `NoOpsProducer` and `NoOpsConsumer` from `backend/shared/kafka`.
6. Enforce websocket origin allowlist in rt-gateway handshake before websocket upgrade.
7. Map catalog/auth upstream dependency failures to deterministic dependency-unavailable semantics.
8. Validate restart recovery by re-instantiating repositories from the same durable state file.
9. Validate concurrent command consistency against durable adapters.

Expected outcome:
1. Invalid runtime adapter combinations fail startup deterministically before serving traffic.
2. Local profile applies localhost Kafka fallback and deterministic durable state-store path defaults.
3. Services that use non-Kafka transport defaults reuse shared noop transport implementations with consistent behavior.
4. Session lifecycle, queue snapshot/version, idempotency records, and playback transition sequence survive repository restart.
5. Concurrent queue adds and playback commands preserve consistent durable version progression.
6. Unknown websocket origins are rejected with `403` and `forbidden_origin` payload.
7. Catalog/auth 5xx and transport failures are mapped to deterministic dependency-unavailable errors.

Edge cases:
1. `KAFKA_CONSUMER_BACKEND=noop` is rejected outside test profile.
2. Empty `WS_ALLOWED_ORIGINS` is rejected outside test profile.
3. `STATE_STORE_BACKEND=inmemory` is rejected outside test profile.
4. Idempotent queue-add retry after restart returns cached prior result and does not duplicate items.

Validation evidence:
1. `cd backend/jam-service && go test ./...`
2. `cd backend/playback-service && go test ./...`
3. `cd backend/rt-gateway && go test ./...`
4. `cd backend/shared && go test ./kafka -run TestNoOps -count=1`
5. `cd backend/rt-gateway && go test ./internal/kafka -run TestNoopConsumerStartReturnsContextCanceled -count=1`
6. `cd backend/jam-service && go test ./internal/repository -run TestDurableQueueRepository_RestartRecoversSessionQueueAndIdempotency -count=1`
7. `cd backend/playback-service && go test ./internal/repository -run TestDurablePlaybackRepository_RestartRecoversTransitions -count=1`
8. `cd backend/jam-service && go test ./internal/repository -run TestDurableQueueRepository_ConcurrentAddsRemainConsistent -count=1`
9. `cd backend/playback-service && go test ./internal/repository -run TestDurablePlaybackRepository_ConcurrentCommandsRemainConsistent -count=1`
10. `cd backend/rt-gateway && go test ./internal/server -run TestWebsocketFanout_EndToEndViaConsumerLoop -count=1`

### Flow 9: API-Gateway AuthN + API-Service BFF Orchestration

**Entry point**: `api-gateway` (port 8085) is the sole public ingress. Direct external calls to `api-service` are blocked at the network layer.

Steps:
1. Client sends protected gateway route request (for example `POST /v1/bff/mvp/sessions/{sessionId}/orchestration`) with `Authorization: Bearer <token>` or with `auth_token`/`token` cookie when header auth is absent.
2. api-gateway strips any client-supplied `X-Auth-*` headers (spoof prevention).
3. api-gateway resolves auth credentials by preferring `Authorization`, then falling back to `auth_token`/`token` cookie values.
4. api-gateway calls `POST /internal/v1/auth/validate` on auth-service with the resolved bearer token; applies `GATEWAY_TIMEOUT_AUTH_MS` timeout.
5. On validate success, api-gateway checks `sessionState == "valid"`; injects `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-SessionState`, `X-Auth-Scope` into the forwarded request.
6. api-gateway proxies the request to api-service while preserving `Authorization` for migration compatibility.
7. api-service reads identity from `X-Auth-*` headers (does not call auth-service); rejects orchestration calls if `X-Auth-UserId` is absent or `sessionState` is not `valid`.
8. api-service fetches jam session state from jam-service as a required dependency, forwarding `X-Auth-*` headers.
9. If `trackId` is present, api-service resolves catalog metadata as optional enrichment.
10. Reject request with `400 invalid_input` when `playbackCommand` appears in orchestration payload.
11. Aggregate dependency outputs into one deterministic BFF envelope.

Expected outcome:
1. Successful path returns `200` with `success=true`, normalized `claims` (from `X-Auth-*` headers), jam `sessionState`, and dependency status map.
2. Missing/invalid token → api-gateway returns `401 invalid_token` before forwarding.
3. Missing both `Authorization` and supported auth cookies → api-gateway returns `401 missing_credentials` before forwarding.
4. auth-service timeout → api-gateway returns `503 auth_service_unavailable`.
5. Revoked session (`sessionState != valid`) → api-gateway returns `401 invalid_token`.
6. Required dependency (jam) timeout → api-service returns `503 dependency_timeout`.
7. Optional catalog failure → api-service returns `200` with `data.partial=true`.

Edge cases:
1. Client-injected `X-Auth-UserId` is stripped by api-gateway and replaced with validated value.
2. Request reaching api-service without `X-Auth-UserId` (gateway bypassed) returns `401 unauthorized`.
3. Empty `sessionId` path parameter returns `400 invalid_input`.
4. `playbackCommand` in orchestration body returns `400 invalid_input`.
5. During migration compatibility, `Authorization` is preserved while `X-Auth-*` headers are injected for downstream path transition.
6. When both `Authorization` and auth cookie are present, gateway validates using `Authorization` value first.

Startup order:
1. `auth-service` (port 8081)
2. `api-service` (port 8084, internal only)
3. `api-gateway` (port 8085, public)

Validation evidence:
1. `go test video-streaming/backend/api-gateway/...`
2. `go test video-streaming/backend/api-service/...`
3. `go test video-streaming/backend/api-gateway/internal/gateway -run TestAuthnMiddleware -count=1`
4. `go test video-streaming/backend/api-service/internal/bff -run TestOrchestrationSuccessAcrossDependencies -count=1`
5. `go test video-streaming/backend/api-service/internal/bff -run TestOrchestrationRejectsMissingIdentityHeaders -count=1`
6. `go test video-streaming/backend/api-service/internal/bff -run TestOrchestrationRejectsPlaybackCommandPayload -count=1`
7. `go test video-streaming/backend/api-service/internal/bff -run TestHTTPCatalogClientLookupTrack -count=1`
8. `go test video-streaming/backend/api-gateway/internal/gateway -run TestIntegration_CookieFallback_ProxiesToAPIService -count=1`
9. `go test video-streaming/backend/api-gateway/internal/gateway -run TestIntegration_OpenAPIJSONRoute_IsPublicAndServed -count=1`
10. `go test video-streaming/backend/api-service/internal/bff -run TestOpenAPISpec_IncludesDelegatedBFFRouteFamilies -count=1`
11. `cd frontend && bun run test -- src/app/api/realtime/ws-config/route.test.ts`

### Flow 14: Gateway and API-Service Swagger/OpenAPI Operational Visibility

Steps:
1. Open api-gateway Swagger UI at `GET /swagger` and fetch OpenAPI JSON at `GET /swagger/openapi.json`.
2. Open api-service Swagger UI at `GET /swagger` and fetch OpenAPI JSON at `GET /swagger/openapi.json`.
3. Verify gateway spec includes health and representative delegated ingress route families.
4. Verify api-service spec includes orchestration plus delegated jam/playback/catalog/realtime route families used by BFF-first frontend calls.

Expected outcome:
1. Swagger UI pages render without auth gating on operational doc routes.
2. OpenAPI JSON responses return `200` and include expected route families.
3. Spec coverage reflects current BFF-first routing behavior.

Edge cases:
1. Unsupported methods on `/swagger` or `/swagger/openapi.json` return `405 method not allowed`.
2. OpenAPI marshal failures return deterministic `500` responses.

Validation evidence:
1. `go test video-streaming/backend/api-gateway/internal/gateway -run TestIntegration_SwaggerUIRoute_IsPublicAndServed -count=1`
2. `go test video-streaming/backend/api-gateway/internal/gateway -run TestIntegration_OpenAPIJSONRoute_IsPublicAndServed -count=1`
3. `go test video-streaming/backend/api-service/internal/bff -run TestRouter_SwaggerAndOpenAPIRoutes_AreServed -count=1`

### Flow 10: Frontend Phase 1 Jam UI Routing and API Boundary

Steps:
1. Open Lobby page (`/`) and validate create/join actions call frontend API routes only.
2. Create a jam and verify redirect to `/jam/{jamId}`.
3. On Jam Room page, verify queue/playback/participants/diagnostics sections render from orchestration state.
4. Trigger queue, playback, and catalog interactions and confirm client calls stay within `/api/*` route boundary.
5. Verify non-auth frontend API routes are delegated through `api-gateway -> api-service (BFF)` before downstream service calls.
6. Verify realtime bootstrap uses `/api/realtime/ws-config` and resolves ws config through BFF before websocket connect.
7. Verify websocket connect uses `WS /v1/bff/mvp/realtime/ws` through `api-gateway -> api-service (BFF) -> rt-gateway` and does not target rt-gateway directly.

Expected outcome:
1. Lobby and Jam Room render without compile/runtime errors.
2. Browser-side callers use frontend API routes only (no direct backend service URL usage).
3. Room UI surfaces degraded or blocking errors in diagnostics and alerts when present.
4. Queue and playback controls enforce host/session-ended guard behavior at UI level.
5. Realtime ws bootstrap configuration is sourced through BFF-first HTTP hop and websocket ingress is routed through gateway/BFF proxy path.

Edge cases:
1. Invalid jam ID in lobby join form is rejected client-side.
2. Realtime degraded bootstrap keeps room usable with periodic state refresh fallback.

Validation evidence:
1. `cd frontend && bun lint`
2. `cd frontend && bun build`
3. `cd frontend && bun test`

### Flow 15: Realtime Websocket Ingress Through Gateway/BFF Proxy

Steps:
1. Request realtime bootstrap config via frontend route `GET /api/realtime/ws-config`.
2. Verify returned `wsUrl` points to gateway/BFF websocket path (`/v1/bff/mvp/realtime/ws`) instead of direct rt-gateway `/ws`.
3. Open websocket using returned `wsUrl` with `sessionId` and `lastSeenVersion` query parameters.
4. Verify ingress path is proxied `api-gateway -> api-service (BFF) -> rt-gateway /ws`.

Expected outcome:
1. Frontend never targets direct rt-gateway websocket URL.
2. Realtime fanout semantics (ordered delivery, gap recovery) remain unchanged after proxy ingress.

Edge cases:
1. Missing `sessionId` still fails deterministically at bootstrap route.
2. Non-websocket-compatible method on BFF websocket path is rejected deterministically.

Validation evidence:
1. `go test video-streaming/backend/api-service/internal/bff -run TestProxyHandler_RealtimeWSProxy_RewritesToRTGatewayWSPath -count=1`
2. `go test video-streaming/backend/api-service/internal/bff -run TestProxyHandler_RealtimeWSConfig -count=1`
3. `cd frontend && bun run test -- src/app/api/realtime/ws-config/route.test.ts`

### Flow 11: Moderation Controls (Mute/Kick, Realtime Visibility, Blocked Actions)

Steps:
1. Host sends moderation command through frontend API routes: `POST /api/jam/{jamId}/moderation/mute` or `POST /api/jam/{jamId}/moderation/kick`.
2. Frontend route validates auth claims and forwards command through `api-gateway -> api-service` delegated jam route.
3. jam-service validates host ownership and applies moderation transition:
	- mute: mark target participant as `muted=true`
	- kick: remove target participant from active participants
4. jam-service publishes moderation audit envelope to `jam.moderation.events` with `action`, `actorUserId`, `targetUserId`, `reason`, and `occurredAt`.
5. rt-gateway consumes moderation topic and invokes moderation hook (default no-op) before fanout processing.
6. rt-gateway fans out moderation event and clients refresh state snapshot path for room consistency.
7. Muted/kicked participants attempting blocked queue mutations receive deterministic moderation-blocked response.

Expected outcome:
1. Non-host moderation attempts are rejected deterministically with host-only semantics.
2. Successful moderation commands update participant projection and audit event stream.
3. Room subscribers receive moderation updates through realtime fanout.
4. Muted or kicked users cannot perform blocked queue mutations and receive `moderation_blocked`.

Edge cases:
1. Missing/invalid moderation payload returns deterministic `invalid_input`.
2. Host cannot moderate invalid target identities (for example: missing participant or host self) and receives deterministic request/domain error mapping.
3. Moderation topic events with stale aggregate versions are suppressed by fanout versioning.

### Flow 17: Centralized Policy AuthZ Guard for Moderation and Permission (ES-P2-004)

Steps:
1. api-gateway validates token and injects normalized identity headers (`X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-SessionState`, `X-Auth-Scope`).
2. api-service delegated jam policy route performs host-only authorization pre-check by loading jam session state and comparing `X-Auth-UserId` to `session.hostUserId`.
3. For host-only policy commands invoked by non-host actors, api-service fails fast with deterministic `403 host_only` and does not proxy command to jam-service.
4. For host actors, api-service forwards policy command to jam-service and command business side effects proceed normally.
5. jam-service host-only guard remains active as defense-in-depth for direct/internal paths that bypass delegated api-service route.

Expected outcome:
1. Non-host actors cannot execute delegated host-only moderation or permission policy actions.
2. Denied host-only commands are rejected upstream in api-service before downstream mutation paths.
3. Host commands continue to reach jam-service and preserve existing moderation behavior.
4. jam-service still rejects non-host commands on direct/internal path access.

Edge cases:
1. Missing or invalid gateway identity headers fail authorization with deterministic unauthorized semantics.
2. Jam state dependency timeout/unavailable during pre-check maps to deterministic dependency-unavailable semantics.
3. Actor not present in participant registry is treated as non-host and denied host-only policy commands.

Validation evidence:
1. `cd backend/api-service && go test ./...`
2. `cd backend/api-service && go test ./internal/bff -run TestProxyHandler_ModerationRoute_NonHostDeniedAtAPIService -count=1`
3. `cd backend/api-service && go test ./internal/bff -run TestProxyHandler_ModerationRoute_HostForwardedByAPIService -count=1`
4. `cd backend/jam-service && go test ./internal/service -run TestModerationNonHostDeniedFastWithHostOnlyAndDeniedAudit -count=1`

Validation evidence:
1. `cd backend/jam-service && go test ./... -count=1`
2. `cd backend/jam-service && go test ./internal/handler -run TestModerationEndpointsAndBlockedQueueCommand -count=1`
3. `cd backend/jam-service && go test ./internal/service -run TestModerationPublishesAuditEvent -count=1`
4. `cd backend/rt-gateway && go test ./... -count=1`
5. `cd backend/rt-gateway && go test ./internal/server -run TestModerationTopicInvokesHookAndFansOut -count=1`
11. `cd frontend && bun run test -- src/components/jam/jam-room-client.test.tsx`

### Flow 18: Permission Projection and Permission-Aware Command Gating (ES-P2-001)

Steps:
1. Host queries and updates guest permission projection via `GET/POST /api/jam/{jamId}/permissions`.
2. Frontend route validates auth claims and forwards through `api-gateway -> api-service (BFF) -> jam-service`.
3. api-service enforces centralized permission-aware prechecks before forwarding protected guest playback and queue reorder paths.
4. jam-service persists permission projection updates in session state and durable store.
5. Accepted permission updates publish `jam.permission.updated` envelopes to Kafka topic `jam.permission.events`.
6. rt-gateway consumes permission topic records and broadcasts realtime updates to room subscribers.

Expected outcome:
1. Guest playback and reorder commands are rejected deterministically when projection flags are disabled.
2. Enabling projection flags allows protected guest commands without changing host ownership semantics.
3. Rejected permission updates do not mutate state and do not publish permission events.
4. Realtime subscribers receive permission updates without requiring page reload.

Edge cases:
1. Empty or invalid permission update payload returns deterministic `400 invalid_input`.
2. Non-host permission update attempts are denied upstream with deterministic `403 host_only` or `403 permission_denied` semantics.
3. Permission event fanout remains isolated to active room subscribers.

Validation evidence:
1. `cd backend/jam-service && go test ./...`
2. `cd backend/api-service && go test ./...`
3. `cd backend/rt-gateway && go test ./internal/server -run TestPermissionTopicFansOutToSubscribers -count=1`
4. `cd frontend && bun run test -- src/components/jam/jam-room-client.test.tsx src/app/api/jam/[jamId]/permissions/route.test.ts`

### Flow 12: Frontend Gateway-Aligned Auth Boundary and Cookie Session Flow

Steps:
1. User submits credential payload through frontend route `POST /api/auth/login`.
2. Frontend route validates request with Zod and forwards to api-gateway `POST /v1/auth/login`.
3. Frontend sets `auth_token` and `refresh_token` as HttpOnly cookies, plus `csrf_token` cookie for state-changing auth calls.
4. Frontend refresh path `POST /api/auth/refresh` requires `X-CSRF-Token` header matching `csrf_token` cookie and forwards refresh token to api-gateway `POST /v1/auth/refresh`.
5. api-gateway forwards auth public lifecycle calls to auth-service; frontend rewrites cookies with rotated token pair.
6. Frontend logout path `POST /api/auth/logout` enforces CSRF match, forwards to api-gateway `POST /v1/auth/logout`, and clears all auth cookies.
7. Frontend `GET /api/auth/me` forwards to api-gateway `GET /v1/auth/me` and returns normalized claims while preserving frontend API boundary semantics.
8. Jam SSR bootstrap route `POST /api/bff/jam/{jamId}/orchestration` forwards to api-gateway `POST /v1/bff/mvp/sessions/{jamId}/orchestration`.

Expected outcome:
1. Valid login returns claims envelope and cookie-backed auth context.
2. Invalid login input returns deterministic `400 invalid_input`.
3. Invalid credentials return deterministic `401 unauthorized` without leaking sensitive details.
4. Refresh with valid CSRF + refresh context rotates session cookies successfully.
5. Refresh/logout with missing or mismatched CSRF token returns deterministic `403 forbidden`.
6. Logout clears cookies even when upstream session context is invalid.

Edge cases:
1. Upstream auth response shape drift is normalized to `502 dependency_invalid_response`.
2. Missing refresh token context returns deterministic `401 unauthorized`.
3. Refresh token replay is enforced upstream as unauthorized/reuse-detected semantics and returned through frontend envelope mapping.
4. Missing auth context on `GET /api/auth/me` returns deterministic `401 unauthorized` without upstream call.

Validation evidence:
1. `cd backend/auth-service && go test ./... -count=1`
2. `cd frontend && bun run test -- src/app/api/auth/login/route.test.ts src/app/api/auth/refresh/route.test.ts src/app/api/auth/logout/route.test.ts src/app/api/auth/me/route.test.ts src/app/api/bff/jam/[jamId]/orchestration/route.test.ts`
3. `cd frontend && bun run lint`
4. `cd frontend && bun run build`

### Flow 13: Frontend shadcn Primitive Conformance Enforcement

Steps:
1. Maintain approved primitive inventory in `frontend/config/shadcn-primitive-inventory.json`.
2. Register only approved deviations in `frontend/config/shadcn-exceptions.json` with complete metadata (`componentPath`, `owner`, `rationale`, `reviewStatus`).
3. Run `bun run check:shadcn` to validate primitive file naming, import conformance, and duplicate primitive detection.
4. Run `bun run lint` to enforce conformance checker + eslint in one deterministic frontend gate.

Expected outcome:
1. Unknown primitives in `frontend/src/components/ui/**` fail validation unless covered by approved exception metadata.
2. Duplicate primitive filenames outside `components/ui` fail validation with actionable file-path output.
3. Imports to unapproved `@/components/ui/*` paths fail validation unless explicitly excepted.
4. Fully conformant code paths pass `check:shadcn` and continue lint pipeline.

Edge cases:
1. Exception entries with missing owner/rationale/review status fail as configuration errors.
2. Temporary exceptions remain valid only while explicitly tracked in exception registry.
3. Test/spec files are excluded from conformance scanning to avoid false positives.

Validation evidence:
1. `cd frontend && bun run test -- src/lib/ui/governance.test.ts`
2. `cd frontend && bun run check:shadcn`
3. `cd frontend && bun run lint`
4. `cd frontend && bun run test`
5. `cd frontend && bun run build`

## Assumptions Logged

1. This baseline excludes proposed but not-yet-implemented flows.
2. Service contracts are based on current tests/specs in repository at initialization time.

## Update Protocol

1. When `/opsx:apply` introduces a new or modified execution flow, add or refine one flow section above.
2. Update expected outcomes and edge cases based on real tests, not planned behavior.
3. If two flows overlap, merge into one generalized pattern and keep old wording out.
