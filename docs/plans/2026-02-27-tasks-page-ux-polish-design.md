# Tasks Page UX Polish — Design Document

**Date**: 2026-02-27
**Branch**: `ui/inner-app-ux-polish`
**Approach**: Surgical Polish (modify existing files in-place)
**Goal**: Match the Tasks page design to the updated Chat page glass morphism style

## Design Decisions

- **Mobile kanban**: Keep horizontal scroll with improved snap and column sizing
- **Controls bar**: Collapsible filters on mobile (search + create always visible)
- **Aesthetic**: Glass morphism + gradients (matching Chat page)
- **Approach**: In-place modifications, no new components

---

## Section 1: Glass Morphism on Modals

Both NewTaskModal and TaskDetailsModal get glass treatment:
- Card: `glass-panel rounded-2xl shadow-2xl`
- Backdrop: `bg-black/40 backdrop-blur-md` (upgrade from `backdrop-blur-sm`)

## Section 2: Responsive Padding Normalization

Apply `p-2 sm:p-3 md:p-4` scale:
- Header bar: `p-2 sm:p-3 md:p-4`
- Kanban columns: `p-2.5 sm:p-3 md:p-4`
- Task cards: `p-3 sm:p-3.5 md:p-4`
- Modal content: `p-3 sm:p-4 md:p-6`
- Form inputs: `px-2.5 sm:px-3 md:px-4 py-1.5 sm:py-2`

## Section 3: Collapsible Filter Bar on Mobile

- Mobile (< md): search + create visible, filters behind toggle
- Desktop (md+): all controls in one row
- New `showFilters` ref, toggle button `md:hidden`

## Section 4: Kanban Column Width Scaling

- Current: `w-72 sm:w-80`
- New: `w-64 sm:w-72 md:w-80 lg:w-[340px]`

## Section 5: Task Card Glass Enhancement

- `bg-makoclaw-surface/30 border-makoclaw-border/20 backdrop-blur-md`
- `ring-1 ring-white/[0.03] hover:ring-white/[0.08]`
- `hover:border-makoclaw-accent/30 hover:-translate-y-[2px]`

## Section 6: Touch Targets & Button Consistency

- All buttons: `min-h-[36px] min-w-[36px]`
- Consistent icon sizing with Chat page toolbar

## Files to Modify

1. `pkg/web/frontend/src/views/TasksView.vue` — Sections 2, 3, 6
2. `pkg/web/frontend/src/components/KanbanColumn.vue` — Sections 2, 4, 5
3. `pkg/web/frontend/src/components/NewTaskModal.vue` — Sections 1, 2, 6
4. `pkg/web/frontend/src/components/TaskDetailsModal.vue` — Sections 1, 2, 6
