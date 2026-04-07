## 1. Contract and Flow Baseline

- [x] 1.1 Finalize and approve sequence + mapping contract for mandatory BFF-first non-auth HTTP flows.
- [x] 1.2 Enumerate frontend route families that must move to BFF-first path (jam, playback, catalog, realtime bootstrap).

## 2. API-service BFF Routing Surface

- [x] 2.1 Add BFF handlers/routes for jam command/query delegation currently called directly by frontend routes.
- [x] 2.2 Add BFF handlers/routes for playback command delegation.
- [x] 2.3 Add BFF handlers/routes for catalog lookup delegation.
- [x] 2.4 Add BFF handler/route for realtime ws-config bootstrap delegation.
- [x] 2.5 Enforce identity propagation and deterministic error envelope mapping on all new BFF route families.

## 3. Frontend Route Rewiring

- [x] 3.1 Rewire frontend jam API routes to call BFF-first upstream paths via api-gateway.
- [x] 3.2 Rewire frontend playback API route to call BFF-first upstream path via api-gateway.
- [x] 3.3 Rewire frontend catalog lookup route to call BFF-first upstream path via api-gateway.
- [x] 3.4 Rewire frontend realtime ws-config route to call BFF-first upstream path via api-gateway.

## 4. Validation and Documentation Alignment

- [x] 4.1 Add/update backend tests for BFF route delegation and identity propagation.
- [x] 4.2 Add/update frontend route tests to assert BFF-first service selection.
- [x] 4.3 Update `docs/frontend-backend-sequence.md` route mapping and notes to match final implementation.
- [x] 4.4 Update `docs/runbooks/run.md` with new BFF-first test/verification flows.
- [x] 4.5 Run focused backend/frontend validation suites for affected flows.
