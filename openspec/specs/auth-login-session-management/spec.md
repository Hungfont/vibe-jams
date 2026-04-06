# auth-login-session-management Specification

## Purpose
TBD - created by archiving change es-p2-006-auth-login-session-flow. Update Purpose after archive.
## Requirements
### Requirement: Auth-service SHALL provide public login endpoint
The system SHALL expose `POST /v1/auth/login` that validates submitted credentials and issues a short-lived access token plus rotating refresh session.

#### Scenario: Login succeeds with valid credentials
- **WHEN** a user submits valid credentials to `POST /v1/auth/login`
- **THEN** auth-service returns success with access token metadata and refresh session context suitable for secure cookie issuance

#### Scenario: Login fails with invalid credentials
- **WHEN** a user submits invalid credentials to `POST /v1/auth/login`
- **THEN** auth-service returns deterministic unauthorized error without leaking credential validation internals

### Requirement: Refresh flow SHALL use opaque token rotation and revocation
The system SHALL expose `POST /v1/auth/refresh` and SHALL rotate opaque refresh tokens on every successful refresh.

#### Scenario: Refresh succeeds and rotates token
- **WHEN** a valid non-revoked refresh token is submitted
- **THEN** auth-service issues new access token, rotates refresh token, and revokes the old refresh token

#### Scenario: Reuse detection revokes token family
- **WHEN** a previously rotated refresh token is replayed
- **THEN** auth-service revokes the refresh token family/session and returns deterministic unauthorized response

### Requirement: Logout flow SHALL revoke active refresh session
The system SHALL expose `POST /v1/auth/logout` to revoke the active refresh session and prevent further refresh.

#### Scenario: Logout succeeds
- **WHEN** an authenticated session calls `POST /v1/auth/logout`
- **THEN** auth-service revokes refresh session state and returns success

#### Scenario: Logout with invalid session context
- **WHEN** logout request is missing valid session context
- **THEN** auth-service returns deterministic unauthorized response

### Requirement: Auth profile endpoint SHALL expose normalized current user claims
The system SHALL expose `GET /v1/auth/me` that returns normalized claims for the currently authenticated user.

#### Scenario: Me endpoint success
- **WHEN** valid authenticated context is provided
- **THEN** endpoint returns `userId`, `plan`, `sessionState`, and `scope`

#### Scenario: Me endpoint unauthorized
- **WHEN** no valid auth context is provided
- **THEN** endpoint returns deterministic unauthorized error

### Requirement: Non-test runtime SHALL persist refresh/session state outside memory
The system SHALL persist refresh/session state using configured durable backend and MUST reject in-memory mode outside test profile.

#### Scenario: Non-test startup with in-memory backend
- **WHEN** runtime profile is non-test and auth store backend is in-memory
- **THEN** auth-service startup fails with deterministic invalid-runtime-adapter error

#### Scenario: Non-test startup with durable backend
- **WHEN** runtime profile is non-test and auth store backend is postgres
- **THEN** auth-service starts and persists refresh/session lifecycle in durable storage

### Requirement: Auth lifecycle SHALL enforce abuse controls and emit audit events
Auth-service SHALL enforce rate limiting and lockout/backoff and SHALL emit audit records for login/refresh/logout outcomes.

#### Scenario: Excessive failed login attempts
- **WHEN** repeated login failures exceed configured threshold
- **THEN** auth-service applies lockout/backoff behavior and returns deterministic throttled/blocked response

#### Scenario: Auth lifecycle event audit
- **WHEN** login, refresh, logout, or refresh-reuse events occur
- **THEN** auth-service emits auditable event records with actor/session metadata and outcome

