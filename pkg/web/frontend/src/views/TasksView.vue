<template>
  <div class="flex flex-col h-full bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh (Subtle) -->
    <div class="absolute inset-0 pointer-events-none opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-blue-500/20 via-transparent to-transparent"></div>

    <!-- Filters & Controls -->
    <div class="glass-sticky top-0 z-20 p-2 sm:p-3 md:p-4 border-b border-makoclaw-border/20">
      <div class="flex flex-col gap-2 sm:gap-3">
        <!-- Always visible row: Search + filter toggle + create -->
        <div class="flex items-center gap-2 sm:gap-3">
          <!-- Search -->
          <div class="flex-1 relative group">
            <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-makoclaw-text-secondary group-focus-within:text-makoclaw-accent transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search tasks..."
              class="w-full pl-9 sm:pl-10 pr-3 sm:pr-4 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent transition-all text-sm backdrop-blur-sm min-h-[36px]"
            >
          </div>

          <!-- Filter toggle (mobile only) -->
          <button
            @click="showFilters = !showFilters"
            class="md:hidden p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-xl border border-makoclaw-border/50 hover:bg-makoclaw-accent/10 hover:text-makoclaw-accent transition-all"
            :class="showFilters ? 'bg-makoclaw-accent/10 text-makoclaw-accent border-makoclaw-accent/30' : 'text-makoclaw-text-secondary'"
            title="Toggle filters"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
            </svg>
          </button>

          <!-- Create task button -->
          <button
            @click="showNewTaskModal = true"
            class="px-3 sm:px-5 py-2 min-h-[36px] bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl transition-all shadow-lg shadow-makoclaw-accent/20 hover:shadow-makoclaw-accent/40 text-sm font-bold flex items-center gap-1.5 sm:gap-2 active:scale-95 flex-shrink-0"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
            <span class="hidden sm:inline">New Task</span>
          </button>
        </div>

        <!-- Collapsible filters row (hidden on mobile by default) -->
        <div :class="['flex items-center gap-2 sm:gap-3 overflow-x-auto scrollbar-hide transition-all duration-200', showFilters ? '' : 'hidden md:flex']">
          <select
            v-model="sortBy"
            @change="taskStore.setSortBy(sortBy)"
            class="px-3 sm:px-4 py-2 min-h-[36px] bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 text-sm hover:border-makoclaw-accent/30 transition-all cursor-pointer backdrop-blur-sm outline-none"
          >
            <option value="recent">Recent First</option>
            <option value="oldest">Oldest First</option>
            <option value="a-z">Name (A-Z)</option>
            <option value="z-a">Name (Z-A)</option>
          </select>

          <select
            v-model="statusFilter"
            @change="taskStore.setFilter('status', statusFilter)"
            class="px-3 sm:px-4 py-2 min-h-[36px] bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 text-sm hover:border-makoclaw-accent/30 transition-all cursor-pointer backdrop-blur-sm outline-none"
          >
            <option value="">All Statuses</option>
            <option value="backlog">Backlog</option>
            <option value="todo">To Do</option>
            <option value="in_progress">In Progress</option>
            <option value="review">Review</option>
            <option value="done">Done</option>
          </select>

          <label class="flex items-center gap-2 px-3 py-1.5 sm:py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl backdrop-blur-sm cursor-pointer min-h-[36px] flex-shrink-0">
            <input
              type="checkbox"
              v-model="showArchived"
              class="rounded border-makoclaw-border bg-makoclaw-surface text-makoclaw-accent focus:ring-makoclaw-accent transition-all cursor-pointer"
            >
            <span class="text-xs sm:text-sm text-makoclaw-text-secondary whitespace-nowrap">Archived</span>
          </label>
        </div>
      </div>
    </div>

    <!-- Kanban Board -->
    <div class="flex-1 overflow-x-auto p-2 sm:p-3 md:p-4" style="scroll-snap-type: x mandatory; -webkit-overflow-scrolling: touch;">
      <div class="flex gap-4 min-h-full">
        <!-- Column: Backlog -->
        <div class="flex-shrink-0 w-64 sm:w-72 md:w-80 lg:w-[340px]" style="scroll-snap-align: start;">
          <KanbanColumn
            status="backlog"
            title="Backlog"
            :tasks="taskStore.tasksByStatus.backlog"
            @task-click="openTaskDetails"
            @task-drop="moveTask"
          />
        </div>

        <!-- Column: To Do -->
        <div class="flex-shrink-0 w-64 sm:w-72 md:w-80 lg:w-[340px]" style="scroll-snap-align: start;">
          <KanbanColumn
            status="todo"
            title="To Do"
            :tasks="taskStore.tasksByStatus.todo"
            @task-click="openTaskDetails"
            @task-drop="moveTask"
          />
        </div>

        <!-- Column: In Progress -->
        <div class="flex-shrink-0 w-64 sm:w-72 md:w-80 lg:w-[340px]" style="scroll-snap-align: start;">
          <KanbanColumn
            status="in_progress"
            title="In Progress"
            :tasks="taskStore.tasksByStatus.in_progress"
            @task-click="openTaskDetails"
            @task-drop="moveTask"
          />
        </div>

        <!-- Column: Review -->
        <div class="flex-shrink-0 w-64 sm:w-72 md:w-80 lg:w-[340px]" style="scroll-snap-align: start;">
          <KanbanColumn
            status="review"
            title="Review"
            :tasks="taskStore.tasksByStatus.review"
            @task-click="openTaskDetails"
            @task-drop="moveTask"
          />
        </div>

        <!-- Column: Done -->
        <div class="flex-shrink-0 w-64 sm:w-72 md:w-80 lg:w-[340px]" style="scroll-snap-align: start;">
          <KanbanColumn
            status="done"
            title="Done"
            :tasks="taskStore.tasksByStatus.done"
            @task-click="openTaskDetails"
            @task-drop="moveTask"
          />
        </div>
      </div>
    </div>

    <!-- New Task Modal -->
    <NewTaskModal
      v-if="showNewTaskModal"
      @close="showNewTaskModal = false"
      @created="handleTaskCreated"
    />

    <!-- Task Details Modal -->
    <TaskDetailsModal
      v-if="selectedTask"
      :task="selectedTask"
      @close="selectedTask = null"
      @updated="handleTaskUpdated"
      @deleted="handleTaskDeleted"
      @archived="handleTaskArchived"
      @unarchived="handleTaskUnarchived"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, watch, onBeforeUnmount } from 'vue'
import { useTaskStore } from '../stores/taskStore'
import { TaskWebSocket } from '../services/websocketService'
import taskService from '../services/taskService'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'
import KanbanColumn from '../components/Tasks/KanbanColumn.vue'
import NewTaskModal from '../components/Tasks/NewTaskModal.vue'
import TaskDetailsModal from '../components/Tasks/TaskDetailsModal.vue'

const taskStore = useTaskStore()
const toast = useToast()
const searchQuery = ref('')
const statusFilter = ref('')
const sortBy = ref('recent')
const showArchived = ref(false)
const showNewTaskModal = ref(false)
const selectedTask = ref(null)
const showExportMenu = ref(false)
const showFilters = ref(false)
const exportDropdownRef = ref(null)

const taskWs = new TaskWebSocket()
const handleTaskWsUpdate = (message) => {
  if (message.type === 'task_updated') {
    taskStore.updateTask(message.task_id || message.task.id, message.task)
  } else if (message.type === 'task_created') {
    taskStore.addTask(message.task)
  } else if (message.type === 'task_deleted') {
    taskStore.removeTask(message.task_id)
  }
}

const handleExport = (format) => {
  advancedService.exportTasks(format)
  showExportMenu.value = false
  toast.success(`Exporting tasks as ${format.toUpperCase()}`)
}

// Close export menu on outside click
const handleClickOutside = (e) => {
  if (exportDropdownRef.value && !exportDropdownRef.value.contains(e.target)) {
    showExportMenu.value = false
  }
}

const fetchTasks = async () => {
  try {
    const data = await taskService.fetchTasks(showArchived.value)
    const tasks = data.tasks || []
    
    taskStore.setTasks(
      tasks.map(t => ({
        ...t,
        created_at: t.created_at || new Date().toISOString()
      }))
    )
  } catch (error) {
    console.error('Failed to fetch tasks:', error)
  }
}

onMounted(async () => {
  document.addEventListener('click', handleClickOutside)
  
  // Clear any stale filters in the store to ensure tasks are visible
  taskStore.clearFilter()
  searchQuery.value = ''
  statusFilter.value = ''
  
  await fetchTasks()

  // Connect to WebSocket for real-time updates
  try {
    await taskWs.connect()
    taskStore.setWebSocket(taskWs)

    taskWs.on('update', handleTaskWsUpdate)
  } catch (error) {
    console.error('Failed to connect to tasks WebSocket:', error)
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  taskWs.off('update', handleTaskWsUpdate)
  taskWs.disconnect()
})

// Watch filters
watch([showArchived], () => {
  fetchTasks()
})

// Watch search query
watch(searchQuery, (newVal) => {
  taskStore.setFilter('search', newVal)
})

const openTaskDetails = (task) => {
  selectedTask.value = task
}

const moveTask = async (taskId, newStatus, sourceStatus) => {
  if (sourceStatus === newStatus) {
    return
  }
  try {
    await taskService.updateTaskStatus(taskId, newStatus)
    taskStore.updateTask(taskId, { status: newStatus })
    toast.info(`Task moved to ${newStatus.replace('_', ' ')}`)
  } catch (error) {
    console.error('Failed to move task:', error)
    toast.error('Failed to move task')
  }
}

const handleTaskCreated = async (taskData) => {
  try {
    const newTask = await taskService.createTask(taskData.title, taskData.description, taskData.status)
    taskStore.addTask(newTask)
    showNewTaskModal.value = false
    toast.success('Task created')
  } catch (error) {
    console.error('Failed to create task:', error)
    toast.error('Failed to create task')
  }
}

const handleTaskUpdated = async (taskId, updates) => {
  try {
    await taskService.updateTask(taskId, updates)
    taskStore.updateTask(taskId, updates)
    toast.success('Task updated')
  } catch (error) {
    console.error('Failed to update task:', error)
    toast.error('Failed to update task')
  }
}

const handleTaskDeleted = async (taskId) => {
  try {
    await taskService.deleteTask(taskId)
    taskStore.removeTask(taskId)
    selectedTask.value = null
    toast.success('Task deleted')
  } catch (error) {
    console.error('Failed to delete task:', error)
    toast.error('Failed to delete task')
  }
}

const handleTaskArchived = async (taskId) => {
  try {
    await taskService.archiveTask(taskId)
    if (!showArchived.value) {
      taskStore.removeTask(taskId)
    } else {
      taskStore.updateTask(taskId, { archived: true })
    }
    selectedTask.value = null
    toast.success('Task archived')
  } catch (error) {
    console.error('Failed to archive task:', error)
    toast.error('Failed to archive task')
  }
}

const handleTaskUnarchived = async (taskId) => {
  try {
    await taskService.unarchiveTask(taskId)
    taskStore.updateTask(taskId, { archived: false })
    selectedTask.value = null
    toast.success('Task unarchived')
  } catch (error) {
    console.error('Failed to unarchive task:', error)
    toast.error('Failed to unarchive task')
  }
}
</script>


