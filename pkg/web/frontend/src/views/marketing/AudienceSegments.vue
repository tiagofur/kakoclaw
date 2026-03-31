<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h3 class="text-lg font-semibold text-white">Segments</h3>
      <button
        class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl font-medium"
        @click="showForm = true; form = { name: '', rules: [] }"
      >+ Create Segment</button>
    </div>

    <!-- Segment Builder -->
    <div v-if="showForm" class="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-5 space-y-4">
      <input v-model="form.name" placeholder="Segment Name *" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">

      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold text-white/50 uppercase tracking-wider">Rules</span>
          <button class="text-xs text-pink-400 hover:text-pink-300" @click="form.rules.push({ field: 'email', operator: 'contains', value: '' })">+ Add Rule</button>
        </div>
        <div v-for="(rule, idx) in form.rules" :key="idx" class="flex gap-2 items-center">
          <select v-model="rule.field" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white text-sm focus:border-pink-500/50 focus:outline-none">
            <option v-for="f in ['email', 'first_name', 'last_name', 'phone', 'company', 'title', 'tags', 'status']" :key="f" :value="f">{{ f }}</option>
          </select>
          <select v-model="rule.operator" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white text-sm focus:border-pink-500/50 focus:outline-none">
            <option v-for="op in ['equals', 'not_equals', 'contains', 'starts_with', 'ends_with', 'greater_than', 'less_than', 'in_list', 'not_in_list']" :key="op" :value="op">{{ op }}</option>
          </select>
          <input v-model="rule.value" placeholder="Value" class="flex-1 px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
          <button class="text-red-400 hover:text-red-300" @click="form.rules.splice(idx, 1)">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>
        <div v-if="form.rules.length === 0" class="text-xs text-white/30 text-center py-2">No rules yet. Click "+ Add Rule" to define targeting criteria.</div>
      </div>

      <div class="flex items-center justify-between">
        <p class="text-xs text-white/30">Save the segment to preview the matching contact count.</p>
        <div class="flex gap-2">
          <button class="px-4 py-2 text-sm text-white/60 hover:text-white" @click="showForm = false">Cancel</button>
          <button class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 text-white rounded-xl font-medium" @click="saveSegment">Save Segment</button>
        </div>
      </div>
    </div>

    <div v-if="store.audienceSegments.length === 0" class="text-center py-12 text-white/40 text-sm">No segments yet. Create a segment to target specific audiences.</div>
    <div v-else class="grid gap-3">
      <div
        v-for="seg in store.audienceSegments"
        :key="seg.id"
        class="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-5 flex items-center justify-between hover:border-pink-500/20 transition-colors"
      >
        <div>
          <h4 class="text-sm font-semibold text-white">{{ seg.name }}</h4>
          <p class="text-xs text-white/40 mt-0.5">
            {{ JSON.parse(seg.rules || '[]').length }} rule(s) ·
            <span v-if="previewCounts[seg.id] !== undefined" class="text-pink-400 font-medium">{{ previewCounts[seg.id] }} contacts (live)</span>
            <span v-else>{{ seg.contact_count }} contacts</span>
          </p>
        </div>
        <div class="flex gap-2">
          <button
            class="px-3 py-1.5 text-xs bg-white/5 text-white/50 rounded-lg hover:text-white transition-colors flex items-center gap-1.5"
            :class="{ 'opacity-50': previewing[seg.id] }"
            :disabled="previewing[seg.id]"
            @click="recount(seg)"
          >
            <svg v-if="previewing[seg.id]" class="animate-spin w-3 h-3" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/></svg>
            <svg v-else class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
            Recount
          </button>
          <button class="px-3 py-1.5 text-xs bg-red-500/20 text-red-400 rounded-lg hover:bg-red-500/30" @click="deleteSegment(seg.id)">Delete</button>
        </div>
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
const form = ref({ name: '', rules: [] })
const previewCounts = ref({})
const previewing = ref({})

async function saveSegment() {
  if (!form.value.name) { toast.error('Name is required'); return }
  try {
    await store.createAudienceSegment({ name: form.value.name, rules: JSON.stringify(form.value.rules) })
    showForm.value = false
    form.value = { name: '', rules: [] }
    emit('reload')
  } catch {
    toast.error('Failed to create segment')
  }
}

async function deleteSegment(id) {
  try {
    await store.deleteAudienceSegment(id)
    emit('reload')
  } catch {
    toast.error('Failed to delete segment')
  }
}

async function recount(seg) {
  if (previewing.value[seg.id]) return
  previewing.value[seg.id] = true
  try {
    const rules = JSON.parse(seg.rules || '[]')
    const result = await store.previewAudienceSegment(seg.id, rules)
    if (result !== null) {
      previewCounts.value[seg.id] = result.count ?? result.contact_count ?? 0
    }
  } catch {
    toast.error('Failed to recount segment')
  } finally {
    previewing.value[seg.id] = false
  }
}
</script>
