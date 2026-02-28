---
name: frontend-design
description: >
  MakoClaw Frontend Design System guide. Use when creating or updating Vue components
  to maintain visual consistency across all screens. Covers glass morphism, gradients,
  color tokens, spacing, and component patterns.
license: MIT
metadata:
  author: makoclaw-team
  version: "1.0"
---

## Purpose

You are following the **MakoClaw Frontend Design System** — a modern, glass-morphism-based visual language for the Vue 3 + Tailwind frontend. This skill ensures visual consistency across all screens.

## Design Principles

1. **Glass Morphism** — Surfaces use `backdrop-blur-xl` with semi-transparent backgrounds
2. **Gradient Accents** — Dual-gradient backgrounds give depth without overwhelming
3. **Subtle Animations** — Smooth transitions (300ms ease) and micro-interactions
4. **Dark/Light Mode Ready** — All colors use CSS variables that adapt to theme
5. **Mobile-First Responsive** — Touch targets ≥ 40px, fluid layouts

---

## Color System

All colors use the `makoclaw-*` Tailwind tokens defined via CSS variables:

| Token | Purpose | Light Mode | Dark Mode |
|-------|---------|------------|-----------|
| `makoclaw-bg` | Page background | Slate 50 | Slate 900 |
| `makoclaw-surface` | Cards, panels | White | Slate 800 |
| `makoclaw-surface-hover` | Hover state | Slate 100 | Slate 700 |
| `makoclaw-border` | Borders | Slate 200 | Slate 700 |
| `makoclaw-accent` | Primary actions | Blue 500 | Blue 500 |
| `makoclaw-accent-hover` | Primary hover | Blue 600 | Blue 600 |
| `makoclaw-success` | Success states | Emerald 500 | Emerald 500 |
| `makoclaw-warning` | Warning states | Amber 500 | Amber 500 |
| `makoclaw-error` | Error states | Red 500 | Red 500 |
| `makoclaw-text` | Primary text | Slate 900 | Slate 50 |
| `makoclaw-text-secondary` | Secondary text | Slate 500 | Slate 400 |

### Per-Screen Accent Colors

Each major screen has a unique gradient identity while using the shared color system:

| Screen | Primary Gradient | Secondary Gradient |
|--------|-----------------|-------------------|
| Dashboard | `makoclaw-accent/30` | `indigo-500/20` |
| Chat | `makoclaw-accent/40` (blue) | `purple-500/30` |
| Tasks | `blue-500/30` | `emerald-500/20` |
| Skills | `purple-500/30` | `pink-500/20` |
| Cron | `cyan-500/30` | `blue-500/20` |
| Files | `indigo-500/30` | `violet-500/20` |
| Knowledge | `teal-500/30` | `emerald-500/20` |
| MCP | `orange-500/30` | `amber-500/20` |
| Workflows | `rose-500/30` | `fuchsia-500/20` |
| Agents | `lime-500/30` | `green-500/20` |
| Settings | `makoclaw-accent/30` | `blue-500/20` |

---

## Background Patterns

### Dual-Gradient Mesh

Every page uses a subtle dual-gradient background:

```html
<!-- Background Gradient Mesh -->
<div class="absolute inset-0 pointer-events-none">
  <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-{primary}/30 via-transparent to-transparent"></div>
  <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-{secondary}/20 via-transparent to-transparent"></div>
</div>
```

- Use `opacity-20` to `opacity-30` for subtlety
- Position gradients at opposite corners for depth
- Primary gradient slightly stronger than secondary

---

## Page Header Pattern

Every screen should have a consistent header structure:

```html
<div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
  <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
    <div class="flex items-center gap-3">
      <!-- Icon Container -->
      <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-{primary}/20 to-{secondary}/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-{primary}/10">
        <svg class="w-5 h-5 sm:w-6 sm:h-6 text-{primary}" ...>
      </div>

      <!-- Title -->
      <div class="flex-1 min-w-0">
        <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-{accent} bg-clip-text text-transparent">
          Page Title
        </h1>
        <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5">Subtitle</p>
      </div>

      <!-- Primary Action -->
      <button class="px-4 sm:px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-{primary} to-{primary-dark} hover:from-{primary-dark} hover:to-{primary-darker} text-white rounded-xl transition-all shadow-lg shadow-{primary}/25 text-sm font-bold flex items-center gap-2 active:scale-95">
        <svg class="w-4 h-4" ...></svg>
        <span class="hidden sm:inline">Action</span>
      </button>
    </div>
  </div>
</div>
```

---

## Component Patterns

### Glass Panel

```html
<div class="glass-panel p-6 rounded-2xl">
  <!-- Content -->
</div>

<!-- Or manually -->
<div class="bg-makoclaw-surface/30 backdrop-blur-2xl border border-white/5 shadow-2xl ring-1 ring-white/10 rounded-2xl p-6">
  <!-- Content -->
</div>
```

### Cards

```html
<!-- Static card -->
<div class="card p-4">
  <!-- bg-makoclaw-surface border border-makoclaw-border rounded-xl -->
</div>

<!-- Interactive card with hover -->
<div class="card-interactive p-4">
  <!-- hover:shadow-md hover:-translate-y-[1px] hover:border-makoclaw-accent/20 -->
</div>
```

### Buttons

```html
<!-- Primary -->
<button class="btn-primary">Action</button>
<!-- px-4 py-2 bg-makoclaw-accent text-white rounded-lg shadow-sm -->

<!-- Gradient Primary (for prominent actions) -->
<button class="px-4 py-2.5 bg-gradient-to-r from-{color}-500 to-{color}-600 hover:from-{color}-600 hover:to-{color}-700 text-white rounded-xl shadow-lg shadow-{color}-500/25 font-bold active:scale-95">
  Action
</button>

<!-- Secondary -->
<button class="btn-secondary">Cancel</button>

<!-- Ghost -->
<button class="btn-ghost">More</button>

<!-- Danger -->
<button class="btn-danger">Delete</button>
```

### Input Fields

```html
<input
  class="w-full pl-10 pr-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all text-sm backdrop-blur-sm min-h-[40px]"
  placeholder="Search..."
>
```

### Badges

```html
<span class="badge-success">Active</span>
<span class="badge-warning">Pending</span>
<span class="badge-error">Failed</span>
<span class="badge-info">New</span>
<span class="badge-neutral">Draft</span>
```

### Empty States

```html
<div class="flex flex-col items-center justify-center py-12 text-center">
  <div class="relative">
    <!-- Optional glow -->
    <div class="absolute inset-0 bg-gradient-to-br from-{primary}/30 to-{secondary}/30 rounded-3xl blur-2xl opacity-50"></div>
    <div class="relative glass-panel p-8 rounded-2xl shadow-2xl ring-1 ring-white/10">
      <div class="w-16 h-16 mx-auto rounded-2xl bg-gradient-to-br from-{primary}/20 to-{secondary}/20 flex items-center justify-center ring-1 ring-white/20">
        <svg class="w-8 h-8 text-{primary}" ...></svg>
      </div>
    </div>
  </div>
  <h3 class="text-lg font-bold text-makoclaw-text mt-6">Title</h3>
  <p class="text-sm text-makoclaw-text-secondary/70 mt-2 max-w-xs">Description</p>
</div>
```

### Modals/Popovers

```html
<div class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl shadow-2xl overflow-hidden ring-1 ring-white/10 animate-scaleIn">
  <!-- Header -->
  <div class="p-4 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent">
    <div class="flex items-center gap-2">
      <div class="w-7 h-7 rounded-lg bg-gradient-to-br from-{primary}/20 to-{secondary}/20 flex items-center justify-center">
        <svg class="w-3.5 h-3.5 text-{primary}" ...></svg>
      </div>
      <h3 class="text-sm font-bold text-makoclaw-text">Title</h3>
    </div>
  </div>
  <!-- Content -->
  <div class="p-4">
    <!-- ... -->
  </div>
</div>
```

### List Items with Selection

```html
<button
  :class="[
    'w-full text-left px-2.5 py-2 rounded-lg text-sm transition-all duration-200',
    isSelected
      ? 'bg-gradient-to-r from-makoclaw-accent/15 to-transparent text-makoclaw-text border-l-2 border-makoclaw-accent'
      : 'hover:bg-makoclaw-surface/50 text-makoclaw-text-secondary hover:text-makoclaw-text border-l-2 border-transparent'
  ]"
>
  <!-- Content -->
</button>
```

---

## Spacing Scale

Use Tailwind's default spacing with these guidelines:

| Context | Padding | Gap |
|---------|---------|-----|
| Page container | `p-4 sm:p-6` | — |
| Card internal | `p-4` or `p-5` | `gap-3` or `gap-4` |
| Compact toolbar | `p-2` or `p-3` | `gap-2` |
| List items | `px-3 py-2` | `space-y-1` |
| Modal header | `p-4` | `gap-2` |

---

## Animation Classes

Available in `globals.css` and `tailwind.config.js`:

| Class | Effect |
|-------|--------|
| `animate-fadeIn` | Fade in from transparent |
| `animate-slideUp` | Fade + slide up |
| `animate-scaleIn` | Scale from 95% to 100% |
| `animate-fadeSlide` | Fade + translateY |
| `animate-skeleton` | Loading shimmer |
| `animate-subtlePulse` | Gentle pulse for indicators |

Vue transitions: `page-*`, `fade-*`, `modal-*`, `slide-*`, `list-*`, `expand-*`

---

## Z-Index Layers

Defined in `tailwind.config.js`:

| Layer | z-index | Use |
|-------|---------|-----|
| `z-sticky` | 20 | Sticky headers |
| `z-overlay-backdrop` | 40 | Overlay backdrops |
| `z-sidebar` | 45 | Mobile sidebar |
| `z-dropdown` | 50 | Dropdown menus |
| `z-modal` | 60 | Modal dialogs |
| `z-modal-nested` | 70 | Nested modals |
| `z-toast` | 80 | Toast notifications |

---

## Responsive Breakpoints

Use Tailwind's mobile-first breakpoints:

| Breakpoint | Min-width | Target |
|------------|-----------|--------|
| (none) | 0px | Mobile phones |
| `sm:` | 640px | Large phones / small tablets |
| `md:` | 768px | Tablets |
| `lg:` | 1024px | Desktop |

Ensure touch targets are at least 40px (`min-h-[40px] min-w-[40px]`).

---

## Checklist for New Screens

- [ ] Add dual-gradient background mesh with screen-specific colors
- [ ] Create page header with icon container, gradient title, and primary action
- [ ] Use `glass-sticky` for sticky headers
- [ ] Apply `custom-scrollbar` to scrollable areas
- [ ] Use `card-interactive` for clickable cards
- [ ] Ensure proper empty states with glow effect
- [ ] Test in both dark and light mode
- [ ] Verify mobile responsiveness (touch targets, spacing)
- [ ] Use appropriate z-index layers for overlays

---

## File References

- **CSS Variables**: `pkg/web/frontend/src/styles/globals.css`
- **Tailwind Config**: `pkg/web/frontend/tailwind.config.js`
- **Example Screens**:
  - `SkillsView.vue` — Marketplace patterns, tabs, rating widget
  - `CronView.vue` — Job cards, schedule badges, day selector
  - `ChatView.vue` — Sidebar, messages, input area, popovers
  - `TasksView.vue` — Kanban header, search/filters

---

## Quick Reference: Common Classes

```
# Backgrounds
bg-makoclaw-bg/40                    # Semi-transparent page bg
bg-makoclaw-surface/30               # Glass panel bg
bg-gradient-to-br from-X/20 to-Y/20  # Gradient accent

# Borders
border border-makoclaw-border/50     # Subtle border
ring-1 ring-white/10                 # Glass ring effect

# Blur
backdrop-blur-xl                     # Standard blur
backdrop-blur-2xl                    # Heavy blur for modals

# Shadows
shadow-lg shadow-{color}-500/25      # Colored glow shadow
shadow-2xl                           # Deep shadow for elevated elements

# Transitions
transition-all duration-200          # Standard transition
active:scale-95                      # Button press feedback
```
