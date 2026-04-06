## 1. Auth-Service Public AuthN APIs

- [x] 1.1 Add auth-service domain models for login credentials, access token claims with scope, refresh session state, and audit event payloads
- [x] 1.2 Implement `POST /v1/auth/login` with credential validation, short-lived access token issuance, and refresh token creation
- [x] 1.3 Implement `POST /v1/auth/refresh` with opaque refresh token rotation and reuse-detection family revocation
- [x] 1.4 Implement `POST /v1/auth/logout` and `GET /v1/auth/me` with deterministic auth error mapping
- [x] 1.5 Keep `/internal/v1/auth/validate` behavior stable while validating new scope-aware claims contract

## 2. Durable Session Store and Security Controls

- [x] 2.1 Add persistent auth store adapter (postgres) for refresh/session state with runtime backend interface
- [x] 2.2 Enforce runtime policy that rejects in-memory auth store in non-test profiles
- [x] 2.3 Add rate limiting and lockout/backoff for login and refresh endpoints
- [x] 2.4 Add audit logging/events for login success/failure, refresh success/failure, logout, and refresh-reuse detection
- [x] 2.5 Add JWT signing key ring and `kid`-based validation path for key rotation compatibility

## 3. Frontend API Boundary and Login UX

- [x] 3.1 Add frontend routes `POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`, and `GET /api/auth/me`
- [x] 3.2 Implement secure cookie handling (HttpOnly, Secure, SameSite) and CSRF protection on state-changing auth routes
- [x] 3.3 Add typed auth client helpers and zod schemas for login/session flows
- [x] 3.4 Build `/login` page UI with Spotify-like visual structure using shadcn primitives and Tailwind utilities
- [x] 3.5 Add frontend tests for login form validation, submit/error mapping, and API route envelope behavior

## 4. Integration, Docs, and Validation

- [x] 4.1 Update `docs/frontend-backend-sequence.md` with auth login/refresh/logout/me route mapping and flow notes
- [x] 4.2 Update `docs/runbooks/run.md` with login/session lifecycle test flow and security edge-case outcomes
- [x] 4.3 Run auth-service test suite and targeted endpoint tests for login/refresh/logout/me
- [x] 4.4 Run frontend lint/test/build and verify contract parity between frontend auth routes and auth-service APIs
