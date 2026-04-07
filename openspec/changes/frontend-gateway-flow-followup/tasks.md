## 1. Frontend Gateway Config Alignment

- [x] 1.1 Add api-gateway base URL configuration for frontend server routes (`API_GATEWAY_URL`) with deterministic local default.
- [x] 1.2 Ensure auth and BFF route handlers consume gateway-targeted service selection without changing browser `/api/**` contracts.

## 2. Frontend Auth Route Upstream Alignment

- [x] 2.1 Update `POST /api/auth/login` route handler to forward to api-gateway auth login path.
- [x] 2.2 Update `POST /api/auth/refresh` route handler to forward to api-gateway auth refresh path.
- [x] 2.3 Update `POST /api/auth/logout` route handler to forward to api-gateway auth logout path.
- [x] 2.4 Update `GET /api/auth/me` route handler to forward to api-gateway auth me path.

## 3. Frontend BFF Orchestration Upstream Alignment

- [x] 3.1 Update `POST /api/bff/jam/{jamId}/orchestration` route handler to forward through api-gateway BFF orchestration path.
- [x] 3.2 Preserve existing envelope parsing and error normalization behavior for orchestration responses.

## 4. Tests and Regression Coverage

- [x] 4.1 Add or update auth route tests to verify gateway-targeted upstream service selection and paths.
- [x] 4.2 Add or update orchestration route tests to verify gateway-targeted upstream service selection.
- [x] 4.3 Run frontend tests and verify touched route handlers pass.

## 5. Documentation and Runbook Alignment

- [x] 5.1 Update `docs/frontend-backend-sequence.md` for gateway-first auth/BFF frontend server routing.
- [x] 5.2 Update `docs/runbooks/run.md` with concise test flow for validating frontend gateway alignment.
