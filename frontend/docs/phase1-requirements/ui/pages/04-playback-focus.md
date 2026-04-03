# UI Requirement: Playback Focus

## URL
- /jam/[jamId]?view=playback

## Objective
1. Provide host-only playback controls with clear role feedback.

## Required UI components (shadcn)
1. Button
2. Slider
3. Tooltip
4. Toast
5. Alert

## API callers
1. POST /api/jam/[jamId]/playback/commands -> POST /v1/jam/sessions/{sessionId}/playback/commands
2. WebSocket /ws?sessionId={id}&lastSeenVersion={optional} -> rt-gateway

## Requirements
1. Only host can execute control commands.
2. Payload must include command, clientEventId, expectedQueueVersion.
3. host_only, unauthorized, version_conflict, session_ended must map clearly.
4. Accepted state should appear quickly, then realtime confirms authoritative state.

## Processing flow
1. Host clicks control.
2. UI submits command via frontend API route.
3. Route returns accepted or deterministic error.
4. UI shows pending or error state.
5. Realtime event confirms final playback transition.
