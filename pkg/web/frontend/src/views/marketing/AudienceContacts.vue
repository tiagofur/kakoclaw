<template>
  <div class="space-y-4">
    <!-- Toolbar -->
    <div class="flex flex-wrap items-center gap-3">
      <input
        v-model="search"
        type="text"
        placeholder="Search contacts..."
        class="flex-1 min-w-[200px] px-4 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none"
        @input="searchDebounced()"
      >
      <select
        v-model="statusFilter"
        class="px-4 py-2 bg-white/5 border border-white/10 rounded-xl text-white text-sm focus:border-pink-500/50 focus:outline-none"
        @change="searchDebounced()"
      >
        <option value="">All Status</option>
        <option value="active">Active</option>
        <option value="unsubscribed">Unsubscribed</option>
        <option value="bounced">Bounced</option>
        <option value="complained">Complained</option>
      </select>
      <button
        class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl font-medium"
        @click="showForm = true; form = { email: '', first_name: '', last_name: '', phone: '', company: '', title: '' }"
      >+ Add Contact</button>
      <label class="px-4 py-2 text-sm bg-white/5 border border-white/10 rounded-xl text-white/70 hover:text-white cursor-pointer transition-colors">
        Import CSV
        <input type="file" accept=".csv" class="hidden" @change="importCSV">
      </label>
      <button
        class="px-4 py-2 text-sm bg-white/5 border border-white/10 rounded-xl text-white/70 hover:text-white transition-colors"
        @click="store.exportAudienceContacts({ search, status: statusFilter })"
      >Export CSV</button>
    </div>

    <!-- Create Contact Form -->
    <div v-if="showForm" class="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-5 space-y-3">
      <div class="grid grid-cols-2 gap-3">
        <input v-model="form.email" placeholder="Email *" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
        <input v-model="form.first_name" placeholder="First Name" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
        <input v-model="form.last_name" placeholder="Last Name" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
        <input v-model="form.phone" placeholder="Phone" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
        <input v-model="form.company" placeholder="Company" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
        <input v-model="form.title" placeholder="Title" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
      </div>
      <div class="flex gap-2 justify-end">
        <button class="px-4 py-2 text-sm text-white/60 hover:text-white" @click="showForm = false">Cancel</button>
        <button class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 text-white rounded-xl font-medium" @click="saveContact">Save</button>
      </div>
    </div>

    <!-- Contact Detail -->
    <div v-if="selectedContact" class="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-5 space-y-3">
      <div class="flex items-center justify-between mb-2">
        <h3 class="text-sm font-bold text-white">Contact Detail</h3>
        <div class="flex gap-2">
          <button class="px-3 py-1.5 text-xs bg-red-500/20 text-red-400 rounded-lg hover:bg-red-500/30" @click="deleteContact">Delete</button>
          <button class="text-white/40 hover:text-white" @click="selectedContact = null">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div v-for="field in ['email', 'first_name', 'last_name', 'phone', 'company', 'title', 'status']" :key="field">
          <label class="block text-[10px] font-bold uppercase tracking-wider text-white/40 mb-1">{{ field.replace('_', ' ') }}</label>
          <input v-model="selectedContact[field]" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white text-sm focus:border-pink-500/50 focus:outline-none">
        </div>
      </div>
      <div class="flex justify-end">
        <button class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 text-white rounded-xl font-medium" @click="updateContact">Save Changes</button>
      </div>
    </div>

    <!-- Contacts Table -->
    <div v-if="store.audienceContactsLoading" class="flex items-center justify-center py-12">
      <svg class="animate-spin w-6 h-6 text-pink-400" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
      </svg>
    </div>
    <div v-else-if="store.audienceContacts.length === 0" class="text-center py-12 text-white/40 text-sm">No contacts found. Add your first contact or import a CSV.</div>
    <div v-else class="overflow-x-auto bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl">
      <table class="w-full text-left">
        <thead>
          <tr class="border-b border-white/10">
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Email</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Name</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Company</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Status</th>
            <th class="px-4 py-3 text-xs font-bold text-white/50 uppercase tracking-wider">Tags</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="contact in store.audienceContacts"
            :key="contact.id"
            class="border-b border-white/5 hover:bg-white/5 cursor-pointer transition-colors"
            @click="selectedContact = { ...contact }"
          >
            <td class="px-4 py-3 text-sm text-white">{{ contact.email }}</td>
            <td class="px-4 py-3 text-sm text-white/70">{{ [contact.first_name, contact.last_name].filter(Boolean).join(' ') || '—' }}</td>
            <td class="px-4 py-3 text-sm text-white/70">{{ contact.company || '—' }}</td>
            <td class="px-4 py-3">
              <span class="px-2 py-1 rounded-full text-xs font-medium" :class="{ 'bg-green-500/20 text-green-400': contact.status === 'active', 'bg-yellow-500/20 text-yellow-400': contact.status === 'unsubscribed', 'bg-red-500/20 text-red-400': ['bounced', 'complained'].includes(contact.status) }">{{ contact.status }}</span>
            </td>
            <td class="px-4 py-3 text-sm text-white/50">{{ contact.tags || '—' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="store.audienceContactsTotal > 25" class="flex items-center justify-center gap-4 text-sm">
      <button class="px-3 py-1.5 rounded-lg bg-white/5 border border-white/10 text-white/70 hover:text-white disabled:opacity-30" :disabled="store.audienceContactsPage <= 1" @click="store.fetchAudienceContacts({ search, status: statusFilter, page: store.audienceContactsPage - 1 })">Prev</button>
      <span class="text-white/50">Page {{ store.audienceContactsPage }} of {{ Math.ceil(store.audienceContactsTotal / 25) }}</span>
      <button class="px-3 py-1.5 rounded-lg bg-white/5 border border-white/10 text-white/70 hover:text-white disabled:opacity-30" :disabled="store.audienceContactsPage >= Math.ceil(store.audienceContactsTotal / 25)" @click="store.fetchAudienceContacts({ search, status: statusFilter, page: store.audienceContactsPage + 1 })">Next</button>
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

const search = ref('')
const statusFilter = ref('')
const showForm = ref(false)
const form = ref({ email: '', first_name: '', last_name: '', phone: '', company: '', title: '' })
const selectedContact = ref(null)
let searchTimeout = null

function searchDebounced() {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    store.fetchAudienceContacts({ search: search.value, status: statusFilter.value, page: 1 })
  }, 300)
}

async function saveContact() {
  if (!form.value.email) { toast.error('Email is required'); return }
  try {
    await store.createAudienceContact(form.value)
    showForm.value = false
    form.value = { email: '', first_name: '', last_name: '', phone: '', company: '', title: '' }
    emit('reload')
  } catch {
    toast.error('Failed to create contact')
  }
}

async function updateContact() {
  try {
    await store.updateAudienceContact(selectedContact.value.id, selectedContact.value)
    emit('reload')
  } catch {
    toast.error('Failed to update contact')
  }
}

async function deleteContact() {
  try {
    await store.deleteAudienceContact(selectedContact.value.id)
    selectedContact.value = null
    emit('reload')
  } catch {
    toast.error('Failed to delete contact')
  }
}

async function importCSV(e) {
  const f = e.target.files?.[0]
  if (!f) return
  const fd = new FormData()
  fd.append('file', f)
  try {
    await store.importAudienceContacts(fd)
    e.target.value = ''
    emit('reload')
    toast.success('Contacts imported')
  } catch {
    toast.error('Failed to import contacts')
  }
}
</script>
