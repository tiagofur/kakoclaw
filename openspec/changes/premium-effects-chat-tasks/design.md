# Design: Premium Visual Effects for Chat & Tasks Pages

**Change:** `premium-effects-chat-tasks`
**Status:** Draft
**Date:** 2026-02-27

## Overview

The Dashboard page features a premium visual language — gradient glow blobs, animated hover underlines, glassmorphism overlays, icon hover transforms, staggered fade-in animations, and decorative background elements. The Chat and Tasks pages currently lack these effects and feel visually flat by comparison. This design brings strategic Dashboard-level polish to both pages without overwhelming them.

---

## Dashboard Premium Effects Catalog

The following effects exist in `DashboardView.vue` and are absent or underutilized in Chat/Tasks:

| # | Effect | Dashboard Implementation | Present in Chat? | Present in Tasks? |
|---|--------|--------------------------|-------------------|-------------------|
| 1 | **Decorative background blobs** | Animated `blur-[100px]` circles with `bg-makoclaw-accent/10`, `bg-blue-500/5`, floating keyframe animation | Partial (static radial gradient only) | Partial (static radial gradient only) |
| 2 | **Stat card gradient glow** | `absolute -top-12 -right-12 w-32 h-32 rounded-full opacity-20 blur-[30px]` behind icons | No | No |
| 3 | **Animated bottom-line on hover** | `absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r group-hover:w-full transition-all duration-500` | No | No |
| 4 | **Icon hover transform** | `group-hover:scale-110 group-hover:rotate-3` with `duration-500` on icon containers | No | No |
| 5 | **Trend badges** | Pill with SVG arrow, `bg-makoclaw-success/10 border-makoclaw-success/20 backdrop-blur-md` | No | No |
| 6 | **Staggered fade-in-up animation** | Custom `animate-fade-in-up` with `animation-delay` per element | No | No |
| 7 | **Glass panel hover elevation** | `hover:shadow-[0_8px_30px_rgb(0,0,0,0.06)] hover:-translate-y-1` on cards | Partial (MessageBubble has `hover:scale-[1.002]`) | Partial (task cards have `hover:-translate-y-[2px]`) |
| 8 | **Section header with pulse dot** | `w-2 h-2 rounded-full bg-blue-500 animate-pulse` beside section titles | No | No |
| 9 | **Soft background gradient on panels** | `bg-gradient-to-br from-indigo-500/5 to-purple-500/5 opacity-50` as overlay | No | No |
| 10 | **Activity item hover glow line** | `absolute left-0 w-1 h-0 bg-gradient-to-b group-hover:h-2/3` on list items | No | No |
| 11 | **`drop-shadow-sm` on icons and numbers** | Applied to stat numbers and icon components | No | No |
| 12 | **Hover text color transition to accent** | `group-hover:text-makoclaw-accent transition-colors` on titles | Yes (task cards) | Yes (task cards) |

---

## Enhancements

### Enhancement 1: Animated Background Blob for Chat Empty State

**Target:** `pkg/web/frontend/src/views/ChatView.vue` (template, around line 2-4)
**Dashboard Reference:** Lines 4-5 — two floating `blur-[100px]` blobs with a `float` keyframe animation creating ambient motion
**Current:**
```html
<div class="absolute inset-0 pointer-events-none opacity-20 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-makoclaw-accent/30 via-transparent to-transparent"></div>
```
**New:**
```html
<!-- Background Gradient Mesh (Subtle) -->
<div class="absolute inset-0 pointer-events-none opacity-20 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-makoclaw-accent/30 via-transparent to-transparent"></div>
<!-- Decorative floating blob -->
<div class="absolute -top-24 -right-24 w-80 h-80 bg-makoclaw-accent/8 blur-[100px] rounded-full pointer-events-none" style="animation: float 20s infinite ease-in-out"></div>
```
**Rationale:** The static radial gradient is barely visible. Adding a single animated blob gives the Chat page ambient life matching the Dashboard. Only one blob to keep the chat content area visually clean.

---

### Enhancement 2: Premium Chat Empty State with Gradient Glow

**Target:** `pkg/web/frontend/src/views/ChatView.vue` (template, lines 151-159)
**Dashboard Reference:** Welcome banner (lines 46-66) — `bg-gradient-to-br from-makoclaw-accent via-blue-600 to-indigo-700`, radial gradient overlay, animated background mesh
**Current:**
```html
<div v-if="messages.length === 0" class="flex flex-col items-center justify-center h-full text-makoclaw-text-secondary animate-fadeIn">
  <div class="glass-panel p-6 sm:p-8 rounded-2xl mb-4 shadow-lg shadow-makoclaw-accent/5">
    <svg class="w-10 h-10 sm:w-12 sm:h-12 text-makoclaw-accent animate-pulse" style="animation-duration: 3s;" ...>
    </svg>
  </div>
  <p class="text-base sm:text-lg font-semibold text-makoclaw-text mt-2">Start a conversation</p>
  <p class="text-xs sm:text-sm text-makoclaw-text-secondary/70 mt-1">Ask anything or run a task</p>
</div>
```
**New:**
```html
<div v-if="messages.length === 0" class="flex flex-col items-center justify-center h-full text-makoclaw-text-secondary animate-fadeIn">
  <div class="relative glass-panel p-6 sm:p-8 rounded-2xl mb-4 shadow-lg shadow-makoclaw-accent/5 group">
    <!-- Gradient glow behind icon -->
    <div class="absolute -top-8 -right-8 w-24 h-24 bg-gradient-to-br from-makoclaw-accent to-blue-500 rounded-full opacity-15 blur-[25px] group-hover:opacity-30 group-hover:scale-110 transition-all duration-500"></div>
    <svg class="w-10 h-10 sm:w-12 sm:h-12 text-makoclaw-accent drop-shadow-sm relative z-10 transition-transform duration-500 group-hover:scale-110 group-hover:rotate-3" style="animation: pulse 3s ease-in-out infinite;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
    </svg>
    <!-- Animated bottom-line -->
    <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover:w-full transition-all duration-500 opacity-70 rounded-b-2xl"></div>
  </div>
  <p class="text-base sm:text-lg font-semibold text-makoclaw-text mt-2">Start a conversation</p>
  <p class="text-xs sm:text-sm text-makoclaw-text-secondary/70 mt-1">Ask anything or run a task</p>
</div>
```
**Rationale:** The empty state is seen every time a user starts a new chat. Adding the Dashboard's gradient glow blob, icon hover transform (scale + rotate), and animated bottom-line makes it feel premium and inviting. The `group` hover wrapper ensures all effects activate together.

---

### Enhancement 3: Chat Input Area Accent Line

**Target:** `pkg/web/frontend/src/views/ChatView.vue` (template, line 192)
**Dashboard Reference:** Lines 77, 144 — animated bottom-line `h-[2px] w-0 group-hover:w-full bg-gradient-to-r from-makoclaw-accent to-blue-500`
**Current:**
```html
<div class="border-t border-makoclaw-border/30 bg-makoclaw-surface/60 backdrop-blur-xl p-2 sm:p-3 md:p-4 z-20 relative ring-1 ring-white/[0.05]">
```
**New:**
```html
<div class="border-t border-makoclaw-border/30 bg-makoclaw-surface/60 backdrop-blur-xl p-2 sm:p-3 md:p-4 z-20 relative ring-1 ring-white/[0.05] group/input">
  <!-- Gradient accent line at the top of input area -->
  <div class="absolute top-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent via-blue-500 to-indigo-500 group-focus-within/input:w-full transition-all duration-700 opacity-60"></div>
```
**Rationale:** When the user focuses on the textarea, a gradient accent line smoothly expands across the top of the input area, providing a premium focus indicator. This mirrors the Dashboard's animated bottom-line pattern but triggers on focus-within rather than hover, which is more appropriate for an input context.

---

### Enhancement 4: Assistant Message Bubble Subtle Glow

**Target:** `pkg/web/frontend/src/components/MessageBubble.vue` (template, lines 8-14)
**Dashboard Reference:** Lines 74-76 — soft gradient glow blob behind stat card icons with hover intensification
**Current (assistant bubble):**
```html
<div
  :class="[
    'max-w-[92%] sm:max-w-[88%] md:max-w-[85%] lg:max-w-3xl xl:max-w-4xl 2xl:max-w-5xl px-3 sm:px-4 md:px-5 py-2 sm:py-2.5 md:py-3 shadow-lg transition-all duration-300 transform hover:scale-[1.002] animate-slideUp',
    msg.role === 'user'
      ? 'bg-gradient-to-br from-makoclaw-accent to-makoclaw-accent-hover text-white rounded-2xl rounded-br-none shadow-makoclaw-accent/10'
      : 'glass-panel text-makoclaw-text rounded-2xl rounded-bl-none shadow-lg shadow-black/[0.03] dark:shadow-black/20'
  ]"
>
```
**New:**
```html
<div
  :class="[
    'relative max-w-[92%] sm:max-w-[88%] md:max-w-[85%] lg:max-w-3xl xl:max-w-4xl 2xl:max-w-5xl px-3 sm:px-4 md:px-5 py-2 sm:py-2.5 md:py-3 shadow-lg transition-all duration-300 transform hover:scale-[1.002] animate-slideUp overflow-hidden group/bubble',
    msg.role === 'user'
      ? 'bg-gradient-to-br from-makoclaw-accent to-makoclaw-accent-hover text-white rounded-2xl rounded-br-none shadow-makoclaw-accent/10'
      : 'glass-panel text-makoclaw-text rounded-2xl rounded-bl-none shadow-lg shadow-black/[0.03] dark:shadow-black/20'
  ]"
>
  <!-- Hover accent line for assistant messages -->
  <div v-if="msg.role === 'assistant'" class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover/bubble:w-full transition-all duration-500 opacity-50"></div>
```
**Rationale:** Adding the animated bottom-line on assistant message bubbles creates a subtle premium hover indicator that mirrors the Dashboard's card treatment. The `relative` and `overflow-hidden` additions contain the line. User bubbles already have a strong gradient background and don't need this effect.

---

### Enhancement 5: Specialist Panel Card Premium Treatment

**Target:** `pkg/web/frontend/src/components/Chat/SpecialistsPanel.vue` (template, lines 29-33)
**Dashboard Reference:** Lines 156-168 — launchpad action cards with gradient icon containers, hover glow overlays, `group-hover:scale-110` on icons, internal opacity overlay
**Current:**
```html
<div
  v-for="specialist in specialists"
  :key="specialist.name"
  class="flex items-start gap-3 p-2 sm:p-2.5 rounded-xl bg-makoclaw-surface/20 hover:bg-makoclaw-surface/40 border border-makoclaw-border/10 hover:border-makoclaw-border/20 backdrop-blur-sm transition-all duration-200"
>
```
**New:**
```html
<div
  v-for="specialist in specialists"
  :key="specialist.name"
  class="flex items-start gap-3 p-2 sm:p-2.5 rounded-xl bg-makoclaw-surface/20 hover:bg-makoclaw-surface/40 border border-makoclaw-border/10 hover:border-makoclaw-border/20 backdrop-blur-sm transition-all duration-300 group/spec relative overflow-hidden"
>
  <!-- Hover glow line (left edge) -->
  <div class="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-0 bg-gradient-to-b from-transparent via-makoclaw-accent/40 to-transparent group-hover/spec:h-2/3 transition-all duration-300"></div>
```
**Rationale:** Mirrors the Dashboard's activity feed items (lines 190-192) which use a left-edge glow line on hover. This gives specialist cards a premium interactive feel when browsing them, without disrupting the compact layout.

---

### Enhancement 6: Kanban Column Header Status Glow Dot

**Target:** `pkg/web/frontend/src/components/Tasks/KanbanColumn.vue` (template, lines 4-14)
**Dashboard Reference:** Lines 111-112, 125-126 — `w-2 h-2 rounded-full bg-blue-500 animate-pulse` beside section headers
**Current:**
```html
<div class="mb-4 pb-0 flex flex-col">
  <div class="pb-3 flex items-center justify-between">
    <h3 class="font-bold text-xs uppercase tracking-[0.2em] text-makoclaw-text-secondary flex items-center gap-2 opacity-80">
      {{ title }}
    </h3>
    <span class="text-[10px] bg-makoclaw-bg/50 font-bold text-makoclaw-accent px-2.5 py-1 rounded-full border border-makoclaw-accent/10 shadow-sm">
        {{ tasks.length }}
    </span>
  </div>
  <div class="h-[1px] bg-gradient-to-r from-transparent via-makoclaw-accent/10 to-transparent"></div>
</div>
```
**New:**
```html
<div class="mb-4 pb-0 flex flex-col">
  <div class="pb-3 flex items-center justify-between">
    <h3 class="font-bold text-xs uppercase tracking-[0.2em] text-makoclaw-text-secondary flex items-center gap-2 opacity-80">
      <span class="w-2 h-2 rounded-full" :class="statusDotColor"></span>
      {{ title }}
    </h3>
    <span class="text-[10px] bg-makoclaw-bg/50 font-bold text-makoclaw-accent px-2.5 py-1 rounded-full border border-makoclaw-accent/10 shadow-sm backdrop-blur-md">
        {{ tasks.length }}
    </span>
  </div>
  <div class="h-[1px] bg-gradient-to-r from-transparent via-makoclaw-accent/10 to-transparent"></div>
</div>
```
**New computed (script):**
```js
const statusDotColor = computed(() => {
  const colors = {
    'backlog': 'bg-makoclaw-text-secondary/40',
    'todo': 'bg-makoclaw-warning',
    'in_progress': 'bg-makoclaw-accent animate-pulse',
    'review': 'bg-amber-500 animate-pulse',
    'done': 'bg-makoclaw-success'
  }
  return colors[props.status] || 'bg-makoclaw-text-secondary/40'
})
```
**Rationale:** The Dashboard uses colored pulse dots beside section headers (Model Intelligence, Operations Status). Adding a status-colored dot to each Kanban column header instantly communicates the column's purpose and draws the eye. Only `in_progress` and `review` pulse, matching the Dashboard's pattern of animating active states.

---

### Enhancement 7: Task Card Animated Bottom-Line & Gradient Glow

**Target:** `pkg/web/frontend/src/components/Tasks/KanbanColumn.vue` (template, lines 22-30)
**Dashboard Reference:** Lines 70-77 — stat cards with `absolute bottom-0 left-0 h-[2px] w-0 group-hover:w-full` animated bottom-line, plus `absolute -top-12 -right-12 w-32 h-32 opacity-20 blur-[30px]` gradient glow
**Current:**
```html
<div
  v-for="task in tasks"
  :key="task.id"
  draggable="true"
  @dragstart="dragStart($event, task)"
  @click="$emit('task-click', task)"
  class="bg-makoclaw-surface/30 border border-makoclaw-border/20 rounded-xl p-3 sm:p-3.5 md:p-4 cursor-grab active:cursor-grabbing hover:border-makoclaw-accent/30 hover:shadow-xl hover:-translate-y-[2px] transition-all duration-200 group relative overflow-hidden backdrop-blur-md ring-1 ring-white/[0.03] hover:ring-white/[0.08]"
>
  <div class="absolute inset-x-0 top-0 h-1 bg-gradient-to-r from-makoclaw-accent to-blue-500 opacity-0 group-hover:opacity-100 transition-opacity"></div>
```
**New:**
```html
<div
  v-for="task in tasks"
  :key="task.id"
  draggable="true"
  @dragstart="dragStart($event, task)"
  @click="$emit('task-click', task)"
  class="bg-makoclaw-surface/30 border border-makoclaw-border/20 rounded-xl p-3 sm:p-3.5 md:p-4 cursor-grab active:cursor-grabbing hover:border-makoclaw-accent/30 hover:shadow-xl hover:-translate-y-[2px] transition-all duration-300 group relative overflow-hidden backdrop-blur-md ring-1 ring-white/[0.03] hover:ring-white/[0.08]"
>
  <!-- Top accent line (existing, unchanged) -->
  <div class="absolute inset-x-0 top-0 h-1 bg-gradient-to-r from-makoclaw-accent to-blue-500 opacity-0 group-hover:opacity-100 transition-opacity"></div>
  <!-- Animated bottom-line on hover -->
  <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover:w-full transition-all duration-500 opacity-60"></div>
  <!-- Soft gradient glow (top-right corner) -->
  <div class="absolute -top-8 -right-8 w-20 h-20 bg-gradient-to-br from-makoclaw-accent to-blue-500 rounded-full opacity-0 blur-[20px] group-hover:opacity-15 transition-all duration-500"></div>
```
**Rationale:** Task cards currently have a top accent line but lack the bottom-line and gradient glow effects that make Dashboard stat cards feel premium. Adding both effects (at smaller scale than Dashboard since task cards are smaller) creates visual consistency. The `duration-200` is bumped to `duration-300` to smooth out the multi-effect hover.

---

### Enhancement 8: Tasks Page Background Floating Blob

**Target:** `pkg/web/frontend/src/views/TasksView.vue` (template, lines 2-4)
**Dashboard Reference:** Lines 4-5 — animated floating blobs with `blur-[100px]` and `float` keyframe animation
**Current:**
```html
<div class="flex flex-col h-full bg-makoclaw-bg relative overflow-hidden">
  <!-- Background Gradient Mesh (Subtle) -->
  <div class="absolute inset-0 pointer-events-none opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-blue-500/20 via-transparent to-transparent"></div>
```
**New:**
```html
<div class="flex flex-col h-full bg-makoclaw-bg relative overflow-hidden">
  <!-- Background Gradient Mesh (Subtle) -->
  <div class="absolute inset-0 pointer-events-none opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-blue-500/20 via-transparent to-transparent"></div>
  <!-- Decorative floating blob -->
  <div class="absolute -bottom-24 -right-24 w-72 h-72 bg-indigo-500/8 blur-[80px] rounded-full pointer-events-none" style="animation: float 18s infinite ease-in-out"></div>
```
**Also add `<style scoped>` section** (TasksView.vue currently has none):
```css
<style scoped>
@keyframes float {
  0%, 100% { transform: translateY(0) translateX(0); }
  50% { transform: translateY(-15px) translateX(8px); }
}
</style>
```
**Rationale:** Matching the Dashboard's floating blob treatment. Positioned at bottom-right to complement the existing top-left radial gradient. Slightly smaller (w-72 vs w-96) and less blur (80px vs 100px) to keep the Kanban board readable.

---

### Enhancement 9: Chat Page Float Keyframe (Required for Enhancement 1)

**Target:** `pkg/web/frontend/src/views/ChatView.vue` (needs `<style scoped>` addition)
**Dashboard Reference:** Lines 422-425 — `@keyframes float` animation
**Current:** ChatView.vue has no `<style scoped>` section.
**New:** Add after `</script>`:
```css
<style scoped>
@keyframes float {
  0%, 100% { transform: translateY(0) translateX(0); }
  50% { transform: translateY(-20px) translateX(10px); }
}
</style>
```
**Rationale:** Required to support the floating blob added in Enhancement 1. Identical to the Dashboard's keyframe.

---

### Enhancement 10: New Task Modal Header Accent Line

**Target:** `pkg/web/frontend/src/components/Tasks/NewTaskModal.vue` (template, lines 4-16)
**Dashboard Reference:** Lines 46-66 — welcome banner with gradient background and decorative elements
**Current:**
```html
<div class="glass-panel rounded-2xl max-w-md w-full shadow-2xl">
  <!-- Header -->
  <div class="flex items-center justify-between p-3 sm:p-4 border-b border-makoclaw-border/30">
    <h3 class="text-lg font-semibold">Create New Task</h3>
```
**New:**
```html
<div class="glass-panel rounded-2xl max-w-md w-full shadow-2xl relative overflow-hidden">
  <!-- Subtle gradient glow at top -->
  <div class="absolute -top-12 left-1/2 -translate-x-1/2 w-48 h-24 bg-gradient-to-br from-makoclaw-accent to-indigo-500 rounded-full opacity-10 blur-[30px] pointer-events-none"></div>
  <!-- Header -->
  <div class="flex items-center justify-between p-3 sm:p-4 border-b border-makoclaw-border/30 relative z-10">
    <h3 class="text-lg font-semibold">Create New Task</h3>
```
**Rationale:** Adds a soft centered gradient glow at the top of the modal, mirroring the Dashboard's soft glow elements. Subtle enough to not distract from the form content while making the modal feel more premium.

---

### Enhancement 11: Task Details Modal Header Glow

**Target:** `pkg/web/frontend/src/components/Tasks/TaskDetailsModal.vue` (template, lines 3-5)
**Dashboard Reference:** Same soft glow pattern as stat cards (lines 74-76)
**Current:**
```html
<div class="glass-panel rounded-2xl max-w-2xl md:max-w-3xl lg:max-w-4xl w-full shadow-2xl my-4">
  <!-- Header -->
  <div class="flex items-center justify-between p-3 sm:p-4 border-b border-makoclaw-border/30">
```
**New:**
```html
<div class="glass-panel rounded-2xl max-w-2xl md:max-w-3xl lg:max-w-4xl w-full shadow-2xl my-4 relative overflow-hidden">
  <!-- Subtle gradient glow at top -->
  <div class="absolute -top-12 left-1/2 -translate-x-1/2 w-64 h-24 bg-gradient-to-br from-makoclaw-accent to-indigo-500 rounded-full opacity-10 blur-[30px] pointer-events-none"></div>
  <!-- Header -->
  <div class="flex items-center justify-between p-3 sm:p-4 border-b border-makoclaw-border/30 relative z-10">
```
**Rationale:** Same treatment as Enhancement 10 but slightly wider (w-64 vs w-48) for the larger modal. Consistent premium glow across both task modals.

---

### Enhancement 12: Kanban Column Panel Hover Elevation

**Target:** `pkg/web/frontend/src/components/Tasks/KanbanColumn.vue` (template, line 2)
**Dashboard Reference:** Lines 71, 105, 119, 135 — `hover:shadow-[0_8px_30px_rgb(0,0,0,0.06)] hover:-translate-y-1` on glass panels
**Current:**
```html
<div class="glass-panel rounded-2xl p-2.5 sm:p-3 md:p-4 flex flex-col h-full min-w-[256px] sm:min-w-[288px] md:min-w-[320px] shadow-sm">
```
**New:**
```html
<div class="glass-panel rounded-2xl p-2.5 sm:p-3 md:p-4 flex flex-col h-full min-w-[256px] sm:min-w-[288px] md:min-w-[320px] shadow-sm transition-all duration-500 hover:shadow-[0_8px_30px_rgb(0,0,0,0.04)]">
```
**Rationale:** Adds the Dashboard's hover shadow elevation to Kanban columns. Uses a softer shadow value (`0.04` vs `0.06`) since columns are large elements and aggressive shadows would be distracting. No translate-y since columns are side-by-side in a scroll container and vertical movement would be jarring.

---

### Enhancement 13: Create Task Button Premium Treatment

**Target:** `pkg/web/frontend/src/views/TasksView.vue` (template, lines 37-43)
**Dashboard Reference:** Line 62 — CTA button with `shadow-xl hover:shadow-white/20 hover:-translate-y-0.5 active:scale-95`
**Current:**
```html
<button
  @click="showNewTaskModal = true"
  class="px-3 sm:px-5 py-2 min-h-[36px] bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl transition-all shadow-lg shadow-makoclaw-accent/20 hover:shadow-makoclaw-accent/40 text-sm font-bold flex items-center gap-1.5 sm:gap-2 active:scale-95 flex-shrink-0"
>
```
**New:**
```html
<button
  @click="showNewTaskModal = true"
  class="px-3 sm:px-5 py-2 min-h-[36px] bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl transition-all shadow-lg shadow-makoclaw-accent/20 hover:shadow-makoclaw-accent/40 hover:-translate-y-0.5 text-sm font-bold flex items-center gap-1.5 sm:gap-2 active:scale-95 flex-shrink-0"
>
```
**Rationale:** Adding `hover:-translate-y-0.5` matches the Dashboard's CTA button lift effect. Small change, but creates a premium feel on the primary action button.

---

### Enhancement 14: Specialist Panel Header Border Treatment

**Target:** `pkg/web/frontend/src/components/Chat/SpecialistsPanel.vue` (template, line 2)
**Dashboard Reference:** Lines 139, 154, 177 — section headers use `border-b border-makoclaw-border/20 pb-2` with uppercase tracking
**Current:**
```html
<div class="glass-panel rounded-lg p-4 mb-4">
```
**New:**
```html
<div class="glass-panel rounded-lg p-4 mb-4 relative overflow-hidden">
  <!-- Soft glow at top-left -->
  <div class="absolute -top-8 -left-8 w-20 h-20 bg-blue-500/10 blur-[20px] rounded-full pointer-events-none"></div>
```
**Rationale:** Mirrors the Dashboard's chart panels (lines 107, 121) which each have a small soft glow blob in the top-left corner. Keeps the specialist panel visually consistent with the premium language.

---

## Summary

| Metric | Value |
|--------|-------|
| **Total Enhancements** | 14 |
| **Files Affected** | 6 |
| **Dashboard Effects Adopted** | 7 distinct effect types |

### Files Affected

1. `pkg/web/frontend/src/views/ChatView.vue` — Enhancements 1, 2, 3, 9
2. `pkg/web/frontend/src/views/TasksView.vue` — Enhancements 8, 13
3. `pkg/web/frontend/src/components/MessageBubble.vue` — Enhancement 4
4. `pkg/web/frontend/src/components/Chat/SpecialistsPanel.vue` — Enhancements 5, 14
5. `pkg/web/frontend/src/components/Tasks/KanbanColumn.vue` — Enhancements 6, 7, 12
6. `pkg/web/frontend/src/components/Tasks/NewTaskModal.vue` — Enhancement 10
7. `pkg/web/frontend/src/components/Tasks/TaskDetailsModal.vue` — Enhancement 11

### Dashboard Effects Being Adopted

| Effect | Enhancements | Where Applied |
|--------|-------------|---------------|
| Animated floating blobs | 1, 8 | Chat background, Tasks background |
| Gradient glow behind elements | 2, 7, 10, 11, 14 | Chat empty state, task cards, modals, specialist panel |
| Animated bottom-line on hover | 2, 3, 4, 7 | Chat empty state, input area focus, message bubbles, task cards |
| Icon hover transform (scale+rotate) | 2 | Chat empty state icon |
| Hover shadow elevation | 12 | Kanban columns |
| Button lift on hover | 13 | Create task button |
| Activity item left glow line | 5 | Specialist cards |

### Effects Intentionally NOT Applied

- **Staggered fade-in-up animation** — Chat messages already use `animate-slideUp` and task cards are dynamically loaded; staggered entrance would cause layout jank on real-time updates
- **Trend badges with SVG arrows** — No statistical trend data exists in Chat or Tasks
- **Drop-shadow on large text** — Chat doesn't have large stat numbers; would be misapplied
- **Pulse dots on all headers** — Only applied to Kanban columns where status context is meaningful (Enhancement 6)

### Performance Considerations

- All blur effects use `pointer-events-none` to avoid layout thrashing
- Glow blobs are absolutely positioned with fixed sizes (no dynamic reflows)
- Animations use `transform` and `opacity` only (GPU-composited, no layout triggers)
- No `animate-pulse` on large elements; pulse only on small 2px dots
- `will-change` is not needed since transitions are brief hover effects
