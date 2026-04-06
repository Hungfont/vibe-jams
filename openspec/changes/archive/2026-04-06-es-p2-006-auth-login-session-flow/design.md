## Context

The current auth-service only exposes internal token validation (`/internal/v1/auth/validate`) and relies on fixed in-memory token fixtures. The frontend currently validates and forwards auth context through frontend-owned API routes but does not own a user-facing login/session lifecycle.

This change introduces production-grade AuthN flows (login, refresh, logout, me), persistent refresh/session state, and a Spotify-inspired login UX while preserving existing internal validate semantics used by service-to-service guards.

## Goals / Non-Goals

**Goals:**
- Add auth-service public AuthN APIs: login, refresh, logout, me.
- Keep `/internal/v1/auth/validate` unchanged for internal guard compatibility.
- Issue short-lived access tokens with stable claims and scope.
- Implement opaque rotating refresh token lifecycle with revoke list.
- Persist refresh/session state in Postgres runtime path and disallow in-memory backend in non-test profiles.
- Add frontend API boundary routes for auth flows and a Spotify-like login page.
- Harden security controls: rate limit, lockout/backoff, audit events, JWT key rotation support.

**Non-Goals:**
- No social login/OAuth provider integration in this change.
- No MFA or device fingerprint trust framework in this change.
- No redesign of jam/playback authorization policies beyond consuming validated claims.

## Decisions

1. Public-vs-internal auth API split
- Public endpoints in auth-service serve login lifecycle:
  - `POST /v1/auth/login`
  - `POST /v1/auth/refresh`
  - `POST /v1/auth/logout`
  - `GET /v1/auth/me`
- Internal endpoint remains:
  - `POST /internal/v1/auth/validate`
- Rationale: isolates browser AuthN from service guard validation, minimizes breakage to current flows.

2. Access token + refresh token model
- Access token: JWT, short TTL (default 10m), includes `userId`, `plan`, `sessionState`, `scope`, `sid`, `exp`, and `kid`.
- Refresh token: opaque random token, stored as hash server-side; one-time rotation per refresh.
- Reuse detection revokes the full token family/session.
- Rationale: limits blast radius of access token leaks and supports deterministic revocation.

3. Persistent session backend and runtime policy
- Introduce `SessionStore` interface with runtime implementations:
  - `postgres` (default non-test runtime)
  - `inmemory` (test/dev only)
- Enforce startup failure when non-test profile uses `inmemory`.
- Rationale: aligns with existing project runtime-hardening pattern.

4. Frontend-owned boundary and secure cookies
- Browser submits to frontend routes (`/api/auth/*`) only.
- Frontend route forwards to auth-service and sets HttpOnly, Secure, SameSite cookies.
- CSRF token check required on refresh/logout when cookie auth is used.
- Rationale: avoids direct browser-to-service coupling and reduces XSS token exposure.

5. Login UX implementation strategy
- Implement `/login` page using existing shadcn primitives + Tailwind, visually close to Spotify login hierarchy (dark canvas, centered auth card, brand-forward CTA, divider/options layout).
- Keep implementation accessible (keyboard/focus/labels) and responsive.
- Rationale: satisfy requested visual direction while preserving project component standards.

6. Abuse and audit controls
- Add rate limit keys per IP + normalized identity (`email`/`userId`).
- Add lockout/backoff window after repeated failures.
- Emit audit events for login success/failure, refresh success/failure, logout, refresh-reuse detection.
- Rationale: baseline abuse resistance and operational traceability.

## Risks / Trade-offs

- [Risk: JWT key mis-rotation invalidates active sessions] -> Mitigation: key ring with active + previous keys and `kid`-based verification window.
- [Risk: Refresh token theft before rotation] -> Mitigation: hashed storage, token family reuse detection, immediate family revocation.
- [Risk: Cookie-based AuthN introduces CSRF surface] -> Mitigation: SameSite policy + CSRF token validation on state-changing routes.
- [Risk: Login rate limiter blocks legitimate spikes] -> Mitigation: identity-aware thresholds, bounded backoff, explicit audit observability.
- [Risk: Frontend/backend contract drift] -> Mitigation: shared envelope shape, route-level tests, docs sequence update in same change.

## Migration Plan

1. Add auth-service domain models and store interfaces for user credential check, sessions, refresh token rotation, and audit log write path.
2. Implement public auth endpoints while retaining internal validate endpoint.
3. Introduce config/runtime policy for auth store backend and JWT key set.
4. Add frontend `/api/auth/login|refresh|logout|me` routes and typed client helpers.
5. Add `/login` page and form flow with zod validation + deterministic error mapping.
6. Update sequence/runbook docs and run backend/frontend validations.

Rollback:
- Disable new public auth routes and frontend login route usage via feature flags.
- Keep internal validate path intact so existing jam/playback flows continue.
- Revoke newly issued refresh sessions if rollback is triggered after partial rollout.

## Open Questions

- Should scope be represented as array claim (`scope[]`) or space-delimited string (`scope`)?
- Should login lockout state be global per identity across IPs, or hybrid identity+IP window only?
- Should `GET /v1/auth/me` rely only on bearer header, or support cookie fallback as primary path?
