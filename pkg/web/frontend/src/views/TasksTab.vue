<template>
  <div class="flex flex-col h-full bg-picoclaw-bg">
    <!-- Filters & Controls -->
    <div class="border-b border-picoclaw-border bg-picoclaw-surface p-4 space-y-3">
      <!-- Search & Sort Row -->
      <div class="flex gap-2 flex-col md:flex-row">
        <div class="flex-1 relative">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search tasks..."
            class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded focus-ring text-sm"
          />
        </div>

        <select
          v-model="sortBy"
          @change="taskStore.setSortBy(sortBy)"
          class="px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded focus-ring text-sm"
        >
          <option value="recent">Recientes</option>
          <option value="oldest">Antiguos</option>
          <option value="a-z">A-Z</option>
          <option value="z-a">Z-A</option>
        </select>

        <select
          v-model="statusFilter"
          @change="taskStore.setFilter('status', statusFilter)"
          class="px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded focus-ring text-sm"
        >
          <option value="">All Status</option>
          <option value="backlog">Backlog</option>
          <option value="todo">To Do</option>
          <option value="in_progress">In Progress</option>
          <option value="review">Review</option>
          <option value="done">Done</option>
        </select>

        <button
          @click="showNewTaskModal = true"
          class="px-4 py-2 bg-picoclaw-accent hover:bg-picoclaw-accent-hover text-white rounded transition-smooth text-sm font-medium"
        >
          New Task
        </button>
      </div>
    </div>

    <!-- Kanban Board -->
    <div class="flex-1 overflow-x-auto p-4">
      <div class="flex gap-4 min-h-full">
        <!-- Column: Backlog -->
        <div class="flex-shrink-0 w-80">
          <KanbanColumn
            status="backlog"
            title="Backlog"
            :tasks="taskStore.tasksByStatus.backlog"
            @task-click="openTaskDetails"
            @task-drop="moveTask"
          />
        </div>

        <!-- Column: To Do -->
        <div class="flex-shrink-0 w-80">
          <KanbanColumn
            status="todo"
            title="To Do"
            :tasks="taskStore.tasksByStatus.todo"
            @task-click="openTaskDetails"
            @task-drop="moveTask"
          />
        </div>

        <!-- Column: In Progress -->
        <div class="flex-shrink-0 w-80">
          <KanbanColumn
            status="in_progress"
            title="In Progress"
            :tasks="taskStore.tasksByStatus.in_progress"
            @task-click="openTaskDetails"
            @task-drop="moveTask"
          />
        </div>

        <!-- Column: Review -->
        <div class="flex-shrink-0 w-80">
          <KanbanColumn
            status="review"
            title="Review"
            :tasks="taskStore.tasksByStatus.review"
            @task-click="openTaskDetails"
            @task-drop="moveTask"
          />
        </div>

        <!-- Column: Done -->
        <div class="flex-shrink-0 w-80">
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
    />
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useTaskStore } from '../stores/taskStore'
import { TaskWebSocket } from '../services/websocketService'
import taskService from '../services/taskService'
import KanbanColumn from '../components/Tasks/KanbanColumn.vue'
import NewTaskModal from '../components/Tasks/NewTaskModal.vue'
import TaskDetailsModal from '../components/Tasks/TaskDetailsModal.vue'

const taskStore = useTaskStore()
const searchQuery = ref('')
const statusFilter = ref('')
const sortBy = ref('recent')
const showNewTaskModal = ref(false)
const selectedTask = ref(null)

const taskWs = new TaskWebSocket()

onMounted(async () => {
  // Fetch initial tasks
  try {
    const tasks = await taskService.fetchTasks()
    taskStore.setTasks(
      tasks.map(t => ({
        ...t,
        created_at: t.created_at || new Date().toISOString()
      }))
    )
  } catch (error) {
    console.error('Failed to fetch tasks:', error)
  }

  // Connect to WebSocket for real-time updates
  try {
    await taskWs.connect()
    taskStore.setWebSocket(taskWs)

    taskWs.on('update', (message) => {
      if (message.type === 'task_updated') {
        taskStore.updateTask(message.task.id, message.task)
      } else if (message.type === 'task_created') {
        taskStore.addTask(message.task)
      } else if (message.type === 'task_deleted') {
        taskStore.removeTask(message.task_id)
      }
    })
  } catch (error) {
    console.error('Failed to connect to tasks WebSocket:', error)
  }
})

// Watch search query
watch(searchQuery, (newVal) => {
  taskStore.setFilter('search', newVal)
})

const openTaskDetails = (task) => {
  selectedTask.value = task
}

const moveTask = async (taskId, newStatus) => {
  try {
    await taskService.updateTaskStatus(taskId, newStatus)
    taskStore.updateTask(taskId, { status: newStatus })
  } catch (error) {
    console.error('Failed to move task:', error)
  }
}

const handleTaskCreated = async (taskData) => {
  try {
    const newTask = await taskService.createTask(taskData.title, taskData.description, taskData.status)
    taskStore.addTask(newTask)
    showNewTaskModal.value = false
  } catch (error) {
    console.error('Failed to create task:', error)
  }
}

const handleTaskUpdated = async (taskId, updates) => {
  try {
    await taskService.updateTask(taskId, updates)
    taskStore.updateTask(taskId, updates)
  } catch (error) {
    console.error('Failed to update task:', error)
  }
}

const handleTaskDeleted = async (taskId) => {
  try {
    await taskService.deleteTask(taskId)
    taskStore.removeTask(taskId)
    selectedTask.value = null
  } catch (error) {
    console.error('Failed to delete task:', error)
  }
}
</script>
