## Context

Phase 1 backend contracts are already available for Jam lifecycle, queue mutations, playback commands, BFF orchestration, and realtime fanout. Current frontend planning exists in docs, but implementation lacks a formal OpenSpec contract that maps each UI page URL to deterministic API callers, state transitions, and error behavior.

This change targets a Spotify-like web experience for Jam workflows without modifying backend APIs. The frontend must use App Router pages and frontend API routes as the only browser boundary.

## Goals / Non-Goals

**Goals:**
- Define page-by-page frontend behavior for Phase 1 URLs (`/` and `/jam/[jamId]` variants).
- Define deterministic API caller mapping from frontend routes to backend endpoints.
- Define consistent UX handling for lifecycle, queue, playback, realtime recovery, and degraded dependencies.
- Define shadcn component baseline that supports Spotify-like interaction patterns.

**Non-Goals:**
- No backend API contract changes.
- No mobile app design.
- No changes to Kafka topic contracts or realtime server internals.
- No redesign of auth model beyond frontend usage rules.

## Decisions

### Decision 1: Browser-to-backend access is prohibited; browser-to-frontend-route is required
- Rationale: preserves auth/session forwarding control, unified error normalization, and endpoint portability.
- Alternative considered: direct browser calls to backend services.
- Why not chosen: duplicates auth handling across components and weakens deterministic envelope mapping.

### Decision 2: Jam room initial render uses orchestration first, then realtime incremental updates
- Rationale: one aggregated payload reduces initial data divergence and supports partial dependency semantics.
- Alternative considered: loading each service separately from client.
- Why not chosen: higher complexity and increased inconsistency during first paint.

### Decision 3: Realtime reducer is version-driven and snapshot-recovery aware
- Rationale: aggregateVersion and queueVersion are already system-level ordering primitives.
- Alternative considered: event-time ordering with best-effort merges.
- Why not chosen: non-deterministic under network jitter and reconnect gaps.

### Decision 4: Spotify-like layout is implemented using shadcn composition, not custom primitives
- Rationale: consistent accessibility, predictable maintenance, and adherence to frontend rules.
- Alternative considered: custom UI primitives.
- Why not chosen: duplicates functionality and increases visual/behavior drift.

### Decision 5: Page-based requirement separation
- Rationale: each URL has distinct goals, API callers, and failure modes, enabling incremental implementation and testability.
- Alternative considered: monolithic UI spec.
- Why not chosen: harder to sequence and validate dependency-based rollout.

### Decision 6: OpenSpec workflow execution is bound to custom agent context
- Rationale: enforcing explicit context selection prevents missing frontend/backend rules and ensures relevant skills/instructions/prompts are loaded deterministically before artifact generation or implementation.
- Alternative considered: rely on implicit default context selection.
- Why not chosen: inconsistent context loading can lead to path mismatch, missed rules, and non-deterministic implementation behavior.

## Risks / Trade-offs

- [Risk] Orchestration partial responses may be misunderstood as full success in UI -> Mitigation: explicit degraded panel, dependency status badges, and action-level disable rules.
- [Risk] Realtime event gaps can desync queue/playback state -> Mitigation: mandatory snapshot fallback path with replace-not-merge semantics.
- [Risk] Role and entitlement checks duplicated in multiple components -> Mitigation: central policy helpers and route-level error normalization.
- [Risk] Spotify-like visual parity can inflate scope -> Mitigation: prioritize interaction parity and information hierarchy over exact visual cloning.
- [Risk] Workflow drift where commands and skills reference different context paths -> Mitigation: keep command layer, skill layer, and agent instruction scope synchronized in `.github` setup files.

## Migration Plan

1. Add frontend API routes in dependency order: auth, jam lifecycle, queue, playback, orchestration, realtime config.
2. Implement pages in order: lobby, jam room shell, queue focus, playback focus, participants, diagnostics.
3. Add reducer logic for version ordering and recovery.
4. Validate end-to-end flows: create/join, queue mutations with conflict recovery, host playback, reconnect recovery, degraded dependency handling.
5. Bind OpenSpec command and skill workflows to custom agent resource maps (rules, instructions, skills, prompts).
6. Roll out behind frontend feature flags if needed.

Rollback strategy:
- Revert to lobby-only behavior and disable jam room advanced actions by feature flags while retaining existing baseline routes.

## Open Questions

- Should queue and playback focus views be query-driven tabs only, or independent nested routes in App Router?
- Should diagnostics be always visible as a side panel or only visible under a dedicated query view?
- Should realtime ws bootstrap include heartbeat thresholds from frontend route config for adaptive reconnect?
