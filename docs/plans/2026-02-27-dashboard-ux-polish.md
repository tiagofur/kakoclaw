# Dashboard UX Polish Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Polish the Dashboard page to match the Chat & Tasks glass morphism design language

**Architecture:** Single-file surgical edits to DashboardView.vue — responsive spacing, icon sizing, text normalization

**Tech Stack:** Vue 3, Tailwind CSS

---

### Task 1: Header & Content Area Spacing

**Files:**
- Modify: `pkg/web/frontend/src/views/DashboardView.vue`

**Changes:**
- Line 8: Header padding `p-6` → `p-2 sm:p-3 md:p-4`
- Line 11: Title `text-2xl` → `text-xl sm:text-2xl`
- Line 20: Refresh button add `min-h-[36px] min-w-[36px]`, border → `border-makoclaw-border/50`
- Line 30: Content area `p-4 md:p-6 space-y-5` → `p-2 sm:p-3 md:p-4 space-y-3 sm:space-y-4`

### Task 2: Welcome Banner Polish

**Files:**
- Modify: `pkg/web/frontend/src/views/DashboardView.vue`

**Changes:**
- Line 46: Banner padding `p-6 md:p-8` → `p-4 sm:p-5 md:p-6`
- Line 53: Decoration SVG `w-44 h-44` → `w-32 h-32`
- Line 58: Heading `text-2xl md:text-3xl` → `text-xl sm:text-2xl`
- Line 62: CTA buttons add `min-h-[36px]`

### Task 3: Stats Grid Normalization

**Files:**
- Modify: `pkg/web/frontend/src/views/DashboardView.vue`

**Changes:**
- Line 69: Grid gap stays `gap-3 md:gap-4`
- Line 74: Icon container `p-2` → `p-2.5`
- Line 75: Icon `w-4 h-4` → `w-5 h-5`
- Line 77: Trend badge `text-[10px]` → `text-xs`
- Line 82: Label `text-[10px]` → `text-xs`
- Line 83: Value `text-3xl` → `text-2xl`

### Task 4: Chart Panels & Detailed Metrics

**Files:**
- Modify: `pkg/web/frontend/src/views/DashboardView.vue`

**Changes:**
- Lines 89,92,93: Grid gaps `gap-5` → `gap-3 sm:gap-4`
- Lines 94,105: Chart panel padding `p-5` → `p-3 sm:p-4 md:p-5`
- Lines 95,106: Header margin `mb-5` → `mb-3 sm:mb-4`
- Lines 99,110: Chart min-height `min-h-[300px]` → `min-h-[200px] sm:min-h-[250px] md:min-h-[300px]`
- Line 118: Metrics panel padding `p-5` → `p-3 sm:p-4 md:p-5`
- Line 119: Metrics header `mb-5` → `mb-3 sm:mb-4`
- Line 120: Metrics grid `gap-5` → `gap-3 sm:gap-4`
- Line 122: Metric label `text-[10px]` → `text-xs`

### Task 5: Action Launchpad & Activity Feed

**Files:**
- Modify: `pkg/web/frontend/src/views/DashboardView.vue`

**Changes:**
- Line 133: Launchpad panel `p-5` → `p-3 sm:p-4`
- Line 134: Header `mb-4` → `mb-3`
- Line 139: Icon container `w-8 h-8` → `w-9 h-9`
- Line 140: Icon `w-4 h-4` → `w-5 h-5`
- Line 142: Label `text-[11px]` → `text-xs`
- Line 148: Activity feed `p-5` → `p-3 sm:p-4`
- Line 149: Feed header `mb-4` → `mb-3`
- Line 151: "All History" link `text-[10px]` → `text-xs`
- Line 164: Activity item icon `w-8 h-8` → `w-9 h-9`
- Line 165: Activity item icon inner `w-4 h-4` → `w-5 h-5`
- Line 168: Activity title `text-[11px]` → `text-xs`
- Line 170: Activity meta `text-[10px]` → `text-xs`
- Line 172: Activity type `text-[10px]` → `text-xs`

### Task 6: Loading Skeleton Alignment

**Files:**
- Modify: `pkg/web/frontend/src/views/DashboardView.vue`

**Changes:**
- Line 33: Skeleton spacing `space-y-5` → `space-y-3 sm:space-y-4`
- Line 34: Welcome skeleton `h-48 rounded-3xl` → `h-36 sm:h-44 rounded-2xl`
- Line 36: Stat skeleton `h-32 rounded-2xl` → `h-28 sm:h-32 rounded-xl`
- Line 38: Chart skeleton `gap-6 rounded-3xl` → `gap-3 sm:gap-4 rounded-2xl`
- Line 39-40: Chart skeleton `h-80 rounded-3xl` → `h-60 sm:h-72 rounded-2xl`
