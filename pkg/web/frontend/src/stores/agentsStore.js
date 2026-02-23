import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useToast } from '../composables/useToast'
import { useAuthStore } from './authStore'
import api from '../services/api'

export const useAgentsStore = defineStore('agents', () => {
  const toast = useToast()
  const authStore = useAuthStore()

  const loading = ref(false)
  const specialists = ref([])
  const orchestrator = ref(null)
  const metrics = ref(null)
  const selectedSpecialist = ref(null)
  const showSpecialistModal = ref(false)
  const specialistFormMode = ref('create')
  const aiGenerating = ref(false)

  const orchestrated = computed(() => orchestrator.value?.enabled || false)
  const specialistsCount = computed(() => specialists.value.length)

  async function fetchAgents() {
    loading.value = true
    try {
      const response = await api.get('/agents')
      specialists.value = response.data.specialists || []
      orchestrator.value = response.data.orchestrator || null
    } catch (error) {
      toast.error('Failed to load agents')
      console.error(error)
    } finally {
      loading.value = false
    }
  }

  async function fetchSpecialistDetails(name) {
    loading.value = true
    try {
      const response = await api.get(`/agents/${name}`)
      return response.data
    } catch (error) {
      toast.error('Failed to load specialist details')
      console.error(error)
      return null
    } finally {
      loading.value = false
    }
  }

  async function createSpecialist(specialistData) {
    loading.value = true
    try {
      await api.post('/agents/specialist', specialistData)
      await fetchAgents()
      toast.success('Specialist created successfully')
      return true
    } catch (error) {
      toast.error('Failed to create specialist')
      console.error(error)
      return false
    } finally {
      loading.value = false
    }
  }

  async function updateSpecialist(name, specialistData) {
    loading.value = true
    try {
      await api.put(`/agents/specialist/${name}`, specialistData)
      await fetchAgents()
      toast.success('Specialist updated successfully')
      return true
    } catch (error) {
      toast.error('Failed to update specialist')
      console.error(error)
      return false
    } finally {
      loading.value = false
    }
  }

  async function deleteSpecialist(name) {
    loading.value = true
    try {
      await api.delete(`/agents/specialist/${name}`)
      await fetchAgents()
      toast.success('Specialist deleted successfully')
      return true
    } catch (error) {
      toast.error('Failed to delete specialist')
      console.error(error)
      return false
    } finally {
      loading.value = false
    }
  }

  async function generateSpecialistWithAI(description) {
    aiGenerating.value = true
    try {
      const response = await api.post('/agents/generate', { description })
      return response.data.specialist
    } catch (error) {
      toast.error('Failed to generate specialist with AI')
      console.error(error)
      return null
    } finally {
      aiGenerating.value = false
    }
  }

  async function testSpecialist(name, testMessage) {
    loading.value = true
    try {
      const response = await api.post(`/agents/${name}/test`, { message: testMessage })
      return response.data
    } catch (error) {
      toast.error('Failed to test specialist')
      console.error(error)
      return null
    } finally {
      loading.value = false
    }
  }

  async function fetchMetrics(period = '7d') {
    loading.value = true
    try {
      const response = await api.get(`/agents/metrics?period=${period}`)
      metrics.value = response.data
    } catch (error) {
      toast.error('Failed to load metrics')
      console.error(error)
    } finally {
      loading.value = false
    }
  }

  async function updateOrchestrator(config) {
    loading.value = true
    try {
      await api.put('/agents/orchestrator', config)
      orchestrator.value = { ...orchestrator.value, ...config }
      toast.success('Orchestrator updated successfully')
      return true
    } catch (error) {
      toast.error('Failed to update orchestrator')
      console.error(error)
      return false
    } finally {
      loading.value = false
    }
  }

  function openSpecialistModal(mode = 'create', specialist = null) {
    specialistFormMode.value = mode
    selectedSpecialist.value = specialist
    showSpecialistModal.value = true
  }

  function closeSpecialistModal() {
    showSpecialistModal.value = false
    selectedSpecialist.value = null
    specialistFormMode.value = 'create'
  }

  function getSpecialistIcon(name) {
    const icons = {
      developer: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4 4" />',
      documentation: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2 2v11a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 002-2M9 5a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />',
      testing: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />',
      devops: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002 2M5 12a2 2 0 00-2 2m-2-4h.01M9 16h.01" />',
      analyst: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 8v8m-4-5v5m4-4v5a2 2 0 00-2-2h8a2 2 0 002-2z" />',
      researcher: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />'
    }
    return icons[name] || '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7 7z" />'
  }

  function getSpecialistColor(name) {
    const colors = {
      developer: 'text-blue-400',
      documentation: 'text-blue-400',
      testing: 'text-amber-400',
      devops: 'text-purple-400',
      analyst: 'text-cyan-400',
      researcher: 'text-pink-400'
    }
    return colors[name] || 'text-gray-400'
  }

  function getSpecialistBgColor(name) {
    const colors = {
      developer: 'bg-blue-500/10',
      documentation: 'bg-blue-500/10',
      testing: 'bg-amber-500/10',
      devops: 'bg-purple-500/10',
      analyst: 'bg-cyan-500/10',
      researcher: 'bg-pink-500/10'
    }
    return colors[name] || 'bg-gray-500/10'
  }

  return {
    loading,
    specialists,
    orchestrator,
    metrics,
    selectedSpecialist,
    showSpecialistModal,
    specialistFormMode,
    aiGenerating,
    orchestrated,
    specialistsCount,
    fetchAgents,
    fetchSpecialistDetails,
    createSpecialist,
    updateSpecialist,
    deleteSpecialist,
    generateSpecialistWithAI,
    testSpecialist,
    fetchMetrics,
    updateOrchestrator,
    openSpecialistModal,
    closeSpecialistModal,
    getSpecialistIcon,
    getSpecialistColor,
    getSpecialistBgColor
  }
})

