<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-indigo-50/20 dark:from-makoclaw-bg dark:via-makoclaw-surface dark:to-makoclaw-bg">

    <!-- Promo Banner -->
    <div class="bg-gradient-to-r from-blue-600 to-cyan-500 text-white text-center py-2.5 px-4 text-sm font-medium">
      <span class="hidden sm:inline">Early Adopter Launch — </span>
      <span class="font-bold">$5/mo</span>
      <span class="line-through opacity-70 ml-1">$10/mo</span>
      <span class="mx-2 opacity-50">|</span>
      <router-link to="/signup" class="underline hover:no-underline font-semibold">Lock this price forever →</router-link>
    </div>

    <!-- Navigation -->
    <nav class="sticky top-0 z-sticky border-b bg-white/80 backdrop-blur-md border-slate-200/80 dark:bg-makoclaw-surface/80 dark:backdrop-blur-sm dark:border-makoclaw-border">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white font-bold text-lg">🦈</div>
            <h1 class="text-2xl font-bold bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">
              MakoClaw
            </h1>
          </div>
          <div class="flex items-center gap-2 sm:gap-4">
            <router-link
              to="/login"
              class="px-3 sm:px-4 py-2 text-slate-600 hover:text-blue-500 dark:text-makoclaw-text dark:hover:text-blue-500 transition-colors font-medium text-sm whitespace-nowrap"
            >
              Sign In
            </router-link>
            <router-link
              to="/signup"
              class="px-3 sm:px-6 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition-all shadow-lg shadow-blue-500/20 font-medium text-sm whitespace-nowrap"
            >
              <span class="sm:hidden">Start</span>
              <span class="hidden sm:inline">Get Started</span>
            </router-link>
          </div>
        </div>
      </div>
    </nav>

    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">

      <!-- ==================== HERO ==================== -->
      <section class="py-16 sm:py-24">
        <div class="text-center max-w-4xl mx-auto">
          <div class="inline-flex items-center gap-2 px-4 py-2 bg-blue-500/10 rounded-full mb-6">
            <span class="w-2 h-2 bg-blue-500 rounded-full animate-pulse"></span>
            <span class="text-sm text-blue-500 dark:text-blue-400 font-medium">Zero Configuration Required</span>
          </div>
          <h2 class="text-4xl sm:text-6xl lg:text-7xl font-bold mb-6 leading-tight">
            <span class="text-slate-900 dark:text-makoclaw-text">Your AI assistant </span>
            <span class="bg-gradient-to-r from-blue-500 via-cyan-500 to-teal-400 bg-clip-text text-transparent">across all channels.</span>
          </h2>
          <p class="text-xl sm:text-2xl text-slate-500 dark:text-makoclaw-text-secondary mb-4 leading-relaxed max-w-3xl mx-auto">
            No code, no servers, no Docker. Just login, connect your AI provider, and you're live on Telegram, Discord, Slack, WhatsApp and more.
          </p>
          <p class="text-lg text-slate-400 dark:text-makoclaw-text-secondary/70 mb-10">
            Ready in under 5 minutes.
          </p>
          <div class="flex items-center justify-center gap-4 flex-wrap">
            <router-link
              to="/signup"
              class="px-10 py-4 bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 text-white text-lg rounded-xl transition-all shadow-2xl shadow-blue-500/30 font-bold active:scale-[0.98]"
            >
              Start Free — Setup in 5 min
            </router-link>
            <a
              href="#how-it-works"
              class="px-10 py-4 border-2 rounded-xl transition-all font-medium text-lg border-slate-200 hover:border-blue-400 text-slate-700 hover:text-blue-600 dark:bg-makoclaw-surface dark:border-makoclaw-border dark:hover:border-blue-500 dark:text-makoclaw-text"
            >
              How it works
            </a>
          </div>
        </div>
      </section>

      <!-- ==================== HOW IT WORKS ==================== -->
      <section id="how-it-works" class="py-16 sm:py-20">
        <h3 class="text-3xl sm:text-4xl font-bold text-center mb-4">
          <span class="text-slate-900 dark:text-makoclaw-text">Live in </span>
          <span class="bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">3 simple steps</span>
        </h3>
        <p class="text-center text-slate-500 dark:text-makoclaw-text-secondary mb-12 text-lg max-w-2xl mx-auto">
          Forget Docker, VPS, and command lines. MakoClaw handles everything — you just configure.
        </p>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
          <!-- Step 1 -->
          <div class="relative text-center">
            <div class="w-16 h-16 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white text-2xl font-bold shadow-lg shadow-blue-500/25">1</div>
            <h4 class="text-xl font-bold text-slate-900 dark:text-makoclaw-text mb-3">Create your account</h4>
            <p class="text-slate-500 dark:text-makoclaw-text-secondary">
              Sign up in seconds. No credit card required. No sensitive server data needed.
            </p>
            <div class="hidden md:block absolute top-8 -right-4 w-8 text-slate-300 dark:text-makoclaw-border">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>
            </div>
          </div>

          <!-- Step 2 -->
          <div class="relative text-center">
            <div class="w-16 h-16 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-cyan-500 to-teal-500 flex items-center justify-center text-white text-2xl font-bold shadow-lg shadow-cyan-500/25">2</div>
            <h4 class="text-xl font-bold text-slate-900 dark:text-makoclaw-text mb-3">Connect your AI</h4>
            <p class="text-slate-500 dark:text-makoclaw-text-secondary">
              Add your API key from OpenAI, Anthropic, Groq, or any provider. You only pay for what you use.
            </p>
            <div class="hidden md:block absolute top-8 -right-4 w-8 text-slate-300 dark:text-makoclaw-border">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>
            </div>
          </div>

          <!-- Step 3 -->
          <div class="text-center">
            <div class="w-16 h-16 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-teal-500 to-emerald-500 flex items-center justify-center text-white text-2xl font-bold shadow-lg shadow-teal-500/25">3</div>
            <h4 class="text-xl font-bold text-slate-900 dark:text-makoclaw-text mb-3">Activate your channels</h4>
            <p class="text-slate-500 dark:text-makoclaw-text-secondary">
              Pick the channels where your bot should respond — Telegram, Discord, Slack, WhatsApp, and more. Done.
            </p>
          </div>
        </div>
      </section>

      <!-- ==================== COMPARISON TABLE ==================== -->
      <section class="py-16 sm:py-20">
        <h3 class="text-3xl sm:text-4xl font-bold text-center mb-4">
          <span class="text-slate-900 dark:text-makoclaw-text">Why </span>
          <span class="bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">MakoClaw?</span>
        </h3>
        <p class="text-center text-slate-500 dark:text-makoclaw-text-secondary mb-12 text-lg">
          Save time and headaches for the price of a coffee per month.
        </p>

        <div class="max-w-3xl mx-auto rounded-2xl overflow-hidden border shadow-xl bg-white/80 border-slate-200 shadow-slate-200/50 dark:bg-makoclaw-surface/80 dark:border-makoclaw-border dark:shadow-black/10">
          <div class="grid grid-cols-3 text-center font-bold text-sm sm:text-base">
            <div class="p-4 sm:p-5 bg-slate-50 dark:bg-makoclaw-bg/50 text-slate-500 dark:text-makoclaw-text-secondary">Feature</div>
            <div class="p-4 sm:p-5 bg-slate-50 dark:bg-makoclaw-bg/50 text-slate-500 dark:text-makoclaw-text-secondary">Self-Hosted</div>
            <div class="p-4 sm:p-5 bg-gradient-to-r from-blue-500 to-cyan-500 text-white">MakoClaw SaaS</div>
          </div>
          <div v-for="(row, i) in comparisonRows" :key="i" class="grid grid-cols-3 text-center text-sm sm:text-base border-t border-slate-100 dark:border-makoclaw-border/50">
            <div class="p-4 sm:p-5 font-medium text-slate-700 dark:text-makoclaw-text text-left pl-6">{{ row.feature }}</div>
            <div class="p-4 sm:p-5 text-slate-400 dark:text-makoclaw-text-secondary">{{ row.selfHosted }}</div>
            <div class="p-4 sm:p-5 text-blue-600 dark:text-blue-400 font-semibold bg-blue-50/50 dark:bg-blue-500/5">{{ row.saas }}</div>
          </div>
        </div>
      </section>

      <!-- ==================== FEATURES GRID ==================== -->
      <section class="py-16 sm:py-20">
        <h3 class="text-3xl sm:text-4xl font-bold text-center mb-4">
          <span class="bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">Everything you need</span>
        </h3>
        <p class="text-center text-slate-500 dark:text-makoclaw-text-secondary mb-12 text-lg">Powerful tools for building autonomous AI systems</p>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div v-for="feat in features" :key="feat.title"
            class="rounded-2xl p-7 border transition-all bg-white/60 border-slate-200/80 hover:border-blue-300 hover:shadow-lg hover:shadow-blue-100/50 dark:bg-makoclaw-surface/50 dark:border-makoclaw-border dark:hover:border-blue-500/50 dark:hover:shadow-none"
          >
            <div class="w-12 h-12 rounded-xl flex items-center justify-center mb-4" :class="feat.iconBg">
              <span class="text-2xl">{{ feat.icon }}</span>
            </div>
            <h4 class="text-lg font-bold text-slate-900 dark:text-makoclaw-text mb-2">{{ feat.title }}</h4>
            <p class="text-slate-500 dark:text-makoclaw-text-secondary text-sm leading-relaxed">{{ feat.desc }}</p>
          </div>
        </div>
      </section>

      <!-- ==================== CHANNELS ==================== -->
      <section class="py-16 sm:py-20">
        <h3 class="text-3xl font-bold text-center mb-4">
          <span class="bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">10+ Channels</span>
        </h3>
        <p class="text-center text-slate-500 dark:text-makoclaw-text-secondary mb-10 text-lg">One agent, everywhere your users are</p>

        <div class="grid grid-cols-3 sm:grid-cols-5 gap-3 sm:gap-4 max-w-3xl mx-auto">
          <div v-for="ch in channels" :key="ch.name"
            class="rounded-xl p-3 sm:p-4 text-center border transition-all bg-white/60 border-slate-200/80 hover:border-blue-300 dark:bg-makoclaw-surface/50 dark:border-makoclaw-border dark:hover:border-blue-500/50"
          >
            <div class="text-2xl sm:text-3xl mb-1.5">{{ ch.emoji }}</div>
            <div class="font-medium text-xs sm:text-sm text-slate-700 dark:text-makoclaw-text">{{ ch.name }}</div>
          </div>
        </div>
      </section>

      <!-- ==================== PRICING ==================== -->
      <section id="pricing" class="py-16 sm:py-20">
        <h3 class="text-3xl sm:text-4xl font-bold text-center mb-4">
          <span class="text-slate-900 dark:text-makoclaw-text">Early Adopter </span>
          <span class="bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">Pricing</span>
        </h3>
        <p class="text-center text-slate-500 dark:text-makoclaw-text-secondary mb-12 text-lg">
          First users lock this price forever. It will go up for new signups.
        </p>

        <div class="max-w-md mx-auto">
          <div class="rounded-2xl p-8 sm:p-10 border-2 text-center relative overflow-hidden bg-white border-blue-500/30 shadow-2xl shadow-blue-100/50 dark:bg-makoclaw-surface dark:border-blue-500/30 dark:shadow-blue-500/5">
            <div class="absolute top-0 right-0 bg-gradient-to-l from-blue-500 to-cyan-500 text-white text-xs font-bold px-4 py-1.5 rounded-bl-xl">
              LAUNCH PRICE
            </div>

            <h4 class="text-lg font-bold text-slate-900 dark:text-makoclaw-text mb-6">MakoClaw Pro</h4>

            <div class="mb-6">
              <span class="text-2xl text-slate-400 dark:text-makoclaw-text-secondary line-through mr-2">$10</span>
              <span class="text-6xl font-bold bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">$5</span>
              <span class="text-slate-500 dark:text-makoclaw-text-secondary text-lg">/month</span>
            </div>

            <p class="text-sm text-slate-400 dark:text-makoclaw-text-secondary mb-8">
              Locked forever for early adopters
            </p>

            <ul class="text-left space-y-3 mb-8 text-sm">
              <li v-for="perk in pricingPerks" :key="perk" class="flex items-center gap-3">
                <svg class="w-5 h-5 text-blue-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                </svg>
                <span class="text-slate-600 dark:text-makoclaw-text-secondary">{{ perk }}</span>
              </li>
            </ul>

            <router-link
              to="/signup"
              class="block w-full py-4 bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 text-white font-bold rounded-xl transition-all shadow-lg shadow-blue-500/25 hover:shadow-blue-500/40 active:scale-[0.98] text-lg"
            >
              Get Started Now
            </router-link>

            <p class="text-xs text-slate-400 dark:text-makoclaw-text-secondary/60 mt-4">No credit card required to start</p>
          </div>
        </div>
      </section>

      <!-- ==================== TRUST & SECURITY ==================== -->
      <section class="py-16 sm:py-20">
        <h3 class="text-3xl sm:text-4xl font-bold text-center mb-12">
          <span class="text-slate-900 dark:text-makoclaw-text">Built on </span>
          <span class="bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">trust</span>
        </h3>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-4xl mx-auto">
          <div v-for="trust in trustSignals" :key="trust.title"
            class="rounded-2xl p-7 border text-center bg-white/60 border-slate-200/80 dark:bg-makoclaw-surface/50 dark:border-makoclaw-border"
          >
            <div class="w-14 h-14 mx-auto mb-4 rounded-xl flex items-center justify-center" :class="trust.iconBg">
              <svg class="w-7 h-7" :class="trust.iconColor" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="trust.iconPath"></path>
              </svg>
            </div>
            <h4 class="text-lg font-bold text-slate-900 dark:text-makoclaw-text mb-2">{{ trust.title }}</h4>
            <p class="text-sm text-slate-500 dark:text-makoclaw-text-secondary leading-relaxed">{{ trust.desc }}</p>
          </div>
        </div>
      </section>

      <!-- ==================== FAQ ==================== -->
      <section class="py-16 sm:py-20">
        <h3 class="text-3xl sm:text-4xl font-bold text-center mb-12">
          <span class="text-slate-900 dark:text-makoclaw-text">Frequently Asked </span>
          <span class="bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">Questions</span>
        </h3>

        <div class="max-w-2xl mx-auto space-y-3">
          <div v-for="(faq, i) in faqs" :key="i"
            class="rounded-xl border overflow-hidden transition-all bg-white/60 border-slate-200/80 dark:bg-makoclaw-surface/50 dark:border-makoclaw-border"
          >
            <button
              @click="toggleFaq(i)"
              class="w-full flex items-center justify-between p-5 text-left font-medium text-slate-900 dark:text-makoclaw-text hover:bg-slate-50/50 dark:hover:bg-makoclaw-bg/30 transition-colors"
            >
              <span>{{ faq.q }}</span>
              <svg class="w-5 h-5 flex-shrink-0 ml-4 transition-transform duration-200 text-slate-400 dark:text-makoclaw-text-secondary"
                :class="{ 'rotate-180': openFaq === i }"
                fill="none" stroke="currentColor" viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
              </svg>
            </button>
            <Transition name="expand">
              <div v-if="openFaq === i" class="px-5 pb-5 text-sm text-slate-500 dark:text-makoclaw-text-secondary leading-relaxed">
                {{ faq.a }}
              </div>
            </Transition>
          </div>
        </div>
      </section>

      <!-- ==================== STATS ==================== -->
      <section class="py-16">
        <div class="grid grid-cols-2 md:grid-cols-5 gap-8 text-center">
          <div>
            <div class="text-4xl font-bold text-blue-500 mb-2">20+</div>
            <div class="text-slate-500 dark:text-makoclaw-text-secondary text-sm">Built-in Tools</div>
          </div>
          <div>
            <div class="text-4xl font-bold text-cyan-500 mb-2">10+</div>
            <div class="text-slate-500 dark:text-makoclaw-text-secondary text-sm">Channels</div>
          </div>
          <div>
            <div class="text-4xl font-bold text-teal-500 mb-2">∞</div>
            <div class="text-slate-500 dark:text-makoclaw-text-secondary text-sm">Workflows</div>
          </div>
          <div>
            <div class="text-4xl font-bold text-indigo-500 mb-2">&lt;1s</div>
            <div class="text-slate-500 dark:text-makoclaw-text-secondary text-sm">Boot Time</div>
          </div>
          <div>
            <div class="text-4xl font-bold text-blue-500 mb-2">&lt;10MB</div>
            <div class="text-slate-500 dark:text-makoclaw-text-secondary text-sm">RAM Usage</div>
          </div>
        </div>
      </section>

      <!-- ==================== FINAL CTA ==================== -->
      <section class="py-16 sm:py-20">
        <div class="rounded-3xl p-10 sm:p-14 border-2 text-center bg-gradient-to-br from-white to-blue-50/50 border-blue-200/50 shadow-2xl shadow-blue-100/30 dark:from-makoclaw-surface dark:to-makoclaw-bg dark:border-blue-500/30 dark:shadow-blue-500/5">
          <h3 class="text-3xl sm:text-4xl font-bold text-slate-900 dark:text-makoclaw-text mb-4">Ready to deploy your AI shark?</h3>
          <p class="text-lg text-slate-500 dark:text-makoclaw-text-secondary mb-8 max-w-xl mx-auto">
            Join the early adopters building the future of autonomous AI agents with MakoClaw.
          </p>
          <router-link
            to="/signup"
            class="inline-block px-12 py-4 bg-gradient-to-r from-blue-500 to-cyan-500 hover:from-blue-600 hover:to-cyan-600 text-white text-lg rounded-xl transition-all shadow-2xl shadow-blue-500/30 font-bold active:scale-[0.98]"
          >
            Create Free Account →
          </router-link>
        </div>
      </section>

    </div>

    <!-- Footer -->
    <footer class="border-t mt-8 border-slate-200 dark:border-makoclaw-border">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div class="text-center text-slate-400 dark:text-makoclaw-text-secondary text-sm">
          <p>&copy; 2026 MakoClaw. The Apex AI Agent Platform.</p>
          <p class="mt-2">Your data, your API keys, your infrastructure.</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const openFaq = ref(null)
const toggleFaq = (i) => {
  openFaq.value = openFaq.value === i ? null : i
}

const comparisonRows = [
  { feature: 'Installation', selfHosted: 'Docker / VPS', saas: 'Zero install' },
  { feature: 'Monthly cost', selfHosted: 'VPS $5-10+', saas: '$5/mo (promo)' },
  { feature: 'Maintenance', selfHosted: 'You do it', saas: 'We handle it' },
  { feature: 'Updates', selfHosted: 'Manual', saas: 'Automatic' },
  { feature: 'Setup time', selfHosted: '1-2 hours', saas: '< 5 minutes' },
  { feature: 'Technical skill', selfHosted: 'Required', saas: 'None needed' },
]

const features = [
  { icon: '🤖', title: 'Multi-Agent System', desc: 'Create specialized AI agents. The Orchestrator auto-delegates tasks to the right specialist.', iconBg: 'bg-blue-100 dark:bg-blue-500/10' },
  { icon: '💬', title: '10+ Channels', desc: 'Web, Telegram, Discord, Slack, WhatsApp, Signal, and more — one agent, everywhere.', iconBg: 'bg-cyan-100 dark:bg-cyan-500/10' },
  { icon: '⚡', title: 'Visual Workflows', desc: 'Build automation pipelines with drag-and-drop. Connect tools, prompts, and conditions visually.', iconBg: 'bg-teal-100 dark:bg-teal-500/10' },
  { icon: '📋', title: 'Kanban Task Board', desc: 'Full task management with Backlog, To Do, In Progress, Review, Done. AI creates and updates tasks.', iconBg: 'bg-indigo-100 dark:bg-indigo-500/10' },
  { icon: '📚', title: 'Knowledge Base (RAG)', desc: 'Upload documents for semantic search. Supports PDF, TXT, MD, JSON, CSV, and more.', iconBg: 'bg-blue-100 dark:bg-blue-500/10' },
  { icon: '⏰', title: 'Cron Jobs & Scheduling', desc: 'Schedule tasks and automated reports. Your AI works even when you sleep.', iconBg: 'bg-cyan-100 dark:bg-cyan-500/10' },
]

const channels = [
  { emoji: '🌐', name: 'Web' },
  { emoji: '✈️', name: 'Telegram' },
  { emoji: '🎮', name: 'Discord' },
  { emoji: '💼', name: 'Slack' },
  { emoji: '📱', name: 'WhatsApp' },
  { emoji: '🔐', name: 'Signal' },
  { emoji: '🐧', name: 'QQ' },
  { emoji: '💬', name: 'DingTalk' },
  { emoji: '📋', name: 'Feishu' },
  { emoji: '📷', name: 'MaixCam' },
]

const pricingPerks = [
  'Unlimited AI agents & specialists',
  'All 10+ channels included',
  'Knowledge base with RAG search',
  'Visual workflow builder',
  'Kanban task management',
  'Cron jobs & scheduling',
  'Automatic updates & maintenance',
  'Priority support',
]

const trustSignals = [
  {
    title: 'Encrypted API Keys',
    desc: 'Your API keys are encrypted at rest and never shared. Only you have access.',
    iconBg: 'bg-blue-100 dark:bg-blue-500/10',
    iconColor: 'text-blue-500',
    iconPath: 'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z',
  },
  {
    title: 'Channel Privacy',
    desc: 'We only read messages from channels you explicitly authorize. Nothing more.',
    iconBg: 'bg-cyan-100 dark:bg-cyan-500/10',
    iconColor: 'text-cyan-500',
    iconPath: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z',
  },
  {
    title: 'Your Data, Your Control',
    desc: 'Export or delete your data at any time. No vendor lock-in, no hidden data usage.',
    iconBg: 'bg-teal-100 dark:bg-teal-500/10',
    iconColor: 'text-teal-500',
    iconPath: 'M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4',
  },
]

const faqs = [
  { q: 'Do I need a credit card to start?', a: 'No. You can create an account and explore MakoClaw completely free. We only ask for payment when you\'re ready to go live with your channels.' },
  { q: 'Can I use my own API key?', a: 'Absolutely. MakoClaw supports OpenAI, Anthropic, Groq, Ollama, and many more providers. You bring your own API key and only pay the provider for what you use.' },
  { q: 'What happens if the price goes up?', a: 'Early adopters lock the $5/month price forever. Even when we raise prices for new users, your rate stays the same as long as your account is active.' },
  { q: 'How is this different from self-hosting?', a: 'With self-hosting you need a VPS, Docker, manual updates, and troubleshooting. MakoClaw handles all infrastructure — you just configure your AI and channels through a clean web interface.' },
  { q: 'Can I migrate from a self-hosted setup?', a: 'Yes. MakoClaw uses the same configuration format. You can import your existing config, knowledge base, and sessions with minimal effort.' },
  { q: 'Is my data safe?', a: 'Your API keys are encrypted at rest. We only access channels you explicitly authorize. You can export or delete all your data at any time.' },
]
</script>

<style scoped>
.expand-enter-active,
.expand-leave-active {
  transition: all 0.2s ease;
  overflow: hidden;
}
.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  padding-bottom: 0;
}
.expand-enter-to,
.expand-leave-from {
  opacity: 1;
  max-height: 200px;
}
</style>
