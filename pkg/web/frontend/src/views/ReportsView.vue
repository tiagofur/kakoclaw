<template>
  <div class="h-full flex flex-col max-w-6xl mx-auto w-full p-4 md:p-8">
    <div class="flex-none mb-8">
      <h2 class="text-3xl font-bold bg-gradient-to-r from-kakoclaw-accent to-emerald-500 bg-clip-text text-transparent mb-2">Reports</h2>
      <p class="text-kakoclaw-text-secondary">Generate reports and view agent performance analytics.</p>
    </div>

    <div class="space-y-6">
      <!-- Agent Performance Section -->
      <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl shadow-sm p-6">
        <div class="flex items-center justify-between mb-6">
          <h3 class="font-bold text-kakoclaw-text flex items-center gap-2">
            <svg class="w-5 h-5 text-kakoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
            Agent Performance
          </h3>
          <select
            v-model="metricsPeriod"
            @change="fetchAgentMetrics"
            class="px-3 py-2 bg-kakoclaw-bg/60 border border-kakoclaw-border/50 rounded-lg text-sm focus:ring-2 focus:ring-kakoclaw-accent/30 focus:border-kakoclaw-accent outline-none cursor-pointer"
          >
            <option value="24h">Last 24 Hours</option>
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
          </select>
        </div>

        <div v-if="loadingMetrics" class="flex items-center justify-center py-8">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-kakoclaw-accent"></div>
        </div>

        <div v-else-if="agentMetrics" class="space-y-6">
          <!-- Overview Cards -->
          <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div class="p-4 bg-kakoclaw-bg/40 rounded-lg">
              <p class="text-[10px] font-bold uppercase tracking-wider text-kakoclaw-text-secondary mb-1">Total Cost</p>
              <p class="text-xl font-bold text-kakoclaw-text">{{ formatCost(agentMetrics.total_cost || 0) }}</p>
            </div>
            <div class="p-4 bg-kakoclaw-bg/40 rounded-lg">
              <p class="text-[10px] font-bold uppercase tracking-wider text-kakoclaw-text-secondary mb-1">Total Calls</p>
              <p class="text-xl font-bold text-kakoclaw-text">{{ agentMetrics.total_calls || 0 }}</p>
            </div>
            <div class="p-4 bg-kakoclaw-bg/40 rounded-lg">
              <p class="text-[10px] font-bold uppercase tracking-wider text-kakoclaw-text-secondary mb-1">Total Tokens</p>
              <p class="text-xl font-bold text-kakoclaw-text">{{ formatNumber(agentMetrics.total_tokens || 0) }}</p>
            </div>
            <div class="p-4 bg-kakoclaw-bg/40 rounded-lg">
              <p class="text-[10px] font-bold uppercase tracking-wider text-kakoclaw-text-secondary mb-1">Active Agents</p>
              <p class="text-xl font-bold text-kakoclaw-text">{{ agentMetrics.by_specialist ? Object.keys(agentMetrics.by_specialist).length : 0 }}</p>
            </div>
          </div>

          <!-- Cost by Specialist Table -->
          <div>
            <h4 class="font-bold text-kakoclaw-text mb-3">Cost by Specialist</h4>
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-kakoclaw-border">
                    <th class="text-left py-2 px-3 text-xs font-bold uppercase tracking-wider text-kakoclaw-text-secondary">Specialist</th>
                    <th class="text-right py-2 px-3 text-xs font-bold uppercase tracking-wider text-kakoclaw-text-secondary">Calls</th>
                    <th class="text-right py-2 px-3 text-xs font-bold uppercase tracking-wider text-kakoclaw-text-secondary">Tokens</th>
                    <th class="text-right py-2 px-3 text-xs font-bold uppercase tracking-wider text-kakoclaw-text-secondary">Avg Cost</th>
                    <th class="text-right py-2 px-3 text-xs font-bold uppercase tracking-wider text-kakoclaw-text-secondary">Total Cost</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="(metrics, name) in (agentMetrics.by_specialist || {})"
                    :key="name"
                    class="border-b border-kakoclaw-border/50 hover:bg-kakoclaw-bg/30"
                  >
                    <td class="py-3 px-3">
                      <div class="flex items-center gap-2">
                        <div :class="`w-6 h-6 rounded ${getSpecialistBgColor(name)} flex items-center justify-center`">
                          <svg class="w-3.5 h-3.5" :class="getSpecialistColor(name)" fill="none" stroke="currentColor" viewBox="0 0 24 24" v-html="getSpecialistIcon(name)"></svg>
                        </div>
                        <span class="font-medium text-kakoclaw-text capitalize">{{ name }}</span>
                      </div>
                    </td>
                    <td class="text-right py-3 px-3 text-kakoclaw-text">{{ metrics.calls || 0 }}</td>
                    <td class="text-right py-3 px-3 text-kakoclaw-text">{{ formatNumber(metrics.tokens || 0) }}</td>
                    <td class="text-right py-3 px-3 text-kakoclaw-text">{{ formatCost(metrics.avg_cost || 0) }}</td>
                    <td class="text-right py-3 px-3 font-bold text-kakoclaw-text">{{ formatCost(metrics.cost || 0) }}</td>
                  </tr>
                  <tr v-if="!agentMetrics.by_specialist || Object.keys(agentMetrics.by_specialist).length === 0">
                    <td colspan="5" class="py-8 text-center text-kakoclaw-text-secondary">No agent data available</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <div v-else class="p-8 text-center text-kakoclaw-text-secondary">
          <svg class="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p>No metrics available for the selected period</p>
        </div>
      </div>

      <!-- Email Report Section -->
      <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl shadow-sm p-6 space-y-6">
          <h3 class="font-bold text-kakoclaw-text flex items-center gap-2">
            <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
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
                      class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-4 py-2 outline-none focus:border-kakoclaw-accent transition-colors"
                  />
              </div>

              <div>
                  <label class="block text-sm font-medium mb-1">Subject</label>
                  <input
                      v-model="subject"
                      type="text"
                      placeholder="Weekly Summary / Project Update"
                      class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-4 py-2 outline-none focus:border-kakoclaw-accent transition-colors"
                  />
              </div>

              <div>
                  <label class="block text-sm font-medium mb-1">Content (Markdown)</label>
                  <textarea
                      v-model="body"
                      rows="10"
                      placeholder="Write your report here..."
                      class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-4 py-2 outline-none focus:border-kakoclaw-accent transition-colors resize-none font-mono text-sm"
                  ></textarea>
              </div>
          </div>

          <div class="flex justify-end pt-4 border-t border-kakoclaw-border">
              <button
                  @click="sendReport"
                  :disabled="sending || !subject || !body"
                  class="flex items-center gap-2 px-6 py-2.5 bg-kakoclaw-accent text-white rounded-lg hover:bg-kakoclaw-accent/90 transition-all font-medium disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-kakoclaw-accent/20"
              >
                  <div v-if="sending" class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                  <span v-else>Send Report</span>
                  <svg v-if="!sending" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" /></svg>
              </button>
          </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useChatStore } from '../stores/chatStore'
import { useAgentsStore } from '../stores/agentsStore'
import { ChatWebSocket } from '../services/websocketService'
import { useRouter } from 'vue-router'
import { useToast } from '../composables/useToast'

const chatStore = useChatStore()
const agentsStore = useAgentsStore()
const router = useRouter()
const toast = useToast()

const to = ref('')
const subject = ref('')
const body = ref('')
const sending = ref(false)
const metricsPeriod = ref('7d')
const loadingMetrics = ref(false)
const agentMetrics = ref(null)

onMounted(() => {
  fetchAgentMetrics()
})

async function fetchAgentMetrics() {
  loadingMetrics.value = true
  await agentsStore.fetchMetrics(metricsPeriod.value)
  agentMetrics.value = agentsStore.metrics
  loadingMetrics.value = false
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

const sendReport = async () => {
    if (!subject.value || !body.value) return

    sending.value = true

    // We send a command to agent to use the email tool
    const prompt = `Please send an email report using the 'send_email_report' tool.
Subject: ${subject.value}
To: ${to.value}
Body:
${body.value}`

    try {
        // Try using existing chatStore WebSocket first
        if (chatStore.sendMessage(prompt)) {
            toast.success('Report sent to agent')
            router.push('/chat')
            return
        }

        // If chatStore WS not connected, create a temporary one
        const tempWs = new ChatWebSocket()
        await tempWs.connect()
        tempWs.send({
            type: 'message',
            content: prompt,
            session_id: 'web:chat:report:' + Date.now().toString(36)
        })
        tempWs.disconnect()
        toast.success('Report sent to agent')
        router.push('/chat')
    } catch (err) {
        console.error("Failed to send report command:", err)
        toast.error('Failed to send command to agent. Make sure the chat is connected.')
    } finally {
        sending.value = false
    }
}
</script>
