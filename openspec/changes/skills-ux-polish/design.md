# Design: skills-ux-polish

**Change:** skills-ux-polish
**Status:** Draft
**Date:** 2026-02-27
**Target file:** `pkg/web/frontend/src/views/SkillsView.vue`

---

## 1. Technical Approach

Apply the established glass morphism design language to the Skills page (`SkillsView.vue`) to match the visual consistency of the already-polished Dashboard, Chat, and Tasks pages. This is a **CSS-only refactor** — no logic, data flow, or API changes. All modifications are class-level substitutions within the existing template structure.

The Skills page is a single-file component (1150 lines) with four tab panels (Installed, Marketplace, Bundles, My Submissions), three modals (Create Skill, Edit Skill, View Skill, Submit to Marketplace), and loading/empty states. Every section requires audit and correction.

### Design Principles Applied

1. **Glass morphism first** — `glass-panel` for content panels, `glass-sticky` for the header
2. **Responsive spacing** — `p-2 sm:p-3 md:p-4` progressive padding, never static `p-4`/`p-5`/`p-6`
3. **Touch-friendly** — `min-h-[36px] min-w-[36px]` on all interactive elements
4. **Subdued borders** — `/20` or `/30` opacity, never full opacity
5. **Standard text scale** — only `text-xs`, `text-sm`, `text-base`; no custom `text-[10px]`
6. **Consistent corners** — `rounded-xl` for buttons/controls, `rounded-2xl` for panels
7. **No gradient text** on page titles — plain `text-makoclaw-text`
8. **Subtle hover** — `hover:-translate-y-0.5`, `active:scale-95`
9. **Glass button backgrounds** — `bg-makoclaw-surface/40` not full-opacity

---

## 2. Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Scope | CSS classes only | No logic changes needed; behavior is correct |
| File count | 1 file modified | All Skills UI lives in `SkillsView.vue` |
| Shared classes | Reuse `glass-panel`, `glass-sticky`, `card-interactive` from `globals.css` | Consistency with Dashboard/Tasks |
| Tab bar | Keep existing `tab-button` classes but wrap in glass container | Tab system already has shared CSS |
| Modals | Upgrade to glass morphism + `rounded-2xl` + `shadow-xl` | Match modern modal feel |
| Loading skeletons | Upgrade to `glass-panel` style with correct borders/corners | Match Dashboard skeleton pattern |
| Empty states | Standardize with glass panel wrapping + consistent typography | Visual consistency |
| Localization strings | **Preserve as-is** (some buttons say "Ver", "Editar", "Desinstalar", "Guardar Cambios") | Not in scope for UX polish |

---

## 3. File Changes

| File | Action | Description |
|------|--------|-------------|
| `pkg/web/frontend/src/views/SkillsView.vue` | Modify | Apply glass morphism, responsive spacing, touch targets, consistent typography across all sections |

No new files. No deleted files.

---

## 4. Detailed Change Specification

### 4.1 Page Root Container (line 2)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 1 | `class="h-full flex flex-col bg-makoclaw-bg"` | `class="h-full flex flex-col bg-makoclaw-bg relative overflow-hidden"` | Match Dashboard root; enable decorative blobs |

Add decorative background blobs after the root div opens (before the header), matching Dashboard:

```html
<!-- Decorative background blobs -->
<div class="absolute -top-24 -right-24 w-96 h-96 bg-makoclaw-accent/10 blur-[100px] rounded-full pointer-events-none"></div>
<div class="absolute top-1/2 -left-24 w-72 h-72 bg-blue-500/5 blur-[80px] rounded-full pointer-events-none"></div>
```

Note: No `animate-pulse` or animation on blobs (matches Dashboard pattern of static or CSS-only `float` animation via scoped styles).

### 4.2 Header (line 4)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 2 | `class="flex-none p-4 border-b border-makoclaw-border bg-makoclaw-surface flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"` | `class="flex-none p-2 sm:p-3 md:p-4 glass-sticky z-20 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 sm:gap-3"` | Responsive padding; glass-sticky instead of solid bg; subdued border via glass-sticky; z-20 for stacking |

### 4.3 Page Title (line 6)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 3 | `class="text-xl font-bold bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent"` | `class="text-lg sm:text-xl font-bold tracking-tight text-makoclaw-text"` | No gradient text on titles; responsive size; plain text color; tracking-tight for polish |

### 4.4 Page Subtitle (line 7)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 4 | `class="text-sm text-makoclaw-text-secondary mt-1"` | `class="text-xs text-makoclaw-text-secondary/70 mt-0.5"` | Standard subtitle: text-xs, /70 opacity, tighter margin |

### 4.5 Create Skill Button (lines 10-18)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 5 | `class="px-4 py-2 text-sm font-medium bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent-hover transition-colors flex items-center gap-2 shadow-sm"` | `class="px-3 sm:px-4 py-2 min-h-[36px] text-sm font-bold bg-makoclaw-accent text-white rounded-xl hover:bg-makoclaw-accent-hover transition-all shadow-lg shadow-makoclaw-accent/20 hover:shadow-makoclaw-accent/40 flex items-center gap-1.5 sm:gap-2 active:scale-95 flex-shrink-0"` | Touch target; rounded-xl; active:scale-95; shadow-lg with accent glow; responsive gap; font-bold per Tasks page pattern |

### 4.6 Tab Container (line 19)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 6 | `class="flex bg-makoclaw-bg rounded-lg p-1 border border-makoclaw-border"` | `class="flex bg-makoclaw-bg/40 backdrop-blur-sm rounded-xl p-1 border border-makoclaw-border/20"` | Glass-like tab container; rounded-xl; subdued border; backdrop blur |

### 4.7 Tab Buttons (lines 20-39) — all four tabs

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 7 | Tab buttons have no `min-h` | Add `min-h-[36px]` to each tab button class string | Touch targets on all interactive elements |

Each tab button should include `min-h-[36px]` in its class binding. The shared `tab-button` class in globals.css already provides the base styling. We just need to add the touch target inline since `tab-button` doesn't include it.

### 4.8 Content Area (line 45)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 8 | `class="flex-1 overflow-auto p-4 md:p-6 custom-scrollbar"` | `class="flex-1 overflow-auto p-2 sm:p-3 md:p-4 custom-scrollbar relative z-10"` | Responsive 3-step padding; z-10 above blobs; remove oversized md:p-6 |

### 4.9 Loading Skeleton (lines 47-61)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 9 | `class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"` (skeleton grid) | `class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 sm:gap-3"` | Responsive gaps |
| 10 | `class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5"` (skeleton card) | `class="glass-panel rounded-2xl p-3 sm:p-4"` | Glass morphism; rounded-2xl for panels; responsive padding |

### 4.10 Installed Skills — Empty State (lines 68-71)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 11 | `class="text-center py-12 text-makoclaw-text-secondary"` (empty wrapper) | `class="glass-panel rounded-2xl p-6 sm:p-8 text-center"` | Wrap empty state in glass panel |
| 12 | `class="text-lg"` (empty title) | `class="text-sm font-medium text-makoclaw-text"` | Correct text scale; standard size |
| 13 | `class="text-sm mt-2"` (empty subtitle) | `class="text-xs text-makoclaw-text-secondary/70 mt-1.5"` | Match subtitle convention |

### 4.11 Installed Skills Grid (line 72)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 14 | `class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"` | `class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 sm:gap-3"` | Responsive grid gaps |

### 4.12 Installed Skill Cards (lines 73-115)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 15 | `class="card-interactive p-5"` | `class="glass-panel rounded-2xl p-3 sm:p-4 transition-all duration-300 hover:shadow-lg hover:shadow-makoclaw-accent/10 hover:-translate-y-0.5 group"` | Glass panel; responsive padding; subtle hover lift; rounded-2xl |
| 16 | `class="font-semibold truncate"` (skill name h3) | `class="text-sm font-bold text-makoclaw-text truncate"` | Explicit text size; font-bold for emphasis |
| 17 | `class="text-sm text-makoclaw-text-secondary mt-1 line-clamp-2"` (description) | `class="text-xs text-makoclaw-text-secondary/70 mt-1 line-clamp-2"` | text-xs for descriptions; /70 opacity |
| 18 | `class="text-xs text-makoclaw-text-secondary"` (usage count) | `class="text-xs text-makoclaw-text-secondary/50 mt-0.5"` | Slightly dimmer; margin top |

### 4.13 Source Badge (lines 86-92)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 19 | `class="ml-2 px-2 py-0.5 text-xs rounded-full flex-shrink-0"` | `class="ml-2 px-2 py-0.5 text-xs rounded-full flex-shrink-0 border border-makoclaw-border/20"` | Add subtle border for visual definition |

### 4.14 Installed Skill Action Buttons (lines 94-114)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 20 | `class="flex items-center gap-2 mt-4 flex-wrap"` (actions container) | `class="flex items-center gap-2 mt-3 flex-wrap"` | Slightly tighter spacing |
| 21 | `class="px-3 py-1.5 text-xs bg-makoclaw-bg rounded-lg ..."` (View button) | `class="px-3 py-1.5 min-h-[36px] text-xs bg-makoclaw-surface/40 rounded-xl hover:bg-makoclaw-border/50 transition-all active:scale-95"` | Touch target; rounded-xl; glass bg; active:scale-95 |
| 22 | `class="px-3 py-1.5 text-xs bg-makoclaw-accent/10 text-makoclaw-accent rounded-lg ..."` (Edit button) | `class="px-3 py-1.5 min-h-[36px] text-xs bg-makoclaw-accent/10 text-makoclaw-accent rounded-xl hover:bg-makoclaw-accent/20 transition-all active:scale-95"` | Touch target; rounded-xl; active:scale-95 |
| 23 | `class="px-3 py-1.5 text-xs bg-green-500/10 text-green-400 rounded-lg ..."` (Submit button) | `class="px-3 py-1.5 min-h-[36px] text-xs bg-green-500/10 text-green-400 rounded-xl hover:bg-green-500/20 transition-all active:scale-95"` | Touch target; rounded-xl; active:scale-95 |
| 24 | `class="px-3 py-1.5 text-xs text-red-400 bg-red-500/10 rounded-lg ..."` (Uninstall button) | `class="px-3 py-1.5 min-h-[36px] text-xs text-red-400 bg-red-500/10 rounded-xl hover:bg-red-500/20 transition-all active:scale-95"` | Touch target; rounded-xl; active:scale-95 |

### 4.15 Marketplace Loading Spinner (line 122)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 25 | `class="flex items-center justify-center py-12"` | `class="flex items-center justify-center py-8 sm:py-12"` | Responsive vertical padding |

### 4.16 Marketplace Empty State (lines 124-127)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 26 | `class="text-center py-12 text-makoclaw-text-secondary"` | `class="glass-panel rounded-2xl p-6 sm:p-8 text-center"` | Glass panel wrapping |
| 27 | `class="text-lg"` (empty title) | `class="text-sm font-medium text-makoclaw-text"` | Correct text scale |
| 28 | `class="text-sm mt-2"` (empty subtitle) | `class="text-xs text-makoclaw-text-secondary/70 mt-1.5"` | Match subtitle convention |

### 4.17 Trending Section (lines 130-139)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 29 | `class="mb-6"` (trending wrapper) | `class="mb-3 sm:mb-4"` | Responsive margin; less excessive |
| 30 | `class="text-sm font-medium text-makoclaw-text-secondary mb-3"` (heading) | `class="text-xs font-medium uppercase tracking-wide text-makoclaw-text-secondary/70 mb-2 sm:mb-3 px-1"` | Match Dashboard section heading pattern |
| 31 | `class="grid grid-cols-1 sm:grid-cols-3 gap-3"` | `class="grid grid-cols-1 sm:grid-cols-3 gap-2 sm:gap-3"` | Responsive gaps |
| 32 | `class="card p-3"` (trending card) | `class="glass-panel rounded-xl p-3 sm:p-4 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-makoclaw-accent/10"` | Glass panel; hover effects |
| 33 | `class="font-medium text-sm"` (trending name) | `class="text-sm font-bold text-makoclaw-text"` | Explicit color; font-bold |
| 34 | `class="text-xs text-makoclaw-text-secondary mt-1 line-clamp-2"` | `class="text-xs text-makoclaw-text-secondary/70 mt-1 line-clamp-2"` | /70 opacity |
| 35 | `class="text-xs text-makoclaw-text-secondary mt-2"` (install count) | `class="text-xs text-makoclaw-text-secondary/50 mt-1.5"` | Dimmer; tighter margin |

### 4.18 Marketplace Skills Grid (line 140)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 36 | `class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"` | `class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 sm:gap-3"` | Responsive grid gaps |

### 4.19 Marketplace Skill Cards (lines 141-224)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 37 | `class="card-interactive p-5"` | `class="glass-panel rounded-2xl p-3 sm:p-4 transition-all duration-300 hover:shadow-lg hover:shadow-makoclaw-accent/10 hover:-translate-y-0.5 group"` | Glass panel; responsive padding; hover effects |
| 38 | `class="font-semibold flex-1"` (marketplace skill name) | `class="text-sm font-bold text-makoclaw-text flex-1"` | Explicit sizing/color |

### 4.20 Security Score Badge (lines 148-156)

No change needed — the conditional coloring and size are already correct.

### 4.21 Marketplace Skill Metadata

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 39 | `class="text-sm text-makoclaw-text-secondary mt-1 line-clamp-2"` (description) | `class="text-xs text-makoclaw-text-secondary/70 mt-1 line-clamp-2"` | text-xs; /70 opacity |
| 40 | `class="text-xs text-makoclaw-text-secondary mt-2"` (author) | `class="text-xs text-makoclaw-text-secondary/50 mt-1.5"` | Dimmer; tighter margin |
| 41 | `class="px-2 py-0.5 text-xs bg-makoclaw-bg rounded-full text-makoclaw-text-secondary"` (tags) | `class="px-2 py-0.5 text-xs bg-makoclaw-surface/40 rounded-full text-makoclaw-text-secondary/70 border border-makoclaw-border/20"` | Glass bg; border; /70 opacity |
| 42 | `class="flex items-center gap-2 mt-4"` (action row) | `class="flex items-center gap-2 mt-3"` | Tighter margin |

### 4.22 Marketplace Install Button (line 170)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 43 | `class="px-4 py-1.5 text-sm bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors disabled:opacity-50"` | `class="px-4 py-1.5 min-h-[36px] text-sm font-medium bg-makoclaw-accent text-white rounded-xl hover:bg-makoclaw-accent-hover transition-all shadow-sm disabled:opacity-50 active:scale-95"` | Touch target; rounded-xl; proper hover; active:scale-95 |

### 4.23 Fork Button (lines 175-181)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 44 | `class="btn-ghost text-xs px-2 py-1"` | `class="px-3 py-1.5 min-h-[36px] text-xs bg-makoclaw-surface/40 rounded-xl hover:bg-makoclaw-border/50 transition-all active:scale-95 text-makoclaw-text-secondary hover:text-makoclaw-text"` | Touch target; glass bg; rounded-xl; explicit hover colors |

### 4.24 Rate Button (lines 188-193)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 45 | `class="text-xs px-2 py-1 bg-makoclaw-bg rounded-lg hover:bg-makoclaw-border/50 transition-colors"` | `class="text-xs px-3 py-1.5 min-h-[36px] bg-makoclaw-surface/40 rounded-xl hover:bg-makoclaw-border/50 transition-all active:scale-95"` | Touch target; glass bg; rounded-xl |

### 4.25 Inline Rating Widget (line 197)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 46 | `class="mt-3 p-3 bg-makoclaw-bg rounded-lg border border-makoclaw-border"` | `class="mt-3 p-3 bg-makoclaw-bg/60 backdrop-blur-sm rounded-xl border border-makoclaw-border/20"` | Glass-like; rounded-xl; subdued border |

### 4.26 Rating Submit Button (line 220)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 47 | `class="text-xs px-3 py-1 bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors disabled:opacity-50"` | `class="text-xs px-3 py-1.5 min-h-[36px] bg-makoclaw-accent text-white rounded-xl hover:bg-makoclaw-accent-hover transition-all disabled:opacity-50 active:scale-95"` | Touch target; rounded-xl; proper hover |

### 4.27 Bundles Loading/Empty States (lines 231-237)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 48 | `class="text-center py-12 text-makoclaw-text-secondary"` (empty) | `class="glass-panel rounded-2xl p-6 sm:p-8 text-center"` | Glass panel wrapping |
| 49 | `class="text-lg"` (empty title) | `class="text-sm font-medium text-makoclaw-text"` | Correct text scale |
| 50 | `class="text-sm mt-2"` (empty subtitle) | `class="text-xs text-makoclaw-text-secondary/70 mt-1.5"` | Match subtitle convention |

### 4.28 Bundles Grid (line 238)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 51 | `class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"` | `class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 sm:gap-3"` | Responsive gaps |

### 4.29 Bundle Cards (lines 239-263)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 52 | `class="card p-4 flex flex-col gap-3"` | `class="glass-panel rounded-2xl p-3 sm:p-4 flex flex-col gap-2 sm:gap-3 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-makoclaw-accent/10"` | Glass panel; responsive; hover effects |
| 53 | `class="text-2xl"` (bundle icon) | `class="text-xl sm:text-2xl"` | Responsive icon size |
| 54 | `class="font-medium text-sm"` (bundle name) | `class="text-sm font-bold text-makoclaw-text"` | font-bold; explicit color |
| 55 | `class="text-xs text-makoclaw-text-secondary mt-1 line-clamp-2"` (desc) | `class="text-xs text-makoclaw-text-secondary/70 mt-1 line-clamp-2"` | /70 opacity |
| 56 | `class="btn-primary text-xs w-full"` (Install Bundle button) | `class="px-4 py-2 min-h-[36px] text-xs font-medium bg-makoclaw-accent text-white rounded-xl hover:bg-makoclaw-accent-hover transition-all shadow-sm active:scale-95 w-full disabled:opacity-50"` | Inline glass-compatible button; touch target; rounded-xl |

### 4.30 My Submissions Empty State (lines 271-274)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 57 | `class="text-center py-12 text-makoclaw-text-secondary"` | `class="glass-panel rounded-2xl p-6 sm:p-8 text-center"` | Glass panel wrapping |
| 58 | `class="text-lg"` | `class="text-sm font-medium text-makoclaw-text"` | Correct text scale |
| 59 | `class="text-sm mt-2"` | `class="text-xs text-makoclaw-text-secondary/70 mt-1.5"` | Match subtitle convention |

### 4.31 My Submissions Grid (line 275)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 60 | `class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"` | `class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 sm:gap-3"` | Responsive gaps |

### 4.32 Submission Cards (lines 276-303)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 61 | `class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5"` | `class="glass-panel rounded-2xl p-3 sm:p-4 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-makoclaw-accent/10"` | Glass panel; responsive padding; hover |
| 62 | `class="font-semibold flex-1"` (submission name) | `class="text-sm font-bold text-makoclaw-text flex-1"` | Explicit sizing/color |
| 63 | `class="text-sm text-makoclaw-text-secondary mt-1 line-clamp-2"` (desc) | `class="text-xs text-makoclaw-text-secondary/70 mt-1 line-clamp-2"` | text-xs; /70 opacity |
| 64 | `class="text-xs text-makoclaw-text-secondary mt-2"` (security score) | `class="text-xs text-makoclaw-text-secondary/50 mt-1.5"` | Dimmer; tighter margin |

### 4.33 Create Skill Modal Container (line 313)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 65 | `class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-3xl w-full max-h-[90vh] flex flex-col shadow-2xl"` | `class="glass-panel rounded-2xl max-w-3xl w-full max-h-[90vh] flex flex-col"` | Glass panel (includes shadow-xl); rounded-2xl; remove shadow-2xl |

### 4.34 Create Skill Modal Header (line 315)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 66 | `class="flex items-center justify-between p-4 border-b border-makoclaw-border"` | `class="flex items-center justify-between p-3 sm:p-4 border-b border-makoclaw-border/20"` | Responsive padding; subdued border |

### 4.35 Create Skill Modal Title (line 320)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 67 | `class="font-bold text-lg"` | `class="font-bold text-sm sm:text-base text-makoclaw-text"` | Standard text scale; responsive |

### 4.36 Modal Close Button (line 322)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 68 | `class="p-1.5 hover:bg-makoclaw-bg rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"` | `class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center hover:bg-makoclaw-bg/60 rounded-xl text-makoclaw-text-secondary hover:text-makoclaw-text transition-all active:scale-95"` | Touch target; rounded-xl; glass bg on hover |

### 4.37 Modal Content Area (line 328)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 69 | `class="flex-1 overflow-auto p-5 custom-scrollbar"` | `class="flex-1 overflow-auto p-3 sm:p-4 custom-scrollbar"` | Responsive padding |

### 4.38 AI Assistant Section (line 332)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 70 | `class="bg-gradient-to-r from-makoclaw-accent/10 to-blue-500/10 border border-makoclaw-accent/30 rounded-xl p-5 mb-6"` | `class="bg-gradient-to-r from-makoclaw-accent/10 to-blue-500/10 border border-makoclaw-accent/20 rounded-xl p-3 sm:p-4 mb-3 sm:mb-4"` | Responsive padding/margin; subdued border |
| 71 | `class="w-10 h-10 rounded-lg ..."` (AI icon container) | `class="w-8 h-8 rounded-lg ..."` | Proportionate: w-8 h-8 with w-4 h-4 icon = 50% fill ratio |

### 4.39 "New" Badge (line 342)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 72 | `class="px-2 py-0.5 text-[10px] font-bold uppercase bg-makoclaw-accent text-white rounded-full"` | `class="px-2 py-0.5 text-xs font-bold uppercase bg-makoclaw-accent text-white rounded-full"` | No custom text sizes — use `text-xs` |

### 4.40 AI Generate Button (lines 352-365)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 73 | `class="w-full px-4 py-2.5 bg-gradient-to-r from-makoclaw-accent to-blue-600 hover:from-makoclaw-accent-hover hover:to-blue-700 text-white rounded-xl font-bold shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98]"` | `class="w-full px-4 py-2.5 min-h-[36px] bg-gradient-to-r from-makoclaw-accent to-blue-600 hover:from-makoclaw-accent-hover hover:to-blue-700 text-white rounded-xl font-bold shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed active:scale-95"` | Touch target; `active:scale-95` not `active:scale-[0.98]` for consistency |

### 4.41 Form Input Fields (lines 386, 397-401, 410-414)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 74 | `class="w-full px-3 py-2.5 bg-makoclaw-bg border border-makoclaw-border rounded-lg ..."` | `class="w-full px-3 py-2.5 min-h-[36px] bg-makoclaw-bg/60 border border-makoclaw-border/20 rounded-xl ..."` | Touch target; glass bg; rounded-xl; subdued border |

Apply to all three form fields (name input, description textarea, context textarea).

### 4.42 Modal Footer (line 441)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 75 | `class="p-4 border-t border-makoclaw-border flex justify-end gap-3"` | `class="p-3 sm:p-4 border-t border-makoclaw-border/20 flex justify-end gap-2 sm:gap-3"` | Responsive padding; subdued border |

### 4.43 Generate Button in Footer (lines 448-459)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 76 | `class="px-5 py-2 text-sm font-semibold bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent-hover shadow-lg shadow-makoclaw-accent/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2"` | `class="px-4 sm:px-5 py-2 min-h-[36px] text-sm font-bold bg-makoclaw-accent text-white rounded-xl hover:bg-makoclaw-accent-hover shadow-lg shadow-makoclaw-accent/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2 active:scale-95"` | Touch target; rounded-xl; font-bold; active:scale-95 |

### 4.44 Save Button in Footer (lines 462-473)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 77 | `class="px-5 py-2 text-sm font-semibold bg-green-600 text-white rounded-lg hover:bg-green-500 shadow-lg shadow-green-600/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2"` | `class="px-4 sm:px-5 py-2 min-h-[36px] text-sm font-bold bg-green-600 text-white rounded-xl hover:bg-green-500 shadow-lg shadow-green-600/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2 active:scale-95"` | Touch target; rounded-xl; font-bold; active:scale-95 |

### 4.45 Edit Skill Modal (lines 480-515)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 78 | `class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-4xl w-full h-[85vh] flex flex-col shadow-2xl"` | `class="glass-panel rounded-2xl max-w-4xl w-full h-[85vh] flex flex-col"` | Glass panel; rounded-2xl; remove shadow-2xl (glass-panel has shadow-xl) |
| 79 | `class="flex items-center justify-between p-4 border-b border-makoclaw-border"` (header) | `class="flex items-center justify-between p-3 sm:p-4 border-b border-makoclaw-border/20"` | Responsive; subdued border |
| 80 | `class="font-bold text-lg text-makoclaw-accent"` (title) | `class="font-bold text-sm sm:text-base text-makoclaw-accent"` | Standard text scale |
| 81 | `class="p-1 hover:bg-makoclaw-bg rounded text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"` (close btn) | `class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center hover:bg-makoclaw-bg/60 rounded-xl text-makoclaw-text-secondary hover:text-makoclaw-text transition-all active:scale-95"` | Touch target; rounded-xl |
| 82 | `class="flex-1 overflow-hidden p-4"` (content) | `class="flex-1 overflow-hidden p-3 sm:p-4"` | Responsive padding |
| 83 | `class="flex-1 w-full p-4 bg-makoclaw-bg border border-makoclaw-border rounded-xl text-sm font-mono focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent transition-all resize-none shadow-inner custom-scrollbar"` (textarea) | `class="flex-1 w-full p-3 sm:p-4 bg-makoclaw-bg/60 border border-makoclaw-border/20 rounded-xl text-sm font-mono focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent transition-all resize-none custom-scrollbar"` | Glass bg; subdued border; remove shadow-inner |
| 84 | `class="p-4 border-t border-makoclaw-border flex justify-end gap-3"` (footer) | `class="p-3 sm:p-4 border-t border-makoclaw-border/20 flex justify-end gap-2 sm:gap-3"` | Responsive; subdued border |
| 85 | `class="px-6 py-2 text-sm font-semibold bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent-hover shadow-lg shadow-makoclaw-accent/20 disabled:opacity-50 transition-all flex items-center gap-2"` (save btn) | `class="px-4 sm:px-5 py-2 min-h-[36px] text-sm font-bold bg-makoclaw-accent text-white rounded-xl hover:bg-makoclaw-accent-hover shadow-lg shadow-makoclaw-accent/20 disabled:opacity-50 transition-all flex items-center gap-2 active:scale-95"` | Touch target; rounded-xl; font-bold |

### 4.46 View Skill Modal (lines 517-530)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 86 | `class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-2xl w-full max-h-[80vh] flex flex-col shadow-xl"` | `class="glass-panel rounded-2xl max-w-2xl w-full max-h-[80vh] flex flex-col"` | Glass panel; rounded-2xl |
| 87 | `class="flex items-center justify-between p-4 border-b border-makoclaw-border"` (header) | `class="flex items-center justify-between p-3 sm:p-4 border-b border-makoclaw-border/20"` | Responsive; subdued border |
| 88 | `class="font-semibold"` (title) | `class="text-sm sm:text-base font-bold text-makoclaw-text"` | Standard scale; explicit color |
| 89 | `class="p-1 hover:bg-makoclaw-bg rounded text-makoclaw-text-secondary hover:text-makoclaw-text"` (close) — uses `&times;` text | `class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center hover:bg-makoclaw-bg/60 rounded-xl text-makoclaw-text-secondary hover:text-makoclaw-text transition-all active:scale-95"` | Touch target; rounded-xl |
| 90 | `class="flex-1 overflow-auto p-4 custom-scrollbar"` (content) | `class="flex-1 overflow-auto p-3 sm:p-4 custom-scrollbar"` | Responsive |

### 4.47 Submit to Marketplace Modal (lines 532-634)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 91 | `class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-lg w-full max-h-[80vh] flex flex-col shadow-xl"` | `class="glass-panel rounded-2xl max-w-lg w-full max-h-[80vh] flex flex-col"` | Glass panel; rounded-2xl |
| 92 | `class="flex items-center justify-between p-4 border-b border-makoclaw-border"` (header) | `class="flex items-center justify-between p-3 sm:p-4 border-b border-makoclaw-border/20"` | Responsive; subdued border |
| 93 | `class="font-semibold"` (title) | `class="text-sm sm:text-base font-bold text-makoclaw-text"` | Standard scale |
| 94 | `class="p-1 hover:bg-makoclaw-bg rounded text-makoclaw-text-secondary hover:text-makoclaw-text"` (close) | `class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center hover:bg-makoclaw-bg/60 rounded-xl text-makoclaw-text-secondary hover:text-makoclaw-text transition-all active:scale-95"` | Touch target; rounded-xl |
| 95 | `class="flex-1 overflow-auto p-4 custom-scrollbar space-y-4"` (content) | `class="flex-1 overflow-auto p-3 sm:p-4 custom-scrollbar space-y-3 sm:space-y-4"` | Responsive |

### 4.48 Submit Modal — Security Scan Panel (line 547)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 96 | `class="p-3 rounded-lg"` | `class="p-3 rounded-xl"` | Consistent rounded-xl |

### 4.49 Submit Modal — Form Fields (lines 576, 614-618)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 97 | `class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm"` (select, input) | `class="w-full px-3 py-2 min-h-[36px] bg-makoclaw-bg/60 border border-makoclaw-border/20 rounded-xl text-sm"` | Touch target; glass bg; rounded-xl; subdued border |

### 4.50 Submit Modal — Visibility Radio Options (lines 590-607)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 98 | Active: `border-makoclaw-accent bg-makoclaw-accent/5`, Inactive: `border-makoclaw-border hover:border-makoclaw-accent/50` | Active: `border-makoclaw-accent/30 bg-makoclaw-accent/5`, Inactive: `border-makoclaw-border/20 hover:border-makoclaw-accent/30` | Subdued borders |
| 99 | `class="... p-3 rounded-lg border ..."` | `class="... p-3 rounded-xl border ..."` | rounded-xl |

### 4.51 Submit Modal Footer (line 622)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 100 | `class="p-4 border-t border-makoclaw-border flex justify-end gap-3"` | `class="p-3 sm:p-4 border-t border-makoclaw-border/20 flex justify-end gap-2 sm:gap-3"` | Responsive; subdued border |

### 4.52 Submit Button (line 627)

| # | Current | New | Rationale |
|---|---------|-----|-----------|
| 101 | `class="px-4 py-2 text-sm font-medium bg-green-600 text-white rounded-lg hover:bg-green-500 disabled:opacity-50 disabled:cursor-not-allowed"` | `class="px-4 py-2 min-h-[36px] text-sm font-bold bg-green-600 text-white rounded-xl hover:bg-green-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all active:scale-95"` | Touch target; rounded-xl; font-bold; active:scale-95 |

### 4.53 Scoped Styles Addition

Add scoped animation styles to match Dashboard:

```css
<style scoped>
.line-clamp-2 { display: -webkit-box; -webkit-line-clamp: 2; line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

/* Custom scrollbar override for more subtle look */
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.1);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.2);
}
</style>
```

---

## 5. Summary of All Issues Found

| # | Section | Issue Category | Severity |
|---|---------|---------------|----------|
| 1 | Root container | Missing `relative overflow-hidden` for decorative elements | Medium |
| 2 | Header | Static `p-4` instead of responsive padding | High |
| 3 | Header | Solid `bg-makoclaw-surface` instead of `glass-sticky` | High |
| 4 | Title | Gradient text instead of plain `text-makoclaw-text` | High |
| 5 | Subtitle | `text-sm` instead of `text-xs text-makoclaw-text-secondary/70` | Medium |
| 6 | Create Skill button | Missing touch target, `rounded-lg` not `rounded-xl`, no active:scale | Medium |
| 7 | Tab container | Full-opacity border, `rounded-lg` not `rounded-xl`, no glass bg | Medium |
| 8 | Tab buttons | Missing `min-h-[36px]` touch targets | Medium |
| 9 | Content area | Static `p-4 md:p-6` instead of `p-2 sm:p-3 md:p-4` | High |
| 10 | Loading skeleton | `bg-makoclaw-surface` instead of `glass-panel`; static `gap-4` | Medium |
| 11-13 | Empty states (x4) | No glass panel wrapping; `text-lg` title oversized | Medium |
| 14 | Skill grids (x5) | Static `gap-4` instead of responsive `gap-2 sm:gap-3` | Medium |
| 15 | Installed skill cards | `card-interactive` instead of `glass-panel`; static `p-5` | High |
| 16 | Skill card titles | No explicit text size | Low |
| 17 | Card descriptions | `text-sm` instead of `text-xs` | Medium |
| 18 | Source badges | Missing border | Low |
| 19-24 | Action buttons (x4) | Missing touch targets; `rounded-lg` not `rounded-xl`; no active:scale | High |
| 25 | Marketplace loading | Static vertical padding | Low |
| 26-28 | Marketplace empty state | No glass panel; oversized text | Medium |
| 29-35 | Trending section | No glass panel; static gaps; non-standard heading style | Medium |
| 36 | Marketplace grid | Static gap | Medium |
| 37-38 | Marketplace cards | `card-interactive` instead of `glass-panel`; static `p-5` | High |
| 39-41 | Marketplace metadata | `text-sm` instead of `text-xs`; no glass bg on tags | Medium |
| 42 | Marketplace action row | Static `mt-4` | Low |
| 43 | Install button | Missing touch target; `rounded-lg` | Medium |
| 44 | Fork button | Missing touch target; no explicit styling | Medium |
| 45 | Rate button | Missing touch target; `rounded-lg` | Medium |
| 46 | Rating widget | `rounded-lg`; full-opacity border; solid bg | Medium |
| 47 | Rating submit button | Missing touch target; `rounded-lg` | Medium |
| 48-50 | Bundles empty state | No glass panel; oversized text | Medium |
| 51 | Bundles grid | Static gap | Medium |
| 52-56 | Bundle cards | `card` instead of `glass-panel`; static padding | High |
| 57-59 | Submissions empty state | No glass panel; oversized text | Medium |
| 60 | Submissions grid | Static gap | Medium |
| 61-64 | Submission cards | `bg-makoclaw-surface` instead of `glass-panel`; static `p-5` | High |
| 65-69 | Create modal | `shadow-2xl`; solid bg; full-opacity borders; static padding | High |
| 70-73 | AI section | Oversized icon; `text-[10px]` custom size; static padding | Medium |
| 74 | Form inputs (x3) | Solid bg; full-opacity border; `rounded-lg` | Medium |
| 75-77 | Modal footer buttons | Missing touch targets; `rounded-lg`; no active:scale | Medium |
| 78-85 | Edit modal | `shadow-2xl`; solid bg; full borders; `shadow-inner` textarea | High |
| 86-90 | View modal | Solid bg; full borders; no touch target on close | Medium |
| 91-101 | Submit modal | Solid bg; full borders; no touch targets; `rounded-lg` everywhere | High |

**Total issues found: 101 individual class changes across 53 numbered entries**

---

## 6. Testing Strategy

### Visual Verification
- [ ] Compare Skills page side-by-side with Dashboard page in both light and dark themes
- [ ] Verify glass morphism effect is visible with backdrop-blur on all panels
- [ ] Check decorative background blobs are visible but subtle
- [ ] Confirm no gradient text on page title

### Responsive Testing
- [ ] Mobile (320px): Verify responsive padding (p-2) on all sections; stacked layout for header
- [ ] Tablet (768px): Verify intermediate padding (p-3); grid collapses correctly
- [ ] Desktop (1280px): Verify full padding (p-4); 3-column grid displays properly

### Touch Target Verification
- [ ] All buttons have at least 36px height (inspect with dev tools)
- [ ] Tab buttons are tappable on mobile
- [ ] Modal close buttons are easily tappable
- [ ] Form submit buttons meet minimum touch size

### Modal Testing
- [ ] Create Skill modal: glass panel, responsive padding, all buttons accessible
- [ ] Edit Skill modal: glass panel, textarea readable, save button reachable
- [ ] View Skill modal: glass panel, content scrollable, close button tappable
- [ ] Submit to Marketplace modal: glass panel, radio buttons selectable, form inputs styled

### State Testing
- [ ] Loading skeleton: glass panel style with correct corners and borders
- [ ] Empty states (all 4 tabs): glass panel wrapping, correct text sizes
- [ ] Error states: red text visible against glass background
- [ ] Disabled buttons: 50% opacity visible against glass background

### Cross-tab Testing
- [ ] Installed tab: skill cards with glass morphism, action buttons styled
- [ ] Marketplace tab: trending section + grid with glass panels
- [ ] Bundles tab: bundle cards with glass panels
- [ ] My Submissions tab: submission cards with glass panels

### Performance Check
- [ ] No `animate-pulse` on large decorative elements
- [ ] `backdrop-blur-xl` on panels does not cause scroll jank (test on low-end device)
- [ ] Tab transitions remain smooth

---

## 7. Recommended Next Step

Proceed to `sdd-tasks` to generate the implementation task list, then `sdd-apply` to implement the changes in `pkg/web/frontend/src/views/SkillsView.vue`.
