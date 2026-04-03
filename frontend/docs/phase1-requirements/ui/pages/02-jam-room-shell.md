# UI Requirement: Jam Room Shell

## URL
- /jam/[jamId]

## Objective
1. Render full room context in one entry.
2. Keep core UI working under partial dependency failures.

## Required UI components (shadcn)
1. Card
2. Tabs
3. Badge
4. Separator
5. Skeleton
6. Alert

## API callers
1. POST /api/bff/jam/[jamId]/orchestration -> POST /v1/bff/mvp/sessions/{sessionId}/orchestration
2. GET /api/jam/[jamId]/state -> GET /api/v1/jams/{jamId}/state
3. GET /api/realtime/ws-config -> frontend ws bootstrap
4. WebSocket /ws?sessionId={id}&lastSeenVersion={optional} -> rt-gateway

## Requirements
1. Initial load should use one orchestration call.
2. UI must show session status, host, participants, queueVersion.
3. partial=true must not break core room rendering.
4. session_ended must switch room to read-only mode.

## Processing flow
1. Page loads orchestration payload.
2. Core sections render immediately.
3. Client opens websocket and starts incremental sync.
4. Degraded issues render in diagnostics area without full block.
