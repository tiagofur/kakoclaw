<template>
  <div class="flex flex-col h-full bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-teal-500/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-emerald-500/20 via-transparent to-transparent" />
    </div>

    <!-- Page Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
        <div class="flex items-center gap-3">
          <!-- Icon Container -->
          <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-teal-500/20 to-emerald-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-teal-500/10">
            <svg
              class="w-5 h-5 sm:w-6 sm:h-6 text-teal-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 6.042A8.967 8.967 0 006 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 016 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 016-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0018 18a8.967 8.967 0 00-6 2.292m0-14.25v14.25"
              />
            </svg>
          </div>

          <!-- Title -->
          <div class="flex-1 min-w-0">
            <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-teal-400 bg-clip-text text-transparent">
              Knowledge Base
            </h1>
            <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5">
              Upload documents for AI-powered search
            </p>
          </div>

          <!-- Document Count Badge -->
          <div class="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 rounded-xl">
            <svg
              class="w-4 h-4 text-teal-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
            <span class="text-sm font-medium text-makoclaw-text">{{ documents.length }}</span>
            <span class="text-sm text-makoclaw-text-secondary">document{{ documents.length !== 1 ? 's' : '' }}</span>
          </div>

          <!-- Upload Button -->
          <button
            class="px-4 sm:px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-teal-500 to-teal-600 hover:from-teal-600 hover:to-teal-700 text-white rounded-xl transition-all shadow-lg shadow-teal-500/25 hover:shadow-teal-500/40 text-sm font-bold flex items-center gap-2 active:scale-95 flex-shrink-0"
            @click="$refs.fileInput.click()"
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
                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
              />
            </svg>
            <span class="hidden sm:inline">Upload</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-4 sm:p-6 custom-scrollbar">
      <!-- Loading Skeleton -->
      <div
        v-if="loading"
        class="space-y-4"
      >
        <div
          v-for="i in 3"
          :key="i"
          class="glass-panel p-5 rounded-xl"
        >
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="skeleton h-4 w-48 mb-3 rounded" />
              <div class="flex gap-2">
                <div class="skeleton h-5 w-16 rounded-full" />
                <div class="skeleton h-5 w-12 rounded-full" />
                <div class="skeleton h-5 w-20 rounded-full" />
              </div>
            </div>
          </div>
          <div class="flex items-center justify-between mt-4">
            <div class="skeleton h-3 w-24 rounded" />
            <div class="flex gap-2">
              <div class="skeleton h-7 w-14 rounded-lg" />
              <div class="skeleton h-7 w-16 rounded-lg" />
            </div>
          </div>
        </div>
      </div>

      <template v-else>
        <!-- Upload Area -->
        <div
          class="relative group mb-6 cursor-pointer"
          @dragover.prevent="dragOver = true"
          @dragleave.prevent="dragOver = false"
          @drop.prevent="handleDrop"
          @click="$refs.fileInput.click()"
        >
          <!-- Glow effect on hover/drag -->
          <div
            class="absolute -inset-1 bg-gradient-to-r from-teal-500/20 to-emerald-500/20 rounded-2xl blur-xl transition-opacity duration-300"
            :class="dragOver || uploading ? 'opacity-100' : 'opacity-0 group-hover:opacity-50'"
          />

          <div
            class="relative glass-panel p-8 rounded-xl text-center transition-all duration-200"
            :class="[
              uploading ? 'ring-2 ring-teal-500 animate-subtlePulse' : '',
              dragOver ? 'ring-2 ring-teal-500 bg-teal-500/5' : 'hover:border-teal-500/30'
            ]"
          >
            <input
              ref="fileInput"
              type="file"
              class="hidden"
              accept=".txt,.md,.pdf,.json,.csv,.html,.xml,.yaml,.yml,.log"
              multiple
              @change="handleFileSelect"
            >

            <div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-gradient-to-br from-teal-500/20 to-emerald-500/20 flex items-center justify-center ring-1 ring-white/10">
              <svg
                class="w-7 h-7 text-teal-400"
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
            </div>

            <p class="text-makoclaw-text font-medium">
              <span
                v-if="uploading"
                class="text-teal-400"
              >Uploading...</span>
              <span v-else>Drop files here or <span class="text-teal-400">click to browse</span></span>
            </p>
            <p class="text-xs text-makoclaw-text-secondary mt-2">
              Supports TXT, MD, PDF, JSON, CSV, HTML, XML, YAML, LOG
            </p>
          </div>
        </div>

        <!-- Search Section -->
        <div class="glass-panel p-4 rounded-xl mb-6">
          <div class="flex gap-2 sm:gap-3">
            <div class="flex-1 relative group">
              <svg
                class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-makoclaw-text-secondary group-focus-within:text-teal-400 transition-colors"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                />
              </svg>
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Search knowledge base..."
                class="w-full pl-10 pr-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-teal-500/30 focus:border-teal-500/50 transition-all text-sm backdrop-blur-sm min-h-[40px]"
                @keyup.enter="runSearch"
              >
            </div>
            <button
              :disabled="!searchQuery.trim()"
              class="px-4 sm:px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-teal-500 to-teal-600 hover:from-teal-600 hover:to-teal-700 text-white rounded-xl transition-all shadow-lg shadow-teal-500/25 text-sm font-bold flex items-center gap-2 active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed disabled:shadow-none"
              @click="runSearch"
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
                  d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                />
              </svg>
              <span class="hidden sm:inline">Search</span>
            </button>
          </div>

          <!-- Search Results -->
          <div
            v-if="searchResults.length > 0"
            class="mt-4 pt-4 border-t border-makoclaw-border/30"
          >
            <h3 class="text-sm font-semibold text-makoclaw-text mb-3 flex items-center gap-2">
              <svg
                class="w-4 h-4 text-teal-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              {{ searchResults.length }} result{{ searchResults.length !== 1 ? 's' : '' }} found
            </h3>
            <div class="space-y-3 max-h-80 overflow-y-auto custom-scrollbar">
              <div
                v-for="(result, idx) in searchResults"
                :key="idx"
                class="search-result-card p-4 animate-fadeSlide"
                :style="{ animationDelay: idx * 50 + 'ms' }"
              >
                <div class="flex items-center gap-2 mb-2">
                  <span class="px-2 py-0.5 text-xs font-medium text-teal-400 bg-teal-500/10 rounded-full">{{ result.document_name }}</span>
                  <span class="text-xs text-makoclaw-text-secondary">chunk #{{ result.position }}</span>
                  <span class="text-xs text-makoclaw-text-secondary ml-auto px-2 py-0.5 bg-makoclaw-bg/50 rounded-full">
                    score: {{ result.rank?.toFixed(2) }}
                  </span>
                </div>
                <p class="text-sm text-makoclaw-text whitespace-pre-wrap line-clamp-4">
                  {{ result.content }}
                </p>
              </div>
            </div>
          </div>

          <div
            v-else-if="searchPerformed && searchResults.length === 0"
            class="mt-4 pt-4 border-t border-makoclaw-border/30"
          >
            <div class="flex items-center justify-center gap-2 py-4 text-makoclaw-text-secondary text-sm">
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
                  d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              No results found for "{{ lastSearchQuery }}"
            </div>
          </div>
        </div>

        <!-- Documents Section -->
        <div>
          <h3 class="text-sm font-semibold text-makoclaw-text mb-4 flex items-center gap-2">
            <svg
              class="w-4 h-4 text-teal-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
              />
            </svg>
            Documents
          </h3>

          <!-- Empty State -->
          <div
            v-if="documents.length === 0"
            class="flex flex-col items-center justify-center py-16 text-center"
          >
            <div class="relative">
              <!-- Glow effect -->
              <div class="absolute inset-0 bg-gradient-to-br from-teal-500/30 to-emerald-500/30 rounded-3xl blur-2xl opacity-50" />
              <div class="relative glass-panel p-8 rounded-2xl shadow-2xl ring-1 ring-white/10">
                <div class="w-16 h-16 mx-auto rounded-2xl bg-gradient-to-br from-teal-500/20 to-emerald-500/20 flex items-center justify-center ring-1 ring-white/20">
                  <svg
                    class="w-8 h-8 text-teal-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="1.5"
                      d="M12 6.042A8.967 8.967 0 006 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 016 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 016-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0018 18a8.967 8.967 0 00-6 2.292m0-14.25v14.25"
                    />
                  </svg>
                </div>
              </div>
            </div>
            <h3 class="text-lg font-bold text-makoclaw-text mt-6">
              No documents yet
            </h3>
            <p class="text-sm text-makoclaw-text-secondary/70 mt-2 max-w-xs">
              Add documents to build your knowledge base for AI-powered search.
            </p>
            <button
              class="mt-6 px-5 py-2.5 bg-gradient-to-r from-teal-500 to-teal-600 hover:from-teal-600 hover:to-teal-700 text-white rounded-xl transition-all shadow-lg shadow-teal-500/25 hover:shadow-teal-500/40 text-sm font-bold flex items-center gap-2 active:scale-95"
              @click="$refs.fileInput.click()"
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
                  d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
                />
              </svg>
              Upload Documents
            </button>
          </div>

          <!-- Documents Grid -->
          <div
            v-else
            class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
          >
            <div
              v-for="doc in documents"
              :key="doc.id"
              class="document-card group"
            >
              <div class="flex items-start gap-3">
                <!-- Document Icon -->
                <div class="w-10 h-10 flex-shrink-0 rounded-lg bg-gradient-to-br from-teal-500/20 to-emerald-500/20 flex items-center justify-center ring-1 ring-white/10 group-hover:ring-teal-500/30 transition-all">
                  <svg
                    class="w-5 h-5 text-teal-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                </div>

                <div class="flex-1 min-w-0">
                  <h4
                    class="font-semibold text-makoclaw-text truncate group-hover:text-teal-400 transition-colors"
                    :title="doc.name"
                  >
                    {{ doc.name }}
                  </h4>
                  <div class="flex flex-wrap gap-1.5 mt-2">
                    <span class="px-2 py-0.5 text-xs bg-makoclaw-bg/50 text-makoclaw-text-secondary rounded-full">{{ doc.mime_type || 'text/plain' }}</span>
                    <span class="px-2 py-0.5 text-xs bg-makoclaw-bg/50 text-makoclaw-text-secondary rounded-full">{{ formatSize(doc.size) }}</span>
                    <span class="px-2 py-0.5 text-xs bg-teal-500/10 text-teal-400 rounded-full">{{ doc.chunk_count }} chunk{{ doc.chunk_count !== 1 ? 's' : '' }}</span>
                  </div>
                </div>
              </div>

              <div class="flex items-center justify-between mt-4 pt-3 border-t border-makoclaw-border/20">
                <span class="text-xs text-makoclaw-text-secondary">{{ formatDate(doc.created_at) }}</span>
                <div class="flex gap-2">
                  <button
                    class="px-3 py-1.5 text-xs font-medium text-makoclaw-text bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-lg hover:bg-teal-500/10 hover:text-teal-400 hover:border-teal-500/30 transition-all"
                    @click="openDocViewer(doc)"
                  >
                    View
                  </button>
                  <button
                    :disabled="deleting === doc.id"
                    class="px-3 py-1.5 text-xs font-medium text-red-400 bg-red-500/10 border border-red-500/20 rounded-lg hover:bg-red-500/20 transition-all disabled:opacity-50"
                    @click="deleteDoc(doc.id, doc.name)"
                  >
                    <span v-if="deleting === doc.id">...</span>
                    <span v-else>Delete</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Document Viewer Modal -->
    <Transition name="modal">
      <div
        v-if="selectedDoc"
        class="fixed inset-0 bg-black/60 backdrop-blur-sm z-modal flex items-center justify-center p-4"
        @click.self="closeDocViewer"
      >
        <div class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl w-full max-w-4xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden ring-1 ring-white/10 animate-scaleIn">
          <!-- Modal Header -->
          <div class="p-4 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent flex-none">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-9 h-9 rounded-lg bg-gradient-to-br from-teal-500/20 to-emerald-500/20 flex items-center justify-center ring-1 ring-white/10">
                  <svg
                    class="w-4.5 h-4.5 text-teal-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                </div>
                <div>
                  <h3 class="font-bold text-makoclaw-text">
                    {{ selectedDoc.name }}
                  </h3>
                  <p class="text-xs text-makoclaw-text-secondary mt-0.5">
                    {{ selectedDoc.chunk_count }} chunks • {{ formatSize(selectedDoc.size) }}
                  </p>
                </div>
              </div>
              <button
                class="p-2 hover:bg-makoclaw-border/30 rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
                @click="closeDocViewer"
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
          </div>

          <!-- Modal Content -->
          <div class="flex-1 overflow-auto p-4 sm:p-6 custom-scrollbar">
            <div
              v-if="loadingChunks"
              class="flex justify-center items-center h-40"
            >
              <div class="animate-spin rounded-full h-8 w-8 border-2 border-makoclaw-border border-t-teal-500" />
            </div>
            <div
              v-else
              class="space-y-4"
            >
              <div
                v-for="chunk in docChunks"
                :key="chunk.id"
                class="chunk-card"
              >
                <div class="flex items-center justify-between mb-3 pb-2 border-b border-makoclaw-border/30">
                  <div class="flex items-center gap-2">
                    <span class="w-6 h-6 rounded-md bg-teal-500/10 flex items-center justify-center text-xs font-bold text-teal-400">
                      {{ chunk.position + 1 }}
                    </span>
                    <h5 class="text-xs font-semibold text-makoclaw-text-secondary uppercase tracking-wider">
                      Chunk
                    </h5>
                  </div>
                  <div v-if="editingChunkId !== chunk.id">
                    <button
                      class="text-xs text-teal-400 hover:text-teal-300 flex items-center gap-1 transition-colors"
                      @click="startEditingChunk(chunk)"
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
                          d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
                        />
                      </svg>
                      Edit
                    </button>
                  </div>
                  <div
                    v-else
                    class="flex gap-2"
                  >
                    <button
                      class="text-xs text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
                      @click="cancelEditingChunk"
                    >
                      Cancel
                    </button>
                    <button
                      :disabled="savingChunk"
                      class="text-xs bg-teal-500 text-white px-3 py-1 rounded-lg hover:bg-teal-600 disabled:opacity-50 transition-colors"
                      @click="saveChunk(chunk.id)"
                    >
                      {{ savingChunk ? 'Saving...' : 'Save' }}
                    </button>
                  </div>
                </div>

                <textarea
                  v-if="editingChunkId === chunk.id"
                  v-model="editChunkContent"
                  class="w-full h-40 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-lg p-3 text-sm font-mono focus:ring-2 focus:ring-teal-500/30 focus:border-teal-500/50 outline-none text-makoclaw-text resize-y transition-all"
                />
                <div
                  v-else
                  class="text-sm text-makoclaw-text whitespace-pre-wrap font-mono leading-relaxed bg-makoclaw-bg/30 p-3 rounded-lg border border-makoclaw-border/20"
                >
                  {{ chunk.content }}
                </div>
              </div>
            </div>
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
const loading = ref(true)
const uploading = ref(false)
const dragOver = ref(false)
const deleting = ref(null)
const documents = ref([])
const searchQuery = ref('')
const lastSearchQuery = ref('')
const searchResults = ref([])
const searchPerformed = ref(false)

// Chunk Viewer/Editor state
const selectedDoc = ref(null)
const docChunks = ref([])
const loadingChunks = ref(false)
const editingChunkId = ref(null)
const editChunkContent = ref('')
const savingChunk = ref(false)

const loadDocuments = async () => {
  loading.value = true
  try {
    const data = await advancedService.fetchKnowledgeDocs()
    documents.value = data.documents || []
  } catch (err) {
    console.error('Failed to load knowledge documents:', err)
    toast.error('Failed to load documents')
  } finally {
    loading.value = false
  }
}

const uploadFiles = async (files) => {
  if (!files || files.length === 0) return
  uploading.value = true
  let successCount = 0
  let failCount = 0
  for (const file of files) {
    try {
      await advancedService.uploadKnowledgeDoc(file)
      successCount++
    } catch (err) {
      console.error(`Failed to upload ${file.name}:`, err)
      failCount++
    }
  }
  uploading.value = false
  if (successCount > 0) {
    toast.success(`Uploaded ${successCount} document${successCount !== 1 ? 's' : ''}`)
    await loadDocuments()
  }
  if (failCount > 0) {
    toast.error(`Failed to upload ${failCount} file${failCount !== 1 ? 's' : ''}`)
  }
}

const handleDrop = (e) => {
  dragOver.value = false
  const files = e.dataTransfer?.files
  if (files) uploadFiles(Array.from(files))
}

const handleFileSelect = (e) => {
  const files = e.target?.files
  if (files) uploadFiles(Array.from(files))
  e.target.value = '' // reset so same file can be re-uploaded
}

const deleteDoc = async (id, name) => {
  if (!confirm(`Delete "${name}"? This cannot be undone.`)) return
  deleting.value = id
  try {
    await advancedService.deleteKnowledgeDoc(id)
    toast.success('Document deleted')
    await loadDocuments()
  } catch (err) {
    console.error('Failed to delete document:', err)
    toast.error('Failed to delete document')
  } finally {
    deleting.value = null
  }
}

// Viewer and Chunk Edit Logic
const openDocViewer = async (doc) => {
  selectedDoc.value = doc
  loadingChunks.value = true
  editingChunkId.value = null
  try {
    const data = await advancedService.fetchKnowledgeChunks(doc.id)
    docChunks.value = data.chunks || []
  } catch (err) {
    console.error('Failed to load doc chunks:', err)
    toast.error('Failed to load document chunks')
    selectedDoc.value = null
  } finally {
    loadingChunks.value = false
  }
}

const closeDocViewer = () => {
  selectedDoc.value = null
  docChunks.value = []
  editingChunkId.value = null
}

const startEditingChunk = (chunk) => {
  editingChunkId.value = chunk.id
  editChunkContent.value = chunk.content
}

const cancelEditingChunk = () => {
  editingChunkId.value = null
  editChunkContent.value = ''
}

const saveChunk = async (chunkId) => {
  const newContent = editChunkContent.value.trim()
  if (!newContent) {
    toast.error('Chunk content cannot be empty')
    return
  }

  savingChunk.value = true
  try {
    await advancedService.updateKnowledgeChunk(chunkId, newContent)

    // update local state
    const chunkIndex = docChunks.value.findIndex(c => c.id === chunkId)
    if (chunkIndex !== -1) {
      docChunks.value[chunkIndex].content = newContent
    }

    toast.success('Chunk updated')
    editingChunkId.value = null
  } catch (err) {
    console.error('Failed to update chunk', err)
    toast.error('Failed to update chunk')
  } finally {
    savingChunk.value = false
  }
}

const runSearch = async () => {
  const q = searchQuery.value.trim()
  if (!q) return
  lastSearchQuery.value = q
  searchPerformed.value = true
  try {
    const data = await advancedService.searchKnowledge(q)
    searchResults.value = data.results || []
  } catch (err) {
    console.error('Search failed:', err)
    toast.error('Search failed')
    searchResults.value = []
  }
}

const formatSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let size = bytes
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024
    i++
  }
  return `${size.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

onMounted(() => loadDocuments())
</script>

<style scoped>
.line-clamp-4 {
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.document-card {
  @apply p-4 rounded-xl bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 hover:bg-makoclaw-surface/50 hover:border-teal-500/20 transition-all duration-200;
}

.search-result-card {
  @apply rounded-xl bg-makoclaw-bg/40 border border-makoclaw-border/30 hover:border-teal-500/20 transition-all;
}

.chunk-card {
  @apply p-4 rounded-xl bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 hover:border-teal-500/20 transition-all;
}
</style>
