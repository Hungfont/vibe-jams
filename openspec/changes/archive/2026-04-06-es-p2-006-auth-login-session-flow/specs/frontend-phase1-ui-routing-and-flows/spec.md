## ADDED Requirements

### Requirement: Frontend SHALL expose auth session routes through frontend-owned API boundary
The frontend SHALL expose auth lifecycle API routes and SHALL keep browser-to-backend communication within frontend-owned `/api/**` routes.

#### Scenario: Browser login submission
- **WHEN** a user submits login form from frontend login page
- **THEN** browser calls `POST /api/auth/login`, and frontend route forwards request to auth-service public login endpoint

#### Scenario: Browser session lifecycle actions
- **WHEN** frontend triggers refresh, logout, or me retrieval
- **THEN** browser calls `POST /api/auth/refresh`, `POST /api/auth/logout`, or `GET /api/auth/me` and routes normalize success/error envelopes

### Requirement: Login page SHALL provide Spotify-like visual hierarchy with accessible form behavior
The frontend SHALL render `/login` using Spotify-like dark visual hierarchy and SHALL preserve accessibility and deterministic submit state behavior.

#### Scenario: Login page first render
- **WHEN** user navigates to `/login`
- **THEN** page renders brand-forward dark layout, credential form, and deterministic loading/error/success UI states

#### Scenario: Login form validation and submit
- **WHEN** user submits invalid or valid credentials
- **THEN** client-side zod validation blocks invalid payloads, valid payloads submit through frontend auth route, and server errors map to actionable messages
