# AGENTS Guide

## Rule Invocation Policy
- The rule `docs-story-workflow` is opt-in only.
- Only apply it when explicitly referenced with `@docs-story-workflow`.
- Do not auto-apply this rule from file paths, filename patterns, or inferred context.

## When to Use `@docs-story-workflow`
- While editing story documents in `docs/story/`.
- While expanding user stories (US) to engineer stories (ES).
- While creating or updating per-ES files in `docs/story/engineer-stories/`.
- While aligning related story planning details in `docs/lld/` for story workflow tasks.

## Execution Notes
- If `@docs-story-workflow` is not explicitly present in the request, do not enforce this rule.
- Keep outputs consistent with the existing docs story format when the rule is invoked.

## Security Guidelines
Before ANY commit:

No hardcoded secrets (API keys, passwords, tokens)
All user inputs validated
SQL injection prevention (parameterized queries)
XSS prevention (sanitized HTML)
CSRF protection enabled
Authentication/authorization verified
Rate limiting on all endpoints
Error messages don't leak sensitive data
Secret management: NEVER hardcode secrets. Use environment variables or a secret manager. Validate required secrets at startup. Rotate any exposed secrets immediately.

If security issue found: STOP → use security-reviewer agent → fix CRITICAL issues → rotate exposed secrets → review codebase for similar issues.


## Coding Style
Immutability (CRITICAL): Always create new objects, never mutate. Return new copies with changes applied.

File organization: Many small files over few large ones. 200-400 lines typical, 800 max. Organize by feature/domain, not by type. High cohesion, low coupling.

Error handling: Handle errors at every level. Provide user-friendly messages in UI code. Log detailed context server-side. Never silently swallow errors.

Input validation: Validate all user input at system boundaries. Use schema-based validation. Fail fast with clear messages. Never trust external data.

Code quality checklist:

Functions small (<50 lines), files focused (<800 lines)
No deep nesting (>4 levels)
Proper error handling, no hardcoded values
Readable, well-named identifiers


## Architecture Patterns
API response format: Consistent envelope with success indicator, data payload, error message, and pagination metadata.

Repository pattern: Encapsulate data access behind standard interface (findAll, findById, create, update, delete). Business logic depends on abstract interface, not storage mechanism.

Skeleton projects: Search for battle-tested templates, evaluate with parallel agents (security, extensibility, relevance), clone best match, iterate within proven structure.