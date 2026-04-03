# Jam Service Requirement File

## Dependency rank
- Rank: 3
- Dependency type: core state owner
- Depends on: auth-service, catalog-service

## Backend APIs used
- POST /api/v1/jams/create
- POST /api/v1/jams/{jamId}/join
- POST /api/v1/jams/{jamId}/leave
- POST /api/v1/jams/{jamId}/end
- GET /api/v1/jams/{jamId}/state
- POST /api/v1/jams/{jamId}/queue/add
- POST /api/v1/jams/{jamId}/queue/remove
- POST /api/v1/jams/{jamId}/queue/reorder
- GET /api/v1/jams/{jamId}/queue/snapshot
- GET /healthz

## Frontend requirements
1. Support complete lifecycle create, join, leave, end.
2. Host leave must transition everyone to ended room state.
3. Queue mutations must be queueVersion-aware.
4. Reorder must recover from version_conflict via snapshot refetch.
5. Add flow must preserve idempotent retry safety.
6. Session ended must disable all write actions immediately.

## Processing flow
1. Lobby create or join sends request through frontend API route.
2. Successful response navigates to room with jamId.
3. Room loads session and queue snapshot.
4. Queue actions call add/remove/reorder routes with required payload.
5. On conflict, frontend fetches latest snapshot and re-renders queue.
6. On host leave or end, room switches read-only and shows ended reason.

## Spotify-like shadcn component mapping
- Dialog: create, join, leave, end confirmations.
- ScrollArea: long queue list behavior similar to playlist stacks.
- DropdownMenu: queue item actions remove or move.
- Skeleton: queue refresh state during conflict recovery.
- Separator and Badge: participant and status grouping in header.

## Frontend router and page mapping

### App pages
- / -> frontend/src/app/page.tsx
- /jam/[jamId] -> frontend/src/app/jam/[jamId]/page.tsx

### App API routes
- POST /api/jam/create -> frontend/src/app/api/jam/create/route.ts -> POST /api/v1/jams/create
- POST /api/jam/[jamId]/join -> frontend/src/app/api/jam/[jamId]/join/route.ts -> POST /api/v1/jams/{jamId}/join
- POST /api/jam/[jamId]/leave -> frontend/src/app/api/jam/[jamId]/leave/route.ts -> POST /api/v1/jams/{jamId}/leave
- POST /api/jam/[jamId]/end -> frontend/src/app/api/jam/[jamId]/end/route.ts -> POST /api/v1/jams/{jamId}/end
- GET /api/jam/[jamId]/state -> frontend/src/app/api/jam/[jamId]/state/route.ts -> GET /api/v1/jams/{jamId}/state
- POST /api/jam/[jamId]/queue/add -> frontend/src/app/api/jam/[jamId]/queue/add/route.ts -> POST /api/v1/jams/{jamId}/queue/add
- POST /api/jam/[jamId]/queue/remove -> frontend/src/app/api/jam/[jamId]/queue/remove/route.ts -> POST /api/v1/jams/{jamId}/queue/remove
- POST /api/jam/[jamId]/queue/reorder -> frontend/src/app/api/jam/[jamId]/queue/reorder/route.ts -> POST /api/v1/jams/{jamId}/queue/reorder
- GET /api/jam/[jamId]/queue/snapshot -> frontend/src/app/api/jam/[jamId]/queue/snapshot/route.ts -> GET /api/v1/jams/{jamId}/queue/snapshot

### Router flow
1. User creates or joins from /.
2. Frontend route returns jamId then navigates to /jam/[jamId].
3. /jam/[jamId] loads state and queue snapshot through API routes above.
4. Queue write routes normalize version_conflict and session_ended.
5. UI executes snapshot refetch path when conflict happens.

## Success criteria
1. Lifecycle transitions in UI match backend session state transitions.
2. Queue mutations are consistent with backend queueVersion.
3. Session ended status prevents additional writes with clear feedback.
