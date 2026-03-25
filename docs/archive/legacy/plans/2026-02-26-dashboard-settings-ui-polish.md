# Dashboard & Settings UI Polish — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Reduce visual weight across Dashboard and Settings (9 files) via surgical proportional scaling — smaller icons, tighter padding, clear typographic hierarchy — while preserving the modern/fun character.

**Architecture:** Pure Tailwind class replacements. No structural HTML changes. No logic changes. No new components. The design system tokens are: icon containers shrink ~30%, card padding -25%, `font-black` reserved for page titles only, border radii step down one level, spacing reduced uniformly.

**Tech Stack:** Vue 3, Tailwind CSS, Vite. Dev server: `cd pkg/web/frontend && npm run dev` → `http://localhost:5173`

---

## Reference: The 5 Token Rules

Apply these consistently everywhere across all 9 files. When in doubt, use these:

| Token | Before | After |
|---|---|---|
| Section icon (gradient bg) | `p-3 rounded-2xl` + `w-6 h-6` | `p-2 rounded-xl` + `w-4 h-4` |
| Standalone large icon | `w-14 h-14 rounded-2xl` | `w-10 h-10 rounded-xl` |
| Card padding | `p-8 md:p-10` | `p-5 md:p-6` |
| Card border radius | `rounded-[2rem]` | `rounded-2xl` |
| Section header label | `font-black uppercase tracking-[0.2em]` | `font-medium uppercase tracking-wide` |
| Section header title | `text-lg font-black italic` | `text-base font-semibold` |
| Form inputs | `rounded-2xl px-5 py-3.5 border-2` | `rounded-xl px-4 py-2.5 border` |
| Primary button | `py-4 font-black uppercase tracking-widest` | `py-2.5 font-semibold` |
| Spacing (gaps/mb) | `gap-8`, `mb-8`, `mb-10` | `gap-5`, `mb-5`, `mb-6` |
| Stat number | `text-4xl font-black` | `text-3xl font-bold` |

---

## Task 1: DashboardView — Welcome Banner

**File:** `pkg/web/frontend/src/views/DashboardView.vue`

**Step 1: Shrink the banner padding and heading**

Find:
```
class="relative overflow-hidden bg-gradient-to-br from-makoclaw-accent via-blue-600 to-indigo-700 rounded-[2rem] p-8 md:p-12 shadow-2xl shadow-makoclaw-accent/20 group animate-fade-in-up"
```
Replace with:
```
class="relative overflow-hidden bg-gradient-to-br from-makoclaw-accent via-blue-600 to-indigo-700 rounded-2xl p-6 md:p-8 shadow-2xl shadow-makoclaw-accent/20 group animate-fade-in-up"
```

**Step 2: Shrink banner heading**

Find:
```
<h3 class="text-3xl md:text-5xl font-black text-white leading-tight">Welcome back,<br/>{{ authStore.user?.username || 'Commander' }}</h3>
```
Replace with:
```
<h3 class="text-2xl md:text-3xl font-black text-white leading-tight">Welcome back,<br/>{{ authStore.user?.username || 'Commander' }}</h3>
```

**Step 3: Shrink banner body text and reduce background bolt**

Find:
```
<p class="text-white/70 mt-4 text-sm md:text-lg max-w-xl leading-relaxed">
```
Replace with:
```
<p class="text-white/70 mt-3 text-sm max-w-xl leading-relaxed">
```

Find:
```
class="absolute top-0 right-0 p-8 transform rotate-12 opacity-10 transition-transform group-hover:rotate-45 group-hover:scale-125 duration-1000 hidden md:block"
```
Replace with:
```
class="absolute top-0 right-0 p-8 transform rotate-12 opacity-[0.07] transition-transform group-hover:rotate-45 group-hover:scale-110 duration-1000 hidden md:block"
```

Find (the bolt SVG w-64):
```
<svg class="w-64 h-64"
```
Replace with:
```
<svg class="w-44 h-44"
```

**Step 4: Shrink banner buttons**

Find:
```
<router-link to="/chat" class="px-6 py-3 bg-white text-makoclaw-accent rounded-xl font-bold shadow-xl hover:shadow-white/20 transition-all hover:-translate-y-1 active:scale-95 text-sm uppercase tracking-wider">Launch New Session</router-link>
               <router-link to="/tasks" class="px-6 py-3 bg-white/10 backdrop-blur-md border border-white/20 text-white rounded-xl font-bold hover:bg-white/20 transition-all active:scale-95 text-sm uppercase tracking-wider">Tasks Dashboard</router-link>
```
Replace with:
```
<router-link to="/chat" class="px-5 py-2 bg-white text-makoclaw-accent rounded-xl font-semibold shadow-xl hover:shadow-white/20 transition-all hover:-translate-y-0.5 active:scale-95 text-sm">Launch New Session</router-link>
               <router-link to="/tasks" class="px-5 py-2 bg-white/10 backdrop-blur-md border border-white/20 text-white rounded-xl font-medium hover:bg-white/20 transition-all active:scale-95 text-sm">Tasks Dashboard</router-link>
```

Also tighten the button wrapper margin:

Find:
```
<div class="mt-8 flex flex-wrap gap-4">
```
Replace with:
```
<div class="mt-5 flex flex-wrap gap-3">
```

**Step 5: Verify**

Start dev server if not running: `cd pkg/web/frontend && npm run dev`
Navigate to Dashboard. Banner should be noticeably more compact. Heading should be smaller, bolt icon more subtle.

---

## Task 2: DashboardView — Stat Cards

**File:** `pkg/web/frontend/src/views/DashboardView.vue`

**Step 1: Shrink stat card container and spacing**

Find:
```
class="glass-panel rounded-3xl p-6 transition-all duration-500 hover:shadow-2xl hover:shadow-makoclaw-accent/10 hover:-translate-y-2 group animate-fade-in-up"
```
Replace with:
```
class="glass-panel rounded-xl p-4 transition-all duration-300 hover:shadow-xl hover:shadow-makoclaw-accent/10 hover:-translate-y-1 group animate-fade-in-up"
```

**Step 2: Shrink stat icon container**

Find:
```
<div class="p-3 rounded-2xl bg-gradient-to-br transition-colors duration-500" :class="stat.iconBg">
                <component :is="stat.icon" class="w-6 h-6 text-white" />
```
Replace with:
```
<div class="p-2 rounded-xl bg-gradient-to-br transition-colors duration-300" :class="stat.iconBg">
                <component :is="stat.icon" class="w-4 h-4 text-white" />
```

**Step 3: Shrink stat number and label**

Find:
```
<div class="mt-6">
              <div class="text-[10px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/60">{{ stat.label }}</div>
              <div class="text-4xl font-black mt-1 bg-gradient-to-br from-makoclaw-text to-makoclaw-text-secondary bg-clip-text text-transparent">{{ stat.value }}</div>
```
Replace with:
```
<div class="mt-4">
              <div class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60">{{ stat.label }}</div>
              <div class="text-3xl font-bold mt-1 bg-gradient-to-br from-makoclaw-text to-makoclaw-text-secondary bg-clip-text text-transparent">{{ stat.value }}</div>
```

**Step 4: Tighten the overall stats grid gap**

Find:
```
<div class="grid grid-cols-2 lg:grid-cols-4 gap-4 md:gap-6">
```
Replace with:
```
<div class="grid grid-cols-2 lg:grid-cols-4 gap-3 md:gap-4">
```

**Step 5: Verify**

Stat cards should now look compact and proportional. Numbers still prominent but not overwhelming.

---

## Task 3: DashboardView — Launchpad & Activity Feed

**File:** `pkg/web/frontend/src/views/DashboardView.vue`

**Step 1: Shrink launchpad card**

Find:
```
<div class="glass-panel rounded-[2rem] p-8 animate-fade-in-up" style="animation-delay: 400ms">
              <h3 class="text-xs font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70 mb-6 px-2">Action Launchpad</h3>
```
Replace with:
```
<div class="glass-panel rounded-2xl p-5 animate-fade-in-up" style="animation-delay: 400ms">
              <h3 class="text-xs font-medium uppercase tracking-wide text-makoclaw-text-secondary/70 mb-4 px-1">Action Launchpad</h3>
```

**Step 2: Shrink launchpad action items**

Find:
```
class="flex flex-col items-center justify-center p-4 rounded-[1.5rem] border border-makoclaw-border hover:border-makoclaw-accent/40 bg-makoclaw-surface/30 hover:bg-makoclaw-accent/5 transition-all group active:scale-95"
```
Replace with:
```
class="flex flex-col items-center justify-center p-3 rounded-xl border border-makoclaw-border hover:border-makoclaw-accent/40 bg-makoclaw-surface/30 hover:bg-makoclaw-accent/5 transition-all group active:scale-95"
```

Find:
```
<div class="w-10 h-10 rounded-xl flex items-center justify-center mb-3 transition-transform group-hover:scale-110 group-hover:rotate-3" :class="action.color">
                     <component :is="action.icon" class="w-5 h-5 text-white" />
                   </div>
                   <span class="text-[11px] font-black uppercase text-makoclaw-text-secondary/80 group-hover:text-makoclaw-accent">{{ action.label }}</span>
```
Replace with:
```
<div class="w-8 h-8 rounded-lg flex items-center justify-center mb-2 transition-transform group-hover:scale-105 group-hover:rotate-3" :class="action.color">
                     <component :is="action.icon" class="w-4 h-4 text-white" />
                   </div>
                   <span class="text-[11px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/80 group-hover:text-makoclaw-accent">{{ action.label }}</span>
```

**Step 3: Shrink activity feed card**

Find:
```
<div class="glass-panel rounded-[2rem] p-8 animate-fade-in-up flex flex-col h-[500px]" style="animation-delay: 550ms">
              <div class="flex items-center justify-between mb-8 px-2">
                <h3 class="text-xs font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70">Recent Pulse</h3>
```
Replace with:
```
<div class="glass-panel rounded-2xl p-5 animate-fade-in-up flex flex-col h-[420px]" style="animation-delay: 550ms">
              <div class="flex items-center justify-between mb-4 px-1">
                <h3 class="text-xs font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">Recent Pulse</h3>
```

**Step 4: Shrink activity feed items**

Find:
```
class="flex items-center gap-4 p-4 rounded-2xl bg-makoclaw-surface/40 border border-makoclaw-border/30 hover:border-makoclaw-accent/30 transition-all group cursor-pointer"
```
Replace with:
```
class="flex items-center gap-3 p-3 rounded-xl bg-makoclaw-surface/40 border border-makoclaw-border/30 hover:border-makoclaw-accent/30 transition-all group cursor-pointer"
```

Find:
```
<div class="w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0 transition-colors" :class="item.iconBg">
                    <component :is="item.icon" class="w-5 h-5 text-white" />
```
Replace with:
```
<div class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 transition-colors" :class="item.iconBg">
                    <component :is="item.icon" class="w-4 h-4 text-white" />
```

**Step 5: Shrink charts cards and Live System Metrics**

Find:
```
<div class="glass-panel rounded-[2rem] p-8 flex flex-col animate-fade-in-up" style="animation-delay: 600ms">
                <div class="flex items-center justify-between mb-8">
                  <h3 class="text-xs font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70">Model Intelligence</h3>
```
Replace with:
```
<div class="glass-panel rounded-2xl p-5 flex flex-col animate-fade-in-up" style="animation-delay: 600ms">
                <div class="flex items-center justify-between mb-5">
                  <h3 class="text-xs font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">Model Intelligence</h3>
```

Find:
```
<div class="glass-panel rounded-[2rem] p-8 flex flex-col animate-fade-in-up" style="animation-delay: 750ms">
                 <div class="flex items-center justify-between mb-8">
                  <h3 class="text-xs font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70">Operations Status</h3>
```
Replace with:
```
<div class="glass-panel rounded-2xl p-5 flex flex-col animate-fade-in-up" style="animation-delay: 750ms">
                 <div class="flex items-center justify-between mb-5">
                  <h3 class="text-xs font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">Operations Status</h3>
```

Find:
```
<div class="glass-panel rounded-[2rem] p-8 animate-fade-in-up" style="animation-delay: 900ms">
              <h3 class="text-xs font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70 mb-8 px-2">Live System Metrics</h3>
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-8">
                <div v-for="metric in detailedMetrics" :key="metric.label" class="p-4 rounded-3xl hover:bg-makoclaw-surface/50 transition-colors group">
                  <div class="text-[10px] font-bold text-makoclaw-text-secondary/50 uppercase tracking-widest mb-1 group-hover:text-makoclaw-accent transition-colors">{{ metric.label }}</div>
                  <div class="text-2xl font-black text-makoclaw-text">{{ metric.value }}</div>
```
Replace with:
```
<div class="glass-panel rounded-2xl p-5 animate-fade-in-up" style="animation-delay: 900ms">
              <h3 class="text-xs font-medium uppercase tracking-wide text-makoclaw-text-secondary/70 mb-5 px-1">Live System Metrics</h3>
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-5">
                <div v-for="metric in detailedMetrics" :key="metric.label" class="p-3 rounded-xl hover:bg-makoclaw-surface/50 transition-colors group">
                  <div class="text-[10px] font-medium text-makoclaw-text-secondary/50 uppercase tracking-wide mb-1 group-hover:text-makoclaw-accent transition-colors">{{ metric.label }}</div>
                  <div class="text-xl font-bold text-makoclaw-text">{{ metric.value }}</div>
```

**Step 6: Tighten overall content spacing**

Find:
```
<div class="flex-1 overflow-auto p-4 md:p-8 space-y-8 custom-scrollbar relative z-10">
```
Replace with:
```
<div class="flex-1 overflow-auto p-4 md:p-6 space-y-5 custom-scrollbar relative z-10">
```

Find:
```
<div class="grid grid-cols-1 xl:grid-cols-3 gap-8">
```
Replace with:
```
<div class="grid grid-cols-1 xl:grid-cols-3 gap-5">
```

Find the two inner grid columns:
```
<div class="xl:col-span-2 space-y-8">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
```
Replace with:
```
<div class="xl:col-span-2 space-y-5">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
```

Find:
```
<div class="space-y-8">
            <!-- Launchpad (Replaces Quick Actions) -->
```
Replace with:
```
<div class="space-y-5">
            <!-- Launchpad (Replaces Quick Actions) -->
```

**Step 7: Verify full Dashboard**

Navigate to `/` on the dev server. The dashboard should feel airy but proportional — no element should dominate. Commit after verification.

```bash
git add pkg/web/frontend/src/views/DashboardView.vue
git commit -m "ui: reduce dashboard visual weight — tighter spacing, smaller icons, proportional typography"
```

---

## Task 4: SettingsView — Sidebar Navigation

**File:** `pkg/web/frontend/src/views/SettingsView.vue`

**Step 1: Tighten sidebar nav items**

Find:
```
class="w-full flex items-center gap-4 px-4 py-3.5 rounded-2xl text-[11px] font-black uppercase tracking-widest transition-all group"
```
Replace with:
```
class="w-full flex items-center gap-3 px-4 py-2.5 rounded-xl text-[11px] font-medium uppercase tracking-wide transition-all group"
```

**Step 2: Shrink sidebar user card avatar**

Find:
```
<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-makoclaw-accent to-blue-600 flex items-center justify-center text-white font-black">
```
Replace with:
```
<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-makoclaw-accent to-blue-600 flex items-center justify-center text-white font-bold text-xs">
```

**Step 3: Tighten sidebar user card**

Find:
```
<div class="glass-panel rounded-2xl p-4 border border-makoclaw-accent/10">
          <div class="flex items-center gap-3">
```
Replace with:
```
<div class="glass-panel rounded-xl p-3 border border-makoclaw-accent/10">
          <div class="flex items-center gap-2.5">
```

**Step 4: Tighten main content area padding**

Find:
```
<main class="flex-1 overflow-auto custom-scrollbar p-6 md:p-10 relative z-10">
```
Replace with:
```
<main class="flex-1 overflow-auto custom-scrollbar p-5 md:p-8 relative z-10">
```

**Step 5: Shrink desktop content header margin**

Find:
```
<div class="hidden lg:flex items-center justify-between mb-10 animate-fade-in-up">
              <div>
                <h1 class="text-3xl font-black text-makoclaw-text">{{ activeTabLabel }}</h1>
```
Replace with:
```
<div class="hidden lg:flex items-center justify-between mb-6 animate-fade-in-up">
              <div>
                <h1 class="text-2xl font-black text-makoclaw-text">{{ activeTabLabel }}</h1>
```

**Step 6: Shrink users admin section card**

Find:
```
<div class="glass-panel rounded-[2rem] p-8">
                       <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4 mb-8">
```
Replace with:
```
<div class="glass-panel rounded-2xl p-5">
                       <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4 mb-5">
```

Find (the "User Hub" label):
```
<h3 class="text-xs font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70 italic">User Hub</h3>
```
Replace with:
```
<h3 class="text-xs font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">User Hub</h3>
```

**Step 7: Verify sidebar and settings layout**

Settings page should feel less spacious. Sidebar nav items more compact. Content area less padded. Commit.

```bash
git add pkg/web/frontend/src/views/SettingsView.vue
git commit -m "ui: tighten settings sidebar nav and main layout proportions"
```

---

## Task 5: AgentSettingsTab

**File:** `pkg/web/frontend/src/components/settings/AgentSettingsTab.vue`

**Step 1: Shrink orchestrator card**

Find:
```
<div class="glass-panel rounded-[2rem] p-8 md:p-10 border border-makoclaw-border/50 relative overflow-hidden group">
```
Replace with:
```
<div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group">
```

**Step 2: Shrink section icon + header + mb**

Find:
```
<div class="flex items-center justify-between mb-8">
          <div class="flex items-center gap-4">
            <div class="p-3 rounded-2xl bg-gradient-to-br from-makoclaw-accent to-indigo-600 shadow-lg shadow-makoclaw-accent/20">
              <IconOrchestrator class="w-6 h-6 text-white" />
            </div>
            <div>
              <h3 class="text-[11px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70">Intelligence Matrix</h3>
              <p class="text-lg font-black text-makoclaw-text italic leading-none">Orchestrator Protocol</p>
```
Replace with:
```
<div class="flex items-center justify-between mb-5">
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-xl bg-gradient-to-br from-makoclaw-accent to-indigo-600 shadow-md shadow-makoclaw-accent/20">
              <IconOrchestrator class="w-4 h-4 text-white" />
            </div>
            <div>
              <h3 class="text-[11px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">Intelligence Matrix</h3>
              <p class="text-base font-semibold text-makoclaw-text leading-none">Orchestrator Protocol</p>
```

**Step 3: Shrink description mb**

Find:
```
<p class="text-sm font-medium text-makoclaw-text-secondary/60 mb-10 leading-relaxed max-w-2xl">
```
Replace with:
```
<p class="text-sm font-medium text-makoclaw-text-secondary/60 mb-6 leading-relaxed max-w-2xl">
```

**Step 4: Shrink orchestrator fields grid**

Find:
```
<div class="space-y-8 pt-6 border-t border-makoclaw-border/30">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
```
Replace with:
```
<div class="space-y-5 pt-5 border-t border-makoclaw-border/30">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
```

**Step 5: Shrink orchestrator inputs**

Find (two selects in orchestrator — note there are 2, use replace_all carefully; target the ones in the orchestrator section):
```
class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer"
```
Replace with:
```
class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer"
```

Find (number inputs):
```
class="w-full bg-makoclaw-bg/20 border-2 border-makoclaw-border/30 rounded-2xl px-5 py-3 text-sm font-black text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all"
```
Replace with:
```
class="w-full bg-makoclaw-bg/20 border border-makoclaw-border/30 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all"
```

Find number input labels:
```
class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1"
```
Replace all with:
```
class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1"
```

**Step 6: Shrink specialists section (second card)**

Find:
```
<div class="glass-panel rounded-[2rem] p-8 md:p-10 border border-makoclaw-border/50 relative overflow-hidden">
      <div class="flex flex-col sm:flex-row items-center justify-between gap-6 mb-10">
        <div class="flex items-center gap-4">
          <div class="p-3 rounded-2xl bg-gradient-to-br from-blue-500 to-cyan-600 shadow-lg shadow-blue-500/20">
            <IconSpecialists class="w-6 h-6 text-white" />
          </div>
          <div>
            <h3 class="text-[11px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70 italic">Specialist Fleet</h3>
```
Replace with:
```
<div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden">
      <div class="flex flex-col sm:flex-row items-center justify-between gap-4 mb-5">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-600 shadow-md shadow-blue-500/20">
            <IconSpecialists class="w-4 h-4 text-white" />
          </div>
          <div>
            <h3 class="text-[11px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">Specialist Fleet</h3>
```

**Step 7: Shrink outer spacing**

Find:
```
<div class="space-y-8 max-w-4xl mx-auto animate-fade-in-up">
```
Replace with:
```
<div class="space-y-5 max-w-4xl mx-auto animate-fade-in-up">
```

**Step 8: Commit**

```bash
git add pkg/web/frontend/src/components/settings/AgentSettingsTab.vue
git commit -m "ui: reduce AgentSettingsTab icon and card proportions"
```

---

## Task 6: ChannelsSettingsTab

**File:** `pkg/web/frontend/src/components/settings/ChannelsSettingsTab.vue`

**Step 1: Shrink channel cards**

Find:
```
class="glass-panel rounded-[2rem] p-8 transition-all duration-500 hover:shadow-2xl hover:shadow-makoclaw-accent/5 hover:-translate-y-2 group relative overflow-hidden flex flex-col h-full border border-makoclaw-border/50 hover:border-makoclaw-accent/40"
```
Replace with:
```
class="glass-panel rounded-2xl p-5 transition-all duration-300 hover:shadow-xl hover:shadow-makoclaw-accent/5 hover:-translate-y-1 group relative overflow-hidden flex flex-col h-full border border-makoclaw-border/50 hover:border-makoclaw-accent/40"
```

**Step 2: Shrink channel icon**

Find:
```
class="w-14 h-14 rounded-2xl flex items-center justify-center bg-makoclaw-surface border-2 border-makoclaw-border/50 text-makoclaw-text-secondary transition-all duration-500 group-hover:scale-110 group-hover:rotate-3 shadow-lg"
                :class="{'!bg-makoclaw-accent !border-makoclaw-accent/20 !text-white shadow-makoclaw-accent/40': channels[channel.id]?.enabled}"
```
Replace with:
```
class="w-10 h-10 rounded-xl flex items-center justify-center bg-makoclaw-surface border border-makoclaw-border/50 text-makoclaw-text-secondary transition-all duration-300 group-hover:scale-105 group-hover:rotate-3 shadow-md"
                :class="{'!bg-makoclaw-accent !border-makoclaw-accent/20 !text-white shadow-makoclaw-accent/30': channels[channel.id]?.enabled}"
```

**Step 3: Shrink channel name and subtitle**

Find:
```
<h3 class="font-black text-lg text-makoclaw-text tracking-tight italic">{{ channel.name }}</h3>
                <span class="text-[9px] font-bold uppercase tracking-widest text-makoclaw-text-secondary/50">Frequency: VHF-B</span>
```
Replace with:
```
<h3 class="font-semibold text-base text-makoclaw-text tracking-tight">{{ channel.name }}</h3>
                <span class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/50">Uplink Channel</span>
```

**Step 4: Shrink channel card header mb and gap**

Find:
```
<div class="flex items-center justify-between mb-8 relative z-10">
            <div class="flex items-center gap-4">
```
Replace with:
```
<div class="flex items-center justify-between mb-4 relative z-10">
            <div class="flex items-center gap-3">
```

**Step 5: Shrink channel description and config button mb**

Find:
```
<p class="text-xs font-medium text-makoclaw-text-secondary/60 mb-10 leading-relaxed h-10 line-clamp-2 relative z-10 group-hover:text-makoclaw-text-secondary transition-colors">
```
Replace with:
```
<p class="text-xs font-medium text-makoclaw-text-secondary/60 mb-5 leading-relaxed h-10 line-clamp-2 relative z-10 group-hover:text-makoclaw-text-secondary transition-colors">
```

**Step 6: Shrink config button and active indicator**

Find:
```
class="flex-1 py-4 text-[10px] font-black uppercase tracking-[0.2em] bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl hover:border-makoclaw-accent/40 hover:bg-makoclaw-surface transition-all flex items-center justify-center text-makoclaw-text group/btn active:scale-95"
```
Replace with:
```
class="flex-1 py-2.5 text-[10px] font-medium uppercase tracking-wide bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl hover:border-makoclaw-accent/40 hover:bg-makoclaw-surface transition-all flex items-center justify-center text-makoclaw-text group/btn active:scale-95"
```

Find (active status square):
```
class="w-14 h-14 rounded-2xl bg-makoclaw-accent/10 border-2 border-makoclaw-accent/20 flex items-center justify-center text-makoclaw-accent shadow-inner"
```
Replace with:
```
class="w-10 h-10 rounded-xl bg-makoclaw-accent/10 border border-makoclaw-accent/20 flex items-center justify-center text-makoclaw-accent"
```

**Step 7: Shrink grid gap and section header**

Find:
```
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
```
Replace with:
```
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
```

Find (section header):
```
<span class="text-[9px] font-black uppercase tracking-[0.4em] text-makoclaw-text-secondary">Neural Uplink Channels</span>
```
Replace with:
```
<span class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary">Uplink Channels</span>
```

**Step 8: Shrink security note card**

Find:
```
<div class="glass-panel p-6 rounded-[2rem] bg-amber-500/5 border border-amber-500/10 flex items-center gap-6 mt-12 animate-fade-in-up"
```
Replace with:
```
<div class="glass-panel p-4 rounded-2xl bg-amber-500/5 border border-amber-500/10 flex items-center gap-4 mt-6 animate-fade-in-up"
```

Find (security note icon):
```
<div class="p-4 rounded-2xl bg-amber-500/10 text-amber-500 shadow-xl shadow-amber-500/5">
           <IconShield class="w-6 h-6" />
```
Replace with:
```
<div class="p-2 rounded-xl bg-amber-500/10 text-amber-500">
           <IconShield class="w-4 h-4" />
```

**Step 9: Shrink outer spacing**

Find:
```
<div class="space-y-8 max-w-5xl mx-auto animate-fade-in-up">
```
Replace with:
```
<div class="space-y-5 max-w-5xl mx-auto animate-fade-in-up">
```

**Step 10: Commit**

```bash
git add pkg/web/frontend/src/components/settings/ChannelsSettingsTab.vue
git commit -m "ui: reduce ChannelsSettingsTab channel card and icon proportions"
```

---

## Task 7: ProvidersSettingsTab

**File:** `pkg/web/frontend/src/components/settings/ProvidersSettingsTab.vue`

**Step 1: Shrink provider cards**

Find:
```
class="glass-panel rounded-[2rem] p-8 md:p-10 border border-makoclaw-border/50 relative overflow-hidden group hover:border-makoclaw-accent/30 transition-all duration-500"
```
Replace with:
```
class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group hover:border-makoclaw-accent/30 transition-all duration-300"
```

**Step 2: Shrink provider icon, name, and header mb**

Find:
```
<div class="flex items-center justify-between mb-8 relative z-10">
        <div class="flex items-center gap-4">
          <div class="w-14 h-14 rounded-2xl bg-makoclaw-surface border-2 border-makoclaw-border/50 flex items-center justify-center text-makoclaw-accent shadow-xl group-hover:scale-110 group-hover:rotate-3 transition-all duration-500">
             <span class="text-xs font-black uppercase tracking-tighter">{{ name.substring(0,2) }}</span>
          </div>
          <div>
            <h3 class="text-xl font-black capitalize text-makoclaw-text tracking-tight italic">{{ name }}</h3>
```
Replace with:
```
<div class="flex items-center justify-between mb-5 relative z-10">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-makoclaw-surface border border-makoclaw-border/50 flex items-center justify-center text-makoclaw-accent shadow-md group-hover:scale-105 group-hover:rotate-3 transition-all duration-300">
             <span class="text-xs font-bold uppercase">{{ name.substring(0,2) }}</span>
          </div>
          <div>
            <h3 class="text-base font-semibold capitalize text-makoclaw-text tracking-tight">{{ name }}</h3>
```

**Step 3: Shrink models config button**

Find:
```
class="p-3 rounded-2xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all hover:scale-110 active:scale-95 group/btn"
```
Replace with:
```
class="p-2 rounded-xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all hover:scale-105 active:scale-95 group/btn"
```

**Step 4: Shrink inputs and button**

Find (API key input):
```
class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-6 py-4 text-sm font-bold outline-none focus:border-makoclaw-accent text-makoclaw-text backdrop-blur-sm transition-all"
```
Replace with:
```
class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium outline-none focus:border-makoclaw-accent text-makoclaw-text backdrop-blur-sm transition-all"
```

Find (base URL input):
```
class="w-full bg-makoclaw-bg/20 border-2 border-makoclaw-border/30 rounded-2xl px-6 py-4 text-sm font-bold outline-none focus:border-makoclaw-accent text-makoclaw-text transition-all"
```
Replace with:
```
class="w-full bg-makoclaw-bg/20 border border-makoclaw-border/30 rounded-xl px-4 py-2.5 text-sm font-medium outline-none focus:border-makoclaw-accent text-makoclaw-text transition-all"
```

Find (Connect button):
```
class="w-full px-6 py-4 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-2xl font-black uppercase tracking-widest shadow-xl shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-3 active:scale-95 disabled:opacity-50 group/save"
```
Replace with:
```
class="w-full px-5 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl font-semibold shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 active:scale-95 disabled:opacity-50 group/save"
```

**Step 5: Shrink form labels**

Find (Access Protocol label):
```
<label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Access Protocol (API Key)</label>
```
Replace with:
```
<label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">API Key</label>
```

Find (Gateway Entry label):
```
<label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Gateway Entry (Base URL)</label>
```
Replace with:
```
<label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">Base URL</label>
```

**Step 6: Shrink grid gap and outer spacing**

Find:
```
<div class="grid grid-cols-1 lg:grid-cols-12 gap-8 items-end relative z-10">
```
Replace with:
```
<div class="grid grid-cols-1 lg:grid-cols-12 gap-4 items-end relative z-10">
```

Find:
```
<div class="space-y-8 max-w-4xl mx-auto animate-fade-in-up">
```
Replace with:
```
<div class="space-y-5 max-w-4xl mx-auto animate-fade-in-up">
```

**Step 7: Commit**

```bash
git add pkg/web/frontend/src/components/settings/ProvidersSettingsTab.vue
git commit -m "ui: reduce ProvidersSettingsTab provider card and input proportions"
```

---

## Task 8: ProfileSettingsTab

**File:** `pkg/web/frontend/src/components/settings/ProfileSettingsTab.vue`

**Step 1: Shrink profile card**

Find:
```
<div class="glass-panel rounded-[2rem] p-8 md:p-10 border border-makoclaw-border/50 relative overflow-hidden group">
```
Replace with:
```
<div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group">
```

**Step 2: Shrink section icon and header**

Find:
```
<div class="flex items-center gap-4 mb-10">
           <div class="p-3 rounded-2xl bg-gradient-to-br from-makoclaw-accent to-blue-600 shadow-lg shadow-makoclaw-accent/20">
             <IconProfile class="w-6 h-6 text-white" />
           </div>
           <div>
             <h3 class="text-xs font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70">Bio-Identity</h3>
             <p class="text-[10px] font-bold text-makoclaw-accent uppercase tracking-widest mt-0.5">Core Account Profile</p>
           </div>
```
Replace with:
```
<div class="flex items-center gap-3 mb-5">
           <div class="p-2 rounded-xl bg-gradient-to-br from-makoclaw-accent to-blue-600 shadow-md shadow-makoclaw-accent/20">
             <IconProfile class="w-4 h-4 text-white" />
           </div>
           <div>
             <h3 class="text-xs font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">Profile</h3>
             <p class="text-[10px] font-medium text-makoclaw-accent uppercase tracking-wide mt-0.5">Account Settings</p>
           </div>
```

**Step 3: Shrink inner content spacing**

Find:
```
<div v-else class="space-y-8">
```
Replace with:
```
<div v-else class="space-y-5">
```

**Step 4: Shrink grid gap**

Find:
```
<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
```
Replace with:
```
<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
```

**Step 5: Shrink inputs**

Find (username and email inputs have same class):
```
class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all"
```
Replace with (replace_all: true):
```
class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all"
```

**Step 6: Shrink info grid and labels**

Find:
```
<div class="grid grid-cols-2 gap-6 p-4 rounded-3xl bg-makoclaw-bg/20 border border-makoclaw-border/30">
```
Replace with:
```
<div class="grid grid-cols-2 gap-4 p-3 rounded-xl bg-makoclaw-bg/20 border border-makoclaw-border/30">
```

Find:
```
<span class="text-[9px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40 block mb-1">Authorization</span>
```
Replace with:
```
<span class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/40 block mb-1">Role</span>
```

Find:
```
<span class="text-[9px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40 block mb-1">Joined Matrix</span>
```
Replace with:
```
<span class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/40 block mb-1">Member Since</span>
```

**Step 7: Shrink buttons**

Find:
```
<button @click="saveProfile" :disabled="saving" class="px-8 py-4 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-2xl font-black uppercase tracking-widest shadow-xl shadow-makoclaw-accent/20 transition-all flex items-center justify-center disabled:opacity-50 active:scale-95 group">
```
Replace with:
```
<button @click="saveProfile" :disabled="saving" class="px-5 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl font-semibold shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center justify-center disabled:opacity-50 active:scale-95 group">
```

Find:
```
<button @click="showChangePassword = true" class="px-8 py-4 bg-makoclaw-surface/50 border-2 border-makoclaw-border/50 text-makoclaw-text rounded-2xl font-black uppercase tracking-widest hover:border-makoclaw-accent/40 hover:bg-makoclaw-surface transition-all flex items-center justify-center group active:scale-95">
```
Replace with:
```
<button @click="showChangePassword = true" class="px-5 py-2.5 bg-makoclaw-surface/50 border border-makoclaw-border/50 text-makoclaw-text rounded-xl font-medium hover:border-makoclaw-accent/40 hover:bg-makoclaw-surface transition-all flex items-center justify-center group active:scale-95">
```

Find (button wrapper):
```
<div class="pt-4 grid grid-cols-1 sm:grid-cols-2 gap-4">
```
Replace with:
```
<div class="pt-3 grid grid-cols-1 sm:grid-cols-2 gap-3">
```

**Step 8: Shrink info card at bottom**

Find:
```
<div class="glass-panel rounded-[2rem] p-6 bg-blue-500/5 border border-blue-500/10">
       <div class="flex items-start gap-4">
```
Replace with:
```
<div class="glass-panel rounded-2xl p-4 bg-blue-500/5 border border-blue-500/10">
       <div class="flex items-start gap-3">
```

**Step 9: Shrink outer spacing and form labels**

Find:
```
<div class="space-y-8 max-w-2xl mx-auto animate-fade-in-up">
```
Replace with:
```
<div class="space-y-5 max-w-2xl mx-auto animate-fade-in-up">
```

Any remaining `font-black uppercase tracking-widest` on labels → `font-medium uppercase tracking-wide`.

**Step 10: Commit**

```bash
git add pkg/web/frontend/src/components/settings/ProfileSettingsTab.vue
git commit -m "ui: reduce ProfileSettingsTab card and form proportions"
```

---

## Task 9: ToolPermissionsTab + AuditLogTab

**Files:**
- `pkg/web/frontend/src/components/settings/ToolPermissionsTab.vue`
- `pkg/web/frontend/src/components/settings/AuditLogTab.vue`

### ToolPermissionsTab

**Step 1: Shrink card**

Find:
```
<div class="glass-panel rounded-[2rem] p-8 md:p-10 border border-makoclaw-border/50 relative overflow-hidden group">
```
Replace with:
```
<div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group">
```

**Step 2: Shrink header**

Find:
```
<div class="flex flex-col sm:flex-row items-center justify-between gap-6 mb-10">
          <div class="flex items-center gap-4">
            <div class="p-3 rounded-2xl bg-gradient-to-br from-makoclaw-accent to-indigo-600 shadow-lg shadow-makoclaw-accent/20">
              <IconShield class="w-6 h-6 text-white" />
            </div>
            <div>
              <h3 class="text-[11px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70 italic">Authorization Matrix</h3>
              <p class="text-xs font-bold text-makoclaw-accent uppercase tracking-widest mt-0.5">Role-Based Intelligence Access</p>
```
Replace with:
```
<div class="flex flex-col sm:flex-row items-center justify-between gap-4 mb-5">
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-xl bg-gradient-to-br from-makoclaw-accent to-indigo-600 shadow-md shadow-makoclaw-accent/20">
              <IconShield class="w-4 h-4 text-white" />
            </div>
            <div>
              <h3 class="text-[11px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">Permissions</h3>
              <p class="text-xs font-medium text-makoclaw-accent uppercase tracking-wide mt-0.5">Role-Based Access Control</p>
```

**Step 3: Shrink save button**

Find:
```
class="w-full sm:w-auto px-8 py-4 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-2xl font-black uppercase tracking-widest shadow-xl shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-3 active:scale-95 disabled:opacity-30 group/save"
```
Replace with:
```
class="w-full sm:w-auto px-5 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl font-semibold shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 active:scale-95 disabled:opacity-30 group/save"
```

**Step 4: Shrink info banner and table**

Find:
```
<div class="bg-blue-500/5 border border-blue-500/10 rounded-[1.5rem] p-6 flex items-start gap-4">
```
Replace with:
```
<div class="bg-blue-500/5 border border-blue-500/10 rounded-xl p-4 flex items-start gap-3">
```

Find (table container):
```
<div class="overflow-hidden rounded-[2rem] border border-makoclaw-border/50 bg-makoclaw-bg/20">
```
Replace with:
```
<div class="overflow-hidden rounded-xl border border-makoclaw-border/50 bg-makoclaw-bg/20">
```

Find (table header cells — there may be several, use replace_all):
```
class="px-8 py-5 text-[10px] font-black uppercase tracking-[0.3em] text-makoclaw-text-secondary/50"
```
Replace with:
```
class="px-5 py-3 text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/50"
```

Find (remaining th with center alignment):
```
class="px-6 py-5 text-center text-[10px] font-black uppercase tracking-[0.3em] text-makoclaw-text-secondary/50"
```
Replace with (replace_all: true):
```
class="px-4 py-3 text-center text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/50"
```

**Step 5: Shrink `space-y-10` inside card**

Find:
```
<div v-else class="space-y-10">
```
Replace with:
```
<div v-else class="space-y-6">
```

Find:
```
<div class="space-y-8 max-w-6xl mx-auto animate-fade-in-up">
```
Replace with:
```
<div class="space-y-5 max-w-6xl mx-auto animate-fade-in-up">
```

### AuditLogTab

**Step 6: Shrink filter card**

Find:
```
<div class="glass-panel rounded-[2rem] p-8 border border-makoclaw-border/50 relative overflow-hidden group">
```
Replace with:
```
<div class="glass-panel rounded-2xl p-5 border border-makoclaw-border/50 relative overflow-hidden group">
```

**Step 7: Shrink filter title and grid**

Find:
```
<h3 class="text-[11px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70 italic mb-8 flex items-center gap-3">
```
Replace with:
```
<h3 class="text-[11px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/70 mb-5 flex items-center gap-2">
```

Find (filter grid):
```
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
```
Replace with:
```
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
```

**Step 8: Shrink filter selects and button**

Find (filter selects — same class x3):
```
class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer"
```
Replace with (replace_all: true):
```
class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer"
```

Find (Apply Trace button):
```
class="w-full px-6 py-3.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-2xl font-black uppercase tracking-widest shadow-xl shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 active:scale-95"
```
Replace with:
```
class="w-full px-5 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl font-semibold shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 active:scale-95"
```

**Step 9: Shrink audit table card and header**

Find:
```
<div class="glass-panel rounded-[2.5rem] p-1 border border-makoclaw-border/50 overflow-hidden relative">
      <div class="p-8 pb-4 flex items-center justify-between">
        <div>
          <h3 class="text-xl font-black text-makoclaw-text italic flex items-center gap-3">
             <IconAudit class="w-6 h-6 text-makoclaw-accent" />
```
Replace with:
```
<div class="glass-panel rounded-2xl border border-makoclaw-border/50 overflow-hidden relative">
      <div class="p-5 pb-3 flex items-center justify-between">
        <div>
          <h3 class="text-base font-semibold text-makoclaw-text flex items-center gap-2">
             <IconAudit class="w-4 h-4 text-makoclaw-accent" />
```

**Step 10: Shrink outer spacing**

Find:
```
<div class="space-y-8 max-w-7xl mx-auto animate-fade-in-up">
```
Replace with:
```
<div class="space-y-5 max-w-7xl mx-auto animate-fade-in-up">
```

**Step 11: Commit**

```bash
git add pkg/web/frontend/src/components/settings/ToolPermissionsTab.vue pkg/web/frontend/src/components/settings/AuditLogTab.vue
git commit -m "ui: reduce ToolPermissionsTab and AuditLogTab visual weight"
```

---

## Task 10: SpecialistFormModal

**File:** `pkg/web/frontend/src/components/settings/SpecialistFormModal.vue`

**Step 1: Shrink modal container**

Find:
```
class="glass-panel border border-makoclaw-border/50 rounded-[2.5rem] shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col relative animate-zoom"
```
Replace with:
```
class="glass-panel border border-makoclaw-border/50 rounded-2xl shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col relative animate-zoom"
```

**Step 2: Shrink modal header**

Find:
```
<div class="relative z-10 flex justify-between items-center p-8 pb-4">
            <div>
              <h3 class="text-2xl font-black text-makoclaw-text italic tracking-tight flex items-center gap-3">
```
Replace with:
```
<div class="relative z-10 flex justify-between items-center p-5 pb-3">
            <div>
              <h3 class="text-xl font-semibold text-makoclaw-text flex items-center gap-2">
```

**Step 3: Shrink modal title icon**

Find (in h3 gap-3 — icon inside the title):
```
<div :class="`p-2 rounded-xl bg-gradient-to-br ${mode === 'create' ? 'from-makoclaw-accent to-indigo-600' : 'from-blue-500 to-cyan-500'} shadow-lg shadow-makoclaw-accent/20 text-white` ">
                  <IconPlus v-if="mode === 'create'" class="w-5 h-5" />
                  <IconEdit v-else class="w-5 h-5" />
```
Replace with:
```
<div :class="`p-1.5 rounded-lg bg-gradient-to-br ${mode === 'create' ? 'from-makoclaw-accent to-indigo-600' : 'from-blue-500 to-cyan-500'} shadow-md text-white` ">
                  <IconPlus v-if="mode === 'create'" class="w-4 h-4" />
                  <IconEdit v-else class="w-4 h-4" />
```

**Step 4: Shrink subtitle**

Find:
```
<p class="text-[10px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/40 mt-1">
```
Replace with:
```
<p class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/40 mt-1">
```

**Step 5: Shrink close button**

Find:
```
<button @click="close" class="p-3 rounded-2xl bg-makoclaw-bg/60 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-white transition-all hover:scale-110 active:scale-90">
              <IconClose class="w-6 h-6" />
```
Replace with:
```
<button @click="close" class="p-2 rounded-xl bg-makoclaw-bg/60 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-white transition-all hover:scale-105 active:scale-90">
              <IconClose class="w-5 h-5" />
```

**Step 6: Shrink modal content spacing**

Find:
```
<div class="flex-1 overflow-y-auto p-8 pt-4 space-y-10 custom-scrollbar relative z-10 text-makoclaw-text">
```
Replace with:
```
<div class="flex-1 overflow-y-auto p-5 pt-3 space-y-6 custom-scrollbar relative z-10 text-makoclaw-text">
```

**Step 7: Shrink AI generation box**

Find:
```
<div class="relative bg-makoclaw-surface/40 backdrop-blur-md border border-makoclaw-accent/20 rounded-[2rem] p-8 overflow-hidden">
                <div class="flex flex-col sm:flex-row items-start gap-6">
                  <div class="w-14 h-14 rounded-2xl bg-makoclaw-accent/20 flex items-center justify-center flex-shrink-0 animate-pulse border border-makoclaw-accent/30">
                    <IconBrain class="w-7 h-7 text-makoclaw-accent" />
                  </div>
```
Replace with:
```
<div class="relative bg-makoclaw-surface/40 backdrop-blur-md border border-makoclaw-accent/20 rounded-2xl p-5 overflow-hidden">
                <div class="flex flex-col sm:flex-row items-start gap-4">
                  <div class="w-10 h-10 rounded-xl bg-makoclaw-accent/20 flex items-center justify-center flex-shrink-0 animate-pulse border border-makoclaw-accent/30">
                    <IconBrain class="w-5 h-5 text-makoclaw-accent" />
                  </div>
```

**Step 8: Shrink AI title and textarea**

Find:
```
<h4 class="text-xs font-black uppercase tracking-widest text-makoclaw-text">Neural Synthesis</h4>
```
Replace with:
```
<h4 class="text-xs font-semibold uppercase tracking-wide text-makoclaw-text">AI Generate</h4>
```

Find (textarea):
```
class="w-full px-6 py-4 bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl text-sm font-medium focus:border-makoclaw-accent outline-none text-makoclaw-text placeholder:text-makoclaw-text-secondary/30 transition-all resize-none shadow-inner"
```
Replace with:
```
class="w-full px-4 py-3 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm font-medium focus:border-makoclaw-accent outline-none text-makoclaw-text placeholder:text-makoclaw-text-secondary/30 transition-all resize-none"
```

**Step 9: Shrink AI generate button**

Find:
```
class="w-full px-8 py-4 bg-gradient-to-r from-makoclaw-accent to-blue-600 hover:from-makoclaw-accent-hover hover:to-blue-700 text-white rounded-2xl font-black uppercase tracking-[0.2em] shadow-2xl shadow-makoclaw-accent/30 transition-all flex items-center justify-center gap-3 disabled:opacity-30 disabled:cursor-not-allowed group/btn"
```
Replace with:
```
class="w-full px-5 py-2.5 bg-gradient-to-r from-makoclaw-accent to-blue-600 hover:from-makoclaw-accent-hover hover:to-blue-700 text-white rounded-xl font-semibold shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 disabled:opacity-30 disabled:cursor-not-allowed group/btn"
```

**Step 10: Apply standard input/label/button reductions to the rest of the modal form**

In the rest of the file (form fields for name, model, tools, system prompt, etc.) apply the same token rules:
- `border-2 ... rounded-2xl px-5 py-3.5` → `border ... rounded-xl px-4 py-2.5`
- `font-black uppercase tracking-widest` on labels → `font-medium uppercase tracking-wide`
- `py-4 px-8 rounded-2xl font-black uppercase tracking-widest` on buttons → `py-2.5 px-5 rounded-xl font-semibold`
- `space-y-8`, `gap-8` → `space-y-5`, `gap-5`
- `rounded-[2rem]` → `rounded-2xl`

**Step 11: Commit**

```bash
git add pkg/web/frontend/src/components/settings/SpecialistFormModal.vue
git commit -m "ui: reduce SpecialistFormModal proportions and visual weight"
```

---

## Task 11: Final Verification

**Step 1: Run dev server**

```bash
cd pkg/web/frontend && npm run dev
```

**Step 2: Verify each screen**

Check these pages in order:
1. `/` (Dashboard) — banner compact, stats smaller, launchpad proportional
2. `/settings` → Profile tab — card compact, inputs slim
3. `/settings` → Agents tab — icon containers small, orchestrator fields clean
4. `/settings` → Channels tab — channel cards proportional, icons not huge
5. `/settings` → Providers tab — provider cards clean
6. `/settings` → Permissions tab (admin) — table clean
7. `/settings` → Audit tab (admin) — filters compact
8. Agents tab → "Add Specialist" modal — modal compact

**Step 3: Check for any missed instances**

Search for remaining `rounded-[2rem]`:
```bash
grep -r "rounded-\[2rem\]" pkg/web/frontend/src/views/DashboardView.vue pkg/web/frontend/src/views/SettingsView.vue pkg/web/frontend/src/components/settings/
```
Replace any remaining with `rounded-2xl`.

Search for remaining `font-black uppercase tracking-widest` on non-title elements:
```bash
grep -n "tracking-widest" pkg/web/frontend/src/views/DashboardView.vue pkg/web/frontend/src/components/settings/*.vue
```
Replace any remaining labels/buttons with `tracking-wide` + `font-medium`/`font-semibold` as appropriate.

**Step 4: Final commit**

```bash
git add pkg/web/frontend/src/views/ pkg/web/frontend/src/components/settings/
git commit -m "ui: final polish — catch any remaining oversized tokens across dashboard and settings"
```
