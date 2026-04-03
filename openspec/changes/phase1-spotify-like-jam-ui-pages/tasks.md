## 1. Frontend Route Foundation

- [x] 1.1 Define shared frontend API envelope and normalized error mapper for all route handlers
- [x] 1.2 Implement auth validation frontend route (`/api/auth/validate`) with cookie/session forwarding
- [x] 1.3 Implement catalog lookup frontend route (`/api/catalog/tracks/[trackId]`) with deterministic track error mapping
- [x] 1.4 Add shared frontend route utilities for backend base URLs, timeout handling, and response parsing

## 2. Jam Lifecycle and Room Data Routes

- [x] 2.1 Implement jam lifecycle frontend routes (`create`, `join`, `leave`, `end`)
- [x] 2.2 Implement jam state frontend route (`/api/jam/[jamId]/state`)
- [x] 2.3 Implement queue frontend routes (`add`, `remove`, `reorder`, `snapshot`) with version conflict normalization
- [x] 2.4 Implement playback command frontend route (`/api/jam/[jamId]/playback/commands`) with host-only and session-ended mapping
- [x] 2.5 Implement BFF orchestration frontend route (`/api/bff/jam/[jamId]/orchestration`) with partial and issue propagation
- [x] 2.6 Implement realtime ws bootstrap route (`/api/realtime/ws-config`) for client connect params

## 3. Lobby and Room Shell Pages

- [x] 3.1 Build Lobby page (`/`) with create and join actions using shadcn Card, Tabs, Input, Button, Alert, and Toast
- [x] 3.2 Implement lobby success navigation flow to `/jam/[jamId]`
- [x] 3.3 Build Jam Room shell page (`/jam/[jamId]`) with orchestration first-load and base section layout
- [x] 3.4 Add room-level degraded dependency panel and session ended read-only mode

## 4. Queue and Playback Focused UI

- [x] 4.1 Build queue-focused room section with shadcn ScrollArea, Card, DropdownMenu, Input, Button, Alert, and Toast
- [x] 4.2 Implement queue mutation UX with queueVersion-aware conflict recovery (snapshot then retry)
- [x] 4.3 Build playback-focused room section with shadcn Button controls, Slider, Tooltip, Toast, and Alert
- [x] 4.4 Implement host-only playback command UX with accepted pending state and deterministic rejection handling

## 5. Participants, Realtime Sync, and Diagnostics

- [x] 5.1 Build participants and role section using shadcn Avatar, Badge, Tooltip, and Separator
- [x] 5.2 Implement websocket client integration with monotonic aggregateVersion reducer
- [x] 5.3 Implement gap detection and snapshot replacement recovery flow
- [x] 5.4 Build diagnostics section for unauthorized, premium_required, host_only, version_conflict, and dependency degradation paths

## 6. Validation and Readiness

- [x] 6.1 Add page-level integration tests for lobby create or join and jam room first render
- [x] 6.2 Add interaction tests for queue conflict recovery and playback host-only behavior
- [x] 6.3 Add realtime tests for stale or duplicate event ignore and snapshot recovery apply
- [x] 6.4 Validate that all page URLs use frontend API routes only and never call backend endpoints directly from browser code
- [x] 6.5 Update frontend docs to reflect final page URL behavior and API caller mapping

## 7. OpenSpec Custom-Agent Context Setup

- [x] 7.1 Update OpenSpec command workflows to resolve custom agent context and mapped resources before execution
- [x] 7.2 Update OpenSpec skill workflows to bind Frontend/Backend/Agent resource loading rules
- [x] 7.3 Align backend instruction path references with repository instruction layout
