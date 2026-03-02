<template>
  <div class="h-full flex flex-col bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-violet-500/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-fuchsia-500/20 via-transparent to-transparent" />
    </div>

    <!-- Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
        <div class="flex items-center gap-3">
          <!-- Icon Container -->
          <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-violet-500/20 to-fuchsia-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-violet-500/10">
            <svg
              class="w-5 h-5 sm:w-6 sm:h-6 text-violet-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"
              />
            </svg>
          </div>

          <div class="flex-1 min-w-0">
            <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-violet-400 bg-clip-text text-transparent">
              Memory Management
            </h1>
            <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5 hidden sm:block">
              Persistent context and daily observations
            </p>
          </div>

          <!-- Tabs -->
          <div class="flex bg-makoclaw-bg/50 rounded-xl p-1 border border-makoclaw-border/50 flex-shrink-0">
            <button
              class="px-4 py-2 rounded-lg text-sm font-bold transition-all"
              :class="[activeTab === 'longterm' ? 'bg-gradient-to-r from-violet-500/20 to-fuchsia-500/20 text-violet-400 shadow-sm ring-1 ring-white/10' : 'text-makoclaw-text-secondary hover:text-makoclaw-text']"
              @click="activeTab = 'longterm'"
            >
              Long-Term
            </button>
            <button
              class="px-4 py-2 rounded-lg text-sm font-bold transition-all"
              :class="[activeTab === 'daily' ? 'bg-gradient-to-r from-violet-500/20 to-fuchsia-500/20 text-violet-400 shadow-sm ring-1 ring-white/10' : 'text-makoclaw-text-secondary hover:text-makoclaw-text']"
              @click="activeTab = 'daily'"
            >
              Daily Notes
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ===== Long-Term Memory ===== -->
    <div
      v-if="activeTab === 'longterm'"
      class="flex-1 flex flex-col p-4 sm:p-6 gap-4 overflow-hidden relative z-10"
    >
      <!-- Long-term Memory Section -->
      <section class="glass-panel p-6 rounded-2xl shadow-xl shadow-violet-500/5 relative overflow-hidden group/editor flex-1 flex flex-col min-h-0">
        <div class="flex justify-between items-center mb-4">
          <div class="flex items-center gap-3">
            <div class="p-2 bg-violet-500/10 rounded-xl text-violet-400">
              <svg
                class="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
              /></svg>
            </div>
            <div>
              <h3 class="text-lg font-bold text-makoclaw-text">
                Long-term Memory
              </h3>
              <p class="text-xs text-makoclaw-text-secondary opacity-70">
                Core knowledge and persistent context
              </p>
            </div>
          </div>
          <button
            v-if="saving"
            class="px-4 py-2 bg-gradient-to-r from-violet-500 to-fuchsia-500 text-white text-sm font-semibold rounded-xl hover:shadow-lg hover:shadow-violet-500/20 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0"
            @click="saveLongTerm"
          >
            Save Changes
          </button>
        </div>

        <div class="relative group/textarea flex-1 flex flex-col min-h-0">
          <!-- Search match indicator -->
          <div
            v-if="ltSearch && ltMatchCount > 0"
            class="absolute top-2 right-2 px-2 py-1 bg-violet-500/20 border border-violet-500/30 rounded text-[10px] text-violet-400 font-mono z-10 animate-pulse"
          >
            {{ ltMatchCount }} matches
          </div>
          <textarea
            v-model="longTermContent"
            class="w-full flex-1 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl p-4 text-sm text-makoclaw-text outline-none focus:border-violet-400/50 focus:ring-2 focus:ring-violet-400/20 transition-all duration-300 resize-none font-mono custom-scrollbar"
            :class="{ 'border-violet-400/40': ltSearch && ltMatchCount > 0 }"
            placeholder="No long-term memory records found..."
          />
          <!-- Focus accent line -->
          <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-violet-500 to-fuchsia-500 transition-all duration-500 opacity-50 focus-within:w-full" />
        </div>
      </section>

      <!-- Search feedback for Long-term memory -->
      <div
        v-if="ltSearch"
        class="px-2 flex items-center gap-2 text-[10px] text-makoclaw-text-secondary opacity-60"
      >
        <svg
          class="w-3 h-3 text-violet-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span>Searching for "{{ ltSearch }}" in core knowledge</span>
      </div>
    </div>

    <!-- ===== Daily Notes ===== -->
    <div
      v-else
      class="flex-1 flex flex-col p-4 md:p-6 overflow-hidden gap-4 relative z-10"
    >
      <!-- Toolbar -->
      <div class="flex flex-wrap items-center gap-3">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-violet-500/10 rounded-xl text-violet-400">
            <svg
              class="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
            /></svg>
          </div>
          <div>
            <h3 class="text-lg font-bold text-makoclaw-text flex items-center gap-2">
              Daily Notes
              <div class="w-2 h-2 bg-violet-500 rounded-full animate-pulse shadow-[0_0_8px_rgba(139,92,246,0.5)]" />
            </h3>
            <p class="text-xs text-makoclaw-text-secondary opacity-70">
              Timeline of automatic summaries and observations
            </p>
          </div>
        </div>
        <div class="relative">
          <svg
            class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-makoclaw-text-secondary"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
          <input
            v-model="dailySearch"
            type="text"
            placeholder="Search notes..."
            class="pl-9 pr-3 py-1.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm outline-none focus:border-violet-400/50 text-makoclaw-text w-44"
          >
        </div>
        <select
          v-model="days"
          class="bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl px-3 py-1.5 text-sm outline-none focus:border-violet-400/50 text-makoclaw-text cursor-pointer"
          @change="loadDaily"
        >
          <option :value="3">
            Last 3 Days
          </option>
          <option :value="7">
            Last 7 Days
          </option>
          <option :value="14">
            Last 14 Days
          </option>
          <option :value="30">
            Last 30 Days
          </option>
        </select>
      </div>

      <!-- Timeline -->
      <div class="flex-1 overflow-auto custom-scrollbar">
        <div
          v-if="loadingDaily"
          class="flex items-center justify-center h-40"
        >
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-makoclaw-accent" />
        </div>

        <div
          v-else-if="filteredDailySections.length === 0"
          class="flex flex-col items-center justify-center h-40 text-makoclaw-text-secondary"
        >
          <svg
            class="w-10 h-10 mb-3 opacity-30"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          /></svg>
          <p class="text-sm">
            {{ dailySearch ? 'No matches found' : 'No daily notes available' }}
          </p>
        </div>

        <!-- Timeline entries -->
        <div
          v-else
          class="relative pl-8 space-y-6"
        >
          <div
            v-for="section in filteredDailySections"
            :key="section.date"
            class="flex"
          >
            <div class="flex-none mr-4 pt-1 flex flex-col items-center">
              <div class="w-2.5 h-2.5 bg-makoclaw-accent rounded-full shadow-[0_0_10px_rgba(var(--makoclaw-accent-rgb),0.4)] animate-pulse" />
              <div class="w-px flex-1 bg-makoclaw-border/50 my-2" />
            </div>

            <div class="flex-1 pb-8 group/note">
              <!-- Note card -->
              <div class="glass-panel p-4 rounded-xl relative overflow-hidden transition-all duration-300 hover:border-violet-400/30 group-hover/note:-translate-y-1 group-hover/note:shadow-lg group-hover/note:shadow-violet-500/5">
                <!-- Hover glow line -->
                <div class="absolute left-0 top-0 bottom-0 w-1 bg-violet-500/0 group-hover/note:bg-violet-500/40 transition-all duration-300" />
                <!-- Animated bottom-line -->
                <div class="absolute bottom-0 left-0 h-[1.5px] w-0 bg-gradient-to-r from-violet-500 to-fuchsia-500 group-hover/note:w-full transition-all duration-500 opacity-40" />
                
                <div class="flex justify-between items-start mb-3">
                  <div class="flex items-center gap-2">
                    <div class="p-1.5 bg-violet-500/5 rounded-lg text-violet-400 opacity-60 group-hover/note:opacity-100 transition-all duration-300">
                      <svg
                        class="w-3.5 h-3.5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                      /></svg>
                    </div>
                    <h4 class="font-bold text-sm text-violet-400 tracking-wide">
                      {{ section.label }}
                    </h4>
                  </div>
                  <div class="flex items-center gap-1">
                    <button
                      class="opacity-0 group-hover/note:opacity-100 p-1.5 hover:bg-violet-500/10 text-makoclaw-text-secondary hover:text-violet-400 rounded-lg transition-all duration-300"
                      title="Copy to clipboard"
                      @click="copyNote(section.content)"
                    >
                      <svg
                        class="w-3.5 h-3.5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                      /></svg>
                    </button>
                    <button
                      class="opacity-0 group-hover/note:opacity-100 p-1.5 hover:bg-makoclaw-error/10 text-makoclaw-text-secondary hover:text-makoclaw-error rounded-lg transition-all duration-300"
                      title="Delete note"
                      @click="deleteNote(section.date)"
                    >
                      <svg
                        class="w-3.5 h-3.5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                      /></svg>
                    </button>
                  </div>
                </div>
                <!-- Content area with glass background -->
                <div class="relative group/content">
                  <div
                    v-if="dailySearch"
                    class="w-full bg-makoclaw-bg/30 border border-makoclaw-border/50 rounded-lg p-3 text-sm text-makoclaw-text outline-none min-h-[100px] font-mono leading-relaxed"
                    v-html="highlightText(section.content, dailySearch)"
                  />
                  <textarea
                    v-else
                    v-model="section.content"
                    class="w-full bg-makoclaw-bg/30 border border-makoclaw-border/50 rounded-lg p-3 text-sm text-makoclaw-text outline-none focus:border-violet-400/50 transition-all duration-300 resize-none min-h-[100px]"
                    @change="updateNote(section)"
                  />
                  <!-- Focus indicator -->
                  <div class="absolute bottom-0 left-0 h-px w-0 bg-violet-500 transition-all duration-500 focus-within:w-full" />
                </div>
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

const deleteNote = async (date) => {
  // Placeholder for delete functionality
  toast.info(`Delete note for ${date} (not implemented)`)
}

const updateNote = async (note) => {
  // Placeholder for update functionality
  toast.info(`Update note for ${note.date} (not implemented)`)
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

<style scoped>
.glass-sticky {
  background: rgba(var(--makoclaw-surface-rgb), 0.7);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}
</style>
