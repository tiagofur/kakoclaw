<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h3 class="text-lg font-semibold text-white">Contact Lists</h3>
      <button
        class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl font-medium"
        @click="showForm = true; form = { name: '', description: '', type: 'static' }"
      >+ Create List</button>
    </div>

    <!-- Create List Form -->
    <div v-if="showForm" class="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-5 space-y-3">
      <input v-model="form.name" placeholder="List Name *" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
      <input v-model="form.description" placeholder="Description" class="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 text-sm focus:border-pink-500/50 focus:outline-none">
      <select v-model="form.type" class="px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-white text-sm focus:border-pink-500/50 focus:outline-none">
        <option value="static">Static</option>
        <option value="dynamic">Dynamic</option>
      </select>
      <div class="flex gap-2 justify-end">
        <button class="px-4 py-2 text-sm text-white/60 hover:text-white" @click="showForm = false">Cancel</button>
        <button class="px-4 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 text-white rounded-xl font-medium" @click="createList">Create</button>
      </div>
    </div>

    <div v-if="store.audienceLists.length === 0" class="text-center py-12 text-white/40 text-sm">No lists yet. Create your first contact list.</div>
    <div v-else class="grid gap-3">
      <div
        v-for="list in store.audienceLists"
        :key="list.id"
        class="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-5 flex items-center justify-between hover:border-pink-500/20 transition-colors"
      >
        <div>
          <h4 class="text-sm font-semibold text-white">{{ list.name }}</h4>
          <p class="text-xs text-white/40 mt-0.5">{{ list.description || 'No description' }} · {{ list.type }}</p>
        </div>
        <button class="px-3 py-1.5 text-xs bg-red-500/20 text-red-400 rounded-lg hover:bg-red-500/30" @click="deleteList(list.id)">Delete</button>
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
const form = ref({ name: '', description: '', type: 'static' })

async function createList() {
  if (!form.value.name) { toast.error('Name is required'); return }
  try {
    await store.createAudienceList(form.value)
    showForm.value = false
    form.value = { name: '', description: '', type: 'static' }
    emit('reload')
  } catch {
    toast.error('Failed to create list')
  }
}

async function deleteList(id) {
  try {
    await store.deleteAudienceList(id)
    emit('reload')
  } catch {
    toast.error('Failed to delete list')
  }
}
</script>
