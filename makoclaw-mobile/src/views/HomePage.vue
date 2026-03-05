<template>
  <ion-page>
    <ion-header class="ion-no-border">
      <ion-toolbar color="step-50">
        <ion-buttons slot="start">
          <ion-button>
            <ion-icon :icon="menuOutline" color="primary"></ion-icon>
          </ion-button>
        </ion-buttons>
        <div class="flex items-center justify-center w-full">
          <!-- Placeholder for dynamic agent header -->
          <div class="flex items-center space-x-2">
            <div
              class="w-8 h-8 rounded-full bg-makoclaw-accent/20 flex items-center justify-center"
            >
              <span class="text-makoclaw-accent font-bold text-sm">O</span>
            </div>
            <div class="flex flex-col">
              <span class="text-sm font-bold text-makoclaw-text leading-tight"
                >Orchestrator</span
              >
              <span
                class="text-[10px] text-makoclaw-text-secondary leading-tight"
                >Active</span
              >
            </div>
          </div>
        </div>
        <ion-buttons slot="end">
          <ion-button>
            <ion-icon :icon="peopleOutline" color="primary"></ion-icon>
          </ion-button>
        </ion-buttons>
      </ion-toolbar>
    </ion-header>

    <ion-content color="background" class="custom-scrollbar">
      <div class="flex flex-col p-4 space-y-4 pb-24">
        <!-- Welcome Message -->
        <div class="flex justify-start">
          <div
            class="max-w-[85%] bg-makoclaw-surface border border-makoclaw-border rounded-2xl rounded-tl-sm p-3 shadow-sm"
          >
            <p class="text-sm text-makoclaw-text">
              Hello! I'm your MakoClaw mobile assistant. How can I help you
              today?
            </p>
          </div>
        </div>

        <!-- System Message -->
        <div class="flex justify-center w-full my-2">
          <div
            class="px-3 py-1 bg-makoclaw-surface-hover rounded-full border border-makoclaw-border/50 shadow-sm"
          >
            <span class="text-xs text-makoclaw-text-secondary"
              >Orchestrator assigned to task</span
            >
          </div>
        </div>

        <!-- User Message -->
        <div class="flex justify-end">
          <div
            class="max-w-[85%] bg-makoclaw-accent text-white rounded-2xl rounded-tr-sm p-3 shadow-sm"
          >
            <p class="text-sm">
              Can you search my Kanban board for pending tasks?
            </p>
          </div>
        </div>
      </div>
    </ion-content>

    <!-- Chat Input Area -->
    <div
      class="absolute bottom-0 left-0 right-0 bg-makoclaw-surface border-t border-makoclaw-border p-3 safe-bottom z-10 glass-sticky"
    >
      <div class="flex items-end space-x-2">
        <button
          class="p-2.5 rounded-full text-makoclaw-text-secondary hover:text-makoclaw-accent hover:bg-makoclaw-accent/10 transition-colors flex-shrink-0"
        >
          <ion-icon :icon="addOutline" class="text-xl"></ion-icon>
        </button>
        <div
          class="flex-1 bg-makoclaw-bg border border-makoclaw-border rounded-xl relative flex items-center min-h-[44px]"
        >
          <textarea
            class="w-full bg-transparent border-none focus:ring-0 text-sm text-makoclaw-text px-3 py-2.5 resize-none max-h-32 custom-scrollbar outline-none"
            placeholder="Message MakoClaw..."
            rows="1"
          ></textarea>
        </div>
        <button
          class="p-2.5 rounded-full bg-makoclaw-accent text-white hover:bg-makoclaw-accent-hover transition-colors shadow-sm flex-shrink-0"
        >
          <ion-icon :icon="sendOutline" class="text-xl"></ion-icon>
        </button>
      </div>
    </div>
  </ion-page>
</template>

<script setup lang="ts">
import {
  IonContent,
  IonHeader,
  IonPage,
  IonToolbar,
  IonButtons,
  IonButton,
  IonIcon,
} from "@ionic/vue";
import {
  menuOutline,
  peopleOutline,
  addOutline,
  sendOutline,
} from "ionicons/icons";
</script>

<style scoped>
/* Ensure content doesn't hide behind absolute input */
ion-content::part(scroll) {
  padding-bottom: 80px;
}
</style>
