import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useTaskStore = defineStore('tasks', () => {
  const tasks = ref([])
  const selectedTask = ref(null)
  const filter = ref({
    search: '',
    status: '',
    dateFrom: null,
    dateTo: null
  })
  const sortBy = ref('recent')
  const isLoading = ref(false)
  const ws = ref(null)

  const filteredTasks = computed(() => {
    let result = [...tasks.value]

    // Filter by search
    if (filter.value.search) {
      const search = filter.value.search.toLowerCase()
      result = result.filter(t => 
        t.title.toLowerCase().includes(search) || 
        (t.description && t.description.toLowerCase().includes(search))
      )
    }

    // Filter by status
    if (filter.value.status) {
      result = result.filter(t => t.status === filter.value.status)
    }

    // Filter by date range
    if (filter.value.dateFrom || filter.value.dateTo) {
      result = result.filter(t => {
        const taskDate = new Date(t.created_at)
        const from = filter.value.dateFrom ? new Date(filter.value.dateFrom) : null
        const to = filter.value.dateTo ? new Date(filter.value.dateTo) : null
        
        if (from && taskDate < from) return false
        if (to && taskDate > to) return false
        return true
      })
    }

    return result
  })

  const tasksByStatus = computed(() => {
    const statuses = ['backlog', 'todo', 'in_progress', 'review', 'done']
    const grouped = {}
    
    statuses.forEach(status => {
      grouped[status] = filteredTasks.value.filter(t => t.status === status)
    })
    
    // Apply sorting
    Object.keys(grouped).forEach(status => {
      grouped[status] = sortTasks(grouped[status])
    })
    
    return grouped
  })

  function sortTasks(taskList) {
    const sorted = [...taskList]
    
    switch (sortBy.value) {
      case 'recent':
        return sorted.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
      case 'oldest':
        return sorted.sort((a, b) => new Date(a.created_at) - new Date(b.created_at))
      case 'a-z':
        return sorted.sort((a, b) => a.title.localeCompare(b.title))
      case 'z-a':
        return sorted.sort((a, b) => b.title.localeCompare(a.title))
      default:
        return sorted
    }
  }

  function setTasks(newTasks) {
    tasks.value = newTasks
  }

  function addTask(task) {
    tasks.value.push(task)
  }

  function updateTask(id, updates) {
    const idx = tasks.value.findIndex(t => t.id === id)
    if (idx !== -1) {
      tasks.value[idx] = { ...tasks.value[idx], ...updates }
    }
  }

  function removeTask(id) {
    tasks.value = tasks.value.filter(t => t.id !== id)
  }

  function setSelectedTask(task) {
    selectedTask.value = task
  }

  function setFilter(key, value) {
    filter.value[key] = value
  }

  function clearFilter() {
    filter.value = {
      search: '',
      status: '',
      dateFrom: null,
      dateTo: null
    }
  }

  function setSortBy(value) {
    sortBy.value = value
  }

  function setLoading(loading) {
    isLoading.value = loading
  }

  function setWebSocket(websocket) {
    ws.value = websocket
  }

  return {
    tasks,
    selectedTask,
    filter,
    sortBy,
    isLoading,
    ws,
    filteredTasks,
    tasksByStatus,
    setTasks,
    addTask,
    updateTask,
    removeTask,
    setSelectedTask,
    setFilter,
    clearFilter,
    setSortBy,
    setLoading,
    setWebSocket
  }
})
