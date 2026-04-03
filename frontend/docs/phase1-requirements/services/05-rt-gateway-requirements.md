# RT Gateway Requirement File

## Dependency rank
- Rank: 5
- Dependency type: realtime sync layer
- Depends on: jam-service snapshot contract, queue and playback events

## Backend APIs used
- GET /ws?sessionId={id}&lastSeenVersion={optional}
- GET /metrics/fanout
- GET /healthz

## Frontend requirements
1. WebSocket subscribe must include valid sessionId.
2. Reconnect should include lastSeenVersion when available.
3. Client must apply strict monotonic aggregateVersion ordering.
4. Duplicate and stale events must be ignored.
5. On version gap, frontend must fetch and apply snapshot recovery.
6. jam.session.snapshot with recovery=true must replace local room projection.

## Processing flow
1. Jam room connects websocket after initial hydration.
2. Client applies incremental events if version equals local plus one.
3. Client ignores stale or duplicate versions.
4. When gap is detected, snapshot recovery path is executed.
5. After snapshot apply, client resumes incremental event processing.

## Spotify-like shadcn component mapping
- Badge: connection and sync status.
- Toast: snapshot recovery notification.
- Skeleton: short resync placeholders without full page block.
- Alert: persistent degraded connection warning if retries fail.

## Frontend router and page mapping

### App pages
- /jam/[jamId] -> frontend/src/app/jam/[jamId]/page.tsx

### App API routes
- GET /api/realtime/ws-config -> frontend/src/app/api/realtime/ws-config/route.ts -> returns ws endpoint, sessionId, and lastSeenVersion for client connect bootstrap

### Realtime transport path
- WebSocket connect from browser to rt-gateway: GET /ws?sessionId={id}&lastSeenVersion={optional}

### Router flow
1. /jam/[jamId] asks /api/realtime/ws-config for connection bootstrap.
2. Client opens websocket to rt-gateway using returned params.
3. Event reducer applies monotonic aggregateVersion updates.
4. On gap, UI fetches /api/jam/[jamId]/state and replaces local projection.
5. Reconnect reuses lastSeenVersion cursor and resumes incremental updates.

## Success criteria
1. Reconnect and gap recovery maintain consistent room state.
2. Out-of-order events do not corrupt queue or playback UI.
3. Recovery state is visible but non-disruptive.
