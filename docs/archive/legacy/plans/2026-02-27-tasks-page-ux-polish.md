# Tasks Page UX Polish — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Match the Tasks page's visual design to the updated Chat page glass morphism style for consistent app-wide UX.

**Architecture:** Surgical in-place modifications to 4 existing Vue components. No new files. All changes are CSS/template level plus one new ref for the collapsible filter.

**Tech Stack:** Vue 3, Tailwind CSS, existing `makoclaw-*` design tokens, `glass-panel`/`glass-sticky` utility classes.

**Design Doc:** `docs/plans/2026-02-27-tasks-page-ux-polish-design.md`

---

### Task 1: Glass Morphism on NewTaskModal

**Files:**
- Modify: `pkg/web/frontend/src/components/Tasks/NewTaskModal.vue`

**Step 1: Upgrade modal backdrop and card**

Find the overlay (line 3):
```
class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center p-4 z-modal"
```
Replace with:
```
class="fixed inset-0 bg-black/40 backdrop-blur-md flex items-center justify-center p-3 sm:p-4 z-modal"
```

Find the modal card (line 4):
```
class="bg-makoclaw-surface border border-makoclaw-border rounded-lg max-w-md w-full shadow-lg"
```
Replace with:
```
class="glass-panel rounded-2xl max-w-md w-full shadow-2xl"
```

**Step 2: Upgrade close button touch target**

Find close button (around line 8-11):
```
class="p-1 hover:bg-makoclaw-border rounded transition-smooth"
```
Replace with:
```
class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center hover:bg-makoclaw-accent/10 rounded-xl text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all"
```

**Step 3: Upgrade form container and header padding**

Find header (line 6):
```
class="flex items-center justify-between p-4 border-b border-makoclaw-border"
```
Replace with:
```
class="flex items-center justify-between p-3 sm:p-4 border-b border-makoclaw-border/30"
```

Find form container (line 19):
```
class="p-4 space-y-4"
```
Replace with:
```
class="p-3 sm:p-4 md:p-5 space-y-3 sm:space-y-4"
```

**Step 4: Upgrade form inputs**

Find all input classes (3 occurrences — title input, description textarea, status select):
```
class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded focus-ring text-sm"
```
Replace all with:
```
class="w-full px-2.5 sm:px-3 md:px-4 py-1.5 sm:py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent text-sm transition-all backdrop-blur-sm"
```

Note: the description textarea also has `resize-none` — keep that appended.

**Step 5: Upgrade action buttons**

Find actions footer (line 76):
```
class="flex gap-3 pt-4 border-t border-makoclaw-border"
```
Replace with:
```
class="flex gap-2 sm:gap-3 pt-3 sm:pt-4 border-t border-makoclaw-border/30"
```

Find cancel button:
```
class="flex-1 px-3 py-2 border border-makoclaw-border rounded hover:bg-makoclaw-border transition-smooth"
```
Replace with:
```
class="flex-1 px-3 py-2 min-h-[36px] border border-makoclaw-border/50 rounded-xl hover:bg-makoclaw-bg transition-all text-sm"
```

Find create button:
```
class="flex-1 px-3 py-2 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded transition-smooth disabled:opacity-50"
```
Replace with:
```
class="flex-1 px-3 py-2 min-h-[36px] bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl transition-all shadow-md shadow-makoclaw-accent/20 hover:shadow-makoclaw-accent/40 text-sm font-medium disabled:opacity-50"
```

**Step 6: Commit**

```bash
git add pkg/web/frontend/src/components/Tasks/NewTaskModal.vue
git commit -m "style(tasks): glass morphism + responsive polish on NewTaskModal"
```

---

### Task 2: Glass Morphism on TaskDetailsModal

**Files:**
- Modify: `pkg/web/frontend/src/components/Tasks/TaskDetailsModal.vue`

**Step 1: Upgrade modal backdrop and card**

Find overlay (line 3):
```
class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center p-4 z-modal overflow-y-auto"
```
Replace with:
```
class="fixed inset-0 bg-black/40 backdrop-blur-md flex items-center justify-center p-3 sm:p-4 z-modal overflow-y-auto"
```

Find modal card (line 4):
```
class="bg-makoclaw-surface border border-makoclaw-border rounded-lg max-w-4xl w-full shadow-lg my-4"
```
Replace with:
```
class="glass-panel rounded-2xl max-w-2xl md:max-w-3xl lg:max-w-4xl w-full shadow-2xl my-4"
```

Changes: glass-panel, rounded-2xl, responsive max-width (narrower on small screens).

**Step 2: Upgrade header and close button**

Find header (line 6):
```
class="flex items-center justify-between p-4 border-b border-makoclaw-border"
```
Replace with:
```
class="flex items-center justify-between p-3 sm:p-4 border-b border-makoclaw-border/30"
```

Find close button:
```
class="p-1 hover:bg-makoclaw-border rounded transition-smooth"
```
Replace with:
```
class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center hover:bg-makoclaw-accent/10 rounded-xl text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all"
```

**Step 3: Upgrade content container**

Find (line 28):
```
class="p-4 space-y-4 max-h-[70vh] overflow-y-auto"
```
Replace with:
```
class="p-3 sm:p-4 md:p-6 space-y-3 sm:space-y-4 max-h-[60vh] sm:max-h-[70vh] overflow-y-auto custom-scrollbar"
```

**Step 4: Upgrade form inputs (title, description, status selects)**

Find all 3 occurrences:
```
class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded focus-ring text-sm"
```
Replace all with:
```
class="w-full px-2.5 sm:px-3 md:px-4 py-1.5 sm:py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent text-sm transition-all backdrop-blur-sm"
```

Note: description textarea also has `resize-y` — keep that appended.

**Step 5: Upgrade result textarea**

Find (line 83-89):
```
class="w-full px-4 py-3 bg-makoclaw-bg border border-makoclaw-border rounded-xl focus:ring-2 focus:ring-makoclaw-accent/50 focus:border-makoclaw-accent text-base font-mono resize-y leading-relaxed"
```
Replace with:
```
class="w-full px-3 sm:px-4 py-2 sm:py-3 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent text-sm sm:text-base font-mono resize-y leading-relaxed backdrop-blur-sm transition-all"
```

**Step 6: Upgrade logs container**

Find (line 96):
```
class="bg-makoclaw-bg border border-makoclaw-border rounded p-3 text-xs max-h-48 overflow-y-auto"
```
Replace with:
```
class="bg-makoclaw-bg/50 border border-makoclaw-border/30 rounded-xl p-2 sm:p-3 text-xs max-h-36 sm:max-h-48 overflow-y-auto backdrop-blur-sm custom-scrollbar"
```

**Step 7: Upgrade action buttons**

Find actions footer (line 119):
```
class="border-t border-makoclaw-border p-4 flex gap-2 justify-end"
```
Replace with:
```
class="border-t border-makoclaw-border/30 p-3 sm:p-4 flex gap-1.5 sm:gap-2 justify-end flex-wrap"
```

For ALL 5 action buttons (archive, unarchive, delete, cancel, save), add `min-h-[36px] rounded-xl` and change `rounded` to `rounded-xl`. Specifically:

Archive button — find:
```
class="px-4 py-2 bg-yellow-500/20 text-yellow-500 hover:bg-yellow-500/30 border border-yellow-500/50 rounded transition-smooth text-sm font-medium disabled:opacity-50"
```
Replace with:
```
class="px-3 sm:px-4 py-2 min-h-[36px] bg-yellow-500/20 text-yellow-500 hover:bg-yellow-500/30 border border-yellow-500/50 rounded-xl transition-all text-sm font-medium disabled:opacity-50"
```

Unarchive button — find:
```
class="px-4 py-2 bg-blue-500/20 text-blue-500 hover:bg-blue-500/30 border border-blue-500/50 rounded transition-smooth text-sm font-medium disabled:opacity-50"
```
Replace with:
```
class="px-3 sm:px-4 py-2 min-h-[36px] bg-blue-500/20 text-blue-500 hover:bg-blue-500/30 border border-blue-500/50 rounded-xl transition-all text-sm font-medium disabled:opacity-50"
```

Delete button — find:
```
class="px-4 py-2 bg-makoclaw-error hover:bg-makoclaw-error/80 text-white rounded transition-smooth text-sm font-medium disabled:opacity-50"
```
Replace with:
```
class="px-3 sm:px-4 py-2 min-h-[36px] bg-makoclaw-error hover:bg-makoclaw-error/80 text-white rounded-xl transition-all text-sm font-medium disabled:opacity-50"
```

Cancel button — find:
```
class="px-4 py-2 border border-makoclaw-border rounded hover:bg-makoclaw-border transition-smooth text-sm"
```
Replace with:
```
class="px-3 sm:px-4 py-2 min-h-[36px] border border-makoclaw-border/50 rounded-xl hover:bg-makoclaw-bg transition-all text-sm"
```

Save button — find:
```
class="px-4 py-2 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded transition-smooth text-sm font-medium disabled:opacity-50"
```
Replace with:
```
class="px-3 sm:px-4 py-2 min-h-[36px] bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl transition-all shadow-md shadow-makoclaw-accent/20 hover:shadow-makoclaw-accent/40 text-sm font-medium disabled:opacity-50"
```

**Step 8: Commit**

```bash
git add pkg/web/frontend/src/components/Tasks/TaskDetailsModal.vue
git commit -m "style(tasks): glass morphism + responsive polish on TaskDetailsModal"
```

---

### Task 3: Task Card Glass Enhancement

**Files:**
- Modify: `pkg/web/frontend/src/components/Tasks/KanbanColumn.vue`

**Step 1: Upgrade column container**

Find (line 2):
```
class="glass-panel rounded-2xl p-4 flex flex-col h-full min-w-[288px] sm:min-w-[320px] shadow-sm"
```
Replace with:
```
class="glass-panel rounded-2xl p-2.5 sm:p-3 md:p-4 flex flex-col h-full min-w-[256px] sm:min-w-[288px] md:min-w-[320px] shadow-sm"
```

**Step 2: Upgrade task card glass effect**

Find (line 28):
```
class="bg-makoclaw-surface/50 border border-makoclaw-border/50 rounded-xl p-4 cursor-grab active:cursor-grabbing hover:border-makoclaw-accent/40 hover:shadow-xl hover:-translate-y-[2px] transition-all duration-300 group relative overflow-hidden backdrop-blur-sm"
```
Replace with:
```
class="bg-makoclaw-surface/30 border border-makoclaw-border/20 rounded-xl p-3 sm:p-3.5 md:p-4 cursor-grab active:cursor-grabbing hover:border-makoclaw-accent/30 hover:shadow-xl hover:-translate-y-[2px] transition-all duration-200 group relative overflow-hidden backdrop-blur-md ring-1 ring-white/[0.03] hover:ring-white/[0.08]"
```

Changes:
- `/50` → `/30` and `/20` (more transparent, matches Chat bubbles)
- `backdrop-blur-sm` → `backdrop-blur-md`
- Added `ring-1 ring-white/[0.03] hover:ring-white/[0.08]`
- Responsive padding
- `duration-300` → `duration-200`

**Step 3: Upgrade task list spacing**

Find (line 17-21):
```
class="flex-1 space-y-3 overflow-y-auto px-1 -mx-1"
```
Replace with:
```
class="flex-1 space-y-2 sm:space-y-3 overflow-y-auto px-0.5 sm:px-1 -mx-0.5 sm:-mx-1 custom-scrollbar"
```

**Step 4: Upgrade empty state**

Find (line 54):
```
class="flex flex-col items-center justify-center py-8 text-makoclaw-text-secondary/50 border-2 border-dashed border-makoclaw-border/30 rounded-lg"
```
Replace with:
```
class="flex flex-col items-center justify-center py-6 sm:py-8 text-makoclaw-text-secondary/50 border-2 border-dashed border-makoclaw-border/20 rounded-xl"
```

**Step 5: Commit**

```bash
git add pkg/web/frontend/src/components/Tasks/KanbanColumn.vue
git commit -m "style(tasks): glass-enhanced task cards with responsive padding"
```

---

### Task 4: Collapsible Filter Bar + Header Polish

**Files:**
- Modify: `pkg/web/frontend/src/views/TasksView.vue`

This task adds a `showFilters` ref and restructures the header for mobile collapsibility.

**Step 1: Add `showFilters` ref in script**

In the script section, after line 170 (`const exportDropdownRef = ref(null)`), add:
```javascript
const showFilters = ref(false)
```

**Step 2: Update header padding**

Find (line 7):
```
class="glass-sticky top-0 z-20 p-4 border-b border-makoclaw-border/30"
```
Replace with:
```
class="glass-sticky top-0 z-20 p-2 sm:p-3 md:p-4 border-b border-makoclaw-border/20"
```

**Step 3: Restructure header for collapsible filters**

The current header has one div with all controls. We need to split it into:
1. Always-visible row: search + filter toggle (mobile) + create button
2. Collapsible row: sort, status filter, archived checkbox (hidden on mobile by default)

Find the filter controls wrapper. The current structure (around lines 8-67) is a single `<div class="flex flex-col sm:flex-row gap-3 md:gap-4">` containing everything.

Replace the entire filter controls section (from the opening `<div class="flex flex-col sm:flex-row gap-3 md:gap-4">` to its closing `</div>` — the one that wraps search, sort, filter, archived, and new-task button) with:

```html
      <div class="flex flex-col gap-2 sm:gap-3">
        <!-- Always visible row: Search + filter toggle + create -->
        <div class="flex items-center gap-2 sm:gap-3">
          <!-- Search -->
          <div class="flex-1 relative group">
            <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-makoclaw-text-secondary group-focus-within:text-makoclaw-accent transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search tasks..."
              class="w-full pl-9 sm:pl-10 pr-3 sm:pr-4 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent transition-all text-sm backdrop-blur-sm min-h-[36px]"
            >
          </div>

          <!-- Filter toggle (mobile only) -->
          <button
            @click="showFilters = !showFilters"
            class="md:hidden p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-xl border border-makoclaw-border/50 hover:bg-makoclaw-accent/10 hover:text-makoclaw-accent transition-all"
            :class="showFilters ? 'bg-makoclaw-accent/10 text-makoclaw-accent border-makoclaw-accent/30' : 'text-makoclaw-text-secondary'"
            title="Toggle filters"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
            </svg>
          </button>

          <!-- Create task button -->
          <button
            @click="showNewTaskModal = true"
            class="px-3 sm:px-5 py-2 min-h-[36px] bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl transition-all shadow-lg shadow-makoclaw-accent/20 hover:shadow-makoclaw-accent/40 text-sm font-bold flex items-center gap-1.5 sm:gap-2 active:scale-95 flex-shrink-0"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
            <span class="hidden sm:inline">New Task</span>
          </button>
        </div>

        <!-- Collapsible filters row (hidden on mobile by default) -->
        <div :class="['flex items-center gap-2 sm:gap-3 overflow-x-auto scrollbar-hide transition-all duration-200', showFilters ? '' : 'hidden md:flex']">
          <select
            v-model="sortBy"
            class="px-3 sm:px-4 py-2 min-h-[36px] bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 text-sm hover:border-makoclaw-accent/30 transition-all cursor-pointer backdrop-blur-sm outline-none"
          >
            <option value="recent">Recent first</option>
            <option value="oldest">Oldest first</option>
            <option value="title">By title</option>
            <option value="status">By status</option>
          </select>

          <select
            v-model="statusFilter"
            class="px-3 sm:px-4 py-2 min-h-[36px] bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 text-sm hover:border-makoclaw-accent/30 transition-all cursor-pointer backdrop-blur-sm outline-none"
          >
            <option value="">All statuses</option>
            <option value="backlog">Backlog</option>
            <option value="todo">To Do</option>
            <option value="in_progress">In Progress</option>
            <option value="review">Review</option>
            <option value="done">Done</option>
          </select>

          <label class="flex items-center gap-2 px-3 py-1.5 sm:py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl backdrop-blur-sm cursor-pointer min-h-[36px] flex-shrink-0">
            <input
              type="checkbox"
              v-model="showArchived"
              class="rounded border-makoclaw-border bg-makoclaw-surface text-makoclaw-accent focus:ring-makoclaw-accent transition-all cursor-pointer"
            >
            <span class="text-xs sm:text-sm text-makoclaw-text-secondary whitespace-nowrap">Archived</span>
          </label>
        </div>
      </div>
```

**Step 4: Update kanban board padding**

Find (line 71):
```
class="flex-1 overflow-x-auto p-4"
```
Replace with:
```
class="flex-1 overflow-x-auto p-2 sm:p-3 md:p-4"
```

**Step 5: Update kanban column widths**

Find ALL 5 occurrences (lines 74, 85, 96, 107, 118):
```
class="flex-shrink-0 w-72 sm:w-80"
```
Replace ALL with:
```
class="flex-shrink-0 w-64 sm:w-72 md:w-80 lg:w-[340px]"
```

**Step 6: Commit**

```bash
git add pkg/web/frontend/src/views/TasksView.vue
git commit -m "style(tasks): collapsible filters, responsive kanban columns, spacing polish"
```

---

### Task 5: Build and Visual QA

**Files:** None (verification only)

**Step 1: Run production build**

```bash
cd pkg/web/frontend && npm run build
```

Expected: Build succeeds with no errors.

**Step 2: Visual QA checklist**

- [ ] 390px: Only search + create visible in header, filter toggle icon works
- [ ] 390px: Kanban columns 256px wide, smooth horizontal scroll
- [ ] 390px: Task cards have responsive padding (p-3)
- [ ] 390px: NewTaskModal glass effect, inputs scale properly
- [ ] 390px: TaskDetailsModal fits in viewport, scrollable
- [ ] 768px: All filters visible in one row
- [ ] 768px: Kanban columns 288px wide
- [ ] 1024px+: Columns 320px, modals properly centered
- [ ] Glass effects: both modals have frosted glass look
- [ ] Task cards: subtle ring glow, enhanced hover state
- [ ] Dark mode: shadows and glass effects correct
- [ ] Consistent feel with Chat page (same glass, same spacing scale)
