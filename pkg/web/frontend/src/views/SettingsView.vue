<template>
  <div class="h-full flex flex-col bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-makoclaw-accent/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-15 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-blue-500/20 via-transparent to-transparent" />
    </div>

    <!-- Mobile Header -->
    <div class="lg:hidden glass-sticky top-0 z-30 border-b border-makoclaw-border/20">
      <div class="px-4 pt-4 pb-3">
        <div class="flex items-center gap-3 mb-3">
          <!-- Icon Container -->
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-makoclaw-accent/20 to-blue-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-makoclaw-accent/10">
            <svg
              class="w-5 h-5 text-makoclaw-accent"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
              />
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
              />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <h1 class="text-xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-makoclaw-accent bg-clip-text text-transparent">
              Settings
            </h1>
            <p class="text-xs text-makoclaw-text-secondary mt-0.5">
              Configure your workspace
            </p>
          </div>
          <button
            class="p-2.5 min-h-[40px] min-w-[40px] rounded-xl bg-makoclaw-surface/50 border border-makoclaw-border/50 hover:bg-makoclaw-surface-hover transition-all flex items-center justify-center"
            @click="loadData"
          >
            <ArrowPathIcon class="w-4 h-4 text-makoclaw-text-secondary" />
          </button>
        </div>

        <!-- Mobile Tabs (Horizontal Scroll) -->
        <div class="flex gap-2 overflow-x-auto pb-1 scrollbar-none snap-x -mx-1 px-1">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="flex-none px-3.5 py-2 rounded-xl text-xs font-medium transition-all snap-start min-h-[36px]"
            :class="[activeTab === tab.key
              ? 'bg-gradient-to-r from-makoclaw-accent to-makoclaw-accent/80 text-white shadow-lg shadow-makoclaw-accent/25'
              : 'bg-makoclaw-surface/40 text-makoclaw-text-secondary border border-makoclaw-border/30 hover:border-makoclaw-accent/30 hover:text-makoclaw-text backdrop-blur-sm']"
            @click="activeTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </div>
      </div>
    </div>

    <div class="flex-1 flex overflow-hidden relative">
      <!-- Desktop Sidebar Navigation -->
      <aside class="hidden lg:flex flex-col w-64 flex-none border-r border-makoclaw-border/20 bg-makoclaw-surface/20 backdrop-blur-sm z-20">
        <!-- Sidebar Header -->
        <div class="p-5 border-b border-makoclaw-border/20">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-makoclaw-accent/20 to-blue-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-makoclaw-accent/10">
              <svg
                class="w-5 h-5 text-makoclaw-accent"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                />
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                />
              </svg>
            </div>
            <div>
              <h2 class="text-lg font-bold text-makoclaw-text">
                Settings
              </h2>
              <p class="text-xs text-makoclaw-text-secondary">
                Configure your workspace
              </p>
            </div>
          </div>
        </div>

        <!-- Navigation -->
        <nav class="flex-1 p-3 space-y-1 overflow-y-auto custom-scrollbar">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-200 group text-left"
            :class="activeTab === tab.key
              ? 'bg-gradient-to-r from-makoclaw-accent/15 to-transparent text-makoclaw-text border-l-2 border-makoclaw-accent'
              : 'text-makoclaw-text-secondary hover:bg-makoclaw-surface/50 hover:text-makoclaw-text border-l-2 border-transparent'"
            @click="activeTab = tab.key"
          >
            <div
              class="w-8 h-8 rounded-lg flex items-center justify-center transition-all"
              :class="activeTab === tab.key ? 'bg-makoclaw-accent/20' : 'bg-makoclaw-bg/50'"
            >
              <component
                :is="getTabIcon(tab.key)"
                class="w-4 h-4"
                :class="activeTab === tab.key ? 'text-makoclaw-accent' : 'text-makoclaw-text-secondary group-hover:text-makoclaw-text'"
              />
            </div>
            <span class="text-sm font-medium">{{ tab.label }}</span>
          </button>
        </nav>
      </aside>

      <!-- Main Content Area -->
      <main class="flex-1 min-w-0 overflow-y-auto custom-scrollbar relative">
        <div class="max-w-4xl mx-auto p-4 sm:p-6 lg:p-8">
          <!-- Content Header -->
          <header class="mb-8">
            <div class="flex items-center gap-4">
              <!-- Icon Container -->
              <div class="w-12 h-12 sm:w-14 sm:h-14 rounded-xl bg-gradient-to-br from-makoclaw-accent/20 to-blue-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-makoclaw-accent/10">
                <component
                  :is="activeTabIcon"
                  class="w-6 h-6 sm:w-7 sm:h-7 text-makoclaw-accent"
                />
              </div>

              <!-- Title + Description -->
              <div class="flex-1 min-w-0">
                <h2 class="text-2xl sm:text-3xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-makoclaw-accent bg-clip-text text-transparent">
                  {{ activeTabLabel }}
                </h2>
                <p class="text-sm text-makoclaw-text-secondary mt-1">
                  {{ activeTabDescription }}
                </p>
              </div>
            </div>
          </header>

          <!-- Tab Content -->
          <div class="relative min-h-[400px]">
            <transition
              enter-active-class="transition duration-300 ease-out"
              enter-from-class="opacity-0 translate-y-4"
              enter-to-class="opacity-100 translate-y-0"
              leave-active-class="transition duration-200 ease-in"
              leave-from-class="opacity-100 translate-y-0"
              leave-to-class="opacity-0 -translate-y-4"
              mode="out-in"
            >
              <component
                :is="activeTabComponent"
                :key="activeTab"
                :agents="configData?.agents || { defaults: {}, orchestrator: {}, specialists: {} }"
                :providers="configData?.providers || {}"
                :channels="configData?.channels || {}"
                :config-data="configData"
                :providers-list="providersList"
                :users-list="usersList"
                :saving="saving"
                @save="saveConfig"
                @toggle="toggleChannel"
                @refresh-config="fetchUserConfig"
                @config="activeTab === 'social_media' ? openSocialConfig($event) : openChannelConfig($event)"
                @delete="deleteSocialAccount"
                @add-user="openUserModal()"
                @edit-user="openUserModal"
                @block-user="openBlockModal"
                @unblock-user="openUnblockModal"
                @delete-user="deleteUserLocal"
              />
            </transition>
          </div>
        </div>
      </main>
    </div>

    <!-- Modals -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="showChannelModal || showSocialModal || showUserModal || showBlockModal || showUnblockModal"
          class="fixed inset-0 z-modal flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
          @click.self="closeAllModals"
        >
          <!-- Channel Config Modal -->
          <div
            v-if="showChannelModal"
            class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl shadow-2xl ring-1 ring-white/10 w-full max-w-lg max-h-[90vh] flex flex-col animate-scaleIn"
          >
            <div class="p-5 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent shrink-0">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-makoclaw-accent/20 to-blue-500/20 flex items-center justify-center ring-1 ring-white/10 text-2xl">
                    {{ selectedChannel?.icon }}
                  </div>
                  <div>
                    <h3 class="text-lg font-bold text-makoclaw-text">
                      Configure {{ selectedChannel?.name }}
                    </h3>
                    <p class="text-xs text-makoclaw-text-secondary">
                      Set up your channel credentials
                    </p>
                  </div>
                </div>
                <button
                  class="p-2 rounded-lg hover:bg-makoclaw-surface-hover transition-colors"
                  @click="showChannelModal = false"
                >
                  <XMarkIcon class="w-5 h-5 text-makoclaw-text-secondary" />
                </button>
              </div>
            </div>

            <div class="p-5 space-y-4 overflow-y-auto min-h-0">
              <div
                v-if="['telegram', 'discord'].includes(selectedChannel?.id)"
                class="space-y-4"
              >
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                    Bot Token
                  </label>
                  <input
                    v-model="channelForm.token"
                    type="password"
                    :placeholder="channelForm._hasExistingToken ? '•••••••• (saved — leave empty to keep)' : 'Enter bot token...'"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                  <p
                    v-if="channelForm._hasExistingToken"
                    class="text-xs text-makoclaw-success/70 mt-1.5 flex items-center gap-1"
                  >
                    ✓ Token configured — leave empty to keep current token
                  </p>
                </div>
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                    Allowed Users
                  </label>
                  <input
                    v-model="channelForm.allow_from"
                    type="text"
                    placeholder="ID-1, ID-2, @username"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                  <p class="text-xs text-makoclaw-text-secondary/60 mt-1.5">
                    Comma-separated list of authorized user IDs
                  </p>
                </div>
              </div>

              <!-- Email channel form -->
              <div
                v-else-if="selectedChannel?.id === 'email'"
                class="space-y-4"
              >
                <div class="grid grid-cols-2 gap-3">
                  <div>
                    <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                      IMAP Host
                    </label>
                    <input
                      v-model="channelForm.imap_host"
                      type="text"
                      placeholder="imap.gmail.com"
                      class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                    >
                  </div>
                  <div>
                    <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                      IMAP Port
                    </label>
                    <input
                      v-model.number="channelForm.imap_port"
                      type="number"
                      placeholder="993"
                      class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                    >
                  </div>
                </div>
                <div class="grid grid-cols-2 gap-3">
                  <div>
                    <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                      SMTP Host
                    </label>
                    <input
                      v-model="channelForm.smtp_host"
                      type="text"
                      placeholder="smtp.gmail.com"
                      class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                    >
                  </div>
                  <div>
                    <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                      SMTP Port
                    </label>
                    <input
                      v-model.number="channelForm.smtp_port"
                      type="number"
                      placeholder="587"
                      class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                    >
                  </div>
                </div>
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                    Username
                  </label>
                  <input
                    v-model="channelForm.username"
                    type="text"
                    placeholder="you@gmail.com"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                </div>
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                    Password
                  </label>
                  <input
                    v-model="channelForm.password"
                    type="password"
                    :placeholder="channelForm._hasExistingToken ? '•••••••• (saved)' : 'App password...'"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                  <p
                    v-if="channelForm._hasExistingToken"
                    class="text-xs text-makoclaw-success/70 mt-1.5 flex items-center gap-1"
                  >
                    Password configured — leave empty to keep current
                  </p>
                </div>
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                    From Address
                  </label>
                  <input
                    v-model="channelForm.from"
                    type="email"
                    placeholder="MakoClaw <you@gmail.com>"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                </div>
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                    Allowed Senders (your email)
                  </label>
                  <input
                    v-model="channelForm.allow_from"
                    type="text"
                    placeholder="you@gmail.com"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                  <p class="text-xs text-makoclaw-text-secondary/60 mt-1.5">
                    Only emails from these addresses will be processed
                  </p>
                </div>
                <div class="grid grid-cols-2 gap-3">
                  <div>
                    <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                      Poll Interval (seconds)
                    </label>
                    <input
                      v-model.number="channelForm.poll_interval_seconds"
                      type="number"
                      placeholder="60"
                      class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                    >
                  </div>
                  <div class="flex items-end pb-1">
                    <label class="flex items-center gap-3 cursor-pointer">
                      <input
                        v-model="channelForm.mark_as_read"
                        type="checkbox"
                        class="w-4 h-4 rounded border-makoclaw-border accent-makoclaw-accent"
                      >
                      <span class="text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider">
                        Mark as Read
                      </span>
                    </label>
                  </div>
                </div>
              </div>

              <!-- Slack specific form -->
              <div
                v-else-if="selectedChannel?.id === 'slack'"
                class="space-y-4"
              >
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                    Bot Token (xoxb-...)
                  </label>
                  <input
                    v-model="channelForm.bot_token"
                    type="password"
                    :placeholder="channelForm._hasExistingToken ? '•••••••• (saved — leave empty to keep)' : 'xoxb-...'"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                  <p
                    v-if="channelForm._hasExistingToken"
                    class="text-xs text-makoclaw-success/70 mt-1.5 flex items-center gap-1"
                  >
                    ✓ Token configured — leave empty to keep current token
                  </p>
                </div>
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                    App Token (xapp-...)
                  </label>
                  <input
                    v-model="channelForm.app_token"
                    type="password"
                    placeholder="xapp-..."
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                </div>
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                    Allowed User IDs
                  </label>
                  <input
                    v-model="channelForm.allow_from"
                    type="text"
                    placeholder="U01234, U56789"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                  <p class="text-xs text-makoclaw-text-secondary/60 mt-1.5">
                    Comma-separated list of authorized Slack user IDs
                  </p>
                </div>
              </div>
            </div>

            <div class="p-5 border-t border-makoclaw-border/30 flex items-center justify-end gap-3 bg-makoclaw-bg/30 shrink-0">
              <button
                class="px-4 py-2.5 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors min-h-[40px]"
                @click="showChannelModal = false"
              >
                Cancel
              </button>
              <button
                class="px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-makoclaw-accent to-makoclaw-accent/80 hover:from-makoclaw-accent-hover hover:to-makoclaw-accent text-white rounded-xl text-sm font-bold shadow-lg shadow-makoclaw-accent/25 transition-all active:scale-95"
                @click="saveChannelConfig"
              >
                Save & Restart
              </button>
            </div>
          </div>

          <!-- Social Media Config Modal -->
          <div
            v-if="showSocialModal"
            class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl shadow-2xl ring-1 ring-white/10 w-full max-w-lg max-h-[90vh] flex flex-col animate-scaleIn"
          >
            <div class="p-5 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent shrink-0">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-makoclaw-accent/20 to-blue-500/20 flex items-center justify-center ring-1 ring-white/10 text-2xl">
                    {{ selectedPlatform?.icon }}
                  </div>
                  <div>
                    <h3 class="text-lg font-bold text-makoclaw-text">
                      {{ socialIsNew ? `Add ${selectedPlatform?.name} account` : `Edit ${selectedPlatform?.id}:${socialAlias}` }}
                    </h3>
                    <p class="text-xs text-makoclaw-text-secondary font-mono">
                      {{ socialIsNew ? selectedPlatform?.name : `${selectedPlatform?.id}:${socialAlias}` }}
                    </p>
                  </div>
                </div>
                <button
                  class="p-2 rounded-lg hover:bg-makoclaw-surface-hover transition-colors"
                  @click="showSocialModal = false"
                >
                  <XMarkIcon class="w-5 h-5 text-makoclaw-text-secondary" />
                </button>
              </div>
            </div>

            <div class="p-5 space-y-4 overflow-y-auto min-h-0">
              <!-- Alias field (editable only when creating a new account) -->
              <div>
                <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                  Account Alias
                  <span class="text-makoclaw-text-secondary/40 normal-case font-medium ml-1">(e.g. personal, brand)</span>
                </label>
                <input
                  v-model="socialAlias"
                  type="text"
                  :disabled="!socialIsNew"
                  placeholder="personal"
                  class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm disabled:opacity-50 disabled:cursor-not-allowed font-mono"
                >
              </div>

              <!-- Twitter/X fields -->
              <div v-if="selectedPlatform?.id === 'twitter'" class="space-y-4">
                <div v-for="field in [
                  { key: 'api_key', label: 'API Key' },
                  { key: 'api_secret', label: 'API Secret' },
                  { key: 'access_token', label: 'Access Token' },
                  { key: 'access_token_secret', label: 'Access Token Secret' }
                ]" :key="field.key">
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                    {{ field.label }}
                  </label>
                  <input
                    v-model="socialForm[field.key]"
                    type="password"
                    :placeholder="socialForm._isConfigured ? '•••••••• (leave empty to keep)' : `Enter ${field.label}...`"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                </div>
              </div>

              <!-- Bluesky fields -->
              <div v-else-if="selectedPlatform?.id === 'bluesky'" class="space-y-4">
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Handle</label>
                  <input
                    v-model="socialForm.handle"
                    type="text"
                    placeholder="user.bsky.social"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                </div>
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">App Password</label>
                  <input
                    v-model="socialForm.app_password"
                    type="password"
                    :placeholder="socialForm._isConfigured ? '•••••••• (leave empty to keep)' : 'xxxx-xxxx-xxxx-xxxx'"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                </div>
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">PDS URL <span class="text-makoclaw-text-secondary/40 normal-case font-medium">(optional)</span></label>
                  <input
                    v-model="socialForm.pds_url"
                    type="text"
                    placeholder="https://bsky.social"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                </div>
              </div>

              <!-- LinkedIn fields -->
              <div v-else-if="selectedPlatform?.id === 'linkedin'" class="space-y-4">
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Access Token</label>
                  <input
                    v-model="socialForm.access_token"
                    type="password"
                    :placeholder="socialForm._isConfigured ? '•••••••• (leave empty to keep)' : 'OAuth 2.0 Bearer token...'"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                </div>
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Author URN</label>
                  <input
                    v-model="socialForm.author_urn"
                    type="text"
                    placeholder="urn:li:person:abc123 or urn:li:organization:456"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                  <p class="text-xs text-makoclaw-text-secondary/60 mt-1.5">
                    Your LinkedIn person or organization URN
                  </p>
                </div>
              </div>

              <!-- Facebook fields -->
              <div v-else-if="selectedPlatform?.id === 'facebook'" class="space-y-4">
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Page Access Token</label>
                  <input
                    v-model="socialForm.page_access_token"
                    type="password"
                    :placeholder="socialForm._isConfigured ? '•••••••• (leave empty to keep)' : 'EAAxxxx...'"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                </div>
                <div>
                  <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Page ID</label>
                  <input
                    v-model="socialForm.page_id"
                    type="text"
                    placeholder="123456789"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                  >
                </div>
              </div>

              <!-- Configured hint -->
              <p
                v-if="socialForm._isConfigured"
                class="text-xs text-makoclaw-success/70 flex items-center gap-1"
              >
                ✓ Credentials saved — leave fields empty to keep existing values
              </p>

              <!-- Test Result -->
              <div
                v-if="socialTestResult"
                class="flex items-center gap-3 p-3 rounded-xl text-sm font-medium"
                :class="socialTestResult.ok
                  ? 'bg-makoclaw-success/10 border border-makoclaw-success/20 text-makoclaw-success'
                  : 'bg-red-500/10 border border-red-500/20 text-red-400'"
              >
                <span>{{ socialTestResult.ok ? '✅ Connection successful' : `❌ ${socialTestResult.error}` }}</span>
              </div>
            </div>

            <div class="p-5 border-t border-makoclaw-border/30 flex items-center justify-between gap-3 bg-makoclaw-bg/30 shrink-0">
              <button
                :disabled="socialTesting"
                class="px-4 py-2.5 min-h-[40px] bg-makoclaw-surface border border-makoclaw-border/50 hover:border-makoclaw-accent/40 text-makoclaw-text rounded-xl text-xs font-bold uppercase tracking-widest transition-all active:scale-95 disabled:opacity-50 flex items-center gap-2"
                @click="testSocialConnection"
              >
                <ArrowPathIcon v-if="socialTesting" class="w-4 h-4 animate-spin" />
                <span v-else>Test Connection</span>
              </button>
              <div class="flex items-center gap-3">
                <button
                  class="px-4 py-2.5 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors min-h-[40px]"
                  @click="showSocialModal = false"
                >
                  Cancel
                </button>
                <button
                  class="px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-makoclaw-accent to-makoclaw-accent/80 hover:from-makoclaw-accent-hover hover:to-makoclaw-accent text-white rounded-xl text-sm font-bold shadow-lg shadow-makoclaw-accent/25 transition-all active:scale-95"
                  @click="saveSocialConfig"
                >
                  Save
                </button>
              </div>
            </div>
          </div>

          <!-- User Modal -->
          <div
            v-if="showUserModal"
            class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl shadow-2xl ring-1 ring-white/10 w-full max-w-lg overflow-hidden animate-scaleIn"
          >
            <div class="p-5 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-makoclaw-accent/20 to-blue-500/20 flex items-center justify-center ring-1 ring-white/10">
                  <UserIcon class="w-5 h-5 text-makoclaw-accent" />
                </div>
                <div>
                  <h3 class="text-lg font-bold text-makoclaw-text">
                    {{ userForm.id ? 'Edit User' : 'Create User' }}
                  </h3>
                  <p class="text-xs text-makoclaw-text-secondary">
                    {{ userForm.id ? 'Update user credentials' : 'Add a new user account' }}
                  </p>
                </div>
              </div>
            </div>

            <div class="p-5 space-y-4">
              <div>
                <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                  Username
                </label>
                <input
                  v-model="userForm.username"
                  :disabled="!!userForm.id"
                  placeholder="Enter username..."
                  class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm disabled:opacity-50 disabled:cursor-not-allowed"
                >
              </div>
              <div>
                <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                  Password
                </label>
                <input
                  v-model="userForm.password"
                  type="password"
                  placeholder="Enter password..."
                  class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all min-h-[40px] backdrop-blur-sm"
                >
              </div>
            </div>

            <div class="p-5 border-t border-makoclaw-border/30 flex items-center justify-end gap-3 bg-makoclaw-bg/30">
              <button
                class="px-4 py-2.5 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors min-h-[40px]"
                @click="showUserModal = false"
              >
                Cancel
              </button>
              <button
                class="px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-makoclaw-accent to-makoclaw-accent/80 hover:from-makoclaw-accent-hover hover:to-makoclaw-accent text-white rounded-xl text-sm font-bold shadow-lg shadow-makoclaw-accent/25 transition-all active:scale-95"
                @click="saveUser"
              >
                {{ userForm.id ? 'Update' : 'Create' }}
              </button>
            </div>
          </div>

          <!-- Block Modal -->
          <div
            v-if="showBlockModal"
            class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl shadow-2xl ring-1 ring-white/10 w-full max-w-md overflow-hidden animate-scaleIn"
          >
            <div class="p-6 text-center">
              <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gradient-to-br from-red-500/20 to-red-600/20 flex items-center justify-center ring-1 ring-red-500/30">
                <NoSymbolIcon class="w-8 h-8 text-red-400" />
              </div>
              <h3 class="text-xl font-bold text-makoclaw-text mb-2">
                Block User?
              </h3>
              <p class="text-sm text-makoclaw-text-secondary mb-5">
                This will immediately revoke access for <span class="font-bold text-makoclaw-text">@{{ blockForm.user?.username }}</span>
              </p>
              <textarea
                v-model="blockForm.reason"
                placeholder="Reason for blocking (optional)..."
                class="w-full px-4 py-3 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-red-500/30 focus:border-red-500/50 transition-all min-h-[100px] resize-none backdrop-blur-sm mb-5"
              />
              <div class="flex gap-3">
                <button
                  class="flex-1 py-2.5 min-h-[40px] text-sm font-medium text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors rounded-xl border border-makoclaw-border/50 hover:bg-makoclaw-surface/50"
                  @click="showBlockModal = false"
                >
                  Cancel
                </button>
                <button
                  class="flex-1 py-2.5 min-h-[40px] bg-gradient-to-r from-red-500 to-red-600 hover:from-red-600 hover:to-red-700 text-white rounded-xl text-sm font-bold shadow-lg shadow-red-500/25 transition-all active:scale-95"
                  @click="blockUser"
                >
                  Block User
                </button>
              </div>
            </div>
          </div>

          <!-- Unblock Modal -->
          <div
            v-if="showUnblockModal"
            class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl shadow-2xl ring-1 ring-white/10 w-full max-w-md overflow-hidden animate-scaleIn"
          >
            <div class="p-6 text-center">
              <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gradient-to-br from-makoclaw-success/20 to-emerald-500/20 flex items-center justify-center ring-1 ring-makoclaw-success/30">
                <CheckCircleIcon class="w-8 h-8 text-makoclaw-success" />
              </div>
              <h3 class="text-xl font-bold text-makoclaw-text mb-2">
                Unblock User?
              </h3>
              <p class="text-sm text-makoclaw-text-secondary mb-5">
                Restore access for <span class="font-bold text-makoclaw-text">@{{ unblockForm.user?.username }}</span>
              </p>
              <div class="flex gap-3">
                <button
                  class="flex-1 py-2.5 min-h-[40px] text-sm font-medium text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors rounded-xl border border-makoclaw-border/50 hover:bg-makoclaw-surface/50"
                  @click="showUnblockModal = false"
                >
                  Cancel
                </button>
                <button
                  class="flex-1 py-2.5 min-h-[40px] bg-gradient-to-r from-makoclaw-success to-emerald-500 hover:from-emerald-500 hover:to-emerald-600 text-white rounded-xl text-sm font-bold shadow-lg shadow-makoclaw-success/25 transition-all active:scale-95"
                  @click="unblockUser"
                >
                  Unblock User
                </button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  UserIcon,
  ShieldCheckIcon,
  BeakerIcon,
  ChatBubbleLeftRightIcon,
  Cog6ToothIcon,
  ArrowPathIcon,
  UserGroupIcon,
  DocumentMagnifyingGlassIcon,
  GlobeAltIcon,
  SparklesIcon,
  MegaphoneIcon,
  WrenchIcon,
  XMarkIcon,
  NoSymbolIcon,
  CheckCircleIcon
} from '@heroicons/vue/24/outline'

import ProfileSettingsTab from '../components/Settings/ProfileSettingsTab.vue'
import SoulSettingsTab from '../components/Settings/SoulSettingsTab.vue'
import AgentSettingsTab from '../components/Settings/AgentSettingsTab.vue'
import ProvidersSettingsTab from '../components/Settings/ProvidersSettingsTab.vue'
import ChannelsSettingsTab from '../components/Settings/ChannelsSettingsTab.vue'
import SocialMediaSettingsTab from '../components/Settings/SocialMediaSettingsTab.vue'
import ToolPermissionsTab from '../components/Settings/ToolPermissionsTab.vue'
import AuditLogTab from '../components/Settings/AuditLogTab.vue'
import CoreSystemTab from '../components/Settings/CoreSystemTab.vue'
import AdminUsersTab from '../components/Settings/AdminUsersTab.vue'
import ToolsSettingsTab from '../components/Settings/ToolsSettingsTab.vue'

import advancedService from '../services/advancedService'
import usersService from '../services/usersService'
import { useToast } from '../composables/useToast'
import { useAuthStore } from '../stores/authStore'

const toast = useToast()
const authStore = useAuthStore()
const route = useRoute()
const loading = ref(true)
const saving = ref(false)
const configData = ref(null)
const providersList = ref([])
const activeTab = ref('profile')

// Tabs Logic
const tabs = computed(() => {
  const base = [
    { key: 'profile', label: 'Profile' },
    { key: 'soul', label: 'Soul' },
    { key: 'agents', label: 'Agents' },
    { key: 'providers', label: 'Providers' },
    { key: 'channels', label: 'Channels' },
    { key: 'social_media', label: 'Social Media' },
    { key: 'tools', label: 'Tools' }
  ]
  if (authStore.user?.role === 'admin') {
    base.push({ key: 'users', label: 'Users' })
    base.push({ key: 'permissions', label: 'Permissions' })
    base.push({ key: 'audit', label: 'Audit Log' })
  }
  base.push({ key: 'system', label: 'System' })
  return base
})

const getTabIcon = (key) => {
  const map = {
    profile: UserIcon,
    soul: SparklesIcon,
    agents: BeakerIcon,
    providers: Cog6ToothIcon,
    channels: ChatBubbleLeftRightIcon,
    social_media: MegaphoneIcon,
    tools: WrenchIcon,
    users: UserGroupIcon,
    permissions: ShieldCheckIcon,
    audit: DocumentMagnifyingGlassIcon,
    system: GlobeAltIcon
  }
  return map[key] || GlobeAltIcon
}

const activeTabLabel = computed(() => tabs.value.find(t => t.key === activeTab.value)?.label || 'Settings')

const activeTabDescription = computed(() => {
  const map = {
    profile: 'Manage your account and preferences',
    soul: 'Define your agent\'s personality, identity, and values',
    agents: 'Configure AI specialists and orchestration',
    providers: 'Connect LLM providers and models',
    channels: 'Manage communication channels',
    social_media: 'Connect social media platforms for publishing',
    tools: 'Developer mode, programmer tools, and browser automation',
    users: 'Manage registered user accounts',
    permissions: 'Configure tool access and permissions',
    audit: 'Review system activity logs',
    system: 'Core system settings and maintenance'
  }
  return map[activeTab.value] || 'System configuration'
})

const activeTabIcon = computed(() => getTabIcon(activeTab.value))

const activeTabComponent = computed(() => {
  const map = {
    profile: ProfileSettingsTab,
    soul: SoulSettingsTab,
    agents: AgentSettingsTab,
    providers: ProvidersSettingsTab,
    channels: ChannelsSettingsTab,
    social_media: SocialMediaSettingsTab,
    tools: ToolsSettingsTab,
    users: AdminUsersTab,
    permissions: ToolPermissionsTab,
    audit: AuditLogTab,
    system: CoreSystemTab
  }
  return map[activeTab.value] || null
})

const setActiveTabFromRoute = () => {
  const requestedTab = String(route.query.tab || '')
  if (requestedTab && tabs.value.some((tab) => tab.key === requestedTab)) {
    activeTab.value = requestedTab
  }
}

// Data Fetching & State
const usersList = ref([])
const showUserModal = ref(false)
const userForm = ref({})
const savingUser = ref(false)
const showBlockModal = ref(false)
const showUnblockModal = ref(false)
const blockForm = ref({ user: null, reason: '' })
const unblockForm = ref({ user: null })
const showChannelModal = ref(false)
const selectedChannel = ref(null)
const channelForm = ref({})

// Social Media modal state
const showSocialModal = ref(false)
const selectedPlatform = ref(null)
const socialAlias = ref('')
const socialIsNew = ref(true)
const socialForm = ref({})
const socialTesting = ref(false)
const socialTestResult = ref(null) // null | { ok: bool, error?: string }

const closeAllModals = () => {
  showChannelModal.value = false
  showSocialModal.value = false
  showUserModal.value = false
  showBlockModal.value = false
  showUnblockModal.value = false
}

const loadData = async () => {
  loading.value = true
  try {
    const promises = [advancedService.fetchUserConfig(), advancedService.fetchModels()]
    if (authStore.user?.role === 'admin') promises.push(usersService.listUsers())
    const results = await Promise.all(promises)
    configData.value = results[0].config || {}
    providersList.value = results[1].providers || []
    if (results[2]) usersList.value = results[2]
  } catch {
    toast.error('Failed to load settings')
  } finally {
    loading.value = false
  }
}

const saveConfig = async (payload) => {
  saving.value = true
  try {
    const hasSystemSettings = payload.web || payload.gateway || payload.storage
    const hasUserSettings = payload.tools || payload.channels || payload.agents || payload.providers
    if (hasSystemSettings && authStore.user?.role === 'admin') await advancedService.updateConfig(payload)
    if (hasUserSettings) await advancedService.updateUserConfig(payload)
    toast.success('Settings saved')
    setTimeout(loadData, 500)
  } catch (err) {
    toast.error('Failed to save: ' + err.message)
  } finally {
    saving.value = false
  }
}

const openUserModal = (user = null) => {
  userForm.value = user ? { id: user.id, username: user.username, password: '', role: user.role } : { username: '', password: '', role: 'user' }
  showUserModal.value = true
}

const openBlockModal = (user) => {
  blockForm.value = { user, reason: '' }
  showBlockModal.value = true
}

const openUnblockModal = (user) => {
  unblockForm.value = { user }
  showUnblockModal.value = true
}

const deleteUserLocal = async (user) => {
  if (!confirm(`Delete user @${user.username}?`)) return
  try {
    await usersService.deleteUser(user.id)
    toast.success('User deleted')
    loadData()
  } catch {
    toast.error('Failed to delete user')
  }
}

const saveUser = async () => {
  savingUser.value = true
  try {
    if (userForm.value.id) await usersService.updateUser(userForm.value.id, userForm.value.password, userForm.value.role)
    else await usersService.createUser(userForm.value.username, userForm.value.password, userForm.value.role)
    toast.success('User saved')
    showUserModal.value = false
    loadData()
  } catch {
    toast.error('Failed to save user')
  } finally {
    savingUser.value = false
  }
}

const blockUser = async () => {
  await usersService.blockUser(blockForm.value.user.id, blockForm.value.reason)
  toast.success('User blocked')
  showBlockModal.value = false
  loadData()
}

const unblockUser = async () => {
  await usersService.unblockUser(unblockForm.value.user.id)
  toast.success('User unblocked')
  showUnblockModal.value = false
  loadData()
}

const saveChannelConfig = async () => {
  const formData = { ...channelForm.value, enabled: true }
  // Don't send empty token if one already exists (would overwrite it with blank)
  if (formData.token === '' && formData._hasExistingToken) {
    delete formData.token
  }
  // For Slack: don't send empty bot_token if one already exists
  if (formData.bot_token === '' && formData._hasExistingToken) {
    delete formData.bot_token
  }
  // Remove empty app_token for Slack (don't overwrite if blank)
  if (formData.app_token === '') {
    delete formData.app_token
  }
  // For Email channel: don't send empty password if one already exists
  if (formData.password === '' && formData._hasExistingToken) {
    delete formData.password
  }
  // Remove internal tracking field before sending to backend
  delete formData._hasExistingToken
  // Convert comma-separated allow_from string to array for backend
  if (typeof formData.allow_from === 'string') {
    formData.allow_from = formData.allow_from.split(',').map(s => s.trim()).filter(Boolean)
  }
  const payload = { channels: { [selectedChannel.value.id]: formData } }
  await saveConfig(payload)
  showChannelModal.value = false
}

const openChannelConfig = (c) => {
  selectedChannel.value = c
  const existing = configData.value?.channels?.[c.id] || {}
  if (c.id === 'email') {
    channelForm.value = {
      imap_host: existing.imap_host || '',
      imap_port: existing.imap_port || 993,
      smtp_host: existing.smtp_host || '',
      smtp_port: existing.smtp_port || 587,
      username: existing.username || '',
      password: '',
      from: existing.from || '',
      allow_from: existing.allow_from || '',
      mailbox: existing.mailbox || 'INBOX',
      poll_interval_seconds: existing.poll_interval_seconds || 60,
      mark_as_read: existing.mark_as_read !== false,
      _hasExistingToken: existing.configured || false
    }
  } else if (c.id === 'slack') {
    channelForm.value = {
      bot_token: '',
      app_token: '',
      allow_from: existing.allow_from || '',
      _hasExistingToken: existing.configured || false
    }
  } else {
    channelForm.value = {
      token: '',  // Always empty — backend never returns real token (redacted)
      allow_from: existing.allow_from || '',
      _hasExistingToken: existing.configured || false
    }
  }
  showChannelModal.value = true
}

const toggleChannel = async (id) => {
  await saveConfig({ channels: { [id]: { enabled: !configData.value.channels[id]?.enabled } } })
}

// Social Media handlers
const platformFields = {
  twitter:  ['api_key', 'api_secret', 'access_token', 'access_token_secret'],
  bluesky:  ['handle', 'app_password', 'pds_url'],
  linkedin: ['access_token', 'author_urn'],
  facebook: ['page_access_token', 'page_id']
}

// { platform, alias } — alias is null when adding a new account
const openSocialConfig = ({ platform, alias }) => {
  selectedPlatform.value = platform
  socialAlias.value = alias ?? ''
  socialIsNew.value = !alias
  socialTestResult.value = null
  const fields = platformFields[platform.id] || []
  const form = {}
  fields.forEach(f => { form[f] = '' })
  const existing = configData.value?.tools?.social_media?.[platform.id]?.[alias]
  form._isConfigured = existing?.configured === true
  socialForm.value = form
  showSocialModal.value = true
}

const saveSocialConfig = async () => {
  const alias = socialAlias.value.trim()
  if (!alias) return
  const formData = { ...socialForm.value }
  delete formData._isConfigured
  // Remove empty strings — backend will keep existing secrets
  Object.keys(formData).forEach(k => { if (formData[k] === '') delete formData[k] })
  await saveConfig({ tools: { social_media: { [selectedPlatform.value.id]: { [alias]: formData } } } })
  showSocialModal.value = false
}

const deleteSocialAccount = async ({ platform, alias }) => {
  await saveConfig({ tools: { social_media: { [platform.id]: { [alias]: null } } } })
}

const testSocialConnection = async () => {
  const alias = socialAlias.value.trim()
  if (!alias) return
  socialTesting.value = true
  socialTestResult.value = null
  try {
    const res = await fetch(`/api/v1/social-media/${selectedPlatform.value.id}/${alias}/test`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    })
    const data = await res.json()
    socialTestResult.value = data
  } catch (err) {
    socialTestResult.value = { ok: false, error: err.message }
  } finally {
    socialTesting.value = false
  }
}

onMounted(() => {
  setActiveTabFromRoute()
  loadData()
})

watch(() => route.query.tab, setActiveTabFromRoute)
</script>

<style scoped>
.scrollbar-none::-webkit-scrollbar { display: none; }
.scrollbar-none { -ms-overflow-style: none; scrollbar-width: none; }

.modal-enter-active, .modal-leave-active { transition: opacity 0.3s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
