# MakoClaw Frontend Design System

This document defines the visual design language for the MakoClaw Vue 3 + Tailwind CSS frontend. Follow these guidelines to maintain consistency across all screens.

## Quick Start

When creating or updating a Vue component:

1. **Use the skill**: Invoke `/frontend-design` in Claude Code sessions
2. **Reference files**:
   - CSS Variables: `pkg/web/frontend/src/styles/globals.css`
   - Tailwind Config: `pkg/web/frontend/tailwind.config.js`

---

## Design Principles

| Principle | Implementation |
|-----------|---------------|
| **Glass Morphism** | `backdrop-blur-xl` with semi-transparent backgrounds |
| **Gradient Accents** | Dual-gradient backgrounds at opposite corners |
| **Subtle Animations** | 300ms transitions, micro-interactions |
| **Theme Adaptive** | CSS variables that work in dark/light mode |
| **Mobile-First** | Touch targets ≥ 40px, fluid layouts |

---

## Color Tokens

All colors use `makoclaw-*` Tailwind classes backed by CSS variables:

```
makoclaw-bg              → Page background
makoclaw-surface         → Cards, panels
makoclaw-surface-hover   → Hover states
makoclaw-border          → Borders
makoclaw-accent          → Primary actions (blue)
makoclaw-accent-hover    → Primary hover
makoclaw-success         → Success (emerald)
makoclaw-warning         → Warning (amber)
makoclaw-error           → Error (red)
makoclaw-text            → Primary text
makoclaw-text-secondary  → Secondary text
```

### Screen-Specific Gradients

Each screen has a unique gradient identity:

| Screen | Primary | Secondary |
|--------|---------|-----------|
| **Dashboard** | `makoclaw-accent/30` | `indigo-500/20` |
| **Chat** | `makoclaw-accent/40` | `purple-500/30` |
| **Tasks** | `blue-500/30` | `emerald-500/20` |
| **Skills** | `purple-500/30` | `pink-500/20` |
| **Cron** | `cyan-500/30` | `blue-500/20` |
| **Files** | `indigo-500/30` | `violet-500/20` |
| **Knowledge** | `teal-500/30` | `emerald-500/20` |
| **MCP** | `orange-500/30` | `amber-500/20` |
| **Workflows** | `rose-500/30` | `fuchsia-500/20` |
| **Agents** | `lime-500/30` | `green-500/20` |
| **Settings** | `makoclaw-accent/30` | `blue-500/20` |

---

## Page Structure

### Background Gradient Mesh

Every page starts with this structure:

```vue
<template>
  <div class="flex flex-col h-full bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-{primary}/30 via-transparent to-transparent"></div>
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-{secondary}/20 via-transparent to-transparent"></div>
    </div>

    <!-- Page Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <!-- ... -->
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-4 sm:p-6">
      <!-- ... -->
    </div>
  </div>
</template>
```

### Page Header

```vue
<div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
  <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
    <div class="flex items-center gap-3">
      <!-- Icon Container -->
      <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-{primary}/20 to-{secondary}/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-{primary}/10">
        <svg class="w-5 h-5 sm:w-6 sm:h-6 text-{primary}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <!-- icon path -->
        </svg>
      </div>

      <!-- Title -->
      <div class="flex-1 min-w-0">
        <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-{accent} bg-clip-text text-transparent">
          Page Title
        </h1>
        <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5">Subtitle description</p>
      </div>

      <!-- Primary Action Button -->
      <button class="px-4 sm:px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-{primary} to-{primary-600} hover:from-{primary-600} hover:to-{primary-700} text-white rounded-xl transition-all shadow-lg shadow-{primary}/25 hover:shadow-{primary}/40 text-sm font-bold flex items-center gap-2 active:scale-95 flex-shrink-0">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        <span class="hidden sm:inline">New Item</span>
      </button>
    </div>
  </div>
</div>
```

---

## Component Library

### Glass Panel

```html
<div class="glass-panel p-6 rounded-2xl">
  <!-- Content -->
</div>

<!-- Equivalent to: -->
<div class="bg-makoclaw-surface/30 backdrop-blur-2xl border border-white/5 shadow-2xl ring-1 ring-white/10 rounded-2xl p-6">
```

### Cards

```html
<!-- Static card -->
<div class="card p-4">...</div>

<!-- Interactive card -->
<div class="card-interactive p-4">...</div>
```

### Buttons

```html
<!-- Primary (gradient) -->
<button class="px-4 py-2.5 bg-gradient-to-r from-makoclaw-accent to-blue-600 hover:from-blue-600 hover:to-blue-700 text-white rounded-xl shadow-lg shadow-makoclaw-accent/25 font-bold active:scale-95">
  Primary Action
</button>

<!-- Standard primary -->
<button class="btn-primary">Action</button>

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

### Modals/Popovers

```html
<div class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl shadow-2xl overflow-hidden ring-1 ring-white/10 animate-scaleIn">
  <!-- Header -->
  <div class="p-4 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent">
    <div class="flex items-center gap-2">
      <div class="w-7 h-7 rounded-lg bg-gradient-to-br from-{primary}/20 to-{secondary}/20 flex items-center justify-center">
        <svg class="w-3.5 h-3.5 text-{primary}" ...></svg>
      </div>
      <h3 class="text-sm font-bold text-makoclaw-text">Modal Title</h3>
    </div>
  </div>
  <!-- Content -->
  <div class="p-4">...</div>
</div>
```

### Empty States

```html
<div class="flex flex-col items-center justify-center py-12 text-center">
  <div class="relative">
    <!-- Glow effect -->
    <div class="absolute inset-0 bg-gradient-to-br from-{primary}/30 to-{secondary}/30 rounded-3xl blur-2xl opacity-50"></div>
    <div class="relative glass-panel p-8 rounded-2xl shadow-2xl ring-1 ring-white/10">
      <div class="w-16 h-16 mx-auto rounded-2xl bg-gradient-to-br from-{primary}/20 to-{secondary}/20 flex items-center justify-center ring-1 ring-white/20">
        <svg class="w-8 h-8 text-{primary}" ...></svg>
      </div>
    </div>
  </div>
  <h3 class="text-lg font-bold text-makoclaw-text mt-6">Nothing here yet</h3>
  <p class="text-sm text-makoclaw-text-secondary/70 mt-2 max-w-xs">Description text</p>
</div>
```

### Selected List Item

```html
<button
  :class="[
    'w-full text-left px-2.5 py-2 rounded-lg text-sm transition-all duration-200',
    isSelected
      ? 'bg-gradient-to-r from-makoclaw-accent/15 to-transparent text-makoclaw-text border-l-2 border-makoclaw-accent'
      : 'hover:bg-makoclaw-surface/50 text-makoclaw-text-secondary hover:text-makoclaw-text border-l-2 border-transparent'
  ]"
>
```

---

## Z-Index Layers

| Layer | z-index | Purpose |
|-------|---------|---------|
| `z-sticky` | 20 | Sticky headers |
| `z-overlay-backdrop` | 40 | Overlay backdrops |
| `z-sidebar` | 45 | Mobile sidebar |
| `z-dropdown` | 50 | Dropdown menus |
| `z-modal` | 60 | Modal dialogs |
| `z-modal-nested` | 70 | Nested modals |
| `z-toast` | 80 | Toast notifications |

---

## Animations

### Tailwind Animations

```
animate-fadeIn       → Fade in
animate-slideUp      → Slide up + fade
animate-scaleIn      → Scale from 95%
animate-fadeSlide    → Fade + translateY
animate-skeleton     → Loading shimmer
animate-subtlePulse  → Gentle pulse
```

### Vue Transitions

```html
<Transition name="page">...</Transition>    <!-- Route transitions -->
<Transition name="fade">...</Transition>    <!-- Simple fade -->
<Transition name="modal">...</Transition>   <!-- Modal scale + fade -->
<Transition name="slide">...</Transition>   <!-- Sidebar slide -->
<Transition name="list">...</Transition>    <!-- List items -->
<Transition name="expand">...</Transition>  <!-- Collapsible sections -->
```

---

## Responsive Design

### Breakpoints

| Prefix | Min-width | Target |
|--------|-----------|--------|
| (none) | 0px | Mobile phones |
| `sm:` | 640px | Large phones |
| `md:` | 768px | Tablets |
| `lg:` | 1024px | Desktop |

### Touch Targets

Always ensure interactive elements have minimum 40px dimensions:

```html
<button class="p-2.5 min-h-[40px] min-w-[40px]">
```

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

## Example Screens

Study these files for reference implementations:

| Screen | File | Notable Patterns |
|--------|------|-----------------|
| Skills | `SkillsView.vue` | Tabs with counts, marketplace cards, rating widget, bundles |
| Cron | `CronView.vue` | Job cards with type icons, schedule badges, day selector |
| Chat | `ChatView.vue` | Sidebar with sessions, message bubbles, input area, tools popover |
| Tasks | `TasksView.vue` | Kanban header, search with filters, status-based styling |

---

## Common Class Combinations

```css
/* Glass backgrounds */
bg-makoclaw-surface/40 backdrop-blur-2xl

/* Subtle borders */
border border-makoclaw-border/50 ring-1 ring-white/10

/* Gradient icon container */
bg-gradient-to-br from-{color}/20 to-{color2}/20 ring-1 ring-white/10

/* Colored shadow */
shadow-lg shadow-{color}-500/25

/* Focus states */
focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50

/* Hover lift */
hover:shadow-md hover:-translate-y-[1px]

/* Button press */
active:scale-95
```
