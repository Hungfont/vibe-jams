## Context

After BFF-first rewiring, some gateway-bound requests rely on cookie session context rather than explicit Authorization header. Gateway middleware currently requires `Authorization` for protected routes and returns `missing_credentials` when absent. In parallel, Swagger/OpenAPI docs for api-service and api-gateway do not fully represent the expanded route surface and operational expectations.

## Goals / Non-Goals

**Goals:**
- Eliminate false `missing_credentials` failures for valid session-cookie contexts at gateway boundary.
- Preserve deterministic unauthorized behavior when no credentials are present.
- Provide up-to-date Swagger/OpenAPI docs for api-gateway and api-service BFF routes.
- Keep contracts explicit for BFF-first route families.

**Non-Goals:**
- Redesign auth token issuance or storage.
- Introduce external OpenAPI generation toolchains.
- Rework frontend UX flows beyond auth-context forwarding compatibility.

## Decisions

1. Gateway credential extraction fallback
- Decision: read bearer from `Authorization` first, then fallback to `auth_token`/`token` cookie for protected routes.
- Rationale: aligns with existing frontend server-route cookie boundary and avoids breaking protected calls that legitimately rely on cookie transport.

2. Keep error code semantics stable
- Decision: retain `missing_credentials` only when both header and cookie token are absent.
- Rationale: preserves client handling logic and deterministic error contract.

3. Swagger updates in-code
- Decision: extend existing in-code OpenAPI docs (api-service) and add equivalent gateway OpenAPI/Swagger endpoints.
- Rationale: current services already ship static in-code swagger surfaces; keeping same pattern is low risk.

4. Test coverage expansion
- Decision: add tests for cookie fallback auth behavior and swagger endpoint availability/content.
- Rationale: prevents regressions and documents expected behavior in executable form.

## Risks / Trade-offs

- [Risk] Cookie fallback broadens accepted credential path at gateway. -> Mitigation: only parse known auth cookie names and keep route protection intact.
- [Risk] Swagger docs drift again as routes evolve. -> Mitigation: add tests checking critical documented paths and keep runbook evidence updated.
- [Risk] Mixed auth transport assumptions across services. -> Mitigation: explicitly document precedence (Authorization over cookie token).

## Migration Plan

1. Implement and test gateway credential fallback logic.
2. Update api-service OpenAPI route documentation.
3. Add gateway swagger/openapi endpoints and docs.
4. Run focused backend/frontend tests.
5. Update runbook/sequence references if behavior text changes.

## Open Questions

- Should gateway explicitly reject malformed bearer tokens extracted from cookie with a dedicated error code versus `invalid_token`?
