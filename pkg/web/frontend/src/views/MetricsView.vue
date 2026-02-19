<template>
  <div class="h-full flex flex-col max-w-5xl mx-auto w-full p-4 md:p-8 overflow-y-auto">
    <div class="flex-none mb-8">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-3xl font-bold bg-gradient-to-r from-kakoclaw-accent to-purple-500 bg-clip-text text-transparent mb-2">Metrics</h2>
          <p class="text-kakoclaw-text-secondary">In-process observability for LLM calls, tool executions, and agent runs.</p>
        </div>
        <button
          @click="loadMetrics"
          :disabled="loading"
          class="flex items-center gap-2 px-4 py-2 bg-kakoclaw-accent text-white rounded-lg hover:bg-kakoclaw-accent/90 transition-all font-medium disabled:opacity-50"
        >
          <div v-if="loading" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
          <span>Refresh</span>
        </button>
      </div>
    </div>

    <!-- Error state -->
    <div v-if="error" class="bg-red-500/10 border border-red-500/30 rounded-xl p-4 mb-6 text-red-400">
      {{ error }}
    </div>

    <!-- Uptime -->
    <div v-if="metrics" class="text-sm text-kakoclaw-text-secondary mb-6">
      Uptime: {{ formatUptime(metrics.uptime_seconds) }} | Started: {{ formatDate(metrics.started_at) }}
    </div>

    <!-- Summary Cards -->
    <div v-if="metrics" class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
      <!-- LLM Card -->
      <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 rounded-lg bg-blue-500/10 flex items-center justify-center">
            <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
          </div>
          <h3 class="text-lg font-semibold">LLM Calls</h3>
        </div>
        <div class="space-y-2 text-sm">
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Total calls</span><span class="font-mono font-medium">{{ metrics.llm_calls }}</span></div>
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Errors</span><span class="font-mono font-medium" :class="metrics.llm_errors > 0 ? 'text-red-400' : ''">{{ metrics.llm_errors }}</span></div>
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Avg latency</span><span class="font-mono font-medium">{{ formatMs(metrics.llm_avg_ms) }}</span></div>
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Tokens in</span><span class="font-mono font-medium">{{ formatNumber(metrics.llm_tokens_in) }}</span></div>
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Tokens out</span><span class="font-mono font-medium">{{ formatNumber(metrics.llm_tokens_out) }}</span></div>
        </div>
      </div>

      <!-- Tool Card -->
      <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 rounded-lg bg-green-500/10 flex items-center justify-center">
            <svg class="w-5 h-5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
          </div>
          <h3 class="text-lg font-semibold">Tool Calls</h3>
        </div>
        <div class="space-y-2 text-sm">
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Total calls</span><span class="font-mono font-medium">{{ metrics.tool_calls }}</span></div>
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Errors</span><span class="font-mono font-medium" :class="metrics.tool_errors > 0 ? 'text-red-400' : ''">{{ metrics.tool_errors }}</span></div>
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Avg latency</span><span class="font-mono font-medium">{{ formatMs(metrics.tool_avg_ms) }}</span></div>
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Total time</span><span class="font-mono font-medium">{{ formatMs(metrics.tool_total_ms) }}</span></div>
        </div>
      </div>

      <!-- Agent Card -->
      <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 rounded-lg bg-purple-500/10 flex items-center justify-center">
            <svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
          </div>
          <h3 class="text-lg font-semibold">Agent Runs</h3>
        </div>
        <div class="space-y-2 text-sm">
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Total runs</span><span class="font-mono font-medium">{{ metrics.agent_runs }}</span></div>
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Errors</span><span class="font-mono font-medium" :class="metrics.agent_errors > 0 ? 'text-red-400' : ''">{{ metrics.agent_errors }}</span></div>
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Avg latency</span><span class="font-mono font-medium">{{ formatMs(metrics.agent_avg_ms) }}</span></div>
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Total iterations</span><span class="font-mono font-medium">{{ metrics.agent_iterations_total }}</span></div>
          <div class="flex justify-between"><span class="text-kakoclaw-text-secondary">Avg iterations</span><span class="font-mono font-medium">{{ metrics.agent_avg_iterations?.toFixed(1) || '0' }}</span></div>
        </div>
      </div>
    </div>

    <!-- Per-model breakdown -->
    <div v-if="metrics && Object.keys(metrics.llm_by_model || {}).length > 0" class="mb-8">
      <h3 class="text-xl font-semibold mb-4">LLM by Model</h3>
      <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl overflow-hidden">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-kakoclaw-border text-kakoclaw-text-secondary">
              <th class="text-left px-4 py-3 font-medium">Model</th>
              <th class="text-right px-4 py-3 font-medium">Calls</th>
              <th class="text-right px-4 py-3 font-medium">Errors</th>
              <th class="text-right px-4 py-3 font-medium">Avg ms</th>
              <th class="text-right px-4 py-3 font-medium">Tokens In</th>
              <th class="text-right px-4 py-3 font-medium">Tokens Out</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(m, name) in metrics.llm_by_model" :key="name" class="border-b border-kakoclaw-border/50 hover:bg-kakoclaw-bg/50">
              <td class="px-4 py-3 font-mono text-xs">{{ name }}</td>
              <td class="px-4 py-3 text-right font-mono">{{ m.calls }}</td>
              <td class="px-4 py-3 text-right font-mono" :class="m.errors > 0 ? 'text-red-400' : ''">{{ m.errors }}</td>
              <td class="px-4 py-3 text-right font-mono">{{ formatMs(m.avg_ms) }}</td>
              <td class="px-4 py-3 text-right font-mono">{{ formatNumber(m.tokens_in) }}</td>
              <td class="px-4 py-3 text-right font-mono">{{ formatNumber(m.tokens_out) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Per-tool breakdown -->
    <div v-if="metrics && Object.keys(metrics.tool_by_name || {}).length > 0" class="mb-8">
      <h3 class="text-xl font-semibold mb-4">Tools by Name</h3>
      <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl overflow-hidden">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-kakoclaw-border text-kakoclaw-text-secondary">
              <th class="text-left px-4 py-3 font-medium">Tool</th>
              <th class="text-right px-4 py-3 font-medium">Calls</th>
              <th class="text-right px-4 py-3 font-medium">Errors</th>
              <th class="text-right px-4 py-3 font-medium">Avg ms</th>
              <th class="text-right px-4 py-3 font-medium">Total ms</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(t, name) in metrics.tool_by_name" :key="name" class="border-b border-kakoclaw-border/50 hover:bg-kakoclaw-bg/50">
              <td class="px-4 py-3 font-mono text-xs">{{ name }}</td>
              <td class="px-4 py-3 text-right font-mono">{{ t.calls }}</td>
              <td class="px-4 py-3 text-right font-mono" :class="t.errors > 0 ? 'text-red-400' : ''">{{ t.errors }}</td>
              <td class="px-4 py-3 text-right font-mono">{{ formatMs(t.avg_ms) }}</td>
              <td class="px-4 py-3 text-right font-mono">{{ formatMs(t.total_ms) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Recent Events -->
    <div v-if="metrics && (metrics.recent_events || []).length > 0" class="mb-8">
      <h3 class="text-xl font-semibold mb-4">Recent Events (last {{ metrics.recent_events.length }})</h3>
      <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl overflow-hidden max-h-96 overflow-y-auto">
        <table class="w-full text-sm">
          <thead class="sticky top-0 bg-kakoclaw-surface">
            <tr class="border-b border-kakoclaw-border text-kakoclaw-text-secondary">
              <th class="text-left px-4 py-3 font-medium">Time</th>
              <th class="text-left px-4 py-3 font-medium">Type</th>
              <th class="text-left px-4 py-3 font-medium">Detail</th>
              <th class="text-right px-4 py-3 font-medium">Duration</th>
              <th class="text-left px-4 py-3 font-medium">Error</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(evt, i) in [...metrics.recent_events].reverse()" :key="i" class="border-b border-kakoclaw-border/50 hover:bg-kakoclaw-bg/50">
              <td class="px-4 py-2 text-xs text-kakoclaw-text-secondary whitespace-nowrap">{{ formatTime(evt.timestamp) }}</td>
              <td class="px-4 py-2">
                <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium"
                  :class="{
                    'bg-blue-500/10 text-blue-400': evt.type === 'llm_call',
                    'bg-green-500/10 text-green-400': evt.type === 'tool_call',
                    'bg-purple-500/10 text-purple-400': evt.type === 'agent_run',
                    'bg-red-500/10 text-red-400': evt.type === 'error'
                  }"
                >{{ evt.type }}</span>
              </td>
              <td class="px-4 py-2 font-mono text-xs">{{ evt.model || evt.tool || '-' }}</td>
              <td class="px-4 py-2 text-right font-mono text-xs">{{ formatMs(evt.duration_ms) }}</td>
              <td class="px-4 py-2 text-xs text-red-400 max-w-xs truncate">{{ evt.error || '' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Empty state -->
    <div v-if="metrics && metrics.llm_calls === 0 && metrics.tool_calls === 0 && metrics.agent_runs === 0" class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-12 text-center">
      <svg class="w-16 h-16 mx-auto mb-4 text-kakoclaw-text-secondary opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16 8v8m-4-5v5m-4-2v2m-2 4h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
      <p class="text-kakoclaw-text-secondary">No metrics collected yet. Start a chat to see LLM and tool usage data.</p>
    </div>

    <!-- Loading state -->
    <div v-if="!metrics && loading" class="flex items-center justify-center py-20">
      <div class="w-8 h-8 border-2 border-kakoclaw-accent/30 border-t-kakoclaw-accent rounded-full animate-spin"></div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import advancedService from '../services/advancedService'

const metrics = ref(null)
const loading = ref(false)
const error = ref('')

async function loadMetrics() {
  loading.value = true
  error.value = ''
  try {
    metrics.value = await advancedService.fetchMetrics()
  } catch (e) {
    error.value = 'Failed to load metrics: ' + (e.response?.data?.error || e.message)
  } finally {
    loading.value = false
  }
}

function formatMs(ms) {
  if (!ms && ms !== 0) return '-'
  if (ms < 1000) return ms + 'ms'
  return (ms / 1000).toFixed(1) + 's'
}

function formatNumber(n) {
  if (!n && n !== 0) return '0'
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return String(n)
}

function formatUptime(seconds) {
  if (!seconds) return '0s'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  if (h > 0) return `${h}h ${m}m`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

function formatDate(iso) {
  if (!iso) return '-'
  return new Date(iso).toLocaleString()
}

function formatTime(iso) {
  if (!iso) return '-'
  return new Date(iso).toLocaleTimeString()
}

let refreshInterval = null

onMounted(() => {
  loadMetrics()
  refreshInterval = setInterval(loadMetrics, 30000)
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
})
</script>
