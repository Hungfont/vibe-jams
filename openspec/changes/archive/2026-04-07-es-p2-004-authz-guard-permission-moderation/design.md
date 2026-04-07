## Context

Jam policy flows currently defer host-only authorization to jam-service, which makes denials happen late in the request path and duplicates policy assumptions across service entrypoints. ES-P2-001 and ES-P2-002 need one upstream authoritative decision layer so delegated command routes apply the same host policy before downstream mutations.

This change is backend cross-cutting because it introduces reusable authorization guard interfaces in api-service delegated routes, updates moderation and permission authorization boundary behavior, and keeps jam-service checks as defense-in-depth. It depends on stable auth claim fields from auth-service and jam session state lookup used for host role determination.

## Goals / Non-Goals

**Goals:**
- Centralize host or guest authorization decisions for permission and moderation command entrypoints at api-service.
- Ensure host-only policy actions fail fast with deterministic `403 host_only`.
- Provide reusable api-service guard interfaces for ES-P2-001 and ES-P2-002 integration without duplicating host policy logic.
- Keep jam-service guard active as downstream safety net.

**Non-Goals:**
- Implementing permission projection business side effects owned by ES-P2-001.
- Rewriting moderation state transitions or event fanout behavior owned by ES-P2-002.
- Changing auth-service token validation protocol or claim issuance format.
- Removing jam-service guard protections.

## Decisions

1. Introduce centralized authorization guard interface in api-service delegated route layer
- Decision: define one reusable guard contract in api-service that evaluates actor claims, delegated command intent, and jam session host context before forwarding policy commands.
- Rationale: keeps policy logic in one upstream place and allows ES-P2-001 or ES-P2-002 to consume consistent decisions.
- Alternatives considered:
  - Keep per-handler host checks: rejected due to duplicate logic and drift risk.
  - Push all authorization to auth-service: rejected because command-level host role context is session-local and requires jam session state.

2. Fail-fast host-only denials with normalized contract
- Decision: enforce `403 host_only` in api-service before downstream proxying when guard denies host-only actions.
- Rationale: deterministic denial semantics and reduced partial side-effect risk.
- Alternatives considered:
  - Return mixed 401/403 by handler: rejected due to inconsistent client behavior.

3. Preserve downstream defense-in-depth guard
- Decision: keep jam-service host-only guard checks active after upstream api-service authorization.
- Rationale: protects direct/internal call paths and prevents policy bypass if upstream routing changes.
- Alternatives considered:
  - Remove jam-service guard entirely: rejected due to increased bypass risk.

4. Keep side effects in dependent stories
- Decision: ES-P2-004 only defines and wires authorization decision contracts and boundary ownership; ES-P2-001 and ES-P2-002 retain business transitions.
- Rationale: preserves story ownership boundaries and minimizes refactor blast radius.

## Risks / Trade-offs

- [Risk] api-service introduces extra jam state lookup before forwarding policy commands. -> Mitigation: bounded timeout and deterministic dependency error mapping.
- [Risk] Existing moderation and permission handlers may still contain local policy checks after migration. -> Mitigation: add integration tests proving api-service guard path is authoritative for delegated routes.
- [Trade-off] Additional upstream guard abstraction increases interface count. -> Mitigation: keep interfaces narrow and route-family oriented.
- [Risk] Claim contract changes in auth-service can break upstream guard evaluation. -> Mitigation: maintain strict dependency on normalized auth claim contract and fail fast when required fields are missing.
