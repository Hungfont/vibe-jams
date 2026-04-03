# Frontend Phase 1 HLD (2 Layers)

This document expands architecture detail into two layers:
- Layer 1: Logical architecture and ownership boundaries.
- Layer 2: Runtime sequence flows for end-to-end behavior.

## Scope
- Phase 1 web frontend only.
- Spotify-like Jam experience using shadcn components.
- Browser calls only through frontend API routes.

## Service requirement files by dependency rank
1. ../services/00-service-dependency-ranking.md
2. ../services/01-auth-service-requirements.md
3. ../services/02-catalog-service-requirements.md
4. ../services/03-jam-service-requirements.md
5. ../services/04-playback-service-requirements.md
6. ../services/05-rt-gateway-requirements.md
7. ../services/06-api-service-bff-requirements.md

## Layer 1: Logical Architecture

### 1.1 Component and service boundaries

```mermaid
flowchart LR
  subgraph UI[Client UI Layer]
    LOB[Lobby Page]
    ROOM[Jam Room Page]
    BAR[Now Playing Bar]
    REC[Realtime Recovery UX]
    ERR[Error and Degraded UX]
  end

  subgraph FE[Frontend Server Layer]
    ROUTES[App Router API Routes]
    MAP[Error and Contract Mapper]
    AUTHCTX[Cookie Session Forwarder]
  end

  subgraph BE[Backend Service Layer]
    BFF[api-service BFF]
    JAM[jam-service]
    PLAY[playback-service]
    AUTH[auth-service]
    CAT[catalog-service]
    RT[rt-gateway websocket]
  end

  LOB --> ROUTES
  ROOM --> ROUTES
  BAR --> ROUTES
  ROUTES --> AUTHCTX
  ROUTES --> MAP

  ROUTES --> BFF
  ROUTES --> JAM
  ROUTES --> PLAY
  ROUTES --> AUTH
  ROUTES --> CAT

  ROOM --> RT
  RT --> REC
  REC --> ROOM
  MAP --> ERR
  ERR --> ROOM
```

### 1.2 State ownership model
1. Authoritative room state comes from jam snapshot and BFF orchestration.
2. Queue and playback transitions are accepted only from backend responses and ordered realtime events.
3. Client local state is a projection, not source of truth.
4. aggregateVersion controls realtime ordering.
5. queueVersion controls command concurrency and stale conflict recovery.

### 1.3 Error contract and UX mapping
- Blocking states:
  - unauthorized
  - premium_required
  - session_ended
- Action-level states:
  - host_only
  - version_conflict
  - track_not_found
  - track_unavailable
- Dependency states:
  - dependency_timeout
  - dependency_unavailable
  - partial=true from BFF

## Layer 2: Runtime Sequence Flows

### 2.1 Create or join -> room hydrate

```mermaid
sequenceDiagram
  participant U as User
  participant P as Lobby Page
  participant F as Frontend API Route
  participant A as auth-service
  participant J as jam-service
  participant B as api-service BFF
  participant R as Jam Room

  U->>P: create or join session
  P->>F: submit action
  F->>A: validate session context
  A-->>F: claims userId plan sessionState
  F->>J: create or join request
  J-->>F: jamId and session state
  F-->>R: navigate to room jamId
  R->>F: orchestration request
  F->>B: POST orchestration
  B-->>F: sessionState queue playback track partial issues
  F-->>R: initial hydrated room state
```

### 2.2 Queue and playback command flow

```mermaid
sequenceDiagram
  participant U as User
  participant R as Jam Room UI
  participant F as Frontend API Route
  participant J as jam-service
  participant P as playback-service
  participant RT as rt-gateway

  U->>R: queue add remove reorder
  R->>F: command with queueVersion
  F->>J: queue mutation
  J-->>F: updated snapshot queueVersion
  F-->>R: render updated queue

  U->>R: host playback command
  R->>F: command with expectedQueueVersion
  F->>P: playback command
  P-->>F: accepted or error
  F-->>R: accepted state or deterministic error
  RT-->>R: ordered playback and queue updates
```

### 2.3 Realtime gap and reconnect recovery

```mermaid
sequenceDiagram
  participant R as Jam Room UI
  participant W as WebSocket Client
  participant G as rt-gateway
  participant J as jam-service

  R->>W: connect with sessionId and lastSeenVersion
  W->>G: GET ws sessionId lastSeenVersion
  G-->>W: incremental events by aggregateVersion

  alt gap detected or stale reconnect
    G->>J: fetch session snapshot
    J-->>G: session and queue aggregateVersion
    G-->>W: jam.session.snapshot recovery=true
    W-->>R: replace local state and resume incremental apply
  end
```

### 2.4 Session ended flow

```mermaid
sequenceDiagram
  participant H as Host
  participant R as Room UI
  participant F as Frontend API Route
  participant J as jam-service
  participant G as rt-gateway

  H->>R: leave or end session
  R->>F: lifecycle action
  F->>J: leave or end
  J-->>F: ended session snapshot
  G-->>R: ended update event
  R-->>R: switch room to read-only and disable write actions
```

## Spotify-like component guidance (shadcn)
1. Shell layout: sidebar + content + persistent player bar.
2. Lobby actions: Card, Tabs, Input, Button, Alert, Toast.
3. Room queue: ScrollArea, Card, DropdownMenu, Skeleton.
4. Playback controls: Button group, Slider, Tooltip.
5. Lifecycle confirmations: Dialog and Sheet for mobile.
6. Realtime feedback: Badge for connection, Toast for recovery.

## Acceptance checklist
1. Create and join route to hydrated room successfully.
2. Queue mutation and playback commands honor queueVersion and role checks.
3. Realtime events are monotonic and duplicates are ignored.
4. Gap and reconnect recovery replaces stale local state.
5. Ended sessions enforce read-only UI and deterministic messaging.
