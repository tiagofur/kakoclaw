<template>
  <div class="space-y-5 animate-fade-in-up">
    <!-- Developer Mode -->
    <div class="glass-panel rounded-2xl p-5 border border-makoclaw-border/50 transition-all duration-300 hover:border-makoclaw-accent/30 hover:shadow-lg">
      <div class="flex items-center justify-between mb-5">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-xl bg-yellow-500/10 text-yellow-400">
            <CommandLineIcon class="w-5 h-5" />
          </div>
          <div>
            <h3 class="text-xs font-medium tracking-wide text-makoclaw-text-secondary/70 uppercase">
              Developer Mode
            </h3>
            <p class="text-[10px] text-makoclaw-text-secondary/50 mt-0.5">
              Full programming agent capabilities
            </p>
          </div>
        </div>
        <label class="relative inline-flex items-center cursor-pointer">
          <input
            v-model="developerMode"
            type="checkbox"
            class="sr-only peer"
            @change="saveDeveloperMode"
          >
          <div class="w-12 h-7 rounded-full bg-makoclaw-border/70 peer-checked:bg-makoclaw-accent after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-6 after:w-6 after:transition-all peer-checked:after:translate-x-5"></div>
        </label>
      </div>
      <div class="space-y-3 text-xs text-makoclaw-text-secondary/70">
        <div class="flex items-center gap-2 px-4 py-3 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30">
          <ClockIcon class="w-4 h-4 text-makoclaw-accent shrink-0" />
          <span><span class="font-bold text-makoclaw-text">Extended timeout:</span> Shell commands run up to 5 minutes (vs 60s default)</span>
        </div>
        <div class="flex items-center gap-2 px-4 py-3 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30">
          <DocumentTextIcon class="w-4 h-4 text-makoclaw-accent shrink-0" />
          <span><span class="font-bold text-makoclaw-text">Extended output:</span> 50K character limit for build/test output (vs 10K default)</span>
        </div>
        <div class="flex items-center gap-2 px-4 py-3 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30">
          <FolderOpenIcon class="w-4 h-4 text-makoclaw-accent shrink-0" />
          <span><span class="font-bold text-makoclaw-text">Unrestricted workspace:</span> Access files outside workspace directory</span>
        </div>
      </div>
    </div>

    <!-- Programmer Tools -->
    <div class="glass-panel rounded-2xl p-5 border border-makoclaw-border/50 transition-all duration-300 hover:shadow-lg">
      <div class="flex items-center gap-3 mb-5">
        <div class="p-2 rounded-xl bg-blue-500/10 text-blue-400">
          <WrenchScrewdriverIcon class="w-5 h-5" />
        </div>
        <h3 class="text-xs font-medium tracking-wide text-makoclaw-text-secondary/70 uppercase">
          Programmer Tools
        </h3>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div
          v-for="tool in programmerTools"
          :key="tool.name"
          class="px-4 py-4 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30 transition-all hover:border-makoclaw-accent/20"
        >
          <div class="flex items-center gap-2 mb-2">
            <component :is="tool.icon" class="w-4 h-4 text-makoclaw-accent" />
            <span class="text-xs font-black text-makoclaw-text">{{ tool.name }}</span>
            <span
              v-if="tool.confirmation"
              class="ml-auto text-[9px] font-medium px-2 py-0.5 rounded-full bg-yellow-500/10 text-yellow-400 border border-yellow-500/20"
            >
              Semi-auto
            </span>
            <span
              v-else
              class="ml-auto text-[9px] font-medium px-2 py-0.5 rounded-full bg-green-500/10 text-green-400 border border-green-500/20"
            >
              Full-auto
            </span>
          </div>
          <p class="text-[10px] text-makoclaw-text-secondary/60 leading-relaxed">{{ tool.description }}</p>
        </div>
      </div>
    </div>

    <!-- Semi-Autonomous Behavior -->
    <div class="glass-panel rounded-2xl p-5 border border-makoclaw-border/50 transition-all duration-300 hover:shadow-lg">
      <div class="flex items-center gap-3 mb-5">
        <div class="p-2 rounded-xl bg-green-500/10 text-green-400">
          <ShieldCheckIcon class="w-5 h-5" />
        </div>
        <h3 class="text-xs font-medium tracking-wide text-makoclaw-text-secondary/70 uppercase">
          Semi-Autonomous Safety
        </h3>
      </div>
      <div class="space-y-2">
        <div class="flex items-center gap-3 px-4 py-3 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30">
          <CheckCircleIcon class="w-4 h-4 text-green-400 shrink-0" />
          <span class="text-xs text-makoclaw-text-secondary/70"><span class="font-bold text-green-400">Auto:</span> git commit, branch, checkout, stash, build, test, lint, format</span>
        </div>
        <div class="flex items-center gap-3 px-4 py-3 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30">
          <ExclamationTriangleIcon class="w-4 h-4 text-yellow-400 shrink-0" />
          <span class="text-xs text-makoclaw-text-secondary/70"><span class="font-bold text-yellow-400">Confirm:</span> git push, force push, rebase, create PR, delete branch</span>
        </div>
        <div class="flex items-center gap-3 px-4 py-3 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30">
          <ExclamationTriangleIcon class="w-4 h-4 text-yellow-400 shrink-0" />
          <span class="text-xs text-makoclaw-text-secondary/70"><span class="font-bold text-yellow-400">Confirm:</span> mobile archive, install on device, docker prune</span>
        </div>
        <div class="flex items-center gap-3 px-4 py-3 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30">
          <XCircleIcon class="w-4 h-4 text-red-400 shrink-0" />
          <span class="text-xs text-makoclaw-text-secondary/70"><span class="font-bold text-red-400">Blocked:</span> rm -rf, force push to main/master, disk format, shutdown</span>
        </div>
      </div>
    </div>

    <!-- Playwright MCP Setup -->
    <div class="glass-panel rounded-2xl p-5 border border-makoclaw-border/50 transition-all duration-300 hover:shadow-lg">
      <div class="flex items-center gap-3 mb-5">
        <div class="p-2 rounded-xl bg-purple-500/10 text-purple-400">
          <GlobeAltIcon class="w-5 h-5" />
        </div>
        <h3 class="text-xs font-medium tracking-wide text-makoclaw-text-secondary/70 uppercase">
          Browser Automation (Playwright MCP)
        </h3>
      </div>
      <p class="text-[10px] text-makoclaw-text-secondary/60 mb-4">
        Enable full browser automation (screenshots, click, fill, JS eval) by configuring the Playwright MCP server.
        Without it, only basic web fetching and E2E test detection are available.
      </p>
      <div class="px-4 py-3 rounded-2xl bg-makoclaw-bg/50 border border-makoclaw-border/30">
        <p class="text-[10px] font-mono text-makoclaw-text-secondary/70 mb-1">~/.MakoClaw/mcp/playwright/mcp.json</p>
        <pre class="text-[10px] font-mono text-makoclaw-accent leading-relaxed">{
  "command": "npx",
  "args": ["-y", "@anthropic/mcp-server-playwright"]
}</pre>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import {
  CommandLineIcon,
  ClockIcon,
  DocumentTextIcon,
  FolderOpenIcon,
  WrenchScrewdriverIcon,
  ShieldCheckIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  XCircleIcon,
  GlobeAltIcon,
  CodeBracketIcon,
  CpuChipIcon,
  DevicePhoneMobileIcon,
  CubeIcon
} from '@heroicons/vue/24/outline'

const props = defineProps({
  configData: Object,
  saving: Boolean
})
const emit = defineEmits(['save'])

const developerMode = ref(false)

watch(() => props.configData, (newVal) => {
  if (newVal?.tools) {
    developerMode.value = !!newVal.tools.developer_mode
  }
}, { immediate: true })

const saveDeveloperMode = () => {
  emit('save', { tools: { developer_mode: developerMode.value } })
}

const programmerTools = [
  {
    name: 'git',
    icon: CodeBracketIcon,
    description: 'Clone, commit, push, branch, merge, rebase, stash, tag, create PR. Native Git with safety controls.',
    confirmation: true
  },
  {
    name: 'project',
    icon: CpuChipIcon,
    description: 'Auto-detect project type, run build/test/lint/format across Go, Node, Python, Rust, Swift, Kotlin, Ruby, Java.',
    confirmation: false
  },
  {
    name: 'browser',
    icon: GlobeAltIcon,
    description: 'Web automation via Playwright MCP. Screenshots, click, fill forms, run E2E tests, evaluate JavaScript.',
    confirmation: false
  },
  {
    name: 'mobile',
    icon: DevicePhoneMobileIcon,
    description: 'iOS (xcodebuild, fastlane, simctl) and Android (gradle, adb) builds, tests, simulator/emulator.',
    confirmation: true
  },
  {
    name: 'docker',
    icon: CubeIcon,
    description: 'Build images, run containers, compose up/down, view logs, exec commands, inspect, prune.',
    confirmation: true
  }
]
</script>
