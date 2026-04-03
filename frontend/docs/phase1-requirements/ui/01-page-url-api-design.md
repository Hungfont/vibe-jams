# Phase 1 UI Design by Page and URL

This document is now split into per-page requirement files.

## Separated UI requirement files
1. pages/00-ui-pages-index.md
2. pages/01-lobby-create-join.md
3. pages/02-jam-room-shell.md
4. pages/03-queue-focus.md
5. pages/04-playback-focus.md
6. pages/05-participants-role.md
7. pages/06-diagnostics-degraded.md

## Notes
1. Each file contains URL, objective, required shadcn components, API caller mapping, and processing flow.
2. Keep browser calls through frontend API routes only.
3. Keep deterministic error mapping consistent with backend response codes.

## Implemented mapping snapshot (opsx-apply)

1. `/`
	- UI: Lobby create and join tabs.
	- Caller routes: `POST /api/auth/validate`, `POST /api/jam/create`, `POST /api/jam/{jamId}/join`.
	- Navigation: success routes to `/jam/{jamId}`.

2. `/jam/{jamId}`
	- UI: shell + queue/playback/participants/diagnostics sections.
	- Server first-load: `POST /api/bff/jam/{jamId}/orchestration`.
	- Caller routes: `GET /api/jam/{jamId}/state`, `POST /api/jam/{jamId}/queue/add`, `POST /api/jam/{jamId}/queue/remove`, `POST /api/jam/{jamId}/queue/reorder`, `GET /api/jam/{jamId}/queue/snapshot`, `POST /api/jam/{jamId}/playback/commands`, `POST /api/jam/{jamId}/end`.
	- Realtime bootstrap: `GET /api/realtime/ws-config`.
