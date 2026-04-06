---
description: "Use when building or updating frontend UI components. Enforce shadcn/ui-first component usage and avoid custom components when shadcn/ui provides an equivalent."
applyTo: "frontend/src/**/*.ts, frontend/src/**/*.tsx, frontend/src/**/*.js, frontend/src/**/*.jsx"
---
# shadcn/ui-First Component Policy

This project uses shadcn/ui as the default and exclusive component system for UI primitives.

## Required Rules

- Always use shadcn/ui components when an equivalent exists (for example: button, input, select, dialog, drawer, tabs, tooltip, table, form, toast).
- Do not create custom UI primitive components that duplicate existing shadcn/ui behavior.
- Prefer composing feature components from shadcn/ui building blocks instead of re-implementing base UI patterns.
- If an equivalent component is not available in the current codebase, add the shadcn/ui component first, then compose from it.
- Only create a custom component when there is no shadcn/ui equivalent and the user has requested or approved that exception.

## Canonical Primitive Inventory

- The approved primitive inventory is defined in `frontend/config/shadcn-primitive-inventory.json`.
- New primitive files under `frontend/src/components/ui/**` must match approved inventory names unless exception metadata is registered.

## Exception Workflow

- Exceptions must be recorded in `frontend/config/shadcn-exceptions.json`.
- Every exception entry must include:
	- `componentPath`
	- `owner`
	- `rationale`
	- `reviewStatus` (`approved` or `temporary`)
- Missing metadata or invalid review status is non-compliant and must fail conformance checks.

## Contributor Guidance

- Before adding a primitive, check inventory first and reuse existing primitives in `frontend/src/components/ui/**`.
- Prefer feature-level composition in `frontend/src/components/**` over creating new primitive files.
- If no equivalent exists, request approval and add complete exception metadata before implementation.

## Implementation Notes

- Keep imports consistent with the repository's shadcn/ui import path conventions.
- Preserve accessibility behavior from shadcn/ui primitives; do not regress keyboard/focus semantics.
- Keep styling changes aligned with existing design tokens and utility classes.
