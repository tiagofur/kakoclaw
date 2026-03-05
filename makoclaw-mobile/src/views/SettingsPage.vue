<template>
  <ion-page>
    <ion-header class="ion-no-border">
      <ion-toolbar color="step-50">
        <ion-buttons slot="start">
          <ion-back-button
            default-href="/home"
            color="primary"
          ></ion-back-button>
        </ion-buttons>
        <ion-title class="text-makoclaw-text font-bold">Settings</ion-title>
      </ion-toolbar>
    </ion-header>

    <ion-content color="background" class="custom-scrollbar">
      <div class="p-4 space-y-6">
        <!-- Account Section -->
        <section>
          <h2
            class="text-xs font-semibold text-makoclaw-text-secondary uppercase tracking-wider mb-3 px-1"
          >
            Account
          </h2>
          <div class="card space-y-0 divide-y divide-makoclaw-border/50">
            <div class="flex items-center justify-between py-3 px-1">
              <div class="flex items-center space-x-3">
                <div
                  class="w-10 h-10 rounded-full bg-makoclaw-accent/20 flex items-center justify-center"
                >
                  <span class="text-makoclaw-accent font-bold text-sm">
                    {{ user?.username?.charAt(0).toUpperCase() || "?" }}
                  </span>
                </div>
                <div>
                  <p class="text-sm font-semibold text-makoclaw-text">
                    {{ user?.username || "Not signed in" }}
                  </p>
                  <p class="text-xs text-makoclaw-text-secondary">
                    {{ user?.role || "" }}
                  </p>
                </div>
              </div>
              <button
                @click="handleLogout"
                class="px-3 py-1.5 text-xs font-medium text-red-400 bg-red-400/10 rounded-lg border border-red-400/20 hover:bg-red-400/20 transition-colors"
              >
                Sign Out
              </button>
            </div>
          </div>
        </section>

        <!-- Server Section -->
        <section>
          <h2
            class="text-xs font-semibold text-makoclaw-text-secondary uppercase tracking-wider mb-3 px-1"
          >
            Server
          </h2>
          <div class="card space-y-3">
            <div>
              <label
                class="block text-xs font-medium text-makoclaw-text-secondary mb-1.5"
                >Server URL</label
              >
              <div
                class="flex items-center bg-makoclaw-bg border border-makoclaw-border rounded-xl px-3 min-h-[44px]"
              >
                <ion-icon
                  :icon="serverOutline"
                  class="text-makoclaw-text-secondary mr-2 text-lg flex-shrink-0"
                ></ion-icon>
                <input
                  v-model="editServerUrl"
                  type="url"
                  placeholder="http://192.168.1.100:8080"
                  class="flex-1 bg-transparent text-sm text-makoclaw-text outline-none border-none py-2.5"
                />
              </div>
            </div>
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-2">
                <div
                  class="w-2 h-2 rounded-full"
                  :class="
                    serverStatus === 'connected'
                      ? 'bg-green-400'
                      : serverStatus === 'checking'
                        ? 'bg-yellow-400 animate-pulse'
                        : 'bg-red-400'
                  "
                ></div>
                <span class="text-xs text-makoclaw-text-secondary">
                  {{
                    serverStatus === "connected"
                      ? "Connected"
                      : serverStatus === "checking"
                        ? "Checking..."
                        : "Disconnected"
                  }}
                </span>
              </div>
              <button
                @click="saveAndTestServer"
                class="px-3 py-1.5 text-xs font-medium text-makoclaw-accent bg-makoclaw-accent/10 rounded-lg border border-makoclaw-accent/20 hover:bg-makoclaw-accent/20 transition-colors"
              >
                Save & Test
              </button>
            </div>
          </div>
        </section>

        <!-- Appearance Section -->
        <section>
          <h2
            class="text-xs font-semibold text-makoclaw-text-secondary uppercase tracking-wider mb-3 px-1"
          >
            Appearance
          </h2>
          <div class="card">
            <div class="flex items-center justify-between">
              <span class="text-sm text-makoclaw-text">Theme</span>
              <ion-segment
                :value="theme"
                @ionChange="handleThemeChange"
                class="max-w-[200px]"
                color="primary"
              >
                <ion-segment-button value="system">
                  <ion-label class="text-[11px]">System</ion-label>
                </ion-segment-button>
                <ion-segment-button value="dark">
                  <ion-label class="text-[11px]">Dark</ion-label>
                </ion-segment-button>
                <ion-segment-button value="light">
                  <ion-label class="text-[11px]">Light</ion-label>
                </ion-segment-button>
              </ion-segment>
            </div>
          </div>
        </section>

        <!-- Notifications Section -->
        <section>
          <h2
            class="text-xs font-semibold text-makoclaw-text-secondary uppercase tracking-wider mb-3 px-1"
          >
            Notifications
          </h2>
          <div class="card">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-makoclaw-text">Push Notifications</p>
                <p class="text-xs text-makoclaw-text-secondary">
                  Receive alerts for responses & tasks
                </p>
              </div>
              <ion-toggle
                :checked="notificationsEnabled"
                @ionChange="handleNotificationsToggle"
                color="primary"
              ></ion-toggle>
            </div>
          </div>
        </section>

        <!-- About Section -->
        <section>
          <h2
            class="text-xs font-semibold text-makoclaw-text-secondary uppercase tracking-wider mb-3 px-1"
          >
            About
          </h2>
          <div class="card space-y-0 divide-y divide-makoclaw-border/50">
            <div class="flex items-center justify-between py-2.5 px-1">
              <span class="text-sm text-makoclaw-text-secondary">Version</span>
              <span class="text-sm text-makoclaw-text">0.1.0</span>
            </div>
            <div class="flex items-center justify-between py-2.5 px-1">
              <span class="text-sm text-makoclaw-text-secondary"
                >Providers</span
              >
              <span class="text-sm text-makoclaw-text">{{
                providerCount
              }}</span>
            </div>
          </div>
        </section>
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import {
  IonPage,
  IonHeader,
  IonToolbar,
  IonTitle,
  IonContent,
  IonButtons,
  IonBackButton,
  IonIcon,
  IonSegment,
  IonSegmentButton,
  IonLabel,
  IonToggle,
} from "@ionic/vue";
import { serverOutline } from "ionicons/icons";
import { useConfigStore } from "../stores/configStore";
import { useAuthStore } from "../stores/authStore";

const router = useRouter();
const configStore = useConfigStore();
const authStore = useAuthStore();

const editServerUrl = ref(configStore.serverUrl);
const serverStatus = ref<"connected" | "disconnected" | "checking">(
  "disconnected",
);
const theme = ref(configStore.theme);
const notificationsEnabled = ref(configStore.notificationsEnabled);

const user = computed(() => authStore.user);
const providerCount = computed(() => configStore.activeProviders.length);

onMounted(async () => {
  editServerUrl.value = configStore.serverUrl;
  // Check status on load
  await testServer();
});

async function saveAndTestServer() {
  await configStore.setServerUrl(editServerUrl.value);
  await testServer();
}

async function testServer() {
  serverStatus.value = "checking";
  try {
    await configStore.checkStatus();
    serverStatus.value = "connected";
  } catch {
    serverStatus.value = "disconnected";
  }
}

function handleThemeChange(event: any) {
  const newTheme = event.detail.value as "dark" | "light" | "system";
  theme.value = newTheme;
  configStore.setTheme(newTheme);
}

function handleNotificationsToggle(event: any) {
  const enabled = event.detail.checked;
  notificationsEnabled.value = enabled;
  configStore.setNotificationsEnabled(enabled);
}

async function handleLogout() {
  await authStore.logout();
  router.replace("/login");
}
</script>
