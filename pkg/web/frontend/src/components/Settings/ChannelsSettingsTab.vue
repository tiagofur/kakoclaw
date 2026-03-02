<template>
  <div class="space-y-5 max-w-4xl mx-auto animate-fade-in-up pb-10">
    <!-- Section Header Decor -->
    <div class="flex items-center gap-4 mb-2 opacity-50">
      <div class="h-[1px] flex-1 bg-gradient-to-r from-transparent to-makoclaw-border" />
      <span class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary">
        Channels
      </span>
      <div class="h-[1px] flex-1 bg-gradient-to-l from-transparent to-makoclaw-border" />
    </div>

    <!-- Available Network Channels Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div
        v-for="channel in availableChannels"
        :key="channel.id"
        class="glass-panel rounded-2xl p-5 transition-all duration-300 hover:shadow-xl hover:shadow-makoclaw-accent/5 hover:-translate-y-1 group relative overflow-hidden flex flex-col h-full border border-makoclaw-border/50 hover:border-makoclaw-accent/40"
      >
        <!-- Background Glow -->
        <div
          class="absolute -top-12 -right-12 w-32 h-32 blur-[50px] rounded-full transition-all duration-700 opacity-0 group-hover:opacity-20"
          :class="channels[channel.id]?.enabled ? 'bg-makoclaw-accent' : 'bg-makoclaw-text-secondary/50'"
        />

        <div class="flex items-center justify-between mb-8 relative z-10">
          <div class="flex items-center gap-3">
            <div
              class="w-10 h-10 rounded-xl flex items-center justify-center bg-makoclaw-surface border border-makoclaw-border/50 text-makoclaw-text-secondary transition-all duration-300 group-hover:scale-105 group-hover:rotate-3 shadow-md text-xl"
              :class="{'!bg-makoclaw-accent !border-makoclaw-accent/20 !text-white shadow-makoclaw-accent/30': channels[channel.id]?.enabled}"
            >
              {{ channel.icon }}
            </div>
            <div>
              <h3 class="font-semibold text-base text-makoclaw-text tracking-tight">
                {{ channel.name }}
              </h3>
              <span class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/50">
                Uplink Channel
              </span>
            </div>
          </div>

          <!-- Modern Toggle -->
          <button
            class="relative inline-flex h-7 w-12 items-center rounded-full transition-all duration-500 focus:outline-none"
            :class="channels[channel.id]?.enabled ? 'bg-makoclaw-accent' : 'bg-makoclaw-bg/80 border-2 border-makoclaw-border'"
            @click="$emit('toggle', channel.id)"
          >
            <span
              class="inline-block h-5 w-5 rounded-full bg-white shadow-lg transform transition-transform duration-500"
              :class="channels[channel.id]?.enabled ? 'translate-x-[22px]' : 'translate-x-[2px]'"
            />
          </button>
        </div>

        <div class="mt-auto flex items-center gap-3 relative z-10">
          <button
            class="flex-1 py-3 text-xs font-black uppercase tracking-widest bg-makoclaw-surface border border-makoclaw-border/10 rounded-xl hover:border-makoclaw-accent/40 hover:bg-makoclaw-surface transition-all flex items-center justify-center text-makoclaw-text group/btn active:scale-95 shadow-lg"
            @click="$emit('config', channel)"
          >
            <AdjustmentsHorizontalIcon class="h-4 w-4 mr-3 transition-transform group-hover/btn:rotate-180 duration-1000" />
            Tune Channel
          </button>
          
          <div
            v-if="channels[channel.id]?.enabled"
            class="w-12 h-12 rounded-xl bg-makoclaw-accent/10 border border-makoclaw-accent/20 flex items-center justify-center text-makoclaw-accent shadow-inner shrink-0"
            title="Active Link"
          >
            <div class="relative">
              <BoltIcon class="w-5 h-5 animate-pulse" />
              <span class="absolute -top-1 -right-1 w-2.5 h-2.5 bg-makoclaw-accent rounded-full animate-ping" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Email Communication / Mailbox Agent -->
    <div class="glass-panel p-6 rounded-3xl mt-8 border border-makoclaw-border/50 transition-all hover:border-cyan-500/30 relative overflow-hidden group">
      <!-- Decor -->
      <div class="absolute top-0 right-0 w-64 h-64 bg-cyan-500/5 rounded-full blur-[80px] -translate-y-1/2 translate-x-1/2 group-hover:bg-cyan-500/10 transition-colors duration-700" />

      <div class="flex items-center justify-between mb-8 relative z-10 flex-wrap gap-4">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-2xl bg-cyan-500/10 text-cyan-400 flex items-center justify-center border border-cyan-500/20 shadow-inner">
            <EnvelopeIcon class="w-6 h-6" />
          </div>
          <div>
            <h3 class="text-lg font-black tracking-tight text-makoclaw-text uppercase italic">
              Transmission Mailbox
            </h3>
            <p class="text-[10px] font-medium uppercase tracking-widest text-makoclaw-text-secondary/60">
              Automated Bot Dispatch & Reports
            </p>
          </div>
        </div>
        <button
          :disabled="saving"
          class="px-5 py-2.5 rounded-xl bg-cyan-600 hover:bg-cyan-500 text-white text-xs font-bold uppercase tracking-widest transition-all active:scale-95 shadow-lg shadow-cyan-500/20 flex items-center gap-2 group/save disabled:opacity-50"
          @click="$emit('save', { tools: configData?.tools })"
        >
          <ArrowPathIcon
            v-if="saving"
            class="w-4 h-4 animate-spin"
          />
          <PaperAirplaneIcon
            v-else
            class="w-4 h-4 group-hover/save:translate-x-1 group-hover/save:-translate-y-1 transition-transform"
          />
          {{ saving ? 'Syncing...' : 'Sync Mailbox' }}
        </button>
      </div>

      <div
        v-if="configData?.tools?.email"
        class="grid grid-cols-1 md:grid-cols-2 gap-6 relative z-10"
      >
        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary/80 flex items-center gap-2">
            <UserIcon class="w-3.5 h-3.5 text-cyan-400" /> Bot Dispatch Address (From)
          </label>
          <input
            v-model="configData.tools.email.from"
            type="email"
            placeholder="bot@makoclaw.local"
            class="w-full bg-makoclaw-bg/60 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-medium text-makoclaw-text focus:border-cyan-500 transition-colors outline-none hover:bg-makoclaw-bg/80"
          >
        </div>
        
        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary/80 flex items-center gap-2">
            <UsersIcon class="w-3.5 h-3.5 text-cyan-400" /> Default Recipient (To)
          </label>
          <input
            v-model="configData.tools.email.to"
            type="email"
            placeholder="admin@domain.com"
            class="w-full bg-makoclaw-bg/60 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-medium text-makoclaw-text focus:border-cyan-500 transition-colors outline-none hover:bg-makoclaw-bg/80"
          >
        </div>

        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary/80">SMTP Server Host</label>
          <input
            v-model="configData.tools.email.host"
            type="text"
            placeholder="smtp.gmail.com"
            class="w-full bg-makoclaw-bg/60 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-medium text-makoclaw-text focus:border-cyan-500 transition-colors outline-none hover:bg-makoclaw-bg/80"
          >
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <label class="text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary/80">Port</label>
            <input
              v-model.number="configData.tools.email.port"
              type="number"
              placeholder="587"
              class="w-full bg-makoclaw-bg/60 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-medium text-makoclaw-text focus:border-cyan-500 transition-colors outline-none hover:bg-makoclaw-bg/80"
            >
          </div>
          <div class="space-y-2 flex flex-col justify-end pb-3">
            <label class="flex items-center gap-3 cursor-pointer group/toggle">
              <div
                class="relative w-12 h-6 rounded-full transition-colors duration-300"
                :class="configData.tools.email.enabled ? 'bg-cyan-500' : 'bg-makoclaw-border/80 border border-makoclaw-border'"
              >
                <div
                  class="absolute inset-y-1 w-4 h-4 rounded-full bg-white transition-transform duration-300 shadow-sm"
                  :class="configData.tools.email.enabled ? 'left-1 translate-x-6' : 'left-1 translate-x-0'"
                />
              </div>
              <span
                class="text-[10px] font-black uppercase tracking-widest transition-colors"
                :class="configData.tools.email.enabled ? 'text-cyan-400' : 'text-makoclaw-text-secondary/50'"
              >Enable Link</span>
              <input
                v-model="configData.tools.email.enabled"
                type="checkbox"
                class="hidden"
              >
            </label>
          </div>
        </div>

        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary/80">SMTP Username</label>
          <input
            v-model="configData.tools.email.username"
            type="text"
            placeholder="Auth Username"
            class="w-full bg-makoclaw-bg/60 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-medium text-makoclaw-text focus:border-cyan-500 transition-colors outline-none hover:bg-makoclaw-bg/80"
          >
        </div>

        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary/80">SMTP Password</label>
          <input
            v-model="configData.tools.email.password"
            type="password"
            placeholder="••••••••••••"
            class="w-full bg-makoclaw-bg/60 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-black tracking-widest text-makoclaw-text focus:border-cyan-500 transition-colors outline-none hover:bg-makoclaw-bg/80"
          >
        </div>
      </div>
    </div>

    <!-- Security Note -->
    <div
      class="glass-panel p-4 rounded-2xl bg-amber-500/5 border border-amber-500/10 flex items-center gap-4 mt-6 animate-fade-in-up"
      style="animation-delay: 0.2s"
    >
      <div class="p-2.5 rounded-xl bg-amber-500/10 text-amber-500 shadow-lg shadow-amber-500/5 shrink-0">
        <ShieldCheckIcon class="w-5 h-5" />
      </div>
      <div>
        <h4 class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text mb-1">
          Protocol Warning
        </h4>
        <p class="text-xs font-medium text-makoclaw-text-secondary/50 leading-relaxed">
          Changing neural channel configurations may temporarily sever active uplinks. Ensure backup encryption is enabled for sensitive transmissions.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { watch } from 'vue'
import { 
  AdjustmentsHorizontalIcon, 
  BoltIcon, 
  ShieldCheckIcon,
  EnvelopeIcon,
  PaperAirplaneIcon,
  ArrowPathIcon,
  UserIcon,
  UsersIcon
} from '@heroicons/vue/24/outline'

const props = defineProps({
  channels: { type: Object, default: () => ({}) },
  configData: { type: Object, default: () => ({}) },
  saving: { type: Boolean, default: false }
})

const availableChannels = [
  { id: 'telegram', name: 'Telegram', icon: '🤖' },
  { id: 'discord', name: 'Discord', icon: '💬' },
  { id: 'slack', name: 'Slack', icon: '💼' },
  { id: 'whatsapp', name: 'WhatsApp', icon: '📞' },
  { id: 'feishu', name: 'Feishu / Lark', icon: '🦅' },
  { id: 'dingtalk', name: 'DingTalk', icon: '🔔' },
  { id: 'qq', name: 'Tencent QQ', icon: '🐧' },
  { id: 'signal', name: 'Signal', icon: '🔒' },
  { id: 'maixcam', name: 'MaixCam Vision', icon: '👁️' }
]

const ensureEmailConfig = () => {
  if (props.configData && !props.configData.tools) {
    props.configData.tools = {}
  }
  if (props.configData && props.configData.tools && !props.configData.tools.email) {
    props.configData.tools.email = {
      enabled: false, host: '', port: 587, username: '', password: '', from: '', to: ''
    }
  }
}

watch(() => props.configData, ensureEmailConfig, { deep: true, immediate: true })

defineEmits(['toggle', 'config', 'save'])
</script>

<style scoped>
@keyframes fade-in-up {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
.animate-fade-in-up {
  animation: fade-in-up 0.5s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>
