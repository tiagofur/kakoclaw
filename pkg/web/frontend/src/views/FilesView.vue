<template>
  <div class="flex flex-col h-full bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-indigo-500/30 via-transparent to-transparent"></div>
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-violet-500/20 via-transparent to-transparent"></div>
    </div>

    <!-- Page Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <!-- Title Section -->
      <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-indigo-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-indigo-500/10">
            <svg class="w-5 h-5 sm:w-6 sm:h-6 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-indigo-400 bg-clip-text text-transparent">Files</h1>
            <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5">Browse and manage workspace files</p>
          </div>

          <!-- Action Buttons -->
          <div class="flex items-center gap-2">
            <input type="file" ref="fileInput" class="hidden" @change="handleFileUpload" multiple />
            <button
              @click="showNewFolderModal = true"
              class="px-3 sm:px-4 py-2 min-h-[40px] bg-makoclaw-surface/50 border border-makoclaw-border/50 text-makoclaw-text rounded-xl hover:bg-makoclaw-surface-hover hover:border-indigo-500/30 transition-all text-sm font-medium flex items-center gap-2 backdrop-blur-sm"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
              </svg>
              <span class="hidden sm:inline">New Folder</span>
            </button>
            <button
              @click="$refs.fileInput.click()"
              :disabled="uploading"
              class="px-4 sm:px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-indigo-500 to-violet-500 hover:from-indigo-600 hover:to-violet-600 text-white rounded-xl transition-all shadow-lg shadow-indigo-500/25 hover:shadow-indigo-500/40 text-sm font-bold flex items-center gap-2 active:scale-95 disabled:opacity-50"
            >
              <svg v-if="uploading" class="animate-spin h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0l-4 4m4-4v12" />
              </svg>
              <span class="hidden sm:inline">{{ uploading ? 'Uploading...' : 'Upload' }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Search & Path Bar -->
      <div class="px-4 sm:px-6 pb-3 sm:pb-4">
        <div class="flex flex-col sm:flex-row gap-3">
          <!-- Search -->
          <div class="flex-1 relative group">
            <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-makoclaw-text-secondary group-focus-within:text-indigo-400 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search files..."
              class="w-full pl-10 pr-10 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-indigo-500/30 focus:border-indigo-500/50 transition-all text-sm backdrop-blur-sm min-h-[40px]"
              @input="debouncedSearch"
            />
            <button
              v-if="searchQuery"
              @click="clearSearch"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Current Path Badge -->
          <div class="flex items-center gap-2 px-3 py-2 bg-makoclaw-surface/40 border border-makoclaw-border/30 rounded-xl backdrop-blur-sm min-h-[40px]">
            <svg class="w-4 h-4 text-indigo-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
            </svg>
            <span class="text-sm font-mono text-makoclaw-text-secondary truncate max-w-[200px]">/{{ currentPath || 'workspace' }}</span>
          </div>
        </div>

        <!-- Search Status -->
        <p v-if="isSearching" class="text-xs text-indigo-400 mt-2 flex items-center gap-1.5">
          <svg class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
          </svg>
          Searching...
        </p>
        <p v-else-if="searchQuery && entries !== null" class="text-xs text-makoclaw-text-secondary mt-2">
          {{ entries.length }} result{{ entries.length !== 1 ? 's' : '' }} for "{{ searchQuery }}"
        </p>
      </div>

      <!-- Breadcrumb -->
      <div class="px-4 sm:px-6 pb-3 flex items-center gap-1.5 text-sm overflow-x-auto scrollbar-hide">
        <button @click="navigateTo('')" class="px-2 py-1 rounded-lg text-indigo-400 hover:bg-indigo-500/10 transition-all flex-shrink-0 font-medium">
          workspace
        </button>
        <template v-for="(part, i) in breadcrumbs" :key="i">
          <svg class="w-4 h-4 text-makoclaw-text-secondary/50 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
          <button @click="navigateTo(breadcrumbs.slice(0, i + 1).join('/'))" class="px-2 py-1 rounded-lg text-indigo-400 hover:bg-indigo-500/10 transition-all flex-shrink-0 font-medium">
            {{ part }}
          </button>
        </template>
      </div>
    </div>

    <!-- Content Area -->
    <div
      class="flex-1 overflow-auto custom-scrollbar relative"
      :class="{ 'bg-indigo-500/5': isDragging }"
      @dragover.prevent="isDragging = true"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="onDrop"
    >
      <!-- Drag overlay -->
      <div v-if="isDragging" class="absolute inset-0 z-50 flex items-center justify-center pointer-events-none">
        <div class="bg-gradient-to-r from-indigo-500 to-violet-500 text-white px-6 py-4 rounded-2xl shadow-2xl animate-subtlePulse flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-white/20 flex items-center justify-center">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
            </svg>
          </div>
          <div>
            <p class="font-bold">Drop files here</p>
            <p class="text-sm text-white/70">Upload to /{{ currentPath || 'workspace' }}</p>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-16">
        <div class="flex flex-col items-center gap-4">
          <div class="w-10 h-10 border-2 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
          <p class="text-sm text-makoclaw-text-secondary">Loading files...</p>
        </div>
      </div>

      <template v-else>
        <!-- Directory Listing -->
        <div v-if="entries !== null" class="p-4 sm:p-6">
          <!-- Parent Directory -->
          <button
            v-if="currentPath && !searchQuery"
            @click="navigateUp()"
            class="w-full flex items-center gap-3 px-4 py-3 mb-2 rounded-xl hover:bg-makoclaw-surface/50 transition-all text-left group"
          >
            <div class="w-10 h-10 rounded-xl bg-makoclaw-surface/50 flex items-center justify-center group-hover:bg-indigo-500/10 transition-colors">
              <svg class="w-5 h-5 text-makoclaw-text-secondary group-hover:text-indigo-400 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 17l-5-5m0 0l5-5m-5 5h12" />
              </svg>
            </div>
            <span class="text-makoclaw-text-secondary group-hover:text-makoclaw-text transition-colors font-medium">..</span>
          </button>

          <!-- Files & Folders Grid -->
          <div class="space-y-2">
            <div
              v-for="entry in entries"
              :key="entry.path"
              class="file-card group"
            >
              <!-- Entry Content -->
              <div
                class="flex-1 flex items-center gap-3 min-w-0 cursor-pointer"
                @click="entry.is_dir ? navigateTo(entry.path) : viewFile(entry.path)"
              >
                <!-- Icon -->
                <div :class="[
                  'w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0 transition-colors',
                  entry.is_dir ? 'bg-amber-500/10 group-hover:bg-amber-500/20' : 'bg-makoclaw-surface/50 group-hover:bg-indigo-500/10'
                ]">
                  <svg v-if="entry.is_dir" class="w-5 h-5 text-amber-400" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
                  </svg>
                  <svg v-else class="w-5 h-5 text-makoclaw-text-secondary group-hover:text-indigo-400 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>

                <!-- Name & Info -->
                <div class="flex-1 min-w-0">
                  <!-- Inline Rename -->
                  <input
                    v-if="renamingEntry === entry.path"
                    v-model="renameValue"
                    type="text"
                    class="w-full text-sm font-medium bg-makoclaw-bg border border-indigo-500/50 rounded-lg px-2.5 py-1.5 focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
                    @keyup.enter="confirmRename(entry)"
                    @keyup.esc="cancelRename"
                    @blur="cancelRename"
                    @click.stop
                    ref="renameInput"
                  />
                  <span
                    v-else
                    class="text-sm font-medium truncate block group-hover:text-indigo-400 transition-colors"
                    @dblclick.stop="startRename(entry)"
                    title="Double-click to rename"
                  >{{ entry.name }}</span>
                  <div class="flex items-center gap-2 mt-0.5">
                    <span v-if="!entry.is_dir" class="text-xs text-makoclaw-text-secondary">{{ formatSize(entry.size) }}</span>
                    <span class="text-xs text-makoclaw-text-secondary/60">{{ formatDate(entry.mod_time) }}</span>
                  </div>
                </div>

                <!-- System Badge -->
                <span
                  v-if="entry.is_system"
                  class="px-2 py-1 text-xs rounded-lg bg-amber-500/10 text-amber-400 border border-amber-500/20 flex-shrink-0"
                  title="System-managed file"
                >
                  System
                </span>
              </div>

              <!-- Action Buttons -->
              <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                <button
                  @click.stop="startRename(entry)"
                  :disabled="entry.is_read_only"
                  :title="entry.is_read_only ? 'System file (read-only)' : 'Rename'"
                  :class="entry.is_read_only ? 'opacity-30 cursor-not-allowed' : 'hover:text-indigo-400 hover:bg-indigo-500/10'"
                  class="p-2 text-makoclaw-text-secondary rounded-lg transition-all"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </button>
                <button
                  @click.stop="downloadEntry(entry.path)"
                  class="p-2 text-makoclaw-text-secondary hover:text-indigo-400 hover:bg-indigo-500/10 rounded-lg transition-all"
                  title="Download"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                  </svg>
                </button>
                <button
                  @click.stop="confirmDelete(entry)"
                  :disabled="deletingPath === entry.path || entry.is_read_only"
                  :title="entry.is_read_only ? 'System file (read-only)' : 'Delete'"
                  :class="entry.is_read_only ? 'opacity-30 cursor-not-allowed' : 'hover:text-red-400 hover:bg-red-500/10'"
                  class="p-2 text-makoclaw-text-secondary rounded-lg transition-all"
                >
                  <svg v-if="deletingPath === entry.path" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <!-- Empty State -->
          <div v-if="entries.length === 0" class="flex flex-col items-center justify-center py-16 text-center">
            <div class="relative">
              <div class="absolute inset-0 bg-gradient-to-br from-indigo-500/30 to-violet-500/30 rounded-3xl blur-2xl opacity-50"></div>
              <div class="relative glass-panel p-8 rounded-2xl shadow-2xl ring-1 ring-white/10">
                <div class="w-16 h-16 mx-auto rounded-2xl bg-gradient-to-br from-indigo-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/20">
                  <svg class="w-8 h-8 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                  </svg>
                </div>
              </div>
            </div>
            <h3 class="text-lg font-bold text-makoclaw-text mt-6">
              {{ searchQuery ? 'No results found' : 'Empty folder' }}
            </h3>
            <p class="text-sm text-makoclaw-text-secondary/70 mt-2 max-w-xs">
              {{ searchQuery ? `No files found matching "${searchQuery}"` : 'Drop files here or click Upload to add files' }}
            </p>
          </div>
        </div>

        <!-- File Viewer -->
        <div v-if="fileContent !== null" class="p-4 sm:p-6">
          <div class="glass-panel rounded-2xl overflow-hidden">
            <!-- File Header -->
            <div class="p-4 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent flex flex-wrap items-center justify-between gap-3">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500/20 to-violet-500/20 flex items-center justify-center">
                  <svg class="w-5 h-5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <div>
                  <h3 class="font-bold text-makoclaw-text">{{ fileName }}</h3>
                  <p class="text-xs text-makoclaw-text-secondary">{{ formatSize(fileSize) }}</p>
                </div>
              </div>

              <div class="flex items-center gap-2">
                <!-- Read-only Badge -->
                <span
                  v-if="isFileReadOnly"
                  class="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-amber-500/10 text-amber-400 rounded-lg border border-amber-500/20"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                  Read-only
                </span>

                <!-- Edit Button -->
                <button
                  v-else-if="isTextFile && !isEditing"
                  @click="startEditing"
                  class="flex items-center gap-2 px-3 py-1.5 text-sm bg-indigo-500/10 border border-indigo-500/30 rounded-lg hover:bg-indigo-500/20 transition-colors text-indigo-400"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                  Edit
                </button>

                <!-- Save/Cancel when Editing -->
                <template v-if="isEditing">
                  <button
                    @click="cancelEditing"
                    class="px-3 py-1.5 text-sm bg-makoclaw-surface/50 border border-makoclaw-border/50 rounded-lg hover:bg-makoclaw-surface-hover transition-colors text-makoclaw-text-secondary"
                  >
                    Cancel
                  </button>
                  <button
                    @click="saveFileContent"
                    :disabled="savingFile"
                    class="flex items-center gap-2 px-4 py-1.5 text-sm bg-gradient-to-r from-indigo-500 to-violet-500 text-white rounded-lg hover:from-indigo-600 hover:to-violet-600 transition-colors disabled:opacity-50"
                  >
                    <svg v-if="savingFile" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                    </svg>
                    {{ savingFile ? 'Saving...' : 'Save' }}
                  </button>
                </template>

                <!-- Download -->
                <button
                  @click="downloadEntry(currentPath)"
                  class="flex items-center gap-2 px-3 py-1.5 text-sm bg-makoclaw-surface/50 border border-makoclaw-border/50 rounded-lg hover:bg-makoclaw-surface-hover transition-colors text-makoclaw-text-secondary"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                  </svg>
                  Download
                </button>
              </div>
            </div>

            <!-- File Content -->
            <div class="p-4">
              <div v-if="fileError" class="text-amber-400 text-sm mb-4 p-3 bg-amber-500/10 rounded-lg border border-amber-500/20">
                {{ fileError }}
              </div>
              <textarea
                v-else-if="isEditing"
                v-model="editContent"
                class="w-full bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl p-4 text-sm font-mono overflow-auto min-h-[50vh] max-h-[70vh] whitespace-pre-wrap focus:outline-none focus:ring-2 focus:ring-indigo-500/30 focus:border-indigo-500/50 resize-y"
              ></textarea>
              <pre v-else class="bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl p-4 text-sm font-mono overflow-auto max-h-[70vh] whitespace-pre-wrap">{{ fileContent }}</pre>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- New Folder Modal -->
    <Transition name="modal">
      <div v-if="showNewFolderModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm z-modal flex items-center justify-center p-4">
        <div class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl w-full max-w-md shadow-2xl ring-1 ring-white/10 animate-scaleIn">
          <div class="p-4 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500/20 to-violet-500/20 flex items-center justify-center">
                <svg class="w-4 h-4 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
                </svg>
              </div>
              <h3 class="font-bold text-lg">Create New Folder</h3>
            </div>
          </div>
          <div class="p-4">
            <label class="block text-sm text-makoclaw-text-secondary mb-2">Folder name</label>
            <input
              v-model="newFolderName"
              type="text"
              placeholder="Enter folder name"
              class="w-full px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/30 focus:border-indigo-500/50"
              @keyup.enter="createFolder"
              ref="folderNameInput"
            />
            <p class="text-xs text-makoclaw-text-secondary mt-2">
              Will be created in: <span class="font-mono text-indigo-400">/{{ currentPath || 'workspace' }}</span>
            </p>
          </div>
          <div class="p-4 border-t border-makoclaw-border/30 flex justify-end gap-3">
            <button
              @click="showNewFolderModal = false; newFolderName = ''"
              class="px-4 py-2 text-sm bg-makoclaw-surface/50 border border-makoclaw-border/50 rounded-xl hover:bg-makoclaw-surface-hover transition-colors"
            >
              Cancel
            </button>
            <button
              @click="createFolder"
              :disabled="!newFolderName.trim() || creatingFolder"
              class="px-4 py-2 text-sm bg-gradient-to-r from-indigo-500 to-violet-500 text-white rounded-xl hover:from-indigo-600 hover:to-violet-600 transition-colors disabled:opacity-50"
            >
              {{ creatingFolder ? 'Creating...' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Delete Confirmation Modal -->
    <Transition name="modal">
      <div v-if="deleteConfirmEntry" class="fixed inset-0 bg-black/50 backdrop-blur-sm z-modal flex items-center justify-center p-4">
        <div class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl w-full max-w-md shadow-2xl ring-1 ring-white/10 animate-scaleIn">
          <div class="p-4 border-b border-makoclaw-border/30 bg-gradient-to-r from-red-500/10 to-transparent">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg bg-red-500/20 flex items-center justify-center">
                <svg class="w-4 h-4 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </div>
              <h3 class="font-bold text-lg text-red-400">Confirm Delete</h3>
            </div>
          </div>
          <div class="p-4">
            <p class="text-sm">
              Are you sure you want to delete
              <span class="font-semibold text-makoclaw-text">"{{ deleteConfirmEntry.name }}"</span>?
            </p>
            <p v-if="deleteConfirmEntry.is_dir" class="text-xs text-amber-400 mt-3 p-2 bg-amber-500/10 rounded-lg border border-amber-500/20">
              Warning: This will delete the folder and all its contents.
            </p>
            <p class="text-xs text-makoclaw-text-secondary mt-3">
              This action cannot be undone.
            </p>
          </div>
          <div class="p-4 border-t border-makoclaw-border/30 flex justify-end gap-3">
            <button
              @click="deleteConfirmEntry = null"
              class="px-4 py-2 text-sm bg-makoclaw-surface/50 border border-makoclaw-border/50 rounded-xl hover:bg-makoclaw-surface-hover transition-colors"
            >
              Cancel
            </button>
            <button
              @click="deleteEntry"
              :disabled="deletingPath"
              class="px-4 py-2 text-sm bg-red-500 text-white rounded-xl hover:bg-red-600 transition-colors disabled:opacity-50"
            >
              {{ deletingPath ? 'Deleting...' : 'Delete' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'

const toast = useToast()
const loading = ref(true)
const uploading = ref(false)
const isDragging = ref(false)
const currentPath = ref('')
const entries = ref(null)
const fileContent = ref(null)
const fileName = ref('')
const fileSize = ref(0)
const fileError = ref(null)

// Search state
const searchQuery = ref('')
const isSearching = ref(false)
let searchTimeout = null

// New folder state
const showNewFolderModal = ref(false)
const newFolderName = ref('')
const creatingFolder = ref(false)
const folderNameInput = ref(null)

// Delete state
const deleteConfirmEntry = ref(null)
const deletingPath = ref(null)

// Rename state
const renamingEntry = ref(null)
const renameValue = ref('')
const renameInput = ref(null)

// File editing state
const isEditing = ref(false)
const editContent = ref('')
const savingFile = ref(false)

// Text file extensions that can be edited
const textExtensions = ['.txt', '.md', '.json', '.js', '.ts', '.jsx', '.tsx', '.vue', '.css', '.scss', '.html', '.xml', '.yaml', '.yml', '.toml', '.ini', '.cfg', '.conf', '.log', '.sh', '.bash', '.zsh', '.py', '.go', '.rs', '.java', '.c', '.cpp', '.h', '.hpp', '.rb', '.php', '.sql', '.env', '.gitignore', '.dockerfile', '.makefile']

const isTextFile = computed(() => {
  if (!fileName.value) return false
  const name = fileName.value.toLowerCase()
  return textExtensions.some(ext => name.endsWith(ext)) || !name.includes('.')
})

const isFileReadOnly = computed(() => {
  const path = currentPath.value
  const systemPaths = ['cron/jobs.json', 'database.db']
  const systemDirs = ['sessions/', 'memory/', 'knowledge/', '.context/']
  return systemPaths.includes(path) || systemDirs.some(dir => path.startsWith(dir))
})

// Watch for modal open to focus input
watch(showNewFolderModal, async (isOpen) => {
  if (isOpen) {
    await nextTick()
    folderNameInput.value?.focus()
  }
})

const handleFileUpload = async (event) => {
  const files = event.target.files
  if (!files || files.length === 0) return
  await uploadFiles(files)
  event.target.value = ''
}

const onDrop = async (event) => {
  isDragging.value = false
  const files = event.dataTransfer.files
  if (!files || files.length === 0) return
  await uploadFiles(files)
}

const uploadFiles = async (files) => {
  uploading.value = true
  let successCount = 0

  for (const file of Array.from(files)) {
    try {
      await advancedService.uploadFile(currentPath.value, file)
      successCount++
    } catch (err) {
      console.error(`Failed to upload ${file.name}:`, err)
      toast.error(`Failed to upload ${file.name}`)
    }
  }

  if (successCount > 0) {
    toast.success(`Successfully uploaded ${successCount} file(s)`)
    await navigateTo(currentPath.value)
  }
  uploading.value = false
}

const breadcrumbs = computed(() => {
  if (!currentPath.value) return []
  return currentPath.value.split('/').filter(Boolean)
})

const navigateTo = async (path) => {
  loading.value = true
  fileContent.value = null
  isEditing.value = false
  editContent.value = ''
  currentPath.value = path
  searchQuery.value = ''
  try {
    const data = await advancedService.fetchFiles(path)
    if (data.entries !== undefined) {
      entries.value = data.entries || []
    } else if (data.content !== undefined) {
      entries.value = null
      fileContent.value = data.content
      fileName.value = data.name
      fileSize.value = data.size
      fileError.value = null
    } else if (data.error) {
      entries.value = null
      fileContent.value = null
      fileError.value = data.error
      fileName.value = data.name
      fileSize.value = data.size
    }
  } catch (err) {
    console.error('Failed to browse files:', err)
    toast.error('Failed to load files')
  } finally {
    loading.value = false
  }
}

const viewFile = async (path) => {
  await navigateTo(path)
}

const downloadEntry = (path) => {
  advancedService.downloadFile(path)
}

const navigateUp = () => {
  const parts = currentPath.value.split('/').filter(Boolean)
  parts.pop()
  navigateTo(parts.join('/'))
}

const debouncedSearch = () => {
  if (searchTimeout) clearTimeout(searchTimeout)

  if (!searchQuery.value.trim()) {
    navigateTo(currentPath.value)
    return
  }

  isSearching.value = true
  searchTimeout = setTimeout(async () => {
    try {
      const data = await advancedService.searchFiles(currentPath.value, searchQuery.value)
      entries.value = data.entries || []
    } catch (err) {
      console.error('Search failed:', err)
      toast.error('Search failed')
    } finally {
      isSearching.value = false
    }
  }, 300)
}

const clearSearch = () => {
  searchQuery.value = ''
  navigateTo(currentPath.value)
}

const createFolder = async () => {
  const folderName = newFolderName.value.trim()
  if (!folderName) return

  creatingFolder.value = true
  try {
    const path = currentPath.value ? `${currentPath.value}/${folderName}` : folderName
    await advancedService.createFolder(path)
    toast.success(`Folder "${folderName}" created`)
    showNewFolderModal.value = false
    newFolderName.value = ''
    await navigateTo(currentPath.value)
  } catch (err) {
    console.error('Failed to create folder:', err)
    toast.error('Failed to create folder')
  } finally {
    creatingFolder.value = false
  }
}

const confirmDelete = (entry) => {
  deleteConfirmEntry.value = entry
}

const deleteEntry = async () => {
  if (!deleteConfirmEntry.value) return

  const entry = deleteConfirmEntry.value
  deletingPath.value = entry.path

  try {
    await advancedService.deleteFile(entry.path)
    toast.success(`"${entry.name}" deleted`)
    deleteConfirmEntry.value = null
    await navigateTo(currentPath.value)
  } catch (err) {
    console.error('Failed to delete:', err)
    toast.error('Failed to delete')
  } finally {
    deletingPath.value = null
  }
}

const startRename = async (entry) => {
  renamingEntry.value = entry.path
  renameValue.value = entry.name
  await nextTick()
  const input = document.querySelector('input[type="text"][class*="border-indigo"]')
  if (input) {
    input.focus()
    const dotIndex = entry.is_dir ? -1 : entry.name.lastIndexOf('.')
    if (dotIndex > 0) {
      input.setSelectionRange(0, dotIndex)
    } else {
      input.select()
    }
  }
}

const cancelRename = () => {
  renamingEntry.value = null
  renameValue.value = ''
}

const confirmRename = async (entry) => {
  const newName = renameValue.value.trim()
  if (!newName || newName === entry.name) {
    cancelRename()
    return
  }

  try {
    await advancedService.renameFile(entry.path, newName)
    toast.success(`Renamed to "${newName}"`)
    cancelRename()
    await navigateTo(currentPath.value)
  } catch (err) {
    console.error('Failed to rename:', err)
    toast.error('Failed to rename')
    cancelRename()
  }
}

const startEditing = () => {
  editContent.value = fileContent.value
  isEditing.value = true
}

const cancelEditing = () => {
  isEditing.value = false
  editContent.value = ''
}

const saveFileContent = async () => {
  savingFile.value = true
  try {
    await advancedService.updateFileContent(currentPath.value, editContent.value)
    fileContent.value = editContent.value
    isEditing.value = false
    toast.success('File saved')
  } catch (err) {
    console.error('Failed to save file:', err)
    toast.error('Failed to save file')
  } finally {
    savingFile.value = false
  }
}

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) {
    bytes /= 1024
    i++
  }
  return `${bytes.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString()
}

onMounted(() => navigateTo(''))
</script>

<style scoped>
.file-card {
  @apply flex items-center gap-3 px-4 py-3 rounded-xl bg-makoclaw-surface/30 border border-makoclaw-border/30 hover:bg-makoclaw-surface/50 hover:border-indigo-500/20 transition-all duration-200;
}
</style>
