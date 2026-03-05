<template>
  <ion-page>
    <ion-header class="ion-no-border">
      <ion-toolbar color="step-50">
        <ion-buttons slot="start">
          <ion-button router-link="/home">
            <ion-icon :icon="chevronBackOutline" color="primary"></ion-icon>
          </ion-button>
        </ion-buttons>
        <ion-title class="text-makoclaw-text font-bold">Tasks</ion-title>
        <ion-buttons slot="end">
          <ion-button>
            <ion-icon :icon="addOutline" color="primary"></ion-icon>
          </ion-button>
        </ion-buttons>
      </ion-toolbar>

      <!-- Segment control for Kanban columns -->
      <ion-toolbar color="step-50">
        <ion-segment v-model="selectedSegment" class="px-2" color="primary">
          <ion-segment-button value="todo">
            <ion-label class="text-xs"
              >To Do <span class="badge-neutral ml-1">3</span></ion-label
            >
          </ion-segment-button>
          <ion-segment-button value="inprogress">
            <ion-label class="text-xs"
              >In Progress <span class="badge-neutral ml-1">1</span></ion-label
            >
          </ion-segment-button>
          <ion-segment-button value="done">
            <ion-label class="text-xs"
              >Done <span class="badge-neutral ml-1">5</span></ion-label
            >
          </ion-segment-button>
        </ion-segment>
      </ion-toolbar>
    </ion-header>

    <ion-content color="background" class="custom-scrollbar">
      <div class="p-4 space-y-3">
        <div v-if="selectedSegment === 'todo'" class="space-y-3">
          <!-- Task Card -->
          <div class="card card-interactive list-item-interactive">
            <div class="flex justify-between items-start mb-2">
              <h3 class="text-sm font-semibold text-makoclaw-text">
                Optimize Mobile Layout
              </h3>
              <span class="badge-warning">High</span>
            </div>
            <p class="text-xs text-makoclaw-text-secondary mb-3">
              Ensure all Vue components look perfect on smaller screens.
            </p>
            <div
              class="flex items-center justify-between text-[11px] text-makoclaw-text-secondary"
            >
              <div class="flex items-center space-x-1">
                <ion-icon :icon="calendarOutline"></ion-icon>
                <span>Today</span>
              </div>
              <div
                class="w-6 h-6 rounded-full bg-makoclaw-surface-hover flex items-center justify-center border border-makoclaw-border"
              >
                <ion-icon :icon="personOutline" class="text-xs"></ion-icon>
              </div>
            </div>
          </div>

          <!-- Task Card -->
          <div class="card card-interactive list-item-interactive">
            <div class="flex justify-between items-start mb-2">
              <h3 class="text-sm font-semibold text-makoclaw-text">
                Setup capacitor config
              </h3>
              <span class="badge-success">Low</span>
            </div>
            <p class="text-xs text-makoclaw-text-secondary mb-3">
              Configure deep links and native permissions.
            </p>
          </div>
        </div>

        <div v-else-if="selectedSegment === 'inprogress'" class="space-y-3">
          <div class="empty-state">
            <ion-icon
              :icon="folderOpenOutline"
              class="empty-state-icon"
            ></ion-icon>
            <h3 class="empty-state-title">No Active Tasks</h3>
            <p class="empty-state-text">Drag tasks here or start working.</p>
          </div>
        </div>
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { ref } from "vue";
import {
  IonContent,
  IonHeader,
  IonPage,
  IonTitle,
  IonToolbar,
  IonButtons,
  IonButton,
  IonIcon,
  IonSegment,
  IonSegmentButton,
  IonLabel,
} from "@ionic/vue";
import {
  chevronBackOutline,
  addOutline,
  calendarOutline,
  personOutline,
  folderOpenOutline,
} from "ionicons/icons";

const selectedSegment = ref("todo");
</script>
