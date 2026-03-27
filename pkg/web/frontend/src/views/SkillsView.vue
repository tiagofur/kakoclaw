<template>
  <div class="flex flex-col h-full bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-purple-500/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-pink-500/20 via-transparent to-transparent" />
    </div>

    <!-- Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
        <!-- Title Row -->
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-purple-500/10">
            <svg
              class="w-5 h-5 sm:w-6 sm:h-6 text-purple-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M13 10V3L4 14h7v7l9-11h-7z"
              />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-purple-400 bg-clip-text text-transparent">
              Skills
            </h1>
            <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5 hidden sm:block">
              Extend your agent with powerful capabilities
            </p>
          </div>
          <button
            class="px-4 sm:px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-purple-500 to-purple-600 hover:from-purple-600 hover:to-purple-700 text-white rounded-xl transition-all shadow-lg shadow-purple-500/25 hover:shadow-purple-500/40 text-sm font-bold flex items-center gap-2 active:scale-95 flex-shrink-0"
            @click="openGenerateModal"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              />
            </svg>
            <span class="hidden sm:inline">Create Skill</span>
          </button>
        </div>
      </div>

      <!-- Tabs Row -->
      <div class="px-4 sm:px-6 pb-3 sm:pb-4">
        <div class="flex items-center gap-1 p-1 bg-makoclaw-bg/50 rounded-xl border border-makoclaw-border/30 overflow-x-auto scrollbar-hide">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            :class="[
              'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all whitespace-nowrap',
              activeTab === tab.id
                ? 'bg-makoclaw-surface shadow-sm text-makoclaw-text border border-makoclaw-border/50'
                : 'text-makoclaw-text-secondary hover:text-makoclaw-text hover:bg-makoclaw-surface/50'
            ]"
            @click="activeTab = tab.id; tab.load?.()"
          >
            <component
              :is="tab.icon"
              class="w-4 h-4"
            />
            {{ tab.label }}
            <span
              v-if="tab.count !== undefined"
              class="text-[10px] px-1.5 py-0.5 rounded-full bg-makoclaw-accent/10 text-makoclaw-accent font-medium"
            >
              {{ tab.count }}
            </span>
          </button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-4 sm:p-6 custom-scrollbar relative">
      <!-- Loading State -->
      <div
        v-if="loading"
        class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
      >
        <div
          v-for="i in 6"
          :key="i"
          class="skill-card-skeleton"
        >
          <div class="flex items-start gap-3 mb-3">
            <div class="skeleton w-10 h-10 rounded-xl" />
            <div class="flex-1 space-y-2">
              <div class="skeleton h-4 w-32 rounded" />
              <div class="skeleton h-3 w-full rounded" />
            </div>
          </div>
          <div class="flex gap-2 mt-auto">
            <div class="skeleton h-8 w-16 rounded-lg" />
            <div class="skeleton h-8 w-16 rounded-lg" />
          </div>
        </div>
      </div>

      <template v-else>
        <Transition
          name="fade"
          mode="out-in"
        >
          <!-- Installed Skills Tab -->
          <div
            v-if="activeTab === 'installed'"
            key="installed"
            class="animate-fadeIn"
          >
            <!-- Empty State -->
            <div
              v-if="skills.length === 0"
              class="flex flex-col items-center justify-center py-12 text-center"
            >
              <div class="relative">
                <div class="absolute inset-0 bg-gradient-to-br from-purple-500/30 to-pink-500/30 rounded-3xl blur-2xl opacity-50" />
                <div class="relative glass-panel p-8 rounded-2xl shadow-2xl ring-1 ring-white/10">
                  <div class="w-16 h-16 mx-auto rounded-2xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 flex items-center justify-center ring-1 ring-white/20">
                    <svg
                      class="w-8 h-8 text-purple-400"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.5"
                        d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                      />
                    </svg>
                  </div>
                </div>
              </div>
              <h3 class="text-lg font-bold text-makoclaw-text mt-6">
                No skills installed
              </h3>
              <p class="text-sm text-makoclaw-text-secondary/70 mt-2 max-w-xs">
                Browse the marketplace to discover and install powerful skills for your agent.
              </p>
              <button
                class="mt-6 px-5 py-2.5 bg-gradient-to-r from-purple-500 to-purple-600 hover:from-purple-600 hover:to-purple-700 text-white rounded-xl font-bold shadow-lg shadow-purple-500/25 transition-all active:scale-95 flex items-center gap-2"
                @click="activeTab = 'marketplace'; loadMarketplace()"
              >
                Browse Marketplace
              </button>
            </div>

            <!-- Skills Grid -->
            <div
              v-else
              class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
            >
              <div
                v-for="skill in skills"
                :key="skill.name"
                class="skill-card group"
              >
                <!-- Card Header -->
                <div class="flex items-start gap-3 mb-3">
                  <div class="skill-icon">
                    <svg
                      class="w-5 h-5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.5"
                        d="M13 10V3L4 14h7v7l9-11h-7z"
                      />
                    </svg>
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2 mb-1">
                      <h3 class="font-semibold text-makoclaw-text truncate">
                        {{ skill.name }}
                      </h3>
                      <span :class="sourceClasses(skill.source)">{{ skill.source }}</span>
                    </div>
                    <p class="text-sm text-makoclaw-text-secondary line-clamp-2">
                      {{ skill.description || 'No description' }}
                    </p>
                  </div>
                </div>

                <!-- Disabled Security Badge -->
                <div
                  v-if="isDisabled(skill)"
                  class="mt-2 flex items-center gap-2 bg-red-500/10 border border-red-500/20 rounded-lg px-3 py-2"
                >
                  <span class="text-red-400 text-xs font-black uppercase tracking-widest">⚠ Deshabilitada</span>
                  <span class="text-red-300/70 text-xs">Esta skill fue desactivada por razones de seguridad. Se recomienda desinstalarla.</span>
                </div>

                <!-- Usage Stats -->
                <div
                  v-if="getUsageCount(skill.name) > 0"
                  class="flex items-center gap-2 text-xs text-makoclaw-text-secondary mb-3"
                >
                  <svg
                    class="w-3.5 h-3.5"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
                    />
                  </svg>
                  <span>Used {{ getUsageCount(skill.name) }}x</span>
                </div>

                <!-- Actions -->
                <div class="flex items-center gap-2 mt-auto pt-3 border-t border-makoclaw-border/30">
                  <button
                    class="skill-action-btn"
                    @click="viewSkill(skill.name)"
                  >
                    <svg
                      class="w-3.5 h-3.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                      />
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                      />
                    </svg>
                    View
                  </button>
                  <button
                    v-if="skill.source === 'workspace' || skill.source === 'user'"
                    class="skill-action-btn text-makoclaw-accent"
                    @click="editSkill(skill.name)"
                  >
                    <svg
                      class="w-3.5 h-3.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                      />
                    </svg>
                    Edit
                  </button>
                  <button
                    v-if="skill.source === 'workspace' || skill.source === 'user'"
                    class="skill-action-btn text-emerald-400"
                    @click="openSubmitModal(skill)"
                  >
                    <svg
                      class="w-3.5 h-3.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                      />
                    </svg>
                    Share
                  </button>
                  <button
                    v-if="skill.source === 'workspace' || skill.source === 'user'"
                    class="skill-action-btn text-makoclaw-error ml-auto"
                    @click="confirmUninstall(skill.name)"
                  >
                    <svg
                      class="w-3.5 h-3.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                      />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Marketplace Tab -->
          <div
            v-else-if="activeTab === 'marketplace'"
            key="marketplace"
            class="animate-fadeIn"
          >
            <div
              v-if="loadingMarketplace"
              class="flex items-center justify-center py-16"
            >
              <div class="flex flex-col items-center gap-3">
                <div class="w-10 h-10 rounded-full border-2 border-makoclaw-accent/30 border-t-makoclaw-accent animate-spin" />
                <span class="text-sm text-makoclaw-text-secondary">Loading marketplace...</span>
              </div>
            </div>

            <div
              v-else-if="marketplaceSkills.length === 0"
              class="empty-state py-16"
            >
              <div class="w-20 h-20 rounded-2xl bg-gradient-to-br from-purple-500/10 to-makoclaw-accent/10 flex items-center justify-center mb-4 border border-white/5">
                <svg
                  class="w-10 h-10 text-purple-400/60"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="1.5"
                    d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"
                  />
                </svg>
              </div>
              <h3 class="empty-state-title">
                Marketplace is empty
              </h3>
              <p class="empty-state-text">
                Be the first to share your skills with the community!
              </p>
            </div>

            <template v-else>
              <!-- Trending Section -->
              <div
                v-if="trendingSkills.length > 0"
                class="mb-6"
              >
                <div class="flex items-center gap-2 mb-3">
                  <svg
                    class="w-4 h-4 text-amber-400"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path d="M12.395 2.553a1 1 0 00-1.45-.385c-.345.23-.614.558-.822.88-.214.33-.403.713-.57 1.116-.334.804-.614 1.768-.84 2.734a31.365 31.365 0 00-.613 3.58 2.64 2.64 0 01-.945-1.067c-.328-.68-.398-1.534-.398-2.654A1 1 0 005.05 6.05 6.981 6.981 0 003 11a7 7 0 1011.95-4.95c-.592-.591-.98-.985-1.348-1.467-.363-.476-.724-1.063-1.207-2.03zM12.12 15.12A3 3 0 017 13s.879.5 2.5.5c0-1 .5-4 1.25-4.5.5 1 .786 1.293 1.371 1.879A2.99 2.99 0 0113 13a2.99 2.99 0 01-.879 2.121z" />
                  </svg>
                  <h3 class="text-sm font-semibold text-makoclaw-text">
                    Trending
                  </h3>
                </div>
                <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
                  <div
                    v-for="skill in trendingSkills"
                    :key="'trending-' + skill.slug"
                    class="trending-card"
                  >
                    <div class="flex items-center gap-2 mb-2">
                      <span class="text-amber-400 text-xs">{{ skill.install_count }} installs</span>
                    </div>
                    <h4 class="font-medium text-sm text-makoclaw-text">
                      {{ skill.name }}
                    </h4>
                    <p class="text-xs text-makoclaw-text-secondary mt-1 line-clamp-2">
                      {{ skill.description }}
                    </p>
                  </div>
                </div>
              </div>

              <!-- All Skills Grid -->
              <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                <div
                  v-for="skill in marketplaceSkills"
                  :key="skill.slug || skill.name"
                  class="marketplace-card group"
                >
                  <!-- Header -->
                  <div class="flex items-start justify-between gap-3 mb-3">
                    <div class="flex-1 min-w-0">
                      <h3 class="font-semibold text-makoclaw-text truncate mb-1">
                        {{ skill.name }}
                      </h3>
                      <p class="text-xs text-makoclaw-text-secondary">
                        by {{ skill.author || 'unknown' }}
                      </p>
                    </div>
                    <div class="flex items-center gap-1.5">
                      <span
                        v-if="skill.source === 'builtin'"
                        class="px-2 py-0.5 text-[10px] font-medium rounded-full bg-amber-500/10 text-amber-400"
                      >
                        Built-in
                      </span>
                      <div
                        v-if="skill.security_score !== undefined && skill.source !== 'builtin'"
                        :class="securityScoreClasses(skill.security_score)"
                      >
                        {{ skill.security_score }}
                      </div>
                    </div>
                  </div>

                  <!-- Description -->
                  <p class="text-sm text-makoclaw-text-secondary line-clamp-2 mb-3">
                    {{ skill.description }}
                  </p>

                  <!-- Tags -->
                  <div
                    v-if="skill.tags && skill.tags.length"
                    class="flex flex-wrap gap-1.5 mb-3"
                  >
                    <span
                      v-for="tag in skill.tags.slice(0, 3)"
                      :key="tag"
                      class="skill-tag"
                    >{{ tag }}</span>
                    <span
                      v-if="skill.tags.length > 3"
                      class="skill-tag"
                    >+{{ skill.tags.length - 3 }}</span>
                  </div>

                  <!-- Stats & Actions -->
                  <div class="flex items-center justify-between mt-auto pt-3 border-t border-makoclaw-border/30">
                    <div class="flex items-center gap-3 text-xs text-makoclaw-text-secondary">
                      <span
                        v-if="skill.install_count"
                        class="flex items-center gap-1"
                      >
                        <svg
                          class="w-3.5 h-3.5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                          />
                        </svg>
                        {{ skill.install_count }}
                      </span>
                      <span
                        v-if="skill.rating_count > 0"
                        class="flex items-center gap-1 text-amber-400"
                      >
                        <svg
                          class="w-3.5 h-3.5 fill-current"
                          viewBox="0 0 20 20"
                        >
                          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                        </svg>
                        {{ skill.average_rating?.toFixed(1) ?? '—' }}
                      </span>
                    </div>
                    <div class="flex items-center gap-2">
                      <button
                        :disabled="forking === skill.slug"
                        class="skill-action-btn opacity-0 group-hover:opacity-100"
                        title="Fork to workspace"
                        @click="forkSkill(skill)"
                      >
                        <svg
                          class="w-3.5 h-3.5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                          />
                        </svg>
                      </button>
                      <span
                        v-if="skill.installed"
                        class="px-2.5 py-1 text-[10px] font-bold rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                      >
                        Installed
                      </span>
                      <button
                        v-else
                        :disabled="installing === skill.slug"
                        class="install-btn"
                        @click="installMarketplaceSkill(skill.slug)"
                      >
                        <svg
                          v-if="installing === skill.slug"
                          class="w-3.5 h-3.5 animate-spin"
                          fill="none"
                          viewBox="0 0 24 24"
                        >
                          <circle
                            class="opacity-25"
                            cx="12"
                            cy="12"
                            r="10"
                            stroke="currentColor"
                            stroke-width="4"
                          />
                          <path
                            class="opacity-75"
                            fill="currentColor"
                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                          />
                        </svg>
                        <span v-else>Install</span>
                      </button>
                    </div>
                  </div>

                  <!-- Rating Widget (expandable) -->
                  <Transition name="expand">
                    <div
                      v-if="ratingOpenSlug === skill.slug"
                      class="rating-widget"
                    >
                      <p class="text-xs font-medium text-makoclaw-text-secondary mb-2">
                        Rate this skill
                      </p>
                      <div class="flex gap-1 mb-2">
                        <button
                          v-for="star in 5"
                          :key="star"
                          class="text-xl leading-none transition-colors hover:scale-110"
                          :class="(pendingRating[skill.slug]?.stars ?? 0) >= star ? 'text-amber-400' : 'text-makoclaw-border'"
                          @click="setStars(skill.slug, star)"
                        >
                          &#9733;
                        </button>
                      </div>
                      <textarea
                        v-model="pendingRating[skill.slug].review"
                        placeholder="Optional review..."
                        maxlength="500"
                        rows="2"
                        class="input-field text-xs resize-none"
                      />
                      <div class="flex justify-end gap-2 mt-2">
                        <button
                          class="btn-ghost text-xs py-1.5"
                          @click="ratingOpenSlug = null"
                        >
                          Cancel
                        </button>
                        <button
                          :disabled="submittingRating || !(pendingRating[skill.slug]?.stars > 0)"
                          class="btn-primary text-xs py-1.5"
                          @click="submitRating(skill)"
                        >
                          {{ submittingRating ? 'Submitting...' : 'Submit' }}
                        </button>
                      </div>
                    </div>
                  </Transition>

                  <!-- Rate Button -->
                  <button
                    v-if="ratingOpenSlug !== skill.slug"
                    class="w-full mt-3 text-xs text-makoclaw-text-secondary hover:text-amber-400 transition-colors py-2 rounded-lg hover:bg-makoclaw-bg/50"
                    @click="toggleRatingWidget(skill.slug)"
                  >
                    Rate this skill
                  </button>
                </div>
              </div>
            </template>
          </div>

          <!-- Bundles Tab -->
          <div
            v-else-if="activeTab === 'bundles'"
            key="bundles"
            class="animate-fadeIn"
          >
            <div
              v-if="loadingBundles"
              class="flex items-center justify-center py-16"
            >
              <div class="flex flex-col items-center gap-3">
                <div class="w-10 h-10 rounded-full border-2 border-makoclaw-accent/30 border-t-makoclaw-accent animate-spin" />
                <span class="text-sm text-makoclaw-text-secondary">Loading bundles...</span>
              </div>
            </div>

            <div
              v-else-if="bundles.length === 0"
              class="empty-state py-16"
            >
              <div class="w-20 h-20 rounded-2xl bg-gradient-to-br from-purple-500/10 to-makoclaw-accent/10 flex items-center justify-center mb-4 border border-white/5">
                <svg
                  class="w-10 h-10 text-purple-400/60"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="1.5"
                    d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                  />
                </svg>
              </div>
              <h3 class="empty-state-title">
                No bundles available
              </h3>
              <p class="empty-state-text">
                Bundles are curated skill collections for one-click setup.
              </p>
            </div>

            <div
              v-else
              class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
            >
              <div
                v-for="bundle in bundles"
                :key="bundle.slug"
                class="bundle-card"
              >
                <div class="flex items-start gap-3 mb-3">
                  <span class="text-3xl">{{ bundle.icon }}</span>
                  <div class="flex-1 min-w-0">
                    <h3 class="font-semibold text-makoclaw-text">
                      {{ bundle.name }}
                    </h3>
                    <p class="text-xs text-makoclaw-text-secondary mt-1 line-clamp-2">
                      {{ bundle.description }}
                    </p>
                  </div>
                </div>
                <div class="flex items-center gap-3 text-xs text-makoclaw-text-secondary mb-4">
                  <span class="flex items-center gap-1">
                    <svg
                      class="w-3.5 h-3.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M13 10V3L4 14h7v7l9-11h-7z"
                      />
                    </svg>
                    {{ bundle.skill_slugs?.length ?? 0 }} skills
                  </span>
                  <span class="flex items-center gap-1">
                    <svg
                      class="w-3.5 h-3.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                      />
                    </svg>
                    {{ bundle.install_count }} installs
                  </span>
                </div>
                <button
                  :disabled="installingBundle === bundle.slug"
                  class="w-full btn-primary"
                  @click="installBundle(bundle)"
                >
                  <span
                    v-if="installingBundle === bundle.slug"
                    class="flex items-center justify-center gap-2"
                  >
                    <svg
                      class="w-4 h-4 animate-spin"
                      fill="none"
                      viewBox="0 0 24 24"
                    >
                      <circle
                        class="opacity-25"
                        cx="12"
                        cy="12"
                        r="10"
                        stroke="currentColor"
                        stroke-width="4"
                      />
                      <path
                        class="opacity-75"
                        fill="currentColor"
                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                      />
                    </svg>
                    Installing...
                  </span>
                  <span v-else>Install Bundle</span>
                </button>
              </div>
            </div>
          </div>

          <!-- My Submissions Tab -->
          <div
            v-else-if="activeTab === 'submissions'"
            key="submissions"
            class="animate-fadeIn"
          >
            <div
              v-if="loadingSubmissions"
              class="flex items-center justify-center py-16"
            >
              <div class="flex flex-col items-center gap-3">
                <div class="w-10 h-10 rounded-full border-2 border-makoclaw-accent/30 border-t-makoclaw-accent animate-spin" />
                <span class="text-sm text-makoclaw-text-secondary">Loading submissions...</span>
              </div>
            </div>

            <div
              v-else-if="mySubmissions.length === 0"
              class="empty-state py-16"
            >
              <div class="w-20 h-20 rounded-2xl bg-gradient-to-br from-purple-500/10 to-makoclaw-accent/10 flex items-center justify-center mb-4 border border-white/5">
                <svg
                  class="w-10 h-10 text-purple-400/60"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="1.5"
                    d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                  />
                </svg>
              </div>
              <h3 class="empty-state-title">
                No submissions yet
              </h3>
              <p class="empty-state-text">
                Share your skills with the community by submitting them to the marketplace.
              </p>
            </div>

            <div
              v-else
              class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
            >
              <div
                v-for="sub in mySubmissions"
                :key="sub.id"
                class="submission-card"
              >
                <div class="flex items-start justify-between gap-3 mb-2">
                  <h3 class="font-semibold text-makoclaw-text truncate">
                    {{ sub.skill_name }}
                  </h3>
                  <div class="flex items-center gap-1.5 flex-shrink-0">
                    <span
                      v-if="sub.visibility === 'private'"
                      class="badge-neutral"
                      title="Private"
                    >
                      <svg
                        class="w-3 h-3"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                        />
                      </svg>
                    </span>
                    <span :class="statusClasses(sub.status)">{{ sub.status }}</span>
                  </div>
                </div>
                <p class="text-sm text-makoclaw-text-secondary line-clamp-2 mb-3">
                  {{ sub.description }}
                </p>
                <div class="flex items-center gap-2 text-xs text-makoclaw-text-secondary">
                  <span>Security: {{ sub.security_score }}/100</span>
                </div>
                <div
                  v-if="sub.reviewer_notes"
                  class="mt-3 p-2 bg-amber-500/10 border border-amber-500/20 rounded-lg text-xs text-amber-300"
                >
                  {{ sub.reviewer_notes }}
                </div>
              </div>
            </div>
          </div>
        </Transition>
      </template>
    </div>

    <!-- Create Skill Modal -->
    <Transition name="modal">
      <div
        v-if="showGenerateModal"
        class="modal-backdrop"
        @click.self="closeGenerateModal"
      >
        <div class="modal-content max-w-3xl">
          <div class="modal-header">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500/20 to-makoclaw-accent/20 flex items-center justify-center">
                <svg
                  class="w-5 h-5 text-purple-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 4v16m8-8H4"
                  />
                </svg>
              </div>
              <div>
                <h2 class="font-bold text-lg text-makoclaw-text">
                  Create New Skill
                </h2>
                <p class="text-xs text-makoclaw-text-secondary">
                  Build powerful capabilities for your agent
                </p>
              </div>
            </div>
            <button
              class="modal-close-btn"
              @click="closeGenerateModal"
            >
              <svg
                class="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          <div class="modal-body custom-scrollbar">
            <!-- Step 1: Input Form -->
            <div
              v-if="!generatedPreview"
              class="space-y-6"
            >
              <!-- AI Generation Section -->
              <div class="ai-generate-section">
                <div class="flex items-start gap-4">
                  <div class="ai-icon">
                    <svg
                      class="w-6 h-6 text-purple-400"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.5"
                        d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                      />
                    </svg>
                  </div>
                  <div class="flex-1">
                    <h3 class="font-semibold text-makoclaw-text mb-1 flex items-center gap-2">
                      Generate with AI
                      <span class="px-2 py-0.5 text-[10px] font-bold uppercase bg-purple-500/20 text-purple-300 rounded-full">Recommended</span>
                    </h3>
                    <p class="text-sm text-makoclaw-text-secondary mb-3">
                      Describe what you want and AI will create the skill configuration.
                    </p>
                    <textarea
                      v-model="aiPrompt"
                      rows="3"
                      placeholder="e.g., 'Create a skill that helps me review pull requests on GitHub quickly.'"
                      class="input-field resize-none"
                    />
                    <button
                      :disabled="!aiPrompt.trim() || aiGenerating"
                      class="w-full mt-3 py-3 bg-gradient-to-r from-purple-500 to-makoclaw-accent text-white rounded-xl font-semibold shadow-lg shadow-purple-500/20 hover:shadow-purple-500/30 transition-all disabled:opacity-50 flex items-center justify-center gap-2"
                      @click="generateSkillConfigWithAI"
                    >
                      <svg
                        v-if="aiGenerating"
                        class="animate-spin w-4 h-4"
                        fill="none"
                        viewBox="0 0 24 24"
                      >
                        <circle
                          class="opacity-25"
                          cx="12"
                          cy="12"
                          r="10"
                          stroke="currentColor"
                          stroke-width="4"
                        />
                        <path
                          class="opacity-75"
                          fill="currentColor"
                          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                        />
                      </svg>
                      <svg
                        v-else
                        class="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M13 10V3L4 14h7v7l9-11h-7z"
                        />
                      </svg>
                      {{ aiGenerating ? 'Generating...' : 'Generate with AI' }}
                    </button>
                  </div>
                </div>
              </div>

              <!-- Divider -->
              <div class="flex items-center gap-4">
                <div class="flex-1 h-px bg-makoclaw-border/50" />
                <span class="text-xs font-medium uppercase tracking-wider text-makoclaw-text-secondary">Or configure manually</span>
                <div class="flex-1 h-px bg-makoclaw-border/50" />
              </div>

              <!-- Manual Form -->
              <div class="space-y-4">
                <div>
                  <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Skill Name <span class="text-makoclaw-error">*</span></label>
                  <input
                    v-model="generateForm.name"
                    type="text"
                    placeholder="e.g., jira-assistant, code-reviewer"
                    class="input-field"
                    :class="{ 'border-makoclaw-error': generateFormErrors.name }"
                    @input="validateSkillName"
                  >
                  <p
                    v-if="generateFormErrors.name"
                    class="text-xs text-makoclaw-error mt-1"
                  >
                    {{ generateFormErrors.name }}
                  </p>
                  <p
                    v-else
                    class="text-xs text-makoclaw-text-secondary mt-1"
                  >
                    Lowercase letters, numbers, and hyphens only
                  </p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Description <span class="text-makoclaw-error">*</span></label>
                  <textarea
                    v-model="generateForm.description"
                    rows="3"
                    placeholder="Describe what this skill does and when to use it..."
                    class="input-field resize-none"
                    :class="{ 'border-makoclaw-error': generateFormErrors.description }"
                  />
                  <p
                    v-if="generateFormErrors.description"
                    class="text-xs text-makoclaw-error mt-1"
                  >
                    {{ generateFormErrors.description }}
                  </p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Additional Context <span class="text-makoclaw-text-secondary font-normal">(optional)</span></label>
                  <textarea
                    v-model="generateForm.prompt"
                    rows="4"
                    placeholder="Commands, examples, constraints, required tools..."
                    class="input-field resize-none"
                  />
                </div>
              </div>

              <p
                v-if="generateError"
                class="text-sm text-makoclaw-error bg-makoclaw-error/10 p-3 rounded-lg border border-makoclaw-error/20"
              >
                {{ generateError }}
              </p>
            </div>

            <!-- Step 2: Preview -->
            <div
              v-else
              class="space-y-4"
            >
              <div class="flex items-center justify-between">
                <p class="text-sm text-makoclaw-text-secondary">
                  Review and edit the generated SKILL.md:
                </p>
                <button
                  class="text-sm text-makoclaw-accent hover:underline"
                  @click="generatedPreview = ''"
                >
                  Back to form
                </button>
              </div>
              <textarea
                v-model="generatedPreview"
                rows="20"
                class="input-field font-mono text-xs resize-none"
                spellcheck="false"
              />
            </div>
          </div>

          <div class="modal-footer">
            <button
              class="btn-ghost"
              @click="closeGenerateModal"
            >
              Cancel
            </button>
            <button
              v-if="!generatedPreview"
              :disabled="generating || !generateForm.name || !generateForm.description"
              class="btn-primary flex items-center gap-2"
              @click="handleGenerate"
            >
              <svg
                v-if="generating"
                class="w-4 h-4 animate-spin"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                />
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                />
              </svg>
              {{ generating ? 'Generating...' : 'Generate Skill' }}
            </button>
            <button
              v-else
              :disabled="savingGenerated || !generatedPreview.trim()"
              class="px-5 py-2 text-sm font-semibold bg-emerald-500 text-white rounded-lg hover:bg-emerald-400 transition-all flex items-center gap-2 disabled:opacity-50"
              @click="handleSaveGenerated()"
            >
              <svg
                v-if="savingGenerated"
                class="w-4 h-4 animate-spin"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                />
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                />
              </svg>
              {{ savingGenerated ? 'Saving...' : 'Save Skill' }}
            </button>
          </div>

          <!-- Overwrite confirmation -->
          <div
            v-if="showOverwriteConfirm"
            class="flex items-center justify-between p-3 bg-amber-500/10 border border-amber-500/20 rounded-xl"
          >
            <span class="text-sm text-amber-300">A skill with this name already exists. Overwrite?</span>
            <div class="flex gap-2">
              <button
                class="px-3 py-1 text-xs font-medium bg-amber-500 text-white rounded-lg hover:bg-amber-400 transition-all"
                @click="handleSaveGenerated(true)"
              >
                Overwrite
              </button>
              <button
                class="px-3 py-1 text-xs font-medium text-makoclaw-text-secondary hover:text-makoclaw-text transition-all"
                @click="showOverwriteConfirm = false"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Uninstall Confirmation -->
    <Transition name="modal">
      <div
        v-if="pendingUninstall"
        class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-4 px-5 py-3 bg-makoclaw-surface/95 backdrop-blur-xl border border-makoclaw-border/30 rounded-2xl shadow-2xl"
      >
        <span class="text-sm text-makoclaw-text">Uninstall <strong>{{ pendingUninstall }}</strong>?</span>
        <button
          class="px-3 py-1.5 text-xs font-semibold bg-makoclaw-error text-white rounded-lg hover:bg-red-500 transition-all"
          @click="uninstallSkill(pendingUninstall)"
        >
          Uninstall
        </button>
        <button
          class="px-3 py-1.5 text-xs font-medium text-makoclaw-text-secondary hover:text-makoclaw-text transition-all"
          @click="cancelUninstall"
        >
          Cancel
        </button>
      </div>
    </Transition>

    <!-- Editor Modal -->
    <Transition name="modal">
      <div
        v-if="editingSkill"
        class="modal-backdrop"
        @click.self="editingSkill = null"
      >
        <div class="modal-content max-w-4xl h-[85vh]">
          <div class="modal-header">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-makoclaw-accent/20 flex items-center justify-center">
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
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  />
                </svg>
              </div>
              <div>
                <h2 class="font-bold text-lg text-makoclaw-text">
                  Edit Skill
                </h2>
                <p class="text-xs text-makoclaw-text-secondary font-mono">
                  {{ editingSkill.name }}
                </p>
              </div>
            </div>
            <button
              class="modal-close-btn"
              @click="editingSkill = null"
            >
              <svg
                class="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
          <div class="flex-1 overflow-hidden p-4">
            <textarea
              v-model="editContent"
              class="w-full h-full input-field font-mono text-sm resize-none"
              spellcheck="false"
            />
          </div>
          <div class="modal-footer">
            <button
              class="btn-ghost"
              @click="editingSkill = null"
            >
              Cancel
            </button>
            <button
              :disabled="savingSkill"
              class="btn-primary flex items-center gap-2"
              @click="saveEditedSkill"
            >
              <svg
                v-if="savingSkill"
                class="w-4 h-4 animate-spin"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                />
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                />
              </svg>
              {{ savingSkill ? 'Saving...' : 'Save Changes' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Viewer Modal -->
    <Transition name="modal">
      <div
        v-if="viewingSkill"
        class="modal-backdrop"
        @click.self="viewingSkill = null"
      >
        <div class="modal-content max-w-2xl max-h-[80vh]">
          <div class="modal-header">
            <h2 class="font-bold text-lg text-makoclaw-text">
              {{ viewingSkill.name }}
            </h2>
            <button
              class="modal-close-btn"
              @click="viewingSkill = null"
            >
              <svg
                class="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
          <div class="modal-body custom-scrollbar">
            <pre class="whitespace-pre-wrap text-sm font-mono text-makoclaw-text-secondary bg-makoclaw-bg/50 p-4 rounded-lg">{{ viewingSkill.content }}</pre>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Submit Modal -->
    <Transition name="modal">
      <div
        v-if="showSubmitModal"
        class="modal-backdrop"
        @click.self="closeSubmitModal"
      >
        <div class="modal-content max-w-lg max-h-[80vh]">
          <div class="modal-header">
            <h2 class="font-bold text-lg text-makoclaw-text">
              Share to Marketplace
            </h2>
            <button
              class="modal-close-btn"
              @click="closeSubmitModal"
            >
              <svg
                class="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
          <div class="modal-body custom-scrollbar space-y-4">
            <div>
              <p class="font-medium text-makoclaw-text">
                {{ submittingSkill?.name }}
              </p>
              <p class="text-sm text-makoclaw-text-secondary">
                {{ submittingSkill?.description || 'No description' }}
              </p>
            </div>

            <!-- Security Scan Error -->
            <div
              v-if="scanResult?.error && submitForm.visibility !== 'private'"
              class="p-4 rounded-xl bg-amber-500/10 border border-amber-500/20"
            >
              <p class="text-sm text-amber-400">{{ scanResult.error }}</p>
            </div>

            <!-- Security Scan -->
            <div
              v-if="scanResult && !scanResult.error && submitForm.visibility !== 'private'"
              class="p-4 rounded-xl"
              :class="scanResult.passed ? 'bg-emerald-500/10 border border-emerald-500/20' : 'bg-makoclaw-error/10 border border-makoclaw-error/20'"
            >
              <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-medium text-makoclaw-text">Security Score</span>
                <span
                  class="text-xl font-bold"
                  :class="scanResult.score >= 80 ? 'text-emerald-400' : scanResult.score >= 60 ? 'text-amber-400' : 'text-makoclaw-error'"
                >
                  {{ scanResult.score }}/100
                </span>
              </div>
              <p
                class="text-xs"
                :class="scanResult.passed ? 'text-emerald-400' : 'text-makoclaw-error'"
              >
                {{ scanResult.passed ? 'Passed security checks' : 'Security issues detected' }}
              </p>
              <div
                v-if="scanResult.findings && scanResult.findings.length"
                class="mt-3 space-y-1"
              >
                <div
                  v-for="(finding, idx) in scanResult.findings.slice(0, 3)"
                  :key="idx"
                  class="text-xs p-2 bg-makoclaw-bg/50 rounded-lg"
                >
                  <span
                    class="font-medium"
                    :class="{
                      'text-makoclaw-error': finding.severity === 'critical',
                      'text-orange-400': finding.severity === 'high',
                      'text-amber-400': finding.severity === 'medium',
                      'text-makoclaw-text-secondary': finding.severity === 'low'
                    }"
                  >{{ finding.severity.toUpperCase() }}:</span>
                  {{ finding.title }}
                </div>
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Category</label>
              <select
                v-model="submitForm.category"
                class="input-field"
              >
                <option value="general">
                  General
                </option>
                <option value="development">
                  Development
                </option>
                <option value="devops">
                  DevOps
                </option>
                <option value="productivity">
                  Productivity
                </option>
                <option value="integrations">
                  Integrations
                </option>
                <option value="ai-agents">
                  AI Agents
                </option>
              </select>
            </div>

            <div>
              <label class="block text-sm font-medium text-makoclaw-text mb-2">Visibility</label>
              <div class="space-y-2">
                <label
                  class="visibility-option"
                  :class="submitForm.visibility === 'public' ? 'visibility-option-active' : ''"
                >
                  <input
                    v-model="submitForm.visibility"
                    type="radio"
                    value="public"
                    class="sr-only"
                  >
                  <div
                    class="visibility-radio"
                    :class="submitForm.visibility === 'public' ? 'visibility-radio-active' : ''"
                  />
                  <div>
                    <div class="text-sm font-medium text-makoclaw-text">Public</div>
                    <div class="text-xs text-makoclaw-text-secondary">Visible to everyone</div>
                  </div>
                </label>
                <label
                  class="visibility-option"
                  :class="submitForm.visibility === 'private' ? 'visibility-option-active' : ''"
                >
                  <input
                    v-model="submitForm.visibility"
                    type="radio"
                    value="private"
                    class="sr-only"
                  >
                  <div
                    class="visibility-radio"
                    :class="submitForm.visibility === 'private' ? 'visibility-radio-active' : ''"
                  />
                  <div>
                    <div class="text-sm font-medium text-makoclaw-text flex items-center gap-1.5">
                      <svg
                        class="w-3.5 h-3.5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                        />
                      </svg>
                      Private
                    </div>
                    <div class="text-xs text-makoclaw-text-secondary">Only visible to you</div>
                  </div>
                </label>
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Tags</label>
              <input
                v-model="submitForm.tags"
                type="text"
                placeholder="github, automation, code-review"
                class="input-field"
              >
              <p class="text-xs text-makoclaw-text-secondary mt-1">
                Comma-separated
              </p>
            </div>
          </div>
          <div class="modal-footer">
            <button
              class="btn-ghost"
              @click="closeSubmitModal"
            >
              Cancel
            </button>
            <button
              :disabled="submitting || (scanResult && !scanResult.passed && submitForm.visibility !== 'private')"
              class="px-5 py-2 text-sm font-semibold bg-emerald-500 text-white rounded-lg hover:bg-emerald-400 transition-all disabled:opacity-50"
              @click="handleSubmitToMarketplace"
            >
              {{ submitting ? 'Submitting...' : 'Submit' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'
import { useConfigStore } from '../stores/configStore'

const toast = useToast()
const configStore = useConfigStore()

// Tab icons as render functions
const InstalledIcon = () => h('svg', { class: 'w-4 h-4', fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
  h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z' })
])
const MarketIcon = () => h('svg', { class: 'w-4 h-4', fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
  h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z' })
])
const BundleIcon = () => h('svg', { class: 'w-4 h-4', fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
  h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10' })
])
const SubmitIcon = () => h('svg', { class: 'w-4 h-4', fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
  h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12' })
])

const activeTab = ref('installed')
const loading = ref(true)
const loadingMarketplace = ref(false)
const loadingSubmissions = ref(false)
const loadingBundles = ref(false)
const skills = ref([])
const marketplaceSkills = ref([])
const mySubmissions = ref([])
const bundles = ref([])
const trendingSkills = ref([])
const installing = ref(null)
const forking = ref(null)
const installingBundle = ref(null)
const usageStats = ref([])
const disabledSlugs = ref(new Set())

const tabs = computed(() => [
  { id: 'installed', label: 'Installed', icon: InstalledIcon, count: skills.value.length },
  { id: 'marketplace', label: 'Marketplace', icon: MarketIcon, load: loadMarketplace },
  { id: 'bundles', label: 'Bundles', icon: BundleIcon, load: loadBundles },
  { id: 'submissions', label: 'My Submissions', icon: SubmitIcon, load: loadMySubmissions }
])

// Modals
const viewingSkill = ref(null)
const editingSkill = ref(null)
const editContent = ref('')
const savingSkill = ref(false)
const showSubmitModal = ref(false)
const submittingSkill = ref(null)
const submitForm = ref({ category: 'general', tags: '', visibility: 'public' })
const submitting = ref(false)
const scanResult = ref(null)
const showGenerateModal = ref(false)
const generating = ref(false)
const savingGenerated = ref(false)
const generateError = ref('')
const generatedPreview = ref('')
const aiPrompt = ref('')
const aiGenerating = ref(false)
const generateForm = ref({ name: '', description: '', prompt: '' })
const generateFormErrors = ref({ name: '', description: '' })

// Rating
const ratingOpenSlug = ref(null)
const pendingRating = ref({})
const submittingRating = ref(false)

// Helpers
const getUsageCount = (skillName) => {
  const stat = usageStats.value.find(s => s.skill_name === skillName)
  return stat ? stat.load_count : 0
}

const isDisabled = (skill) => disabledSlugs.value.has(skill.slug || skill.name)

const sourceClasses = (source) => {
  const base = 'px-2 py-0.5 text-[10px] font-medium rounded-full'
  if (source === 'user') return `${base} bg-emerald-500/10 text-emerald-400`
  if (source === 'workspace') return `${base} bg-makoclaw-accent/10 text-makoclaw-accent`
  if (source === 'global') return `${base} bg-purple-500/10 text-purple-400`
  if (source === 'builtin') return `${base} bg-amber-500/10 text-amber-400`
  return `${base} bg-makoclaw-text-secondary/10 text-makoclaw-text-secondary`
}

const statusClasses = (status) => {
  const base = 'px-2 py-0.5 text-[10px] font-medium rounded-full'
  if (status === 'approved') return `${base} bg-emerald-500/10 text-emerald-400`
  if (status === 'pending' || status === 'needs_review') return `${base} bg-amber-500/10 text-amber-400`
  return `${base} bg-makoclaw-error/10 text-makoclaw-error`
}

const securityScoreClasses = (score) => {
  const base = 'px-2 py-1 text-xs font-bold rounded-lg'
  if (score >= 80) return `${base} bg-emerald-500/10 text-emerald-400`
  if (score >= 60) return `${base} bg-amber-500/10 text-amber-400`
  return `${base} bg-makoclaw-error/10 text-makoclaw-error`
}

// Rating
const toggleRatingWidget = (slug) => {
  if (ratingOpenSlug.value === slug) {
    ratingOpenSlug.value = null
  } else {
    ratingOpenSlug.value = slug
    if (!pendingRating.value[slug]) {
      pendingRating.value[slug] = { stars: 0, review: '' }
    }
  }
}

const setStars = (slug, stars) => {
  if (!pendingRating.value[slug]) pendingRating.value[slug] = { stars: 0, review: '' }
  pendingRating.value[slug].stars = stars
}

const submitRating = async (skill) => {
  const slug = skill.slug
  const pr = pendingRating.value[slug]
  if (!pr || pr.stars < 1 || pr.stars > 5) {
    toast.error('Please select a rating (1-5)')
    return
  }
  submittingRating.value = true
  try {
    const summary = await advancedService.rateSkill(slug, pr.stars, pr.review || '')
    const idx = marketplaceSkills.value.findIndex(s => s.slug === slug)
    if (idx !== -1) {
      marketplaceSkills.value[idx].average_rating = summary.average_rating
      marketplaceSkills.value[idx].rating_count = summary.rating_count
    }
    ratingOpenSlug.value = null
    toast.success('Rating submitted!')
  } catch {
    toast.error('Failed to submit rating')
  } finally {
    submittingRating.value = false
  }
}

// Helpers
const getErrorMessage = (err, fallback) => {
  const data = err?.response?.data
  if (typeof data === 'string' && data) return data
  if (data?.error) return data.error
  return fallback
}

// Load functions
const loadSkills = async () => {
  loading.value = true
  try {
    const data = await advancedService.fetchSkills()
    skills.value = data.skills || []
  } catch (err) {
    console.error('Failed to load skills:', err)
    toast.error(getErrorMessage(err, 'Failed to load skills'))
  } finally {
    loading.value = false
  }
  try {
    const alertData = await advancedService.fetchSecurityAlerts()
    disabledSlugs.value = new Set(alertData.disabled_slugs || [])
  } catch {
    // non-critical — silently ignore
  }
}

const loadMarketplace = async () => {
  loadingMarketplace.value = true
  try {
    const data = await advancedService.fetchMarketplaceSkills()
    marketplaceSkills.value = data.skills || []
  } catch (err) {
    marketplaceSkills.value = []
    toast.error(getErrorMessage(err, 'Failed to load marketplace'))
  } finally {
    loadingMarketplace.value = false
  }
  // Trending
  advancedService.fetchMarketplaceSkills({ sort: 'trending', limit: 3, page: 1 }).then(data => {
    trendingSkills.value = (data.skills || []).slice(0, 3)
  }).catch(() => {})
}

const loadMySubmissions = async () => {
  loadingSubmissions.value = true
  try {
    const data = await advancedService.fetchMySubmissions()
    mySubmissions.value = data.submissions || []
  } catch {
    toast.error('Failed to load submissions')
  } finally {
    loadingSubmissions.value = false
  }
}

const loadBundles = async () => {
  loadingBundles.value = true
  try {
    const data = await advancedService.fetchMarketplaceBundles()
    bundles.value = data.bundles || []
  } catch (err) {
    console.error('Failed to load bundles:', err)
    toast.error(getErrorMessage(err, 'Failed to load bundles'))
  } finally {
    loadingBundles.value = false
  }
}

// Actions
const viewSkill = async (name) => {
  try {
    const data = await advancedService.viewSkill(name)
    viewingSkill.value = data
  } catch {
    toast.error('Failed to load skill')
  }
}

const editSkill = async (name) => {
  try {
    const data = await advancedService.viewSkill(name)
    editingSkill.value = { name: data.name }
    editContent.value = data.content
  } catch {
    toast.error('Failed to load skill')
  }
}

const saveEditedSkill = async () => {
  if (!editContent.value.trim()) return
  savingSkill.value = true
  try {
    await advancedService.createSkill({
      name: editingSkill.value.name,
      content: editContent.value,
      overwrite: true
    })
    toast.success('Skill updated')
    editingSkill.value = null
    await loadSkills()
  } catch (err) {
    const msg = err?.response?.data?.error || err?.response?.data || 'Failed to update skill'
    toast.error(msg)
  } finally {
    savingSkill.value = false
  }
}

const pendingUninstall = ref(null)

const confirmUninstall = (name) => {
  pendingUninstall.value = name
}

const cancelUninstall = () => {
  pendingUninstall.value = null
}

const uninstallSkill = async (name) => {
  pendingUninstall.value = null
  try {
    await advancedService.uninstallSkill(name)
    toast.success('Skill uninstalled')
    await loadSkills()
    await configStore.checkStatus()
  } catch (err) {
    toast.error(getErrorMessage(err, 'Failed to uninstall skill'))
  }
}

const installMarketplaceSkill = async (slug) => {
  installing.value = slug
  try {
    const skill = marketplaceSkills.value.find(s => s.slug === slug)
    const result = await advancedService.installMarketplaceSkill(slug)
    const autoInstalled = result.auto_installed || []
    if (autoInstalled.length > 0) {
      toast.success(`Installed ${skill ? skill.name : slug}. Also: ${autoInstalled.join(', ')}`)
    } else {
      toast.success('Skill installed')
    }
    // Mark as installed in marketplace view
    if (skill) skill.installed = true
    await loadSkills()
  } catch {
    toast.error('Failed to install skill')
  } finally {
    installing.value = null
  }
}

const forkSkill = async (skill) => {
  forking.value = skill.slug
  try {
    const result = await advancedService.forkSkill(skill.slug)
    toast.success(result.message || `Forked as '${result.skill_name}'`)
    skill.fork_count = (skill.fork_count || 0) + 1
    await loadSkills()
  } catch (err) {
    toast.error(getErrorMessage(err, 'Failed to fork skill'))
  } finally {
    forking.value = null
  }
}

const installBundle = async (bundle) => {
  installingBundle.value = bundle.slug
  try {
    const result = await advancedService.installBundle(bundle.slug)
    const installed = result.installed || []
    if (installed.length > 0) {
      toast.success(`Installed ${installed.length} skill(s) from "${bundle.name}"`)
    } else {
      toast.info('All skills already installed')
    }
    bundle.install_count = (bundle.install_count || 0) + 1
    await loadSkills()
  } catch (err) {
    toast.error(getErrorMessage(err, 'Failed to install bundle'))
  } finally {
    installingBundle.value = null
  }
}

// Submit modal
const openSubmitModal = async (skill) => {
  try {
    const data = await advancedService.viewSkill(skill.name)
    submittingSkill.value = { ...skill, content: data.content }
    submitForm.value = { category: 'general', tags: '', visibility: 'public' }
    scanResult.value = null
    showSubmitModal.value = true
    try {
      scanResult.value = await advancedService.scanSkill(data.content)
    } catch {
      scanResult.value = { error: 'Security scan unavailable. You can still submit.' }
    }
  } catch {
    toast.error('Failed to load skill')
  }
}

const closeSubmitModal = () => {
  showSubmitModal.value = false
  submittingSkill.value = null
  scanResult.value = null
}

const handleSubmitToMarketplace = async () => {
  if (!submittingSkill.value) return
  submitting.value = true
  try {
    const tags = submitForm.value.tags.split(',').map(t => t.trim()).filter(Boolean)
    const result = await advancedService.submitToMarketplace({
      name: submittingSkill.value.name,
      description: submittingSkill.value.description || '',
      content: submittingSkill.value.content,
      category: submitForm.value.category,
      tags,
      visibility: submitForm.value.visibility
    })
    toast.success(result.message || 'Skill submitted')
    closeSubmitModal()
    await loadMySubmissions()
    activeTab.value = 'submissions'
  } catch (err) {
    toast.error(getErrorMessage(err, 'Failed to submit skill'))
  } finally {
    submitting.value = false
  }
}

// Generate modal
const openGenerateModal = () => {
  showGenerateModal.value = true
  generateError.value = ''
  generatedPreview.value = ''
  generateForm.value = { name: '', description: '', prompt: '' }
  generateFormErrors.value = { name: '', description: '' }
  aiPrompt.value = ''
}

const closeGenerateModal = () => {
  showGenerateModal.value = false
  generating.value = false
  savingGenerated.value = false
}

const validateSkillName = () => {
  const name = generateForm.value.name.trim()
  if (!name) {
    generateFormErrors.value.name = ''
    return
  }
  const validPattern = /^[a-z0-9]+(-[a-z0-9]+)*$/
  if (!validPattern.test(name)) {
    generateFormErrors.value.name = 'Use lowercase, numbers, and hyphens only'
  } else if (name.length > 64) {
    generateFormErrors.value.name = 'Max 64 characters'
  } else {
    generateFormErrors.value.name = ''
  }
}

const generateSkillConfigWithAI = async () => {
  if (!aiPrompt.value.trim()) return
  aiGenerating.value = true
  generateError.value = ''
  try {
    const data = await advancedService.generateSkillConfig(aiPrompt.value)
    if (data) {
      if (data.name) generateForm.value.name = data.name
      if (data.description) generateForm.value.description = data.description
      if (data.prompt) generateForm.value.prompt = data.prompt
      validateSkillName()
      toast.success('Configuration generated!')
      aiPrompt.value = ''
    }
  } catch (err) {
    generateError.value = err?.response?.data || 'Failed to generate configuration'
  } finally {
    aiGenerating.value = false
  }
}

const handleGenerate = async () => {
  generateError.value = ''
  generateFormErrors.value = { name: '', description: '' }
  const name = generateForm.value.name.trim()
  const description = generateForm.value.description.trim()
  if (!name) {
    generateFormErrors.value.name = 'Name is required'
    return
  }
  const validPattern = /^[a-z0-9]+(-[a-z0-9]+)*$/
  if (!validPattern.test(name)) {
    generateFormErrors.value.name = 'Use lowercase, numbers, and hyphens only'
    return
  }
  if (!description) {
    generateFormErrors.value.description = 'Description is required'
    return
  }
  generating.value = true
  try {
    let goal = description
    if (generateForm.value.prompt.trim()) {
      goal += '\n\nAdditional context:\n' + generateForm.value.prompt.trim()
    }
    const data = await advancedService.generateSkillDraft({
      name: name,
      goal: goal,
      capabilities: '',
      constraints: '',
      tools: '',
      examples: ''
    })
    if (data.name) generateForm.value.name = data.name
    generatedPreview.value = data.draft || ''
    toast.success('Skill draft generated')
  } catch (err) {
    generateError.value = err?.response?.data || 'Failed to generate skill'
  } finally {
    generating.value = false
  }
}

const showOverwriteConfirm = ref(false)

const handleSaveGenerated = async (overwrite = false) => {
  if (!generatedPreview.value.trim()) {
    generateError.value = 'No content to save'
    return
  }
  savingGenerated.value = true
  generateError.value = ''
  showOverwriteConfirm.value = false
  try {
    await advancedService.createSkill({
      name: generateForm.value.name,
      content: generatedPreview.value,
      overwrite
    })
    toast.success('Skill created!')
    closeGenerateModal()
    await loadSkills()
  } catch (err) {
    if (err?.response?.status === 409 && !overwrite) {
      savingGenerated.value = false
      showOverwriteConfirm.value = true
      return
    }
    const msg = getErrorMessage(err, 'Failed to save skill')
    generateError.value = msg
    toast.error(msg)
  } finally {
    savingGenerated.value = false
  }
}

onMounted(() => {
  loadSkills()
  advancedService.fetchSkillAnalytics().then(data => {
    usageStats.value = data.stats || []
  }).catch(() => {})
})
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* Skill Cards */
.skill-card {
  @apply bg-makoclaw-surface/80 backdrop-blur-sm border border-makoclaw-border/50 rounded-xl p-4
         transition-all duration-200 hover:border-makoclaw-accent/20 hover:shadow-lg hover:shadow-makoclaw-accent/5
         flex flex-col;
}

.skill-card-skeleton {
  @apply bg-makoclaw-surface/50 border border-makoclaw-border/30 rounded-xl p-4;
}

.skill-icon {
  @apply w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500/20 to-makoclaw-accent/20 flex items-center justify-center flex-shrink-0 text-purple-400;
}

.skill-action-btn {
  @apply px-3 py-1.5 text-xs font-medium text-makoclaw-text-secondary bg-makoclaw-bg/50 rounded-lg
         hover:bg-makoclaw-bg transition-all flex items-center gap-1.5;
}

.skill-tag {
  @apply px-2 py-0.5 text-[10px] font-medium rounded-full bg-makoclaw-bg/50 text-makoclaw-text-secondary border border-makoclaw-border/30;
}

/* Marketplace Cards */
.marketplace-card {
  @apply bg-makoclaw-surface/80 backdrop-blur-sm border border-makoclaw-border/50 rounded-xl p-4
         transition-all duration-200 hover:border-makoclaw-accent/20 hover:shadow-lg hover:shadow-makoclaw-accent/5
         flex flex-col;
}

.trending-card {
  @apply bg-gradient-to-br from-amber-500/5 to-orange-500/5 border border-amber-500/20 rounded-xl p-3
         hover:border-amber-500/30 transition-all;
}

.bundle-card {
  @apply bg-makoclaw-surface/80 backdrop-blur-sm border border-makoclaw-border/50 rounded-xl p-4
         transition-all duration-200 hover:border-makoclaw-accent/20 hover:shadow-lg;
}

.submission-card {
  @apply bg-makoclaw-surface/80 backdrop-blur-sm border border-makoclaw-border/50 rounded-xl p-4;
}

.install-btn {
  @apply px-4 py-1.5 text-xs font-semibold bg-makoclaw-accent text-white rounded-lg
         hover:bg-makoclaw-accent-hover transition-all disabled:opacity-50;
}

/* Rating Widget */
.rating-widget {
  @apply mt-3 p-3 bg-makoclaw-bg/50 rounded-xl border border-makoclaw-border/30;
}

/* AI Generate Section */
.ai-generate-section {
  @apply p-5 bg-gradient-to-br from-purple-500/5 to-makoclaw-accent/5 border border-purple-500/20 rounded-xl;
}

.ai-icon {
  @apply w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500/20 to-makoclaw-accent/20 flex items-center justify-center flex-shrink-0;
}

/* Modal Styles */
.modal-backdrop {
  @apply fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-modal p-4;
}

.modal-content {
  @apply bg-makoclaw-surface border border-makoclaw-border/50 rounded-2xl w-full shadow-2xl flex flex-col overflow-hidden;
}

.modal-header {
  @apply flex items-center justify-between p-4 border-b border-makoclaw-border/30;
}

.modal-body {
  @apply flex-1 overflow-auto p-4;
}

.modal-footer {
  @apply flex justify-end gap-3 p-4 border-t border-makoclaw-border/30;
}

.modal-close-btn {
  @apply p-2 hover:bg-makoclaw-bg rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors;
}

/* Visibility Options */
.visibility-option {
  @apply flex items-center gap-3 p-3 rounded-xl border border-makoclaw-border/50 cursor-pointer transition-all;
}

.visibility-option-active {
  @apply border-makoclaw-accent/50 bg-makoclaw-accent/5;
}

.visibility-radio {
  @apply w-4 h-4 rounded-full border-2 border-makoclaw-border transition-all flex-shrink-0;
}

.visibility-radio-active {
  @apply border-makoclaw-accent bg-makoclaw-accent;
}
</style>
