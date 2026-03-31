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
