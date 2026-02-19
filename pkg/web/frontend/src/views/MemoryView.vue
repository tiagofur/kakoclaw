<template>
  <div class="h-full flex flex-col bg-kakoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-kakoclaw-border flex justify-between items-center bg-kakoclaw-surface">
      <h2 class="text-xl font-bold bg-gradient-to-r from-kakoclaw-accent to-purple-500 bg-clip-text text-transparent">Memory Management</h2>
      
      <!-- Tabs -->
      <div class="flex bg-kakoclaw-bg rounded-lg p-1 border border-kakoclaw-border">
          <button 
            @click="activeTab = 'longterm'"
            class="px-4 py-1.5 rounded-md text-sm font-medium transition-all"
            :class="activeTab === 'longterm' ? 'bg-white dark:bg-gray-700 shadow-sm text-kakoclaw-accent' : 'text-kakoclaw-text-secondary hover:text-kakoclaw-text'"
          >
            Long-Term
          </button>
          <button 
            @click="activeTab = 'daily'"
            class="px-4 py-1.5 rounded-md text-sm font-medium transition-all"
            :class="activeTab === 'daily' ? 'bg-white dark:bg-gray-700 shadow-sm text-kakoclaw-accent' : 'text-kakoclaw-text-secondary hover:text-kakoclaw-text'"
          >
            Daily Notes
          </button>
      </div>
    </div>

    <!-- Long Term Memory Tab -->
    <div v-if="activeTab === 'longterm'" class="flex-1 flex flex-col p-6 overflow-hidden">
        <div class="flex items-center justify-between mb-4">
            <div>
                <h3 class="font-semibold text-lg">Core Memory (MEMORY.md)</h3>
                <p class="text-sm text-kakoclaw-text-secondary">Persistent facts and preferences the agent remembers about you.</p>
            </div>
            <button 
                @click="saveLongTerm" 
                :disabled="saving"
                class="flex items-center gap-2 px-4 py-2 bg-kakoclaw-accent text-white rounded-lg hover:bg-kakoclaw-accent/90 transition-colors disabled:opacity-50"
            >
                <div v-if="saving" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                <span v-else>Save Changes</span>
            </button>
        </div>
        
        <div class="flex-1 relative border border-kakoclaw-border rounded-xl overflow-hidden shadow-sm">
             <div v-if="loading" class="absolute inset-0 flex items-center justify-center bg-kakoclaw-surface/50 backdrop-blur-sm z-10">
                 <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-kakoclaw-accent"></div>
             </div>
             <textarea 
                v-model="longTermContent" 
                class="w-full h-full p-4 resize-none bg-kakoclaw-surface outline-none font-mono text-sm leading-relaxed"
                placeholder="Loading memory..."
             ></textarea>
        </div>
    </div>

    <!-- Daily Notes Tab -->
    <div v-else class="flex-1 flex flex-col p-6 overflow-hidden">
        <div class="flex items-center justify-between mb-4">
            <div>
                <h3 class="font-semibold text-lg">Daily Notes</h3>
                <p class="text-sm text-kakoclaw-text-secondary">Auto-generated summaries and activities from recent days.</p>
            </div>
            <select v-model="days" @change="loadDaily" class="bg-kakoclaw-surface border border-kakoclaw-border rounded-lg px-3 py-1.5 text-sm outline-none focus:border-kakoclaw-accent">
                <option :value="3">Last 3 Days</option>
                <option :value="7">Last 7 Days</option>
                <option :value="14">Last 14 Days</option>
                <option :value="30">Last 30 Days</option>
            </select>
        </div>

        <div class="flex-1 border border-kakoclaw-border rounded-xl bg-kakoclaw-surface overflow-auto relative p-4 custom-scrollbar">
             <div v-if="loadingDaily" class="absolute inset-0 flex items-center justify-center bg-kakoclaw-surface/50 backdrop-blur-sm z-10">
                 <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-kakoclaw-accent"></div>
             </div>
             
             <!-- Markdown Content Display -->
             <div class="prose dark:prose-invert max-w-none">
                 <pre class="whitespace-pre-wrap font-mono text-sm">{{ dailyContent }}</pre>
             </div>
        </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
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

const loadLongTerm = async () => {
    loading.value = true
    try {
        const res = await memoryService.getLongTermMemory()
        longTermContent.value = res.data.content
    } catch (err) {
        console.error("Failed to load memory:", err)
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
        console.error("Failed to save memory:", err)
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
        console.error("Failed to load daily notes:", err)
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
.custom-scrollbar::-webkit-scrollbar {
  width: 8px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 4px;
}
</style>
