<template>
  <div class="h-full flex flex-col bg-makoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-makoclaw-border bg-makoclaw-surface flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      <div>
        <h2 class="text-xl font-bold bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent">Skills</h2>
        <p class="text-sm text-makoclaw-text-secondary mt-1">Manage installed skills and browse the marketplace</p>
      </div>
      <div class="flex flex-col sm:flex-row items-start sm:items-center gap-3">
        <button
          @click="openGenerateModal"
          class="px-4 py-2 text-sm font-medium bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent-hover transition-colors flex items-center gap-2 shadow-sm"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          Create Skill
        </button>
        <div class="flex bg-makoclaw-bg rounded-lg p-1 border border-makoclaw-border">
          <button
            @click="activeTab = 'installed'"
            class="tab-button"
            :class="[activeTab === 'installed' ? 'tab-button-active' : 'tab-button-inactive']"
          >Installed</button>
          <button
            @click="activeTab = 'marketplace'; loadMarketplace()"
            class="tab-button"
            :class="[activeTab === 'marketplace' ? 'tab-button-active' : 'tab-button-inactive']"
          >Marketplace</button>
          <button
            @click="activeTab = 'submissions'; loadMySubmissions()"
            class="tab-button"
            :class="[activeTab === 'submissions' ? 'tab-button-active' : 'tab-button-inactive']"
          >My Submissions</button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-4 md:p-6 custom-scrollbar">
      <!-- Loading Skeleton -->
      <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-for="i in 6" :key="i" class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="skeleton h-4 w-32 mb-2 rounded"></div>
              <div class="skeleton h-3 w-full mb-1 rounded"></div>
              <div class="skeleton h-3 w-2/3 rounded"></div>
            </div>
            <div class="skeleton h-5 w-16 rounded-full ml-2"></div>
          </div>
          <div class="flex items-center gap-2 mt-4">
            <div class="skeleton h-7 w-12 rounded-lg"></div>
            <div class="skeleton h-7 w-14 rounded-lg"></div>
          </div>
        </div>
      </div>

      <template v-else>
        <Transition name="fade" mode="out-in">
        <!-- Installed Skills -->
        <div v-if="activeTab === 'installed'" key="installed">
          <div v-if="skills.length === 0" class="text-center py-12 text-makoclaw-text-secondary">
            <p class="text-lg">No skills installed</p>
            <p class="text-sm mt-2">Browse the marketplace to install skills</p>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div
              v-for="skill in skills"
              :key="skill.name"
              class="card-interactive p-5"
            >
              <div class="flex items-start justify-between">
                <div class="flex-1 min-w-0">
                  <h3 class="font-semibold truncate">{{ skill.name }}</h3>
                  <p class="text-sm text-makoclaw-text-secondary mt-1 line-clamp-2">{{ skill.description || 'No description' }}</p>
                </div>
                <span class="ml-2 px-2 py-0.5 text-xs rounded-full flex-shrink-0"
                  :class="{
                    'bg-blue-500/10 text-blue-400': skill.source === 'workspace',
                    'bg-blue-500/10 text-blue-400': skill.source === 'global',
                    'bg-gray-500/10 text-gray-400': skill.source === 'builtin'
                  }"
                >{{ skill.source }}</span>
              </div>
              <div class="flex items-center gap-2 mt-4 flex-wrap">
                <button
                  @click="viewSkill(skill.name)"
                  class="px-3 py-1.5 text-xs bg-makoclaw-bg rounded-lg hover:bg-makoclaw-border/50 transition-colors"
                >Ver</button>
                <button
                  v-if="skill.source === 'workspace'"
                  @click="editSkill(skill.name)"
                  class="px-3 py-1.5 text-xs bg-makoclaw-accent/10 text-makoclaw-accent rounded-lg hover:bg-makoclaw-accent/20 transition-colors"
                >Editar</button>
                <button
                  v-if="skill.source === 'workspace'"
                  @click="openSubmitModal(skill)"
                  class="px-3 py-1.5 text-xs bg-green-500/10 text-green-400 rounded-lg hover:bg-green-500/20 transition-colors"
                >Submit to Marketplace</button>
                <button
                  v-if="skill.source === 'workspace'"
                  @click="uninstallSkill(skill.name)"
                  class="px-3 py-1.5 text-xs text-red-400 bg-red-500/10 rounded-lg hover:bg-red-500/20 transition-colors"
                >Desinstalar</button>
              </div>
            </div>
          </div>
        </div>

        <!-- Marketplace -->
        <div v-else-if="activeTab === 'marketplace'" key="marketplace">
          <div v-if="loadingMarketplace" class="flex items-center justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-makoclaw-accent"></div>
          </div>
          <div v-else-if="marketplaceSkills.length === 0" class="text-center py-12 text-makoclaw-text-secondary">
            <p class="text-lg">No skills available in marketplace</p>
            <p class="text-sm mt-2">Be the first to submit a skill!</p>
          </div>
          <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div
              v-for="skill in marketplaceSkills"
              :key="skill.slug || skill.name"
              class="card-interactive p-5"
            >
              <div class="flex items-start justify-between">
                <h3 class="font-semibold flex-1">{{ skill.name }}</h3>
                <span
                  v-if="skill.security_score !== undefined"
                  class="px-2 py-0.5 text-xs rounded-full flex-shrink-0"
                  :class="{
                    'bg-green-500/10 text-green-400': skill.security_score >= 80,
                    'bg-yellow-500/10 text-yellow-400': skill.security_score >= 60 && skill.security_score < 80,
                    'bg-red-500/10 text-red-400': skill.security_score < 60
                  }"
                >{{ skill.security_score }}/100</span>
              </div>
              <p class="text-sm text-makoclaw-text-secondary mt-1 line-clamp-2">{{ skill.description }}</p>
              <p class="text-xs text-makoclaw-text-secondary mt-2">by {{ skill.author || 'unknown' }}</p>
              <div v-if="skill.tags && skill.tags.length" class="flex flex-wrap gap-1 mt-2">
                <span v-for="tag in skill.tags" :key="tag" class="px-2 py-0.5 text-xs bg-makoclaw-bg rounded-full text-makoclaw-text-secondary">{{ tag }}</span>
              </div>
              <div class="flex items-center gap-2 mt-4">
                <button
                  @click="installMarketplaceSkill(skill.slug)"
                  :disabled="installing === skill.slug"
                  class="px-4 py-1.5 text-sm bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors disabled:opacity-50"
                >
                  <span v-if="installing === skill.slug">Installing...</span>
                  <span v-else>Install</span>
                </button>
                <span v-if="skill.install_count" class="text-xs text-makoclaw-text-secondary">{{ skill.install_count }} installs</span>
                <div class="flex items-center gap-2 ml-auto">
                  <span v-if="skill.rating_count > 0" class="text-xs text-yellow-400">
                    &#9733; {{ skill.average_rating?.toFixed(1) ?? '—' }} ({{ skill.rating_count }})
                  </span>
                  <button
                    @click="toggleRatingWidget(skill.slug)"
                    class="text-xs px-2 py-1 bg-makoclaw-bg rounded-lg hover:bg-makoclaw-border/50 transition-colors"
                    :class="{ 'text-yellow-400': ratingOpenSlug === skill.slug }"
                  >Rate</button>
                </div>
              </div>

              <!-- Inline Rating Widget -->
              <div v-if="ratingOpenSlug === skill.slug" class="mt-3 p-3 bg-makoclaw-bg rounded-lg border border-makoclaw-border">
                <p class="text-xs font-medium mb-2">Your rating</p>
                <div class="flex gap-1 mb-2">
                  <button
                    v-for="star in 5"
                    :key="star"
                    @click="setStars(skill.slug, star)"
                    class="text-xl leading-none transition-colors"
                    :class="(pendingRating[skill.slug]?.stars ?? 0) >= star ? 'text-yellow-400' : 'text-makoclaw-text-secondary'"
                  >&#9733;</button>
                </div>
                <textarea
                  v-model="pendingRating[skill.slug].review"
                  placeholder="Optional review (max 500 chars)"
                  maxlength="500"
                  rows="2"
                  class="w-full px-2 py-1.5 text-xs bg-makoclaw-surface border border-makoclaw-border rounded-lg resize-none focus:ring-1 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent transition-all"
                ></textarea>
                <div class="flex justify-end gap-2 mt-2">
                  <button @click="ratingOpenSlug = null" class="text-xs px-3 py-1 text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">Cancel</button>
                  <button
                    @click="submitRating(skill)"
                    :disabled="submittingRating || !(pendingRating[skill.slug]?.stars > 0)"
                    class="text-xs px-3 py-1 bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors disabled:opacity-50"
                  >{{ submittingRating ? 'Submitting...' : 'Submit Rating' }}</button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- My Submissions -->
        <div v-else-if="activeTab === 'submissions'" key="submissions">
          <div v-if="loadingSubmissions" class="flex items-center justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-makoclaw-accent"></div>
          </div>
          <div v-else-if="mySubmissions.length === 0" class="text-center py-12 text-makoclaw-text-secondary">
            <p class="text-lg">No submissions yet</p>
            <p class="text-sm mt-2">Submit a skill from your installed workspace skills to share with others</p>
          </div>
          <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div
              v-for="sub in mySubmissions"
              :key="sub.id"
              class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5"
            >
              <div class="flex items-start justify-between">
                <h3 class="font-semibold flex-1">{{ sub.skill_name }}</h3>
                <div class="flex items-center gap-1.5 flex-shrink-0">
                  <span
                    v-if="sub.visibility === 'private'"
                    class="px-2 py-0.5 text-xs rounded-full bg-makoclaw-bg text-makoclaw-text-secondary border border-makoclaw-border"
                    title="Private skill"
                  >&#128274;</span>
                  <span
                    class="px-2 py-0.5 text-xs rounded-full"
                    :class="{
                      'bg-green-500/10 text-green-400': sub.status === 'approved',
                      'bg-yellow-500/10 text-yellow-400': sub.status === 'pending' || sub.status === 'needs_review',
                      'bg-red-500/10 text-red-400': sub.status === 'rejected'
                    }"
                  >{{ sub.status }}</span>
                </div>
              </div>
              <p class="text-sm text-makoclaw-text-secondary mt-1 line-clamp-2">{{ sub.description }}</p>
              <p class="text-xs text-makoclaw-text-secondary mt-2">Security Score: {{ sub.security_score }}/100</p>
              <p v-if="sub.reviewer_notes" class="text-xs text-yellow-400 mt-2 bg-yellow-500/10 p-2 rounded">{{ sub.reviewer_notes }}</p>
            </div>
          </div>
        </div>
        </Transition>

      </template>
    </div>

    <!-- Skill Generation Modal -->
    <Transition name="modal">
    <div v-if="showGenerateModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-modal p-4" @click.self="closeGenerateModal">
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-3xl w-full max-h-[90vh] flex flex-col shadow-2xl">
        <!-- Modal Header -->
        <div class="flex items-center justify-between p-4 border-b border-makoclaw-border">
          <div class="flex items-center gap-2">
            <svg class="w-5 h-5 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
            </svg>
            <h3 class="font-bold text-lg">Create New Skill</h3>
          </div>
          <button @click="closeGenerateModal" class="p-1.5 hover:bg-makoclaw-bg rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>

        <!-- Modal Content -->
        <div class="flex-1 overflow-auto p-5 custom-scrollbar">
          <!-- Step 1: Input Form (before generation) -->
          <div v-if="!generatedPreview" class="space-y-4">
            <!-- AI Assistant Section -->
            <div class="bg-gradient-to-r from-makoclaw-accent/10 to-blue-500/10 border border-makoclaw-accent/30 rounded-xl p-5 mb-6">
              <div class="flex items-start gap-3">
                <div class="w-10 h-10 rounded-lg bg-makoclaw-accent/20 flex items-center justify-center flex-shrink-0">
                  <svg class="w-5 h-5 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
                <div class="flex-1">
                  <h4 class="font-bold text-makoclaw-text mb-1 flex items-center gap-2">
                    Create Skill with AI
                    <span class="px-2 py-0.5 text-[10px] font-bold uppercase bg-makoclaw-accent text-white rounded-full">New</span>
                  </h4>
                  <p class="text-sm text-makoclaw-text-secondary mb-3">Describe what you want the skill to do and AI will formulate the form for you.</p>
                  <div class="space-y-3">
                    <textarea
                      v-model="aiPrompt"
                      rows="3"
                      placeholder="e.g., 'Create a skill that helps me review pull requests on GitHub quickly.'"
                      class="w-full px-4 py-3 bg-makoclaw-bg/60 border border-makoclaw-border rounded-xl text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent outline-none text-makoclaw-text resize-none backdrop-blur-sm"
                    ></textarea>
                    <button
                      @click="generateSkillConfigWithAI"
                      :disabled="!aiPrompt.trim() || aiGenerating"
                      class="w-full px-4 py-2.5 bg-gradient-to-r from-makoclaw-accent to-blue-600 hover:from-makoclaw-accent-hover hover:to-blue-700 text-white rounded-xl font-bold shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98]"
                    >
                      <svg v-if="aiGenerating" class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                      </svg>
                      {{ aiGenerating ? 'Generating Configuration...' : 'Generate with AI' }}
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <div class="flex items-center gap-4 mb-4">
              <div class="flex-1 h-px bg-makoclaw-border"></div>
              <span class="text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary">OR CONFIGURE MANUALLY</span>
              <div class="flex-1 h-px bg-makoclaw-border"></div>
            </div>

            <p class="text-sm text-makoclaw-text-secondary">Provide a name, description and context. The AI will generate a SKILL.md template that you can customize.</p>

            <!-- Skill Name -->
            <div>
              <label class="block text-sm font-medium mb-1.5">Skill Name <span class="text-red-400">*</span></label>
              <input
                v-model="generateForm.name"
                type="text"
                placeholder="e.g., jira-assistant, code-reviewer, deploy-helper"
                class="w-full px-3 py-2.5 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent transition-all"
                :class="{ 'border-red-400': generateFormErrors.name }"
                @input="validateSkillName"
              >
              <p v-if="generateFormErrors.name" class="text-xs text-red-400 mt-1">{{ generateFormErrors.name }}</p>
              <p v-else class="text-xs text-makoclaw-text-secondary mt-1">Use lowercase letters, numbers, and hyphens only</p>
            </div>

            <!-- Description -->
            <div>
              <label class="block text-sm font-medium mb-1.5">Description <span class="text-red-400">*</span></label>
              <textarea
                v-model="generateForm.description"
                rows="3"
                placeholder="Describe what this skill does and when the agent should use it..."
                class="w-full px-3 py-2.5 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent transition-all resize-none"
                :class="{ 'border-red-400': generateFormErrors.description }"
              ></textarea>
              <p v-if="generateFormErrors.description" class="text-xs text-red-400 mt-1">{{ generateFormErrors.description }}</p>
            </div>

            <!-- Additional Context (Optional) -->
            <div>
              <label class="block text-sm font-medium mb-1.5">Additional Context <span class="text-makoclaw-text-secondary font-normal">(optional)</span></label>
              <textarea
                v-model="generateForm.prompt"
                rows="4"
                placeholder="Provide additional details: specific commands, example use cases, safety constraints, required tools..."
                class="w-full px-3 py-2.5 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent transition-all resize-none"
              ></textarea>
            </div>

            <!-- Error Message -->
            <p v-if="generateError" class="text-sm text-red-400 bg-red-500/10 px-3 py-2 rounded-lg">{{ generateError }}</p>
          </div>

          <!-- Step 2: Preview (after generation) -->
          <div v-else class="space-y-4">
            <div class="flex items-center justify-between">
              <p class="text-sm text-makoclaw-text-secondary">Review and edit the generated SKILL.md before saving:</p>
              <button
                @click="generatedPreview = ''"
                class="text-xs text-makoclaw-accent hover:text-makoclaw-accent-hover transition-colors"
              >Back to form</button>
            </div>
            <textarea
              v-model="generatedPreview"
              rows="20"
              class="w-full px-4 py-3 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm font-mono focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent transition-all resize-none custom-scrollbar"
              spellcheck="false"
            ></textarea>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 border-t border-makoclaw-border flex justify-end gap-3">
          <button
            @click="closeGenerateModal"
            class="px-4 py-2 text-sm font-medium text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
          >Cancel</button>

          <!-- Generate Button (Step 1) -->
          <button
            v-if="!generatedPreview"
            @click="handleGenerate"
            :disabled="generating || !generateForm.name || !generateForm.description"
            class="px-5 py-2 text-sm font-semibold bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent-hover shadow-lg shadow-makoclaw-accent/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            <span v-if="generating" class="animate-spin h-4 w-4 border-2 border-white/30 border-t-white rounded-full"></span>
            <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            {{ generating ? 'Generating...' : 'Generate Skill' }}
          </button>

          <!-- Save Button (Step 2) -->
          <button
            v-else
            @click="handleSaveGenerated"
            :disabled="savingGenerated || !generatedPreview.trim()"
            class="px-5 py-2 text-sm font-semibold bg-green-600 text-white rounded-lg hover:bg-green-500 shadow-lg shadow-green-600/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            <span v-if="savingGenerated" class="animate-spin h-4 w-4 border-2 border-white/30 border-t-white rounded-full"></span>
            <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
            {{ savingGenerated ? 'Saving...' : 'Save Skill' }}
          </button>
        </div>
      </div>
    </div>
    </Transition>

    <!-- Skill Editor Modal -->
    <Transition name="modal">
    <div v-if="editingSkill" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-modal p-4" @click.self="editingSkill = null">
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-4xl w-full h-[85vh] flex flex-col shadow-2xl">
        <div class="flex items-center justify-between p-4 border-b border-makoclaw-border">
          <div class="flex items-center gap-2">
            <h3 class="font-bold text-lg text-makoclaw-accent">Editar Skill:</h3>
            <span class="font-mono text-sm bg-makoclaw-bg px-2 py-0.5 rounded border border-makoclaw-border">{{ editingSkill.name }}</span>
          </div>
          <button @click="editingSkill = null" class="p-1 hover:bg-makoclaw-bg rounded text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>
        <div class="flex-1 overflow-hidden p-4">
          <div class="h-full flex flex-col gap-3">
             <p class="text-xs text-makoclaw-text-secondary">Modifica el contenido markdown de la skill. Asegúrate de mantener el bloque YAML (frontmatter) al principio.</p>
             <textarea 
               v-model="editContent" 
               class="flex-1 w-full p-4 bg-makoclaw-bg border border-makoclaw-border rounded-xl text-sm font-mono focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent transition-all resize-none shadow-inner custom-scrollbar"
               spellcheck="false"
             ></textarea>
          </div>
        </div>
        <div class="p-4 border-t border-makoclaw-border flex justify-end gap-3">
          <button @click="editingSkill = null" class="px-5 py-2 text-sm font-medium text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">Cancelar</button>
          <button 
            @click="saveEditedSkill" 
            :disabled="savingSkill" 
            class="px-6 py-2 text-sm font-semibold bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent-hover shadow-lg shadow-makoclaw-accent/20 disabled:opacity-50 transition-all flex items-center gap-2"
          >
            <span v-if="savingSkill" class="animate-spin h-3.5 w-3.5 border-2 border-white/30 border-t-white rounded-full"></span>
            {{ savingSkill ? 'Guardando...' : 'Guardar Cambios' }}
          </button>
        </div>
      </div>
    </div>
    </Transition>

    <!-- Skill Viewer Modal -->
    <Transition name="modal">
    <div v-if="viewingSkill" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-modal p-4" @click.self="viewingSkill = null">
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-2xl w-full max-h-[80vh] flex flex-col shadow-xl">
        <div class="flex items-center justify-between p-4 border-b border-makoclaw-border">
          <h3 class="font-semibold">{{ viewingSkill.name }}</h3>
          <button @click="viewingSkill = null" class="p-1 hover:bg-makoclaw-bg rounded text-makoclaw-text-secondary hover:text-makoclaw-text">&times;</button>
        </div>
        <div class="flex-1 overflow-auto p-4 custom-scrollbar">
          <pre class="whitespace-pre-wrap text-sm font-mono">{{ viewingSkill.content }}</pre>
        </div>
      </div>
    </div>
    </Transition>

    <!-- Submit to Marketplace Modal -->
    <Transition name="modal">
    <div v-if="showSubmitModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-modal p-4" @click.self="closeSubmitModal">
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-lg w-full max-h-[80vh] flex flex-col shadow-xl">
        <div class="flex items-center justify-between p-4 border-b border-makoclaw-border">
          <h3 class="font-semibold">Submit to Marketplace</h3>
          <button @click="closeSubmitModal" class="p-1 hover:bg-makoclaw-bg rounded text-makoclaw-text-secondary hover:text-makoclaw-text">&times;</button>
        </div>
        <div class="flex-1 overflow-auto p-4 custom-scrollbar space-y-4">
          <div>
            <p class="font-medium">{{ submittingSkill?.name }}</p>
            <p class="text-sm text-makoclaw-text-secondary">{{ submittingSkill?.description || 'No description' }}</p>
          </div>

          <!-- Security Scan Results -->
          <div v-if="scanResult && submitForm.visibility !== 'private'" class="p-3 rounded-lg" :class="scanResult.passed ? 'bg-green-500/10' : 'bg-red-500/10'">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium">Security Score</span>
              <span class="text-lg font-bold" :class="scanResult.score >= 80 ? 'text-green-400' : scanResult.score >= 60 ? 'text-yellow-400' : 'text-red-400'">
                {{ scanResult.score }}/100
              </span>
            </div>
            <p class="text-xs mt-1" :class="scanResult.passed ? 'text-green-400' : 'text-red-400'">
              {{ scanResult.passed ? 'Passed security checks' : 'Security issues detected' }}
            </p>
            <div v-if="scanResult.findings && scanResult.findings.length" class="mt-2 space-y-1">
              <div v-for="(finding, idx) in scanResult.findings.slice(0, 3)" :key="idx" class="text-xs p-2 bg-makoclaw-bg rounded">
                <span class="font-medium" :class="{
                  'text-red-400': finding.severity === 'critical',
                  'text-orange-400': finding.severity === 'high',
                  'text-yellow-400': finding.severity === 'medium',
                  'text-gray-400': finding.severity === 'low'
                }">{{ finding.severity.toUpperCase() }}:</span>
                {{ finding.title }}
              </div>
              <p v-if="scanResult.findings.length > 3" class="text-xs text-makoclaw-text-secondary">
                And {{ scanResult.findings.length - 3 }} more issues...
              </p>
            </div>
          </div>

          <!-- Category -->
          <div>
            <label class="block text-sm font-medium mb-1">Category</label>
            <select v-model="submitForm.category" class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm">
              <option value="general">General</option>
              <option value="development">Development</option>
              <option value="devops">DevOps</option>
              <option value="productivity">Productivity</option>
              <option value="integrations">Integrations</option>
              <option value="ai-agents">AI Agents</option>
            </select>
          </div>

          <!-- Visibility -->
          <div>
            <label class="block text-sm font-medium mb-2">Visibility</label>
            <div class="space-y-2">
              <label class="flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-colors"
                :class="submitForm.visibility === 'public' ? 'border-makoclaw-accent bg-makoclaw-accent/5' : 'border-makoclaw-border hover:border-makoclaw-accent/50'">
                <input type="radio" v-model="submitForm.visibility" value="public" class="mt-0.5 accent-makoclaw-accent" />
                <div>
                  <div class="text-sm font-medium">Public</div>
                  <div class="text-xs text-makoclaw-text-secondary">Visible to everyone in the marketplace</div>
                </div>
              </label>
              <label class="flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-colors"
                :class="submitForm.visibility === 'private' ? 'border-makoclaw-accent bg-makoclaw-accent/5' : 'border-makoclaw-border hover:border-makoclaw-accent/50'">
                <input type="radio" v-model="submitForm.visibility" value="private" class="mt-0.5 accent-makoclaw-accent" />
                <div>
                  <div class="text-sm font-medium flex items-center gap-1.5">
                    <span>&#128274;</span> Private
                  </div>
                  <div class="text-xs text-makoclaw-text-secondary">Only visible to you — auto-approved instantly</div>
                </div>
              </label>
            </div>
          </div>

          <!-- Tags -->
          <div>
            <label class="block text-sm font-medium mb-1">Tags (comma-separated)</label>
            <input
              v-model="submitForm.tags"
              type="text"
              placeholder="e.g., github, automation, code-review"
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm"
            >
          </div>
        </div>
        <div class="p-4 border-t border-makoclaw-border flex justify-end gap-3">
          <button @click="closeSubmitModal" class="px-4 py-2 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text">Cancel</button>
          <button
            @click="handleSubmitToMarketplace"
            :disabled="submitting || (scanResult && !scanResult.passed && submitForm.visibility !== 'private')"
            class="px-4 py-2 text-sm font-medium bg-green-600 text-white rounded-lg hover:bg-green-500 disabled:opacity-50 disabled:cursor-not-allowed"
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
import { ref, onMounted } from 'vue'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'

const toast = useToast()
const activeTab = ref('installed')
const loading = ref(true)
const loadingAvailable = ref(false)
const loadingMarketplace = ref(false)
const loadingSubmissions = ref(false)
const skills = ref([])
const available = ref([])
const marketplaceSkills = ref([])
const mySubmissions = ref([])
const installing = ref(null)
const viewingSkill = ref(null)
const editingSkill = ref(null)
const editContent = ref('')
const savingSkill = ref(false)

// Rating Widget State
const ratingOpenSlug = ref(null)
const pendingRating = ref({})
const submittingRating = ref(false)

// Submit Modal State
const showSubmitModal = ref(false)
const submittingSkill = ref(null)
const submitForm = ref({ category: 'general', tags: '', visibility: 'public' })
const submitting = ref(false)
const scanResult = ref(null)

// Generation Modal State
const showGenerateModal = ref(false)
const generating = ref(false)
const savingGenerated = ref(false)
const generateError = ref('')
const generatedPreview = ref('')
const aiPrompt = ref('')
const aiGenerating = ref(false)
const generateForm = ref({
  name: '',
  description: '',
  prompt: ''
})
const generateFormErrors = ref({
  name: '',
  description: ''
})

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
  if (!pendingRating.value[slug]) {
    pendingRating.value[slug] = { stars: 0, review: '' }
  }
  pendingRating.value[slug].stars = stars
}

const submitRating = async (skill) => {
  const slug = skill.slug
  const pr = pendingRating.value[slug]
  if (!pr || pr.stars < 1 || pr.stars > 5) {
    toast.error('Please select a star rating (1–5)')
    return
  }
  submittingRating.value = true
  try {
    const summary = await advancedService.rateSkill(slug, pr.stars, pr.review || '')
    // Update the card in-place
    const idx = marketplaceSkills.value.findIndex(s => s.slug === slug)
    if (idx !== -1) {
      marketplaceSkills.value[idx].average_rating = summary.average_rating
      marketplaceSkills.value[idx].rating_count = summary.rating_count
    }
    ratingOpenSlug.value = null
    toast.success('Rating submitted!')
  } catch (err) {
    toast.error('Failed to submit rating')
  } finally {
    submittingRating.value = false
  }
}

const loadSkills = async () => {
  loading.value = true
  try {
    const data = await advancedService.fetchSkills()
    skills.value = data.skills || []
  } catch (err) {
    console.error('Failed to load skills:', err)
  } finally {
    loading.value = false
  }
}

const loadAvailable = async () => {
  if (available.value.length > 0) return
  loadingAvailable.value = true
  try {
    const data = await advancedService.fetchAvailableSkills()
    available.value = data.skills || []
    if (data.warning) {
      toast.info(data.warning)
    }
  } catch (err) {
    console.error('Failed to load marketplace:', err)
    toast.error('Failed to load marketplace')
  } finally {
    loadingAvailable.value = false
  }
}

// New Marketplace functions
const loadMarketplace = async () => {
  loadingMarketplace.value = true
  try {
    const data = await advancedService.fetchMarketplaceSkills()
    marketplaceSkills.value = data.skills || []
  } catch (err) {
    console.error('Failed to load marketplace:', err)
    // Fallback to old endpoint
    await loadAvailable()
    marketplaceSkills.value = available.value
  } finally {
    loadingMarketplace.value = false
  }
}

const loadMySubmissions = async () => {
  loadingSubmissions.value = true
  try {
    const data = await advancedService.fetchMySubmissions()
    mySubmissions.value = data.submissions || []
  } catch (err) {
    console.error('Failed to load submissions:', err)
    toast.error('Failed to load submissions')
  } finally {
    loadingSubmissions.value = false
  }
}

const installMarketplaceSkill = async (slug) => {
  installing.value = slug
  try {
    await advancedService.installMarketplaceSkill(slug)
    toast.success('Skill installed successfully')
    await loadSkills()
  } catch (err) {
    toast.error('Failed to install skill')
  } finally {
    installing.value = null
  }
}

const openSubmitModal = async (skill) => {
  try {
    const data = await advancedService.viewSkill(skill.name)
    submittingSkill.value = { ...skill, content: data.content }
    submitForm.value = { category: 'general', tags: '', visibility: 'public' }
    scanResult.value = null
    showSubmitModal.value = true

    // Run security scan
    try {
      scanResult.value = await advancedService.scanSkill(data.content)
    } catch (err) {
      console.error('Scan failed:', err)
    }
  } catch (err) {
    toast.error('Failed to load skill content')
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

    toast.success(result.message || 'Skill submitted successfully')
    closeSubmitModal()

    // Refresh submissions tab
    await loadMySubmissions()
    activeTab.value = 'submissions'
  } catch (err) {
    toast.error(err?.response?.data || 'Failed to submit skill')
  } finally {
    submitting.value = false
  }
}

const viewSkill = async (name) => {
  try {
    const data = await advancedService.viewSkill(name)
    viewingSkill.value = data
  } catch (err) {
    toast.error('Failed to load skill content')
  }
}

const editSkill = async (name) => {
  try {
    const data = await advancedService.viewSkill(name)
    editingSkill.value = { name: data.name }
    editContent.value = data.content
  } catch (err) {
    toast.error('Failed to load skill for editing')
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
    toast.success('Skill actualizada correctamente')
    editingSkill.value = null
    await loadSkills()
  } catch (err) {
    toast.error(err?.response?.data || 'Error al actualizar la skill')
  } finally {
    savingSkill.value = false
  }
}

const installSkill = async (repo) => {
  installing.value = repo
  try {
    await advancedService.installSkill(repo)
    toast.success('Skill installed successfully')
    await loadSkills()
  } catch (err) {
    toast.error('Failed to install skill')
  } finally {
    installing.value = null
  }
}

const uninstallSkill = async (name) => {
  try {
    await advancedService.uninstallSkill(name)
    toast.success('Skill uninstalled')
    await loadSkills()
  } catch (err) {
    toast.error('Failed to uninstall skill')
  }
}

// Generation Modal Functions
const openGenerateModal = () => {
  showGenerateModal.value = true
  generateError.value = ''
  generatedPreview.value = ''
  generateForm.value = { name: '', description: '', prompt: '' }
  generateFormErrors.value = { name: '', description: '' }
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
  // Check for valid characters (lowercase, numbers, hyphens)
  const validPattern = /^[a-z0-9]+(-[a-z0-9]+)*$/
  if (!validPattern.test(name)) {
    generateFormErrors.value.name = 'Use lowercase letters, numbers, and hyphens only (e.g., my-skill-name)'
  } else if (name.length > 64) {
    generateFormErrors.value.name = 'Name must be 64 characters or less'
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
      toast.success('Skill configuration generated!')
      aiPrompt.value = '' // Clear input after successful generation
    }
  } catch (err) {
    generateError.value = err?.response?.data || 'Failed to generate skill configuration. Please try again.'
  } finally {
    aiGenerating.value = false
  }
}

const handleGenerate = async () => {
  // Validate form
  generateError.value = ''
  generateFormErrors.value = { name: '', description: '' }

  const name = generateForm.value.name.trim()
  const description = generateForm.value.description.trim()

  if (!name) {
    generateFormErrors.value.name = 'Skill name is required'
    return
  }

  // Validate name format
  const validPattern = /^[a-z0-9]+(-[a-z0-9]+)*$/
  if (!validPattern.test(name)) {
    generateFormErrors.value.name = 'Use lowercase letters, numbers, and hyphens only'
    return
  }

  if (!description) {
    generateFormErrors.value.description = 'Description is required'
    return
  }

  generating.value = true
  try {
    // Build goal from description + optional prompt
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

    // Update the form name if the backend normalized it
    if (data.name) {
      generateForm.value.name = data.name
    }

    generatedPreview.value = data.draft || ''
    toast.success('Skill draft generated successfully')
  } catch (err) {
    generateError.value = err?.response?.data || 'Failed to generate skill draft. Please try again.'
  } finally {
    generating.value = false
  }
}

const handleSaveGenerated = async (overwrite = false) => {
  if (!generatedPreview.value.trim()) {
    generateError.value = 'No content to save'
    return
  }

  savingGenerated.value = true
  generateError.value = ''

  try {
    await advancedService.createSkill({
      name: generateForm.value.name,
      content: generatedPreview.value,
      overwrite
    })
    toast.success('Skill created successfully!')
    closeGenerateModal()
    await loadSkills()
  } catch (err) {
    if (err?.response?.status === 409 && !overwrite) {
      const confirmed = confirm('A skill with this name already exists. Do you want to overwrite it?')
      if (confirmed) {
        await handleSaveGenerated(true)
      }
      return
    }
    generateError.value = err?.response?.data || 'Failed to save skill'
  } finally {
    savingGenerated.value = false
  }
}

onMounted(() => loadSkills())
</script>

<style scoped>
.line-clamp-2 { display: -webkit-box; -webkit-line-clamp: 2; line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
</style>


