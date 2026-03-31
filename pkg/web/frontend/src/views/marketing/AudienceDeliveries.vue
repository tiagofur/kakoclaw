<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h3 class="text-lg font-semibold text-white">Email Deliveries</h3>
    </div>

    <!-- Send Form -->
    <div class="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-5 space-y-3">
      <h4 class="text-sm font-bold text-white/70">Send Email to List</h4>
      <select v-model="sendForm.list_id" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white text-sm focus:border-pink-500/50 focus:outline-none">
        <option value="">Select a list...</option>
        <option v-for="list in store.audienceLists" :key="list.id" :value="list.id">{{ list.name }}</option>
      </select>
      <input v-model="sendForm.subject" placeholder="Subject" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
      <textarea v-model="sendForm.body" placeholder="Email body..." rows="4" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none resize-none" />
      <div class="flex justify-end">
        <button class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 text-white rounded-xl font-medium" @click="sendToList">Send</button>
      </div>
    </div>

    <!-- Deliveries Table -->
    <div v-if="store.audienceDeliveriesLoading" class="flex items-center justify-center py-8">
      <svg class="animate-spin w-6 h-6 text-pink-400" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
      </svg>
    </div>
    <div v-else-if="store.audienceDeliveries.length === 0" class="text-center py-8 text-white/40 text-sm">No deliveries yet. Send an email to a list to get started.</div>
    <div v-else class="overflow-x-auto bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl">
      <table class="w-full text-left">
        <thead>
          <tr class="border-b border-white/10">
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Campaign</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Subject</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Status</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Date</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="del in store.audienceDeliveries" :key="del.id" class="border-b border-white/5">
            <td class="px-4 py-3 text-sm text-white/70">{{ del.campaign_name || '—' }}</td>
            <td class="px-4 py-3 text-sm text-white/70">{{ del.subject || '—' }}</td>
            <td class="px-4 py-3">
              <span class="px-2 py-1 rounded-full text-xs font-medium" :class="{ 'bg-gray-500/20 text-gray-400': del.status === 'queued', 'bg-blue-500/20 text-blue-400': del.status === 'sent', 'bg-green-500/20 text-green-400': del.status === 'delivered', 'bg-purple-500/20 text-purple-400': del.status === 'opened', 'bg-pink-500/20 text-pink-400': del.status === 'clicked', 'bg-red-500/20 text-red-400': ['bounced', 'failed'].includes(del.status) }">{{ del.status }}</span>
            </td>
            <td class="px-4 py-3 text-sm text-white/40">{{ del.sent_at ? new Date(del.sent_at).toLocaleDateString() : '—' }}</td>
          </tr>
        </tbody>
      </table>
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

const sendForm = ref({ list_id: '', subject: '', body: '' })

async function sendToList() {
  if (!sendForm.value.list_id || !sendForm.value.subject) {
    toast.error('List and subject are required')
    return
  }
  try {
    await store.sendAudienceEmail(sendForm.value)
    sendForm.value = { list_id: '', subject: '', body: '' }
    emit('reload')
    toast.success('Email queued for delivery')
  } catch {
    toast.error('Failed to send email')
  }
}
</script>
