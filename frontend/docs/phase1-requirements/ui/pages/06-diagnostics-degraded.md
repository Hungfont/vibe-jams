# UI Requirement: Diagnostics and Degraded States

## URL
- /jam/[jamId]?view=diagnostics

## Objective
1. Keep UI stable during backend dependency failures.
2. Provide deterministic, actionable error handling.

## Required UI components (shadcn)
1. Alert
2. Toast
3. Dialog
4. Skeleton

## API callers
1. POST /api/bff/jam/[jamId]/orchestration -> POST /v1/bff/mvp/sessions/{sessionId}/orchestration
2. GET /api/jam/[jamId]/state -> GET /api/v1/jams/{jamId}/state
3. POST /api/auth/validate -> POST /internal/v1/auth/validate

## Requirements
1. unauthorized must trigger re-auth flow.
2. premium_required must trigger upgrade prompt.
3. host_only must disable control with reason.
4. version_conflict must support refresh and retry.
5. dependency_timeout and dependency_unavailable must show degraded panel without blocking core room.

## Processing flow
1. Diagnostics view reads dependencyStatuses and issues.
2. UI shows blocking and non-blocking states separately.
3. Recovery action calls state or orchestration refetch.
4. Core room remains interactive when required data is still valid.
