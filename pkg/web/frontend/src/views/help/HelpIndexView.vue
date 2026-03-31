<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <!-- Hero -->
    <section class="py-16 sm:py-20">
      <div class="text-center max-w-3xl mx-auto">
        <div class="inline-flex items-center gap-2 px-4 py-2 bg-blue-500/10 rounded-full mb-6">
          <span class="w-2 h-2 bg-blue-500 rounded-full animate-pulse"></span>
          <span class="text-sm text-blue-500 dark:text-blue-400 font-medium">Documentation</span>
        </div>
        <h2 class="text-4xl sm:text-5xl font-bold mb-6 leading-tight">
          <span class="text-slate-900 dark:text-makoclaw-text">Learn to use </span>
          <span class="bg-gradient-to-r from-blue-500 via-cyan-500 to-teal-400 bg-clip-text text-transparent">MakoClaw</span>
        </h2>
        <p class="text-xl text-slate-500 dark:text-makoclaw-text-secondary leading-relaxed">
          Guides and tutorials to get the most out of your AI agents.
        </p>
      </div>
    </section>

    <!-- Guides Grid -->
    <section class="pb-20">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">

        <!-- Marketing Campaigns -->
        <router-link to="/help/marketing" class="group block">
          <div class="h-full p-6 rounded-2xl border transition-all duration-200 bg-white/60 backdrop-blur-sm border-slate-200/80 hover:border-blue-300 hover:shadow-xl hover:shadow-blue-500/10 dark:bg-makoclaw-surface/60 dark:border-makoclaw-border dark:hover:border-blue-500/50">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-pink-500 to-rose-500 flex items-center justify-center text-white text-xl mb-4 shadow-lg shadow-pink-500/25 group-hover:scale-110 transition-transform">
              📣
            </div>
            <h3 class="text-lg font-bold text-slate-900 dark:text-makoclaw-text mb-2 group-hover:text-blue-500 transition-colors">Marketing Campaigns</h3>
            <p class="text-sm text-slate-500 dark:text-makoclaw-text-secondary leading-relaxed mb-4">
              Run full marketing campaigns with AI — generate images, post to social media, schedule content, and track analytics.
            </p>
            <div class="flex flex-wrap gap-2">
              <span class="px-2 py-1 text-xs rounded-full bg-pink-500/10 text-pink-600 dark:text-pink-400 font-medium">Image Generation</span>
              <span class="px-2 py-1 text-xs rounded-full bg-blue-500/10 text-blue-600 dark:text-blue-400 font-medium">Social Media</span>
              <span class="px-2 py-1 text-xs rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-medium">Analytics</span>
            </div>
          </div>
        <!-- Dev Studio -->
        <router-link to="/help/dev-studio" class="group block">
          <div class="h-full p-6 rounded-2xl border transition-all duration-200 bg-white/60 backdrop-blur-sm border-slate-200/80 hover:border-cyan-300 hover:shadow-xl hover:shadow-cyan-500/10 dark:bg-makoclaw-surface/60 dark:border-makoclaw-border dark:hover:border-cyan-500/50">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-cyan-400 to-blue-600 flex items-center justify-center text-white text-xl mb-4 shadow-lg shadow-cyan-500/25 group-hover:scale-110 transition-transform">
              ⚡
            </div>
            <h3 class="text-lg font-bold text-slate-900 dark:text-makoclaw-text mb-2 group-hover:text-cyan-500 transition-colors">Dev Studio</h3>
            <p class="text-sm text-slate-500 dark:text-makoclaw-text-secondary leading-relaxed mb-4">
              Aprende a ejecutar Claude Code u OpenCode dentro del ecosistema de MakoClaw, con memoria semántica local (ONNX).
            </p>
            <div class="flex flex-wrap gap-2">
              <span class="px-2 py-1 text-xs rounded-full bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 font-medium">Terminal</span>
              <span class="px-2 py-1 text-xs rounded-full bg-purple-500/10 text-purple-600 dark:text-purple-400 font-medium">Memoria Local</span>
            </div>
          </div>
        </router-link>

        <!-- Coming Soon: Channels -->
        <div class="group block opacity-60 cursor-default">
          <div class="h-full p-6 rounded-2xl border bg-white/40 backdrop-blur-sm border-slate-200/60 dark:bg-makoclaw-surface/40 dark:border-makoclaw-border/60">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-violet-500 to-purple-500 flex items-center justify-center text-white text-xl mb-4 shadow-lg shadow-violet-500/25">
              💬
            </div>
            <h3 class="text-lg font-bold text-slate-900 dark:text-makoclaw-text mb-2">Channels</h3>
            <p class="text-sm text-slate-500 dark:text-makoclaw-text-secondary leading-relaxed mb-4">
              Connect your agents to Telegram, Discord, Slack, WhatsApp, and more.
            </p>
            <span class="inline-flex items-center gap-1 px-3 py-1 text-xs rounded-full bg-slate-500/10 text-slate-500 dark:text-slate-400 font-medium">
              Coming soon
            </span>
          </div>
        </div>

        <!-- Coming Soon: Agents & Specialists -->
        <div class="group block opacity-60 cursor-default">
          <div class="h-full p-6 rounded-2xl border bg-white/40 backdrop-blur-sm border-slate-200/60 dark:bg-makoclaw-surface/40 dark:border-makoclaw-border/60">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-amber-500 to-orange-500 flex items-center justify-center text-white text-xl mb-4 shadow-lg shadow-amber-500/25">
              🤖
            </div>
            <h3 class="text-lg font-bold text-slate-900 dark:text-makoclaw-text mb-2">Agents & Specialists</h3>
            <p class="text-sm text-slate-500 dark:text-makoclaw-text-secondary leading-relaxed mb-4">
              Create specialized agents, configure the orchestrator, and delegate tasks automatically.
            </p>
            <span class="inline-flex items-center gap-1 px-3 py-1 text-xs rounded-full bg-slate-500/10 text-slate-500 dark:text-slate-400 font-medium">
              Coming soon
            </span>
          </div>
        </div>

      </div>
    </section>
  </div>
</template>

<script setup>
</script>
