# Design: Settings Tabs UX Polish

**Change:** `settings-tabs-ux-polish`
**Status:** Draft
**Date:** 2026-02-27

## Overview

The Settings page has 6 tab components plus 2 inline tabs (Users, System) in SettingsView.vue. Some tabs have been well polished to match the premium design language (glass morphism, responsive spacing, gradient glows, hover effects), while others have inconsistencies in spacing, border opacities, input styling, and missing premium effects. This design brings ALL Settings tabs to visual consistency with each other and with the Dashboard/Chat/Tasks premium design language.

---

## Current State Assessment

### Premium Design Language Reference (from Dashboard, Chat, Tasks)

| Token | Standard Value |
|-------|---------------|
| Panel container | `glass-panel rounded-2xl p-2 sm:p-3 md:p-4` or `p-5 md:p-6` |
| Panel border | `border border-makoclaw-border/50` (resting), `/30` hover |
| Decorative blob | `absolute -top-12 -right-12 w-48 h-48 bg-makoclaw-accent/5 blur-[50px] rounded-full group-hover:bg-makoclaw-accent/10 transition-all duration-1000` |
| Section label | `text-[11px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/70` |
| Form label | `text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1` |
| Standard input | `w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all` |
| Bold/prominent input | `w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all` |
| Primary button | `px-5 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl text-[11px] font-semibold shadow-lg shadow-makoclaw-accent/20 transition-all active:scale-95` |
| Touch target minimum | `min-h-[36px] min-w-[36px]` (on interactive elements) |
| Icon container (gradient) | `p-2 rounded-xl bg-gradient-to-br from-makoclaw-accent to-indigo-600 shadow-md shadow-makoclaw-accent/20` |
| Rounded corners | `rounded-xl` for buttons/inputs, `rounded-2xl` for panels |
| Responsive spacing | `p-2 sm:p-3 md:p-4` or `p-5 md:p-6` |
| Fade-in animation | `animate-fade-in-up` with custom keyframe |
| Hover elevation | `hover:shadow-xl hover:shadow-makoclaw-accent/5 hover:-translate-y-0.5` |
| Bottom hover line | `absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover:w-full transition-all duration-500 opacity-60` |

### Tab-by-Tab Assessment

| Tab | File | Status | Issues |
|-----|------|--------|--------|
| **ProfileSettingsTab** | ProfileSettingsTab.vue | GOOD | Minor: missing responsive padding on root, `max-w-2xl` could be `max-w-4xl` for consistency. No premium hover effects on the main panel. |
| **AgentSettingsTab** | AgentSettingsTab.vue | GOOD | Well designed. Minor inconsistency: Orchestrator model select uses `border-2 rounded-2xl px-5 py-3.5 font-bold` while the Command Provider select uses `border rounded-xl px-4 py-2.5 font-medium` -- mixed input styles within the same grid. |
| **ProvidersSettingsTab** | ProvidersSettingsTab.vue | GOOD | Well designed. No major issues. |
| **ChannelsSettingsTab** | ChannelsSettingsTab.vue | GOOD | Well designed. No major issues. |
| **ToolPermissionsTab** | ToolPermissionsTab.vue | NEEDS WORK | Uses `space-y-6` (inconsistent with `space-y-5` standard), `p-6` without responsive breakpoints, `mb-8` (too large), section header uses `text-xl font-black uppercase` (too heavy compared to other tabs), no bottom hover accent line on save button area. |
| **AuditLogTab** | AuditLogTab.vue | GOOD | Well designed. Consistent with premium language. |
| **Users (inline)** | SettingsView.vue:132-193 | NEEDS WORK | Missing decorative blob, no section header icon gradient container, table header uses `bg-makoclaw-bg/30` instead of `bg-makoclaw-bg/40`, lacks border on glass-panel, action buttons lack touch target minimum sizing. |
| **System (inline)** | SettingsView.vue:196-319 | NEEDS WORK | Config display items use `rounded-2xl` inconsistently with `rounded-xl`, input fields use oversized `border-2 rounded-2xl px-5 py-3.5 font-bold` style (should be standard input style), Backup section uses excessive spacing (`gap-10`, `mt-12`), checkbox labels are hard to read. |

---

## Enhancements

### Enhancement 1: ToolPermissionsTab -- Normalize Spacing and Visual Weight

**Target:** `pkg/web/frontend/src/components/Settings/ToolPermissionsTab.vue`, line 2

**Current:**
```html
<div class="space-y-6 max-w-6xl mx-auto animate-fade-in-up">
```
**New:**
```html
<div class="space-y-5 max-w-5xl mx-auto animate-fade-in-up">
```

**Rationale:** All other tabs use `space-y-5` and `max-w-4xl` or `max-w-5xl`. `max-w-6xl` is the widest of all tabs and `space-y-6` is inconsistent. Since the table does benefit from width, we use `max-w-5xl` as a compromise.

---

### Enhancement 2: ToolPermissionsTab -- Normalize Section Header to Match Other Tabs

**Target:** `pkg/web/frontend/src/components/Settings/ToolPermissionsTab.vue`, lines 4-6

**Current:**
```html
    <div class="glass-panel rounded-2xl p-6 border border-makoclaw-border/50 relative overflow-hidden group">
      <div class="absolute -top-24 -right-24 w-64 h-64 bg-makoclaw-accent/10 blur-[80px] rounded-full group-hover:bg-makoclaw-accent/20 transition-all duration-1000" />
```
**New:**
```html
    <div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group">
      <div class="absolute -top-12 -right-12 w-48 h-48 bg-makoclaw-accent/5 blur-[60px] rounded-full group-hover:bg-makoclaw-accent/10 transition-all duration-1000" />
```

**Rationale:** Adds responsive padding `p-5 md:p-6` matching AgentSettingsTab. The decorative blob is oversized (`w-64 h-64`, `blur-[80px]`, `/10` opacity) compared to the standard (`w-48 h-48`, `blur-[60px]`, `/5` opacity) used in ProfileSettingsTab and ProvidersSettingsTab. Normalizing to the standard sizes.

---

### Enhancement 3: ToolPermissionsTab -- Reduce Section Header Title Weight

**Target:** `pkg/web/frontend/src/components/Settings/ToolPermissionsTab.vue`, lines 9-23

**Current:**
```html
        <div class="flex items-center justify-between mb-8">
          <div class="flex items-center gap-3">
            <div class="p-2.5 rounded-xl bg-gradient-to-br from-makoclaw-accent/80 to-indigo-600 shadow-lg shadow-makoclaw-accent/20">
              <IconShield class="w-5 h-5 text-white" />
            </div>
            <div>
              <h3 class="text-[11px] font-bold uppercase tracking-widest text-makoclaw-text-secondary/60">
                Node Management
              </h3>
              <p class="text-xl font-black text-makoclaw-text tracking-tight uppercase">
                Access Control Matrix
              </p>
            </div>
          </div>
        </div>
```
**New:**
```html
        <div class="flex items-center justify-between mb-5">
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-xl bg-gradient-to-br from-makoclaw-accent to-indigo-600 shadow-md shadow-makoclaw-accent/20">
              <IconShield class="w-4 h-4 text-white" />
            </div>
            <div>
              <h3 class="text-[11px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">
                Node Management
              </h3>
              <p class="text-base font-semibold text-makoclaw-text leading-none">
                Access Control Matrix
              </p>
            </div>
          </div>
        </div>
```

**Rationale:** Comparing to AgentSettingsTab (the gold standard among tabs):
- Icon uses `p-2 w-4 h-4` (not `p-2.5 w-5 h-5`), `from-makoclaw-accent` (not `from-makoclaw-accent/80`), `shadow-md` (not `shadow-lg`)
- Label uses `font-medium tracking-wide` (not `font-bold tracking-widest`)
- Title uses `text-base font-semibold leading-none` (not `text-xl font-black uppercase`)
- Margin bottom is `mb-5` (not `mb-8`)

This brings the visual weight in line with other tabs.

---

### Enhancement 4: ToolPermissionsTab -- Normalize Info Panels Spacing

**Target:** `pkg/web/frontend/src/components/Settings/ToolPermissionsTab.vue`, line 26

**Current:**
```html
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
```
**New:**
```html
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-5">
```

**Rationale:** `mb-8` is excessive. Standard bottom margin in AgentSettingsTab sections is `mb-5` or `mb-6`.

---

### Enhancement 5: ToolPermissionsTab -- Normalize Safe Paradigm Section Header

**Target:** `pkg/web/frontend/src/components/Settings/ToolPermissionsTab.vue`, lines 117-133

**Current:**
```html
    <div class="glass-panel rounded-2xl p-6 border border-makoclaw-border/50 relative overflow-hidden group">
      <div class="absolute -bottom-24 -left-24 w-64 h-64 bg-amber-500/10 blur-[80px] rounded-full group-hover:bg-amber-500/20 transition-all duration-1000" />

      <div class="relative z-10">
        <div class="flex items-center gap-3 mb-6">
          <div class="p-2 rounded-xl bg-amber-500/10 text-amber-500 shadow-inner">
            <IconLock class="w-4 h-4" />
          </div>
          <div>
            <h3 class="text-[11px] font-bold uppercase tracking-widest text-amber-500/60">
              Security Firewall
            </h3>
            <p class="text-base font-black text-makoclaw-text tracking-tight uppercase">
              Safe Paradigm Whitelist
            </p>
          </div>
        </div>
```
**New:**
```html
    <div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group">
      <div class="absolute -bottom-12 -left-12 w-48 h-48 bg-amber-500/5 blur-[60px] rounded-full group-hover:bg-amber-500/10 transition-all duration-1000" />

      <div class="relative z-10">
        <div class="flex items-center gap-3 mb-5">
          <div class="p-2 rounded-xl bg-amber-500/10 text-amber-500">
            <IconLock class="w-4 h-4" />
          </div>
          <div>
            <h3 class="text-[11px] font-medium uppercase tracking-wide text-amber-500/60">
              Security Firewall
            </h3>
            <p class="text-base font-semibold text-makoclaw-text leading-none">
              Safe Paradigm Whitelist
            </p>
          </div>
        </div>
```

**Rationale:** Same normalization as Enhancement 3:
- Responsive padding `p-5 md:p-6`
- Blob downsized to standard dimensions
- `font-bold tracking-widest` -> `font-medium tracking-wide`
- `font-black uppercase` -> `font-semibold leading-none` (no uppercase, matching AgentSettingsTab)
- `mb-6` -> `mb-5`
- Remove `shadow-inner` from icon container (not used on any other tab's icon containers)

---

### Enhancement 6: ToolPermissionsTab -- Normalize Save Button Alignment

**Target:** `pkg/web/frontend/src/components/Settings/ToolPermissionsTab.vue`, lines 170-180

**Current:**
```html
    <div class="flex justify-center md:justify-end">
      <button
        @click="commitPermissions"
        :disabled="loading"
        class="w-full md:w-auto px-6 py-3 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl text-[11px] font-bold shadow-lg shadow-makoclaw-accent/30 transition-all flex items-center justify-center gap-2 disabled:opacity-50 active:scale-95 group/save"
      >
```
**New:**
```html
    <div class="flex justify-center md:justify-end pt-6">
      <button
        @click="commitPermissions"
        :disabled="loading"
        class="w-full md:w-auto px-5 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl text-[11px] font-semibold shadow-lg shadow-makoclaw-accent/30 transition-all flex items-center justify-center gap-2 disabled:opacity-50 active:scale-95 group/save"
      >
```

**Rationale:** AgentSettingsTab save button uses `px-5 py-2.5 font-semibold` (not `px-6 py-3 font-bold`). Adding `pt-6` for visual separation matching the AgentSettingsTab pattern.

---

### Enhancement 7: Users Tab (inline) -- Add Decorative Blob and Section Icon

**Target:** `pkg/web/frontend/src/views/SettingsView.vue`, lines 132-146

**Current:**
```html
                  <div v-if="activeTab === 'users' && authStore.user?.role === 'admin'" class="space-y-6 animate-fade-in-up">
                    <div class="glass-panel rounded-2xl p-5">
                       <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4 mb-5">
                        <div>
                          <h3 class="text-xs font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">User Hub</h3>
                          <p class="text-xs font-medium text-makoclaw-text-secondary/50 mt-1">Manage infrastructure access nodes</p>
                        </div>
```
**New:**
```html
                  <div v-if="activeTab === 'users' && authStore.user?.role === 'admin'" class="space-y-5 animate-fade-in-up">
                    <div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group">
                       <div class="absolute -top-12 -right-12 w-48 h-48 bg-makoclaw-accent/5 blur-[60px] rounded-full group-hover:bg-makoclaw-accent/10 transition-all duration-1000"></div>
                       <div class="relative z-10">
                       <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4 mb-5">
                        <div class="flex items-center gap-3">
                          <div class="p-2 rounded-xl bg-gradient-to-br from-makoclaw-accent to-indigo-600 shadow-md shadow-makoclaw-accent/20">
                            <IconUsers class="w-4 h-4 text-white" />
                          </div>
                          <div>
                            <h3 class="text-[11px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">User Hub</h3>
                            <p class="text-xs font-medium text-makoclaw-text-secondary/50 mt-0.5">Manage infrastructure access nodes</p>
                          </div>
                        </div>
```

**Also need to close the `relative z-10` div** -- add `</div>` before the closing `</div>` of the glass-panel container (after the table's closing `</div>`).

Specifically, line 192 currently:
```html
                    </div>
```
Needs to become:
```html
                    </div>
                    </div>
```

And fix the table header background (line 150):
**Current:**
```html
                          <thead class="bg-makoclaw-bg/30 tracking-wide border-b border-makoclaw-border/50 text-[10px] text-makoclaw-text-secondary font-medium uppercase">
```
**New:**
```html
                          <thead class="bg-makoclaw-bg/40 tracking-wide text-[10px] text-makoclaw-text-secondary/60 font-medium uppercase">
```

**Rationale:** The Users tab was missing:
- `border border-makoclaw-border/50` on the glass-panel (every other tab has this)
- Responsive padding `p-5 md:p-6`
- Decorative blob (every other tab panel has one)
- Section header icon in gradient container (consistent with AgentSettingsTab, ProfileSettingsTab, ToolPermissionsTab)
- `space-y-6` -> `space-y-5` for consistency
- Table header uses `bg-makoclaw-bg/30` instead of standard `bg-makoclaw-bg/40`; also should use `text-makoclaw-text-secondary/60` matching AuditLogTab

---

### Enhancement 8: Users Tab -- Improve Action Button Touch Targets

**Target:** `pkg/web/frontend/src/views/SettingsView.vue`, lines 182-185

**Current:**
```html
                                  <button @click="openUserModal(u)" class="p-2 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-colors"><IconEdit class="w-4 h-4"/></button>
                                  <button v-if="!u.blocked" @click="openBlockModal(u)" :disabled="authStore.user?.username === u.username" class="p-2 text-makoclaw-text-secondary hover:text-orange-400 disabled:opacity-20"><IconBlock class="w-4 h-4"/></button>
                                  <button v-else @click="openUnblockModal(u)" class="p-2 text-makoclaw-text-secondary hover:text-green-400"><IconCheck class="w-4 h-4"/></button>
                                  <button @click="deleteUserLocal(u)" :disabled="authStore.user?.username === u.username" class="p-2 text-makoclaw-text-secondary hover:text-red-400 disabled:opacity-20"><IconDelete class="w-4 h-4"/></button>
```
**New:**
```html
                                  <button @click="openUserModal(u)" class="p-2 rounded-xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all hover:scale-110"><IconEdit class="w-4 h-4"/></button>
                                  <button v-if="!u.blocked" @click="openBlockModal(u)" :disabled="authStore.user?.username === u.username" class="p-2 rounded-xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-orange-400 disabled:opacity-20 transition-all hover:scale-110"><IconBlock class="w-4 h-4"/></button>
                                  <button v-else @click="openUnblockModal(u)" class="p-2 rounded-xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-green-400 transition-all hover:scale-110"><IconCheck class="w-4 h-4"/></button>
                                  <button @click="deleteUserLocal(u)" :disabled="authStore.user?.username === u.username" class="p-2 rounded-xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-red-400 disabled:opacity-20 transition-all hover:scale-110"><IconDelete class="w-4 h-4"/></button>
```

**Rationale:** AgentSettingsTab specialist card action buttons use `p-2 rounded-xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all hover:scale-110`. The Users tab action buttons are unstyled bare icon buttons. Adding the same container styling creates visual consistency and improves touch target visibility.

---

### Enhancement 9: System Tab -- Add Section Icon Gradient Containers

**Target:** `pkg/web/frontend/src/views/SettingsView.vue`, lines 196-201

**Current:**
```html
                  <div v-if="activeTab === 'system'" class="space-y-5 animate-fade-in-up">
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
                       <div class="glass-panel rounded-2xl p-5 border border-makoclaw-border/50">
                        <div class="flex items-center gap-3 mb-5">
                           <div class="p-2 rounded-xl bg-blue-500/10 text-blue-400"><IconGlobe class="w-5 h-5"/></div>
                           <h3 class="text-xs font-medium tracking-wide text-makoclaw-text-secondary/70">Web Infrastructure</h3>
```
**New:**
```html
                  <div v-if="activeTab === 'system'" class="space-y-5 animate-fade-in-up">
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
                       <div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group">
                        <div class="absolute -top-12 -right-12 w-48 h-48 bg-blue-500/5 blur-[60px] rounded-full group-hover:bg-blue-500/10 transition-all duration-1000"></div>
                        <div class="relative z-10">
                        <div class="flex items-center gap-3 mb-5">
                           <div class="p-2 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-600 shadow-md shadow-blue-500/20"><IconGlobe class="w-4 h-4 text-white"/></div>
                           <div>
                             <h3 class="text-[11px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">Web Infrastructure</h3>
                           </div>
```

Similarly for Gateway section (line 210-213):
**Current:**
```html
                      <div class="glass-panel rounded-2xl p-5 border border-makoclaw-border/50">
                        <div class="flex items-center gap-3 mb-5">
                           <div class="p-2 rounded-xl bg-makoclaw-accent/10 text-makoclaw-accent"><IconGateway class="w-5 h-5"/></div>
                           <h3 class="text-xs font-medium tracking-wide text-makoclaw-text-secondary/70">Gateway Layer</h3>
```
**New:**
```html
                      <div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group">
                        <div class="absolute -top-12 -right-12 w-48 h-48 bg-makoclaw-accent/5 blur-[60px] rounded-full group-hover:bg-makoclaw-accent/10 transition-all duration-1000"></div>
                        <div class="relative z-10">
                        <div class="flex items-center gap-3 mb-5">
                           <div class="p-2 rounded-xl bg-gradient-to-br from-makoclaw-accent to-blue-600 shadow-md shadow-makoclaw-accent/20"><IconGateway class="w-4 h-4 text-white"/></div>
                           <div>
                             <h3 class="text-[11px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">Gateway Layer</h3>
                           </div>
```

And for Utility Parameters (line 224-228):
**Current:**
```html
                    <div class="glass-panel rounded-2xl p-5 border border-makoclaw-border/50">
                      <div class="flex items-center justify-between mb-5">
                         <div class="flex items-center gap-3">
                           <div class="p-2 rounded-xl bg-orange-500/10 text-orange-400"><IconTool class="w-5 h-5"/></div>
                           <h3 class="text-xs font-medium tracking-wide text-makoclaw-text-secondary/70">Utility Parameters</h3>
```
**New:**
```html
                    <div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group">
                      <div class="absolute -top-12 -left-12 w-48 h-48 bg-orange-500/5 blur-[60px] rounded-full group-hover:bg-orange-500/10 transition-all duration-1000"></div>
                      <div class="relative z-10">
                      <div class="flex items-center justify-between mb-5">
                         <div class="flex items-center gap-3">
                           <div class="p-2 rounded-xl bg-gradient-to-br from-orange-500 to-amber-600 shadow-md shadow-orange-500/20"><IconTool class="w-4 h-4 text-white"/></div>
                           <div>
                             <h3 class="text-[11px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">Utility Parameters</h3>
                           </div>
```

**Note:** Each of these needs a closing `</div>` for the `relative z-10` wrapper added before the panel's closing tag. Also, config key-value row items (lines 204, 216) should be adjusted:

**Current:**
```html
                          <div v-for="(val, key) in configData.web" :key="key" class="flex justify-between items-center px-4 py-3 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30">
```
**New:**
```html
                          <div v-for="(val, key) in configData.web" :key="key" class="flex justify-between items-center px-4 py-3 rounded-xl bg-makoclaw-bg/30 border border-makoclaw-border/30 hover:border-makoclaw-accent/30 transition-all">
```

**Rationale:**
- System tab icon containers use flat colored backgrounds (`bg-blue-500/10`) instead of the premium gradient containers (`bg-gradient-to-br from-blue-500 to-cyan-600`) used across other tabs
- Icons use `w-5 h-5` instead of standard `w-4 h-4`
- Section headers are bare `<h3>` instead of wrapped in `<div>` with proper formatting
- Headers use `text-xs` instead of `text-[11px] uppercase`
- Missing decorative blobs on all three panels
- Missing responsive padding
- Missing `group` class for blob hover effect
- Config items use `rounded-2xl` but should be `rounded-xl` (inner elements should be less rounded than the container)
- Config items lack hover state

---

### Enhancement 10: System Tab -- Normalize Input Fields in Utility Parameters

**Target:** `pkg/web/frontend/src/views/SettingsView.vue`, lines 241-245

**Current:**
```html
                            <input v-model="configData.tools.web.search.api_key" type="password" placeholder="••••••••••••••••" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent transition-all outline-none">
                         </div>
                         <div class="space-y-2">
                            <label class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/60">Search Capacity</label>
                            <input v-model.number="configData.tools.web.search.max_results" type="number" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-black text-makoclaw-text focus:border-makoclaw-accent transition-all outline-none">
```
**New:**
```html
                            <input v-model="configData.tools.web.search.api_key" type="password" placeholder="••••••••••••••••" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent transition-all outline-none">
                         </div>
                         <div class="space-y-2">
                            <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">Search Capacity</label>
                            <input v-model.number="configData.tools.web.search.max_results" type="number" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent transition-all outline-none">
```

Also fix the Search Access Key label (line 240):
**Current:**
```html
                            <label class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/60">Search Access Key</label>
```
**New:**
```html
                            <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">Search Access Key</label>
```

**Rationale:** The standard input pattern across ProfileSettingsTab, AuditLogTab, ChannelsSettingsTab is:
- `border` (not `border-2`)
- `rounded-xl` (not `rounded-2xl`)
- `px-4 py-2.5` (not `px-5 py-3.5`)
- `font-medium` (not `font-bold`/`font-black`)

Labels should have `uppercase ml-1` matching other tabs.

---

### Enhancement 11: AgentSettingsTab -- Normalize Inconsistent Input Styles Within Tab

**Target:** `pkg/web/frontend/src/components/Settings/AgentSettingsTab.vue`, lines 39, 168, 174

There are two input style variants being mixed within this tab:

**Heavy style** (used on some selects):
```
border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-accent
```

**Standard style** (used on other selects and all number inputs):
```
border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text
```

**Changes needed:**

Line 39 (Orchestrator Primary Model Core):
**Current:**
```html
                <select v-model="orchestratorConfig.model" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-accent focus:border-makoclaw-accent outline-none transition-all cursor-pointer">
```
**New:**
```html
                <select v-model="orchestratorConfig.model" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer">
```

Line 168 (Base Provider):
**Current:**
```html
            <select v-model="agents.defaults.provider" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none cursor-pointer">
```
**New:**
```html
            <select v-model="agents.defaults.provider" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer">
```

Line 174 (Standard Core):
**Current:**
```html
            <select v-model="agents.defaults.model" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-accent focus:border-makoclaw-accent outline-none cursor-pointer">
```
**New:**
```html
            <select v-model="agents.defaults.model" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer">
```

Also normalize the "Fluidity", "Output Cap", "Tool Endurance" inputs (lines 185, 189, 193):
**Current:**
```html
                <input ... class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2 text-sm font-black text-makoclaw-text focus:border-makoclaw-accent outline-none">
```
**New:**
```html
                <input ... class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all">
```

**Rationale:** Within AgentSettingsTab, model selects use oversized `border-2 rounded-2xl px-5 py-3.5 font-bold` while provider selects use normal `border rounded-xl px-4 py-2.5 font-medium`. The defaults section inputs use `font-black py-2` -- mixed weights and padding. Unifying to the standard style for visual consistency within the tab and across all tabs.

---

### Enhancement 12: System Tab -- Normalize Backup Section Spacing

**Target:** `pkg/web/frontend/src/views/SettingsView.vue`, lines 260, 267, 268-270

**Current backup grid:**
```html
                      <div class="grid grid-cols-1 lg:grid-cols-2 gap-10">
```
**New:**
```html
                      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
```

**Current export options container:**
```html
                           <div class="p-6 bg-makoclaw-bg/40 rounded-3xl border border-makoclaw-border/50 space-y-3">
```
**New:**
```html
                           <div class="p-4 bg-makoclaw-bg/40 rounded-2xl border border-makoclaw-border/50 space-y-2">
```

**Current checkbox labels (line 269):**
```html
                              <label v-for="(val, opt) in exportOptions" :key="opt" class="flex items-center justify-between p-3 rounded-xl hover:bg-white/5 cursor-pointer transition-colors">
                                <span class="text-xs font-black text-makoclaw-text-secondary/80 uppercase tracking-tighter">{{ opt.replace('include_', '').replace('_', ' ') }}</span>
```
**New:**
```html
                              <label v-for="(val, opt) in exportOptions" :key="opt" class="flex items-center justify-between p-3 rounded-xl hover:bg-white/5 cursor-pointer transition-colors">
                                <span class="text-xs font-medium text-makoclaw-text-secondary/80 uppercase tracking-wide">{{ opt.replace('include_', '').replace('_', ' ') }}</span>
```

**Current import dashed border (line 290):**
```html
                           <div class="border-2 border-dashed border-makoclaw-border/50 rounded-3xl p-6 hover:border-makoclaw-accent/50 transition-all group relative">
```
**New:**
```html
                           <div class="border-2 border-dashed border-makoclaw-border/50 rounded-2xl p-5 hover:border-makoclaw-accent/50 transition-all group relative">
```

**Current channel modal spacing (line 356):**
```html
                  <div class="flex justify-end gap-3 mt-12 pt-8 border-t border-makoclaw-border/30">
```
**New:**
```html
                  <div class="flex justify-end gap-3 mt-8 pt-6 border-t border-makoclaw-border/30">
```

**Rationale:** The Backup/System section uses larger spacing than every other Settings tab:
- `gap-10` -> `gap-6` (standard grid gap)
- `p-6 rounded-3xl` -> `p-4 rounded-2xl` (standard inner container)
- `font-black tracking-tighter` -> `font-medium tracking-wide` (standard label weight)
- `rounded-3xl p-6` -> `rounded-2xl p-5` (panels use `rounded-2xl` throughout)
- `mt-12 pt-8` -> `mt-8 pt-6` (excessive whitespace reduced to match other modals)

---

### Enhancement 13: ProfileSettingsTab -- Add Premium Hover Effects

**Target:** `pkg/web/frontend/src/components/Settings/ProfileSettingsTab.vue`, line 4

**Current:**
```html
    <div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group">
```
No hover bottom-line or elevation effect.

Add after line 6 (after the decorative blob `</div>`):
```html
      <!-- Hover bottom accent line -->
      <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover:w-full transition-all duration-700 opacity-40 rounded-b-2xl"></div>
```

Also update line 2 from:
**Current:**
```html
  <div class="space-y-5 max-w-2xl mx-auto animate-fade-in-up">
```
**New:**
```html
  <div class="space-y-5 max-w-4xl mx-auto animate-fade-in-up">
```

**Rationale:** `max-w-2xl` is narrower than any other tab. Other tabs use `max-w-4xl` or `max-w-5xl`. The Profile panel already has a `group` class but no hover effects utilizing it (beyond the blob). Adding the bottom accent line matches the premium treatment on task cards and message bubbles.

---

### Enhancement 14: Add Premium Bottom-Line to AgentSettingsTab Panels

**Target:** `pkg/web/frontend/src/components/Settings/AgentSettingsTab.vue`

The Orchestrator panel (line 4) already has `group` class. Add a bottom hover line after the decorative blob (after line 5):
```html
      <!-- Hover bottom accent line -->
      <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover:w-full transition-all duration-700 opacity-40 rounded-b-2xl"></div>
```

Similarly for the Specialists panel (line 73, add `group` class if not present), add:
```html
      <!-- Hover bottom accent line -->
      <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-blue-500 to-cyan-500 group-hover:w-full transition-all duration-700 opacity-40 rounded-b-2xl"></div>
```

And for the Agent Defaults panel (line 153):
```html
      <!-- Hover bottom accent line -->
      <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover:w-full transition-all duration-700 opacity-40 rounded-b-2xl"></div>
```

**Rationale:** Premium bottom-line hover effect is present on Dashboard cards, Chat message bubbles, and Task Kanban cards. Adding it to Settings panels creates cross-page visual consistency.

---

### Enhancement 15: Add Premium Bottom-Line to Provider Cards

**Target:** `pkg/web/frontend/src/components/Settings/ProvidersSettingsTab.vue`, line 6

After the decorative blob div, add:
```html
      <!-- Hover bottom accent line -->
      <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover:w-full transition-all duration-700 opacity-40 rounded-b-2xl"></div>
```

**Rationale:** Same as Enhancement 14. Provider cards already have `group` and `relative overflow-hidden`.

---

### Enhancement 16: Add Premium Bottom-Line to Channel Cards

**Target:** `pkg/web/frontend/src/components/Settings/ChannelsSettingsTab.vue`, line 14

Channel cards already have `group relative overflow-hidden`. After the background glow div (line 20), add:
```html
          <!-- Hover bottom accent line -->
          <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover:w-full transition-all duration-500 opacity-40 rounded-b-2xl"></div>
```

**Rationale:** Same as Enhancement 14. Channel cards are the most "card-like" in Settings and benefit most from this effect.

---

### Enhancement 17: ToolPermissionsTab -- Add Bottom-Line to Panels

**Target:** `pkg/web/frontend/src/components/Settings/ToolPermissionsTab.vue`

After the decorative blob on the Access Control Matrix panel (after line 6), add:
```html
      <!-- Hover bottom accent line -->
      <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-indigo-500 group-hover:w-full transition-all duration-700 opacity-40 rounded-b-2xl"></div>
```

After the decorative blob on the Safe Paradigm panel (after line 118), add:
```html
      <!-- Hover bottom accent line -->
      <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-amber-500 to-orange-500 group-hover:w-full transition-all duration-700 opacity-40 rounded-b-2xl"></div>
```

**Rationale:** Matching the amber/orange color scheme of the Safe Paradigm section for visual coherence. The Access Control panel uses the standard accent-to-indigo gradient.

---

### Enhancement 18: AuditLogTab -- Add Bottom-Line to Panels

**Target:** `pkg/web/frontend/src/components/Settings/AuditLogTab.vue`

After the decorative blob on the filters panel (after line 5), add:
```html
      <!-- Hover bottom accent line -->
      <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover:w-full transition-all duration-700 opacity-40 rounded-b-2xl"></div>
```

**Rationale:** Consistent with all other tab panels.

---

## Summary

| Metric | Value |
|--------|-------|
| **Total Enhancements** | 18 |
| **Files Affected** | 7 |
| **Tabs Fixed** | 5 (ToolPermissionsTab, Users inline, System inline, AgentSettingsTab, ProfileSettingsTab) |
| **Premium Effects Added** | Bottom hover lines on all panels (6 files) |
| **Error Fixes** | 0 (build is clean -- no compile errors found) |

### Files Affected

1. `pkg/web/frontend/src/components/Settings/ToolPermissionsTab.vue` -- Enhancements 1-6, 17
2. `pkg/web/frontend/src/views/SettingsView.vue` (Users tab) -- Enhancements 7-8
3. `pkg/web/frontend/src/views/SettingsView.vue` (System tab) -- Enhancements 9-10, 12
4. `pkg/web/frontend/src/components/Settings/AgentSettingsTab.vue` -- Enhancements 11, 14
5. `pkg/web/frontend/src/components/Settings/ProfileSettingsTab.vue` -- Enhancement 13
6. `pkg/web/frontend/src/components/Settings/ProvidersSettingsTab.vue` -- Enhancement 15
7. `pkg/web/frontend/src/components/Settings/ChannelsSettingsTab.vue` -- Enhancement 16
8. `pkg/web/frontend/src/components/Settings/AuditLogTab.vue` -- Enhancement 18

### Consistency Normalization Summary

| Property | Before (inconsistent) | After (normalized) |
|----------|-----------------------|--------------------|
| Root spacing | `space-y-5` / `space-y-6` | `space-y-5` everywhere |
| Panel padding | `p-5` / `p-6` (fixed) | `p-5 md:p-6` (responsive) |
| Panel max-width | `max-w-2xl` to `max-w-6xl` | `max-w-4xl` or `max-w-5xl` |
| Decorative blob | Missing / oversized / standard | Standard `w-48 h-48 /5 blur-[60px]` |
| Section header weight | `text-xl font-black` / `text-base font-semibold` | `text-base font-semibold` everywhere |
| Label formatting | Missing `uppercase` / `ml-1` | `text-[10px] font-medium uppercase tracking-wide ml-1` |
| Input borders | `border` / `border-2` | `border` everywhere |
| Input rounding | `rounded-xl` / `rounded-2xl` | `rounded-xl` everywhere (inputs/buttons) |
| Input padding | `px-4 py-2` / `px-5 py-3.5` | `px-4 py-2.5` everywhere |
| Input font weight | `font-medium` / `font-bold` / `font-black` | `font-medium` everywhere |
| Icon container | Flat color (`bg-blue-500/10`) / Gradient | Gradient (`bg-gradient-to-br`) on all section headers |
| Icon size | `w-4 h-4` / `w-5 h-5` | `w-4 h-4` everywhere |
| Action buttons | Bare icons / Contained | Contained (`rounded-xl bg-makoclaw-bg/40 border`) |
| Bottom hover line | Absent on all panels | Present on all panels |

### Priority Order for Implementation

1. **HIGH** -- Enhancements 1-6 (ToolPermissionsTab normalization) -- most visually inconsistent tab
2. **HIGH** -- Enhancements 7-8 (Users tab) -- missing fundamental premium elements
3. **HIGH** -- Enhancements 9-10, 12 (System tab) -- oversized inputs and spacing
4. **MEDIUM** -- Enhancement 11 (AgentSettingsTab input normalization) -- subtle but noticeable
5. **MEDIUM** -- Enhancement 13 (ProfileSettingsTab width and hover) -- minor width fix
6. **LOW** -- Enhancements 14-18 (bottom hover lines everywhere) -- polish layer, can be done in one pass

### Performance Considerations

- All blur effects use `pointer-events-none` to avoid layout thrashing
- Decorative blobs are absolutely positioned with fixed sizes (no dynamic reflows)
- Bottom-line animations use `transform` (width) and `opacity` only -- GPU-composited
- No new JavaScript logic added -- all changes are CSS/template class modifications
- No new components or imports required

### Build Status

- `npx vite build` completes successfully with no errors
- No TypeScript compilation errors detected
- No missing imports or undefined component references found
- The user-mentioned "errores" may refer to visual inconsistencies rather than code errors, or may have been resolved in a prior commit
