# User Stories - Spotify-Like Jam

## Phase 1 (MVP)

### US-P1-001
- Persona: Premium listener (host)
- Story: As a Premium listener, I want to start a Jam from a playlist so that my group can listen together in real time.
- Acceptance Criteria:
  - Host can create a Jam session with `POST /v1/jam/sessions`.
  - Session response contains `sessionId` and `inviteToken`.
  - Non-premium host attempt returns `403 premium_required`.

### US-P1-002
- Persona: Participant (guest)
- Story: As a participant, I want to join a Jam with an invite token so that I can listen in the same session.
- Acceptance Criteria:
  - Valid token allows join and returns session snapshot.
  - Invalid/expired token returns `404 session_not_found` or equivalent join error.
  - Participant is visible in participant list via realtime updates.

### US-P1-003
- Persona: Guest
- Story: As a guest, I want to add a song to the Jam queue so that everyone can hear my selected track.
- Acceptance Criteria:
  - Guest can add item using `POST /v1/jam/sessions/{sessionId}/queue/items`.
  - Successful add returns `queueItemId` and incremented `queueVersion`.
  - Duplicate client submit with same idempotency key does not create duplicate items.

### US-P1-004
- Persona: Host
- Story: As a host, I want to control playback commands so that I can manage session listening flow.
- Acceptance Criteria:
  - Playback command endpoint accepts valid host actions (`play`, `pause`, `next`, `prev`, `seek`).
  - Unauthorized user receives `403 host_only` when host-only control is enabled.
  - Command acceptance returns `202` with `accepted=true`.

### US-P1-005
- Persona: Any session participant
- Story: As a participant, I want queue and playback updates in realtime so that my UI stays in sync.
- Acceptance Criteria:
  - Client subscribed to `jam:{sessionId}` receives `jam.queue.updated` and `jam.playback.updated`.
  - Events include `eventVersion`.
  - On event gaps, client can recover by fetching session snapshot.

### US-P1-006
- Persona: Host
- Story: As a host, I want the Jam to end when I leave so that the session has clear ownership.
- Acceptance Criteria:
  - Host leave action marks session ended.
  - All active clients receive `jam.ended`.
  - Further queue/playback commands after end are rejected.

### US-P1-007
- Persona: Backend engineer (`jam-service`)
- Story: As a backend engineer, I want auth token and premium entitlement validation to be available so that Jam create/end flows are not blocked by identity checks.
- Acceptance Criteria:
  - Auth validation endpoint/middleware returns stable user identity claims.
  - Premium entitlement is available for host create/end authorization.
  - Invalid or expired token is rejected consistently with `401 unauthorized`.

### US-P1-008
- Persona: Backend engineer (`jam-service`/`playback-service`)
- Story: As a backend engineer, I want track metadata lookup and existence validation from catalog so that queue and playback commands can proceed safely.
- Acceptance Criteria:
  - Service can resolve `trackId` to playable metadata.
  - Missing or unavailable track returns deterministic client error.
  - Queue add/playback pre-check uses catalog response before state mutation.

## Phase 2 (Jam Control and Consistency Upgrades)

### US-P2-001
- Persona: Host
- Story: As a host, I want granular guest permissions so that I can decide who can control playback or queue order.
- Acceptance Criteria:
  - Host can update permissions per action category.
  - Permission changes are reflected in realtime to all participants.
  - Non-host permission changes are rejected with `403 host_only`.

### US-P2-002
- Persona: Host
- Story: As a host, I want to mute or remove disruptive users so that the session remains usable.
- Acceptance Criteria:
  - Host can trigger moderation actions (mute/kick).
  - Moderated user loses corresponding capabilities immediately.
  - Moderation event is traceable via audit stream.

### US-P2-003
- Persona: Guest
- Story: As a guest, I want clear conflict feedback when queue state changed so that I can retry with latest data.
- Acceptance Criteria:
  - Stale reorder requests return `409 version_conflict`.
  - Error response includes enough metadata to request latest snapshot.
  - Retry with latest version can succeed.

### US-P2-004
- Persona: Participant
- Story: As a participant, I want playback synchronization to remain stable across frequent commands so that experience feels consistent.
- Acceptance Criteria:
  - Playback updates include `queueVersion` and `playbackEpoch`.
  - Client applies updates only when versioning rules are satisfied.
  - Drift scenarios are recoverable through snapshot flow.

### US-P2-005
- Persona: Backend engineer (`jam-service`)
- Story: As a backend engineer, I want host/guest authorization guards for permission and moderation actions so that policy enforcement does not block Phase 2 controls.
- Acceptance Criteria:
  - Host-only actions are enforced through centralized authZ guard.
  - Unauthorized actor receives `403 host_only` consistently.
  - Permission and moderation actions include actor identity in audit context.

### US-P2-006
- Persona: Backend engineer (`catalog-service`)
- Story: As a backend engineer, I want catalog access policy checks for queue operations so that restricted or unavailable tracks are filtered early.
- Acceptance Criteria:
  - Queue add/reorder requests validate catalog availability rules.
  - Policy violations return safe and consistent errors.
  - Policy checks do not break existing queue version and idempotency rules.

## Phase 3 (Personalization and Recommendation)

### US-P3-001
- Persona: Participant
- Story: As a participant, I want group-aware recommendations so that suggested songs fit the whole group taste.
- Acceptance Criteria:
  - Recommendation list updates during active Jam.
  - Suggestions reflect group interactions (adds/skips/likes).
  - Recommendation payload is tied to current `sessionId`.

### US-P3-002
- Persona: Host
- Story: As a host, I want recommended songs to be easy to insert into queue so that session momentum stays high.
- Acceptance Criteria:
  - Recommended item can be added to queue with one action.
  - Added recommendation follows same queue versioning/idempotency rules.
  - Queue update is broadcast in realtime after add.

### US-P3-003
- Persona: Product owner
- Story: As a product owner, I want recommendation behavior telemetry so that I can improve ranking quality.
- Acceptance Criteria:
  - Recommendation impression/click/add signals are emitted as analytics events.
  - Events are attributable to session and user context.
  - Metrics enable per-phase quality reporting.

### US-P3-004
- Persona: Backend engineer (`api-service`)
- Story: As a backend engineer, I want recommendation endpoint auth scope validation so that only eligible users can access session-scoped recommendations.
- Acceptance Criteria:
  - Recommendation endpoint validates user/session access context.
  - Unauthorized access to recommendation payload is rejected.
  - Access validation does not increase response latency beyond agreed threshold.

### US-P3-005
- Persona: Backend engineer (`api-service`/`catalog-service`)
- Story: As a backend engineer, I want recommendation outputs enriched with catalog metadata so that clients can render and add suggestions without extra blocking calls.
- Acceptance Criteria:
  - Recommendation responses include required track metadata fields.
  - Missing catalog metadata is handled with fallback or filtered result.
  - Add-to-queue from recommendation remains one-step for client flow.

## Phase 4 (Hardening and Scale)

### US-P4-001
- Persona: Participant
- Story: As a participant, I want Jam to remain responsive during high traffic so that listening experience is reliable.
- Acceptance Criteria:
  - Queue and playback event latency remains within defined SLO at target load.
  - Backpressure handling prevents system-wide event drops.
  - Service degradation behavior is graceful and observable.

### US-P4-002
- Persona: Operations engineer
- Story: As an operations engineer, I want resilient failover behavior so that sessions survive infrastructure incidents.
- Acceptance Criteria:
  - Multi-AZ failover procedure is documented and tested.
  - Recovery objectives are measurable and rehearsed.
  - Critical services can be restored without data corruption.

### US-P4-003
- Persona: Security/compliance owner
- Story: As a compliance owner, I want stronger abuse controls and policy tuning so that platform misuse is reduced.
- Acceptance Criteria:
  - Abuse detection signals are captured and actionable.
  - Rate-limit policies can be tuned per endpoint and session type.
  - Repeated offenders can be blocked through policy workflows.

### US-P4-004
- Persona: Security engineer (`auth-service`)
- Story: As a security engineer, I want identity-aware rate limit and abuse block hooks so that abusive clients can be throttled or blocked without impacting valid sessions.
- Acceptance Criteria:
  - Rate limit can be keyed by user/session identity claims.
  - Abuse decision hooks can trigger temporary block state.
  - Critical Jam flows remain available for non-abusive traffic.

### US-P4-005
- Persona: Reliability engineer (`catalog-service`)
- Story: As a reliability engineer, I want catalog read resilience and fallback behavior during incidents so that playback and recommendations stay available.
- Acceptance Criteria:
  - Catalog failures trigger bounded fallback/degraded response mode.
  - Recovery path preserves consistency once dependency is healthy.
  - Incident runbook includes catalog fallback validation steps.
