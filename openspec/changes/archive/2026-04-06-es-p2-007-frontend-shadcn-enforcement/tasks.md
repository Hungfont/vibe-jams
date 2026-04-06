## 1. Policy Scope and Canonical Primitive Inventory

- [x] 1.1 Align shadcn instruction scope to frontend source globs and document canonical primitive categories
- [x] 1.2 Define exception metadata contract (owner, rationale, review status) for non-shadcn primitive cases
- [x] 1.3 Add contributor-facing guidance for choosing existing approved primitives before creating new UI primitives

## 2. Frontend Conformance Enforcement

- [x] 2.1 Add deterministic frontend conformance check to lint/CI pipeline for duplicate primitive detection
- [x] 2.2 Implement violation reporting with actionable error messages and affected file paths
- [x] 2.3 Add baseline/allowlist handling so approved exceptions are honored while new violations fail

## 3. Incremental Primitive Migration

- [x] 3.1 Audit current `frontend/src/components/ui/**` primitives against approved shadcn-equivalent categories
- [x] 3.2 Migrate highest-risk duplicate primitives to approved shadcn-style implementations or wrappers
- [x] 3.3 Update feature components to consume approved primitive imports only

## 4. Validation and Documentation

- [x] 4.1 Add tests/verification for conformance checker behavior (pass and fail paths)
- [x] 4.2 Update `docs/runbooks/run.md` with new frontend shadcn conformance validation flow
- [x] 4.3 Run frontend lint/test/build with conformance checks enabled and capture final evidence
