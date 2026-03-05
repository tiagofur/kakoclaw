<template>
  <ion-page>
    <ion-content color="background" class="ion-padding" :fullscreen="true">
      <div class="flex flex-col items-center justify-center min-h-full px-6">
        <!-- Logo / Branding -->
        <div class="mb-10 text-center">
          <div
            class="w-20 h-20 rounded-2xl bg-gradient-to-br from-makoclaw-accent to-makoclaw-accent-hover flex items-center justify-center mx-auto mb-4 shadow-lg"
          >
            <span class="text-3xl font-black text-white">M</span>
          </div>
          <h1 class="text-2xl font-bold text-makoclaw-text">MakoClaw</h1>
          <p class="text-sm text-makoclaw-text-secondary mt-1">
            Sign in to your server
          </p>
        </div>

        <!-- Server URL -->
        <div class="w-full max-w-sm mb-6">
          <label
            class="block text-xs font-medium text-makoclaw-text-secondary mb-1.5"
            >Server URL</label
          >
          <div
            class="flex items-center bg-makoclaw-surface border border-makoclaw-border rounded-xl px-3 min-h-[48px]"
          >
            <ion-icon
              :icon="serverOutline"
              class="text-makoclaw-text-secondary mr-2 text-lg flex-shrink-0"
            ></ion-icon>
            <input
              v-model="serverUrl"
              type="url"
              placeholder="http://192.168.1.100:8080"
              class="flex-1 bg-transparent text-sm text-makoclaw-text outline-none border-none py-3"
            />
          </div>
        </div>

        <!-- Login Form -->
        <div class="w-full max-w-sm space-y-4">
          <div>
            <label
              class="block text-xs font-medium text-makoclaw-text-secondary mb-1.5"
              >Username</label
            >
            <div
              class="flex items-center bg-makoclaw-surface border border-makoclaw-border rounded-xl px-3 min-h-[48px]"
            >
              <ion-icon
                :icon="personOutline"
                class="text-makoclaw-text-secondary mr-2 text-lg flex-shrink-0"
              ></ion-icon>
              <input
                v-model="username"
                type="text"
                placeholder="Username"
                autocapitalize="off"
                autocomplete="username"
                class="flex-1 bg-transparent text-sm text-makoclaw-text outline-none border-none py-3"
                @keyup.enter="handleLogin"
              />
            </div>
          </div>
          <div>
            <label
              class="block text-xs font-medium text-makoclaw-text-secondary mb-1.5"
              >Password</label
            >
            <div
              class="flex items-center bg-makoclaw-surface border border-makoclaw-border rounded-xl px-3 min-h-[48px]"
            >
              <ion-icon
                :icon="lockClosedOutline"
                class="text-makoclaw-text-secondary mr-2 text-lg flex-shrink-0"
              ></ion-icon>
              <input
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="Password"
                autocomplete="current-password"
                class="flex-1 bg-transparent text-sm text-makoclaw-text outline-none border-none py-3"
                @keyup.enter="handleLogin"
              />
              <button
                @click="showPassword = !showPassword"
                class="text-makoclaw-text-secondary p-1"
              >
                <ion-icon
                  :icon="showPassword ? eyeOffOutline : eyeOutline"
                  class="text-lg"
                ></ion-icon>
              </button>
            </div>
          </div>

          <!-- Error Message -->
          <div
            v-if="errorMsg"
            class="px-3 py-2 bg-red-500/10 border border-red-500/30 rounded-xl"
          >
            <p class="text-xs text-red-400">{{ errorMsg }}</p>
          </div>

          <!-- Login Button -->
          <button
            @click="handleLogin"
            :disabled="isLoading"
            class="w-full py-3.5 rounded-xl font-semibold text-sm text-white transition-all shadow-md"
            :class="
              isLoading
                ? 'bg-makoclaw-accent/50 cursor-not-allowed'
                : 'bg-makoclaw-accent hover:bg-makoclaw-accent-hover active:scale-[0.98]'
            "
          >
            <span
              v-if="isLoading"
              class="flex items-center justify-center space-x-2"
            >
              <ion-spinner name="dots" class="w-5 h-5"></ion-spinner>
              <span>Signing in...</span>
            </span>
            <span v-else>Sign In</span>
          </button>
        </div>

        <!-- Version -->
        <p class="mt-8 text-[11px] text-makoclaw-text-secondary/50">
          MakoClaw Mobile v0.1.0
        </p>
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { IonContent, IonPage, IonIcon, IonSpinner } from "@ionic/vue";
import {
  serverOutline,
  personOutline,
  lockClosedOutline,
  eyeOutline,
  eyeOffOutline,
} from "ionicons/icons";
import { useConfigStore } from "../stores/configStore";
import authService from "../services/authService";

const router = useRouter();
const configStore = useConfigStore();

const serverUrl = ref(configStore.serverUrl);
const username = ref("");
const password = ref("");
const showPassword = ref(false);
const isLoading = ref(false);
const errorMsg = ref("");

onMounted(async () => {
  await configStore.loadFromStorage();
  serverUrl.value = configStore.serverUrl;
});

async function handleLogin() {
  if (!username.value.trim() || !password.value.trim()) {
    errorMsg.value = "Username and password are required.";
    return;
  }

  errorMsg.value = "";
  isLoading.value = true;

  try {
    // Save server URL before login attempt
    await configStore.setServerUrl(serverUrl.value);

    await authService.login(username.value.trim(), password.value);
    router.replace("/home");
  } catch (error: any) {
    if (error.code === "ERR_NETWORK") {
      errorMsg.value = `Cannot connect to server at ${serverUrl.value}. Check the URL and ensure the server is running.`;
    } else if (error.response?.status === 401) {
      errorMsg.value = "Invalid username or password.";
    } else {
      errorMsg.value =
        error.response?.data?.error || error.message || "Login failed.";
    }
  } finally {
    isLoading.value = false;
  }
}
</script>
