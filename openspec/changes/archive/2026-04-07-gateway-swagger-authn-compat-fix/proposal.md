## Why

The current gateway can return `401 missing_credentials` even when authenticated session context exists in cookies for server-mediated calls, and Swagger coverage for Go services is inconsistent with newly added BFF route families. We need deterministic auth extraction compatibility in api-gateway and updated OpenAPI/Swagger docs for affected Go services.

## What Changes

- Fix api-gateway auth middleware to support bearer token extraction from trusted cookie context when `Authorization` header is missing.
- Keep deterministic `missing_credentials` behavior only when neither bearer header nor supported auth cookie token is present.
- Update api-service BFF OpenAPI to document newly exposed BFF-first route families (jam/playback/catalog/realtime bootstrap).
- Add/refresh api-gateway Swagger/OpenAPI endpoints to describe gateway health, auth routes, and BFF proxy ingress behavior.
- Add/adjust backend and frontend tests validating the auth compatibility fix and swagger endpoint behavior.

## Capabilities

### New Capabilities
- `go-service-swagger-openapi`: standardizes Swagger/OpenAPI exposure for Go services impacted by BFF-first routing.

### Modified Capabilities
- `api-gateway-authn-middleware`: extends credential extraction behavior to support auth-token cookie fallback when header auth is absent.
- `api-service-bff-orchestration`: updates API documentation surface to include non-orchestration BFF route families already exposed by implementation.

## Impact

- `backend/api-gateway/internal/gateway/**` for credential extraction and swagger exposure.
- `backend/api-service/internal/bff/openapi.go` for expanded documented endpoints.
- `frontend/src/app/api/realtime/ws-config/route.ts` (if needed) for auth context forwarding alignment.
- Documentation/testing artifacts and runbook validation evidence.
