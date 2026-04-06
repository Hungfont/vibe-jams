## Why

Current auth flow only validates pre-issued bearer tokens and does not provide a production-grade login/session lifecycle. This blocks real user sign-in UX and secure token refresh/logout behavior needed for scale and abuse control.

## What Changes

- Add public AuthN APIs in auth-service for login, refresh, logout, and profile introspection.
- Keep existing internal auth validation API for service-to-service authorization guards.
- Introduce short-lived access tokens and rotating opaque refresh tokens with server-side revocation state.
- Add persistent refresh/session storage with runtime policy that disallows in-memory store in non-test runtime profiles.
- Add frontend API boundary routes for auth login/refresh/logout/me and a Spotify-like login page UX.
- Issue HttpOnly/Secure/SameSite cookies from frontend server boundary to reduce token exposure in browser JavaScript.
- Add auth security controls: rate limiting, lockout/backoff, audit trail for login lifecycle events, and key rotation support for JWT signing.

## Capabilities

### New Capabilities
- `auth-login-session-management`: Public login, refresh, logout, and me flows with secure token/session lifecycle and revocation semantics.

### Modified Capabilities
- `auth-claim-contract`: Claim contract expands to include scope semantics and stable claim derivation from issued access tokens.
- `frontend-phase1-ui-routing-and-flows`: Frontend route mapping extends with auth login/session routes while preserving frontend-owned API boundary.

## Impact

- Affected code: `backend/auth-service/**`, `frontend/src/app/api/auth/**`, `frontend/src/app/login/**`, `frontend/src/lib/api/**`, `frontend/src/lib/jam/**` (auth usage alignment where needed).
- Affected APIs:
  - New public auth-service endpoints: `POST /v1/auth/login`, `POST /v1/auth/refresh`, `POST /v1/auth/logout`, `GET /v1/auth/me`.
  - Existing internal endpoint retained: `POST /internal/v1/auth/validate`.
- Affected security controls: rate limit keys, lockout policy state, refresh token rotation/revocation list, JWT key management.
- Affected docs: `docs/frontend-backend-sequence.md`, `docs/runbooks/run.md`.
