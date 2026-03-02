# Chat Page UX Polish — Design Document

**Date**: 2026-02-27
**Branch**: `ui/inner-app-ux-polish`
**Approach**: Surgical Polish (modify existing files in-place)

## Goals

- Optimize space usage on all screen sizes (floor: 390px)
- Improve mobile responsiveness and touch targets
- Enhance glass morphism aesthetic with 2026 design trends
- Maintain fun, clean, modern look

## Design Decisions

- **Target floor**: 390px (iPhone 14/15 standard)
- **Input UX**: Compact icon toolbar above textarea (single row always)
- **Desktop messages**: Fluid width scaling (70-80% of available space)
- **Aesthetic**: Glass morphism + gradients (builds on existing accent gradient)

---

## Section 1: Message Bubble Fluid Width

**Current**: `max-w-[90%] sm:max-w-[85%] lg:max-w-2xl` (caps at 672px)

**New**: `max-w-[92%] sm:max-w-[88%] md:max-w-[85%] lg:max-w-3xl xl:max-w-4xl 2xl:max-w-5xl`

| Screen | Width |
|--------|-------|
| 390px mobile | 358px (92%) |
| 768px tablet | 652px (85%) |
| 1024px laptop | 768px (3xl) |
| 1280px desktop | 896px (4xl) |
| 1536px+ ultrawide | 1024px (5xl) |

**Files**: `MessageBubble.vue`

## Section 2: Input Area Redesign

Compact icon toolbar above textarea. All buttons in single row (no wrapping).

```
┌─────────────────────────────────────┐
│ [📎] [💡] [🔧]        [🎤] [Send▶] │  ← 36px min touch targets
│ ┌─────────────────────────────────┐ │
│ │ Type a message...               │ │  ← auto-grow textarea
│ └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

- Left group: file attach, prompt library, tools select
- Right group: mic, send/stop
- `justify-between` layout
- On `md+`: optional text labels beside icons
- Glass container with `backdrop-blur-lg` and subtle top glow

**Files**: `ChatView.vue` (input section)

## Section 3: Top Bar Simplification

Fixed single-row layout (remove `flex-wrap` and `order-` reflow):

```
[☰ sidebar] [Agent Status...] ────── [Model ▾]
```

- Sidebar toggle: fixed left
- Agent status: `flex-1`, truncate with ellipsis
- Model selector: fixed right, `max-w-[140px] md:max-w-[240px]`
- No wrapping — always single row

**Files**: `ChatView.vue` (top bar section)

## Section 4: Glass Morphism Enhancement

- **Assistant bubbles**: `backdrop-blur-xl` + gradient border (white/10 → white/5)
- **Input area**: Glass container, `backdrop-blur-lg`, subtle top border glow
- **Chat history sidebar**: Glass panel with blur (replace solid bg)
- **Specialists panel**: Glass cards per specialist
- **Shadow system**: `shadow-lg shadow-black/5` light / `shadow-black/20` dark

**Files**: `MessageBubble.vue`, `ChatView.vue`, `SpecialistsPanel.vue`

## Section 5: Spacing Normalization

Consistent responsive scale:
- Container padding: `p-2 sm:p-3 md:p-4`
- Gaps: `gap-1.5 sm:gap-2 md:gap-3`
- Message spacing: `space-y-3 md:space-y-4`
- Touch targets: `min-h-[44px] min-w-[44px]` on mobile interactive elements

**Files**: `ChatView.vue`, `MessageBubble.vue`

## Section 6: Empty State Polish

- Gradient accent icon with animated pulse
- Bolder greeting text
- Matches glass morphism aesthetic

**Files**: `ChatView.vue` (empty state section)

---

## Files to Modify

1. `pkg/web/frontend/src/views/ChatView.vue` — Sections 2, 3, 5, 6
2. `pkg/web/frontend/src/components/MessageBubble.vue` — Sections 1, 4, 5
3. `pkg/web/frontend/src/components/Chat/SpecialistsPanel.vue` — Section 4
4. `pkg/web/frontend/src/styles/globals.css` — Section 4 (glass utilities if needed)

## Out of Scope

- Component extraction (deferred to future task)
- New components or files
- Store changes
- Backend changes
