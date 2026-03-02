<template>
  <div class="h-full flex flex-col bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-rose-500/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-red-500/20 via-transparent to-transparent" />
    </div>

    <!-- Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
        <div class="flex items-center gap-3">
          <!-- Icon Container -->
          <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-rose-500/20 to-red-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-rose-500/10">
            <svg
              class="w-5 h-5 sm:w-6 sm:h-6 text-rose-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
          </div>

          <div class="flex-1 min-w-0">
            <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-rose-400 bg-clip-text text-transparent">
              Reports
            </h1>
            <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5 hidden sm:block">
              Analytics and performance insights
            </p>
          </div>
        </div>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto p-4 md:p-6 space-y-6 relative z-10 custom-scrollbar">
      <!-- Top Stats Section -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <!-- Average Response Time -->
        <div class="glass-panel p-5 rounded-2xl relative overflow-hidden group/stat hover:-translate-y-1 transition-all duration-300 shadow-lg shadow-makoclaw-accent/5">
          <!-- Hover glow line -->
          <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover/stat:w-full transition-all duration-500 opacity-50" />
          
          <div class="flex items-center gap-4">
            <div class="p-3 bg-makoclaw-accent/10 rounded-xl text-makoclaw-accent relative">
              <!-- Icon glow -->
              <div class="absolute inset-0 bg-makoclaw-accent/20 blur-lg rounded-full opacity-0 group-hover/stat:opacity-100 transition-opacity duration-500" />
              <svg 
                class="w-6 h-6 relative z-10" 
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <div>
              <p class="text-xs text-makoclaw-text-secondary font-medium uppercase tracking-wider opacity-60">
                Avg Response Time
              </p>
              <h4 class="text-2xl font-bold text-makoclaw-text">
                {{ reportsData?.global_metrics?.avg_response_time || '0.0s' }}
              </h4>
            </div>
          </div>
        </div>

        <!-- Success Rate -->
        <div class="glass-panel p-5 rounded-2xl relative overflow-hidden group/stat hover:-translate-y-1 transition-all duration-300 shadow-lg shadow-makoclaw-accent/5">
          <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-green-500 to-emerald-400 group-hover/stat:w-full transition-all duration-500 opacity-50" />
          <div class="flex items-center gap-4">
            <div class="p-3 bg-green-500/10 rounded-xl text-green-500 relative">
              <div class="absolute inset-0 bg-green-500/20 blur-lg rounded-full opacity-0 group-hover/stat:opacity-100 transition-opacity duration-500" />
              <svg 
                class="w-6 h-6 relative z-10" 
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <div>
              <p class="text-xs text-makoclaw-text-secondary font-medium uppercase tracking-wider opacity-60">
                Success Rate
              </p>
              <h4 class="text-2xl font-bold text-makoclaw-text">
                {{ reportsData?.global_metrics?.success_rate || '0%' }}
              </h4>
            </div>
          </div>
        </div>

        <!-- Total Sessions -->
        <div class="glass-panel p-5 rounded-2xl relative overflow-hidden group/stat hover:-translate-y-1 transition-all duration-300 shadow-lg shadow-makoclaw-accent/5">
          <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-blue-500 to-indigo-400 group-hover/stat:w-full transition-all duration-500 opacity-50" />
          <div class="flex items-center gap-4">
            <div class="p-3 bg-blue-500/10 rounded-xl text-blue-500 relative">
              <div class="absolute inset-0 bg-blue-500/20 blur-lg rounded-full opacity-0 group-hover/stat:opacity-100 transition-opacity duration-500" />
              <svg 
                class="w-6 h-6 relative z-10" 
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M17 8h2a2 2 0 012 2v6a2 2 0 01-2 2h-2v4l-4-4H9a1.994 1.994 0 01-1.414-.586m0 0L11 14h4a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2v4l.586-.586z"
                />
              </svg>
            </div>
            <div>
              <p class="text-xs text-makoclaw-text-secondary font-medium uppercase tracking-wider opacity-60">
                Total Sessions
              </p>
              <h4 class="text-2xl font-bold text-makoclaw-text">
                {{ reportsData?.global_metrics?.total_sessions || '0' }}
              </h4>
            </div>
          </div>
        </div>

        <!-- Errors Count -->
        <div class="glass-panel p-5 rounded-2xl relative overflow-hidden group/stat hover:-translate-y-1 transition-all duration-300 shadow-lg shadow-makoclaw-accent/5">
          <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-red-500 to-rose-400 group-hover/stat:w-full transition-all duration-500 opacity-50" />
          <div class="flex items-center gap-4">
            <div class="p-3 bg-red-500/10 rounded-xl text-red-500 relative">
              <div class="absolute inset-0 bg-red-500/20 blur-lg rounded-full opacity-0 group-hover/stat:opacity-100 transition-opacity duration-500" />
              <svg 
                class="w-6 h-6 relative z-10" 
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                />
              </svg>
            </div>
            <div>
              <p class="text-xs text-makoclaw-text-secondary font-medium uppercase tracking-wider opacity-60">
                System Errors
              </p>
              <h4 class="text-2xl font-bold text-makoclaw-text">
                {{ reportsData?.global_metrics?.error_count || '0' }}
              </h4>
            </div>
          </div>
        </div>
      </div>

      <!-- Agent Performance Table -->
      <section class="glass-panel p-6 rounded-2xl shadow-xl shadow-makoclaw-accent/5 relative overflow-hidden group/table">
        <div class="flex items-center gap-3 mb-6">
          <div class="p-2 bg-makoclaw-accent/10 rounded-xl text-makoclaw-accent">
            <svg
              class="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
              />
            </svg>
          </div>
          <h3 class="text-lg font-bold text-makoclaw-text">
            Agent Performance
          </h3>
          <div class="flex items-center gap-3 ml-auto">
            <select
              v-model="metricsPeriod"
              class="px-3 py-2 bg-makoclaw-bg/60 border border-makoclaw-border/50 rounded-lg text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent outline-none cursor-pointer"
              @change="fetchAgentMetrics"
            >
              <option value="24h">
                Last 24 Hours
              </option>
              <option value="7d">
                Last 7 Days
              </option>
              <option value="30d">
                Last 30 Days
              </option>
            </select>
            <button
              :disabled="!hasSpecialistData"
              class="flex items-center gap-2 px-3 py-2 text-sm bg-makoclaw-bg/60 border border-makoclaw-border/50 rounded-lg hover:border-makoclaw-accent/50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              @click="exportCSV"
            >
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
              Export CSV
            </button>
          </div>
        </div>

        <div
          v-if="loadingMetrics"
          class="flex items-center justify-center py-8"
        >
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-makoclaw-accent" />
        </div>

        <div
          v-else-if="agentMetrics"
          class="space-y-6"
        >
          <!-- Overview Cards -->
          <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div class="p-4 bg-makoclaw-bg/40 rounded-lg">
              <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-1">
                Total Cost
              </p>
              <p class="text-xl font-bold text-makoclaw-text">
                {{ formatCost(agentMetrics.total_cost || 0) }}
              </p>
            </div>
            <div class="p-4 bg-makoclaw-bg/40 rounded-lg">
              <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-1">
                Total Calls
              </p>
              <p class="text-xl font-bold text-makoclaw-text">
                {{ agentMetrics.total_calls || 0 }}
              </p>
            </div>
            <div class="p-4 bg-makoclaw-bg/40 rounded-lg">
              <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-1">
                Total Tokens
              </p>
              <p class="text-xl font-bold text-makoclaw-text">
                {{ formatNumber(agentMetrics.total_tokens || 0) }}
              </p>
            </div>
            <div class="p-4 bg-makoclaw-bg/40 rounded-lg">
              <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-1">
                Active Agents
              </p>
              <p class="text-xl font-bold text-makoclaw-text">
                {{ agentMetrics.by_specialist ? Object.keys(agentMetrics.by_specialist).length : 0 }}
              </p>
            </div>
          </div>

          <!-- Charts Section -->
          <div
            v-if="hasSpecialistData"
            class="grid grid-cols-1 lg:grid-cols-2 gap-6"
          >
            <!-- Cost Breakdown Chart -->
            <div class="p-4 bg-makoclaw-bg/30 rounded-lg">
              <h4 class="font-bold text-makoclaw-text mb-4 text-sm">
                Cost by Specialist
              </h4>
              <div class="space-y-3">
                <div
                  v-for="(metrics, name) in sortedByField('cost')"
                  :key="'cost-' + name"
                  class="flex items-center gap-3"
                >
                  <div class="w-20 text-xs text-makoclaw-text truncate capitalize">
                    {{ name }}
                  </div>
                  <div class="flex-1 h-6 bg-makoclaw-bg/50 rounded overflow-hidden">
                    <div
                      :style="{ width: getCostPercentage(metrics.cost) + '%' }"
                      :class="getBarColor(name)"
                      class="h-full transition-all duration-300"
                    />
                  </div>
                  <div class="w-16 text-right text-xs font-medium text-makoclaw-text">
                    {{ formatCost(metrics.cost) }}
                  </div>
                </div>
              </div>
            </div>

            <!-- Token Usage Chart -->
            <div class="p-4 bg-makoclaw-bg/30 rounded-lg">
              <h4 class="font-bold text-makoclaw-text mb-4 text-sm">
                Token Usage by Specialist
              </h4>
              <div class="space-y-3">
                <div
                  v-for="(metrics, name) in sortedByField('tokens')"
                  :key="'tokens-' + name"
                  class="flex items-center gap-3"
                >
                  <div class="w-20 text-xs text-makoclaw-text truncate capitalize">
                    {{ name }}
                  </div>
                  <div class="flex-1 h-6 bg-makoclaw-bg/50 rounded overflow-hidden">
                    <div
                      :style="{ width: getTokenPercentage(metrics.tokens) + '%' }"
                      :class="getBarColor(name)"
                      class="h-full transition-all duration-300"
                    />
                  </div>
                  <div class="w-20 text-right text-xs font-medium text-makoclaw-text">
                    {{ formatCompact(metrics.tokens) }}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Cost by Specialist Table -->
          <div>
            <h4 class="font-bold text-makoclaw-text mb-3">
              Detailed Breakdown
            </h4>
            <div class="overflow-x-auto rounded-xl border border-makoclaw-border/50">
              <table class="w-full text-sm">
                <thead class="bg-makoclaw-bg/50 text-makoclaw-text-secondary">
                  <tr class="border-b border-makoclaw-border">
                    <th class="text-left py-2 px-3 text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary">
                      Specialist
                    </th>
                    <th class="text-right py-2 px-3 text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary">
                      Calls
                    </th>
                    <th class="text-right py-2 px-3 text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary">
                      Tokens
                    </th>
                    <th class="text-right py-2 px-3 text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary">
                      Avg Cost
                    </th>
                    <th class="text-right py-2 px-3 text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary">
                      Total Cost
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="(metrics, name) in (agentMetrics.by_specialist || {})"
                    :key="name"
                    class="border-b border-makoclaw-border/50 hover:bg-makoclaw-bg/30"
                  >
                    <td class="py-3 px-3">
                      <div class="flex items-center gap-2">
                        <div :class="`w-6 h-6 rounded ${getSpecialistBgColor(name)} flex items-center justify-center`">
                          <svg
                            class="w-3.5 h-3.5"
                            :class="getSpecialistColor(name)"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                            v-html="getSpecialistIcon(name)"
                          />
                        </div>
                        <span class="font-medium text-makoclaw-text capitalize">{{ name }}</span>
                      </div>
                    </td>
                    <td class="text-right py-3 px-3 text-makoclaw-text">
                      {{ metrics.calls || 0 }}
                    </td>
                    <td class="text-right py-3 px-3 text-makoclaw-text">
                      {{ formatNumber(metrics.tokens || 0) }}
                    </td>
                    <td class="text-right py-3 px-3 text-makoclaw-text">
                      {{ formatCost(metrics.avg_cost || 0) }}
                    </td>
                    <td class="text-right py-3 px-3 font-bold text-makoclaw-text">
                      {{ formatCost(metrics.cost || 0) }}
                    </td>
                  </tr>
                  <tr v-if="!agentMetrics.by_specialist || Object.keys(agentMetrics.by_specialist).length === 0">
                    <td
                      colspan="5"
                      class="py-8 text-center text-makoclaw-text-secondary"
                    >
                      No agent data available
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <div
          v-else
          class="p-8 text-center text-makoclaw-text-secondary"
        >
          <svg
            class="w-12 h-12 mx-auto mb-3 opacity-50"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
          <p>No metrics available for the selected period</p>
        </div>
      </section>

      <!-- Email Report Section -->
      <section class="glass-panel p-6 rounded-2xl shadow-xl shadow-makoclaw-accent/5 relative overflow-hidden group/form">
        <div class="flex items-center gap-3 mb-6">
          <div class="p-2 bg-blue-500/10 rounded-xl text-blue-400">
            <svg
              class="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
              />
            </svg>
          </div>
          <h3 class="text-lg font-bold text-makoclaw-text">
            Email Summary Report
          </h3>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-4">
            <div class="relative group/input">
              <label class="block text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider mb-1.5 ml-1">Recipient Email</label>
              <input
                v-model="emailForm.recipient"
                type="email"
                placeholder="analytics@company.com"
                class="w-full bg-makoclaw-bg/50 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm text-makoclaw-text outline-none focus:border-makoclaw-accent/50 transition-all duration-300"
              >
              <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-makoclaw-accent transition-all duration-500 opacity-50 focus-within:w-full" />
            </div>
            
            <div class="relative group/input">
              <label class="block text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider mb-1.5 ml-1">Title / Subject</label>
              <input
                v-model="emailForm.title"
                type="text"
                placeholder="Weekly Performance Update"
                class="w-full bg-makoclaw-bg/50 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm text-makoclaw-text outline-none focus:border-makoclaw-accent/50 transition-all duration-300"
              >
              <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-makoclaw-accent transition-all duration-500 opacity-50 focus-within:w-full" />
            </div>
          </div>
          
          <div class="space-y-4 flex flex-col justify-end">
            <button
              :disabled="sendingReport"
              class="w-full py-3 bg-gradient-to-r from-makoclaw-accent to-blue-500 text-white font-bold rounded-xl hover:shadow-lg hover:shadow-makoclaw-accent/20 transition-all duration-300 flex items-center justify-center gap-2 hover:-translate-y-0.5 active:translate-y-0 disabled:opacity-50"
              @click="sendReport"
            >
              <div
                v-if="sendingReport"
                class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
              />
              <svg
                v-else
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
                />
              </svg>
              {{ sendingReport ? 'Sending...' : 'Send Summary Report' }}
            </button>
            <p class="text-[10px] text-makoclaw-text-secondary text-center opacity-60 px-4">
              A comprehensive daily summary will be compiled and sent to the recipient securely.
            </p>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAgentsStore } from '../stores/agentsStore'
import { useToast } from '../composables/useToast'
import advancedService from '../services/advancedService'

const agentsStore = useAgentsStore()
const toast = useToast()

const emailForm = ref({
  recipient: '',
  title: '',
})
const sendingReport = ref(false)
const metricsPeriod = ref('7d')
const loadingMetrics = ref(false)
const agentMetrics = ref(null)
const reportsData = ref(null) // Added for global metrics

// Chart colors for specialists
const specialistColors = {
  orchestrator: 'bg-purple-500',
  developer: 'bg-blue-500',
  reviewer: 'bg-green-500',
  researcher: 'bg-yellow-500',
  writer: 'bg-pink-500',
  analyst: 'bg-cyan-500',
  default: 'bg-makoclaw-accent'
}

const hasSpecialistData = computed(() => {
  return agentMetrics.value?.by_specialist && Object.keys(agentMetrics.value.by_specialist).length > 0
})

const maxCost = computed(() => {
  if (!agentMetrics.value?.by_specialist) return 0
  return Math.max(...Object.values(agentMetrics.value.by_specialist).map(m => m.cost || 0), 0.0001)
})

const maxTokens = computed(() => {
  if (!agentMetrics.value?.by_specialist) return 0
  return Math.max(...Object.values(agentMetrics.value.by_specialist).map(m => m.tokens || 0), 1)
})

onMounted(() => {
  fetchAgentMetrics()
  fetchReportsData() // Fetch global metrics on mount
})

async function fetchAgentMetrics() {
  loadingMetrics.value = true
  await agentsStore.fetchMetrics(metricsPeriod.value)
  agentMetrics.value = agentsStore.metrics
  loadingMetrics.value = false
}

async function fetchReportsData() {
  try {
    reportsData.value = await advancedService.getReportsData()
  } catch (error) {
    console.error("Failed to fetch reports data:", error)
    toast.error("Failed to load global reports data.")
  }
}

function sortedByField(field) {
  if (!agentMetrics.value?.by_specialist) return {}
  const entries = Object.entries(agentMetrics.value.by_specialist)
  entries.sort((a, b) => (b[1][field] || 0) - (a[1][field] || 0))
  return Object.fromEntries(entries)
}

function getCostPercentage(cost) {
  return Math.max(((cost || 0) / maxCost.value) * 100, 2)
}

function getTokenPercentage(tokens) {
  return Math.max(((tokens || 0) / maxTokens.value) * 100, 2)
}

function getBarColor(name) {
  const normalized = name.toLowerCase()
  for (const [key, color] of Object.entries(specialistColors)) {
    if (normalized.includes(key)) return color
  }
  return specialistColors.default
}

function getSpecialistIcon(name) {
  return agentsStore.getSpecialistIcon(name)
}

function getSpecialistColor(name) {
  return agentsStore.getSpecialistColor(name)
}

function getSpecialistBgColor(name) {
  return agentsStore.getSpecialistBgColor(name)
}

function formatCost(cost) {
  if (!cost) return '$0.00'
  return '$' + cost.toFixed(4)
}

function formatNumber(num) {
  if (!num) return '0'
  return num.toLocaleString()
}

function formatCompact(num) {
  if (!num) return '0'
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return num.toString()
}

function exportCSV() {
  if (!agentMetrics.value?.by_specialist) {
    toast.error('No metrics data to export')
    return
  }

  const specialists = agentMetrics.value.by_specialist
  const rows = [['Specialist', 'Calls', 'Tokens', 'Avg Cost', 'Total Cost']]

  for (const [name, metrics] of Object.entries(specialists)) {
    rows.push([
      name,
      (metrics.calls || 0).toString(),
      (metrics.tokens || 0).toString(),
      (metrics.avg_cost || 0).toFixed(4),
      (metrics.cost || 0).toFixed(4)
    ])
  }

  // Add totals row
  rows.push([
    'TOTAL',
    (agentMetrics.value.total_calls || 0).toString(),
    (agentMetrics.value.total_tokens || 0).toString(),
    '',
    (agentMetrics.value.total_cost || 0).toFixed(4)
  ])

  // Generate CSV content
  const csvContent = rows.map(row => row.map(cell => {
    // Escape cells containing commas or quotes
    if (cell.includes(',') || cell.includes('"') || cell.includes('\n')) {
      return '"' + cell.replace(/"/g, '""') + '"'
    }
    return cell
  }).join(',')).join('\n')

  // Create and download file
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `makoclaw-metrics-${metricsPeriod.value}.csv`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)

  toast.success('CSV exported successfully')
}

const sendReport = async () => {
    if (!emailForm.value.title || !emailForm.value.recipient) {
      toast.error('Recipient and Subject are required.')
      return
    }

    sendingReport.value = true

    try {
        const result = await advancedService.sendReportEmail(emailForm.value.recipient, emailForm.value.title)

        if (result.success) {
            toast.success(result.message || 'Report sent successfully!')
            // Clear form on success
            emailForm.value.recipient = ''
            emailForm.value.title = ''
        } else {
            toast.error(result.error || 'Failed to send report')
        }
    } catch (err) {
        console.error("Failed to send report:", err)
        const errorMsg = err.response?.data?.error || err.message || 'Failed to send email'
        toast.error(errorMsg)
    } finally {
        sendingReport.value = false
    }
}
</script>

<style scoped>
.glass-sticky {
  background: rgba(var(--makoclaw-surface-rgb), 0.7);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}
</style>
