<template>
  <div
    v-if="configStore.securityAlert"
    class="bg-amber-500/10 border-b-2 border-amber-500 animate-slideDown sticky top-0 z-banner backdrop-blur-sm px-6 py-4 shadow-sm"
  >
    <div class="flex items-center gap-4 max-w-[1200px] mx-auto flex-wrap md:flex-nowrap">
      <div class="text-[32px] md:text-2xl shrink-0">
        🔒
      </div>
      <div class="flex-1">
        <h3 class="text-makoclaw-text font-semibold text-lg md:text-base mb-1">
          Alerta de Seguridad
        </h3>
        <p class="text-makoclaw-text-secondary text-sm">
          Una o más skills instaladas fueron desactivadas por razones de seguridad.
          Revisá tus skills y desinstalá las marcadas.
        </p>
      </div>
      <button
        class="bg-amber-500 text-white hover:bg-amber-400 rounded-lg px-5 py-2.5 font-semibold transition-all active:scale-[0.97] whitespace-nowrap w-full md:w-auto"
        @click="router.push('/skills')"
      >
        Revisar Skills
      </button>
    </div>
  </div>
</template>

<script setup>
import { useConfigStore } from '../stores/configStore'
import { useRouter } from 'vue-router'

const configStore = useConfigStore()
const router = useRouter()
</script>
