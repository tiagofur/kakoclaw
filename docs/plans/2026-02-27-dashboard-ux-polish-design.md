# Dashboard Page UX Polish — Design Document

**Date**: 2026-02-27
**Branch**: `ui/inner-app-ux-polish`
**Approach**: Surgical Polish (modify existing file in-place)
**Goal**: Match the Dashboard page design to the updated Chat & Tasks glass morphism style

## Design Decisions

- **Approach**: In-place modifications to single file `DashboardView.vue`
- **Aesthetic**: Glass morphism + gradients (matching Chat & Tasks pages)
- **Responsive spacing**: `p-2 sm:p-3 md:p-4` scale throughout
- **Touch targets**: `min-h-[36px] min-w-[36px]` on all interactive elements
- **Text scale**: Replace non-standard `text-[10px]`, `text-[11px]` with `text-xs`

---

## Section 1: Header Bar

- Padding: `p-6` → `p-2 sm:p-3 md:p-4`
- Title: `text-2xl` → `text-xl sm:text-2xl`
- Refresh button: add `min-h-[36px] min-w-[36px]`, use `border-makoclaw-border/50`

## Section 2: Welcome Banner

- Padding: `p-6 md:p-8` → `p-4 sm:p-5 md:p-6`
- Heading: `text-2xl md:text-3xl` → `text-xl sm:text-2xl`
- Decoration SVG: `w-44 h-44` → `w-32 h-32`
- CTA buttons: add `min-h-[36px]`, reduce padding slightly

## Section 3: Content Area Spacing

- Container: `p-4 md:p-6 space-y-5` → `p-2 sm:p-3 md:p-4 space-y-3 sm:space-y-4`
- Grid gaps: `gap-5` → `gap-3 sm:gap-4` throughout

## Section 4: Stats Grid

- Icon container: `p-2` with `w-4 h-4` → `p-2.5` with `w-5 h-5`
- Stat value: `text-3xl` → `text-2xl`
- Label: `text-[10px]` → `text-xs`
- Trend badge: `text-[10px]` → `text-xs`

## Section 5: Chart Panels

- Panel padding: `p-5` → `p-3 sm:p-4 md:p-5`
- Header margin: `mb-5` → `mb-3 sm:mb-4`
- Chart min-height: `min-h-[300px]` → `min-h-[200px] sm:min-h-[250px] md:min-h-[300px]`

## Section 6: Action Launchpad

- Icon container: `w-8 h-8` → `w-9 h-9`
- Icon: `w-4 h-4` → `w-5 h-5`
- Label: `text-[11px]` → `text-xs`
- Launchpad padding: `p-5` → `p-3 sm:p-4`

## Section 7: Activity Feed

- Feed container padding: `p-5` → `p-3 sm:p-4`
- Item icon container: `w-8 h-8` → `w-9 h-9`
- Item icon: `w-4 h-4` → `w-5 h-5`
- Title: `text-[11px]` → `text-xs`
- Meta text: `text-[10px]` → `text-xs`
- Empty state icon: `w-12 h-12` → `w-10 h-10`

## Section 8: Detailed Metrics & Loading Skeleton

- Metrics grid gap: `gap-5` → `gap-3 sm:gap-4`
- Metrics panel padding: `p-5` → `p-3 sm:p-4 md:p-5`
- Metric labels: `text-[10px]` → `text-xs`
- Skeleton: align border radii (`rounded-3xl` → `rounded-2xl`), adjust heights

## File to Modify

1. `pkg/web/frontend/src/views/DashboardView.vue` — All 8 sections
