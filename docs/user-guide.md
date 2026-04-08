# Video Streaming Jam System User Guide

## 1. What this system is

This repository implements a Spotify-like collaborative Jam system with:

- User authentication and session lifecycle (login, refresh, logout)
- Jam room creation and join flows
- Queue management (add, remove, reorder)
- Playback commands (play, pause, next, prev, seek)
- Host moderation and guest permission controls
- Realtime updates through websocket fanout

The browser talks only to frontend API routes under `/api/**`. The backend entrypoint is `api-gateway`, which then routes to `api-service` (BFF) and downstream services.

## 2. Service map and local ports

Default local service ports in current code:

- `jam-service`: `8080`
- `auth-service`: `8081`
- `playback-service`: `8082`
- `catalog-service`: `8083`
- `api-service` (BFF): `8084`
- `api-gateway` (public ingress): `8085`
- `rt-gateway`:
  - `api-service` and frontend expect `8086` by default
  - rt-gateway process default is `8090`

Important: for realtime to work, `api-service` `RT_GATEWAY_URL` must match the actual rt-gateway port.

## 3. Prerequisites

- Go `1.21` (from `go.work`)
- Bun `>=1.2.0` (from frontend `package.json`)
- Kafka available at `localhost:9092` if using Kafka transport
- Windows Terminal if using `backend/scripts/start-all-services.ps1`

## 4. Start the system

### Option A: Fast startup script (Windows)

From repo root:

```powershell
./backend/scripts/start-all-services.ps1
```

Notes:

- Script opens one tab per backend service.
- If a default port is occupied, it picks a random port.
- Random port reassignment can break service-to-service URLs unless you update dependent env vars.
- Realtime can degrade if rt-gateway is on `8090` while api-service still points to `8086`.

### Option B: Deterministic manual startup (recommended)

Use fixed ports to match frontend/backend defaults. Example in separate terminals:

```powershell
# jam-service
cd backend/jam-service
$env:SERVER_PORT="8080"
go run ./cmd/server
```

```powershell
# auth-service
cd backend/auth-service
$env:SERVER_ADDR=":8081"
go run ./cmd/server
```

```powershell
# playback-service
cd backend/playback-service
$env:SERVER_PORT="8082"
go run ./cmd/server
```

```powershell
# catalog-service
cd backend/catalog-service
$env:SERVER_PORT="8083"
go run ./cmd/server
```

```powershell
# rt-gateway (align with api-service default expectation)
cd backend/rt-gateway
$env:SERVER_PORT="8086"
go run ./cmd/server
```

```powershell
# api-service
cd backend/api-service
$env:SERVER_PORT="8084"
$env:RT_GATEWAY_URL="http://localhost:8086"
go run ./cmd/server
```

```powershell
# api-gateway
cd backend/api-gateway
$env:SERVER_PORT="8085"
go run ./cmd/server
```

### Start frontend

```powershell
cd frontend
bun install
bun dev
```

Open: `http://localhost:3000`

## 5. Demo accounts

From the in-memory credential store:

- Premium user:
  - identity: `premium@example.com`
  - password: `premium-pass`
- Free user:
  - identity: `free@example.com`
  - password: `free-pass`

## 6. End-user flow

### Step 1: Login

- Go to `/login`
- Enter email and password
- On success, frontend sets cookies:
  - `auth_token` (HttpOnly)
  - `refresh_token` (HttpOnly)
  - `csrf_token` (readable by frontend for CSRF header)

### Step 2: Create or join Jam (Lobby `/`)

- Create Jam:
  - Requires premium plan (`premium`, `premium_plus`, `pro`)
  - Non-premium gets `premium_required`
- Join Jam:
  - Enter existing `jamId`

### Step 3: Jam room (`/jam/{jamId}`)

Main tabs and behavior:

- Queue
  - Add track by track ID
  - Remove queue items
  - Reverse queue order (requires host or permission)
- Playback
  - Commands: play, pause, next, prev, seek
  - Requires host or granted playback permission
- Participants
  - Host can mute/kick participants
  - Host can toggle guest permissions:
    - playback
    - reorder
    - volume
- Diagnostics
  - Shows connection state, aggregate version, dependency issues

### Step 4: Realtime updates

- Client requests websocket bootstrap via `/api/realtime/ws-config`
- Then opens websocket via gateway/BFF proxy path
- If realtime gaps are detected, snapshot recovery is triggered automatically

## 7. Catalog test tracks (local fixtures)

Current seeded track IDs in catalog-service:

- `trk_1`: playable, allowed
- `trk_2`: unavailable (`license_blocked`)
- `trk_3`: policy restricted (`region_blocked`)

When catalog validation is enabled in jam/playback services, these IDs are useful for validating acceptance and rejection paths.

## 8. Common errors and meaning

- `unauthorized`
  - Missing/invalid auth context, expired session, or invalid token
- `forbidden`
  - Often CSRF failure on refresh/logout (`X-CSRF-Token` mismatch)
- `premium_required`
  - Non-premium user tried premium-only operation (for example create jam)
- `session_ended`
  - Write command against ended session
- `version_conflict`
  - Optimistic concurrency conflict on queue/playback writes
- `track_not_found` / `track_unavailable` / `track_restricted`
  - Catalog validation rejected the track
- `dependency_unavailable`
  - Downstream service temporarily unreachable

## 9. Health checks

Quick liveness endpoints:

- `http://localhost:8080/healthz` (jam)
- `http://localhost:8081/healthz` (auth)
- `http://localhost:8082/healthz` (playback)
- `http://localhost:8083/healthz` (catalog)
- `http://localhost:8084/healthz` (api-service)
- `http://localhost:8085/healthz` (api-gateway)
- `http://localhost:8086/healthz` or `http://localhost:8090/healthz` (rt-gateway, depending on config)

## 10. Troubleshooting

### Realtime status stays degraded in Jam room

Likely cause: rt-gateway port mismatch.

- Check rt-gateway actual port
- Ensure api-service `RT_GATEWAY_URL` points to that port
- Ensure websocket route `/v1/bff/mvp/realtime/ws-config` is reachable through api-gateway

### Login works but refresh/logout fails with 403

Likely cause: missing or mismatched CSRF header.

- Refresh/logout require `X-CSRF-Token` matching `csrf_token` cookie
- Use frontend flows instead of calling backend endpoints directly from browser scripts

### Create Jam fails for free account

Expected behavior.

- Jam creation is premium-only
- Use premium fixture account for host operations

### Queue/playback behaves inconsistently under concurrent actions

Expected conflict handling path.

- `version_conflict` means client snapshot is stale
- Client should refresh snapshot and retry with latest queue version
