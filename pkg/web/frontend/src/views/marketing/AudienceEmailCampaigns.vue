<template>
  <div class="space-y-4">
    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex items-center gap-3">
        <h3 class="text-lg font-semibold text-white">Email Campaigns</h3>
        <div class="flex gap-1 bg-white/5 rounded-lg p-0.5">
          <button
            class="px-3 py-1.5 text-xs font-medium rounded-md transition-all"
            :class="filter === 'all' ? 'bg-pink-500/20 text-pink-400' : 'text-white/60 hover:text-white/80'"
            @click="filter = 'all'; emit('reload')"
          >All</button>
          <button
            class="px-3 py-1.5 text-xs font-medium rounded-md transition-all"
            :class="filter === 'archived' ? 'bg-pink-500/20 text-pink-400' : 'text-white/60 hover:text-white/80'"
            @click="filter = 'archived'; emit('reload')"
          >Archived</button>
        </div>
      </div>
      <div class="flex gap-2">
        <button
          v-if="filter === 'all'"
          class="px-4 py-2 text-sm bg-white/5 border border-white/10 rounded-xl text-white/70 hover:text-white transition-colors"
          @click="exportCSV"
        >
          <svg class="w-4 h-4 inline mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a2 2 0 002 2h12a2 2 0 002-2v-1M12 4v12m0 0l-4-4m4 4l4-4" /></svg>
          Export CSV
        </button>
        <button
          v-if="filter === 'all'"
          class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl font-medium"
          @click="campaignForm = { active: true, id: null, name: '', subject: '', body_html: '', template_slug: '', list_id: '' }"
        >+ New Campaign</button>
      </div>
    </div>

    <!-- Campaign Form -->
    <div v-if="campaignForm.active" class="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-6 space-y-3">
      <div class="grid grid-cols-2 gap-3">
        <input v-model="campaignForm.name" placeholder="Campaign Name *" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
        <input v-model="campaignForm.subject" placeholder="Subject *" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
      </div>
      <textarea v-model="campaignForm.body_html" placeholder="HTML Body..." rows="6" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none resize-none" />
      <input v-model="campaignForm.template_slug" placeholder="Template Slug (optional)" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
      <select v-model="campaignForm.list_id" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white text-sm focus:border-pink-500/50 focus:outline-none">
        <option value="">Select a list...</option>
        <option v-for="list in store.audienceLists" :key="list.id" :value="list.id">{{ list.name }}</option>
      </select>
      <div class="flex gap-2 justify-end">
        <button class="px-4 py-2 text-sm text-white/60 hover:text-white" @click="campaignForm.active = false">Cancel</button>
        <button class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 text-white rounded-xl font-medium" @click="saveCampaign">Save</button>
      </div>
    </div>

    <!-- Empty states -->
    <div v-if="store.emailCampaignsLoading" class="flex items-center justify-center py-8">
      <svg class="animate-spin w-6 h-6 text-pink-400" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/></svg>
    </div>
    <div v-else-if="filter === 'archived' && filteredCampaigns.length === 0" class="flex flex-col items-center justify-center py-16 text-center">
      <p class="text-sm font-semibold text-white/40">No archived campaigns</p>
      <p class="text-xs text-white/25 mt-1">Archived campaigns will appear here.</p>
    </div>
    <div v-else-if="filteredCampaigns.length === 0" class="text-center py-8 text-white/40 text-sm">No email campaigns yet. Create your first campaign.</div>

    <!-- Campaigns Table -->
    <div v-else class="overflow-x-auto bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl">
      <table class="w-full text-left">
        <thead>
          <tr class="border-b border-white/10">
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Name</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Subject</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Status</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Sent</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Date</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="camp in filteredCampaigns" :key="camp.id" class="border-b border-white/5 hover:bg-white/5 transition-colors">
            <td class="px-4 py-3 text-sm text-white">{{ camp.name }}</td>
            <td class="px-4 py-3 text-sm text-white/70">{{ camp.subject }}</td>
            <td class="px-4 py-3">
              <span class="px-2 py-1 rounded-full text-xs font-medium" :class="{ 'bg-gray-500/20 text-gray-400': camp.status === 'draft', 'bg-blue-500/20 text-blue-400': camp.status === 'sending', 'bg-green-500/20 text-green-400': camp.status === 'sent', 'bg-red-500/20 text-red-400': camp.status === 'failed', 'bg-slate-500/20 text-slate-400': camp.status === 'archived' }">{{ camp.status }}</span>
            </td>
            <td class="px-4 py-3 text-sm text-white/50">{{ camp.sent_count || 0 }}</td>
            <td class="px-4 py-3 text-sm text-white/40">{{ camp.created_at ? new Date(camp.created_at).toLocaleDateString() : '—' }}</td>
            <td class="px-4 py-3">
              <div class="flex gap-2">
                <button v-if="camp.status === 'draft' && filter === 'all'" class="px-3 py-1 text-xs bg-blue-500/20 text-blue-400 rounded-lg hover:bg-blue-500/30" :disabled="sending" @click="sendCampaign(camp)">Send</button>
                <button v-if="camp.status === 'draft' && filter === 'all'" class="px-3 py-1 text-xs bg-white/5 text-white/60 rounded-lg hover:text-white" @click="editCampaign(camp)">Edit</button>
                <button class="px-3 py-1 text-xs bg-violet-500/20 text-violet-400 rounded-lg hover:bg-violet-500/30" @click="toggleVariants(camp)">A/B</button>
                <button class="px-3 py-1 text-xs bg-teal-500/20 text-teal-400 rounded-lg hover:bg-teal-500/30" @click="toggleMemory(camp)">Memory</button>
                <button v-if="filter === 'all' && camp.status !== 'archived'" class="px-3 py-1 text-xs bg-amber-500/20 text-amber-400 rounded-lg hover:bg-amber-500/30" @click="archiveCampaign(camp)">Archive</button>
                <button v-if="filter === 'archived'" class="px-3 py-1 text-xs bg-white/5 text-white/60 rounded-lg hover:text-white" @click="restoreCampaign(camp)">Restore</button>
                <button v-if="filter === 'archived'" class="px-3 py-1 text-xs bg-red-500/20 text-red-400 rounded-lg hover:bg-red-500/30" @click="deleteCampaignPermanently(camp)">Delete</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Campaign Memory Panel -->
    <div v-if="selectedMemoryCampaignId" class="bg-white/5 backdrop-blur-xl border border-teal-500/20 rounded-2xl p-5 space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <h4 class="text-sm font-bold text-white">Campaign Memory</h4>
          <p class="text-xs text-white/40 mt-0.5">{{ store.emailCampaigns.find(c => c.id === selectedMemoryCampaignId)?.name }}</p>
        </div>
        <button class="text-white/40 hover:text-white" @click="selectedMemoryCampaignId = null; store.campaignMemory = []">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>

      <div v-if="store.campaignMemoryLoading" class="flex justify-center py-6">
        <svg class="animate-spin w-5 h-5 text-teal-400" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/></svg>
      </div>

      <div v-else class="space-y-2">
        <div v-if="store.campaignMemory.length === 0" class="text-center py-4 text-white/40 text-sm">No memory entries yet. Add notes, decisions, or agent context below.</div>
        <div v-for="entry in store.campaignMemory" :key="entry.id" class="flex items-start gap-3 p-3 bg-white/5 border border-white/10 rounded-xl group">
          <span class="mt-0.5 px-1.5 py-0.5 text-xs rounded font-medium flex-shrink-0"
                :class="{ 'bg-teal-500/20 text-teal-400': entry.role === 'note', 'bg-blue-500/20 text-blue-400': entry.role === 'agent', 'bg-amber-500/20 text-amber-400': entry.role === 'decision' }">
            {{ entry.role }}
          </span>
          <p class="flex-1 text-sm text-white/80 whitespace-pre-wrap">{{ entry.content }}</p>
          <div class="flex-shrink-0 text-right">
            <p class="text-xs text-white/30">{{ new Date(entry.created_at).toLocaleString() }}</p>
            <button class="mt-1 text-xs text-red-400/60 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-opacity" @click="removeMemoryEntry(entry.id)">Delete</button>
          </div>
        </div>
      </div>

      <form class="flex gap-2" @submit.prevent="submitMemoryEntry">
        <select v-model="newMemoryRole" class="px-2 py-2 bg-white/5 border border-white/10 rounded-xl text-white/70 text-xs focus:border-teal-500/50 focus:outline-none">
          <option value="note">note</option>
          <option value="decision">decision</option>
          <option value="agent">agent</option>
        </select>
        <input v-model="newMemoryContent" type="text" placeholder="Add a note, decision, or agent context..." class="flex-1 px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/30 text-sm focus:border-teal-500/50 focus:outline-none" />
        <button type="submit" :disabled="!newMemoryContent.trim()" class="px-4 py-2 text-sm bg-teal-600 hover:bg-teal-500 disabled:opacity-40 text-white rounded-xl font-medium transition-colors">Add</button>
      </form>
    </div>

    <!-- A/B Variants Panel -->
    <div v-if="selectedVariantCampaignId" class="bg-white/5 backdrop-blur-xl border border-violet-500/20 rounded-2xl p-5 space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <h4 class="text-sm font-bold text-white">A/B Test Variants</h4>
          <p class="text-xs text-white/40 mt-0.5">{{ store.emailCampaigns.find(c => c.id === selectedVariantCampaignId)?.name }}</p>
        </div>
        <button class="text-white/40 hover:text-white" @click="selectedVariantCampaignId = null; store.experimentVariants = []">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>

      <div v-if="store.experimentVariantsLoading" class="flex justify-center py-6">
        <svg class="animate-spin w-5 h-5 text-violet-400" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/></svg>
      </div>
      <div v-else-if="store.experimentVariants.length === 0" class="text-center py-6 text-white/40 text-sm">No variants yet. Add variants to run an A/B test.</div>
      <div v-else class="grid gap-3">
        <div
          v-for="variant in store.experimentVariants"
          :key="variant.id"
          class="bg-white/5 border rounded-xl p-4 space-y-3 transition-colors"
          :class="variant.is_winner ? 'border-yellow-500/40' : 'border-white/10'"
        >
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="flex items-center gap-2">
                <span class="text-sm font-semibold text-white">{{ variant.name }}</span>
                <span v-if="variant.is_winner" class="px-2 py-0.5 text-xs bg-yellow-500/20 text-yellow-400 rounded-full font-medium">Winner</span>
              </div>
              <p class="text-xs text-white/50 mt-0.5">{{ variant.subject }}</p>
            </div>
            <div class="flex items-center gap-2 flex-shrink-0">
              <span class="text-xs bg-violet-500/20 text-violet-400 px-2 py-0.5 rounded-full font-medium">{{ variant.split_percent }}% split</span>
              <button v-if="!variant.is_winner" class="px-3 py-1 text-xs bg-yellow-500/20 text-yellow-400 rounded-lg hover:bg-yellow-500/30" @click="setWinner(selectedVariantCampaignId, variant.id)">Set Winner</button>
            </div>
          </div>
          <div v-if="variantMetrics[variant.id]" class="grid grid-cols-3 gap-3">
            <div class="text-center">
              <p class="text-lg font-bold text-white">{{ variantMetrics[variant.id].sent || 0 }}</p>
              <p class="text-xs text-white/40">Sent</p>
            </div>
            <div class="text-center">
              <p class="text-lg font-bold text-green-400">{{ variantMetrics[variant.id].opened || 0 }}</p>
              <p class="text-xs text-white/40">Opened <span v-if="variantMetrics[variant.id].sent" class="text-green-400/70">({{ Math.round((variantMetrics[variant.id].opened / variantMetrics[variant.id].sent) * 100) }}%)</span></p>
            </div>
            <div class="text-center">
              <p class="text-lg font-bold text-blue-400">{{ variantMetrics[variant.id].clicked || 0 }}</p>
              <p class="text-xs text-white/40">Clicked <span v-if="variantMetrics[variant.id].sent" class="text-blue-400/70">({{ Math.round((variantMetrics[variant.id].clicked / variantMetrics[variant.id].sent) * 100) }}%)</span></p>
            </div>
          </div>
          <div v-else class="text-xs text-white/30 text-center py-2">No delivery data yet</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useMarketingStore } from '../../stores/marketingStore'
import { useToast } from '../../composables/useToast'
import marketingService from '../../services/marketingService'

const emit = defineEmits(['reload'])

const store = useMarketingStore()
const toast = useToast()

const filter = ref('all')
const sending = ref(false)
const campaignForm = ref({ active: false, id: null, name: '', subject: '', body_html: '', template_slug: '', list_id: '' })
const selectedVariantCampaignId = ref(null)
const variantMetrics = ref({})
const selectedMemoryCampaignId = ref(null)
const newMemoryContent = ref('')
const newMemoryRole = ref('note')

const filteredCampaigns = computed(() => {
  if (filter.value === 'archived') return store.emailCampaigns.filter((c) => c.status === 'archived')
  return store.emailCampaigns.filter((c) => c.status !== 'archived')
})

async function saveCampaign() {
  const form = campaignForm.value
  if (!form.name || !form.subject) { toast.error('Name and subject are required'); return }
  try {
    if (form.id) {
      await store.updateEmailCampaign(form.id, { name: form.name, subject: form.subject, body_html: form.body_html, template_slug: form.template_slug || undefined, list_id: form.list_id || undefined })
      toast.success('Campaign updated')
    } else {
      await store.createEmailCampaign({ name: form.name, subject: form.subject, body_html: form.body_html, template_slug: form.template_slug || undefined, list_id: form.list_id || undefined })
      toast.success('Campaign created')
    }
    campaignForm.value = { active: false, id: null, name: '', subject: '', body_html: '', template_slug: '', list_id: '' }
    emit('reload')
  } catch {
    toast.error('Failed to save campaign')
  }
}

function editCampaign(camp) {
  campaignForm.value = { active: true, id: camp.id, name: camp.name || '', subject: camp.subject || '', body_html: camp.body_html || '', template_slug: camp.template_slug || '', list_id: camp.list_id || '' }
}

async function sendCampaign(camp) {
  if (!window.confirm(`Send "${camp.name}" to contacts?`)) return
  sending.value = true
  try {
    await store.sendEmailCampaign(camp.id)
    toast.success('Campaign sending started')
    emit('reload')
  } catch {
    toast.error('Failed to send campaign')
  } finally {
    sending.value = false
  }
}

async function archiveCampaign(camp) {
  if (!window.confirm(`Archive "${camp.name}"?`)) return
  try {
    await store.updateEmailCampaign(camp.id, { ...camp, status: 'archived' })
    toast.success('Campaign archived')
    emit('reload')
  } catch {
    toast.error('Failed to archive campaign')
  }
}

async function restoreCampaign(camp) {
  try {
    await store.updateEmailCampaign(camp.id, { ...camp, status: 'draft' })
    toast.success('Campaign restored to drafts')
    emit('reload')
  } catch {
    toast.error('Failed to restore campaign')
  }
}

async function deleteCampaignPermanently(camp) {
  if (!window.confirm(`Permanently delete "${camp.name}"? This cannot be undone.`)) return
  try {
    await store.deleteEmailCampaign(camp.id)
    toast.success('Campaign permanently deleted')
    emit('reload')
  } catch {
    toast.error('Failed to delete campaign')
  }
}

async function exportCSV() {
  try {
    const response = await marketingService.audienceListCampaigns({ format: 'csv' })
    const blob = new Blob([response.data], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'email-campaigns.csv'
    a.click()
    window.URL.revokeObjectURL(url)
  } catch {
    toast.error('Failed to export campaigns')
  }
}

async function toggleVariants(camp) {
  if (selectedVariantCampaignId.value === camp.id) {
    selectedVariantCampaignId.value = null
    store.experimentVariants = []
    variantMetrics.value = {}
    return
  }
  selectedVariantCampaignId.value = camp.id
  variantMetrics.value = {}
  try {
    await store.fetchExperimentVariants(camp.id)
    for (const variant of store.experimentVariants) {
      try {
        const res = await store.fetchVariantMetrics(camp.id, variant.id)
        if (res) variantMetrics.value[variant.id] = res
      } catch { /* no metrics yet */ }
    }
  } catch {
    toast.error('Failed to load variants')
  }
}

async function setWinner(campaignId, variantId) {
  try {
    await store.setVariantWinner(campaignId, variantId)
    await store.fetchExperimentVariants(campaignId)
    toast.success('Winner set')
  } catch {
    toast.error('Failed to set winner')
  }
}

async function toggleMemory(camp) {
  if (selectedMemoryCampaignId.value === camp.id) {
    selectedMemoryCampaignId.value = null
    store.campaignMemory = []
    return
  }
  selectedMemoryCampaignId.value = camp.id
  try {
    await store.fetchCampaignMemory(camp.id)
  } catch {
    toast.error('Failed to load campaign memory')
  }
}

async function submitMemoryEntry() {
  if (!newMemoryContent.value.trim()) return
  try {
    await store.addCampaignMemoryEntry(selectedMemoryCampaignId.value, newMemoryContent.value.trim(), newMemoryRole.value)
    newMemoryContent.value = ''
  } catch {
    toast.error('Failed to add memory entry')
  }
}

async function removeMemoryEntry(entryId) {
  try {
    await store.deleteCampaignMemoryEntry(selectedMemoryCampaignId.value, entryId)
  } catch {
    toast.error('Failed to delete memory entry')
  }
}
</script>
