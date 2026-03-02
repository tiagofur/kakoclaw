# Design: Dashboard & Settings UI Polish

**Date:** 2026-02-26
**Branch:** ui/inner-app-ux-polish
**Approach:** Option A — Surgical Proportional Scaling

## Problem

The Dashboard and Settings screens feel overwhelming and disproportionate. The same 5 anti-patterns appear across all 9 files:

1. **Icon containers too large** — `w-14 h-14` standalone icons and `p-3 + w-6` gradient icon containers
2. **Excessive padding** — cards `p-8 md:p-10`, inputs `py-4 px-6`, buttons `py-4`
3. **No typographic hierarchy** — `font-black + uppercase + tracking-[0.2em]` on every label, title, and button
4. **Extreme border radii** — `rounded-[2rem]`, `rounded-3xl` everywhere
5. **Too much spacing** — `mb-8`, `mb-10`, `space-y-8`, `gap-8` throughout

## Token Changes (applied consistently across all 9 files)

### Icons

| Context | Before | After |
|---|---|---|
| Section header icons (gradient bg) | `p-3 rounded-2xl` + `w-6 h-6` icon | `p-2 rounded-xl` + `w-4 h-4` icon |
| Channel/Provider standalone icons | `w-14 h-14 rounded-2xl` | `w-10 h-10 rounded-xl` |
| Stat card icons | `p-3 rounded-2xl` + `w-6 h-6` | `p-2 rounded-xl` + `w-4 h-4` |
| Activity feed icons | `w-10 h-10 rounded-xl` + `w-5 h-5` | `w-8 h-8 rounded-lg` + `w-4 h-4` |
| Launchpad action icons | `w-10 h-10 rounded-xl` + `w-5 h-5` | `w-8 h-8 rounded-lg` + `w-4 h-4` |

### Cards & Panels

| Property | Before | After |
|---|---|---|
| Card padding | `p-8 md:p-10` | `p-5 md:p-6` |
| Card border radius | `rounded-[2rem]` | `rounded-2xl` |
| Stat card border radius | `rounded-3xl` | `rounded-xl` |
| Section spacing | `space-y-8` | `space-y-5` |
| Section mb | `mb-8`, `mb-10` | `mb-5`, `mb-6` |
| Grid gap | `gap-8` | `gap-5` |

### Typography

| Context | Before | After |
|---|---|---|
| Page titles (`h1`, `h2`) | `font-black` | `font-black` (keep) |
| Section title (card header) | `text-lg font-black italic` | `text-base font-semibold` |
| Section label (overline) | `text-[11px] font-black uppercase tracking-[0.2em]` | `text-[11px] font-medium uppercase tracking-wide` |
| Form labels | `text-[10px] font-black uppercase tracking-widest` | `text-[10px] font-medium uppercase tracking-wide` |
| Stat number | `text-4xl font-black` | `text-3xl font-bold` |
| Stat label | `text-[10px] font-black uppercase tracking-[0.2em]` | `text-[10px] font-medium uppercase tracking-wide` |
| Buttons (primary) | `font-black uppercase tracking-widest` | `font-semibold` |
| Buttons (secondary) | `font-black uppercase tracking-widest` | `font-medium` |

### Inputs & Forms

| Property | Before | After |
|---|---|---|
| Input padding | `px-6 py-4`, `px-5 py-3.5` | `px-4 py-2.5` |
| Input border radius | `rounded-2xl` | `rounded-xl` |
| Input border | `border-2` | `border` |
| Select padding | `px-5 py-3.5` | `px-4 py-2.5` |

### Dashboard Welcome Banner

| Property | Before | After |
|---|---|---|
| Heading size | `text-3xl md:text-5xl` | `text-2xl md:text-3xl` |
| Padding | `p-8 md:p-12` | `p-6 md:p-8` |
| Body text | `text-sm md:text-lg` | `text-sm` |
| Button style | `px-6 py-3 uppercase font-bold tracking-wider` | `px-5 py-2 font-medium` |
| Background bolt icon | `w-64 h-64 opacity-10` | `w-48 h-48 opacity-[0.07]` |

## Files to Modify

1. `pkg/web/frontend/src/views/DashboardView.vue`
2. `pkg/web/frontend/src/views/SettingsView.vue`
3. `pkg/web/frontend/src/components/settings/AgentSettingsTab.vue`
4. `pkg/web/frontend/src/components/settings/ChannelsSettingsTab.vue`
5. `pkg/web/frontend/src/components/settings/ProvidersSettingsTab.vue`
6. `pkg/web/frontend/src/components/settings/ProfileSettingsTab.vue`
7. `pkg/web/frontend/src/components/settings/ToolPermissionsTab.vue`
8. `pkg/web/frontend/src/components/settings/AuditLogTab.vue`
9. `pkg/web/frontend/src/components/settings/SpecialistFormModal.vue`

## What to Preserve

- All gradient colors on icon backgrounds
- Accent color highlights and status badges
- Hover animations (scale, rotate-3) — just toned down where `scale-110`→`scale-105`
- The personality of copy ("Orchestrator Protocol", "Specialist Fleet", etc.)
- Background blur decorative elements (opacity reduced slightly)
- All functional behavior and data flow

## Success Criteria

- No element dominates the visual hierarchy unintentionally
- Typography has 3 clear levels: title (font-black), heading (font-semibold), label (font-medium)
- Icon containers read as accents, not primary visual elements
- Forms feel compact and professional, not spacious/bloated
- Design is still clearly "fun" and modern — not a generic enterprise UI
