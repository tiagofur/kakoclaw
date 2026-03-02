<template>
  <div class="h-full flex flex-col max-w-6xl mx-auto w-full p-4 md:p-8">
    <div class="flex-none mb-8">
      <h2 class="text-3xl font-bold bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent mb-2">
        Reports
      </h2>
      <p class="text-makoclaw-text-secondary">
        Generate reports and view agent performance analytics.
      </p>
    </div>

    <div class="space-y-6">
      <!-- Agent Performance Section -->
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl shadow-sm p-6">
        <div class="flex items-center justify-between mb-6">
          <h3 class="font-bold text-makoclaw-text flex items-center gap-2">
            <svg
              class="w-5 h-5 text-makoclaw-accent"
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
            Agent Performance
          </h3>
          <div class="flex items-center gap-3">
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
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
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
      </div>

      <!-- Email Report Section -->
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl shadow-sm p-6 space-y-6">
        <h3 class="font-bold text-makoclaw-text flex items-center gap-2">
          <svg
            class="w-5 h-5 text-blue-400"
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
          Send Email Report
        </h3>

        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium mb-1">To (Email)</label>
            <input
              v-model="to"
              type="email"
              placeholder="Defaults to configured recipient"
              class="w-full bg-makoclaw-bg border border-makoclaw-border rounded-lg px-4 py-2 outline-none focus:border-makoclaw-accent transition-colors"
            >
          </div>

          <div>
            <label class="block text-sm font-medium mb-1">Subject</label>
            <input
              v-model="subject"
              type="text"
              placeholder="Weekly Summary / Project Update"
              class="w-full bg-makoclaw-bg border border-makoclaw-border rounded-lg px-4 py-2 outline-none focus:border-makoclaw-accent transition-colors"
            >
          </div>

          <div>
            <label class="block text-sm font-medium mb-1">Content (Markdown)</label>
            <textarea
              v-model="body"
              rows="10"
              placeholder="Write your report here..."
              class="w-full bg-makoclaw-bg border border-makoclaw-border rounded-lg px-4 py-2 outline-none focus:border-makoclaw-accent transition-colors resize-none font-mono text-sm"
            />
          </div>
        </div>

        <div class="flex justify-end pt-4 border-t border-makoclaw-border">
          <button
            :disabled="sending || !subject || !body"
            class="flex items-center gap-2 px-6 py-2.5 bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-all font-medium disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-makoclaw-accent/20"
            @click="sendReport"
          >
            <div
              v-if="sending"
              class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"
            />
            <span v-else>Send Report</span>
            <svg
              v-if="!sending"
              class="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
            /></svg>
          </button>
        </div>
      </div>
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

const to = ref('')
const subject = ref('')
const body = ref('')
const sending = ref(false)
const metricsPeriod = ref('7d')
const loadingMetrics = ref(false)
const agentMetrics = ref(null)

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
})

async function fetchAgentMetrics() {
  loadingMetrics.value = true
  await agentsStore.fetchMetrics(metricsPeriod.value)
  agentMetrics.value = agentsStore.metrics
  loadingMetrics.value = false
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
    if (!subject.value || !body.value) return

    sending.value = true

    try {
        const result = await advancedService.sendReportEmail(to.value, subject.value, body.value)

        if (result.success) {
            toast.success(result.message || 'Report sent successfully!')
            // Clear form on success
            subject.value = ''
            body.value = ''
            to.value = ''
        } else {
            toast.error(result.error || 'Failed to send report')
        }
    } catch (err) {
        console.error("Failed to send report:", err)
        const errorMsg = err.response?.data?.error || err.message || 'Failed to send email'
        toast.error(errorMsg)
    } finally {
        sending.value = false
    }
}
</script>
