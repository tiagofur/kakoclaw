<template>
  <div class="flex h-full bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-pink-500/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-violet-500/20 via-transparent to-transparent" />
    </div>

    <!-- Left Sidebar: Campaign Tree -->
    <div class="relative z-10 w-72 flex-shrink-0 glass-panel border-r border-makoclaw-border/30 flex flex-col overflow-hidden">
      <!-- Sidebar Header -->
      <div class="px-4 pt-4 pb-3 border-b border-makoclaw-border/20">
        <div class="flex items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10">
              <svg class="w-4 h-4 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
              </svg>
            </div>
            <span class="text-sm font-bold text-makoclaw-text">Campaigns</span>
          </div>
          <button
            class="px-2.5 py-1.5 text-xs bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-lg transition-all shadow-sm shadow-pink-500/20 font-semibold flex items-center gap-1"
            @click="showNewModal = true"
          >
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            New
          </button>
        </div>
      </div>

      <!-- Campaign Tree -->
      <div class="flex-1 overflow-y-auto p-2">
        <!-- Loading -->
        <div v-if="loadingList" class="flex items-center justify-center py-12">
          <svg class="animate-spin w-6 h-6 text-pink-400" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
        </div>

        <!-- Empty state -->
        <div v-else-if="accounts.length === 0" class="flex flex-col items-center justify-center py-12 text-center px-4">
          <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10 mb-3">
            <svg class="w-6 h-6 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
            </svg>
          </div>
          <p class="text-sm font-semibold text-makoclaw-text">No campaigns yet</p>
          <p class="text-xs text-makoclaw-text-secondary/70 mt-1">Click + New to create your first campaign</p>
        </div>

        <!-- Account groups -->
        <div v-else class="space-y-1">
          <div v-for="(campaigns, account) in groupedCampaigns" :key="account">
            <!-- Account header -->
            <button
              class="w-full flex items-center gap-2 px-2 py-1.5 rounded-lg text-left hover:bg-makoclaw-surface/50 transition-colors"
              @click="toggleAccount(account)"
            >
              <svg
                class="w-3.5 h-3.5 text-makoclaw-text-secondary transition-transform flex-shrink-0"
                :class="expandedAccounts.has(account) ? 'rotate-90' : ''"
                fill="none" stroke="currentColor" viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
              <svg class="w-4 h-4 text-makoclaw-text-secondary/70 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
              </svg>
              <span class="text-xs font-semibold text-makoclaw-text truncate">{{ account }}</span>
              <span class="ml-auto text-[10px] text-makoclaw-text-secondary/50">{{ campaigns.length }}</span>
            </button>

            <!-- Campaign items -->
            <div v-if="expandedAccounts.has(account)" class="pl-6 space-y-0.5">
              <button
                v-for="campaign in campaigns"
                :key="campaign.campaign"
                class="w-full flex items-center gap-2 px-2 py-2 rounded-lg text-left transition-all text-sm"
                :class="isSelected(campaign) ? 'bg-pink-500/15 text-pink-400' : 'text-makoclaw-text-secondary hover:bg-makoclaw-surface/50 hover:text-makoclaw-text'"
                @click="selectCampaign(campaign)"
              >
                <svg class="w-3.5 h-3.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                </svg>
                <span class="truncate text-xs font-medium">{{ campaign.campaign }}</span>
                <!-- Status dots -->
                <div class="ml-auto flex gap-0.5 flex-shrink-0">
                  <span v-if="campaign.has_strategy" class="w-1.5 h-1.5 rounded-full bg-emerald-400" title="Has strategy" />
                  <span v-if="campaign.has_copy" class="w-1.5 h-1.5 rounded-full bg-blue-400" title="Has copy" />
                  <span v-if="campaign.has_assets" class="w-1.5 h-1.5 rounded-full bg-amber-400" title="Has assets" />
                </div>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Panel -->
    <div class="relative z-10 flex-1 flex flex-col overflow-hidden">
      <!-- Empty / no selection state -->
      <div v-if="!selectedCampaign" class="flex-1 flex items-center justify-center">
        <div class="text-center">
          <div class="relative mx-auto w-20 h-20 mb-6">
            <div class="absolute inset-0 bg-gradient-to-br from-pink-500/30 to-violet-500/30 rounded-3xl blur-2xl opacity-60" />
            <div class="relative w-20 h-20 rounded-2xl bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10">
              <svg class="w-10 h-10 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                  d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
              </svg>
            </div>
          </div>
          <h2 class="text-xl font-bold text-makoclaw-text">Marketing Campaigns</h2>
          <p class="text-sm text-makoclaw-text-secondary mt-2 max-w-xs mx-auto">
            Select a campaign from the sidebar, or create a new one to get started.
          </p>
        </div>
      </div>

      <!-- Campaign detail -->
      <div v-else class="flex-1 flex flex-col overflow-hidden">
        <!-- Campaign header -->
        <div class="glass-sticky px-6 pt-5 pb-0 border-b border-makoclaw-border/20 z-10">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10">
              <svg class="w-5 h-5 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
              </svg>
            </div>
            <div>
              <div class="flex items-center gap-2 text-xs text-makoclaw-text-secondary">
                <span>{{ selectedCampaign.account }}</span>
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </div>
              <h1 class="text-xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-pink-400 bg-clip-text text-transparent">
                {{ selectedCampaign.campaign }}
              </h1>
            </div>
          </div>

          <!-- Tabs -->
          <div class="flex gap-1 overflow-x-auto pb-0 hide-scrollbar">
            <button
              v-for="tab in availableTabs"
              :key="tab.id"
              class="px-3 py-2 text-xs font-semibold rounded-t-lg transition-all whitespace-nowrap border-b-2 -mb-px"
              :class="activeTab === tab.id
                ? 'text-pink-400 border-pink-400 bg-pink-500/5'
                : 'text-makoclaw-text-secondary border-transparent hover:text-makoclaw-text hover:bg-makoclaw-surface/30'"
              @click="activeTab = tab.id"
            >
              {{ tab.label }}
            </button>
          </div>
        </div>

        <!-- Tab content -->
        <div class="flex-1 overflow-y-auto">
          <!-- Loading detail -->
          <div v-if="loadingDetail" class="flex items-center justify-center py-20">
            <svg class="animate-spin w-8 h-8 text-pink-400" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
          </div>

          <div v-else class="p-6">
            <!-- Brief tab -->
            <div v-if="activeTab === 'brief'">
              <div v-if="campaignDetail?.brief" class="glass-panel rounded-2xl p-6">
                <div class="prose prose-invert prose-sm max-w-none">
                  <pre class="whitespace-pre-wrap text-sm text-makoclaw-text font-sans leading-relaxed">{{ campaignDetail.brief }}</pre>
                </div>
              </div>
              <div v-else class="flex flex-col items-center justify-center py-16 text-center">
                <div class="w-14 h-14 rounded-2xl bg-gradient-to-br from-pink-500/10 to-violet-500/10 flex items-center justify-center ring-1 ring-white/5 mb-4">
                  <svg class="w-7 h-7 text-pink-400/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <p class="text-sm text-makoclaw-text-secondary">No brief.md found in this campaign</p>
                <p class="text-xs text-makoclaw-text-secondary/60 mt-1">Ask the agent to generate a brief for this campaign</p>
              </div>
            </div>

            <!-- Strategy tab -->
            <div v-if="activeTab === 'strategy'">
              <div v-if="campaignDetail?.strategy" class="glass-panel rounded-2xl p-6">
                <pre class="whitespace-pre-wrap text-sm text-makoclaw-text font-sans leading-relaxed">{{ campaignDetail.strategy }}</pre>
              </div>
              <div v-else class="flex flex-col items-center justify-center py-16 text-center">
                <div class="w-14 h-14 rounded-2xl bg-gradient-to-br from-pink-500/10 to-violet-500/10 flex items-center justify-center ring-1 ring-white/5 mb-4">
                  <svg class="w-7 h-7 text-pink-400/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                  </svg>
                </div>
                <p class="text-sm text-makoclaw-text-secondary">No strategy.md found</p>
              </div>
            </div>

            <!-- Copy tab -->
            <div v-if="activeTab === 'copy'">
              <div v-if="campaignDetail?.files?.copy?.length" class="grid grid-cols-1 gap-4">
                <div
                  v-for="file in campaignDetail.files.copy"
                  :key="file.path"
                  class="glass-panel rounded-xl cursor-pointer hover:ring-1 hover:ring-pink-500/30 transition-all"
                  @click="previewFile(file)"
                >
                  <div class="p-4 flex items-center gap-3">
                    <div class="w-9 h-9 rounded-lg bg-gradient-to-br from-blue-500/20 to-indigo-500/20 flex items-center justify-center flex-shrink-0">
                      <svg class="w-4 h-4 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                      </svg>
                    </div>
                    <div class="flex-1 min-w-0">
                      <p class="text-sm font-semibold text-makoclaw-text truncate">{{ file.name }}</p>
                      <p class="text-xs text-makoclaw-text-secondary">{{ formatSize(file.size) }}</p>
                    </div>
                    <svg class="w-4 h-4 text-makoclaw-text-secondary/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                    </svg>
                  </div>
                  <!-- Preview if selected -->
                  <div v-if="previewedFile?.path === file.path && filePreviewContent" class="border-t border-makoclaw-border/30 p-4">
                    <pre class="whitespace-pre-wrap text-xs text-makoclaw-text-secondary font-mono leading-relaxed max-h-64 overflow-y-auto">{{ filePreviewContent }}</pre>
                  </div>
                </div>
              </div>
              <div v-else class="flex flex-col items-center justify-center py-16 text-center">
                <p class="text-sm text-makoclaw-text-secondary">No copy files found</p>
              </div>
            </div>

            <!-- Assets tab -->
            <div v-if="activeTab === 'assets'">
              <div v-if="campaignDetail?.files?.assets?.length" class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
                <div
                  v-for="file in campaignDetail.files.assets"
                  :key="file.path"
                  class="glass-panel rounded-xl overflow-hidden cursor-pointer hover:ring-1 hover:ring-pink-500/30 transition-all group"
                  @click="lightboxFile = file"
                >
                  <div class="aspect-square bg-makoclaw-bg/50 flex items-center justify-center overflow-hidden">
                    <img
                      v-if="file.is_image"
                      :src="fileUrl(file)"
                      :alt="file.name"
                      class="w-full h-full object-cover group-hover:scale-105 transition-transform"
                      loading="lazy"
                    >
                    <svg v-else class="w-12 h-12 text-makoclaw-text-secondary/30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                  </div>
                  <div class="p-2">
                    <p class="text-xs text-makoclaw-text truncate font-medium">{{ file.name }}</p>
                    <p class="text-[10px] text-makoclaw-text-secondary">{{ formatSize(file.size) }}</p>
                  </div>
                </div>
              </div>
              <div v-else class="flex flex-col items-center justify-center py-16 text-center">
                <p class="text-sm text-makoclaw-text-secondary">No assets found</p>
              </div>
            </div>

            <!-- Schedule tab -->
            <div v-if="activeTab === 'schedule'">
              <div v-if="campaignDetail?.files?.schedules?.length" class="space-y-3">
                <div
                  v-for="file in campaignDetail.files.schedules"
                  :key="file.path"
                  class="glass-panel rounded-xl overflow-hidden"
                >
                  <div class="p-4 border-b border-makoclaw-border/30 flex items-center justify-between">
                    <span class="text-sm font-semibold text-makoclaw-text">{{ file.name }}</span>
                    <span class="text-xs text-makoclaw-text-secondary">{{ formatSize(file.size) }}</span>
                  </div>
                  <ScheduleContent :url="fileUrl(file)" :token="authStore.token" />
                </div>
              </div>
              <div v-else class="flex flex-col items-center justify-center py-16 text-center">
                <p class="text-sm text-makoclaw-text-secondary">No schedule files found</p>
              </div>
            </div>

            <!-- Analytics tab -->
            <div v-if="activeTab === 'analytics'">
              <div v-if="campaignDetail?.files?.analytics?.length" class="space-y-4">
                <div
                  v-for="file in campaignDetail.files.analytics"
                  :key="file.path"
                  class="glass-panel rounded-xl overflow-hidden"
                >
                  <div class="p-4 border-b border-makoclaw-border/30 flex items-center justify-between">
                    <span class="text-sm font-semibold text-makoclaw-text">{{ file.name }}</span>
                    <span class="text-xs text-makoclaw-text-secondary">{{ formatSize(file.size) }}</span>
                  </div>
                  <ScheduleContent :url="fileUrl(file)" :token="authStore.token" />
                </div>
              </div>
              <div v-else class="flex flex-col items-center justify-center py-16 text-center">
                <p class="text-sm text-makoclaw-text-secondary">No analytics files found</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Lightbox -->
    <Transition name="modal">
      <div
        v-if="lightboxFile"
        class="fixed inset-0 bg-black/80 backdrop-blur-sm z-modal flex items-center justify-center p-4"
        @click.self="lightboxFile = null"
      >
        <div class="relative max-w-4xl w-full">
          <button
            class="absolute -top-10 right-0 text-white/70 hover:text-white transition-colors"
            @click="lightboxFile = null"
          >
            <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
          <img
            v-if="lightboxFile.is_image"
            :src="fileUrl(lightboxFile)"
            :alt="lightboxFile.name"
            class="w-full rounded-2xl shadow-2xl"
          >
          <div class="mt-3 text-center">
            <p class="text-sm text-white/70 font-medium">{{ lightboxFile.name }}</p>
            <p class="text-xs text-white/40">{{ formatSize(lightboxFile.size) }}</p>
          </div>
        </div>
      </div>
    </Transition>

    <!-- New Campaign Modal -->
    <Transition name="modal">
      <div
        v-if="showNewModal"
        class="fixed inset-0 bg-black/50 backdrop-blur-sm z-modal flex items-center justify-center p-4"
        @click.self="showNewModal = false"
      >
        <div class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl w-full max-w-lg shadow-2xl ring-1 ring-white/10 animate-scaleIn">
          <!-- Modal header -->
          <div class="p-5 border-b border-makoclaw-border/30 bg-gradient-to-r from-pink-500/10 to-violet-500/10">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10">
                <svg class="w-5 h-5 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
                </svg>
              </div>
              <div>
                <h2 class="text-base font-bold text-makoclaw-text">New Campaign</h2>
                <p class="text-xs text-makoclaw-text-secondary">Fill in the details below</p>
              </div>
            </div>
          </div>

          <!-- Modal body -->
          <div class="p-5 space-y-4">
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Account</label>
                <input
                  v-model="newCampaign.account"
                  type="text"
                  placeholder="e.g. acme-corp"
                  class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text"
                >
              </div>
              <div>
                <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Campaign</label>
                <input
                  v-model="newCampaign.campaign"
                  type="text"
                  placeholder="e.g. summer-2026"
                  class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text"
                >
              </div>
            </div>

            <div>
              <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Objective</label>
              <input
                v-model="newCampaign.objective"
                type="text"
                placeholder="e.g. increase brand awareness by 30%"
                class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text"
              >
            </div>

            <div>
              <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Platforms</label>
              <div class="flex flex-wrap gap-2">
                <label
                  v-for="platform in platformOptions"
                  :key="platform"
                  class="flex items-center gap-2 px-3 py-1.5 rounded-lg border cursor-pointer transition-all text-xs font-medium"
                  :class="newCampaign.platforms.includes(platform)
                    ? 'bg-pink-500/15 border-pink-500/40 text-pink-400'
                    : 'bg-makoclaw-bg/30 border-makoclaw-border/40 text-makoclaw-text-secondary hover:border-pink-500/30'"
                >
                  <input
                    type="checkbox"
                    class="hidden"
                    :value="platform"
                    :checked="newCampaign.platforms.includes(platform)"
                    @change="togglePlatform(platform)"
                  >
                  {{ platform }}
                </label>
              </div>
            </div>

            <div>
              <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Description</label>
              <textarea
                v-model="newCampaign.description"
                rows="3"
                placeholder="Brief description of the campaign goals..."
                class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text resize-none"
              />
            </div>

            <!-- Agent command hint -->
            <div class="p-3 bg-makoclaw-bg/40 rounded-xl border border-makoclaw-border/30">
              <p class="text-xs text-makoclaw-text-secondary font-semibold mb-1.5">Agent command that will be generated:</p>
              <code class="text-xs text-pink-400 font-mono break-all">{{ generatedCommand }}</code>
            </div>
          </div>

          <!-- Modal footer -->
          <div class="p-5 border-t border-makoclaw-border/30 flex justify-end gap-3">
            <button
              class="px-4 py-2 text-sm bg-makoclaw-surface/50 border border-makoclaw-border/50 rounded-xl hover:bg-makoclaw-surface-hover transition-colors text-makoclaw-text-secondary"
              @click="showNewModal = false"
            >
              Cancel
            </button>
            <button
              :disabled="!newCampaign.account || !newCampaign.campaign"
              class="px-5 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl transition-all disabled:opacity-50 font-semibold shadow-lg shadow-pink-500/20"
              @click="copyCommandAndClose"
            >
              Copy Command
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, defineComponent, h, watch } from 'vue'
import { useAuthStore } from '../stores/authStore'

// Inline component: loads a file URL and renders its content
const ScheduleContent = defineComponent({
  name: 'ScheduleContent',
  props: {
    url: { type: String, required: true },
    token: { type: String, default: '' }
  },
  setup(props) {
    const content = ref(null)
    const loading = ref(true)
    const error = ref(null)

    onMounted(async () => {
      try {
        const res = await fetch(props.url, {
          headers: { Authorization: `Bearer ${props.token}` }
        })
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const text = await res.text()
        // Try to parse as JSON for table rendering
        try {
          const parsed = JSON.parse(text)
          content.value = { type: 'json', data: parsed }
        } catch {
          content.value = { type: 'text', data: text }
        }
      } catch (e) {
        error.value = e.message
      } finally {
        loading.value = false
      }
    })

    return () => {
      if (loading.value) {
        return h('div', { class: 'p-4 text-xs text-makoclaw-text-secondary' }, 'Loading...')
      }
      if (error.value) {
        return h('div', { class: 'p-4 text-xs text-red-400' }, `Error: ${error.value}`)
      }
      if (content.value?.type === 'json') {
        const data = content.value.data
        const isArray = Array.isArray(data)
        if (isArray && data.length > 0 && typeof data[0] === 'object') {
          const keys = Object.keys(data[0])
          return h('div', { class: 'overflow-x-auto' }, [
            h('table', { class: 'w-full text-xs' }, [
              h('thead', {}, [
                h('tr', { class: 'border-b border-makoclaw-border/30' },
                  keys.map(k => h('th', { class: 'px-3 py-2 text-left text-makoclaw-text-secondary font-semibold' }, k))
                )
              ]),
              h('tbody', {},
                data.map((row, i) => h('tr', {
                  key: i,
                  class: 'border-b border-makoclaw-border/10 hover:bg-makoclaw-surface/20'
                },
                  keys.map(k => h('td', { class: 'px-3 py-2 text-makoclaw-text' }, String(row[k] ?? '')))
                ))
              )
            ])
          ])
        }
        return h('pre', { class: 'p-4 text-xs text-makoclaw-text font-mono whitespace-pre-wrap' }, JSON.stringify(data, null, 2))
      }
      return h('pre', { class: 'p-4 text-xs text-makoclaw-text font-sans whitespace-pre-wrap leading-relaxed' }, content.value?.data ?? '')
    }
  }
})

const authStore = useAuthStore()

// ---- State ----
const campaigns = ref([])
const loadingList = ref(false)
const loadingDetail = ref(false)
const selectedCampaign = ref(null)
const campaignDetail = ref(null)
const activeTab = ref('brief')
const expandedAccounts = ref(new Set())
const previewedFile = ref(null)
const filePreviewContent = ref(null)
const lightboxFile = ref(null)
const showNewModal = ref(false)

const newCampaign = ref({
  account: '',
  campaign: '',
  objective: '',
  platforms: [],
  description: ''
})

const platformOptions = ['Twitter', 'LinkedIn', 'Facebook', 'Instagram', 'TikTok', 'YouTube']

// ---- Computed ----
const groupedCampaigns = computed(() => {
  const groups = {}
  for (const c of campaigns.value) {
    if (!groups[c.account]) groups[c.account] = []
    groups[c.account].push(c)
  }
  return groups
})

const accounts = computed(() => Object.keys(groupedCampaigns.value))

const availableTabs = computed(() => {
  if (!selectedCampaign.value) return []
  const tabs = [{ id: 'brief', label: 'Brief' }]
  if (selectedCampaign.value.has_strategy) tabs.push({ id: 'strategy', label: 'Strategy' })
  if (selectedCampaign.value.has_copy) tabs.push({ id: 'copy', label: 'Copy' })
  if (selectedCampaign.value.has_assets) tabs.push({ id: 'assets', label: 'Assets' })
  if (selectedCampaign.value.has_schedule) tabs.push({ id: 'schedule', label: 'Schedule' })
  if (selectedCampaign.value.has_analytics) tabs.push({ id: 'analytics', label: 'Analytics' })
  // Always show all tabs even if empty — they handle their own empty state
  if (!tabs.find(t => t.id === 'copy')) tabs.push({ id: 'copy', label: 'Copy' })
  if (!tabs.find(t => t.id === 'assets')) tabs.push({ id: 'assets', label: 'Assets' })
  if (!tabs.find(t => t.id === 'schedule')) tabs.push({ id: 'schedule', label: 'Schedule' })
  if (!tabs.find(t => t.id === 'analytics')) tabs.push({ id: 'analytics', label: 'Analytics' })
  return tabs
})

const generatedCommand = computed(() => {
  const { account, campaign, objective, platforms, description } = newCampaign.value
  const platformStr = platforms.join(',') || 'twitter,linkedin'
  return `marketing_init_campaign(account="${account || 'account'}", campaign="${campaign || 'campaign-name'}", objective="${objective || 'your objective'}", platforms="${platformStr}", description="${description || 'Campaign description'}")`
})

// ---- Methods ----
function authHeaders() {
  return { Authorization: `Bearer ${authStore.token}` }
}

async function loadCampaigns() {
  loadingList.value = true
  try {
    const res = await fetch('/api/v1/marketing/campaigns', { headers: authHeaders() })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    campaigns.value = data.campaigns || []
    // Auto-expand all accounts
    for (const c of campaigns.value) {
      expandedAccounts.value.add(c.account)
    }
  } catch (e) {
    console.error('Failed to load campaigns:', e)
    campaigns.value = []
  } finally {
    loadingList.value = false
  }
}

async function loadCampaignDetail(campaign) {
  loadingDetail.value = true
  campaignDetail.value = null
  try {
    const res = await fetch(
      `/api/v1/marketing/campaigns/${encodeURIComponent(campaign.account)}/${encodeURIComponent(campaign.campaign)}`,
      { headers: authHeaders() }
    )
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    campaignDetail.value = await res.json()
  } catch (e) {
    console.error('Failed to load campaign detail:', e)
  } finally {
    loadingDetail.value = false
  }
}

function selectCampaign(campaign) {
  selectedCampaign.value = campaign
  activeTab.value = 'brief'
  previewedFile.value = null
  filePreviewContent.value = null
  loadCampaignDetail(campaign)
}

function isSelected(campaign) {
  return selectedCampaign.value?.account === campaign.account &&
    selectedCampaign.value?.campaign === campaign.campaign
}

function toggleAccount(account) {
  if (expandedAccounts.value.has(account)) {
    expandedAccounts.value.delete(account)
  } else {
    expandedAccounts.value.add(account)
  }
}

function fileUrl(file) {
  if (!selectedCampaign.value) return ''
  const { account, campaign } = selectedCampaign.value
  return `/api/v1/marketing/campaigns/${encodeURIComponent(account)}/${encodeURIComponent(campaign)}/files/${file.path}`
}

async function previewFile(file) {
  if (previewedFile.value?.path === file.path) {
    previewedFile.value = null
    filePreviewContent.value = null
    return
  }
  previewedFile.value = file
  filePreviewContent.value = null
  try {
    const res = await fetch(fileUrl(file), { headers: authHeaders() })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    filePreviewContent.value = await res.text()
  } catch (e) {
    filePreviewContent.value = 'Failed to load file content.'
  }
}

function formatSize(bytes) {
  if (bytes === 0) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function togglePlatform(platform) {
  const idx = newCampaign.value.platforms.indexOf(platform)
  if (idx >= 0) {
    newCampaign.value.platforms.splice(idx, 1)
  } else {
    newCampaign.value.platforms.push(platform)
  }
}

async function copyCommandAndClose() {
  try {
    await navigator.clipboard.writeText(generatedCommand.value)
  } catch (e) {
    // clipboard API may not be available in all contexts
  }
  showNewModal.value = false
  // Reset form
  newCampaign.value = { account: '', campaign: '', objective: '', platforms: [], description: '' }
}

// ---- Lifecycle ----
onMounted(() => {
  loadCampaigns()
})
</script>


