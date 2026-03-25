<template>
  <div class="space-y-6 animate-fade-in-up">

    <!-- AI Generation Card -->
    <div class="bg-makoclaw-surface/30 border border-makoclaw-border/50 rounded-xl p-5 hover:border-makoclaw-accent/20 transition-all">
      <div class="flex items-start gap-4">
        <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-makoclaw-accent/20 to-indigo-500/20 flex items-center justify-center flex-shrink-0 ring-1 ring-white/10">
          <svg class="w-5 h-5 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
          </svg>
        </div>
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2 mb-1">
            <h4 class="text-sm font-bold text-makoclaw-text">AI Soul Generator</h4>
            <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold bg-makoclaw-accent/15 text-makoclaw-accent border border-makoclaw-accent/20">AI-Ready</span>
          </div>
          <p class="text-xs text-makoclaw-text-secondary/70 mb-4">
            Describe the personality you want and AI will generate the soul, identity, and user context for you.
          </p>
          <div class="space-y-3">
            <textarea
              v-model="aiPrompt"
              rows="3"
              placeholder="e.g. 'A witty, sarcastic assistant that loves science analogies and speaks like a mentor...'"
              class="w-full px-4 py-3 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 resize-none backdrop-blur-sm"
            />
            <button
              :disabled="!aiPrompt.trim() || generating"
              class="px-5 py-2.5 bg-gradient-to-r from-makoclaw-accent to-indigo-600 hover:from-makoclaw-accent-hover hover:to-indigo-700 text-white rounded-xl text-sm font-bold shadow-lg shadow-makoclaw-accent/25 transition-all flex items-center gap-2 disabled:opacity-40 disabled:cursor-not-allowed active:scale-95"
              @click="generateWithAI"
            >
              <svg v-if="generating" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
              <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              {{ generating ? 'Generating...' : 'Generate with AI' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Section Divider -->
    <div class="flex items-center gap-3">
      <div class="h-px bg-makoclaw-border/30 flex-1" />
      <span class="text-xs font-medium text-makoclaw-text-secondary/50">Manual Configuration</span>
      <div class="h-px bg-makoclaw-border/30 flex-1" />
    </div>

    <!-- Soul Card (SOUL.md) -->
    <div class="bg-makoclaw-surface/30 border border-makoclaw-border/50 rounded-xl p-5 hover:border-makoclaw-accent/20 transition-all">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 flex items-center justify-center ring-1 ring-white/10">
            <svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
            </svg>
          </div>
          <div>
            <h3 class="text-sm font-bold text-makoclaw-text">Soul</h3>
            <p class="text-xs text-makoclaw-text-secondary/60">Core personality and values</p>
          </div>
        </div>
        <button
          class="p-2 rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-text hover:bg-makoclaw-surface-hover transition-colors"
          :title="soulRawMode ? 'Switch to form' : 'Switch to raw editor'"
          @click="soulRawMode = !soulRawMode"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
          </svg>
        </button>
      </div>

      <div v-if="soulRawMode" class="space-y-2">
        <textarea
          v-model="soulRaw"
          rows="12"
          class="w-full px-4 py-3 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm font-mono text-makoclaw-text-secondary leading-relaxed focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none resize-y placeholder:text-makoclaw-text-secondary/30 backdrop-blur-sm"
        />
      </div>
      <div v-else class="space-y-4">
        <!-- Preamble -->
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Introduction</label>
          <textarea
            v-model="soul.preamble"
            rows="2"
            placeholder="I am makoclaw, a lightweight AI assistant..."
            class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none resize-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
          />
        </div>
        <!-- Personality -->
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Personality Traits</label>
          <div class="flex flex-wrap gap-2 mb-2">
            <span
              v-for="(trait, i) in soul.personality"
              :key="i"
              class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-purple-500/10 border border-purple-500/20 rounded-lg text-xs text-purple-300"
            >
              {{ trait }}
              <button class="hover:text-purple-100 transition-colors" @click="soul.personality.splice(i, 1)">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </span>
          </div>
          <div class="flex gap-2">
            <input
              v-model="newPersonalityTrait"
              type="text"
              placeholder="Add trait..."
              class="flex-1 px-4 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
              @keydown.enter.prevent="addTrait('personality')"
            >
            <button
              class="px-3 py-2 bg-purple-500/15 border border-purple-500/30 rounded-xl text-xs font-medium text-purple-300 hover:bg-purple-500/25 transition-colors"
              @click="addTrait('personality')"
            >
              Add
            </button>
          </div>
        </div>
        <!-- Values -->
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Values</label>
          <div class="flex flex-wrap gap-2 mb-2">
            <span
              v-for="(value, i) in soul.values"
              :key="i"
              class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-cyan-500/10 border border-cyan-500/20 rounded-lg text-xs text-cyan-300"
            >
              {{ value }}
              <button class="hover:text-cyan-100 transition-colors" @click="soul.values.splice(i, 1)">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </span>
          </div>
          <div class="flex gap-2">
            <input
              v-model="newValueTrait"
              type="text"
              placeholder="Add value..."
              class="flex-1 px-4 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
              @keydown.enter.prevent="addTrait('values')"
            >
            <button
              class="px-3 py-2 bg-cyan-500/15 border border-cyan-500/30 rounded-xl text-xs font-medium text-cyan-300 hover:bg-cyan-500/25 transition-colors"
              @click="addTrait('values')"
            >
              Add
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Identity Card (IDENTITY.md) -->
    <div class="bg-makoclaw-surface/30 border border-makoclaw-border/50 rounded-xl p-5 hover:border-makoclaw-accent/20 transition-all">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center ring-1 ring-white/10">
            <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V8a2 2 0 00-2-2h-5m-4 0V5a2 2 0 114 0v1m-4 0a2 2 0 104 0m-5 8a2 2 0 100-4 2 2 0 000 4zm0 0c1.306 0 2.417.835 2.83 2M9 14a3.001 3.001 0 00-2.83 2M15 11h3m-3 4h2" />
            </svg>
          </div>
          <div>
            <h3 class="text-sm font-bold text-makoclaw-text">Identity</h3>
            <p class="text-xs text-makoclaw-text-secondary/60">Name, description, and capabilities</p>
          </div>
        </div>
        <button
          class="p-2 rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-text hover:bg-makoclaw-surface-hover transition-colors"
          :title="identityRawMode ? 'Switch to form' : 'Switch to raw editor'"
          @click="identityRawMode = !identityRawMode"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
          </svg>
        </button>
      </div>

      <div v-if="identityRawMode" class="space-y-2">
        <textarea
          v-model="identityRaw"
          rows="14"
          class="w-full px-4 py-3 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm font-mono text-makoclaw-text-secondary leading-relaxed focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none resize-y placeholder:text-makoclaw-text-secondary/30 backdrop-blur-sm"
        />
      </div>
      <div v-else class="space-y-4">
        <!-- Name + Version row -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div class="space-y-1.5">
            <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Name</label>
            <input
              v-model="identity.name"
              type="text"
              placeholder="makoclaw"
              class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
            >
          </div>
          <div class="space-y-1.5">
            <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Version</label>
            <input
              v-model="identity.version"
              type="text"
              placeholder="0.1.0"
              class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
            >
          </div>
        </div>
        <!-- Description -->
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Description</label>
          <textarea
            v-model="identity.description"
            rows="2"
            placeholder="An ultra-efficient, Go-native personal AI assistant..."
            class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none resize-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
          />
        </div>
        <!-- Purpose -->
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Purpose</label>
          <div class="flex flex-wrap gap-2 mb-2">
            <span
              v-for="(item, i) in identity.purpose"
              :key="i"
              class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-blue-500/10 border border-blue-500/20 rounded-lg text-xs text-blue-300"
            >
              {{ item }}
              <button class="hover:text-blue-100 transition-colors" @click="identity.purpose.splice(i, 1)">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </span>
          </div>
          <div class="flex gap-2">
            <input
              v-model="newPurpose"
              type="text"
              placeholder="Add purpose..."
              class="flex-1 px-4 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
              @keydown.enter.prevent="addIdentityTrait('purpose')"
            >
            <button
              class="px-3 py-2 bg-blue-500/15 border border-blue-500/30 rounded-xl text-xs font-medium text-blue-300 hover:bg-blue-500/25 transition-colors"
              @click="addIdentityTrait('purpose')"
            >
              Add
            </button>
          </div>
        </div>
        <!-- Capabilities -->
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Capabilities</label>
          <div class="flex flex-wrap gap-2 mb-2">
            <span
              v-for="(item, i) in identity.capabilities"
              :key="i"
              class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-emerald-500/10 border border-emerald-500/20 rounded-lg text-xs text-emerald-300"
            >
              {{ item }}
              <button class="hover:text-emerald-100 transition-colors" @click="identity.capabilities.splice(i, 1)">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </span>
          </div>
          <div class="flex gap-2">
            <input
              v-model="newCapability"
              type="text"
              placeholder="Add capability..."
              class="flex-1 px-4 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
              @keydown.enter.prevent="addIdentityTrait('capabilities')"
            >
            <button
              class="px-3 py-2 bg-emerald-500/15 border border-emerald-500/30 rounded-xl text-xs font-medium text-emerald-300 hover:bg-emerald-500/25 transition-colors"
              @click="addIdentityTrait('capabilities')"
            >
              Add
            </button>
          </div>
        </div>
        <!-- Philosophy -->
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Philosophy</label>
          <div class="flex flex-wrap gap-2 mb-2">
            <span
              v-for="(item, i) in identity.philosophy"
              :key="i"
              class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-amber-500/10 border border-amber-500/20 rounded-lg text-xs text-amber-300"
            >
              {{ item }}
              <button class="hover:text-amber-100 transition-colors" @click="identity.philosophy.splice(i, 1)">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </span>
          </div>
          <div class="flex gap-2">
            <input
              v-model="newPhilosophy"
              type="text"
              placeholder="Add philosophy..."
              class="flex-1 px-4 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
              @keydown.enter.prevent="addIdentityTrait('philosophy')"
            >
            <button
              class="px-3 py-2 bg-amber-500/15 border border-amber-500/30 rounded-xl text-xs font-medium text-amber-300 hover:bg-amber-500/25 transition-colors"
              @click="addIdentityTrait('philosophy')"
            >
              Add
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- User Context Card (USER.md) -->
    <div class="bg-makoclaw-surface/30 border border-makoclaw-border/50 rounded-xl p-5 hover:border-makoclaw-accent/20 transition-all">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-orange-500/20 to-yellow-500/20 flex items-center justify-center ring-1 ring-white/10">
            <svg class="w-5 h-5 text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <div>
            <h3 class="text-sm font-bold text-makoclaw-text">User Context</h3>
            <p class="text-xs text-makoclaw-text-secondary/60">Information about you for the AI</p>
          </div>
        </div>
        <button
          class="p-2 rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-text hover:bg-makoclaw-surface-hover transition-colors"
          :title="userRawMode ? 'Switch to form' : 'Switch to raw editor'"
          @click="userRawMode = !userRawMode"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
          </svg>
        </button>
      </div>

      <div v-if="userRawMode" class="space-y-2">
        <textarea
          v-model="userRaw"
          rows="10"
          class="w-full px-4 py-3 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm font-mono text-makoclaw-text-secondary leading-relaxed focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none resize-y placeholder:text-makoclaw-text-secondary/30 backdrop-blur-sm"
        />
      </div>
      <div v-else class="space-y-4">
        <!-- Preferences row -->
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div class="space-y-1.5">
            <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Communication Style</label>
            <input
              v-model="userCtx.communicationStyle"
              type="text"
              placeholder="casual / formal / mixed"
              class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
            >
          </div>
          <div class="space-y-1.5">
            <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Timezone</label>
            <input
              v-model="userCtx.timezone"
              type="text"
              placeholder="America/Buenos_Aires"
              class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
            >
          </div>
          <div class="space-y-1.5">
            <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Language</label>
            <input
              v-model="userCtx.language"
              type="text"
              placeholder="English / Spanish / etc."
              class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
            >
          </div>
        </div>
        <!-- Personal info -->
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div class="space-y-1.5">
            <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Name</label>
            <input
              v-model="userCtx.name"
              type="text"
              placeholder="(optional)"
              class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
            >
          </div>
          <div class="space-y-1.5">
            <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Location</label>
            <input
              v-model="userCtx.location"
              type="text"
              placeholder="(optional)"
              class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
            >
          </div>
          <div class="space-y-1.5">
            <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Occupation</label>
            <input
              v-model="userCtx.occupation"
              type="text"
              placeholder="(optional)"
              class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
            >
          </div>
        </div>
        <!-- Learning Goals -->
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">Learning Goals</label>
          <textarea
            v-model="userCtx.learningGoals"
            rows="3"
            placeholder="What you want to learn, areas of interest, preferred interaction style..."
            class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none resize-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
          />
        </div>
      </div>
    </div>

    <!-- Save Bar -->
    <div class="flex items-center justify-between p-4 bg-makoclaw-surface/30 border border-makoclaw-border/50 rounded-xl">
      <p class="text-xs text-makoclaw-text-secondary/60">
        Changes take effect on the next chat message.
      </p>
      <div class="flex items-center gap-3">
        <button
          :disabled="saving"
          class="px-6 py-2.5 bg-gradient-to-r from-makoclaw-accent to-indigo-600 hover:from-makoclaw-accent-hover hover:to-indigo-700 text-white rounded-xl text-sm font-bold shadow-lg shadow-makoclaw-accent/25 transition-all disabled:opacity-40 disabled:cursor-not-allowed active:scale-95"
          @click="saveAll"
        >
          {{ saving ? 'Saving...' : 'Save All' }}
        </button>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import advancedService from '../../services/advancedService'
import { useToast } from '../../composables/useToast'
import {
  parseMarkdownSections,
  parseBulletList,
  serializeBulletList,
  serializeToMarkdown,
  getSection
} from '../../composables/useMarkdownSections'

const toast = useToast()

// State
const loading = ref(true)
const saving = ref(false)
const generating = ref(false)
const aiPrompt = ref('')

// Raw mode toggles
const soulRawMode = ref(false)
const identityRawMode = ref(false)
const userRawMode = ref(false)

// Raw content for raw editor mode
const soulRaw = ref('')
const identityRaw = ref('')
const userRaw = ref('')

// Structured data - Soul
const soul = ref({
  preamble: '',
  personality: [],
  values: []
})
const soulExtra = ref([]) // unknown sections preserved for round-trip

// Structured data - Identity
const identity = ref({
  name: '',
  description: '',
  version: '',
  purpose: [],
  capabilities: [],
  philosophy: [],
  license: '',
  repository: ''
})
const identityExtra = ref([])

// Structured data - User
const userCtx = ref({
  communicationStyle: '',
  timezone: '',
  language: '',
  name: '',
  location: '',
  occupation: '',
  learningGoals: ''
})
const userExtra = ref([])

// New trait inputs
const newPersonalityTrait = ref('')
const newValueTrait = ref('')
const newPurpose = ref('')
const newCapability = ref('')
const newPhilosophy = ref('')

// Sync raw ↔ structured when toggling modes
watch(soulRawMode, (isRaw) => {
  if (isRaw) {
    soulRaw.value = buildSoulMarkdown()
  } else {
    parseSoulContent(soulRaw.value)
  }
})

watch(identityRawMode, (isRaw) => {
  if (isRaw) {
    identityRaw.value = buildIdentityMarkdown()
  } else {
    parseIdentityContent(identityRaw.value)
  }
})

watch(userRawMode, (isRaw) => {
  if (isRaw) {
    userRaw.value = buildUserMarkdown()
  } else {
    parseUserContent(userRaw.value)
  }
})

// Parse helpers
function parseSoulContent(markdown) {
  const parsed = parseMarkdownSections(markdown)
  soul.value.preamble = parsed.preamble
  soul.value.personality = parseBulletList(getSection(parsed.sections, 'Personality'))
  soul.value.values = parseBulletList(getSection(parsed.sections, 'Values'))
  soulExtra.value = parsed.sections.filter(
    s => !['personality', 'values'].includes(s.name.toLowerCase())
  )
}

function parseIdentityContent(markdown) {
  const parsed = parseMarkdownSections(markdown)
  identity.value.name = getSection(parsed.sections, 'Name').trim()
  identity.value.description = getSection(parsed.sections, 'Description').trim()
  identity.value.version = getSection(parsed.sections, 'Version').trim()
  identity.value.purpose = parseBulletList(getSection(parsed.sections, 'Purpose'))
  identity.value.capabilities = parseBulletList(getSection(parsed.sections, 'Capabilities'))
  identity.value.philosophy = parseBulletList(getSection(parsed.sections, 'Philosophy'))
  identity.value.license = getSection(parsed.sections, 'License').trim()
  identity.value.repository = getSection(parsed.sections, 'Repository').trim()
  identityExtra.value = parsed.sections.filter(
    s => !['name', 'description', 'version', 'purpose', 'capabilities', 'philosophy', 'license', 'repository'].includes(s.name.toLowerCase())
  )
}

function parseUserContent(markdown) {
  const parsed = parseMarkdownSections(markdown)

  // Preferences section has key: value bullet format
  const prefsContent = getSection(parsed.sections, 'Preferences')
  const prefs = parseBulletList(prefsContent)
  for (const pref of prefs) {
    const [key, ...rest] = pref.split(':')
    const val = rest.join(':').trim().replace(/^\(|\)$/g, '')
    const k = key.trim().toLowerCase()
    if (k.includes('communication')) userCtx.value.communicationStyle = val
    else if (k.includes('timezone')) userCtx.value.timezone = val
    else if (k.includes('language')) userCtx.value.language = val
  }

  // Personal Information
  const personalContent = getSection(parsed.sections, 'Personal Information')
  const personalItems = parseBulletList(personalContent)
  for (const item of personalItems) {
    const [key, ...rest] = item.split(':')
    const val = rest.join(':').trim().replace(/^\(|\)$/g, '')
    const k = key.trim().toLowerCase()
    if (k === 'name') userCtx.value.name = val
    else if (k === 'location') userCtx.value.location = val
    else if (k === 'occupation') userCtx.value.occupation = val
  }

  // Learning Goals
  const goalsContent = getSection(parsed.sections, 'Learning Goals')
  userCtx.value.learningGoals = goalsContent.trim()

  userExtra.value = parsed.sections.filter(
    s => !['preferences', 'personal information', 'learning goals'].includes(s.name.toLowerCase())
  )
}

// Build markdown helpers
function buildSoulMarkdown() {
  const sections = []
  if (soul.value.personality.length) {
    sections.push({ name: 'Personality', content: serializeBulletList(soul.value.personality) })
  }
  if (soul.value.values.length) {
    sections.push({ name: 'Values', content: serializeBulletList(soul.value.values) })
  }
  sections.push(...soulExtra.value)
  return serializeToMarkdown('Soul', soul.value.preamble, sections)
}

function buildIdentityMarkdown() {
  const sections = []
  if (identity.value.name) sections.push({ name: 'Name', content: identity.value.name })
  if (identity.value.description) sections.push({ name: 'Description', content: identity.value.description })
  if (identity.value.version) sections.push({ name: 'Version', content: identity.value.version })
  if (identity.value.purpose.length) sections.push({ name: 'Purpose', content: serializeBulletList(identity.value.purpose) })
  if (identity.value.capabilities.length) sections.push({ name: 'Capabilities', content: serializeBulletList(identity.value.capabilities) })
  if (identity.value.philosophy.length) sections.push({ name: 'Philosophy', content: serializeBulletList(identity.value.philosophy) })
  if (identity.value.license) sections.push({ name: 'License', content: identity.value.license })
  if (identity.value.repository) sections.push({ name: 'Repository', content: identity.value.repository })
  sections.push(...identityExtra.value)
  return serializeToMarkdown('Identity', '', sections)
}

function buildUserMarkdown() {
  const prefItems = []
  prefItems.push(`Communication style: ${userCtx.value.communicationStyle || '(casual/formal)'}`)
  prefItems.push(`Timezone: ${userCtx.value.timezone || '(your timezone)'}`)
  prefItems.push(`Language: ${userCtx.value.language || '(your preferred language)'}`)

  const personalItems = []
  personalItems.push(`Name: ${userCtx.value.name || '(optional)'}`)
  personalItems.push(`Location: ${userCtx.value.location || '(optional)'}`)
  personalItems.push(`Occupation: ${userCtx.value.occupation || '(optional)'}`)

  const sections = [
    { name: 'Preferences', content: serializeBulletList(prefItems) },
    { name: 'Personal Information', content: serializeBulletList(personalItems) },
    { name: 'Learning Goals', content: userCtx.value.learningGoals || serializeBulletList(['What the user wants to learn from AI', 'Preferred interaction style', 'Areas of interest']) }
  ]
  sections.push(...userExtra.value)
  return serializeToMarkdown('User', 'Information about user goes here.', sections)
}

// Add trait helpers
function addTrait(field) {
  const inputRef = field === 'personality' ? newPersonalityTrait : newValueTrait
  const val = inputRef.value.trim()
  if (!val) return
  soul.value[field].push(val)
  inputRef.value = ''
}

function addIdentityTrait(field) {
  const refMap = { purpose: newPurpose, capabilities: newCapability, philosophy: newPhilosophy }
  const inputRef = refMap[field]
  const val = inputRef.value.trim()
  if (!val) return
  identity.value[field].push(val)
  inputRef.value = ''
}

// Load files
async function loadFiles() {
  loading.value = true
  try {
    const [soulRes, identityRes, userRes] = await Promise.all([
      advancedService.fetchFiles('SOUL.md').catch(() => null),
      advancedService.fetchFiles('IDENTITY.md').catch(() => null),
      advancedService.fetchFiles('USER.md').catch(() => null)
    ])

    if (soulRes?.content) {
      soulRaw.value = soulRes.content
      parseSoulContent(soulRes.content)
    }
    if (identityRes?.content) {
      identityRaw.value = identityRes.content
      parseIdentityContent(identityRes.content)
    }
    if (userRes?.content) {
      userRaw.value = userRes.content
      parseUserContent(userRes.content)
    }
  } catch (err) {
    toast.error('Failed to load soul files: ' + err.message)
  } finally {
    loading.value = false
  }
}

// Save all files
async function saveAll() {
  saving.value = true
  try {
    const soulMd = soulRawMode.value ? soulRaw.value : buildSoulMarkdown()
    const identityMd = identityRawMode.value ? identityRaw.value : buildIdentityMarkdown()
    const userMd = userRawMode.value ? userRaw.value : buildUserMarkdown()

    await Promise.all([
      advancedService.updateFileContent('SOUL.md', soulMd),
      advancedService.updateFileContent('IDENTITY.md', identityMd),
      advancedService.updateFileContent('USER.md', userMd)
    ])

    toast.success('Soul saved successfully')
  } catch (err) {
    toast.error('Failed to save: ' + err.message)
  } finally {
    saving.value = false
  }
}

// AI Generation
async function generateWithAI() {
  if (!aiPrompt.value.trim()) return
  generating.value = true
  try {
    const result = await advancedService.generateSoul(aiPrompt.value.trim())

    if (result.soul) {
      if (result.soul.personality) soul.value.personality = result.soul.personality
      if (result.soul.values) soul.value.values = result.soul.values
      if (result.soul.preamble) soul.value.preamble = result.soul.preamble
    }
    if (result.identity) {
      if (result.identity.name) identity.value.name = result.identity.name
      if (result.identity.description) identity.value.description = result.identity.description
      if (result.identity.purpose) identity.value.purpose = result.identity.purpose
      if (result.identity.capabilities) identity.value.capabilities = result.identity.capabilities
      if (result.identity.philosophy) identity.value.philosophy = result.identity.philosophy
    }
    if (result.user_context) {
      if (result.user_context.communication_style) userCtx.value.communicationStyle = result.user_context.communication_style
      if (result.user_context.timezone) userCtx.value.timezone = result.user_context.timezone
      if (result.user_context.language) userCtx.value.language = result.user_context.language
    }

    toast.success('Soul generated! Review the fields below and save when ready.')
  } catch (err) {
    toast.error('AI generation failed: ' + err.message)
  } finally {
    generating.value = false
  }
}

onMounted(loadFiles)
</script>
