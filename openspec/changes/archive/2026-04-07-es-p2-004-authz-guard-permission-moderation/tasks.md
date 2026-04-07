## 1. Centralized Authorization Guard Contract

- [x] 1.1 Define centralized jam policy authorization guard interfaces for host or guest decisions.
- [x] 1.2 Add reusable decision context model sourced from normalized auth claims and jam role context.
- [x] 1.3 Ensure host-only denials map to deterministic `403 host_only` at command entrypoints.

## 2. Moderation and Permission Entrypoint Integration

- [x] 2.1 Refactor moderation command authorization to consume centralized guard decision path.
- [x] 2.2 Wire permission command authorization hooks to consume centralized guard interfaces for ES-P2-001 integration.
- [x] 2.3 Remove duplicated host policy checks from command handlers where centralized guard is authoritative.

## 3. Audit Metadata and Event Contract Alignment

- [x] 3.1 Extend policy audit payload contract to include actor identity metadata and decision outcome (`accepted` or `denied`).
- [x] 3.2 Emit denied authorization audit entries for host-only policy attempts without business side effects.
- [x] 3.3 Preserve accepted moderation and permission audit entries with standardized actor metadata fields.

## 4. Contract and Behavior Tests

- [x] 4.1 Add tests proving non-host actors cannot execute host-only moderation or permission policy actions.
- [x] 4.2 Add tests proving authorization behavior is consistent across jam command entrypoints.
- [x] 4.3 Add tests validating audit events for denied and accepted policy decisions include actor metadata.

## 5. Validation and Runbook Updates

- [x] 5.1 Execute targeted backend test suites covering centralized authZ guard integration and audit payload compatibility.
- [x] 5.2 Update docs/runbooks/run.md with concise test flow for host-only guard consistency and policy audit metadata verification.

## 6. Boundary Shift to API-Service Authoritative Host-Only Checks

- [x] 6.1 Add api-service delegated-route policy guard interfaces for moderation and permission host-only checks.
- [x] 6.2 Enforce deterministic `403 host_only` in api-service before proxying denied policy commands.
- [x] 6.3 Keep jam-service guard active as defense-in-depth and verify direct jam-service paths remain protected.
- [x] 6.4 Add/extend api-service proxy tests for host and non-host delegated policy command behavior.
- [x] 6.5 Update docs/runbooks/run.md to reflect authoritative api-service host-only enforcement flow and validation commands.
