<template>
  <div class="flex h-full bg-makoclaw-bg relative overflow-hidden">
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-pink-500/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-violet-500/20 via-transparent to-transparent" />
    </div>

    <div class="relative z-10 w-72 flex-shrink-0 glass-panel border-r border-makoclaw-border/30 flex flex-col overflow-hidden">
      <div class="px-4 pt-4 pb-3 border-b border-makoclaw-border/20">
        <div class="flex items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10">
              <svg class="w-4 h-4 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
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

      <div class="flex-1 overflow-y-auto p-2">
        <div v-if="loadingList" class="flex items-center justify-center py-12">
          <svg class="animate-spin w-6 h-6 text-pink-400" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
        </div>

        <div v-else-if="accounts.length === 0" class="flex flex-col items-center justify-center py-12 text-center px-4">
          <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10 mb-3">
            <svg class="w-6 h-6 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
            </svg>
          </div>
          <p class="text-sm font-semibold text-makoclaw-text">No campaigns yet</p>
          <p class="text-xs text-makoclaw-text-secondary/70 mt-1">Click + New to create your first campaign</p>
        </div>

        <div v-else class="space-y-1">
          <div v-for="(accountCampaigns, account) in groupedCampaigns" :key="account">
            <button
              class="w-full flex items-center gap-2 px-2 py-1.5 rounded-lg text-left transition-colors"
              :class="focusedAccount === account && !selectedCampaign ? 'bg-pink-500/10 text-pink-400' : 'hover:bg-makoclaw-surface/50'"
              @click="handleAccountClick(account)"
            >
              <svg
                class="w-3.5 h-3.5 text-makoclaw-text-secondary transition-transform flex-shrink-0"
                :class="expandedAccounts.has(account) ? 'rotate-90' : ''"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
              <svg class="w-4 h-4 text-makoclaw-text-secondary/70 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
              </svg>
              <span class="text-xs font-semibold truncate">{{ account }}</span>
              <span class="ml-auto text-[10px] text-makoclaw-text-secondary/50">{{ accountCampaigns.length }}</span>
            </button>

            <div v-if="expandedAccounts.has(account)" class="pl-6 space-y-0.5">
              <button
                v-for="campaign in accountCampaigns"
                :key="campaign.campaign"
                class="w-full flex items-center gap-2 px-2 py-2 rounded-lg text-left transition-all text-sm"
                :class="isSelected(campaign) ? 'bg-pink-500/15 text-pink-400' : 'text-makoclaw-text-secondary hover:bg-makoclaw-surface/50 hover:text-makoclaw-text'"
                @click="handleSelectCampaign(campaign)"
              >
                <svg class="w-3.5 h-3.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                </svg>
                <span class="truncate text-xs font-medium">{{ campaign.campaign }}</span>
                <div class="ml-auto flex gap-1 items-center flex-shrink-0">
                  <span :class="statusPillClass(campaign.status)">{{ campaign.status || 'draft' }}</span>
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

    <div class="relative z-10 flex-1 flex flex-col overflow-hidden">
      <div v-if="!selectedCampaign" class="flex-1 overflow-y-auto custom-scrollbar p-6">
        <div class="max-w-6xl mx-auto space-y-6">
          <div class="text-center py-6">
            <div class="relative mx-auto w-20 h-20 mb-6">
              <div class="absolute inset-0 bg-gradient-to-br from-pink-500/30 to-violet-500/30 rounded-3xl blur-2xl opacity-60" />
              <div class="relative w-20 h-20 rounded-2xl bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10">
                <svg class="w-10 h-10 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
                </svg>
              </div>
            </div>
            <h2 class="text-xl font-bold text-makoclaw-text">Marketing Campaigns</h2>
            <p class="text-sm text-makoclaw-text-secondary mt-2 max-w-lg mx-auto">
              Select a campaign to manage briefs, strategy, templates and analytics. Or browse the account-level media library below.
            </p>
          </div>

          <div class="glass-panel rounded-2xl p-5 sm:p-6 space-y-4">
            <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
              <div>
                <h3 class="text-sm font-bold text-makoclaw-text">Account Media Library</h3>
                <p class="text-xs text-makoclaw-text-secondary mt-1">Choose an account to upload and organize reusable media.</p>
              </div>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="account in accounts"
                  :key="account"
                  class="px-3 py-2 min-h-[40px] rounded-xl border text-xs font-semibold transition-all"
                  :class="focusedAccount === account ? 'bg-pink-500/15 border-pink-500/40 text-pink-400' : 'bg-makoclaw-bg/40 border-makoclaw-border/40 text-makoclaw-text-secondary hover:border-pink-500/30 hover:text-makoclaw-text'"
                  @click="selectAccountMedia(account)"
                >
                  {{ account }}
                </button>
              </div>
            </div>

            <div v-if="!focusedAccount" class="flex flex-col items-center justify-center py-12 text-center">
              <p class="text-sm font-semibold text-makoclaw-text">Pick an account to view media</p>
              <p class="text-xs text-makoclaw-text-secondary/70 mt-1">The library is stored at account level and can be copied into campaigns later.</p>
            </div>

            <template v-else>
              <div class="flex flex-col lg:flex-row gap-3">
                <div class="flex-1 relative">
                  <input
                    v-model="mediaSearch"
                    type="text"
                    placeholder="Search by tag, filename or description"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 transition-all text-sm backdrop-blur-sm min-h-[40px] text-makoclaw-text"
                  >
                </div>
                <button
                  class="px-4 py-2.5 min-h-[40px] bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl transition-all shadow-lg shadow-pink-500/25 text-sm font-bold flex items-center justify-center gap-2 active:scale-95"
                  @click="showUploadForm = !showUploadForm"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a2 2 0 002 2h12a2 2 0 002-2v-1M16 8l-4-4m0 0L8 8m4-4v12" />
                  </svg>
                  Upload
                </button>
              </div>

              <div v-if="showUploadForm" class="grid grid-cols-1 lg:grid-cols-4 gap-3 p-4 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30">
                <div class="lg:col-span-1">
                  <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">File</label>
                  <input type="file" class="w-full text-xs text-makoclaw-text-secondary" @change="handleMediaFileChange">
                </div>
                <div class="lg:col-span-1">
                  <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Tags</label>
                  <input v-model="mediaUpload.tags" type="text" placeholder="launch, hero" class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text">
                </div>
                <div class="lg:col-span-1">
                  <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Description</label>
                  <input v-model="mediaUpload.description" type="text" placeholder="Hero banner" class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text">
                </div>
                <div class="lg:col-span-1 flex items-end">
                  <button
                    :disabled="!mediaUpload.file || mediaUploading"
                    class="w-full px-4 py-2.5 text-sm bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl transition-all disabled:opacity-50 font-semibold shadow-lg shadow-pink-500/20"
                    @click="uploadMedia"
                  >
                    {{ mediaUploading ? 'Uploading...' : 'Save Upload' }}
                  </button>
                </div>
              </div>

              <div v-if="mediaLoading" class="flex items-center justify-center py-12">
                <svg class="animate-spin w-6 h-6 text-pink-400" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
              </div>

              <div v-else-if="filteredMedia.length" class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4">
                <div v-for="item in filteredMedia" :key="item.filename" class="glass-panel rounded-2xl overflow-hidden border border-makoclaw-border/30 ring-1 ring-white/5">
                  <div class="aspect-video bg-makoclaw-bg/50 flex items-center justify-center overflow-hidden">
                    <img :src="mediaPreviewUrl(item)" :alt="item.filename" class="w-full h-full object-cover">
                  </div>
                  <div class="p-4 space-y-3">
                    <div>
                      <p class="text-sm font-semibold text-makoclaw-text truncate">{{ item.original_name || item.filename }}</p>
                      <p class="text-xs text-makoclaw-text-secondary mt-1">{{ item.description || 'No description yet' }}</p>
                    </div>

                    <div class="flex flex-wrap gap-2 min-h-[24px]">
                      <span v-for="tag in item.tags" :key="`${item.filename}-${tag}`" class="px-2 py-1 rounded-full text-[10px] font-semibold bg-pink-500/10 border border-pink-500/20 text-pink-300">#{{ tag }}</span>
                      <span v-if="!item.tags?.length" class="text-[10px] text-makoclaw-text-secondary/60">No tags</span>
                    </div>

                    <div class="grid grid-cols-1 sm:grid-cols-[1fr_auto] gap-2 items-center">
                      <select
                        v-model="mediaCopyTargets[item.filename]"
                        class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text"
                      >
                        <option value="">Select campaign</option>
                        <option v-for="campaign in accountCampaignOptions" :key="campaign.campaign" :value="campaign.campaign">{{ campaign.campaign }}</option>
                      </select>
                      <button
                        :disabled="!mediaCopyTargets[item.filename] || copyingMedia[item.filename]"
                        class="px-4 py-2 min-h-[40px] text-xs font-semibold rounded-xl bg-makoclaw-surface/50 border border-makoclaw-border/50 hover:border-pink-500/30 hover:text-pink-400 text-makoclaw-text-secondary transition-all disabled:opacity-50"
                        @click="copyMedia(item)"
                      >
                        {{ copyingMedia[item.filename] ? 'Copying...' : 'Copy to Campaign' }}
                      </button>
                    </div>
                    <div class="flex justify-end">
                      <button class="px-3 py-2 min-h-[40px] text-xs font-semibold rounded-xl bg-red-500/10 border border-red-500/20 hover:bg-red-500/20 text-red-400 transition-all" @click="deleteMediaItem(item)">
                        Delete
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              <div v-else class="flex flex-col items-center justify-center py-12 text-center">
                <p class="text-sm font-semibold text-makoclaw-text">No media matched your filter</p>
                <p class="text-xs text-makoclaw-text-secondary/70 mt-1">Upload something new or try another tag search.</p>
              </div>
            </template>
          </div>
        </div>
      </div>

      <div v-else class="flex-1 flex flex-col overflow-hidden">
        <div class="glass-sticky px-6 pt-5 pb-0 border-b border-makoclaw-border/20 z-10">
          <div class="flex items-start justify-between gap-4 mb-4">
            <div class="flex items-center gap-3 min-w-0">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10">
                <svg class="w-5 h-5 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
                </svg>
              </div>
              <div class="min-w-0">
                <div class="flex items-center gap-2 text-xs text-makoclaw-text-secondary">
                  <span>{{ selectedCampaign.account }}</span>
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                  </svg>
                </div>
                <h1 class="text-xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-pink-400 bg-clip-text text-transparent truncate">{{ selectedCampaign.campaign }}</h1>
              </div>
            </div>

            <div class="flex items-center gap-2 flex-wrap justify-end">
              <div class="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-makoclaw-surface/40 border border-makoclaw-border/40 backdrop-blur-sm">
                <span :class="statusPillClass(currentStatus)">{{ currentStatus }}</span>
                <select
                  :value="currentStatus"
                  class="bg-transparent text-xs text-makoclaw-text border-0 focus:outline-none min-h-[32px]"
                  @change="handleStatusChange($event.target.value)"
                >
                  <option v-for="status in statusOptions" :key="status" :value="status">{{ status }}</option>
                </select>
              </div>
              <button
                class="px-3 py-2 min-h-[40px] bg-red-500/10 border border-red-500/20 text-red-400 rounded-xl hover:bg-red-500/20 transition-colors text-sm font-medium"
                @click="handleDeleteCampaign"
              >
                Delete
              </button>
            </div>
          </div>

          <div class="flex gap-1 overflow-x-auto pb-0 hide-scrollbar">
            <button
              v-for="tab in availableTabs"
              :key="tab.id"
              class="px-3 py-2 text-xs font-semibold rounded-t-lg transition-all whitespace-nowrap border-b-2 -mb-px"
              :class="activeTab === tab.id ? 'text-pink-400 border-pink-400 bg-pink-500/5' : 'text-makoclaw-text-secondary border-transparent hover:text-makoclaw-text hover:bg-makoclaw-surface/30'"
              @click="requestTabSwitch(tab.id)"
            >
              {{ tab.label }}
            </button>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto custom-scrollbar">
          <div v-if="loadingDetail" class="flex items-center justify-center py-20">
            <svg class="animate-spin w-8 h-8 text-pink-400" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
          </div>

          <div v-else class="p-6 space-y-6">
            <div v-if="activeTab === 'brief'">
              <div v-if="campaignDetail?.brief || isEditingBrief" class="glass-panel rounded-2xl p-6 space-y-4">
                <div class="flex items-center justify-between gap-2">
                  <h2 class="text-sm font-semibold text-makoclaw-text">Brief</h2>
                  <div class="flex items-center gap-2">
                    <template v-if="isEditingBrief">
                      <button class="px-3 py-1.5 text-sm bg-makoclaw-surface/50 border border-makoclaw-border/50 rounded-lg hover:bg-makoclaw-surface-hover transition-colors text-makoclaw-text-secondary" @click="cancelEditing">Cancel</button>
                      <button class="px-4 py-1.5 text-sm bg-gradient-to-r from-pink-500 to-violet-500 text-white rounded-lg hover:from-pink-600 hover:to-violet-600 transition-colors" @click="saveBrief">Save</button>
                    </template>
                    <button v-else class="px-3 py-1.5 text-sm bg-pink-500/10 border border-pink-500/30 rounded-lg hover:bg-pink-500/20 transition-colors text-pink-400" @click="startEditingBrief">Edit</button>
                  </div>
                </div>

                <textarea
                  v-if="isEditingBrief"
                  v-model="editContent"
                  rows="16"
                  class="w-full bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl p-4 text-sm font-sans overflow-auto whitespace-pre-wrap focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 resize-y"
                />
                <div v-else class="prose prose-invert prose-sm max-w-none">
                  <pre class="whitespace-pre-wrap text-sm text-makoclaw-text font-sans leading-relaxed">{{ campaignDetail?.brief }}</pre>
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

            <div v-if="activeTab === 'strategy'">
              <div v-if="campaignDetail?.strategy || isEditingStrategy" class="glass-panel rounded-2xl p-6 space-y-4">
                <div class="flex items-center justify-between gap-2">
                  <h2 class="text-sm font-semibold text-makoclaw-text">Strategy</h2>
                  <div class="flex items-center gap-2">
                    <template v-if="isEditingStrategy">
                      <button class="px-3 py-1.5 text-sm bg-makoclaw-surface/50 border border-makoclaw-border/50 rounded-lg hover:bg-makoclaw-surface-hover transition-colors text-makoclaw-text-secondary" @click="cancelEditing">Cancel</button>
                      <button class="px-4 py-1.5 text-sm bg-gradient-to-r from-pink-500 to-violet-500 text-white rounded-lg hover:from-pink-600 hover:to-violet-600 transition-colors" @click="saveStrategy">Save</button>
                    </template>
                    <button v-else class="px-3 py-1.5 text-sm bg-pink-500/10 border border-pink-500/30 rounded-lg hover:bg-pink-500/20 transition-colors text-pink-400" @click="startEditingStrategy">Edit</button>
                  </div>
                </div>

                <textarea
                  v-if="isEditingStrategy"
                  v-model="editContent"
                  rows="16"
                  class="w-full bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl p-4 text-sm font-sans overflow-auto whitespace-pre-wrap focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 resize-y"
                />
                <pre v-else class="whitespace-pre-wrap text-sm text-makoclaw-text font-sans leading-relaxed">{{ campaignDetail?.strategy }}</pre>
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
                  <div v-if="previewedFile?.path === file.path && filePreviewContent" class="border-t border-makoclaw-border/30 p-4">
                    <pre class="whitespace-pre-wrap text-xs text-makoclaw-text-secondary font-mono leading-relaxed max-h-64 overflow-y-auto">{{ filePreviewContent }}</pre>
                  </div>
                </div>
              </div>
              <div v-else class="flex flex-col items-center justify-center py-16 text-center">
                <p class="text-sm text-makoclaw-text-secondary">No copy files found</p>
              </div>
            </div>

            <div v-if="activeTab === 'templates'" class="space-y-6">
              <div class="glass-panel rounded-2xl p-5 sm:p-6 space-y-4">
                <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                  <div>
                    <h2 class="text-sm font-bold text-makoclaw-text">Account Templates</h2>
                    <p class="text-xs text-makoclaw-text-secondary mt-1">Reusable Mustache templates for {{ selectedCampaign.account }}</p>
                  </div>
                  <button class="px-4 py-2.5 min-h-[40px] bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl transition-all shadow-lg shadow-pink-500/25 text-sm font-bold flex items-center gap-2 active:scale-95" @click="startNewTemplate">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    New Template
                  </button>
                </div>

                <div v-if="showTemplateForm" class="grid grid-cols-1 xl:grid-cols-2 gap-5 p-4 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30">
                  <div class="space-y-4">
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <div>
                        <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Name</label>
                        <input v-model="templateForm.name" type="text" placeholder="Launch promo" class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text">
                      </div>
                      <div>
                        <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Slug</label>
                        <input v-model="templateForm.slug" :readonly="templateFormMode === 'edit'" type="text" placeholder="launch-promo" class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text" @input="templateSlugTouched = true">
                        <p class="mt-1 text-[10px] text-makoclaw-text-secondary/70">{{ templateFormMode === 'create' ? 'Auto-generated from the template name.' : 'Slug is locked after creation.' }}</p>
                      </div>
                      <div class="sm:col-span-2">
                        <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Platform</label>
                        <select v-model="templateForm.platform" class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text">
                          <option value="">Select platform</option>
                          <option v-for="platform in platformOptions" :key="platform" :value="platform.toLowerCase()">{{ platform }}</option>
                        </select>
                      </div>
                    </div>

                    <div>
                      <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Mustache Content</label>
                      <textarea v-model="templateForm.body" rows="10" placeholder="Hello {{first_name}}, your offer is ready." class="w-full px-3 py-3 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text resize-y font-mono" />
                    </div>

                    <div class="space-y-2">
                      <p class="text-xs font-semibold text-makoclaw-text-secondary">Detected Variables</p>
                      <div class="flex flex-wrap gap-2 min-h-[24px]">
                        <span v-for="variable in templateVariables" :key="variable" class="px-2.5 py-1 rounded-full text-[10px] font-semibold bg-violet-500/10 border border-violet-500/20 text-violet-300">{{ formatTemplateToken(variable) }}</span>
                        <span v-if="!templateVariables.length" class="text-xs text-makoclaw-text-secondary/60">No variables detected yet</span>
                      </div>
                    </div>

                    <div class="flex justify-end gap-3">
                      <button class="px-4 py-2 text-sm bg-makoclaw-surface/50 border border-makoclaw-border/50 rounded-xl hover:bg-makoclaw-surface-hover transition-colors text-makoclaw-text-secondary" @click="cancelTemplateForm">Cancel</button>
                      <button :disabled="savingTemplate" class="px-5 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl transition-all disabled:opacity-50 font-semibold shadow-lg shadow-pink-500/20" @click="submitTemplate">
                        {{ savingTemplate ? 'Saving...' : templateFormMode === 'create' ? 'Create Template' : 'Save Changes' }}
                      </button>
                    </div>
                  </div>

                  <div class="space-y-4">
                    <div class="glass-panel rounded-2xl p-4 border border-makoclaw-border/20">
                      <div class="flex items-center justify-between gap-2 mb-3">
                        <h3 class="text-xs font-bold text-makoclaw-text uppercase tracking-wide">Preview Inputs</h3>
                        <span class="text-[10px] text-makoclaw-text-secondary">Client-side substitution</span>
                      </div>
                      <div v-if="templateVariables.length" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                        <div v-for="variable in templateVariables" :key="`preview-${variable}`">
                          <label class="block text-[10px] font-semibold text-makoclaw-text-secondary mb-1">{{ variable }}</label>
                          <input v-model="templatePreviewValues[variable]" type="text" :placeholder="`Value for ${variable}`" class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text">
              </div>
            </div>
                      <p v-else class="text-xs text-makoclaw-text-secondary/70">Add <code>{{ formatTemplateToken('variable') }}</code> tokens to generate a live preview.</p>
                    </div>

                    <div class="glass-panel rounded-2xl p-4 border border-makoclaw-border/20">
                      <h3 class="text-xs font-bold text-makoclaw-text uppercase tracking-wide mb-3">Rendered Preview</h3>
                      <pre class="whitespace-pre-wrap text-sm text-makoclaw-text font-sans leading-relaxed min-h-[220px]">{{ renderedTemplatePreview }}</pre>
                    </div>
                  </div>
                </div>

                <div v-if="templatesLoading" class="flex items-center justify-center py-12">
                  <svg class="animate-spin w-6 h-6 text-pink-400" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                </div>

                <div v-else-if="templates.length" class="grid grid-cols-1 xl:grid-cols-2 gap-4">
                  <div v-for="template in templates" :key="template.slug" class="glass-panel rounded-2xl p-5 border border-makoclaw-border/30 ring-1 ring-white/5 space-y-4">
                    <div class="flex items-start justify-between gap-3">
                      <div class="min-w-0">
                        <div class="flex flex-wrap items-center gap-2">
                          <h3 class="text-sm font-bold text-makoclaw-text truncate">{{ template.name }}</h3>
                          <span class="px-2 py-1 rounded-full text-[10px] font-semibold bg-pink-500/10 border border-pink-500/20 text-pink-300 uppercase">{{ template.platform || 'generic' }}</span>
                        </div>
                        <p class="text-xs text-makoclaw-text-secondary mt-1">Slug: {{ template.slug }}</p>
                      </div>
                      <div class="flex items-center gap-2">
                        <button class="px-3 py-1.5 text-xs font-semibold rounded-lg bg-makoclaw-surface/50 border border-makoclaw-border/50 hover:border-pink-500/30 hover:text-pink-400 text-makoclaw-text-secondary transition-all" @click="editTemplate(template)">Edit</button>
                        <button class="px-3 py-1.5 text-xs font-semibold rounded-lg bg-red-500/10 border border-red-500/20 hover:bg-red-500/20 text-red-400 transition-all" @click="deleteTemplateItem(template)">Delete</button>
                      </div>
                    </div>

                    <div class="flex flex-wrap gap-2 min-h-[24px]">
                      <span v-for="variable in template.variables" :key="`${template.slug}-${variable}`" class="px-2.5 py-1 rounded-full text-[10px] font-semibold bg-violet-500/10 border border-violet-500/20 text-violet-300">{{ variable }}</span>
                      <span v-if="!template.variables?.length" class="text-xs text-makoclaw-text-secondary/60">No variables</span>
                    </div>

                    <pre class="whitespace-pre-wrap text-sm text-makoclaw-text-secondary font-sans leading-relaxed bg-makoclaw-bg/30 rounded-xl p-3 border border-makoclaw-border/20">{{ template.body }}</pre>
                  </div>
                </div>

                <div v-else class="flex flex-col items-center justify-center py-12 text-center">
                  <p class="text-sm font-semibold text-makoclaw-text">No templates for this account yet</p>
                  <p class="text-xs text-makoclaw-text-secondary/70 mt-1">Create your first reusable template and inject variables like <code>{{ formatTemplateToken('first_name') }}</code>.</p>
                </div>
              </div>
            </div>

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

            <div v-if="activeTab === 'schedule'">
              <div v-if="campaignDetail?.files?.schedules?.length" class="space-y-3">
                <div v-for="file in campaignDetail.files.schedules" :key="file.path" class="glass-panel rounded-xl overflow-hidden">
                  <div class="p-4 border-b border-makoclaw-border/30 flex items-center justify-between">
                    <span class="text-sm font-semibold text-makoclaw-text">{{ file.name }}</span>
                    <span class="text-xs text-makoclaw-text-secondary">{{ formatSize(file.size) }}</span>
                  </div>
                  <ScheduleContent :path="fileWorkspacePath(file)" />
                </div>
              </div>
              <div v-else class="flex flex-col items-center justify-center py-16 text-center">
                <p class="text-sm text-makoclaw-text-secondary">No schedule files found</p>
              </div>
            </div>

            <div v-if="activeTab === 'media'" class="space-y-6">
              <div class="glass-panel rounded-2xl p-5 sm:p-6 space-y-4">
                <div class="flex flex-col lg:flex-row gap-3 lg:items-center lg:justify-between">
                  <div>
                    <h2 class="text-sm font-bold text-makoclaw-text">Media Library</h2>
                    <p class="text-xs text-makoclaw-text-secondary mt-1">Account-level assets for {{ currentAccount }}</p>
                  </div>
                  <div class="flex flex-col sm:flex-row gap-3 w-full lg:w-auto">
                    <input v-model="mediaSearch" type="text" placeholder="Search by tag, filename or description" class="w-full sm:w-72 px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 transition-all text-sm backdrop-blur-sm min-h-[40px] text-makoclaw-text">
                    <button class="px-4 py-2.5 min-h-[40px] bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl transition-all shadow-lg shadow-pink-500/25 text-sm font-bold flex items-center justify-center gap-2 active:scale-95" @click="showUploadForm = !showUploadForm">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a2 2 0 002 2h12a2 2 0 002-2v-1M16 8l-4-4m0 0L8 8m4-4v12" />
                      </svg>
                      Upload
                    </button>
                  </div>
                </div>

                <div v-if="showUploadForm" class="grid grid-cols-1 lg:grid-cols-4 gap-3 p-4 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30">
                  <div class="lg:col-span-1">
                    <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">File</label>
                    <input type="file" class="w-full text-xs text-makoclaw-text-secondary" @change="handleMediaFileChange">
                  </div>
                  <div class="lg:col-span-1">
                    <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Tags</label>
                    <input v-model="mediaUpload.tags" type="text" placeholder="launch, hero" class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text">
                  </div>
                  <div class="lg:col-span-1">
                    <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Description</label>
                    <input v-model="mediaUpload.description" type="text" placeholder="Hero banner" class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text">
                  </div>
                  <div class="lg:col-span-1 flex items-end">
                    <button :disabled="!mediaUpload.file || mediaUploading" class="w-full px-4 py-2.5 text-sm bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl transition-all disabled:opacity-50 font-semibold shadow-lg shadow-pink-500/20" @click="uploadMedia">
                      {{ mediaUploading ? 'Uploading...' : 'Save Upload' }}
                    </button>
                  </div>
                </div>

                <div v-if="mediaLoading" class="flex items-center justify-center py-12">
                  <svg class="animate-spin w-6 h-6 text-pink-400" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                </div>

                <div v-else-if="filteredMedia.length" class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4">
                  <div v-for="item in filteredMedia" :key="item.filename" class="glass-panel rounded-2xl overflow-hidden border border-makoclaw-border/30 ring-1 ring-white/5">
                    <div class="aspect-video bg-makoclaw-bg/50 flex items-center justify-center overflow-hidden">
                      <img :src="mediaPreviewUrl(item)" :alt="item.filename" class="w-full h-full object-cover">
                    </div>
                    <div class="p-4 space-y-3">
                      <div>
                        <p class="text-sm font-semibold text-makoclaw-text truncate">{{ item.original_name || item.filename }}</p>
                        <p class="text-xs text-makoclaw-text-secondary mt-1">{{ item.description || 'No description yet' }}</p>
                      </div>
                      <div class="flex flex-wrap gap-2 min-h-[24px]">
                        <span v-for="tag in item.tags" :key="`${item.filename}-${tag}`" class="px-2 py-1 rounded-full text-[10px] font-semibold bg-pink-500/10 border border-pink-500/20 text-pink-300">#{{ tag }}</span>
                        <span v-if="!item.tags?.length" class="text-[10px] text-makoclaw-text-secondary/60">No tags</span>
                      </div>
                      <div class="grid grid-cols-1 sm:grid-cols-[1fr_auto] gap-2 items-center">
                        <select v-model="mediaCopyTargets[item.filename]" class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text">
                          <option value="">Select campaign</option>
                          <option v-for="campaign in accountCampaignOptions" :key="campaign.campaign" :value="campaign.campaign">{{ campaign.campaign }}</option>
                        </select>
                        <button :disabled="!mediaCopyTargets[item.filename] || copyingMedia[item.filename]" class="px-4 py-2 min-h-[40px] text-xs font-semibold rounded-xl bg-makoclaw-surface/50 border border-makoclaw-border/50 hover:border-pink-500/30 hover:text-pink-400 text-makoclaw-text-secondary transition-all disabled:opacity-50" @click="copyMedia(item)">
                          {{ copyingMedia[item.filename] ? 'Copying...' : 'Copy to Campaign' }}
                        </button>
                      </div>
                      <div class="flex justify-end">
                        <button class="px-3 py-2 min-h-[40px] text-xs font-semibold rounded-xl bg-red-500/10 border border-red-500/20 hover:bg-red-500/20 text-red-400 transition-all" @click="deleteMediaItem(item)">
                          Delete
                        </button>
                      </div>
                    </div>
                  </div>
                </div>

                <div v-else class="flex flex-col items-center justify-center py-16 text-center">
                  <p class="text-sm text-makoclaw-text-secondary">No media found for this account</p>
                </div>
              </div>
            </div>

            <div v-if="activeTab === 'analytics'" class="space-y-6">
              <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
                <div v-for="metric in analyticsKpis" :key="metric.label" class="bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 rounded-2xl p-4 sm:p-5 transition-all duration-300 hover:bg-makoclaw-surface/50 hover:border-pink-500/20 hover:shadow-lg hover:shadow-pink-500/5 group ring-1 ring-white/5 relative overflow-hidden">
                  <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-pink-500 to-violet-500 group-hover:w-full transition-all duration-500 opacity-50" />
                  <div class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary/60 mb-1">{{ metric.label }}</div>
                  <div class="text-xl sm:text-2xl font-bold text-makoclaw-text">{{ metric.value }}</div>
                </div>
              </div>

              <div v-if="analyticsLoading" class="flex items-center justify-center py-12">
                <svg class="animate-spin w-6 h-6 text-pink-400" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
              </div>

              <template v-else-if="analyticsHasEntries">
                <div class="grid grid-cols-1 xl:grid-cols-3 gap-6">
                  <div class="xl:col-span-2 bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 rounded-2xl p-5 ring-1 ring-white/5">
                    <div class="flex items-center justify-between mb-5">
                      <h3 class="text-sm font-bold text-makoclaw-text">Impressions per Platform</h3>
                      <div class="w-2 h-2 rounded-full bg-pink-500 animate-pulse shadow-lg shadow-pink-500/50" />
                    </div>
                    <div class="relative min-h-[320px] flex items-center justify-center">
                      <Bar v-if="analyticsBarData.labels.length" :data="analyticsBarData" :options="analyticsBarOptions" />
                    </div>
                  </div>

                  <div class="bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 rounded-2xl p-5 ring-1 ring-white/5">
                    <div class="flex items-center justify-between mb-5">
                      <h3 class="text-sm font-bold text-makoclaw-text">Engagement Breakdown</h3>
                      <div class="w-2 h-2 rounded-full bg-violet-500 shadow-lg shadow-violet-500/50" />
                    </div>
                    <div class="relative min-h-[320px] flex items-center justify-center">
                      <Doughnut v-if="analyticsDoughnutData.labels.length" :data="analyticsDoughnutData" :options="analyticsDoughnutOptions" />
                      <div v-else class="text-sm text-makoclaw-text-secondary/50">No engagement details yet</div>
                    </div>
                  </div>
                </div>

                <div class="bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 rounded-2xl p-5 ring-1 ring-white/5">
                  <div class="flex items-center justify-between mb-5">
                    <h3 class="text-sm font-bold text-makoclaw-text">Impressions Trend</h3>
                    <div class="w-2 h-2 rounded-full bg-pink-500 shadow-lg shadow-pink-500/50" />
                  </div>
                  <div class="relative min-h-[320px] flex items-center justify-center">
                    <Line v-if="analyticsLineData.labels.length" :data="analyticsLineData" :options="analyticsLineOptions" />
                    <div v-else class="text-sm text-makoclaw-text-secondary/50">No trend data yet</div>
                  </div>
                </div>
              </template>

              <div v-else-if="showLegacyAnalyticsFallback" class="space-y-4">
                <div class="glass-panel rounded-2xl p-5 border border-amber-500/20 bg-amber-500/5">
                  <h3 class="text-sm font-bold text-makoclaw-text">Legacy Analytics Reports</h3>
                  <p class="text-xs text-makoclaw-text-secondary mt-1">No aggregated JSON entries were found, so we are showing the existing report files instead.</p>
                </div>
                <div v-for="file in legacyAnalyticsFiles" :key="file.path" class="glass-panel rounded-xl overflow-hidden">
                  <div class="p-4 border-b border-makoclaw-border/30 flex items-center justify-between">
                    <span class="text-sm font-semibold text-makoclaw-text">{{ file.name }}</span>
                    <span class="text-xs text-makoclaw-text-secondary">{{ formatSize(file.size) }}</span>
                  </div>
                  <ScheduleContent :path="fileWorkspacePath(file)" />
                </div>
              </div>

              <div v-else class="flex flex-col items-center justify-center py-16 text-center px-8">
                <div class="w-14 h-14 rounded-2xl bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10 mb-4">
                  <svg class="w-7 h-7 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg>
                </div>
                <p class="text-sm font-semibold text-makoclaw-text">No analytics data yet</p>
                <p class="text-xs text-makoclaw-text-secondary/60 mt-2 max-w-xs">Ask your AI agent to track performance using the <code class="px-1.5 py-0.5 bg-white/10 rounded text-pink-400 text-[11px]">social_analytics</code> tool. Data will appear here once reports are generated.</p>
              </div>
            </div>

            <!-- Audience Tab -->
            <div v-if="activeTab === 'audience'" class="space-y-6">
              <!-- Sub-tabs -->
              <div class="flex gap-2 border-b border-makoclaw-border/20 pb-3">
                <button
                  v-for="sub in [{ id: 'contacts', label: 'Contacts' }, { id: 'lists', label: 'Lists' }, { id: 'segments', label: 'Segments' }, { id: 'deliveries', label: 'Deliveries' }, { id: 'campaigns', label: 'Campaigns' }, { id: 'automations', label: 'Automations' }, { id: 'accounts', label: 'Accounts' }]"
                  :key="sub.id"
                  class="px-4 py-2 rounded-xl text-sm font-medium transition-all"
                  :class="audienceSubTab === sub.id ? 'bg-pink-500/20 text-pink-400' : 'text-white/60 hover:text-white/80'"
                  @click="audienceSubTab = sub.id"
                >{{ sub.label }}</button>
              </div>

              <!-- Contacts Sub-tab -->
              <AudienceContacts v-if="audienceSubTab === 'contacts'" @reload="loadAudienceData()" />

              <!-- Lists Sub-tab -->
              <AudienceLists v-if="audienceSubTab === 'lists'" @reload="loadAudienceData()" />

              <!-- Segments Sub-tab -->
              <AudienceSegments v-if="audienceSubTab === 'segments'" @reload="loadAudienceData()" />

              <!-- Deliveries Sub-tab -->
              <AudienceDeliveries v-if="audienceSubTab === 'deliveries'" @reload="loadAudienceData()" />

              <!-- Campaigns Sub-tab -->
              <AudienceEmailCampaigns v-if="audienceSubTab === 'campaigns'" @reload="loadAudienceData()" />

              <!-- Automations Sub-tab -->
              <AudienceAutomations v-if="audienceSubTab === 'automations'" @reload="loadAudienceData()" />

              <!-- Accounts Sub-tab -->
              <AudienceAccounts v-if="audienceSubTab === 'accounts'" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <Transition name="modal">
      <div
        v-if="lightboxFile"
        class="fixed inset-0 bg-black/80 backdrop-blur-sm z-modal flex items-center justify-center p-4"
        @click.self="lightboxFile = null"
      >
        <div class="relative max-w-4xl w-full">
          <button class="absolute -top-10 right-0 text-white/70 hover:text-white transition-colors" @click="lightboxFile = null">
            <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
          <img v-if="lightboxFile.is_image" :src="fileUrl(lightboxFile)" :alt="lightboxFile.name" class="w-full rounded-2xl shadow-2xl">
          <div class="mt-3 text-center">
            <p class="text-sm text-white/70 font-medium">{{ lightboxFile.name }}</p>
            <p class="text-xs text-white/40">{{ formatSize(lightboxFile.size) }}</p>
          </div>
        </div>
      </div>
    </Transition>

    <Transition name="modal">
      <div
        v-if="showNewModal"
        class="fixed inset-0 bg-black/50 backdrop-blur-sm z-modal flex items-center justify-center p-4"
        @click.self="showNewModal = false"
      >
        <div class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl w-full max-w-lg shadow-2xl ring-1 ring-white/10 animate-scaleIn">
          <div class="p-5 border-b border-makoclaw-border/30 bg-gradient-to-r from-pink-500/10 to-violet-500/10">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10">
                <svg class="w-5 h-5 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
                </svg>
              </div>
              <div>
                <h2 class="text-base font-bold text-makoclaw-text">New Campaign</h2>
                <p class="text-xs text-makoclaw-text-secondary">Fill in the details below</p>
              </div>
            </div>
          </div>

          <div class="p-5 space-y-4">
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Account</label>
                <input v-model="newCampaign.account" type="text" placeholder="e.g. acme-corp" class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text">
              </div>
              <div>
                <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Campaign</label>
                <input v-model="newCampaign.campaign" type="text" placeholder="e.g. summer-2026" class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text">
              </div>
            </div>

            <div>
              <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Objective</label>
              <input v-model="newCampaign.objective" type="text" placeholder="e.g. increase brand awareness by 30%" class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text">
            </div>

            <div>
              <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Platforms</label>
              <div class="flex flex-wrap gap-2">
                <label
                  v-for="platform in platformOptions"
                  :key="platform"
                  class="flex items-center gap-2 px-3 py-1.5 rounded-lg border cursor-pointer transition-all text-xs font-medium"
                  :class="newCampaign.platforms.includes(platform) ? 'bg-pink-500/15 border-pink-500/40 text-pink-400' : 'bg-makoclaw-bg/30 border-makoclaw-border/40 text-makoclaw-text-secondary hover:border-pink-500/30'"
                >
                  <input type="checkbox" class="hidden" :value="platform" :checked="newCampaign.platforms.includes(platform)" @change="togglePlatform(platform)">
                  {{ platform }}
                </label>
              </div>
            </div>

            <div>
              <label class="block text-xs font-semibold text-makoclaw-text-secondary mb-1.5">Description</label>
              <textarea v-model="newCampaign.description" rows="3" placeholder="Brief description of the campaign goals..." class="w-full px-3 py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 text-makoclaw-text resize-none" />
            </div>

            <div class="p-3 bg-makoclaw-bg/40 rounded-xl border border-makoclaw-border/30">
              <p class="text-xs text-makoclaw-text-secondary font-semibold mb-1.5">Agent command that will be generated:</p>
              <code class="text-xs text-pink-400 font-mono break-all">{{ generatedCommand }}</code>
            </div>
          </div>

          <div class="p-5 border-t border-makoclaw-border/30 flex justify-end gap-3">
            <button class="px-4 py-2 text-sm bg-makoclaw-surface/50 border border-makoclaw-border/50 rounded-xl hover:bg-makoclaw-surface-hover transition-colors text-makoclaw-text-secondary" @click="showNewModal = false">Cancel</button>
            <button
              :disabled="!newCampaign.account || !newCampaign.campaign || creatingCampaign"
              class="px-5 py-2 text-sm bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl transition-all disabled:opacity-50 font-semibold shadow-lg shadow-pink-500/20"
              @click="createCampaignFromModal"
            >
              {{ creatingCampaign ? 'Creating...' : 'Create Campaign' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, onUnmounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { Bar, Doughnut, Line } from 'vue-chartjs'
import { Chart as ChartJS, Title, Tooltip, Legend, ArcElement, CategoryScale, LinearScale, BarElement, LineElement, PointElement } from 'chart.js'
import advancedService from '../services/advancedService'
import { useMarketingStore } from '../stores/marketingStore'
import { useAuthStore } from '../stores/authStore'
import { useToast } from '../composables/useToast'
import { buildFilesApiUrl } from '../utils/imagePreview'
import AudienceContacts from './marketing/AudienceContacts.vue'
import AudienceLists from './marketing/AudienceLists.vue'
import AudienceSegments from './marketing/AudienceSegments.vue'
import AudienceDeliveries from './marketing/AudienceDeliveries.vue'
import AudienceEmailCampaigns from './marketing/AudienceEmailCampaigns.vue'
import AudienceAutomations from './marketing/AudienceAutomations.vue'
import AudienceAccounts from './marketing/AudienceAccounts.vue'

ChartJS.register(Title, Tooltip, Legend, ArcElement, CategoryScale, LinearScale, BarElement, LineElement, PointElement)

const ScheduleContent = defineComponent({
  name: 'ScheduleContent',
  props: {
    path: { type: String, required: true }
  },
  setup(props) {
    const content = ref(null)
    const loading = ref(true)
    const error = ref(null)

    async function loadContent() {
      loading.value = true
      error.value = null

      try {
        const data = await advancedService.fetchFiles(props.path)
        const text = typeof data?.content === 'string'
          ? data.content
          : JSON.stringify(data?.content ?? '', null, 2)

        try {
          content.value = { type: 'json', data: JSON.parse(text) }
        } catch {
          content.value = { type: 'text', data: text }
        }
      } catch (e) {
        error.value = e.message
      } finally {
        loading.value = false
      }
    }

    watch(() => props.path, loadContent, { immediate: true })

    return () => {
      if (loading.value) {
        return h('div', { class: 'p-4 text-xs text-makoclaw-text-secondary' }, 'Loading...')
      }
      if (error.value) {
        return h('div', { class: 'p-4 text-xs text-red-400' }, `Error: ${error.value}`)
      }
      if (content.value?.type === 'json') {
        const data = content.value.data
        if (Array.isArray(data) && data.length > 0 && typeof data[0] === 'object') {
          const keys = Object.keys(data[0])
          return h('div', { class: 'overflow-x-auto' }, [
            h('table', { class: 'w-full text-xs' }, [
              h('thead', {}, [
                h('tr', { class: 'border-b border-makoclaw-border/30' }, keys.map((key) => h('th', { class: 'px-3 py-2 text-left text-makoclaw-text-secondary font-semibold' }, key)))
              ]),
              h('tbody', {}, data.map((row, index) => h('tr', { key: index, class: 'border-b border-makoclaw-border/10 hover:bg-makoclaw-surface/20' }, keys.map((key) => h('td', { class: 'px-3 py-2 text-makoclaw-text' }, String(row[key] ?? ''))))))
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
const toast = useToast()
const store = useMarketingStore()
const { campaigns, selectedCampaign, campaignDetail, templates, media, analyticsSummary, loadingList, loadingDetail, expandedAccounts } = storeToRefs(store)

const activeTab = ref('brief')
const previewedFile = ref(null)
const filePreviewContent = ref(null)
const lightboxFile = ref(null)
const showNewModal = ref(false)
const creatingCampaign = ref(false)
const isEditingBrief = ref(false)
const isEditingStrategy = ref(false)
const editContent = ref('')
const focusedAccount = ref('')
const templatesLoading = ref(false)
const savingTemplate = ref(false)
const showTemplateForm = ref(false)
const templateFormMode = ref('create')
const editingTemplateSlug = ref('')
const templateSlugTouched = ref(false)
const templatePreviewValues = ref({})
const mediaLoading = ref(false)
const mediaUploading = ref(false)
const showUploadForm = ref(false)
const mediaSearch = ref('')
const mediaUpload = ref({ file: null, tags: '', description: '' })
const copyingMedia = ref({})
const mediaCopyTargets = ref({})
const analyticsLoading = ref(false)
const engagementBreakdown = ref({ likes: 0, comments: 0, shares: 0 })
const audienceSubTab = ref('contacts')

const newCampaign = ref(createEmptyCampaign())
const templateForm = ref(createEmptyTemplateForm())

const platformOptions = ['Twitter', 'LinkedIn', 'Facebook', 'Instagram', 'TikTok', 'YouTube']
const statusOptions = ['draft', 'active', 'paused', 'completed', 'archived']

const groupedCampaigns = computed(() => {
  const groups = {}
  for (const campaign of campaigns.value) {
    if (!groups[campaign.account]) groups[campaign.account] = []
    groups[campaign.account].push(campaign)
  }
  return groups
})

const accounts = computed(() => Object.keys(groupedCampaigns.value))
const currentAccount = computed(() => selectedCampaign.value?.account || focusedAccount.value || '')
const accountCampaignOptions = computed(() => campaigns.value.filter((campaign) => campaign.account === currentAccount.value))

const availableTabs = computed(() => ([
  { id: 'brief', label: 'Brief' },
  { id: 'strategy', label: 'Strategy' },
  { id: 'copy', label: 'Copy' },
  { id: 'templates', label: 'Templates' },
  { id: 'assets', label: 'Assets' },
  { id: 'schedule', label: 'Schedule' },
  { id: 'media', label: 'Media Library' },
  { id: 'analytics', label: 'Analytics' },
  { id: 'audience', label: 'Audience' }
]))

const generatedCommand = computed(() => {
  const { account, campaign, objective, platforms, description } = newCampaign.value
  const platformStr = platforms.join(',') || 'twitter,linkedin'
  return `marketing_init_campaign(account="${account || 'account'}", campaign="${campaign || 'campaign-name'}", objective="${objective || 'your objective'}", platforms="${platformStr}", description="${description || 'Campaign description'}")`
})

const currentStatus = computed(() => campaignDetail.value?.status || selectedCampaign.value?.status || 'draft')

const templateVariables = computed(() => extractTemplateVariables(templateForm.value.body))
const renderedTemplatePreview = computed(() => renderTemplatePreview(templateForm.value.body, templatePreviewValues.value))

const filteredMedia = computed(() => {
  const term = mediaSearch.value.trim().toLowerCase()
  if (!term) return media.value
  return media.value.filter((item) => {
    const haystack = [
      item.filename,
      item.original_name,
      item.description,
      ...(item.tags || [])
    ].join(' ').toLowerCase()
    return haystack.includes(term)
  })
})

const analyticsTotals = computed(() => analyticsSummary.value || {
  total_posts: 0,
  total_impressions: 0,
  total_engagements: 0,
  engagement_rate: 0,
  by_platform: {},
  trend_dates: [],
  trend_impressions: [],
  has_legacy_reports: false
})

const analyticsHasEntries = computed(() => {
  const summary = analyticsTotals.value
  return summary.total_posts > 0 || summary.trend_dates?.length > 0 || Object.keys(summary.by_platform || {}).length > 0
})

const legacyAnalyticsFiles = computed(() => {
  const files = campaignDetail.value?.files?.analytics || []
  const legacy = files.filter((file) => !file.name.toLowerCase().endsWith('.json'))
  return legacy.length ? legacy : files
})

const showLegacyAnalyticsFallback = computed(() => !analyticsHasEntries.value && analyticsTotals.value.has_legacy_reports && legacyAnalyticsFiles.value.length > 0)

const analyticsKpis = computed(() => ([
  { label: 'Total Posts', value: formatNumber(analyticsTotals.value.total_posts) },
  { label: 'Total Impressions', value: formatNumber(analyticsTotals.value.total_impressions) },
  { label: 'Total Engagements', value: formatNumber(analyticsTotals.value.total_engagements) },
  { label: 'Engagement Rate', value: formatPercent(analyticsTotals.value.engagement_rate) }
]))

const analyticsBarData = computed(() => {
  const entries = Object.entries(analyticsTotals.value.by_platform || {})
  return {
    labels: entries.map(([platform]) => platform),
    datasets: [{
      data: entries.map(([, stats]) => stats?.impressions || 0),
      backgroundColor: entries.map((_, index) => index % 2 === 0 ? '#ec4899' : '#8b5cf6'),
      borderRadius: 8,
      barPercentage: 0.65
    }]
  }
})

const analyticsDoughnutData = computed(() => {
  const breakdown = engagementBreakdown.value
  const data = [breakdown.likes, breakdown.comments, breakdown.shares]
  const hasData = data.some((value) => value > 0)
  return hasData
    ? {
        labels: ['Likes', 'Comments', 'Shares'],
        datasets: [{
          data,
          backgroundColor: ['#ec4899', '#8b5cf6', '#c084fc'],
          borderColor: 'transparent',
          hoverOffset: 10
        }]
      }
    : { labels: [], datasets: [{ data: [] }] }
})

const analyticsLineData = computed(() => ({
  labels: analyticsTotals.value.trend_dates || [],
  datasets: [{
    label: 'Impressions',
    data: analyticsTotals.value.trend_impressions || [],
    borderColor: '#ec4899',
    backgroundColor: 'rgba(236, 72, 153, 0.18)',
    tension: 0.35,
    fill: true,
    pointBackgroundColor: '#8b5cf6',
    pointBorderColor: '#ffffff',
    pointRadius: 4
  }]
}))

const chartLegend = {
  position: 'bottom',
  labels: {
    color: 'rgba(156, 163, 175, 0.7)',
    font: { size: 10, weight: 'bold', family: 'Inter' },
    padding: 15,
    usePointStyle: true
  }
}

const analyticsDoughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: chartLegend },
  cutout: '72%'
}

const analyticsBarOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    y: {
      beginAtZero: true,
      grid: { color: 'rgba(255,255,255,0.03)', drawBorder: false },
      ticks: { color: 'rgba(156, 163, 175, 0.5)', font: { size: 10 } }
    },
    x: {
      grid: { display: false },
      ticks: { color: 'rgba(156, 163, 175, 0.6)', font: { size: 10, weight: 'bold' } }
    }
  }
}

const analyticsLineOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: chartLegend },
  scales: {
    y: {
      beginAtZero: true,
      grid: { color: 'rgba(255,255,255,0.03)', drawBorder: false },
      ticks: { color: 'rgba(156, 163, 175, 0.5)', font: { size: 10 } }
    },
    x: {
      grid: { color: 'rgba(255,255,255,0.03)', drawBorder: false },
      ticks: { color: 'rgba(156, 163, 175, 0.6)', font: { size: 10, weight: 'bold' } }
    }
  }
}

function createEmptyCampaign() {
  return {
    account: '',
    campaign: '',
    objective: '',
    platforms: [],
    description: ''
  }
}

function createEmptyTemplateForm() {
  return {
    name: '',
    slug: '',
    platform: '',
    body: ''
  }
}

function isSelected(campaign) {
  return selectedCampaign.value?.account === campaign.account && selectedCampaign.value?.campaign === campaign.campaign
}

async function handleSelectCampaign(campaign) {
  focusedAccount.value = campaign.account
  activeTab.value = 'brief'
  previewedFile.value = null
  filePreviewContent.value = null
  cancelEditing()
  await store.selectCampaign(campaign)
}

function handleAccountClick(account) {
  focusedAccount.value = account
  store.toggleAccount(account)
}

function selectAccountMedia(account) {
  focusedAccount.value = account
  mediaSearch.value = ''
}

function togglePlatform(platform) {
  const index = newCampaign.value.platforms.indexOf(platform)
  if (index >= 0) {
    newCampaign.value.platforms.splice(index, 1)
  } else {
    newCampaign.value.platforms.push(platform)
  }
}

function campaignWorkspacePath(fileName = '') {
  if (!selectedCampaign.value) return ''
  const base = `marketing/${selectedCampaign.value.account}/${selectedCampaign.value.campaign}`
  return fileName ? `${base}/${fileName}` : base
}

function fileWorkspacePath(file) {
  return campaignWorkspacePath(file.path)
}

function fileUrl(file) {
  return buildFilesApiUrl(fileWorkspacePath(file), authStore.token)
}

function mediaPreviewUrl(item) {
  return buildFilesApiUrl(item.path, authStore.token)
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
    const data = await advancedService.fetchFiles(fileWorkspacePath(file))
    filePreviewContent.value = typeof data?.content === 'string'
      ? data.content
      : JSON.stringify(data?.content ?? '', null, 2)
  } catch {
    filePreviewContent.value = 'Failed to load file content.'
  }
}

function formatSize(bytes) {
  if (!bytes) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatNumber(num) {
  if (!num) return '0'
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
  if (num >= 1000) return `${(num / 1000).toFixed(1)}k`
  return String(num)
}

function formatPercent(value) {
  return `${((value || 0) * 100).toFixed(1)}%`
}

function statusPillClass(status) {
  switch (status) {
    case 'active':
      return 'badge-success'
    case 'paused':
      return 'badge-warning'
    case 'completed':
      return 'badge-info'
    case 'archived':
      return 'px-2 py-0.5 text-xs font-medium rounded-full bg-slate-500/20 text-slate-300'
    case 'draft':
    default:
      return 'badge-neutral'
  }
}

function sanitizeSlug(value = '') {
  return String(value)
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
}

function extractTemplateVariables(content = '') {
  const matches = content.matchAll(/\{\{\s*([\w.:-]+)\s*\}\}/g)
  return [...new Set(Array.from(matches, (match) => match[1]))]
}

function renderTemplatePreview(content = '', values = {}) {
  return String(content || '').replace(/\{\{\s*([\w.:-]+)\s*\}\}/g, (_, variable) => values[variable] || `«${variable}»`)
}

function formatTemplateToken(variable = '') {
  return `{{${variable}}}`
}

function resetTemplateForm() {
  templateForm.value = createEmptyTemplateForm()
  editingTemplateSlug.value = ''
   templateSlugTouched.value = false
  templatePreviewValues.value = {}
}

function startNewTemplate() {
  templateFormMode.value = 'create'
  showTemplateForm.value = true
  resetTemplateForm()
}

function editTemplate(template) {
  templateFormMode.value = 'edit'
  showTemplateForm.value = true
  editingTemplateSlug.value = template.slug
  templateSlugTouched.value = true
  templateForm.value = {
    name: template.name || '',
    slug: template.slug || '',
    platform: template.platform || '',
    body: template.body || ''
  }
}

function cancelTemplateForm() {
  showTemplateForm.value = false
  resetTemplateForm()
}

async function submitTemplate() {
  const account = selectedCampaign.value?.account
  if (!account) return

  const name = templateForm.value.name.trim()
  const body = templateForm.value.body.trim()
  const platform = templateForm.value.platform.trim()
  const slug = sanitizeSlug(templateForm.value.slug || name)

  if (!name || !body) {
    toast.error('Template name and content are required')
    return
  }

  if (templateFormMode.value === 'create' && !slug) {
    toast.error('Template name must generate a valid slug')
    return
  }

  savingTemplate.value = true

  try {
    if (templateFormMode.value === 'create') {
      await store.createTemplate(account, { slug, name, body, platform })
      toast.success('Template created')
    } else {
      await store.updateTemplate(account, editingTemplateSlug.value, { name, body, platform })
      toast.success('Template updated')
    }
    cancelTemplateForm()
  } catch (error) {
    console.error('Failed to save template:', error)
    toast.error('Failed to save template')
  } finally {
    savingTemplate.value = false
  }
}

async function deleteTemplateItem(template) {
  const account = selectedCampaign.value?.account
  if (!account) return
  if (!window.confirm(`Delete template "${template.name}"?`)) return

  try {
    await store.deleteTemplate(account, template.slug)
    if (editingTemplateSlug.value === template.slug) cancelTemplateForm()
    toast.success('Template deleted')
  } catch (error) {
    console.error('Failed to delete template:', error)
    toast.error('Failed to delete template')
  }
}

function handleMediaFileChange(event) {
  mediaUpload.value.file = event.target.files?.[0] || null
}

function resetMediaUpload() {
  mediaUpload.value = { file: null, tags: '', description: '' }
}

async function uploadMedia() {
  const account = currentAccount.value
  if (!account || !mediaUpload.value.file) return

  const formData = new FormData()
  formData.append('file', mediaUpload.value.file)
  if (mediaUpload.value.tags.trim()) formData.append('tags', mediaUpload.value.tags.trim())
  if (mediaUpload.value.description.trim()) formData.append('description', mediaUpload.value.description.trim())

  mediaUploading.value = true
  try {
    await store.uploadMedia(account, formData)
    resetMediaUpload()
    showUploadForm.value = false
    toast.success('Media uploaded')
  } catch (error) {
    console.error('Failed to upload media:', error)
    toast.error('Failed to upload media')
  } finally {
    mediaUploading.value = false
  }
}

async function copyMedia(item) {
  const account = currentAccount.value
  const targetCampaign = mediaCopyTargets.value[item.filename]
  if (!account || !targetCampaign) return

  copyingMedia.value = { ...copyingMedia.value, [item.filename]: true }

  try {
    await store.copyMedia(account, item.filename, targetCampaign)
    toast.success(`Copied to ${targetCampaign}`)
  } catch (error) {
    console.error('Failed to copy media:', error)
    toast.error('Failed to copy media')
  } finally {
    copyingMedia.value = { ...copyingMedia.value, [item.filename]: false }
  }
}

async function deleteMediaItem(item) {
  const account = currentAccount.value
  if (!account) return
  if (!window.confirm(`Delete media \"${item.original_name || item.filename}\"?`)) return

  try {
    await store.deleteMedia(account, item.filename)
    delete mediaCopyTargets.value[item.filename]
    toast.success('Media deleted')
  } catch (error) {
    console.error('Failed to delete media:', error)
    toast.error('Failed to delete media')
  }
}

function hasUnsavedTemplateChanges() {
  if (!showTemplateForm.value) return false
  const { name, slug, platform, body } = templateForm.value
  return Boolean(name.trim() || slug.trim() || platform.trim() || body.trim())
}

function requestTabSwitch(nextTab) {
  if (nextTab === activeTab.value) return
  if (hasUnsavedTemplateChanges() && !window.confirm('You have unsaved template changes. Switch tabs and discard them?')) return

  if (activeTab.value === 'templates' && nextTab !== 'templates' && showTemplateForm.value) {
    cancelTemplateForm()
  }

  activeTab.value = nextTab
}

async function createCampaignFromModal() {
  creatingCampaign.value = true

  try {
    await store.createCampaign({ ...newCampaign.value })
    showNewModal.value = false
    focusedAccount.value = newCampaign.value.account
    newCampaign.value = createEmptyCampaign()
    toast.success('Campaign created')
  } catch (error) {
    console.error('Failed to create campaign:', error)
    toast.error('Failed to create campaign')
  } finally {
    creatingCampaign.value = false
  }
}

async function handleDeleteCampaign() {
  if (!selectedCampaign.value) return

  const { account, campaign } = selectedCampaign.value
  if (!window.confirm(`Delete campaign "${campaign}" from ${account}?`)) return

  try {
    await store.deleteCampaign(account, campaign)
    activeTab.value = 'brief'
    previewedFile.value = null
    filePreviewContent.value = null
    toast.success('Campaign deleted')
  } catch (error) {
    console.error('Failed to delete campaign:', error)
    toast.error('Failed to delete campaign')
  }
}

async function handleStatusChange(status) {
  if (!selectedCampaign.value || status === currentStatus.value) return

  try {
    await store.updateCampaignStatus(selectedCampaign.value.account, selectedCampaign.value.campaign, status)
    toast.success(`Status updated to ${status}`)
  } catch (error) {
    console.error('Failed to update campaign status:', error)
    toast.error('Failed to update campaign status')
  }
}

function startEditingBrief() {
  isEditingStrategy.value = false
  editContent.value = campaignDetail.value?.brief || ''
  isEditingBrief.value = true
}

function startEditingStrategy() {
  isEditingBrief.value = false
  editContent.value = campaignDetail.value?.strategy || ''
  isEditingStrategy.value = true
}

function cancelEditing() {
  isEditingBrief.value = false
  isEditingStrategy.value = false
  editContent.value = ''
}

async function saveCampaignText(path, field, successMessage) {
  try {
    await advancedService.updateFileContent(path, editContent.value)
    await store.fetchCampaignDetail(selectedCampaign.value)
    if (campaignDetail.value) {
      campaignDetail.value[field] = editContent.value
    }
    cancelEditing()
    toast.success(successMessage)
  } catch (error) {
    console.error(`Failed to save ${field}:`, error)
    toast.error(`Failed to save ${field}`)
  }
}

async function saveBrief() {
  await saveCampaignText(campaignWorkspacePath('brief.md'), 'brief', 'Brief saved')
}

async function saveStrategy() {
  await saveCampaignText(campaignWorkspacePath('strategy.md'), 'strategy', 'Strategy saved')
}

async function loadTemplates(account) {
  if (!account) {
    templates.value = []
    return
  }

  templatesLoading.value = true
  try {
    await store.fetchTemplates(account)
  } catch {
    toast.error('Failed to load templates')
  } finally {
    templatesLoading.value = false
  }
}

async function loadMedia(account) {
  if (!account) {
    media.value = []
    return
  }

  mediaLoading.value = true
  try {
    await store.fetchMedia(account)
  } catch {
    toast.error('Failed to load media')
  } finally {
    mediaLoading.value = false
  }
}

async function loadAnalyticsSummary() {
  if (!selectedCampaign.value) {
    analyticsSummary.value = null
    engagementBreakdown.value = { likes: 0, comments: 0, shares: 0 }
    return
  }

  analyticsLoading.value = true
  try {
    await Promise.all([
      store.fetchAnalyticsSummary(selectedCampaign.value.account, selectedCampaign.value.campaign),
      loadEngagementBreakdown()
    ])
  } catch {
    toast.error('Failed to load analytics summary')
  } finally {
    analyticsLoading.value = false
  }
}

async function loadEngagementBreakdown() {
  const files = campaignDetail.value?.files?.analytics || []
  const totals = { likes: 0, comments: 0, shares: 0 }

  await Promise.all(files
    .filter((file) => file.name.toLowerCase().endsWith('.json'))
    .map(async (file) => {
      try {
        const payload = await advancedService.fetchFiles(fileWorkspacePath(file))
        const content = typeof payload?.content === 'string'
          ? JSON.parse(payload.content)
          : payload?.content
        totals.likes += Number(content?.likes || 0)
        totals.comments += Number(content?.comments || 0)
        totals.shares += Number(content?.shares || 0)
      } catch {
        // Ignore malformed analytics files in the UI aggregation.
      }
    }))

  engagementBreakdown.value = totals
}

async function loadAudienceData() {
  await Promise.allSettled([
    store.fetchAudienceContacts({ page: store.audienceContactsPage }),
    store.fetchAudienceLists(),
    store.fetchAudienceSegments(),
    store.fetchAudienceDeliveries(),
    store.fetchEmailCampaigns(),
    store.fetchAutomations()
  ])
}


watch(templateVariables, (variables) => {
  const next = {}
  variables.forEach((variable) => {
    next[variable] = templatePreviewValues.value[variable] || ''
  })
  templatePreviewValues.value = next
}, { immediate: true })

watch(() => templateForm.value.name, (name) => {
  if (templateFormMode.value !== 'create' || templateSlugTouched.value) return
  templateForm.value.slug = sanitizeSlug(name)
})

watch(currentAccount, async (account) => {
  if (!account) {
    templates.value = []
    media.value = []
    return
  }
  await Promise.allSettled([loadTemplates(account), loadMedia(account)])
}, { immediate: true })

watch(selectedCampaign, async (campaign) => {
  previewedFile.value = null
  filePreviewContent.value = null
  cancelEditing()
  if (campaign?.account) {
    focusedAccount.value = campaign.account
    await loadAnalyticsSummary()
  } else {
    analyticsSummary.value = null
    engagementBreakdown.value = { likes: 0, comments: 0, shares: 0 }
  }
})

watch(activeTab, async (tab) => {
  if (tab === 'analytics' && selectedCampaign.value) {
    await loadAnalyticsSummary()
  }
  if (tab === 'audience') {
    await loadAudienceData()
  }
})

onMounted(async () => {
  await store.fetchCampaigns()
  if (!focusedAccount.value && accounts.value.length) {
    focusedAccount.value = accounts.value[0]
  }
  store.startPolling()
})

onUnmounted(() => {
  store.stopPolling()
})
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.15);
  border-radius: 10px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.25);
}
</style>
