<template>
  <div class="h-full flex flex-col bg-kakoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-kakoclaw-border bg-kakoclaw-surface flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold bg-gradient-to-r from-kakoclaw-accent to-purple-500 bg-clip-text text-transparent">Settings</h2>
        <p class="text-sm text-kakoclaw-text-secondary mt-1">Configure your agent, providers, and channels</p>
      </div>
      <div class="flex bg-kakoclaw-bg rounded-lg p-1 border border-kakoclaw-border overflow-x-auto max-w-[50%] sm:max-w-none">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          class="px-3 py-1.5 rounded-md text-sm font-medium transition-all whitespace-nowrap"
          :class="activeTab === tab.key ? 'bg-white dark:bg-gray-700 shadow-sm text-kakoclaw-accent' : 'text-kakoclaw-text-secondary hover:text-kakoclaw-text'"
        >{{ tab.label }}</button>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-4 sm:p-6 custom-scrollbar">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-kakoclaw-accent"></div>
      </div>

      <template v-else-if="configData">
        <!-- Agent Settings -->
        <div v-if="activeTab === 'agents'" class="space-y-6 max-w-2xl mx-auto">
          <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-6">
            <h3 class="font-semibold mb-6 flex items-center text-kakoclaw-text">
               <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-kakoclaw-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
               </svg>
               Agent Defaults
            </h3>
            
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-kakoclaw-text-secondary mb-1">Default Provider</label>
                <select v-model="configData.agents.defaults.provider" class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-kakoclaw-accent outline-none text-kakoclaw-text">
                  <option v-for="p in providersList" :key="p.name" :value="p.name">{{ p.name }}</option>
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium text-kakoclaw-text-secondary mb-1">Default Model</label>
                <select v-model="configData.agents.defaults.model" class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-kakoclaw-accent outline-none text-kakoclaw-text">
                  <optgroup v-for="p in providersList" :key="p.name" :label="p.name">
                    <option v-for="m in p.models" :key="m.id" :value="m.id">{{ m.id }}</option>
                  </optgroup>
                </select>
              </div>

              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium text-kakoclaw-text-secondary mb-1">Temperature</label>
                  <input v-model.number="configData.agents.defaults.temperature" type="number" step="0.1" min="0" max="2" class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-3 py-2 text-sm outline-none text-kakoclaw-text">
                </div>
                <div>
                  <label class="block text-sm font-medium text-kakoclaw-text-secondary mb-1">Max Tokens</label>
                  <input v-model.number="configData.agents.defaults.max_tokens" type="number" class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-3 py-2 text-sm outline-none text-kakoclaw-text">
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-kakoclaw-text-secondary mb-1">Max Tool Iterations</label>
                <input v-model.number="configData.agents.defaults.max_tool_iterations" type="number" class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-3 py-2 text-sm outline-none text-kakoclaw-text">
              </div>

              <div class="pt-4 border-t border-kakoclaw-border mt-6">
                <button @click="saveConfig({agents: configData.agents})" :disabled="saving" class="w-full bg-kakoclaw-accent text-white py-2 rounded-lg font-medium hover:bg-kakoclaw-accent-hover transition-colors flex items-center justify-center disabled:opacity-50">
                  <span v-if="saving" class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin mr-2"></span>
                  Save Agent Settings
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Providers -->
        <div v-if="activeTab === 'providers'" class="space-y-4 max-w-4xl mx-auto">
          <div
            v-for="(info, name) in configData.providers"
            :key="name"
            class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5"
          >
            <div class="flex items-center justify-between mb-4">
              <h3 class="font-semibold capitalize text-lg flex items-center text-kakoclaw-text">
                 <span class="w-8 h-8 rounded-lg bg-kakoclaw-bg border border-kakoclaw-border flex items-center justify-center mr-3 text-kakoclaw-accent text-xs font-bold">{{ name.substring(0,2).toUpperCase() }}</span>
                 {{ name }}
              </h3>
              <span
                class="px-2 py-0.5 text-xs rounded-full"
                :class="info.configured ? 'bg-emerald-500/10 text-emerald-400' : 'bg-gray-500/10 text-gray-400'"
              >{{ info.configured ? 'Configured' : 'Not configured' }}</span>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 items-end">
              <div>
                <label class="block text-xs font-bold text-kakoclaw-text-secondary mb-1 uppercase tracking-tight">API Key</label>
                <input 
                  v-model="info.api_key" 
                  type="password" 
                  :placeholder="info.configured ? '••••••••••••••••' : 'Enter API Key'" 
                  class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-3 py-2 text-sm outline-none focus:border-kakoclaw-accent text-kakoclaw-text"
                >
              </div>
              <div class="flex gap-2">
                <div class="flex-1">
                  <label class="block text-xs font-bold text-kakoclaw-text-secondary mb-1 uppercase tracking-tight">API Base (optional)</label>
                  <input 
                    v-model="info.api_base" 
                    type="text" 
                    placeholder="https://api..." 
                    class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-3 py-2 text-sm outline-none focus:border-kakoclaw-accent text-kakoclaw-text"
                  >
                </div>
                <button 
                  @click="saveConfig({providers: {[name]: {api_key: info.api_key, api_base: info.api_base}}})" 
                  :disabled="saving"
                  class="bg-kakoclaw-accent text-white px-4 h-9 rounded-lg hover:bg-kakoclaw-accent-hover transition-colors disabled:opacity-50 flex items-center"
                >
                  <span v-if="saving" class="w-3 h-3 border-2 border-white/20 border-t-white rounded-full animate-spin mr-2"></span>
                  Save
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Channels -->
        <div v-if="activeTab === 'channels'" class="space-y-4 max-w-5xl mx-auto">
           <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <div 
                v-for="channel in availableChannels" 
                :key="channel.id"
                class="bg-kakoclaw-surface rounded-xl border border-kakoclaw-border p-5 transition-all hover:shadow-md"
              >
                <div class="flex items-center justify-between mb-4">
                  <div class="flex items-center space-x-3">
                    <div 
                      class="w-10 h-10 rounded-lg flex items-center justify-center bg-kakoclaw-bg border border-kakoclaw-border text-kakoclaw-text-secondary"
                      :class="{'bg-kakoclaw-accent !text-white border-transparent shadow-sm': configData.channels[channel.id]?.enabled}"
                      v-html="channel.icon"
                    ></div>
                    <h3 class="font-medium text-kakoclaw-text">{{ channel.name }}</h3>
                  </div>
                  <button 
                    @click="toggleChannel(channel.id)"
                    class="relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors duration-200 focus:outline-none"
                    :class="configData.channels[channel.id]?.enabled ? 'bg-kakoclaw-success' : 'bg-kakoclaw-border'"
                  >
                    <span 
                      aria-hidden="true" 
                      class="pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition duration-200"
                      :class="configData.channels[channel.id]?.enabled ? 'translate-x-5' : 'translate-x-0'"
                    ></span>
                  </button>
                </div>
                <p class="text-xs text-kakoclaw-text-secondary mb-6 h-8 line-clamp-2">{{ channel.description }}</p>
                <div class="flex gap-2">
                  <button 
                    @click="openChannelConfig(channel)"
                    class="flex-1 py-1.5 text-xs font-medium border border-kakoclaw-border rounded-lg hover:bg-kakoclaw-bg transition-colors flex items-center justify-center text-kakoclaw-text"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    Settings
                  </button>
                  <div 
                    v-if="configData.channels[channel.id]?.enabled"
                    class="px-2 py-1.5 border border-kakoclaw-border rounded-lg bg-emerald-500/10 flex items-center"
                    title="Active"
                  >
                    <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                  </div>
                </div>
              </div>
           </div>
        </div>

        <!-- System -->
        <div v-if="activeTab === 'system'" class="space-y-6 max-w-4xl mx-auto">
          <!-- Web & Gateway -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5">
              <h3 class="font-bold mb-4 text-kakoclaw-text-secondary uppercase tracking-widest text-[10px]">Web Server</h3>
              <div class="space-y-3">
                <div v-for="(val, key) in configData.web" :key="key" class="flex justify-between items-center py-1">
                  <span class="text-sm text-kakoclaw-text-secondary">{{ formatKey(key) }}</span>
                  <span class="text-sm font-mono text-kakoclaw-text">{{ String(val) }}</span>
                </div>
              </div>
            </div>
            <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5">
              <h3 class="font-bold mb-4 text-kakoclaw-text-secondary uppercase tracking-widest text-[10px]">Gateway</h3>
              <div class="space-y-3">
                <div v-for="(val, key) in configData.gateway" :key="key" class="flex justify-between items-center py-1">
                  <span class="text-sm text-kakoclaw-text-secondary">{{ formatKey(key) }}</span>
                  <span class="text-sm font-mono text-kakoclaw-text">{{ String(val) }}</span>
                </div>
              </div>
            </div>
          </div>
          
          <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5">
            <h3 class="font-bold mb-4 text-kakoclaw-text-secondary uppercase tracking-widest text-[10px]">Storage & Backend</h3>
            <div class="space-y-3">
              <div class="flex justify-between items-center py-1">
                <span class="text-sm text-kakoclaw-text-secondary">Database Path</span>
                <span class="text-sm font-mono text-kakoclaw-text text-right truncate ml-4">{{ configData.storage?.path || '(not set)' }}</span>
              </div>
            </div>
          </div>

          <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-6">
            <div class="flex justify-between items-center mb-6">
              <h3 class="font-bold text-kakoclaw-text-secondary uppercase tracking-widest text-[10px]">Search Utilities</h3>
              <button 
                @click="saveConfig({tools: configData.tools})" 
                :disabled="saving"
                class="text-kakoclaw-accent hover:text-kakoclaw-accent-hover text-xs font-bold uppercase disabled:opacity-50"
              >
                {{ saving ? 'Updating...' : 'Save Updates' }}
              </button>
            </div>
            <div class="space-y-5">
               <div>
                  <label class="block text-xs font-semibold text-kakoclaw-text-secondary mb-1.5">Web Search API Key</label>
                  <input v-model="configData.tools.web.search.api_key" type="password" placeholder="••••••••••••••••" class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-3 py-2 text-sm outline-none focus:border-kakoclaw-accent text-kakoclaw-text">
               </div>
               <div>
                  <label class="block text-xs font-semibold text-kakoclaw-text-secondary mb-1.5">Max Search Results</label>
                  <input v-model.number="configData.tools.web.search.max_results" type="number" class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-3 py-2 text-sm outline-none focus:border-kakoclaw-accent text-kakoclaw-text">
               </div>
             </div>
           </div>

          <!-- Backup Section -->
          <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-6">
             <div class="flex justify-between items-center mb-6">
               <h3 class="font-bold text-kakoclaw-text-secondary uppercase tracking-widest text-[10px]">Backup & Restore</h3>
             </div>

             <!-- Export Section -->
             <div class="space-y-4 mb-8">
               <h4 class="font-semibold text-kakoclaw-text flex items-center">
                 <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-kakoclaw-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                   <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                 </svg>
                 Export Backup
               </h4>

               <div class="space-y-3 p-4 bg-kakoclaw-bg/50 rounded-lg">
                 <label class="flex items-center space-x-3">
                   <input type="checkbox" v-model="exportOptions.include_database" checked class="rounded border-kakoclaw-border text-kakoclaw-accent focus:ring-kakoclaw-accent">
                   <span class="text-sm text-kakoclaw-text">Database & Sessions</span>
                 </label>
                 <label class="flex items-center space-x-3">
                   <input type="checkbox" v-model="exportOptions.include_workspace" checked class="rounded border-kakoclaw-border text-kakoclaw-accent focus:ring-kakoclaw-accent">
                   <span class="text-sm text-kakoclaw-text">Workspace & Skills</span>
                 </label>
                 <label class="flex items-center space-x-3">
                   <input type="checkbox" v-model="exportOptions.include_config" class="rounded border-kakoclaw-border text-kakoclaw-accent focus:ring-kakoclaw-accent">
                   <span class="text-sm text-kakoclaw-text">Configuration (config.json)</span>
                 </label>
                 <label class="flex items-center space-x-3">
                   <input type="checkbox" v-model="exportOptions.include_env" class="rounded border-kakoclaw-border text-kakoclaw-accent focus:ring-kakoclaw-accent">
                   <span class="text-sm text-kakoclaw-text">Environment Variables (.env)</span>
                 </label>
                 <p v-if="exportOptions.include_env" class="text-xs text-orange-400 mt-2 flex items-center">
                   <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                     <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                   </svg>
                   ⚠️ Contains sensitive data (API keys, passwords)
                 </p>
               </div>

               <button
                 @click="exportBackup"
                 :disabled="exporting || (!exportOptions.include_database && !exportOptions.include_workspace && !exportOptions.include_config && !exportOptions.include_env)"
                 class="w-full bg-kakoclaw-accent text-white py-2.5 rounded-lg font-medium hover:bg-kakoclaw-accent-hover transition-colors flex items-center justify-center disabled:opacity-50 disabled:cursor-not-allowed"
               >
                 <svg v-if="exporting" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                   <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                   <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                 </svg>
                 <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                   <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                 </svg>
                 {{ exporting ? 'Creating Backup...' : 'Download Backup (.kakoclaw)' }}
               </button>
             </div>

             <!-- Import Section -->
             <div class="space-y-4">
               <h4 class="font-semibold text-kakoclaw-text flex items-center">
                 <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-orange-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                   <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                 </svg>
                 Import Backup
               </h4>

               <div class="border-2 border-dashed border-kakoclaw-border rounded-xl p-8 transition-colors hover:border-kakoclaw-accent/50">
                 <input type="file" @change="handleFileSelect" accept=".kakoclaw" class="hidden" ref="fileInput">
                 <button
                   @click="$refs.fileInput.click()"
                   :disabled="importing"
                   class="w-full flex flex-col items-center justify-center space-y-2 text-kakoclaw-text-secondary hover:text-kakoclaw-text transition-colors disabled:opacity-50"
                 >
                   <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                     <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                   </svg>
                   <span class="font-medium">Select .kakoclaw File</span>
                   <span class="text-xs">or drag and drop here</span>
                 </button>

                 <!-- File Preview -->
                 <div v-if="selectedFile" class="mt-6 space-y-4">
                   <div class="flex items-center justify-between p-3 bg-kakoclaw-bg rounded-lg">
                     <div class="flex items-center space-x-3">
                       <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-kakoclaw-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                         <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                       </svg>
                       <div>
                         <p class="font-medium text-sm text-kakoclaw-text">{{ selectedFile.name }}</p>
                         <p class="text-xs text-kakoclaw-text-secondary">{{ formatBytes(selectedFile.size) }}</p>
                       </div>
                     </div>
                     <button @click="clearSelectedFile" class="text-kakoclaw-text-secondary hover:text-red-400 transition-colors">
                       <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                         <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                       </svg>
                     </button>
                   </div>

                   <!-- Import Options -->
                   <div v-if="validationResult" class="space-y-3 p-4 bg-kakoclaw-bg/50 rounded-lg">
                     <p class="text-sm font-medium text-kakoclaw-text mb-2">Backup Information:</p>
                     <div class="grid grid-cols-2 gap-2 text-xs">
                       <div><span class="text-kakoclaw-text-secondary">Version:</span> <span class="text-kakoclaw-text">{{ validationResult.version }}</span></div>
                       <div><span class="text-kakoclaw-text-secondary">Files:</span> <span class="text-kakoclaw-text">{{ validationResult.total_files }}</span></div>
                       <div><span class="text-kakoclaw-text-secondary">Size:</span> <span class="text-kakoclaw-text">{{ formatBytes(validationResult.data_size_bytes) }}</span></div>
                       <div><span class="text-kakoclaw-text-secondary">Created:</span> <span class="text-kakoclaw-text">{{ formatDate(validationResult.created_at) }}</span></div>
                     </div>

                     <div class="space-y-2 pt-2 border-t border-kakoclaw-border">
                       <label class="flex items-center space-x-2">
                         <input type="checkbox" v-model="importOptions.replace_database" class="rounded border-kakoclaw-border text-kakoclaw-accent focus:ring-kakoclaw-accent" :disabled="!validationResult.includes_database">
                         <span class="text-sm text-kakoclaw-text" :class="{'opacity-50': !validationResult.includes_database}">Replace Database & Sessions</span>
                       </label>
                       <label class="flex items-center space-x-2">
                         <input type="checkbox" v-model="importOptions.replace_workspace" class="rounded border-kakoclaw-border text-kakoclaw-accent focus:ring-kakoclaw-accent" :disabled="!validationResult.includes_workspace">
                         <span class="text-sm text-kakoclaw-text" :class="{'opacity-50': !validationResult.includes_workspace}">Replace Workspace & Skills</span>
                       </label>
                       <label class="flex items-center space-x-2">
                         <input type="checkbox" v-model="importOptions.replace_config" class="rounded border-kakoclaw-border text-kakoclaw-accent focus:ring-kakoclaw-accent" :disabled="!validationResult.includes_config">
                         <span class="text-sm text-kakoclaw-text" :class="{'opacity-50': !validationResult.includes_config}">Replace Configuration</span>
                       </label>
                       <label class="flex items-center space-x-2">
                         <input type="checkbox" v-model="importOptions.replace_env" class="rounded border-kakoclaw-border text-kakoclaw-accent focus:ring-kakoclaw-accent" :disabled="!validationResult.includes_env">
                         <span class="text-sm text-kakoclaw-text" :class="{'opacity-50': !validationResult.includes_env}">Replace Environment Variables</span>
                       </label>
                     </div>

                     <p class="text-xs text-orange-400 flex items-center">
                       <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                         <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                       </svg>
                       Existing files will be backed up automatically
                     </p>
                   </div>

                   <button
                     @click="importBackup"
                     :disabled="importing || !validationResult || (!importOptions.replace_database && !importOptions.replace_workspace && !importOptions.replace_config && !importOptions.replace_env)"
                     class="w-full bg-orange-500 hover:bg-orange-600 text-white py-2.5 rounded-lg font-medium transition-colors flex items-center justify-center disabled:opacity-50 disabled:cursor-not-allowed"
                   >
                     <svg v-if="importing" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                       <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                       <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                     </svg>
                     <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                       <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                     </svg>
                     {{ importing ? 'Importing Backup...' : 'Import Backup' }}
                   </button>
                 </div>
               </div>
             </div>
          </div>
         </div>
       </template>
     </div>

    <!-- Channel Config Modal -->
    <div v-if="showChannelModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="bg-kakoclaw-surface rounded-2xl shadow-2xl w-full max-w-md border border-kakoclaw-border overflow-hidden animate-in fade-in zoom-in duration-200">
        <div class="flex justify-between items-center p-6 border-b border-kakoclaw-border bg-kakoclaw-bg/20">
          <h3 class="text-lg font-bold text-kakoclaw-text flex items-center">
            <span class="w-7 h-7 mr-3 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg flex items-center justify-center text-kakoclaw-text-secondary scale-90" v-html="selectedChannel?.icon"></span>
            Configure {{ selectedChannel?.name }}
          </h3>
          <button @click="showChannelModal = false" class="text-kakoclaw-text-secondary hover:text-kakoclaw-text flex items-center justify-center w-8 h-8 rounded-full hover:bg-kakoclaw-bg transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        <div class="p-6 space-y-5">
          <div v-if="selectedChannel?.id === 'telegram'" class="space-y-4">
             <div>
                <label class="block text-xs font-bold text-kakoclaw-text-secondary mb-1.5 uppercase">Bot Token</label>
                <input v-model="channelForm.token" type="password" placeholder="123456:ABC..." class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-xl text-sm outline-none focus:ring-2 focus:ring-kakoclaw-accent/20 focus:border-kakoclaw-accent text-kakoclaw-text">
             </div>
             <div>
                <label class="block text-xs font-bold text-kakoclaw-text-secondary mb-1.5 uppercase">Allowed Usernames/IDs</label>
                <input v-model="channelForm.allow_from" type="text" placeholder="user1,1234567" class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-xl text-sm outline-none focus:ring-2 focus:ring-kakoclaw-accent/20 focus:border-kakoclaw-accent text-kakoclaw-text">
                <p class="text-[10px] text-kakoclaw-text-secondary mt-1.5 ml-1">Comma separated list of users who can use the bot.</p>
             </div>
          </div>

          <div v-else-if="selectedChannel?.id === 'discord'" class="space-y-4">
             <div>
                <label class="block text-xs font-bold text-kakoclaw-text-secondary mb-1.5 uppercase">Bot Token</label>
                <input v-model="channelForm.token" type="password" placeholder="MTIz..." class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-xl text-sm outline-none focus:ring-2 focus:ring-kakoclaw-accent/20 focus:border-kakoclaw-accent text-kakoclaw-text">
             </div>
             <div>
                <label class="block text-xs font-bold text-kakoclaw-text-secondary mb-1.5 uppercase">Allowed Server/Channel IDs</label>
                <input v-model="channelForm.allow_from" type="text" class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-xl text-sm outline-none focus:ring-2 focus:ring-kakoclaw-accent/20 focus:border-kakoclaw-accent text-kakoclaw-text">
             </div>
          </div>

          <div v-else-if="selectedChannel?.id === 'slack'" class="space-y-4">
             <div>
                <label class="block text-xs font-bold text-kakoclaw-text-secondary mb-1.5 uppercase">Bot Token (xoxb-...)</label>
                <input v-model="channelForm.bot_token" type="password" placeholder="xoxb-..." class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-xl text-sm outline-none focus:ring-2 focus:ring-kakoclaw-accent/20 focus:border-kakoclaw-accent text-kakoclaw-text">
             </div>
          </div>

          <div v-else class="py-10 text-center">
             <div class="w-12 h-12 rounded-full bg-kakoclaw-bg border border-kakoclaw-border flex items-center justify-center mx-auto mb-3 opacity-50">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-kakoclaw-text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                   <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
                </svg>
             </div>
             <p class="text-sm text-kakoclaw-text-secondary font-medium">Advanced settings needed</p>
             <p class="text-xs text-kakoclaw-text-secondary mt-1">Please edit config.json directly for this channel.</p>
          </div>
        </div>

        <div class="flex justify-end space-x-3 p-6 border-t border-kakoclaw-border bg-kakoclaw-bg/20">
          <button @click="showChannelModal = false" class="px-4 py-2 text-sm font-medium text-kakoclaw-text-secondary hover:text-kakoclaw-text transition-colors">Cancel</button>
          <button @click="saveChannelConfig" :disabled="saving" class="px-6 py-2 text-sm font-bold bg-kakoclaw-accent text-white rounded-xl shadow-lg shadow-kakoclaw-accent/20 hover:bg-kakoclaw-accent-hover transition-all flex items-center disabled:opacity-50">
            <span v-if="saving" class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin mr-2"></span>
            Apply & Restart
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'

const toast = useToast()
const loading = ref(true)
const saving = ref(false)
const configData = ref(null)
const providersList = ref([])
const activeTab = ref('agents')

const tabs = [
  { key: 'agents', label: 'General' },
  { key: 'providers', label: 'Providers' },
  { key: 'channels', label: 'Channels' },
  { key: 'system', label: 'System' }
]

const chatIcon = '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6 text-current"><path stroke-linecap="round" stroke-linejoin="round" d="M8.625 12a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0H8.25m4.125 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0H12m4.125 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 0 1-2.555-.337A5.972 5.972 0 0 1 5.41 20.97a5.969 5.969 0 0 1-.474-.065 4.48 4.48 0 0 0 .978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25Z" /></svg>'
const hashIcon = '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6 text-current"><path stroke-linecap="round" stroke-linejoin="round" d="M5.25 8.25h13.5m-13.5 7.5h13.5m-3-10.5-3 15m-3-15-3 15" /></svg>'

const availableChannels = [
  { id: 'telegram', name: 'Telegram', icon: chatIcon, description: 'Chat with agent via Telegram bot.' },
  { id: 'discord', name: 'Discord', icon: hashIcon, description: 'Connect agent to Discord servers.' },
  { id: 'slack', name: 'Slack', icon: hashIcon, description: 'Integrate into Slack workspaces.' },
  { id: 'whatsapp', name: 'WhatsApp', icon: chatIcon, description: 'Connect via WhatsApp bridge.' },
  { id: 'feishu', name: 'Feishu / Lark', icon: chatIcon, description: 'Enterprise collaboration platform.' },
  { id: 'signal', name: 'Signal', icon: chatIcon, description: 'Secure messaging via Signal.' }
]

// Modal stuff
const showChannelModal = ref(false)
const selectedChannel = ref(null)
const channelForm = ref({})

// Backup stuff
const exporting = ref(false)
const importing = ref(false)
const selectedFile = ref(null)
const validationResult = ref(null)
const fileInput = ref(null)

const exportOptions = ref({
  include_database: true,
  include_workspace: true,
  include_config: false,
  include_env: false
})

const importOptions = ref({
  replace_database: true,
  replace_workspace: true,
  replace_config: true,
  replace_env: true
})

const formatKey = (key) => key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())

const formatBytes = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A'
  const date = new Date(dateStr)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
}

const exportBackup = async () => {
  exporting.value = true
  try {
    const params = new URLSearchParams({
      include_database: exportOptions.value.include_database,
      include_workspace: exportOptions.value.include_workspace,
      include_config: exportOptions.value.include_config,
      include_env: exportOptions.value.include_env
    })

    const response = await fetch(`/api/v1/backup/export?${params}`)
    if (!response.ok) {
      throw new Error('Failed to export backup')
    }

    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `kakoclaw-${new Date().toISOString().split('T')[0]}.kakoclaw`
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)

    toast.success('Backup exported successfully')
  } catch (err) {
    console.error(err)
    toast.error('Failed to export backup: ' + err.message)
  } finally {
    exporting.value = false
  }
}

const handleFileSelect = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  if (!file.name.endsWith('.kakoclaw')) {
    toast.error('Please select a .kakoclaw file')
    return
  }

  selectedFile.value = file
  validationResult.value = null

  try {
    const formData = new FormData()
    formData.append('file', file)

    const response = await fetch('/api/v1/backup/validate', {
      method: 'POST',
      body: formData
    })

    if (!response.ok) {
      throw new Error('Failed to validate backup')
    }

    const result = await response.json()
    if (result.valid) {
      validationResult.value = result
      importOptions.value = {
        replace_database: result.includes_database,
        replace_workspace: result.includes_workspace,
        replace_config: result.includes_config,
        replace_env: result.includes_env
      }
    } else {
      toast.error('Invalid backup file: ' + result.error)
      selectedFile.value = null
    }
  } catch (err) {
    console.error(err)
    toast.error('Failed to validate backup: ' + err.message)
    selectedFile.value = null
  }
}

const clearSelectedFile = () => {
  selectedFile.value = null
  validationResult.value = null
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

const importBackup = async () => {
  if (!selectedFile.value || !validationResult.value) {
    toast.error('Please select a valid backup file')
    return
  }

  if (!importOptions.value.replace_database && !importOptions.value.replace_workspace && !importOptions.value.replace_config && !importOptions.value.replace_env) {
    toast.error('Please select at least one item to import')
    return
  }

  importing.value = true
  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)
    formData.append('options', JSON.stringify(importOptions.value))

    const response = await fetch('/api/v1/backup/import', {
      method: 'POST',
      body: formData
    })

    if (!response.ok) {
      throw new Error('Failed to import backup')
    }

    const result = await response.json()
    if (result.ok) {
      toast.success('Backup imported successfully')
      clearSelectedFile()
    } else {
      toast.error('Failed to import backup')
    }
  } catch (err) {
    console.error(err)
    toast.error('Failed to import backup: ' + err.message)
  } finally {
    importing.value = false
  }
}

const loadData = async () => {
  loading.value = true
  try {
    const [cfg, mods] = await Promise.all([
      advancedService.fetchConfig(),
      advancedService.fetchModels()
    ])
    configData.value = cfg.config || {}
    providersList.value = mods.providers || []
  } catch (err) {
    console.error(err)
    toast.error('Failed to load configuration')
  } finally {
    loading.value = false
  }
}

const saveConfig = async (payload) => {
  saving.value = true
  try {
    await advancedService.updateConfig(payload)
    toast.success('Configuration updated successfully')
    // Wait for server to restart channels/processes if needed
    setTimeout(loadData, 500)
  } catch (err) {
    const detail = err.response?.data?.error || err.message
    toast.error('Update failed: ' + detail)
  } finally {
    saving.value = false
  }
}

const toggleChannel = async (id) => {
  const isEnabled = !configData.value.channels[id]?.enabled
  const payload = {
    channels: {
      [id]: { enabled: isEnabled }
    }
  }
  await saveConfig(payload)
}

const openChannelConfig = (channel) => {
  selectedChannel.value = channel
  const current = configData.value.channels[channel.id] || {}
  
  // Initialize form based on channel type
  if (channel.id === 'telegram') {
    channelForm.value = { token: '', allow_from: current.allow_from || '' }
  } else if (channel.id === 'discord') {
    channelForm.value = { token: '', allow_from: current.allow_from || '' }
  } else if (channel.id === 'slack') {
    channelForm.value = { bot_token: '' }
  } else {
    channelForm.value = {}
  }
  
  showChannelModal.value = true
}

const saveChannelConfig = async () => {
  // Only send fields that are not empty (to avoid overwriting with empty tokens)
  const updates = { enabled: configData.value.channels[selectedChannel.value.id]?.enabled }
  for (const [k, v] of Object.entries(channelForm.value)) {
    if (v !== '') updates[k] = v
  }

  const payload = {
    channels: {
      [selectedChannel.value.id]: updates
    }
  }
  await saveConfig(payload)
  showChannelModal.value = false
}

onMounted(loadData)
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background-color: rgba(156, 163, 175, 0.2); border-radius: 10px; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background-color: rgba(156, 163, 175, 0.4); }

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;  
  overflow: hidden;
}
</style>
