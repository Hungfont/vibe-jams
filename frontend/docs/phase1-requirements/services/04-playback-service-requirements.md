# Playback Service Requirement File

## Dependency rank
- Rank: 4
- Dependency type: core command path
- Depends on: auth-service, catalog-service, jam queue version model

## Backend APIs used
- POST /v1/jam/sessions/{sessionId}/playback/commands
- GET /healthz

## Frontend requirements
1. Playback controls are host-only.
2. Commands supported: play, pause, next, prev, seek.
3. Payload must always include command, clientEventId, expectedQueueVersion.
4. host_only and unauthorized must disable or block command interaction.
5. version_conflict must trigger snapshot recovery and retry option.
6. session_ended must block command dispatch.

## Processing flow
1. Host presses playback control in persistent player bar.
2. Frontend API route sends command payload to backend.
3. Accepted response updates command status immediately.
4. Realtime fanout confirms authoritative playback transition.
5. Rejected response maps to deterministic UI reason and fallback action.

## Spotify-like shadcn component mapping
- Button group: playback transport controls.
- Slider: seek and timeline interaction.
- Tooltip: disabled reason for non-host users.
- Toast: accepted and rejected command feedback.
- Alert: blocking errors such as session ended.

## Frontend router and page mapping

### App pages
- /jam/[jamId] -> frontend/src/app/jam/[jamId]/page.tsx

### App API routes
- POST /api/jam/[jamId]/playback/commands -> frontend/src/app/api/jam/[jamId]/playback/commands/route.ts -> POST /v1/jam/sessions/{sessionId}/playback/commands

### Router flow
1. Host interacts with player controls in /jam/[jamId].
2. UI submits command to /api/jam/[jamId]/playback/commands.
3. Route forwards expectedQueueVersion and normalizes response code.
4. UI applies accepted state and waits for realtime confirmation.
5. UI displays deterministic fallback for rejected commands.

## Success criteria
1. Host commands are accepted and synchronized to room state.
2. Non-host users never execute host-only commands.
3. Conflict and ended states produce deterministic fallback UX.
