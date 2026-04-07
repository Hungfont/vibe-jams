## 1. Implementation

- [x] 1.1 Update api-gateway auth middleware to support cookie-token fallback (`auth_token` / `token`) when `Authorization` is missing, while preserving existing public-route bypass and deterministic `missing_credentials` / `invalid_token` mapping.
- [x] 1.2 Add or update api-gateway middleware unit tests covering Authorization priority, cookie fallback success, missing credentials, and invalid token behavior.
- [x] 1.3 Add Swagger/OpenAPI support to api-gateway with `GET /swagger` and `GET /swagger/openapi.json`, including documented health and proxy route families.
- [x] 1.4 Expand api-service BFF OpenAPI document to include delegated BFF route families currently used by frontend BFF-first flow.
- [x] 1.5 Add/adjust tests validating OpenAPI/Swagger route availability for api-gateway and api-service.
- [x] 1.6 Update docs/frontend-backend-sequence.md if endpoint documentation paths or auth credential expectations changed.
- [x] 1.7 Update docs/runbooks/run.md with concise validation/test flow and observed behavior for this change.

## 2. Validation

- [x] 2.1 Run focused backend tests for api-gateway auth middleware and swagger routes.
- [x] 2.2 Run focused backend tests for api-service BFF OpenAPI endpoints.
- [x] 2.3 Run relevant frontend route tests if auth header/cookie forwarding behavior changes any frontend API route assumptions.
- [x] 2.4 Verify OpenSpec validation passes with no blocking issues.
