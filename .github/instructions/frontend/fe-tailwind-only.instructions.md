---
description: "Use when implementing or updating frontend styling. Enforce Tailwind CSS as the only styling approach and disallow non-Tailwind alternatives unless explicitly approved."
applyTo:
  - "frontend/src/**/*.ts"
  - "frontend/src/**/*.tsx"
  - "frontend/src/**/*.js"
  - "frontend/src/**/*.jsx"
  - "frontend/src/**/*.css"
---
# Tailwind-Only Styling Policy

Use Tailwind CSS utilities as the default and required styling system for frontend work.

## Required Rules

- Use Tailwind utility classes for component and layout styling.
- Use className composition patterns compatible with Tailwind for conditional styling.
- Keep global CSS minimal and limited to Tailwind directives, CSS variables, and truly global resets.
- Prefer design tokens through CSS variables and Tailwind-friendly utility usage.

## Disallowed Without Explicit User Approval

- CSS Modules (for example: `*.module.css`, `*.module.scss`).
- CSS-in-JS styling approaches (for example: styled-components, emotion, style objects for regular UI styling).
- New standalone stylesheet files for component-local styling when Tailwind utilities can express the same result.
- Inline style objects for standard visual styling (allowed only for dynamic values Tailwind cannot represent).

## Exception Handling

- If a third-party library requires specific CSS hooks, keep overrides minimal and isolated.
- If Tailwind cannot express a required dynamic value, use the smallest possible inline style and document why in code.
- Any non-Tailwind styling exception must be explicitly requested or approved by the user.
