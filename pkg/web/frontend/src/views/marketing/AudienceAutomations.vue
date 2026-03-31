<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h3 class="text-lg font-semibold text-white">Automations</h3>
      <button
        class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl font-medium"
        @click="showForm = true; form = { name: '', trigger_type: 'list_subscribe', trigger_list_id: '', delay_hours: 0, campaign_id: '', status: 'active' }"
      >+ New Automation</button>
    </div>

    <!-- Create Form -->
    <div v-if="showForm" class="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-6 space-y-3">
      <input v-model="form.name" placeholder="Automation Name *" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
      <div class="grid grid-cols-2 gap-3">
        <div class="space-y-1">
          <label class="text-xs text-white/50 font-medium">Trigger</label>
          <select v-model="form.trigger_type" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white text-sm focus:border-pink-500/50 focus:outline-none">
            <option value="list_subscribe">When contact joins a list</option>
          </select>
        </div>
        <div class="space-y-1">
          <label class="text-xs text-white/50 font-medium">List *</label>
          <select v-model="form.trigger_list_id" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white text-sm focus:border-pink-500/50 focus:outline-none">
            <option value="">Select list...</option>
            <option v-for="list in store.audienceLists" :key="list.id" :value="list.id">{{ list.name }}</option>
          </select>
        </div>
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div class="space-y-1">
          <label class="text-xs text-white/50 font-medium">Send Email Campaign *</label>
          <select v-model="form.campaign_id" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white text-sm focus:border-pink-500/50 focus:outline-none">
            <option value="">Select campaign...</option>
            <option v-for="ec in store.emailCampaigns" :key="ec.id" :value="ec.id">{{ ec.name }}</option>
          </select>
        </div>
        <div class="space-y-1">
          <label class="text-xs text-white/50 font-medium">Delay (hours)</label>
          <input v-model.number="form.delay_hours" type="number" min="0" placeholder="0 = immediate" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
        </div>
      </div>
      <div class="flex gap-2 justify-end">
        <button class="px-4 py-2 text-sm text-white/60 hover:text-white" @click="showForm = false">Cancel</button>
        <button class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 text-white rounded-xl font-medium" @click="saveAutomation">Save</button>
      </div>
    </div>

    <!-- Automations Table -->
    <div v-if="store.automationsLoading" class="flex items-center justify-center py-8">
      <svg class="animate-spin w-6 h-6 text-pink-400" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/></svg>
    </div>
    <div v-else-if="store.automations.length === 0" class="text-center py-8 text-white/40 text-sm">No automations yet. Create your first automation.</div>
    <div v-else class="overflow-x-auto bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl">
      <table class="w-full text-left">
        <thead>
          <tr class="border-b border-white/10">
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Name</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Trigger</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Campaign</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Delay</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Status</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="auto in store.automations" :key="auto.id" class="border-b border-white/5 hover:bg-white/5 transition-colors">
            <td class="px-4 py-3 text-sm text-white">{{ auto.name }}</td>
            <td class="px-4 py-3 text-sm text-white/70">
              <span class="text-xs bg-white/10 px-2 py-0.5 rounded-full">{{ auto.trigger_type || 'list_subscribe' }}</span>
              <span v-if="auto.trigger_list_id" class="ml-1 text-xs text-white/50">→ {{ store.audienceLists.find(l => l.id === auto.trigger_list_id)?.name || '#' + auto.trigger_list_id }}</span>
            </td>
            <td class="px-4 py-3 text-sm text-white/70">
              {{ store.emailCampaigns.find(c => c.id === auto.campaign_id)?.name || (auto.campaign_id ? '#' + auto.campaign_id : '—') }}
            </td>
            <td class="px-4 py-3 text-sm text-white/70">{{ auto.delay_hours > 0 ? auto.delay_hours + 'h' : 'Immediate' }}</td>
            <td class="px-4 py-3">
              <span class="px-2 py-1 rounded-full text-xs font-medium" :class="{ 'bg-green-500/20 text-green-400': auto.status === 'active', 'bg-gray-500/20 text-gray-400': auto.status !== 'active' }">{{ auto.status || 'inactive' }}</span>
            </td>
            <td class="px-4 py-3">
              <div class="flex gap-2">
                <button class="px-3 py-1 text-xs bg-white/5 text-white/60 rounded-lg hover:text-white" @click="viewRuns(auto)">Runs</button>
                <button class="px-3 py-1 text-xs bg-red-500/20 text-red-400 rounded-lg hover:bg-red-500/30" @click="deleteAutomation(auto.id)">Delete</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Automation Runs History -->
    <div v-if="selectedAutomationId" class="space-y-3">
      <div class="flex items-center justify-between">
        <h4 class="text-sm font-bold text-white">Run History</h4>
        <button class="text-white/40 hover:text-white" @click="selectedAutomationId = null; store.automationRuns = []">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>
      <div v-if="store.automationRuns.length === 0" class="text-center py-6 text-white/40 text-sm">No runs recorded for this automation.</div>
      <div v-else class="overflow-x-auto bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl">
        <table class="w-full text-left">
          <thead>
            <tr class="border-b border-white/10">
              <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Status</th>
              <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Scheduled</th>
              <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Processed</th>
              <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Error</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="run in store.automationRuns" :key="run.id" class="border-b border-white/5">
              <td class="px-4 py-3">
                <span class="px-2 py-1 rounded-full text-xs font-medium" :class="{ 'bg-green-500/20 text-green-400': run.status === 'sent', 'bg-blue-500/20 text-blue-400': run.status === 'pending', 'bg-red-500/20 text-red-400': run.status === 'failed' }">{{ run.status }}</span>
              </td>
              <td class="px-4 py-3 text-sm text-white/70">{{ run.scheduled_at ? new Date(run.scheduled_at).toLocaleString() : '—' }}</td>
              <td class="px-4 py-3 text-sm text-white/70">{{ run.processed_at ? new Date(run.processed_at).toLocaleString() : '—' }}</td>
              <td class="px-4 py-3 text-sm text-red-400/80">{{ run.error || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useMarketingStore } from '../../stores/marketingStore'
import { useToast } from '../../composables/useToast'

const emit = defineEmits(['reload'])

const store = useMarketingStore()
const toast = useToast()

const showForm = ref(false)
const form = ref({ name: '', trigger_type: 'list_subscribe', trigger_list_id: '', delay_hours: 0, campaign_id: '', status: 'active' })
const selectedAutomationId = ref(null)

async function saveAutomation() {
  if (!form.value.name) { toast.error('Name is required'); return }
  if (!form.value.trigger_list_id) { toast.error('Select a trigger list'); return }
  if (!form.value.campaign_id) { toast.error('Select an email campaign to send'); return }
  try {
    await store.createAutomation({
      name: form.value.name,
      trigger_type: form.value.trigger_type,
      trigger_list_id: Number(form.value.trigger_list_id),
      delay_hours: form.value.delay_hours || 0,
      campaign_id: Number(form.value.campaign_id),
      status: form.value.status || 'active'
    })
    toast.success('Automation created')
    showForm.value = false
    form.value = { name: '', trigger_type: 'list_subscribe', trigger_list_id: '', delay_hours: 0, campaign_id: '', status: 'active' }
    emit('reload')
  } catch {
    toast.error('Failed to save automation')
  }
}

async function deleteAutomation(id) {
  try {
    await store.deleteAutomation(id)
    emit('reload')
  } catch {
    toast.error('Failed to delete automation')
  }
}

async function viewRuns(auto) {
  selectedAutomationId.value = auto.id
  try {
    await store.fetchAutomationRuns(auto.id)
  } catch {
    toast.error('Failed to load automation runs')
  }
}
</script>
