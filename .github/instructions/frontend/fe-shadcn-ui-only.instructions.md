---
description: "Use when building or updating frontend UI components. Enforce shadcn/ui-first component usage and avoid custom components when shadcn/ui provides an equivalent."
applyTo: "**/*.ts"
---
# shadcn/ui-First Component Policy

This project uses shadcn/ui as the default and exclusive component system for UI primitives.

## Required Rules

- Always use shadcn/ui components when an equivalent exists (for example: button, input, select, dialog, drawer, tabs, tooltip, table, form, toast).
- Do not create custom UI primitive components that duplicate existing shadcn/ui behavior.
- Prefer composing feature components from shadcn/ui building blocks instead of re-implementing base UI patterns.
- If an equivalent component is not available in the current codebase, add the shadcn/ui component first, then compose from it.
- Only create a custom component when there is no shadcn/ui equivalent and the user has requested or approved that exception.

## Implementation Notes

- Keep imports consistent with the repository's shadcn/ui import path conventions.
- Preserve accessibility behavior from shadcn/ui primitives; do not regress keyboard/focus semantics.
- Keep styling changes aligned with existing design tokens and utility classes.
