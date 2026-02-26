<template>
  <div class="h-full flex flex-col bg-makoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-makoclaw-border flex flex-wrap gap-3 items-center justify-between bg-makoclaw-surface">
      <h2 class="text-xl font-bold bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent">Memory Management</h2>

      <!-- Tabs -->
      <div class="flex bg-makoclaw-bg rounded-lg p-1 border border-makoclaw-border">
        <button
          @click="activeTab = 'longterm'"
          class="tab-button"
          :class="[activeTab === 'longterm' ? 'tab-button-active' : 'tab-button-inactive']"
        >Long-Term</button>
        <button
          @click="activeTab = 'daily'"
          class="tab-button"
          :class="[activeTab === 'daily' ? 'tab-button-active' : 'tab-button-inactive']"
        >Daily Notes</button>
      </div>
    </div>

    <!-- ===== Long-Term Memory ===== -->
    <div v-if="activeTab === 'longterm'" class="flex-1 flex flex-col p-4 md:p-6 overflow-hidden gap-4">
      <!-- Toolbar -->
      <div class="flex flex-wrap items-center gap-3">
        <div class="flex-1 relative min-w-[200px]">
          <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-makoclaw-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            v-model="ltSearch"
            type="text"
            placeholder="Search in memory..."
            class="w-full pl-9 pr-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent text-makoclaw-text"
          >
        </div>
        <button
          @click="saveLongTerm"
          :disabled="saving"
          class="flex items-center gap-2 px-4 py-2 bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors disabled:opacity-50 text-sm"
        >
          <div v-if="saving" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" /></svg>
          Save Changes
        </button>
      </div>

      <!-- Search Highlight Info -->
      <div v-if="ltSearch" class="flex items-center gap-2 text-xs text-makoclaw-text-secondary">
        <svg class="w-3.5 h-3.5 text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span>{{ ltMatchCount }} occurrence{{ ltMatchCount !== 1 ? 's' : '' }} of "{{ ltSearch }}" found in the editor below</span>
      </div>

      <!-- Editor -->
      <div class="flex-1 relative border border-makoclaw-border rounded-xl overflow-hidden shadow-sm">
        <div v-if="loading" class="absolute inset-0 flex items-center justify-center bg-makoclaw-surface/50 backdrop-blur-sm z-sticky">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-makoclaw-accent"></div>
        </div>
        <textarea
          v-model="longTermContent"
          class="w-full h-full p-4 resize-none bg-makoclaw-surface outline-none font-mono text-sm leading-relaxed text-makoclaw-text"
          placeholder="Loading memory..."
          :class="{ 'border-2 border-yellow-400/40': ltSearch && ltMatchCount > 0 }"
        ></textarea>
      </div>
    </div>

    <!-- ===== Daily Notes ===== -->
    <div v-else class="flex-1 flex flex-col p-4 md:p-6 overflow-hidden gap-4">
      <!-- Toolbar -->
      <div class="flex flex-wrap items-center gap-3">
        <h3 class="font-semibold text-lg flex-1">Daily Notes</h3>
        <div class="relative">
          <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-makoclaw-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            v-model="dailySearch"
            type="text"
            placeholder="Search notes..."
            class="pl-9 pr-3 py-1.5 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent text-makoclaw-text w-44"
          >
        </div>
        <select v-model="days" @change="loadDaily" class="bg-makoclaw-surface border border-makoclaw-border rounded-lg px-3 py-1.5 text-sm outline-none focus:border-makoclaw-accent text-makoclaw-text">
          <option :value="3">Last 3 Days</option>
          <option :value="7">Last 7 Days</option>
          <option :value="14">Last 14 Days</option>
          <option :value="30">Last 30 Days</option>
        </select>
      </div>

      <!-- Timeline -->
      <div class="flex-1 overflow-auto custom-scrollbar">
        <div v-if="loadingDaily" class="flex items-center justify-center h-40">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-makoclaw-accent"></div>
        </div>

        <div v-else-if="filteredDailySections.length === 0" class="flex flex-col items-center justify-center h-40 text-makoclaw-text-secondary">
          <svg class="w-10 h-10 mb-3 opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
          <p class="text-sm">{{ dailySearch ? 'No matches found' : 'No daily notes available' }}</p>
        </div>

        <!-- Timeline entries -->
        <div v-else class="relative pl-8 space-y-6">
          <!-- Vertical timeline bar -->
          <div class="absolute left-3.5 top-3 bottom-3 w-0.5 bg-makoclaw-border rounded-full"></div>

          <div v-for="section in filteredDailySections" :key="section.date" class="relative">
            <!-- Timeline dot -->
            <div class="absolute -left-[22px] top-1.5 w-3.5 h-3.5 rounded-full border-2 border-makoclaw-accent bg-makoclaw-bg shadow-sm shadow-makoclaw-accent/30 hover:shadow-makoclaw-accent/30 hover:shadow-md transition-shadow duration-200"></div>

            <!-- Note card -->
            <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl overflow-hidden hover:border-makoclaw-accent/30 transition-colors">
              <!-- Card header -->
              <div class="flex items-center justify-between px-4 py-3 bg-makoclaw-bg border-b border-makoclaw-border">
                <div class="flex items-center gap-2">
                  <svg class="w-4 h-4 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
                  <span class="text-sm font-semibold text-makoclaw-text">{{ section.label }}</span>
                </div>
                <button @click="copyNote(section.content)" class="p-1.5 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-colors rounded" title="Copy to clipboard">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
                </button>
              </div>

              <!-- Card body -->
              <div class="px-4 py-3">
                <pre class="whitespace-pre-wrap font-mono text-xs leading-relaxed text-makoclaw-text" v-html="highlightText(section.content, dailySearch)"></pre>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import memoryService from '../services/memoryService'
import { useToast } from '../composables/useToast'

const toast = useToast()

const activeTab = ref('longterm')
const longTermContent = ref('')
const dailyContent = ref('')
const loading = ref(false)
const loadingDaily = ref(false)
const saving = ref(false)
const days = ref(7)

// Search state
const ltSearch = ref('')
const dailySearch = ref('')

// ---- Long-term search computed ----
const ltMatchCount = computed(() => {
  if (!ltSearch.value || !longTermContent.value) return 0
  const regex = new RegExp(ltSearch.value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi')
  return (longTermContent.value.match(regex) || []).length
})

// ---- Daily notes parsing into sections (one per date heading) ----
const dailySections = computed(() => {
  if (!dailyContent.value) return []
  const lines = dailyContent.value.split('\n')
  const sections = []
  let current = null

  for (const line of lines) {
    // Match headings like "# 2026-02-20" or "## 2026-02-20"
    const headingMatch = line.match(/^#{1,3}\s*(\d{4}-\d{2}-\d{2})/)
    if (headingMatch) {
      if (current) sections.push(current)
      const dateStr = headingMatch[1]
      const d = new Date(dateStr + 'T00:00:00')
      const today = new Date()
      const diffDays = Math.floor((today - d) / (1000 * 60 * 60 * 24))
      let label = dateStr
      if (diffDays === 0) label = 'Today — ' + dateStr
      else if (diffDays === 1) label = 'Yesterday — ' + dateStr
      else label = diffDays + ' days ago — ' + dateStr

      current = { date: dateStr, label, content: '' }
    } else if (current) {
      current.content += line + '\n'
    } else {
      // Pre-heading content
      if (!sections.length) sections.push({ date: 'general', label: 'General Notes', content: line + '\n' })
      else sections[0].content += line + '\n'
    }
  }
  if (current) sections.push(current)
  return sections.reverse() // newest first
})

const filteredDailySections = computed(() => {
  if (!dailySearch.value) return dailySections.value
  const q = dailySearch.value.toLowerCase()
  return dailySections.value.filter(s =>
    s.content.toLowerCase().includes(q) || s.label.toLowerCase().includes(q)
  )
})

// Highlight search term in text
const highlightText = (text, query) => {
  if (!query) return escapeHtml(text)
  const escaped = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const regex = new RegExp(`(${escaped})`, 'gi')
  return escapeHtml(text).replace(regex, '<mark class="bg-yellow-400/30 text-yellow-200 rounded px-0.5">$1</mark>')
}

const escapeHtml = (str) =>
  str.replace(/&/g, '&amp;')
     .replace(/</g, '&lt;')
     .replace(/>/g, '&gt;')

const copyNote = async (content) => {
  try {
    await navigator.clipboard.writeText(content)
    toast.success('Copied to clipboard')
  } catch {
    toast.error('Failed to copy')
  }
}

// ---- API calls ----
const loadLongTerm = async () => {
  loading.value = true
  try {
    const res = await memoryService.getLongTermMemory()
    longTermContent.value = res.data.content
  } catch (err) {
    console.error('Failed to load memory:', err)
    toast.error('Failed to load memory')
  } finally {
    loading.value = false
  }
}

const saveLongTerm = async () => {
  saving.value = true
  try {
    await memoryService.updateLongTermMemory(longTermContent.value)
    toast.success('Memory saved successfully')
  } catch (err) {
    console.error('Failed to save memory:', err)
    toast.error('Failed to save memory')
  } finally {
    saving.value = false
  }
}

const loadDaily = async () => {
  loadingDaily.value = true
  try {
    const res = await memoryService.getDailyNotes(days.value)
    dailyContent.value = res.data.content
  } catch (err) {
    console.error('Failed to load daily notes:', err)
    toast.error('Failed to load daily notes')
  } finally {
    loadingDaily.value = false
  }
}

onMounted(() => {
  loadLongTerm()
  loadDaily()
})
</script>

