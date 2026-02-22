<template>
  <div class="min-h-screen bg-gradient-to-br from-kakoclaw-bg via-kakoclaw-surface to-kakoclaw-bg">
    <!-- Navigation -->
    <nav class="border-b border-kakoclaw-border bg-kakoclaw-surface/80 backdrop-blur-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <div class="flex items-center">
            <h1 class="text-2xl font-bold bg-gradient-to-r from-kakoclaw-accent to-emerald-500 bg-clip-text text-transparent">
              KakoClaw
            </h1>
          </div>
          <div class="flex items-center gap-4">
            <router-link
              to="/login"
              class="px-4 py-2 text-kakoclaw-text hover:text-kakoclaw-accent transition-colors font-medium"
            >
              Sign In
            </router-link>
            <router-link
              to="/signup"
              class="px-6 py-2 bg-kakoclaw-accent hover:bg-kakoclaw-accent-hover text-white rounded-lg transition-all shadow-lg shadow-kakoclaw-accent/20 font-medium"
            >
              Get Started
            </router-link>
          </div>
        </div>
      </div>
    </nav>

    <!-- Hero Section -->
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
      <div class="text-center max-w-3xl mx-auto">
        <h2 class="text-5xl sm:text-6xl font-bold mb-6">
          <span class="bg-gradient-to-r from-kakoclaw-accent via-emerald-500 to-teal-400 bg-clip-text text-transparent">
            AI Agent Platform
          </span>
          <br />
          <span class="text-kakoclaw-text">Built for Power Users</span>
        </h2>
        <p class="text-xl text-kakoclaw-text-secondary mb-10 leading-relaxed">
          Deploy autonomous AI agents that execute tasks, manage workflows, and integrate with your favorite tools. 
          All from one powerful control panel.
        </p>
        <div class="flex items-center justify-center gap-4">
          <router-link
            to="/signup"
            class="px-8 py-4 bg-kakoclaw-accent hover:bg-kakoclaw-accent-hover text-white text-lg rounded-xl transition-all shadow-2xl shadow-kakoclaw-accent/30 font-bold"
          >
            Create Free Account
          </router-link>
          <router-link
            to="/login"
            class="px-8 py-4 bg-kakoclaw-surface border-2 border-kakoclaw-border hover:border-kakoclaw-accent text-kakoclaw-text text-lg rounded-xl transition-all font-medium"
          >
            Sign In
          </router-link>
        </div>
      </div>

      <!-- Features Grid -->
      <div class="mt-24 grid grid-cols-1 md:grid-cols-3 gap-8">
        <!-- Feature 1 -->
        <div class="glass-panel rounded-2xl p-8 border border-kakoclaw-border hover:border-kakoclaw-accent/50 transition-all">
          <div class="w-12 h-12 bg-kakoclaw-accent/10 rounded-lg flex items-center justify-center mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-kakoclaw-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <h3 class="text-xl font-bold text-kakoclaw-text mb-2">Autonomous Execution</h3>
          <p class="text-kakoclaw-text-secondary">
            Let AI agents execute complex multi-step tasks while you focus on what matters. Full tool access included.
          </p>
        </div>

        <!-- Feature 2 -->
        <div class="glass-panel rounded-2xl p-8 border border-kakoclaw-border hover:border-kakoclaw-accent/50 transition-all">
          <div class="w-12 h-12 bg-emerald-500/10 rounded-lg flex items-center justify-center mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
            </svg>
          </div>
          <h3 class="text-xl font-bold text-kakoclaw-text mb-2">Multi-Channel Support</h3>
          <p class="text-kakoclaw-text-secondary">
            Connect via Telegram, Discord, Slack, or Web. One agent, multiple interfaces.
          </p>
        </div>

        <!-- Feature 3 -->
        <div class="glass-panel rounded-2xl p-8 border border-kakoclaw-border hover:border-kakoclaw-accent/50 transition-all">
          <div class="w-12 h-12 bg-teal-500/10 rounded-lg flex items-center justify-center mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-teal-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
          </div>
          <h3 class="text-xl font-bold text-kakoclaw-text mb-2">Private & Secure</h3>
          <p class="text-kakoclaw-text-secondary">
            Self-hosted deployment. Your data stays on your infrastructure. Full control guaranteed.
          </p>
        </div>
      </div>

      <!-- Stats -->
      <div class="mt-20 grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
        <div>
          <div class="text-4xl font-bold text-kakoclaw-accent mb-2">20+</div>
          <div class="text-kakoclaw-text-secondary">Built-in Tools</div>
        </div>
        <div>
          <div class="text-4xl font-bold text-emerald-500 mb-2">5+</div>
          <div class="text-kakoclaw-text-secondary">Channels</div>
        </div>
        <div>
          <div class="text-4xl font-bold text-teal-500 mb-2">Unlimited</div>
          <div class="text-kakoclaw-text-secondary">Workflows</div>
        </div>
        <div>
          <div class="text-4xl font-bold text-kakoclaw-accent mb-2">100%</div>
          <div class="text-kakoclaw-text-secondary">Open Source</div>
        </div>
      </div>

      <!-- CTA Section -->
      <div class="mt-24 text-center">
        <div class="glass-panel rounded-3xl p-12 border-2 border-kakoclaw-accent/30">
          <h3 class="text-3xl font-bold text-kakoclaw-text mb-4">Ready to deploy your AI agent?</h3>
          <p class="text-lg text-kakoclaw-text-secondary mb-8">
            Join developers building the future of autonomous AI systems.
          </p>
          <router-link
            to="/signup"
            class="inline-block px-10 py-4 bg-gradient-to-r from-kakoclaw-accent to-emerald-500 hover:from-kakoclaw-accent-hover hover:to-emerald-600 text-white text-lg rounded-xl transition-all shadow-2xl shadow-kakoclaw-accent/40 font-bold"
          >
            Start Building Now ->
          </router-link>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <footer class="border-t border-kakoclaw-border mt-20">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div class="text-center text-kakoclaw-text-secondary text-sm">
          <p>(c) 2026 KakoClaw. Open source AI agent platform.</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup>
// Landing page component
</script>
