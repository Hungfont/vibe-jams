---
description: "Use when implementing or updating backend authentication, authorization, gateway middleware, BFF routing, or protected microservice endpoints. Enforces Client -> API Gateway -> API Service (BFF) -> Microservices flow."
name: "Backend AuthN AuthZ Flow Guardrail"
applyTo:
  - "backend/api-gateway/**/*.go"
  - "backend/api-service/**/*.go"
  - "backend/auth-service/**/*.go"
  - "backend/catalog-service/**/*.go"
  - "backend/jam-service/**/*.go"
  - "backend/playback-service/**/*.go"
  - "backend/rt-gateway/**/*.go"
---
# Backend AuthN/AuthZ Flow Guardrail

- These rules are default preferences and should be followed unless the user explicitly requests a different approach.
- The canonical protected-call path is: Client -> API Gateway -> API Service (BFF) -> Microservices.

## Required Flow Rules

- API Gateway is the only public ingress for backend APIs. Do not expose direct client access to API Service or downstream microservices.
- For protected requests, API Gateway must validate identity with auth-service (`POST /internal/v1/auth/validate`) before forwarding.
- API Gateway should prefer bearer `Authorization` and may fall back to auth cookies when header auth is absent.
- After successful validation, API Gateway must inject auth context headers and forward them downstream:
  - `X-Auth-UserId`
  - `X-Auth-Plan`
  - `X-Auth-SessionState`
  - `X-Auth-Scope`
- API Service (BFF) must treat `X-Auth-*` as the source of identity context. It must reject missing/invalid identity context with deterministic `401 unauthorized`.
- API Service should orchestrate and route protected business calls to microservices. New protected business endpoints should be introduced through BFF routing first.
- Microservices must enforce domain-level authorization (resource ownership, role/plan/session checks) using forwarded identity context and return deterministic `403` for forbidden actions.

## Allowed Exceptions

- Auth lifecycle endpoints (for example login/refresh/logout/me) may terminate at auth-service through API Gateway without BFF orchestration.
- Internal service-to-service operations that are not client-originated may bypass the BFF only when explicitly requested and documented.

## Disallowed Without Explicit User Approval

- Adding direct client -> API Service or client -> microservice routes.
- Calling auth-service token validation directly from API Service for request-time auth checks that should be handled by API Gateway.
- Accepting protected requests in API Service without validated `X-Auth-UserId` context.
- Trusting client-supplied `X-Auth-*` headers without gateway validation.

## Change Management Rule

- When authN/authZ flow contracts, identity headers, or gateway validation behavior changes, update `docs/frontend-backend-sequence.md` in the same change.
- Keep auth behavior consistent across gateway middleware, BFF handlers, and downstream service handlers to avoid split authorization semantics.