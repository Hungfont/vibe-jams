# shadcn Primitive Audit

Date: 2026-04-06

## Scope

- Reviewed directory: `frontend/src/components/ui/**`
- Inventory source: `frontend/config/shadcn-primitive-inventory.json`

## Findings

| Primitive file | Inventory status | Notes |
| --- | --- | --- |
| `alert.tsx` | Approved | Keep as approved primitive.
| `avatar.tsx` | Approved | Keep as approved primitive.
| `badge.tsx` | Approved | Keep as approved primitive.
| `button.tsx` | Approved | High-risk primitive; migrated to normalized shadcn-style variant helper.
| `card.tsx` | Approved | High-risk primitive; migrated to shadcn-style section wrappers.
| `dialog.tsx` | Approved | Keep as approved primitive.
| `dropdown-menu.tsx` | Approved | Keep as approved primitive.
| `input.tsx` | Approved | Keep as approved primitive.
| `scroll-area.tsx` | Approved | Keep as approved primitive.
| `separator.tsx` | Approved | Keep as approved primitive.
| `skeleton.tsx` | Approved | Keep as approved primitive.
| `slider.tsx` | Approved | Keep as approved primitive.
| `tabs.tsx` | Approved | Keep as approved primitive.
| `toast.tsx` | Approved | Keep as approved primitive.
| `tooltip.tsx` | Approved | Keep as approved primitive.
| `use-toast.ts` | Approved + exception | Tracked helper exception metadata in `frontend/config/shadcn-exceptions.json`.

## Current Compliance

- No unknown primitive file names detected under `components/ui`.
- No duplicate approved primitive names detected outside `components/ui`.
- Feature components currently import from `@/components/ui/*` approved primitives only.

## Follow-up

- Keep this audit updated when primitives are added, removed, or exception status changes.