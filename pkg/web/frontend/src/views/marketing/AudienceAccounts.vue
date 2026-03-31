<template>
  <div class="space-y-4">
    <!-- Header -->
    <div class="flex items-center justify-between gap-3">
      <h3 class="text-lg font-semibold text-white">Company Accounts</h3>
      <button
        class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl font-medium"
        @click="openNewForm"
      >+ Add Account</button>
    </div>

    <!-- New account form -->
    <div v-if="showForm" class="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-5 space-y-3">
      <h4 class="text-sm font-bold text-white">{{ editingAccount ? 'Edit Profile' : 'New Account' }}</h4>
      <div class="grid grid-cols-2 gap-3">
        <input v-model="form.account" :disabled="!!editingAccount" placeholder="Account / Brand Name *" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none disabled:opacity-50" />
        <input v-model="form.website" placeholder="Website URL" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none" />
      </div>
      <div class="grid grid-cols-2 gap-3">
        <input v-model="form.industry" placeholder="Industry" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none" />
        <input v-model="form.target_audience" placeholder="Target Audience" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none" />
      </div>
      <textarea v-model="form.description" placeholder="Description..." rows="3" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none resize-none" />
      <div class="flex gap-2 justify-end">
        <button class="px-4 py-2 text-sm text-white/60 hover:text-white" @click="showForm = false">Cancel</button>
        <button class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 text-white rounded-xl font-medium" :disabled="!form.account.trim()" @click="saveProfile">Save</button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="store.companyProfilesLoading" class="flex items-center justify-center py-8">
      <svg class="animate-spin w-6 h-6 text-pink-400" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/></svg>
    </div>

    <!-- Empty state -->
    <div v-else-if="store.companyProfiles.length === 0 && !showForm" class="text-center py-12 text-white/40 text-sm">
      No company profiles yet. Add an account to get started, or click Research to auto-fill from the web.
    </div>

    <!-- Profile cards -->
    <div v-else class="grid gap-4">
      <div
        v-for="profile in store.companyProfiles"
        :key="profile.account"
        class="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-5 space-y-3"
      >
        <!-- Card header -->
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="flex items-center gap-2 flex-wrap">
              <h4 class="text-base font-bold text-white truncate">{{ profile.account }}</h4>
              <span v-if="profile.industry" class="px-2 py-0.5 text-xs bg-pink-500/20 text-pink-400 rounded-full font-medium">{{ profile.industry }}</span>
              <span v-if="profile.researched_at" class="px-2 py-0.5 text-xs bg-emerald-500/20 text-emerald-400 rounded-full font-medium">Researched</span>
            </div>
            <a v-if="profile.website" :href="profile.website" target="_blank" class="text-xs text-cyan-400 hover:text-cyan-300 truncate block mt-0.5">{{ profile.website }}</a>
          </div>
          <div class="flex gap-2 flex-shrink-0">
            <button
              class="px-3 py-1.5 text-xs bg-violet-500/20 text-violet-400 rounded-lg hover:bg-violet-500/30 transition-colors flex items-center gap-1"
              :disabled="researchingAccount === profile.account"
              @click="triggerResearch(profile)"
            >
              <svg v-if="researchingAccount === profile.account" class="animate-spin w-3 h-3" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/></svg>
              <svg v-else class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
              {{ researchingAccount === profile.account ? 'Researching...' : 'Research' }}
            </button>
            <button
              class="px-3 py-1.5 text-xs bg-white/5 text-white/60 rounded-lg hover:text-white transition-colors"
              @click="editProfile(profile)"
            >Edit</button>
          </div>
        </div>

        <!-- Profile details -->
        <div v-if="profile.description || profile.target_audience" class="grid gap-2">
          <p v-if="profile.description" class="text-sm text-white/70">{{ profile.description }}</p>
          <p v-if="profile.target_audience" class="text-xs text-white/50"><span class="text-white/30">Target: </span>{{ profile.target_audience }}</p>
        </div>

        <!-- Research notes -->
        <div v-if="profile.research_notes" class="p-3 bg-violet-500/10 border border-violet-500/20 rounded-xl">
          <p class="text-xs font-semibold text-violet-400 mb-1">Research Notes</p>
          <p class="text-sm text-white/70 whitespace-pre-wrap">{{ profile.research_notes }}</p>
          <p v-if="profile.researched_at" class="text-xs text-white/30 mt-1">{{ new Date(profile.researched_at).toLocaleString() }}</p>
        </div>

        <!-- Social links -->
        <div v-if="parsedSocials(profile).length > 0" class="flex flex-wrap gap-2">
          <a
            v-for="[platform, url] in parsedSocials(profile)"
            :key="platform"
            :href="url"
            target="_blank"
            class="px-2.5 py-1 text-xs bg-white/5 border border-white/10 text-white/60 hover:text-white rounded-lg capitalize transition-colors"
          >{{ platform }}</a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useMarketingStore } from '../../stores/marketingStore'
import { useToast } from '../../composables/useToast'

const emit = defineEmits(['reload'])

const store = useMarketingStore()
const toast = useToast()

const showForm = ref(false)
const editingAccount = ref(null)
const researchingAccount = ref(null)
const form = ref({ account: '', website: '', industry: '', target_audience: '', description: '' })

onMounted(() => store.fetchCompanyProfiles())

function openNewForm() {
  editingAccount.value = null
  form.value = { account: '', website: '', industry: '', target_audience: '', description: '' }
  showForm.value = true
}

function editProfile(profile) {
  editingAccount.value = profile.account
  form.value = {
    account: profile.account,
    website: profile.website || '',
    industry: profile.industry || '',
    target_audience: profile.target_audience || '',
    description: profile.description || ''
  }
  showForm.value = true
}

async function saveProfile() {
  if (!form.value.account.trim()) return
  try {
    await store.saveCompanyProfile(form.value.account, form.value)
    toast.success('Profile saved')
    showForm.value = false
    editingAccount.value = null
  } catch {
    toast.error('Failed to save profile')
  }
}

async function triggerResearch(profile) {
  researchingAccount.value = profile.account
  try {
    await store.researchCompany(profile.account, profile.website || '')
    toast.success('Research complete')
  } catch (e) {
    toast.error('Research failed: ' + (e?.response?.data || e.message || 'unknown error'))
  } finally {
    researchingAccount.value = null
  }
}

function parsedSocials(profile) {
  try {
    const links = JSON.parse(profile.social_links || '{}')
    return Object.entries(links).filter(([, v]) => v)
  } catch {
    return []
  }
}
</script>
