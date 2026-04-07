# jam-policy-authorization-guard Specification

## Purpose
TBD - created by archiving change es-p2-004-authz-guard-permission-moderation. Update Purpose after archive.
## Requirements
### Requirement: Centralized guard SHALL evaluate host or guest policy decisions for jam policy commands at api-service
The api-service delegated route layer SHALL expose a reusable authorization guard interface that evaluates actor claims, jam session role context, and policy command intent for permission and moderation command entrypoints before proxy delegation.

#### Scenario: Permission command consumes centralized guard decision
- **WHEN** a permission policy command is submitted by any actor
- **THEN** api-service evaluates authorization through the centralized guard interface before forwarding to downstream service

#### Scenario: Moderation command consumes centralized guard decision
- **WHEN** a moderation policy command is submitted by any actor
- **THEN** api-service evaluates authorization through the centralized guard interface before forwarding to downstream service

### Requirement: Host-only policy actions SHALL fail fast with deterministic host_only response
For host-only policy actions, the api-service authorization guard SHALL deny non-host actors before downstream proxy side effects and SHALL return deterministic HTTP `403` with error code `host_only`.

#### Scenario: Non-host permission policy action denied
- **WHEN** a non-host actor invokes a host-only permission command
- **THEN** api-service rejects request with `403 host_only` and does not forward command downstream

#### Scenario: Non-host moderation policy action denied
- **WHEN** a non-host actor invokes a host-only moderation command
- **THEN** api-service rejects request with `403 host_only` and does not forward command downstream

### Requirement: Downstream guard SHALL remain active as defense-in-depth
Jam-service host-only guard checks SHALL remain active even when api-service is authoritative for delegated policy routes.

#### Scenario: Direct jam-service call still enforces host-only
- **WHEN** a policy command bypasses api-service and reaches jam-service directly
- **THEN** jam-service guard still denies non-host actor with deterministic `403 host_only`

