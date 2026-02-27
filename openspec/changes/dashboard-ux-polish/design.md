# Design: Dashboard UX Polish

## Technical Approach

Surgical in-place edits to `DashboardView.vue` to align with the established glass morphism design language already implemented in `ChatView.vue` and `TasksView.vue`. The Dashboard currently has disproportionate elements, a header with unnecessary multiline text, oversized decorative icons, inconsistent border opacities, and spacing that does not match the responsive patterns established in sibling pages.

All changes are confined to `pkg/web/frontend/src/views/DashboardView.vue`. No new components, no new CSS classes, no structural changes to the data model or script logic. The shared `glass-panel`, `glass-sticky`, and `custom-scrollbar` utility classes from `globals.css` are already being used correctly — the issues are in the inline Tailwind classes surrounding them.

## Architecture Decisions

| Decision | Rationale |
|----------|-----------|
| Single-file edit only | All issues are presentational within DashboardView.vue; no shared components or CSS changes needed |
| Preserve animation system | The `animate-fade-in-up` stagger animations are well-implemented and unique to Dashboard — keep them |
| Remove decorative background blob animation | The `animate-pulse` on the top-right blob causes constant GPU repaints; replace with static opacity |
| Eliminate `<br/>` in welcome heading | The forced line break at "Welcome back,\<br/>" causes unnecessary multiline on wide screens |
| Reduce decorative SVG icon in banner | The 128x128px bolt icon in the welcome banner is disproportionate — reduce to 96x96 (w-24 h-24) |
| Standardize border opacity to /20 | Dashboard uses `/50` on several interactive borders while Chat/Tasks use `/20`; normalize |
| Keep chart min-heights | The `min-h-[200px] sm:min-h-[250px] md:min-h-[300px]` pattern is functional for Chart.js rendering — no change |

## File Changes

| File | Change Type | Description |
|------|-------------|-------------|
| `pkg/web/frontend/src/views/DashboardView.vue` | Modify | Fix all disproportionate elements, header layout, border opacities, spacing, icon sizes |

## Detailed Change Specification

### Section: Decorative Background Blobs (Lines 4-5)

- **Current**: `class="absolute -top-24 -right-24 w-96 h-96 bg-makoclaw-accent/10 blur-[100px] rounded-full pointer-events-none animate-pulse"`
- **New**: `class="absolute -top-24 -right-24 w-96 h-96 bg-makoclaw-accent/10 blur-[100px] rounded-full pointer-events-none"`
- **Rationale**: Remove `animate-pulse` from the 384px blob. This causes continuous GPU compositing for a purely decorative element. Chat and Tasks pages use static gradient meshes without animation on their background elements.

### Section: Header Bar (Lines 8-27)

#### Header Container (Line 8)
- **Current**: `class="flex-none p-2 sm:p-3 md:p-4 glass-sticky z-20"`
- **New**: `class="flex-none p-2 sm:p-3 md:p-4 glass-sticky z-20"` (no change — already matches pattern)
- **Rationale**: Already correct. The `glass-sticky` class applies `bg-makoclaw-bg/60 backdrop-blur-xl border-b border-makoclaw-border/30`.

#### Page Title (Line 11)
- **Current**: `class="text-xl sm:text-2xl font-black tracking-tight bg-gradient-to-r from-makoclaw-accent via-blue-400 to-cyan-400 bg-clip-text text-transparent"`
- **New**: `class="text-lg sm:text-xl font-bold tracking-tight text-makoclaw-text"`
- **Rationale**: The gradient text is overly decorative for a page header. Chat uses `font-semibold text-xs md:text-sm` for its sidebar headers. Tasks uses standard text. The Dashboard title should be prominent but not flashy — use `text-lg sm:text-xl font-bold` with standard text color. Remove the 3-stop gradient which is inconsistent with every other page header in the app.

#### Subtitle Line (Lines 14-17)
- **Current**: Multi-element `<p>` with `text-xs font-medium text-makoclaw-text-secondary/70 mt-1 flex items-center gap-2` containing an animated pulse dot + "System operational" + username
- **New**: `class="text-xs text-makoclaw-text-secondary/70 mt-0.5 flex items-center gap-1.5"`
- **Rationale**: Reduce `mt-1` to `mt-0.5` for tighter header grouping. Reduce `gap-2` to `gap-1.5`. Remove `font-medium` — subtitle should be lighter weight than title.

#### Refresh Button (Line 20)
- **Current**: `class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-xl bg-makoclaw-surface border border-makoclaw-border/50 hover:border-makoclaw-accent/50 transition-all active:scale-90 group"`
- **New**: `class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-xl bg-makoclaw-surface/40 border border-makoclaw-border/20 hover:border-makoclaw-accent/30 transition-all active:scale-95 group"`
- **Rationale**: Border opacity `/50` is too prominent — normalize to `/20` resting and `/30` hover, matching Chat/Tasks. Change `bg-makoclaw-surface` (full opacity) to `bg-makoclaw-surface/40` to match glass morphism convention. Change `active:scale-90` to `active:scale-95` (Chat/Tasks use `active:scale-95`).

### Section: Welcome Banner (Lines 46-66)

#### Banner Container (Line 46)
- **Current**: `class="relative overflow-hidden bg-gradient-to-br from-makoclaw-accent via-blue-600 to-indigo-700 rounded-2xl p-4 sm:p-5 md:p-6 shadow-2xl shadow-makoclaw-accent/20 group animate-fade-in-up"`
- **New**: `class="relative overflow-hidden bg-gradient-to-br from-makoclaw-accent via-blue-600 to-indigo-700 rounded-2xl p-3 sm:p-4 md:p-5 shadow-xl shadow-makoclaw-accent/20 group animate-fade-in-up"`
- **Rationale**: Reduce padding from `p-4 sm:p-5 md:p-6` to `p-3 sm:p-4 md:p-5` to be less spacious (consistent with panel padding elsewhere). Change `shadow-2xl` to `shadow-xl` — other panels use `shadow-xl` via `glass-panel`.

#### Decorative Bolt Icon (Lines 52-53)
- **Current**: `class="w-32 h-32"` (128x128px SVG)
- **New**: `class="w-20 h-20"` (80x80px SVG)
- **Rationale**: A 128px decorative icon is disproportionately large relative to the banner content. Reducing to 80px keeps the decorative feel while being proportionate.

#### Also the decorative icon container (Line 52)
- **Current**: `class="absolute top-0 right-0 p-8 transform rotate-12 opacity-[0.07] transition-transform group-hover:rotate-45 group-hover:scale-110 duration-1000 hidden md:block"`
- **New**: `class="absolute top-0 right-0 p-6 transform rotate-12 opacity-[0.07] transition-transform group-hover:rotate-45 group-hover:scale-110 duration-1000 hidden md:block"`
- **Rationale**: Reduce `p-8` to `p-6` — the icon container padding is excessive, pushing the icon too far from the corner.

#### "Command Center" Badge (Line 57)
- **Current**: `class="inline-block px-3 py-1 bg-white/10 backdrop-blur-md rounded-full text-xs font-medium text-white uppercase tracking-wide mb-3"`
- **New**: `class="inline-block px-2.5 py-0.5 bg-white/10 backdrop-blur-md rounded-full text-xs font-medium text-white uppercase tracking-wide mb-2"`
- **Rationale**: Reduce `py-1` to `py-0.5` and `px-3` to `px-2.5` for a tighter badge consistent with the badge utilities in globals.css (which use `px-2 py-0.5`). Reduce `mb-3` to `mb-2`.

#### Welcome Heading (Line 58)
- **Current**: `<h3 class="text-xl sm:text-2xl font-black text-white leading-tight">Welcome back,<br/>{{ authStore.user?.username || 'Commander' }}</h3>`
- **New**: `<h3 class="text-lg sm:text-xl font-bold text-white leading-tight">Welcome back, {{ authStore.user?.username || 'Commander' }}</h3>`
- **Rationale**: CRITICAL FIX. The `<br/>` forces a line break that creates unnecessary multiline text (user's primary complaint). Remove it so the greeting flows naturally on one line. Reduce `text-xl sm:text-2xl` to `text-lg sm:text-xl` to be proportionate with the reduced banner size. Change `font-black` to `font-bold` for consistency with other headings.

#### Description Text (Line 59)
- **Current**: `class="text-white/70 mt-2 text-sm max-w-xl leading-relaxed"`
- **New**: `class="text-white/60 mt-1.5 text-xs sm:text-sm max-w-xl leading-relaxed"`
- **Rationale**: Reduce `mt-2` to `mt-1.5` for tighter spacing. Add responsive `text-xs sm:text-sm` for mobile. Reduce opacity from `/70` to `/60` for subtler secondary text.

#### Action Buttons Container (Line 61)
- **Current**: `class="mt-4 flex flex-wrap gap-2 sm:gap-3"`
- **New**: `class="mt-3 flex flex-wrap gap-2"`
- **Rationale**: Reduce `mt-4` to `mt-3`. Simplify gap to just `gap-2` (the `sm:gap-3` creates too much spacing between buttons on medium screens).

### Section: Stats Grid (Lines 69-86)

#### Grid Container (Line 69)
- **Current**: `class="grid grid-cols-2 lg:grid-cols-4 gap-3 md:gap-4"`
- **New**: `class="grid grid-cols-2 lg:grid-cols-4 gap-2 sm:gap-3"`
- **Rationale**: Reduce gap from `gap-3 md:gap-4` to `gap-2 sm:gap-3` — more compact grid matching the tighter responsive spacing pattern.

#### Stat Card (Lines 70-72)
- **Current**: `class="glass-panel rounded-xl p-4 transition-all duration-300 hover:shadow-xl hover:shadow-makoclaw-accent/10 hover:-translate-y-1 group animate-fade-in-up"`
- **New**: `class="glass-panel rounded-xl p-3 sm:p-4 transition-all duration-300 hover:shadow-lg hover:shadow-makoclaw-accent/10 hover:-translate-y-0.5 group animate-fade-in-up"`
- **Rationale**: Add responsive padding `p-3 sm:p-4` (was static `p-4`). Reduce hover shadow from `hover:shadow-xl` to `hover:shadow-lg`. Reduce hover lift from `hover:-translate-y-1` (4px) to `hover:-translate-y-0.5` (2px) — more subtle, matching Chat/Tasks hover patterns.

#### Icon Container (Line 74)
- **Current**: `class="p-2.5 rounded-xl bg-gradient-to-br transition-colors duration-300"`
- **New**: `class="p-2 rounded-lg bg-gradient-to-br transition-colors duration-300"`
- **Rationale**: Reduce `p-2.5` to `p-2` for a tighter icon container. Change `rounded-xl` to `rounded-lg` — icon containers in the launchpad (line 139) use `rounded-lg`, keep consistent.

#### Stat Value (Line 83)
- **Current**: `class="text-2xl font-bold mt-1 bg-gradient-to-br from-makoclaw-text to-makoclaw-text-secondary bg-clip-text text-transparent"`
- **New**: `class="text-xl sm:text-2xl font-bold mt-1 text-makoclaw-text"`
- **Rationale**: Add responsive sizing `text-xl sm:text-2xl` so stat values don't overflow on small screens. Remove the gradient text effect — it makes numbers harder to read and is inconsistent with the rest of the app. Use plain `text-makoclaw-text`.

### Section: Charts — Model Intelligence & Operations Status (Lines 92-115)

#### Chart Panel Padding (Lines 94, 105)
- **Current**: `class="glass-panel rounded-2xl p-3 sm:p-4 md:p-5 flex flex-col animate-fade-in-up"`
- **New**: `class="glass-panel rounded-2xl p-3 sm:p-4 flex flex-col animate-fade-in-up"`
- **Rationale**: Remove `md:p-5` — the 20px padding at md breakpoint is more than other panels which stop at `p-4`. Standardize to `p-3 sm:p-4`.

#### Chart Section Headers (Lines 96, 107)
- **Current**: `class="text-xs font-medium uppercase tracking-wide text-makoclaw-text-secondary/70"`
- **New**: No change needed — this matches the established section header pattern used throughout the page.

### Section: Live System Metrics (Lines 118-127)

#### Metrics Panel (Line 118)
- **Current**: `class="glass-panel rounded-2xl p-3 sm:p-4 md:p-5 animate-fade-in-up"`
- **New**: `class="glass-panel rounded-2xl p-3 sm:p-4 animate-fade-in-up"`
- **Rationale**: Same as charts — remove `md:p-5`, standardize to `p-3 sm:p-4`.

#### Metric Value (Line 123)
- **Current**: `class="text-xl font-bold text-makoclaw-text"`
- **New**: `class="text-lg sm:text-xl font-bold text-makoclaw-text"`
- **Rationale**: Add responsive sizing for mobile.

#### Metric Underline Bar (Line 124)
- **Current**: `class="h-1 w-8 bg-makoclaw-accent/20 rounded-full mt-2 group-hover:w-full transition-all duration-500"`
- **New**: `class="h-0.5 w-6 bg-makoclaw-accent/20 rounded-full mt-1.5 group-hover:w-full transition-all duration-500"`
- **Rationale**: The `h-1` (4px) bar is visually heavy. Reduce to `h-0.5` (2px). Reduce initial width from `w-8` to `w-6`. Reduce `mt-2` to `mt-1.5`.

### Section: Action Launchpad (Lines 132-145)

#### Launchpad Container (Line 133)
- **Current**: `class="glass-panel rounded-2xl p-3 sm:p-4 animate-fade-in-up"`
- **New**: No change — already correct.

#### Launchpad Action Items (Lines 136-143)
- **Current**: `class="flex flex-col items-center justify-center p-3 rounded-xl border border-makoclaw-border/50 hover:border-makoclaw-accent/40 bg-makoclaw-surface/30 hover:bg-makoclaw-accent/5 transition-all group active:scale-95"`
- **New**: `class="flex flex-col items-center justify-center p-3 rounded-xl border border-makoclaw-border/20 hover:border-makoclaw-accent/30 bg-makoclaw-surface/30 hover:bg-makoclaw-accent/5 transition-all group active:scale-95"`
- **Rationale**: Border opacity `/50` is too prominent — normalize to `/20` resting, `/30` hover.

#### Launchpad Icon Container (Line 139)
- **Current**: `class="w-9 h-9 rounded-lg flex items-center justify-center mb-2 transition-transform group-hover:scale-105 group-hover:rotate-3"`
- **New**: `class="w-8 h-8 rounded-lg flex items-center justify-center mb-1.5 transition-transform group-hover:scale-105"`
- **Rationale**: Reduce icon container from `w-9 h-9` (36px) to `w-8 h-8` (32px) — more proportionate in the 2-column grid. Reduce `mb-2` to `mb-1.5`. Remove `group-hover:rotate-3` — the rotation effect is gimmicky and not used elsewhere.

#### Launchpad Icon (Line 140)
- **Current**: `class="w-5 h-5 text-white"`
- **New**: `class="w-4 h-4 text-white"`
- **Rationale**: Inside a 32px container, a 20px icon is too large (62% fill). A 16px icon (50% fill) is more balanced.

### Section: Recent Activity Feed (Lines 148-177)

#### Feed Container (Line 148)
- **Current**: `class="glass-panel rounded-2xl p-3 sm:p-4 animate-fade-in-up flex flex-col h-[420px]"`
- **New**: `class="glass-panel rounded-2xl p-3 sm:p-4 animate-fade-in-up flex flex-col h-[380px]"`
- **Rationale**: Reduce fixed height from 420px to 380px. At 420px on most laptop screens (768-900px height), the feed extends below the fold unnecessarily while the launchpad above has whitespace.

#### Empty State Icon Container (Line 156)
- **Current**: `class="w-10 h-10 rounded-full border-2 border-dashed border-makoclaw-border flex items-center justify-center mb-3"`
- **New**: `class="w-10 h-10 rounded-full border-2 border-dashed border-makoclaw-border/40 flex items-center justify-center mb-3"`
- **Rationale**: Add `/40` opacity to the dashed border — full opacity borders are inconsistent with the glass morphism style.

#### Activity Item (Lines 162-163)
- **Current**: `class="flex items-center gap-3 p-2.5 sm:p-3 rounded-xl bg-makoclaw-surface/40 border border-makoclaw-border/20 hover:border-makoclaw-accent/30 transition-all group cursor-pointer"`
- **New**: No change — this already follows the correct pattern with `/20` border and `/30` hover.

#### Activity Icon Container (Line 164)
- **Current**: `class="w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0 transition-colors"`
- **New**: `class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 transition-colors"`
- **Rationale**: Reduce from `w-9 h-9` (36px) to `w-8 h-8` (32px) to match the launchpad icon container size and be proportionate with the activity row height.

#### Activity Icon (Line 165)
- **Current**: `class="w-5 h-5 text-white"`
- **New**: `class="w-4 h-4 text-white"`
- **Rationale**: Match launchpad icon sizing — 16px inside a 32px container.

### Section: Script — Launchpad Actions Definition (Lines 249-254)

#### Shadow Classes on Launchpad Colors
- **Current**: Each action has `shadow-lg shadow-{color}/20` (e.g., `'bg-makoclaw-accent shadow-lg shadow-makoclaw-accent/20'`)
- **New**: Remove shadow from individual action colors — use just the base color (e.g., `'bg-makoclaw-accent'`)
- **Rationale**: The individual colored shadows on 32px containers create visual noise. The parent hover effect provides sufficient feedback.

Specific changes:
- Line 250: `'bg-makoclaw-accent shadow-lg shadow-makoclaw-accent/20'` → `'bg-makoclaw-accent'`
- Line 251: `'bg-indigo-500 shadow-lg shadow-indigo-500/20'` → `'bg-indigo-500'`
- Line 252: `'bg-cyan-500 shadow-lg shadow-cyan-500/20'` → `'bg-cyan-500'`
- Line 253: `'bg-blue-600 shadow-lg shadow-blue-600/20'` → `'bg-blue-600'`

## Summary of All Issues (27 total)

| # | Category | Location | Issue |
|---|----------|----------|-------|
| 1 | Performance | Background blob | `animate-pulse` on 384px element causes constant GPU repaints |
| 2 | Proportion | Header title | `text-xl sm:text-2xl font-black` with gradient is disproportionately flashy |
| 3 | Spacing | Header subtitle | `mt-1` and `gap-2` too loose for header grouping |
| 4 | Consistency | Refresh button border | `border-makoclaw-border/50` should be `/20` |
| 5 | Consistency | Refresh button bg | `bg-makoclaw-surface` should be `/40` for glass effect |
| 6 | Consistency | Refresh button scale | `active:scale-90` should be `active:scale-95` |
| 7 | Spacing | Welcome banner padding | `p-4 sm:p-5 md:p-6` too spacious — reduce by one step |
| 8 | Proportion | Welcome banner shadow | `shadow-2xl` should be `shadow-xl` |
| 9 | Proportion | Decorative bolt icon | `w-32 h-32` (128px) disproportionate — reduce to `w-20 h-20` |
| 10 | Spacing | Decorative icon padding | `p-8` excessive — reduce to `p-6` |
| 11 | Proportion | Command Center badge | `py-1 px-3 mb-3` too large — reduce to `py-0.5 px-2.5 mb-2` |
| 12 | **Layout** | **Welcome heading `<br/>`** | **Forces unnecessary multiline — REMOVE** |
| 13 | Proportion | Welcome heading size | `text-xl sm:text-2xl font-black` too large — reduce one step |
| 14 | Spacing | Description margin | `mt-2` should be `mt-1.5` |
| 15 | Responsive | Description text | Missing `text-xs sm:text-sm` responsive size |
| 16 | Spacing | Action buttons margin | `mt-4` should be `mt-3` |
| 17 | Spacing | Stats grid gap | `gap-3 md:gap-4` should be `gap-2 sm:gap-3` |
| 18 | Responsive | Stat card padding | Static `p-4` should be `p-3 sm:p-4` |
| 19 | Proportion | Stat card hover | `hover:shadow-xl hover:-translate-y-1` too aggressive |
| 20 | Proportion | Stat icon container | `p-2.5 rounded-xl` slightly oversized |
| 21 | Readability | Stat value | Gradient text on numbers reduces readability |
| 22 | Responsive | Stat value size | Static `text-2xl` should be `text-xl sm:text-2xl` |
| 23 | Consistency | Chart panel padding | `md:p-5` exceeds standard — remove |
| 24 | Responsive | Metric value | Static `text-xl` should be `text-lg sm:text-xl` |
| 25 | Proportion | Metric underline | `h-1 w-8 mt-2` too heavy — reduce |
| 26 | Consistency | Launchpad border | `/50` should be `/20` |
| 27 | Proportion | Launchpad icon | `w-9 h-9` with `w-5 h-5` icon — reduce both |

## Testing Strategy

### Visual Inspection Breakpoints

| Breakpoint | Width | What to Verify |
|-----------|-------|----------------|
| Mobile S | 390px | All text fits single lines, no overflow, touch targets >= 36px, stats grid 2-col readable |
| Tablet | 768px | Welcome banner heading on one line, stats grid still 2-col, charts side-by-side |
| Desktop | 1024px | Stats grid 4-col, main grid 3-col layout (2+1), no excessive whitespace |
| Wide | 1440px | Content doesn't stretch disproportionately, max-widths respected |

### Functional Verification

- [ ] Reload button triggers data refresh (animation still works without `animate-pulse`)
- [ ] All router-links in launchpad and banner navigate correctly
- [ ] Charts render correctly within updated padding constraints
- [ ] Activity feed scrolls correctly in reduced 380px height
- [ ] Stagger entrance animations still fire with correct delays

### Cross-Browser

- Chrome 120+, Firefox 120+, Safari 17+ (backdrop-blur support)
- Light and dark mode both tested (CSS variables handle theming)

## Open Questions

1. **Welcome banner text content**: The descriptive text "Your AI fleet is standing by. All systems are nominal across 10+ channels. What are we building today?" is thematic flavor text. Should this be shortened or made more utilitarian? (Design opinion: keep it — it's personality, not a UX issue.)

2. **Stat labels**: Labels like "Intelligence Fleet", "Active Missions", "Neural Links", "Total Cognition" are thematic but not immediately clear. This is a content/copy concern, not a UX/design concern — out of scope for this change.

3. **Fixed height on activity feed**: The `h-[380px]` fixed height works for most screens but could be improved with `max-h-[380px]` + flex grow to be more adaptive. Consider for a follow-up if the sidebar needs to fill available space dynamically.
