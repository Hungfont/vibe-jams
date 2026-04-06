## Context

BFF orchestration currently mixes read aggregation with optional playback execution, while frontend uses this endpoint for SSR preload and periodic refresh. This mismatch introduces side effects in a read-heavy flow.

## Goals / Non-Goals

**Goals:**
- Make orchestration side-effect free.
- Reject playbackCommand at orchestration boundary with deterministic 400 invalid_input.
- Keep playback mutation available only via dedicated playback endpoint.
- Keep contract/schema/docs aligned.

**Non-Goals:**
- No change to playback command endpoint semantics.
- No new endpoints or dependency additions.

## Decisions

- Validate playbackCommand in orchestration service and fail fast with invalid_input.
- Keep playbackCommand in request DTO only to return explicit deterministic rejection.
- Remove playback field from orchestration response contract and OpenAPI schema.
- Keep optional/degradable behavior for catalog only.

## Risks / Trade-offs

- [Clients still sending playbackCommand] -> 400 responses may surface latent client issues. Mitigation: sequence and runbook updates, plus integration tests.
- [Schema drift between layers] -> mitigation: update backend DTOs, OpenAPI, frontend types, and docs in one change.
