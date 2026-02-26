<template>
  <div class="h-full flex flex-col bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh (Subtle) -->
    <div class="absolute inset-0 pointer-events-none opacity-20 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-makoclaw-accent/30 via-transparent to-transparent"></div>

    <!-- Header -->
    <div class="flex-none p-4 border-b border-makoclaw-border/30 bg-makoclaw-surface/30 backdrop-blur-sm z-20">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-xl font-bold bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent">Multi-Agent System</h2>
          <p class="text-sm text-makoclaw-text-secondary mt-1">Manage your AI specialists and track their performance</p>
        </div>
        <div class="flex items-center gap-3">
          <select
            v-model="metricsPeriod"
            @change="fetchMetrics"
            class="px-3 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-lg text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent outline-none cursor-pointer backdrop-blur-sm"
          >
            <option value="24h">Last 24 Hours</option>
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-4 md:p-6 custom-scrollbar">
      <!-- Loading Skeleton -->
      <div v-if="agentsStore.loading" class="space-y-6">
        <!-- Orchestrator Skeleton -->
        <div class="glass-panel rounded-2xl p-6 border border-makoclaw-border/50">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-3">
              <div class="skeleton w-12 h-12 rounded-xl"></div>
              <div>
                <div class="skeleton h-4 w-28 mb-2 rounded"></div>
                <div class="skeleton h-3 w-48 rounded"></div>
              </div>
            </div>
            <div class="skeleton h-6 w-16 rounded-full"></div>
          </div>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
            <div v-for="i in 4" :key="'orch-'+i" class="p-3 bg-makoclaw-bg/40 rounded-lg">
              <div class="skeleton h-2.5 w-16 mb-2 rounded"></div>
              <div class="skeleton h-4 w-24 rounded"></div>
            </div>
          </div>
        </div>
        <!-- Specialists Skeleton -->
        <div class="space-y-4">
          <div class="skeleton h-5 w-32 rounded"></div>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div v-for="i in 3" :key="'spec-'+i" class="glass-panel rounded-xl p-5 border border-makoclaw-border/50">
              <div class="flex items-start gap-3 mb-3">
                <div class="skeleton w-10 h-10 rounded-lg"></div>
                <div>
                  <div class="skeleton h-4 w-28 mb-2 rounded"></div>
                  <div class="skeleton h-2.5 w-36 rounded"></div>
                </div>
              </div>
              <div class="skeleton h-3 w-full mb-1 rounded"></div>
              <div class="skeleton h-3 w-2/3 mb-4 rounded"></div>
              <div class="flex gap-1 mb-4">
                <div class="skeleton h-4 w-14 rounded-full"></div>
                <div class="skeleton h-4 w-16 rounded-full"></div>
                <div class="skeleton h-4 w-12 rounded-full"></div>
              </div>
              <div class="pt-3 border-t border-makoclaw-border/50 flex items-center justify-between">
                <div class="skeleton h-3 w-20 rounded"></div>
                <div class="skeleton h-3 w-12 rounded"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="space-y-6">
        <!-- Orchestrator Status -->
        <div class="glass-panel rounded-2xl p-6 border" :class="orchestrator.enabled ? 'border-makoclaw-accent/30 bg-makoclaw-accent/5' : 'border-makoclaw-border/50'">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-3">
              <div :class="`w-12 h-12 rounded-xl flex items-center justify-center ${orchestrator.enabled ? 'bg-makoclaw-accent/20' : 'bg-makoclaw-bg/60'}`">
                <svg class="w-6 h-6" :class="orchestrator.enabled ? 'text-makoclaw-accent' : 'text-makoclaw-text-secondary'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
                </svg>
              </div>
              <div>
                <h3 class="font-bold text-makoclaw-text">Orchestrator</h3>
                <p class="text-sm text-makoclaw-text-secondary">{{ orchestrator.enabled ? 'Active - Delegating tasks to specialists' : 'Disabled' }}</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <span :class="`px-3 py-1 text-xs font-bold rounded-full ${orchestrator.enabled ? 'bg-makoclaw-success/20 text-makoclaw-success' : 'bg-makoclaw-text-secondary/20 text-makoclaw-text-secondary'}`">
                {{ orchestrator.enabled ? 'Active' : 'Inactive' }}
              </span>
            </div>
          </div>

          <div v-if="orchestrator.enabled" class="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
            <div class="p-3 bg-makoclaw-bg/40 rounded-lg">
              <p class="text-xs text-makoclaw-text-secondary uppercase tracking-wider mb-1">Provider</p>
              <p class="font-bold text-makoclaw-text">{{ orchestrator.provider }}</p>
            </div>
            <div class="p-3 bg-makoclaw-bg/40 rounded-lg">
              <p class="text-xs text-makoclaw-text-secondary uppercase tracking-wider mb-1">Model</p>
              <p class="font-bold text-makoclaw-text text-sm">{{ orchestrator.model }}</p>
            </div>
            <div class="p-3 bg-makoclaw-bg/40 rounded-lg">
              <p class="text-xs text-makoclaw-text-secondary uppercase tracking-wider mb-1">Temperature</p>
              <p class="font-bold text-makoclaw-text">{{ orchestrator.temperature }}</p>
            </div>
            <div class="p-3 bg-makoclaw-bg/40 rounded-lg">
              <p class="text-xs text-makoclaw-text-secondary uppercase tracking-wider mb-1">Max Retries</p>
              <p class="font-bold text-makoclaw-text">{{ orchestrator.max_delegation_retries }}</p>
            </div>
          </div>

          <div v-else class="mt-4 p-4 bg-makoclaw-bg/40 rounded-lg text-center">
            <p class="text-sm text-makoclaw-text-secondary mb-3">Orchestrator is not enabled. Configure it in Settings > Agents.</p>
            <router-link
              to="/settings?tab=agents"
              class="inline-flex items-center gap-2 px-4 py-2 bg-makoclaw-accent/10 hover:bg-makoclaw-accent/20 text-makoclaw-accent rounded-lg text-sm font-medium transition-colors"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              </svg>
              Configure Orchestrator
            </router-link>
          </div>
        </div>

        <!-- Specialists Grid -->
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="font-bold text-makoclaw-text flex items-center gap-2">
              <svg class="w-5 h-5 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
              Specialists
              <span class="px-2 py-0.5 text-xs font-bold bg-makoclaw-accent/20 text-makoclaw-accent rounded-full">{{ agentsStore.specialists.length }}</span>
            </h3>
          </div>

          <div v-if="agentsStore.specialists.length === 0" class="p-12 bg-makoclaw-bg/40 rounded-2xl border border-dashed border-makoclaw-border text-center">
            <svg class="w-16 h-16 mx-auto text-makoclaw-text-secondary/30 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
            <h4 class="text-lg font-bold text-makoclaw-text mb-2">No Specialists Yet</h4>
            <p class="text-sm text-makoclaw-text-secondary mb-6 max-w-md mx-auto">Create your first AI specialist to start using the multi-agent system. Each specialist can be configured with specific tools and capabilities.</p>
            <router-link
              to="/settings?tab=agents"
              class="inline-flex items-center gap-2 px-6 py-3 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl font-bold shadow-lg shadow-makoclaw-accent/20 transition-all"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
              </svg>
              Create First Specialist
            </router-link>
          </div>

          <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div
              v-for="specialist in agentsStore.specialists"
              :key="specialist.name"
              class="glass-panel rounded-xl p-5 border border-makoclaw-border/50 hover:border-makoclaw-accent/50 hover:shadow-xl hover:shadow-makoclaw-accent/10 transition-all group cursor-pointer"
              @click="openSpecialistDetails(specialist)"
            >
              <div class="flex items-start justify-between mb-3">
                <div class="flex items-center gap-3">
                  <div :class="`w-10 h-10 rounded-lg ${agentsStore.getSpecialistBgColor(specialist.name)} flex items-center justify-center`">
                    <svg class="w-5 h-5" :class="agentsStore.getSpecialistColor(specialist.name)" fill="none" stroke="currentColor" viewBox="0 0 24 24" v-html="agentsStore.getSpecialistIcon(specialist.name)"></svg>
                  </div>
                  <div>
                    <h4 class="font-bold text-makoclaw-text capitalize">{{ specialist.name }}</h4>
                    <p class="text-xs text-makoclaw-text-secondary">{{ specialist.provider }} / {{ specialist.model }}</p>
                  </div>
                </div>
              </div>

              <p class="text-sm text-makoclaw-text-secondary mb-4 line-clamp-2">{{ specialist.description }}</p>

              <div class="flex flex-wrap gap-1 mb-4">
                <span
                  v-for="keyword in (specialist.keywords || []).slice(0, 3)"
                  :key="keyword"
                  class="px-2 py-0.5 text-[10px] font-medium bg-makoclaw-bg/60 text-makoclaw-text-secondary rounded-full"
                >
                  {{ keyword }}
                </span>
              </div>

              <div class="pt-3 border-t border-makoclaw-border/50 flex items-center justify-between text-xs">
                <div class="flex items-center gap-4 text-makoclaw-text-secondary">
                  <span class="flex items-center gap-1">
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    {{ (specialist.tools || []).length }} tools
                  </span>
                  <span class="flex items-center gap-1">
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                    </svg>
                    {{ specialist.temperature }}
                  </span>
                </div>
                <span class="text-makoclaw-accent font-medium group-hover:translate-x-1 transition-transform">View →</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Cost Overview -->
        <div v-if="metrics" class="space-y-4">
          <h3 class="font-bold text-makoclaw-text flex items-center gap-2">
            <svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Cost Overview
            <span class="text-xs text-makoclaw-text-secondary font-normal uppercase tracking-wider">({{ metricsPeriod }})</span>
          </h3>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div class="glass-panel rounded-xl p-5 border border-makoclaw-border/50">
              <div class="flex items-center gap-3 mb-3">
                <div class="w-10 h-10 rounded-lg bg-makoclaw-accent/20 flex items-center justify-center">
                  <svg class="w-5 h-5 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <div>
                  <p class="text-xs text-makoclaw-text-secondary uppercase tracking-wider">Total Cost</p>
                  <p class="text-xl font-bold text-makoclaw-text">{{ formatCost(metrics.total_cost || 0) }}</p>
                </div>
              </div>
            </div>

            <div class="glass-panel rounded-xl p-5 border border-makoclaw-border/50">
              <div class="flex items-center gap-3 mb-3">
                <div class="w-10 h-10 rounded-lg bg-makoclaw-accent/20 flex items-center justify-center">
                  <svg class="w-5 h-5 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
                  </svg>
                </div>
                <div>
                  <p class="text-xs text-makoclaw-text-secondary uppercase tracking-wider">Total Calls</p>
                  <p class="text-xl font-bold text-makoclaw-text">{{ metrics.total_calls || 0 }}</p>
                </div>
              </div>
            </div>

            <div class="glass-panel rounded-xl p-5 border border-makoclaw-border/50">
              <div class="flex items-center gap-3 mb-3">
                <div class="w-10 h-10 rounded-lg bg-makoclaw-accent/20 flex items-center justify-center">
                  <svg class="w-5 h-5 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
                  </svg>
                </div>
                <div>
                  <p class="text-xs text-makoclaw-text-secondary uppercase tracking-wider">Total Tokens</p>
                  <p class="text-xl font-bold text-makoclaw-text">{{ formatNumber(metrics.total_tokens || 0) }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Cost by Specialist -->
          <div class="glass-panel rounded-xl p-5 border border-makoclaw-border/50">
            <h4 class="font-bold text-makoclaw-text mb-4">Cost by Specialist</h4>
            <div class="space-y-3">
              <div
                v-for="(cost, name) in (metrics.by_specialist || {})"
                :key="name"
                class="flex items-center justify-between p-3 bg-makoclaw-bg/40 rounded-lg"
              >
                <div class="flex items-center gap-3">
                  <div :class="`w-8 h-8 rounded-lg ${agentsStore.getSpecialistBgColor(name)} flex items-center justify-center`">
                    <svg class="w-4 h-4" :class="agentsStore.getSpecialistColor(name)" fill="none" stroke="currentColor" viewBox="0 0 24 24" v-html="agentsStore.getSpecialistIcon(name)"></svg>
                  </div>
                  <span class="font-medium text-makoclaw-text capitalize">{{ name }}</span>
                </div>
                <div class="flex items-center gap-4">
                  <span class="text-sm text-makoclaw-text-secondary">{{ cost.calls || 0 }} calls</span>
                  <span class="font-bold text-makoclaw-text">{{ formatCost(cost.cost || 0) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Specialist Details Modal -->
    <SpecialistDetailsModal
      v-if="selectedSpecialist"
      :show="showSpecialistDetails"
      :specialist="selectedSpecialist"
      @close="closeSpecialistDetails"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAgentsStore } from '../stores/agentsStore'
import SpecialistDetailsModal from '../components/Agents/SpecialistDetailsModal.vue'

const agentsStore = useAgentsStore()
const metricsPeriod = ref('7d')
const showSpecialistDetails = ref(false)
const selectedSpecialist = ref(null)

const orchestrator = computed(() => agentsStore.orchestrator || { enabled: false })
const metrics = computed(() => agentsStore.metrics)

onMounted(async () => {
  await agentsStore.fetchAgents()
  await agentsStore.fetchMetrics(metricsPeriod.value)
})

function fetchMetrics() {
  agentsStore.fetchMetrics(metricsPeriod.value)
}

function formatCost(cost) {
  if (!cost) return '$0.00'
  return '$' + cost.toFixed(4)
}

function formatNumber(num) {
  if (!num) return '0'
  return num.toLocaleString()
}

function openSpecialistDetails(specialist) {
  selectedSpecialist.value = specialist
  showSpecialistDetails.value = true
}

function closeSpecialistDetails() {
  showSpecialistDetails.value = false
  selectedSpecialist.value = null
}
</script>


