<template>
  <div class="min-h-screen bg-gradient-to-br from-makoclaw-bg via-makoclaw-surface to-makoclaw-bg">
    <!-- Navigation -->
    <nav class="border-b border-makoclaw-border bg-makoclaw-surface/80 backdrop-blur-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white font-bold text-lg">🦈</div>
            <h1 class="text-2xl font-bold bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">
              MakoClaw
            </h1>
          </div>
          <div class="flex items-center gap-4">
            <router-link
              to="/login"
              class="px-4 py-2 text-makoclaw-text hover:text-blue-500 transition-colors font-medium"
            >
              Sign In
            </router-link>
            <router-link
              to="/signup"
              class="px-6 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition-all shadow-lg shadow-blue-500/20 font-medium"
            >
              Get Started
            </router-link>
          </div>
        </div>
      </div>
    </nav>

    <!-- Hero Section -->
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
      <div class="text-center max-w-4xl mx-auto">
        <div class="inline-flex items-center gap-2 px-4 py-2 bg-blue-500/10 rounded-full mb-6">
          <span class="w-2 h-2 bg-blue-500 rounded-full animate-pulse"></span>
          <span class="text-sm text-blue-400 font-medium">The Apex AI Agent Platform</span>
        </div>
        <h2 class="text-5xl sm:text-7xl font-bold mb-6">
          <span class="bg-gradient-to-r from-blue-500 via-cyan-500 to-teal-400 bg-clip-text text-transparent">
            MakoClaw
          </span>
        </h2>
        <p class="text-2xl text-makoclaw-text mb-4">Your AI Agent Ecosystem</p>
        <p class="text-xl text-makoclaw-text-secondary mb-10 leading-relaxed max-w-3xl mx-auto">
          Deploy autonomous AI agents that execute tasks, manage workflows, and integrate with your favorite tools. 
          Access from Web, Telegram, Discord, Slack, WhatsApp, and more — all from one powerful control panel.
        </p>
        <div class="flex items-center justify-center gap-4 flex-wrap">
          <router-link
            to="/signup"
            class="px-10 py-4 bg-blue-500 hover:bg-blue-600 text-white text-lg rounded-xl transition-all shadow-2xl shadow-blue-500/30 font-bold"
          >
            Create Free Account
          </router-link>
          <router-link
            to="/login"
            class="px-10 py-4 bg-makoclaw-surface border-2 border-makoclaw-border hover:border-blue-500 text-makoclaw-text text-lg rounded-xl transition-all font-medium"
          >
            Sign In
          </router-link>
        </div>
      </div>

      <!-- Main Features Grid -->
      <div class="mt-24 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <!-- Multi-Agent System -->
        <div class="glass-panel rounded-2xl p-8 border border-makoclaw-border hover:border-blue-500/50 transition-all col-span-1 md:col-span-2">
          <div class="flex items-start gap-4">
            <div class="w-14 h-14 bg-blue-500/10 rounded-xl flex items-center justify-center flex-shrink-0">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-7 w-7 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </div>
            <div>
              <h3 class="text-2xl font-bold text-makoclaw-text mb-2">Multi-Agent System</h3>
              <p class="text-makoclaw-text-secondary mb-3">
                Create specialized AI agents for different tasks. Use the Orchestrator to automatically delegate tasks to the right specialist.
              </p>
              <div class="flex flex-wrap gap-2">
                <span class="px-3 py-1 bg-blue-500/10 text-blue-400 rounded-full text-xs">Orchestrator</span>
                <span class="px-3 py-1 bg-cyan-500/10 text-cyan-400 rounded-full text-xs">Specialists</span>
                <span class="px-3 py-1 bg-teal-500/10 text-teal-400 rounded-full text-xs">Auto-delegation</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Multi-Channel Support -->
        <div class="glass-panel rounded-2xl p-8 border border-makoclaw-border hover:border-blue-500/50 transition-all">
          <div class="w-12 h-12 bg-cyan-500/10 rounded-lg flex items-center justify-center mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-cyan-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <h3 class="text-xl font-bold text-makoclaw-text mb-2">9+ Channels</h3>
          <p class="text-makoclaw-text-secondary">
            Web, Telegram, Discord, Slack, WhatsApp, Signal, QQ, DingTalk, Feishu, MaixCam — one agent, everywhere.
          </p>
        </div>

        <!-- Visual Workflows -->
        <div class="glass-panel rounded-2xl p-8 border border-makoclaw-border hover:border-blue-500/50 transition-all">
          <div class="w-12 h-12 bg-blue-500/10 rounded-lg flex items-center justify-center mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <h3 class="text-xl font-bold text-makoclaw-text mb-2">Visual Workflows</h3>
          <p class="text-makoclaw-text-secondary">
            Build automation pipelines with drag-and-drop. Connect tools, prompts, and conditions visually.
          </p>
        </div>

        <!-- Task Management -->
        <div class="glass-panel rounded-2xl p-8 border border-makoclaw-border hover:border-blue-500/50 transition-all">
          <div class="w-12 h-12 bg-cyan-500/10 rounded-lg flex items-center justify-center mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-cyan-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
            </svg>
          </div>
          <h3 class="text-xl font-bold text-makoclaw-text mb-2">Kanban Task Board</h3>
          <p class="text-makoclaw-text-secondary">
            Manage tasks with a full Kanban board. Backlog, To Do, In Progress, Review, Done — AI can create and update tasks.
          </p>
        </div>

        <!-- Knowledge Base -->
        <div class="glass-panel rounded-2xl p-8 border border-makoclaw-border hover:border-blue-500/50 transition-all">
          <div class="w-12 h-12 bg-teal-500/10 rounded-lg flex items-center justify-center mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-teal-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
            </svg>
          </div>
          <h3 class="text-xl font-bold text-makoclaw-text mb-2">Knowledge Base (RAG)</h3>
          <p class="text-makoclaw-text-secondary">
            Upload documents and let AI search them with semantic retrieval. Supports PDF, TXT, MD, JSON, CSV, and more.
          </p>
        </div>

        <!-- File Management -->
        <div class="glass-panel rounded-2xl p-8 border border-makoclaw-border hover:border-blue-500/50 transition-all">
          <div class="w-12 h-12 bg-blue-500/10 rounded-lg flex items-center justify-center mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
          </div>
          <h3 class="text-xl font-bold text-makoclaw-text mb-2">File Management</h3>
          <p class="text-makoclaw-text-secondary">
            Browse, upload, and manage workspace files. Drag & drop support. AI can read, write, edit, and delete files.
          </p>
        </div>
      </div>

      <!-- All Capabilities Section -->
      <div class="mt-24">
        <h3 class="text-3xl font-bold text-center text-makoclaw-text mb-4">
          <span class="bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">The Power of MakoClaw</span>
        </h3>
        <p class="text-center text-makoclaw-text-secondary mb-12 text-lg">Everything you need to build and deploy autonomous AI systems</p>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <!-- AI Tools -->
          <div class="bg-makoclaw-surface/50 rounded-xl p-6 border border-makoclaw-border/50">
            <h4 class="text-lg font-bold text-blue-500 mb-4 flex items-center gap-2">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /></svg>
              AI Tools
            </h4>
            <ul class="space-y-2 text-sm text-makoclaw-text-secondary">
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>Multi-Agent Orchestration</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>Specialist Agents</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>Subagents</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>LLM-Assisted File Editing</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>Memory System</li>
            </ul>
          </div>

          <!-- Productivity -->
          <div class="bg-makoclaw-surface/50 rounded-xl p-6 border border-makoclaw-border/50">
            <h4 class="text-lg font-bold text-cyan-500 mb-4 flex items-center gap-2">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" /></svg>
              Productivity
            </h4>
            <ul class="space-y-2 text-sm text-makoclaw-text-secondary">
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-cyan-500 rounded-full"></span>Kanban Task Board</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-cyan-500 rounded-full"></span>Visual Workflow Builder</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-cyan-500 rounded-full"></span>Cron Jobs & Scheduling</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-cyan-500 rounded-full"></span>Task Management (AI)</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-cyan-500 rounded-full"></span>Reports & Analytics</li>
            </ul>
          </div>

          <!-- Data & Knowledge -->
          <div class="bg-makoclaw-surface/50 rounded-xl p-6 border border-makoclaw-border/50">
            <h4 class="text-lg font-bold text-teal-500 mb-4 flex items-center gap-2">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" /></svg>
              Data & Knowledge
            </h4>
            <ul class="space-y-2 text-sm text-makoclaw-text-secondary">
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-teal-500 rounded-full"></span>Knowledge Base (RAG)</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-teal-500 rounded-full"></span>File Management</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-teal-500 rounded-full"></span>Document Search</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-teal-500 rounded-full"></span>History & Sessions</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-teal-500 rounded-full"></span>Multi-user Support</li>
            </ul>
          </div>

          <!-- Integration -->
          <div class="bg-makoclaw-surface/50 rounded-xl p-6 border border-makoclaw-border/50">
            <h4 class="text-lg font-bold text-indigo-500 mb-4 flex items-center gap-2">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" /></svg>
              Integration
            </h4>
            <ul class="space-y-2 text-sm text-makoclaw-text-secondary">
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-indigo-500 rounded-full"></span>Telegram Bot</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-indigo-500 rounded-full"></span>Discord Bot</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-indigo-500 rounded-full"></span>Slack Integration</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-indigo-500 rounded-full"></span>WhatsApp (Bridge)</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-indigo-500 rounded-full"></span>MCP Protocol Support</li>
            </ul>
          </div>

          <!-- Automation -->
          <div class="bg-makoclaw-surface/50 rounded-xl p-6 border border-makoclaw-border/50">
            <h4 class="text-lg font-bold text-blue-500 mb-4 flex items-center gap-2">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
              Automation
            </h4>
            <ul class="space-y-2 text-sm text-makoclaw-text-secondary">
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>Visual Workflow Builder</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>Shell Command Execution</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>Cron Jobs</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>Custom Skills</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>Trigger Actions</li>
            </ul>
          </div>

          <!-- Communication -->
          <div class="bg-makoclaw-surface/50 rounded-xl p-6 border border-makoclaw-border/50">
            <h4 class="text-lg font-bold text-cyan-500 mb-4 flex items-center gap-2">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
              Communication
            </h4>
            <ul class="space-y-2 text-sm text-makoclaw-text-secondary">
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-cyan-500 rounded-full"></span>Web Chat Interface</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-cyan-500 rounded-full"></span>Cross-channel Messaging</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-cyan-500 rounded-full"></span>Email Support</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-cyan-500 rounded-full"></span>Voice Input (Web Speech)</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-cyan-500 rounded-full"></span>Real-time WebSocket</li>
            </ul>
          </div>

          <!-- Security -->
          <div class="bg-makoclaw-surface/50 rounded-xl p-6 border border-makoclaw-border/50">
            <h4 class="text-lg font-bold text-teal-500 mb-4 flex items-center gap-2">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>
              Security & Privacy
            </h4>
            <ul class="space-y-2 text-sm text-makoclaw-text-secondary">
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-teal-500 rounded-full"></span>Self-hosted</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-teal-500 rounded-full"></span>Multi-user Authentication</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-teal-500 rounded-full"></span>OAuth2 / PKCE</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-teal-500 rounded-full"></span>Session Management</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-teal-500 rounded-full"></span>Your Data, Your Infrastructure</li>
            </ul>
          </div>

          <!-- Technical -->
          <div class="bg-makoclaw-surface/50 rounded-xl p-6 border border-makoclaw-border/50">
            <h4 class="text-lg font-bold text-indigo-500 mb-4 flex items-center gap-2">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" /></svg>
              Technical
            </h4>
            <ul class="space-y-2 text-sm text-makoclaw-text-secondary">
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-indigo-500 rounded-full"></span>Go (Native)</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-indigo-500 rounded-full"></span>&lt; 10MB RAM</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-indigo-500 rounded-full"></span>&lt; 1s Boot</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-indigo-500 rounded-full"></span>REST API</li>
              <li class="flex items-center gap-2"><span class="w-1.5 h-1.5 bg-indigo-500 rounded-full"></span>Docker Ready</li>
            </ul>
          </div>
        </div>
      </div>

      <!-- More Channels -->
      <div class="mt-16">
        <h3 class="text-2xl font-bold text-center text-makoclaw-text mb-8">
          <span class="bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">Available Channels</span>
        </h3>
        <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
          <div class="glass-panel rounded-xl p-4 text-center border border-makoclaw-border hover:border-blue-500/50 transition-all">
            <div class="text-3xl mb-2">🌐</div>
            <div class="font-medium text-makoclaw-text">Web UI</div>
          </div>
          <div class="glass-panel rounded-xl p-4 text-center border border-makoclaw-border hover:border-blue-500/50 transition-all">
            <div class="text-3xl mb-2">✈️</div>
            <div class="font-medium text-makoclaw-text">Telegram</div>
          </div>
          <div class="glass-panel rounded-xl p-4 text-center border border-makoclaw-border hover:border-blue-500/50 transition-all">
            <div class="text-3xl mb-2">🎮</div>
            <div class="font-medium text-makoclaw-text">Discord</div>
          </div>
          <div class="glass-panel rounded-xl p-4 text-center border border-makoclaw-border hover:border-blue-500/50 transition-all">
            <div class="text-3xl mb-2">💼</div>
            <div class="font-medium text-makoclaw-text">Slack</div>
          </div>
          <div class="glass-panel rounded-xl p-4 text-center border border-makoclaw-border hover:border-blue-500/50 transition-all">
            <div class="text-3xl mb-2">📱</div>
            <div class="font-medium text-makoclaw-text">WhatsApp</div>
          </div>
          <div class="glass-panel rounded-xl p-4 text-center border border-makoclaw-border hover:border-blue-500/50 transition-all">
            <div class="text-3xl mb-2">🔐</div>
            <div class="font-medium text-makoclaw-text">Signal</div>
          </div>
          <div class="glass-panel rounded-xl p-4 text-center border border-makoclaw-border hover:border-blue-500/50 transition-all">
            <div class="text-3xl mb-2">🐧</div>
            <div class="font-medium text-makoclaw-text">QQ</div>
          </div>
          <div class="glass-panel rounded-xl p-4 text-center border border-makoclaw-border hover:border-blue-500/50 transition-all">
            <div class="text-3xl mb-2">💬</div>
            <div class="font-medium text-makoclaw-text">DingTalk</div>
          </div>
          <div class="glass-panel rounded-xl p-4 text-center border border-makoclaw-border hover:border-blue-500/50 transition-all">
            <div class="text-3xl mb-2">📋</div>
            <div class="font-medium text-makoclaw-text">Feishu</div>
          </div>
          <div class="glass-panel rounded-xl p-4 text-center border border-makoclaw-border hover:border-blue-500/50 transition-all">
            <div class="text-3xl mb-2">📷</div>
            <div class="font-medium text-makoclaw-text">MaixCam</div>
          </div>
        </div>
      </div>

      <!-- Stats -->
      <div class="mt-20 grid grid-cols-2 md:grid-cols-5 gap-8 text-center">
        <div>
          <div class="text-4xl font-bold text-blue-500 mb-2">20+</div>
          <div class="text-makoclaw-text-secondary">Built-in Tools</div>
        </div>
        <div>
          <div class="text-4xl font-bold text-cyan-500 mb-2">10+</div>
          <div class="text-makoclaw-text-secondary">Channels</div>
        </div>
        <div>
          <div class="text-4xl font-bold text-teal-500 mb-2">∞</div>
          <div class="text-makoclaw-text-secondary">Workflows</div>
        </div>
        <div>
          <div class="text-4xl font-bold text-indigo-500 mb-2">&lt;1s</div>
          <div class="text-makoclaw-text-secondary">Boot Time</div>
        </div>
        <div>
          <div class="text-4xl font-bold text-blue-500 mb-2">&lt;10MB</div>
          <div class="text-makoclaw-text-secondary">RAM Usage</div>
        </div>
      </div>

      <!-- CTA Section -->
      <div class="mt-24 text-center">
        <div class="glass-panel rounded-3xl p-12 border-2 border-blue-500/30">
          <h3 class="text-3xl font-bold text-makoclaw-text mb-4">Ready to deploy your AI shark?</h3>
          <p class="text-lg text-makoclaw-text-secondary mb-8">
            Join developers building the future of autonomous AI systems with MakoClaw.
          </p>
          <router-link
            to="/signup"
            class="inline-block px-12 py-4 bg-gradient-to-r from-blue-500 to-cyan-500 hover:from-blue-600 hover:to-cyan-600 text-white text-lg rounded-xl transition-all shadow-2xl shadow-blue-500/40 font-bold"
          >
            Start Building Now →
          </router-link>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <footer class="border-t border-makoclaw-border mt-20">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div class="text-center text-makoclaw-text-secondary text-sm">
          <p>© 2026 MakoClaw. The Apex AI Agent Platform.</p>
          <p class="mt-2">Open Source — Self-hosted — Your data, your infrastructure.</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup>
// Landing page component
</script>

