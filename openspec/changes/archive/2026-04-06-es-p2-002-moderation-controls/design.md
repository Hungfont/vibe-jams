## Context

Phase 1 provides lifecycle, queue, playback, and realtime baseline but does not include host moderation controls. Participant state exists in jam-service repository snapshots, and rt-gateway already fans out queue and playback event streams. ES-P2-002 introduces moderation as a new command family that must keep state, event fanout, and UI behavior consistent.

## Goals / Non-Goals

**Goals:**
- Add host moderation commands (`mute`, `kick`) in jam-service.
- Ensure muted or kicked users are blocked from moderated actions.
- Publish moderation audit events to `jam.moderation.events` with deterministic payload.
- Fanout moderation updates to room clients via rt-gateway.
- Provide frontend controls and participant-state visibility in Jam Room.

**Non-Goals:**
- Do not redesign full permission matrix from ES-P2-001.
- Do not replace centralized authZ architecture from ES-P2-004.
- Do not introduce new external abuse-detection service in this change.

## Decisions

1. Participant moderation state is stored in jam-service session projection.
- Add `muted` participant flag and kick transition (participant removal).
- Keep host moderation authority checked at command handling boundary.

2. Moderation audit payload is standardized in moderation events.
- Event payload includes `action`, `actorUserId`, `targetUserId`, `reason`, `occurredAt`.
- Publish to dedicated topic `jam.moderation.events` keyed by `sessionId`.

3. Blocked-action enforcement is applied in jam command path.
- Queue command path rejects actor when participant is muted or not in session due to kick.
- Return deterministic moderation-blocked error semantics.

4. Realtime fanout consumes moderation topic alongside queue/playback.
- rt-gateway topic filter includes moderation topic.
- Existing fanout processor broadcasts moderation envelope using shared ordering path.

5. Abuse heuristics extension point is introduced as a consumer hook.
- Add optional hook interface invoked for moderation events in rt-gateway consume loop.
- Default implementation is no-op to avoid runtime dependency changes.

## Risks / Trade-offs

- [Risk: auth source split between frontend route validation and backend command payload actor] -> Mitigation: backend validates host authority for moderation commands and enforces participant-based blocking in repository path.
- [Risk: topic/config drift across shared topic constants, bootstrap script, and gateway config] -> Mitigation: update shared topic baseline and tests in the same change.
- [Risk: moderation updates visible in realtime but stale in snapshot fallback paths] -> Mitigation: ensure session snapshot includes moderation state and is used by reconnect/gap recovery.
- [Risk: blocked-action semantics may overlap with future ES-P2-001 permissions] -> Mitigation: use moderation-specific error code and keep permission controls out of scope.

## Migration Plan

1. Add moderation model fields and repository transitions in jam-service.
2. Add moderation command handlers and service methods with host checks.
3. Add moderation topic constants, producer support, and topic validation fixtures.
4. Extend rt-gateway config/topic subscription and moderation hook invocation.
5. Add frontend moderation routes and Jam Room participant controls.
6. Update sequence/runbook docs and execute targeted backend/frontend tests.

Rollback:
- Disable moderation route usage in frontend and revert jam-service moderation handler wiring while keeping existing queue/playback flows unchanged.

## Open Questions

- Should kicked users be blocked from rejoin attempts in ES-P2-002, or deferred to ES-P2-001/ES-P2-004 policy layer?
- Should moderation blocked-action response use dedicated code (`moderation_blocked`) or reuse existing policy code taxonomy?