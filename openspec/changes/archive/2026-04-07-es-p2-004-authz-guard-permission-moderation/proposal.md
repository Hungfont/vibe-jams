## Why

Permission and moderation command paths currently rely on downstream jam-service checks, which delays host-only denials and allows policy enforcement drift across entrypoints. ES-P2-004 is updated to make api-service the authoritative upstream policy layer so host-only actions are denied before downstream side effects while preserving jam-service checks as defense-in-depth.

## What Changes

- Introduce an api-service policy authorization guard for host-only moderation and permission command families before proxy delegation.
- Require api-service delegated jam policy routes to run host-only checks using gateway-injected identity claims and jam session state lookup.
- Standardize host-only denial behavior to deterministic `403 host_only` at api-service.
- Keep jam-service host-only checks as downstream defense-in-depth and compatibility fallback.
- Preserve business side effects in ES-P2-001 and ES-P2-002; ES-P2-004 updates policy decision ownership boundary only.

## Capabilities

### New Capabilities

- jam-policy-authorization-guard: centralized host or guest authorization guard contract anchored in api-service for jam permission and moderation commands.

### Modified Capabilities

- api-service-bff-microservice-routing: delegated jam policy routes must execute host-only pre-checks before forwarding to jam-service.
- jam-moderation-controls: moderation command authorization must be denied upstream at api-service for non-host actors while preserving jam-service fallback guard.
- auth-claim-contract: claim fields required by policy authorization are consumed consistently in api-service guard decisions.

## Impact

- Backend api-service BFF delegated route authorization path for moderation and permission policy commands.
- Shared policy decision contract consumed by ES-P2-001 and ES-P2-002 via api-service pre-check integration.
- Host-only denial contract alignment across api-gateway identity injection, api-service policy enforcement, and jam-service fallback guard.
- Contract and handler tests for upstream host-only denials plus downstream defense-in-depth behavior.
