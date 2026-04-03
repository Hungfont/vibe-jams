## Why

Phase 1 already has backend contracts for Jam lifecycle, queue commands, playback commands, realtime fanout, and BFF orchestration, but frontend behavior is not yet standardized by page and URL. We need a clear Spotify-like UI contract now so implementation can proceed with deterministic flows, dependency ordering, and consistent API usage.

## What Changes

- Define Phase 1 frontend UI requirements by concrete page URL, including lobby, jam room shell, queue focus, playback focus, participants, and diagnostics views.
- Define per-page API caller mapping from frontend API routes to backend endpoints, including auth, jam, playback, BFF orchestration, catalog lookup, and realtime ws bootstrap.
- Define deterministic UI handling rules for core error states (`unauthorized`, `premium_required`, `host_only`, `version_conflict`, `session_ended`, `track_not_found`, `track_unavailable`, dependency degradation).
- Define Spotify-like shadcn component usage baseline for each page while preserving existing backend contracts.
- Define dependency-ranked delivery order so foundational auth and contract surfaces are built before queue/playback/realtime UX layers.
- Define OpenSpec execution context binding by custom agent so proposal or apply workflow explicitly loads mapped rules, instructions, skills, and prompts for Frontend Engineer, Backend Engineer, and Agent orchestration contexts.

## Capabilities

### New Capabilities
- `frontend-phase1-ui-routing-and-flows`: Page-by-page URL requirements, Spotify-like UI composition, API caller mapping, and deterministic end-to-end flow behavior for Phase 1 frontend.

### Modified Capabilities
- None.

## Impact

- Affected code: `frontend/src/app/**` pages and frontend API route handlers.
- Affected docs: `frontend/docs/phase1-requirements/**` and related implementation guides.
- Affected workflow docs and setup: `.github/copilot-instructions.md`, `.github/commands/opsx-*.md`, `.github/skills/openspec-*/SKILL.md`, `.github/agents/backend.agent.md`.
- API impact: No backend contract changes; frontend route mapping and usage constraints are formalized.
- Dependencies: `api-service`, `auth-service`, `jam-service`, `playback-service`, `catalog-service`, and `rt-gateway` runtime availability for end-to-end flows.
