## 1. API Gateway: New Service Scaffold

- [x] 1.1 Create `backend/api-gateway/` Go module with `go.mod` and `cmd/server/main.go` entry point
- [x] 1.2 Add `internal/config/config.go`: load env vars for `AUTH_SERVICE_URL`, `API_SERVICE_URL`, `GATEWAY_TIMEOUT_AUTH_MS`, `GATEWAY_TIMEOUT_UPSTREAM_MS`, `SERVER_PORT`
- [x] 1.3 Implement `GET /healthz` endpoint returning `{"status":"ok","service":"api-gateway"}`
- [x] 1.4 Wire HTTP server with configurable read/idle/shutdown timeouts matching auth-service pattern

## 2. API Gateway: AuthN Middleware

- [x] 2.1 Implement middleware that strips all inbound `X-Auth-*` headers from every client request before any other processing
- [x] 2.2 Implement public route bypass check: match `POST /v1/auth/login`, `POST /v1/auth/refresh`, `POST /v1/auth/logout`, `GET /v1/auth/me` — skip token validation and forward directly to auth-service
- [x] 2.3 For non-public routes: extract Bearer token from `Authorization` header; return `401 {"error":{"code":"missing_credentials"}}` if absent
- [x] 2.4 Call `POST /internal/v1/auth/validate` on auth-service with `Authorization` header forwarded; apply configurable timeout (`GATEWAY_TIMEOUT_AUTH_MS`)
- [x] 2.5 On auth-service timeout return `503 {"error":{"code":"auth_service_unavailable"}}`; on non-200 return `401 {"error":{"code":"invalid_token"}}`
- [x] 2.6 Check `sessionState` from validate response: if not `valid` (case-insensitive), return `401 {"error":{"code":"invalid_token"}}`
- [x] 2.7 On success inject `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-Scope`, `X-Auth-SessionState` headers; strip `Authorization` header before forwarding

## 3. API Gateway: Routing and Proxy

- [x] 3.1 Implement reverse proxy to api-service for all non-auth routes (apply authN middleware first)
- [x] 3.2 Implement reverse proxy to auth-service for public auth routes (bypass authN middleware)
- [x] 3.3 Apply configurable per-request upstream timeout (`GATEWAY_TIMEOUT_UPSTREAM_MS`); return `504 {"error":{"code":"upstream_timeout"}}` on breach
- [x] 3.4 Forward upstream response status code and body to client unmodified
- [x] 3.5 Write unit tests: missing token, invalid token (auth-service 401), auth-service timeout, valid token with non-valid sessionState, successful proxy with injected headers, client-supplied X-Auth header stripped

## 4. API-Service: Remove Direct Auth Call, Trust Gateway Headers

- [x] 4.1 In `internal/bff/service.go`: remove `authClient.ValidateBearerToken()` call from `Orchestrate()`
- [x] 4.2 Add header extraction helper: read `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-SessionState`, `X-Auth-Scope` from `*http.Request`
- [x] 4.3 Return `401 unauthorized` if `X-Auth-UserId` header is absent or empty
- [x] 4.4 Return `401 unauthorized` if `X-Auth-SessionState` is not `valid` (preserve existing sessionState check semantics)
- [x] 4.5 Populate `sharedauth.Claims` from header values and use in orchestration response (`data.claims`) as before
- [x] 4.6 Remove `HTTPAuthClient` struct and `authClient` dependency from `service.go` and `main.go` if no longer used elsewhere
- [x] 4.7 Update `internal/bff/clients.go`: remove `HTTPAuthClient`; update `JamClient` and `PlaybackClient` to forward `X-Auth-*` headers instead of `Authorization` header
- [x] 4.8 Update api-service unit tests: replace Bearer token setup with `X-Auth-*` header injection; remove auth-service mock

## 5. Deployment and Network Setup
- [x] 5.1 Remove unnecessary code.
- [x] 5.2 Update `docs/runbooks/run.md` to document the new gateway entry point and startup order
- [x] 5.3 Update `docs/frontend-backend-sequence` to document the sequence diagram

## 6. Integration Validation

- [x] 6.1 Integration test: valid token end-to-end through api-gateway → `/internal/v1/auth/validate` → api-service orchestration returns 200 with claims in response
- [x] 6.2 Integration test: request with expired/invalid token → gateway returns `401 invalid_token`
- [x] 6.3 Integration test: request with revoked session (non-valid sessionState) → gateway returns `401 invalid_token`
- [x] 6.4 Integration test: public auth route (`POST /v1/auth/login`) passes through gateway without token verification
- [x] 6.5 Integration test: client-injected `X-Auth-UserId` header is stripped by gateway; downstream receives only gateway-validated values
- [x] 6.6 Integration test: direct request to api-service bypassing gateway (no `X-Auth-UserId`) → api-service returns `401 unauthorized`
