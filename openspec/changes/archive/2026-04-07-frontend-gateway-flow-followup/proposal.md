## Why

Change `api-gateway-authn-authz` is complete and established api-gateway as the entry point for BFF/auth flows. Frontend server routes still contain direct upstream targeting patterns that are inconsistent with the gateway-first contract, so we need a frontend follow-up to align runtime behavior and documentation.

## What Changes

- Route frontend auth lifecycle API handlers (`/api/auth/login`, `/api/auth/refresh`, `/api/auth/logout`, `/api/auth/me`) to `api-gateway` as upstream target.
- Route frontend BFF orchestration API handler (`/api/bff/jam/{jamId}/orchestration`) to `api-gateway` upstream target.
- Introduce explicit frontend gateway base URL config (`API_GATEWAY_URL`) and align defaults to current local runtime topology.
- Keep browser-side API contract unchanged (`/api/**` routes stay stable).
- Update frontend flow docs and runbook test flow to reflect gateway-first behavior.
- Add/adjust tests for route-to-upstream mapping and regression safety.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `frontend-phase1-ui-routing-and-flows`: frontend auth and BFF orchestration upstream behavior is updated to use api-gateway as primary ingress while preserving existing browser route contracts.

## Impact

- Frontend server route handlers under `frontend/src/app/api/auth/**` and `frontend/src/app/api/bff/**`.
- Frontend backend URL config in `frontend/src/lib/api/config.ts` (and related API utility callers/tests).
- Sequence and operational docs: `docs/frontend-backend-sequence.md`, `docs/runbooks/run.md`.
- Frontend tests for auth/orchestration route forwarding behavior.
