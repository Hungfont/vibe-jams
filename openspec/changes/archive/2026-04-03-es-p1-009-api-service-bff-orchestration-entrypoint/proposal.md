## Why

MVP web clients currently need to coordinate multiple backend service calls directly for auth, jam, playback, and catalog flows. This increases client complexity, weakens consistency of error handling, and slows integration delivery.

## What Changes

- Add an API-service BFF orchestration entrypoint for MVP web flows.
- Centralize downstream routing to auth, jam, playback, and catalog service dependencies behind a single BFF-facing API surface.
- Standardize response envelope, timeout boundaries, and dependency failure mapping for orchestrated web flows.
- Define minimal aggregation behavior for combined web-view payloads used by MVP client screens.

## Capabilities

### New Capabilities
- `api-service-bff-orchestration`: BFF entrypoint contracts, orchestration behavior, routing, aggregation, and deterministic failure semantics for MVP web flows.

### Modified Capabilities
- `auth-claim-contract`: Clarify claim propagation requirements from BFF entrypoint to downstream protected routes.
- `catalog-track-validation-api`: Clarify BFF-side consumption contract for catalog lookups in MVP web aggregation flows.

## Impact

- Affected code: `backend/api-service` (routing, orchestration handlers, downstream clients, integration tests).
- Affected APIs: new or expanded BFF endpoint surface consumed by MVP web clients.
- Affected dependencies: runtime connectivity to `auth-service`, `jam-service`, `playback-service`, `catalog-service`.
- Operational impact: adds centralized timeout/error policy and observability points in API-service BFF layer.
