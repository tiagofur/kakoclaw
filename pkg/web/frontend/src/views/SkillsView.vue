<template>
  <div class="h-full flex flex-col bg-kakoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-kakoclaw-border bg-kakoclaw-surface flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold bg-gradient-to-r from-kakoclaw-accent to-purple-500 bg-clip-text text-transparent">Skills</h2>
        <p class="text-sm text-kakoclaw-text-secondary mt-1">Manage installed skills and browse the marketplace</p>
      </div>
      <div class="flex bg-kakoclaw-bg rounded-lg p-1 border border-kakoclaw-border">
        <button
          @click="activeTab = 'installed'"
          class="px-4 py-1.5 rounded-md text-sm font-medium transition-all"
          :class="activeTab === 'installed' ? 'bg-white dark:bg-gray-700 shadow-sm text-kakoclaw-accent' : 'text-kakoclaw-text-secondary hover:text-kakoclaw-text'"
        >Installed</button>
        <button
          @click="activeTab = 'marketplace'; loadAvailable()"
          class="px-4 py-1.5 rounded-md text-sm font-medium transition-all"
          :class="activeTab === 'marketplace' ? 'bg-white dark:bg-gray-700 shadow-sm text-kakoclaw-accent' : 'text-kakoclaw-text-secondary hover:text-kakoclaw-text'"
        >Marketplace</button>
        <button
          @click="activeTab = 'create'"
          class="px-4 py-1.5 rounded-md text-sm font-medium transition-all"
          :class="activeTab === 'create' ? 'bg-white dark:bg-gray-700 shadow-sm text-kakoclaw-accent' : 'text-kakoclaw-text-secondary hover:text-kakoclaw-text'"
        >Create with AI</button>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-6 custom-scrollbar">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-kakoclaw-accent"></div>
      </div>

      <template v-else>
        <!-- Installed Skills -->
        <div v-if="activeTab === 'installed'">
          <div v-if="skills.length === 0" class="text-center py-12 text-kakoclaw-text-secondary">
            <p class="text-lg">No skills installed</p>
            <p class="text-sm mt-2">Browse the marketplace to install skills</p>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div
              v-for="skill in skills"
              :key="skill.name"
              class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5 hover:border-kakoclaw-accent/50 transition-colors"
            >
              <div class="flex items-start justify-between">
                <div class="flex-1 min-w-0">
                  <h3 class="font-semibold truncate">{{ skill.name }}</h3>
                  <p class="text-sm text-kakoclaw-text-secondary mt-1 line-clamp-2">{{ skill.description || 'No description' }}</p>
                </div>
                <span class="ml-2 px-2 py-0.5 text-xs rounded-full flex-shrink-0"
                  :class="{
                    'bg-emerald-500/10 text-emerald-400': skill.source === 'workspace',
                    'bg-blue-500/10 text-blue-400': skill.source === 'global',
                    'bg-gray-500/10 text-gray-400': skill.source === 'builtin'
                  }"
                >{{ skill.source }}</span>
              </div>
              <div class="flex items-center gap-2 mt-4">
                <button
                  @click="viewSkill(skill.name)"
                  class="px-3 py-1.5 text-xs bg-kakoclaw-bg rounded-lg hover:bg-kakoclaw-border/50 transition-colors"
                >Ver</button>
                <button
                  v-if="skill.source === 'workspace'"
                  @click="editSkill(skill.name)"
                  class="px-3 py-1.5 text-xs bg-kakoclaw-accent/10 text-kakoclaw-accent rounded-lg hover:bg-kakoclaw-accent/20 transition-colors"
                >Editar</button>
                <button
                  v-if="skill.source === 'workspace'"
                  @click="uninstallSkill(skill.name)"
                  class="px-3 py-1.5 text-xs text-red-400 bg-red-500/10 rounded-lg hover:bg-red-500/20 transition-colors"
                >Desinstalar</button>
              </div>
            </div>
          </div>
        </div>

        <!-- Marketplace -->
        <div v-if="activeTab === 'marketplace'">
          <div v-if="loadingAvailable" class="flex items-center justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-kakoclaw-accent"></div>
          </div>
          <div v-else-if="available.length === 0" class="text-center py-12 text-kakoclaw-text-secondary">
            <p class="text-lg">No skills available in marketplace</p>
          </div>
          <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div
              v-for="skill in available"
              :key="skill.name"
              class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5"
            >
              <h3 class="font-semibold">{{ skill.name }}</h3>
              <p class="text-sm text-kakoclaw-text-secondary mt-1 line-clamp-2">{{ skill.description }}</p>
              <p class="text-xs text-kakoclaw-text-secondary mt-2">by {{ skill.author || 'unknown' }}</p>
              <div v-if="skill.tags && skill.tags.length" class="flex flex-wrap gap-1 mt-2">
                <span v-for="tag in skill.tags" :key="tag" class="px-2 py-0.5 text-xs bg-kakoclaw-bg rounded-full text-kakoclaw-text-secondary">{{ tag }}</span>
              </div>
              <button
                @click="installSkill(skill.repository)"
                :disabled="installing === skill.repository"
                class="mt-4 px-4 py-1.5 text-sm bg-kakoclaw-accent text-white rounded-lg hover:bg-kakoclaw-accent/90 transition-colors disabled:opacity-50"
              >
                <span v-if="installing === skill.repository">Installing...</span>
                <span v-else>Install</span>
              </button>
            </div>
          </div>
        </div>

        <!-- AI Creator -->
        <div v-if="activeTab === 'create'" class="max-w-4xl mx-auto space-y-4">
          <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5 space-y-4">
            <h3 class="font-semibold text-lg">Create local skill with AI</h3>
            <p class="text-sm text-kakoclaw-text-secondary">Generate a draft, review it, then save it to workspace skills.</p>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <label class="text-xs text-kakoclaw-text-secondary">Skill name (slug)</label>
                <input v-model="creator.name" type="text" placeholder="e.g. jira-assistant" class="mt-1 w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm">
              </div>
              <div>
                <label class="text-xs text-kakoclaw-text-secondary">Goal *</label>
                <input v-model="creator.goal" type="text" placeholder="What this skill should accomplish" class="mt-1 w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm">
              </div>
            </div>

            <div>
              <label class="text-xs text-kakoclaw-text-secondary">Capabilities</label>
              <textarea v-model="creator.capabilities" rows="3" class="mt-1 w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm"></textarea>
            </div>
            <div>
              <label class="text-xs text-kakoclaw-text-secondary">Safety constraints</label>
              <textarea v-model="creator.constraints" rows="3" class="mt-1 w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm"></textarea>
            </div>
            <div>
              <label class="text-xs text-kakoclaw-text-secondary">Tools available</label>
              <input v-model="creator.tools" type="text" placeholder="e.g. shell, web_fetch, filesystem" class="mt-1 w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm">
            </div>
            <div>
              <label class="text-xs text-kakoclaw-text-secondary">Example interactions</label>
              <textarea v-model="creator.examples" rows="3" class="mt-1 w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm"></textarea>
            </div>

            <div class="flex items-center gap-2">
              <button @click="generateDraft" :disabled="creatingDraft" class="px-4 py-2 text-sm bg-kakoclaw-accent text-white rounded-lg hover:bg-kakoclaw-accent/90 disabled:opacity-50">
                <span v-if="creatingDraft">Generating...</span>
                <span v-else>Generate draft</span>
              </button>
              <button v-if="draftContent" @click="saveSkill(false)" :disabled="savingSkill" class="px-4 py-2 text-sm bg-emerald-600 text-white rounded-lg hover:bg-emerald-500 disabled:opacity-50">
                <span v-if="savingSkill">Saving...</span>
                <span v-else>Save skill</span>
              </button>
            </div>

            <p v-if="creatorError" class="text-sm text-red-400">{{ creatorError }}</p>
          </div>

          <div v-if="draftContent" class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5">
            <h4 class="font-medium mb-2">Draft preview (editable)</h4>
            <textarea v-model="draftContent" rows="18" class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm font-mono"></textarea>
          </div>
        </div>
      </template>
    </div>

    <!-- Skill Editor Modal -->
    <div v-if="editingSkill" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="editingSkill = null">
      <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl max-w-4xl w-full h-[85vh] flex flex-col shadow-2xl">
        <div class="flex items-center justify-between p-4 border-b border-kakoclaw-border">
          <div class="flex items-center gap-2">
            <h3 class="font-bold text-lg text-kakoclaw-accent">Editar Skill:</h3>
            <span class="font-mono text-sm bg-kakoclaw-bg px-2 py-0.5 rounded border border-kakoclaw-border">{{ editingSkill.name }}</span>
          </div>
          <button @click="editingSkill = null" class="p-1 hover:bg-kakoclaw-bg rounded text-kakoclaw-text-secondary hover:text-kakoclaw-text transition-colors">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>
        <div class="flex-1 overflow-hidden p-4">
          <div class="h-full flex flex-col gap-3">
             <p class="text-xs text-kakoclaw-text-secondary">Modifica el contenido markdown de la skill. Asegúrate de mantener el bloque YAML (frontmatter) al principio.</p>
             <textarea 
               v-model="editContent" 
               class="flex-1 w-full p-4 bg-kakoclaw-bg border border-kakoclaw-border rounded-xl text-sm font-mono focus:ring-2 focus:ring-kakoclaw-accent/30 focus:border-kakoclaw-accent transition-all resize-none shadow-inner custom-scrollbar"
               spellcheck="false"
             ></textarea>
          </div>
        </div>
        <div class="p-4 border-t border-kakoclaw-border flex justify-end gap-3">
          <button @click="editingSkill = null" class="px-5 py-2 text-sm font-medium text-kakoclaw-text-secondary hover:text-kakoclaw-text transition-colors">Cancelar</button>
          <button 
            @click="saveEditedSkill" 
            :disabled="savingSkill" 
            class="px-6 py-2 text-sm font-semibold bg-kakoclaw-accent text-white rounded-lg hover:bg-kakoclaw-accent-hover shadow-lg shadow-kakoclaw-accent/20 disabled:opacity-50 transition-all flex items-center gap-2"
          >
            <span v-if="savingSkill" class="animate-spin h-3.5 w-3.5 border-2 border-white/30 border-t-white rounded-full"></span>
            {{ savingSkill ? 'Guardando...' : 'Guardar Cambios' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Skill Viewer Modal -->
    <div v-if="viewingSkill" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="viewingSkill = null">
      <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl max-w-2xl w-full max-h-[80vh] flex flex-col shadow-xl">
        <div class="flex items-center justify-between p-4 border-b border-kakoclaw-border">
          <h3 class="font-semibold">{{ viewingSkill.name }}</h3>
          <button @click="viewingSkill = null" class="p-1 hover:bg-kakoclaw-bg rounded text-kakoclaw-text-secondary hover:text-kakoclaw-text">&times;</button>
        </div>
        <div class="flex-1 overflow-auto p-4 custom-scrollbar">
          <pre class="whitespace-pre-wrap text-sm font-mono">{{ viewingSkill.content }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'

const toast = useToast()
const activeTab = ref('installed')
const loading = ref(true)
const loadingAvailable = ref(false)
const skills = ref([])
const available = ref([])
const installing = ref(null)
const viewingSkill = ref(null)
const editingSkill = ref(null)
const editContent = ref('')
const creatingDraft = ref(false)
const savingSkill = ref(false)
const creatorError = ref('')
const draftContent = ref('')
const creator = ref({
  name: '',
  goal: '',
  capabilities: '',
  constraints: '',
  tools: '',
  examples: ''
})

const loadSkills = async () => {
  loading.value = true
  try {
    const data = await advancedService.fetchSkills()
    skills.value = data.skills || []
  } catch (err) {
    console.error('Failed to load skills:', err)
  } finally {
    loading.value = false
  }
}

const loadAvailable = async () => {
  if (available.value.length > 0) return
  loadingAvailable.value = true
  try {
    const data = await advancedService.fetchAvailableSkills()
    available.value = data.skills || []
    if (data.warning) {
      toast.info(data.warning)
    }
  } catch (err) {
    console.error('Failed to load marketplace:', err)
    toast.error('Failed to load marketplace')
  } finally {
    loadingAvailable.value = false
  }
}

const viewSkill = async (name) => {
  try {
    const data = await advancedService.viewSkill(name)
    viewingSkill.value = data
  } catch (err) {
    toast.error('Failed to load skill content')
  }
}

const editSkill = async (name) => {
  try {
    const data = await advancedService.viewSkill(name)
    editingSkill.value = { name: data.name }
    editContent.value = data.content
  } catch (err) {
    toast.error('Failed to load skill for editing')
  }
}

const saveEditedSkill = async () => {
  if (!editContent.value.trim()) return
  
  savingSkill.value = true
  try {
    await advancedService.createSkill({
      name: editingSkill.value.name,
      content: editContent.value,
      overwrite: true
    })
    toast.success('Skill actualizada correctamente')
    editingSkill.value = null
    await loadSkills()
  } catch (err) {
    toast.error(err?.response?.data || 'Error al actualizar la skill')
  } finally {
    savingSkill.value = false
  }
}

const installSkill = async (repo) => {
  installing.value = repo
  try {
    await advancedService.installSkill(repo)
    toast.success('Skill installed successfully')
    await loadSkills()
  } catch (err) {
    toast.error('Failed to install skill')
  } finally {
    installing.value = null
  }
}

const uninstallSkill = async (name) => {
  try {
    await advancedService.uninstallSkill(name)
    toast.success('Skill uninstalled')
    await loadSkills()
  } catch (err) {
    toast.error('Failed to uninstall skill')
  }
}

const generateDraft = async () => {
  creatorError.value = ''
  if (!creator.value.goal.trim()) {
    creatorError.value = 'Goal is required'
    return
  }
  creatingDraft.value = true
  try {
    const data = await advancedService.generateSkillDraft({
      name: creator.value.name,
      goal: creator.value.goal,
      capabilities: creator.value.capabilities,
      constraints: creator.value.constraints,
      tools: creator.value.tools,
      examples: creator.value.examples
    })
    creator.value.name = data.name || creator.value.name
    draftContent.value = data.draft || ''
    toast.success('Draft generated')
  } catch (err) {
    creatorError.value = err?.response?.data || 'Failed to generate draft'
  } finally {
    creatingDraft.value = false
  }
}

const saveSkill = async (overwrite) => {
  creatorError.value = ''
  if (!creator.value.name.trim()) {
    creatorError.value = 'Skill name is required'
    return
  }
  if (!draftContent.value.trim()) {
    creatorError.value = 'Draft content is empty'
    return
  }
  savingSkill.value = true
  try {
    await advancedService.createSkill({
      name: creator.value.name,
      content: draftContent.value,
      overwrite
    })
    toast.success('Skill saved locally')
    draftContent.value = ''
    creator.value = { name: '', goal: '', capabilities: '', constraints: '', tools: '', examples: '' }
    activeTab.value = 'installed'
    await loadSkills()
  } catch (err) {
    if (err?.response?.status === 409 && !overwrite) {
      const confirmed = confirm('Skill already exists. Overwrite it?')
      if (confirmed) {
        await saveSkill(true)
      }
      return
    }
    creatorError.value = err?.response?.data || 'Failed to save skill'
  } finally {
    savingSkill.value = false
  }
}

onMounted(() => loadSkills())
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 8px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background-color: rgba(156, 163, 175, 0.5); border-radius: 4px; }
.line-clamp-2 { display: -webkit-box; -webkit-line-clamp: 2; line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
</style>
