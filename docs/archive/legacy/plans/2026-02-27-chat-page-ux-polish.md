# Chat Page UX Polish — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Improve the Chat page's space efficiency, mobile responsiveness, and visual polish using glass morphism + fluid width design.

**Architecture:** Surgical in-place modifications to 4 existing files. No new components or store changes. All changes are CSS/template level in the Vue frontend.

**Tech Stack:** Vue 3, Tailwind CSS, existing `makoclaw-*` design tokens, `glass`/`glass-panel` utility classes.

**Design Doc:** `docs/plans/2026-02-27-chat-page-ux-polish-design.md`

---

### Task 1: Message Bubble Fluid Width

**Files:**
- Modify: `pkg/web/frontend/src/components/MessageBubble.vue:10`

**Step 1: Update max-width breakpoints on the bubble container**

In `MessageBubble.vue`, line 10, find:
```
'max-w-[90%] sm:max-w-[85%] lg:max-w-2xl px-4 md:px-5 py-2.5 md:py-3 shadow-lg transition-all duration-300 transform hover:scale-[1.002] animate-slideUp',
```

Replace with:
```
'max-w-[92%] sm:max-w-[88%] md:max-w-[85%] lg:max-w-3xl xl:max-w-4xl 2xl:max-w-5xl px-3 sm:px-4 md:px-5 py-2 sm:py-2.5 md:py-3 shadow-lg transition-all duration-300 transform hover:scale-[1.002] animate-slideUp',
```

Changes:
- `max-w-[90%]` → `max-w-[92%]` (more mobile space)
- `sm:max-w-[85%]` → `sm:max-w-[88%]` (more small tablet space)
- Added `md:max-w-[85%]` for tablet
- `lg:max-w-2xl` → `lg:max-w-3xl` (768px instead of 672px)
- Added `xl:max-w-4xl` (896px on desktop)
- Added `2xl:max-w-5xl` (1024px on ultrawide)
- Padding: added `sm:` intermediate steps

**Step 2: Verify in browser**

Run: `cd pkg/web/frontend && npm run dev`
Check at 390px, 768px, 1024px, 1280px, 1536px — bubbles should scale fluidly.

**Step 3: Commit**

```bash
git add pkg/web/frontend/src/components/MessageBubble.vue
git commit -m "style(chat): fluid width message bubbles scaling across breakpoints"
```

---

### Task 2: Glass Morphism Enhancement on Assistant Bubbles

**Files:**
- Modify: `pkg/web/frontend/src/styles/globals.css:119-121`
- Modify: `pkg/web/frontend/src/components/MessageBubble.vue:13`

**Step 1: Enhance the `.glass-panel` utility in globals.css**

In `globals.css`, find (around line 119):
```css
.glass-panel {
  @apply bg-makoclaw-surface/40 backdrop-blur-xl border border-makoclaw-border/30 shadow-xl;
}
```

Replace with:
```css
.glass-panel {
  @apply bg-makoclaw-surface/40 backdrop-blur-xl border border-makoclaw-border/20 shadow-xl ring-1 ring-white/[0.05];
}
```

Changes:
- Border opacity `30` → `20` (subtler)
- Added `ring-1 ring-white/[0.05]` for subtle gradient-like glow on edges

**Step 2: Enhance the assistant bubble shadow in MessageBubble.vue**

In `MessageBubble.vue`, line 13, find:
```
: 'glass-panel text-makoclaw-text rounded-2xl rounded-bl-none shadow-black/5'
```

Replace with:
```
: 'glass-panel text-makoclaw-text rounded-2xl rounded-bl-none shadow-lg shadow-black/[0.03] dark:shadow-black/20'
```

Changes:
- `shadow-black/5` → `shadow-lg shadow-black/[0.03] dark:shadow-black/20` (more depth, dark mode aware)

**Step 3: Commit**

```bash
git add pkg/web/frontend/src/styles/globals.css pkg/web/frontend/src/components/MessageBubble.vue
git commit -m "style(chat): enhanced glass morphism on assistant message bubbles"
```

---

### Task 3: Input Area Redesign — Compact Toolbar

**Files:**
- Modify: `pkg/web/frontend/src/views/ChatView.vue:193-398` (input section)

This is the largest task. The current input area has buttons below the textarea that wrap on mobile. We need to restructure to: icon toolbar above textarea, single row always.

**Step 1: Restructure the input container glass effect**

In `ChatView.vue`, find the input container (around line 193):
```html
<div class="border-t border-makoclaw-border/50 bg-makoclaw-surface/80 backdrop-blur-md p-2.5 md:p-4 z-20 relative">
```

Replace with:
```html
<div class="border-t border-makoclaw-border/30 bg-makoclaw-surface/60 backdrop-blur-xl p-2 sm:p-3 md:p-4 z-20 relative ring-1 ring-white/[0.05]">
```

Changes:
- Border opacity `50` → `30` (subtler)
- `bg-makoclaw-surface/80` → `/60` (more glass transparency)
- `backdrop-blur-md` → `backdrop-blur-xl` (stronger glass)
- Padding: `p-2.5 md:p-4` → `p-2 sm:p-3 md:p-4` (better scaling)
- Added `ring-1 ring-white/[0.05]` for glow

**Step 2: Move buttons above textarea into compact toolbar**

Find the current button row (after textarea, around lines 260-397). The buttons are currently inside a `<div class="flex items-center gap-1.5 md:gap-2 flex-wrap">` after the textarea.

Restructure the form interior so the button toolbar comes BEFORE the textarea:

```html
<!-- Compact toolbar row -->
<div class="flex items-center justify-between gap-1 px-1 mb-1.5">
  <!-- Left group: file, prompts, tools -->
  <div class="flex items-center gap-0.5 sm:gap-1">
    <!-- File attach button -->
    <button type="button" @click="triggerFileInput" ...existing handlers...
      class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-accent hover:bg-makoclaw-accent/10 transition-all"
      title="Attach file">
      <svg class="w-4 h-4" ...existing icon.../>
    </button>
    <!-- Prompt library button -->
    <button type="button" @click="showPromptModal = true"
      class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-accent hover:bg-makoclaw-accent/10 transition-all"
      title="Prompt library">
      <svg class="w-4 h-4" ...existing icon.../>
    </button>
    <!-- Tools select button -->
    <button type="button" @click="showToolsPopover = !showToolsPopover"
      class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-accent hover:bg-makoclaw-accent/10 transition-all relative"
      title="Select tools">
      <svg class="w-4 h-4" ...existing icon.../>
      <span v-if="chatStore.selectedTools.length > 0"
        class="absolute -top-0.5 -right-0.5 w-3.5 h-3.5 bg-makoclaw-accent text-white text-[8px] rounded-full flex items-center justify-center font-bold">
        {{ chatStore.selectedTools.length }}
      </span>
    </button>
  </div>
  <!-- Right group: mic, send/stop -->
  <div class="flex items-center gap-0.5 sm:gap-1">
    <!-- Mic button -->
    <button v-if="voiceSupported" type="button" @click="toggleVoiceRecording"
      class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-lg transition-all"
      :class="isRecording ? 'text-makoclaw-error bg-makoclaw-error/10 animate-pulse' : 'text-makoclaw-text-secondary hover:text-makoclaw-accent hover:bg-makoclaw-accent/10'"
      :title="isRecording ? 'Stop recording' : 'Voice input'">
      <svg class="w-4 h-4" ...existing mic icon.../>
    </button>
    <!-- Send button -->
    <button v-if="!isLoading" type="submit"
      :disabled="!isConnected || !messageInput.trim()"
      class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-lg bg-makoclaw-accent hover:bg-makoclaw-accent-hover disabled:bg-makoclaw-surface disabled:text-makoclaw-text-secondary text-white transition-all shadow-md shadow-makoclaw-accent/20 hover:shadow-makoclaw-accent/40"
      title="Send message">
      <svg class="w-4 h-4" ...existing send icon.../>
    </button>
    <!-- Stop button -->
    <button v-else type="button" @click="stopGeneration"
      class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-lg bg-makoclaw-error/10 hover:bg-makoclaw-error/20 text-makoclaw-error transition-all"
      title="Stop generation">
      <svg class="w-4 h-4" ...existing stop icon.../>
    </button>
  </div>
</div>

<!-- Attachment preview strip (keep above textarea) -->
<div v-if="attachments.length > 0" class="flex flex-wrap gap-1.5 mb-1.5">
  ...existing attachment items...
</div>

<!-- Textarea -->
<textarea ...existing textarea unchanged... />
```

Key design principles:
- All buttons are icon-only with `min-h-[36px] min-w-[36px]` touch targets
- Single row always — no `flex-wrap`
- `justify-between` separates left (actions) from right (send)
- Consistent `p-2` padding on all buttons
- On `md+`: text labels can optionally appear via `<span class="hidden md:inline ml-1 text-xs">` inside buttons

**Step 3: Remove the old button row below textarea**

Delete the old `<div class="flex items-center gap-1.5 md:gap-2 flex-wrap">` that contained the buttons below the textarea. Keep only the restructured toolbar above.

**Step 4: Verify in browser at 390px and 1024px**

Run: `cd pkg/web/frontend && npm run dev`
- At 390px: toolbar should be one row, all icons visible, no wrapping
- At 1024px: same layout with more breathing room

**Step 5: Commit**

```bash
git add pkg/web/frontend/src/views/ChatView.vue
git commit -m "style(chat): compact icon toolbar above textarea for better mobile UX"
```

---

### Task 4: Top Bar Simplification

**Files:**
- Modify: `pkg/web/frontend/src/views/ChatView.vue:99-140`

**Step 1: Replace the flex-wrap top bar with a fixed single-row layout**

Find the top bar (around line 99):
```html
<div class="flex items-center justify-between px-2 md:px-4 py-1.5 md:py-2 border-b border-makoclaw-border/30 bg-makoclaw-surface/30 backdrop-blur-sm z-20 gap-2 flex-wrap">
```

Replace the entire top bar section with:
```html
<div class="flex items-center px-2 sm:px-3 md:px-4 py-1.5 md:py-2 border-b border-makoclaw-border/20 bg-makoclaw-surface/30 backdrop-blur-xl z-20 gap-2">
  <!-- Sidebar toggle (fixed left) -->
  <button @click="toggleSidebar"
    class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center hover:bg-makoclaw-accent/10 rounded-xl text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all flex-shrink-0"
    title="Toggle Sidebar">
    <svg class="w-4 h-4 md:w-5 md:h-5" ...existing sidebar icon.../>
  </button>

  <!-- Agent status (flex-1, truncates) -->
  <div v-if="chatStore.globalIsLoading" class="flex-1 flex items-center gap-1.5 px-2 py-1 bg-makoclaw-accent/10 border border-makoclaw-accent/30 rounded-lg min-w-0">
    <svg class="w-3.5 h-3.5 text-makoclaw-accent animate-spin flex-shrink-0" ...existing spinner.../>
    <span class="text-[10px] md:text-xs font-medium text-makoclaw-accent truncate">Agent working...</span>
  </div>
  <div v-else class="flex-1"></div>

  <!-- Model selector (fixed right) -->
  <div class="flex items-center gap-1 flex-shrink-0">
    <svg class="w-3 h-3 md:w-3.5 md:h-3.5 text-makoclaw-text-secondary flex-shrink-0" ...existing model icon.../>
    <select v-model="chatStore.selectedModel" :disabled="chatStore.allModels.length === 0"
      class="bg-makoclaw-bg/50 border border-makoclaw-border rounded-lg px-2 py-1 text-[10px] md:text-xs text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/50 focus:border-makoclaw-accent transition-all cursor-pointer max-w-[140px] md:max-w-[240px]">
      ...existing options...
    </select>
  </div>
</div>
```

Changes:
- Removed `flex-wrap` — always single row
- Removed `justify-between` — use `flex-1` spacer instead
- `backdrop-blur-sm` → `backdrop-blur-xl` (stronger glass)
- Agent status: `flex-1 min-w-0 truncate` (fills space, truncates)
- Model selector: `max-w-[140px] md:max-w-[240px]` (tighter on mobile)
- Removed `order-` classes entirely
- All interactive elements: `min-h-[36px]` touch target

**Step 2: Verify no wrapping at 390px**

At 390px: sidebar toggle (36px) + spacer + model selector (~140px) should fit in one row.

**Step 3: Commit**

```bash
git add pkg/web/frontend/src/views/ChatView.vue
git commit -m "style(chat): single-row top bar with fixed layout, no wrapping"
```

---

### Task 5: Chat History Sidebar Glass Enhancement

**Files:**
- Modify: `pkg/web/frontend/src/views/ChatView.vue:7-11`

**Step 1: Enhance sidebar glass effect**

Find (around line 8-11):
```
'flex-shrink-0 border-r border-makoclaw-border bg-makoclaw-surface/50 backdrop-blur-md transition-all duration-500 ease-[cubic-bezier(0.4,0,0.2,1)] flex flex-col',
```

Replace with:
```
'flex-shrink-0 border-r border-makoclaw-border/20 bg-makoclaw-surface/30 backdrop-blur-xl transition-all duration-500 ease-[cubic-bezier(0.4,0,0.2,1)] flex flex-col ring-1 ring-white/[0.03]',
```

Changes:
- `border-makoclaw-border` → `border-makoclaw-border/20` (subtler)
- `bg-makoclaw-surface/50` → `/30` (more transparent glass)
- `backdrop-blur-md` → `backdrop-blur-xl` (stronger blur)
- Added `ring-1 ring-white/[0.03]` (subtle glow)

**Step 2: Commit**

```bash
git add pkg/web/frontend/src/views/ChatView.vue
git commit -m "style(chat): glass morphism on chat history sidebar"
```

---

### Task 6: Spacing Normalization

**Files:**
- Modify: `pkg/web/frontend/src/views/ChatView.vue` (multiple sections)
- Modify: `pkg/web/frontend/src/components/MessageBubble.vue`

**Step 1: Normalize messages container spacing**

Find (around line 149):
```
class="flex-1 overflow-y-auto p-3 md:p-4 space-y-4 md:space-y-6 z-10"
```

Replace with:
```
class="flex-1 overflow-y-auto p-2 sm:p-3 md:p-4 space-y-3 md:space-y-4 z-10"
```

Changes:
- Added `sm:` intermediate step
- `space-y-4 md:space-y-6` → `space-y-3 md:space-y-4` (tighter message spacing)

**Step 2: Normalize sidebar header spacing**

Find (around line 13):
```
class="p-2 md:p-4 border-b border-makoclaw-border flex justify-between items-center gap-2"
```

Replace with:
```
class="p-2 sm:p-3 md:p-4 border-b border-makoclaw-border/30 flex justify-between items-center gap-2"
```

**Step 3: Normalize specialists panel wrapper**

Find (around line 143):
```html
<div class="px-3 md:px-4 pt-3">
```

Replace with:
```html
<div class="px-2 sm:px-3 md:px-4 pt-2 sm:pt-3">
```

**Step 4: Normalize message bubble action buttons touch targets**

In `MessageBubble.vue`, find the action buttons area (around line 60-101). Ensure all buttons have `min-h-[36px] min-w-[36px]` on mobile. Find small button classes like `p-1` and update to `p-1.5 sm:p-2`.

**Step 5: Commit**

```bash
git add pkg/web/frontend/src/views/ChatView.vue pkg/web/frontend/src/components/MessageBubble.vue
git commit -m "style(chat): normalize spacing scale across breakpoints"
```

---

### Task 7: Empty State Polish

**Files:**
- Modify: `pkg/web/frontend/src/views/ChatView.vue:152-160`

**Step 1: Enhance empty state with glass morphism and animation**

Find (around lines 152-160):
```html
<div v-if="messages.length === 0" class="flex flex-col items-center justify-center h-full text-makoclaw-text-secondary opacity-60">
  <div class="bg-makoclaw-surface/50 p-6 rounded-full mb-4">
    <svg class="w-12 h-12 text-makoclaw-accent" .../>
  </div>
  <p class="text-lg font-medium">Start a conversation</p>
  <p class="text-sm">Ask anything or run a task</p>
```

Replace with:
```html
<div v-if="messages.length === 0" class="flex flex-col items-center justify-center h-full text-makoclaw-text-secondary animate-fadeIn">
  <div class="glass-panel p-6 sm:p-8 rounded-2xl mb-4 shadow-lg shadow-makoclaw-accent/5">
    <svg class="w-10 h-10 sm:w-12 sm:h-12 text-makoclaw-accent animate-pulse" style="animation-duration: 3s;" ...existing icon.../>
  </div>
  <p class="text-base sm:text-lg font-semibold text-makoclaw-text mt-2">Start a conversation</p>
  <p class="text-xs sm:text-sm text-makoclaw-text-secondary/70 mt-1">Ask anything or run a task</p>
```

Changes:
- Removed `opacity-60` (was making everything faded)
- Icon container: `bg-makoclaw-surface/50` → `glass-panel` with `rounded-2xl`
- Added `shadow-lg shadow-makoclaw-accent/5` for accent glow
- Icon: `animate-pulse` with slow 3s duration
- Responsive icon: `w-10 sm:w-12`
- Title: `font-medium` → `font-semibold`, added `text-makoclaw-text` for better contrast
- Subtitle: added `text-makoclaw-text-secondary/70` + `mt-1` spacing

**Step 2: Commit**

```bash
git add pkg/web/frontend/src/views/ChatView.vue
git commit -m "style(chat): polished empty state with glass morphism and subtle animation"
```

---

### Task 8: Specialists Panel Glass Cards

**Files:**
- Modify: `pkg/web/frontend/src/components/Chat/SpecialistsPanel.vue:32`

**Step 1: Add glass card effect to individual specialist items**

Find (around line 32):
```
class="flex items-start gap-3 p-2 rounded-lg hover:bg-makoclaw-bg/30 transition-colors"
```

Replace with:
```
class="flex items-start gap-3 p-2 sm:p-2.5 rounded-xl bg-makoclaw-surface/20 hover:bg-makoclaw-surface/40 border border-makoclaw-border/10 hover:border-makoclaw-border/20 backdrop-blur-sm transition-all duration-200"
```

Changes:
- `rounded-lg` → `rounded-xl` (softer)
- Added `bg-makoclaw-surface/20` base (subtle glass)
- `hover:bg-makoclaw-bg/30` → `hover:bg-makoclaw-surface/40` (consistent with glass system)
- Added subtle border on hover
- Added `backdrop-blur-sm` for glass depth
- `transition-colors` → `transition-all duration-200`

**Step 2: Commit**

```bash
git add pkg/web/frontend/src/components/Chat/SpecialistsPanel.vue
git commit -m "style(chat): glass card effect on specialist items"
```

---

### Task 9: Build and Visual QA

**Files:** None (verification only)

**Step 1: Run production build**

```bash
cd pkg/web/frontend && npm run build
```

Expected: Build succeeds with no errors.

**Step 2: Visual QA checklist**

Run dev server and check:
- [ ] 390px: Input toolbar is single row, no wrapping
- [ ] 390px: Top bar is single row, model selector fits
- [ ] 390px: Message bubbles use 92% width
- [ ] 768px: Sidebar visible, messages scale to 85%
- [ ] 1024px: Messages use max-w-3xl (768px)
- [ ] 1280px: Messages use max-w-4xl (896px)
- [ ] 1536px: Messages use max-w-5xl (1024px)
- [ ] Glass effects visible on: assistant bubbles, input area, sidebar, specialists
- [ ] Empty state shows glass panel with pulsing icon
- [ ] Dark mode: shadows and glass effects look correct
- [ ] Touch targets: all buttons feel tappable on mobile

**Step 3: Final commit (if any fixes needed)**

```bash
git add -A && git commit -m "fix(chat): visual QA adjustments"
```
