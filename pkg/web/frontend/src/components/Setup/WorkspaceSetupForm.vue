<template>
  <div class="workspace-setup-form animate-fade-in">
    <h2 class="text-2xl font-bold text-white mb-4">
      Set up your Workspace
    </h2>
    <p class="text-slate-300 mb-6">
      Customize your workspace with skills and example files to get started quickly.
    </p>

    <!-- Skills Selection -->
    <div class="mb-8">
      <h3 class="text-lg font-semibold text-white mb-3 flex items-center">
        <span class="text-blue-400 mr-2">🧠</span>
        Install Skills
      </h3>
      <p class="text-sm text-slate-400 mb-4">
        Skills are pre-built capabilities that extend your agent's functionality.
      </p>

      <div class="space-y-3">
        <label
          v-for="skill in availableSkills"
          :key="skill.id"
          class="flex items-start p-4 bg-slate-700/50 rounded-lg cursor-pointer hover:bg-slate-700 transition-colors border border-transparent hover:border-blue-500/30"
        >
          <input
            v-model="selectedSkills"
            type="checkbox"
            :value="skill.id"
            class="mt-1 mr-3 w-5 h-5 text-blue-600 bg-slate-600 border-slate-500 rounded focus:ring-blue-500 focus:ring-2"
          >
          <div class="flex-1">
            <div class="font-medium text-white">{{ skill.name }}</div>
            <div class="text-sm text-slate-400 mt-1">{{ skill.description }}</div>
          </div>
        </label>
      </div>

      <button
        class="mt-4 text-sm text-blue-400 hover:text-blue-300 underline"
        @click="selectAllSkills"
      >
        Select All Skills
      </button>
    </div>

    <!-- Example Files -->
    <div class="mb-6">
      <h3 class="text-lg font-semibold text-white mb-3 flex items-center">
        <span class="text-blue-400 mr-2">📁</span>
        Example Files
      </h3>
      <p class="text-sm text-slate-400 mb-4">
        Create sample files to help you understand how to work with your agent.
      </p>

      <label class="flex items-start p-4 bg-slate-700/50 rounded-lg cursor-pointer hover:bg-slate-700 transition-colors">
        <input
          v-model="includeExampleFiles"
          type="checkbox"
          class="mt-1 mr-3 w-5 h-5 text-blue-600 bg-slate-600 border-slate-500 rounded focus:ring-blue-500 focus:ring-2"
        >
        <div class="flex-1">
          <div class="font-medium text-white">Create Example Files</div>
          <div class="text-sm text-slate-400 mt-1">
            Adds sample markdown files with examples of agent interactions and workflows
          </div>
        </div>
      </label>
    </div>

    <!-- Quick Setup Button -->
    <div class="mb-6">
      <button
        class="w-full py-3 px-4 rounded-lg font-medium bg-slate-700 text-blue-300 hover:bg-slate-600 transition-colors border border-blue-500/30"
        @click="useDefaults"
      >
        ⚡ Use Recommended Defaults
      </button>
      <p class="text-xs text-slate-400 mt-2 text-center">
        Installs {{ recommendedSkills.length }} recommended skills and creates example files
      </p>
    </div>

    <!-- Preview Section -->
    <div
      v-if="selectedSkills.length > 0 || includeExampleFiles"
      class="bg-blue-900/20 border border-blue-500/30 rounded-lg p-4"
    >
      <h4 class="font-medium text-blue-300 mb-2">
        What will be installed:
      </h4>
      <ul class="space-y-1 text-sm text-blue-200">
        <li v-if="selectedSkills.length > 0">
          ✓ {{ selectedSkills.length }} skill{{ selectedSkills.length > 1 ? 's' : '' }}
        </li>
        <li v-if="includeExampleFiles">
          ✓ Example files and documentation
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({ skills: [], exampleFiles: false })
  }
})

const emit = defineEmits(['update:modelValue'])

const selectedSkills = ref(props.modelValue.skills || [])
const includeExampleFiles = ref(props.modelValue.exampleFiles || false)

const availableSkills = [
  {
    id: 'github',
    name: 'GitHub Integration',
    description: 'Create issues, PRs, and manage repositories directly from chat'
  },
  {
    id: 'summarize',
    name: 'Text Summarization',
    description: 'Summarize long documents, articles, or chat histories'
  },
  {
    id: 'weather',
    name: 'Weather Information',
    description: 'Get current weather and forecasts for any location'
  },
  {
    id: 'skill-creator',
    name: 'Skill Creator',
    description: 'Create new custom skills directly from chat'
  },
  {
    id: 'development',
    name: 'Development Tools',
    description: 'Code analysis, refactoring, and debugging assistance'
  },
  {
    id: 'tmux',
    name: 'Tmux Management',
    description: 'Manage tmux sessions and terminal multiplexing'
  }
]

const recommendedSkills = computed(() => {
  return ['github', 'summarize', 'skill-creator']
})

const useDefaults = () => {
  selectedSkills.value = [...recommendedSkills.value]
  includeExampleFiles.value = true
}

const selectAllSkills = () => {
  if (selectedSkills.value.length === availableSkills.length) {
    selectedSkills.value = []
  } else {
    selectedSkills.value = availableSkills.map(s => s.id)
  }
}

// Watch for changes and emit to parent
watch([selectedSkills, includeExampleFiles], () => {
  emit('update:modelValue', {
    skills: selectedSkills.value,
    exampleFiles: includeExampleFiles.value
  })
}, { deep: true })
</script>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.3s ease-in;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
