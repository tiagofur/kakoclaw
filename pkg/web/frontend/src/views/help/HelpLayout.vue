<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-indigo-50/20 dark:from-makoclaw-bg dark:via-makoclaw-surface dark:to-makoclaw-bg">
    <!-- Navigation -->
    <nav class="sticky top-0 z-50 border-b bg-white/80 backdrop-blur-md border-slate-200/80 dark:bg-makoclaw-surface/80 dark:backdrop-blur-sm dark:border-makoclaw-border">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <!-- Logo -->
          <router-link to="/landing" class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white font-bold text-lg">
              🦈
            </div>
            <h1 class="text-2xl font-bold bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">
              MakoClaw
            </h1>
          </router-link>

          <!-- Center nav links -->
          <div class="hidden sm:flex items-center gap-6">
            <router-link to="/help" class="text-sm font-medium transition-colors" :class="$route.path === '/help' ? 'text-blue-500' : 'text-slate-600 hover:text-blue-500 dark:text-makoclaw-text-secondary dark:hover:text-blue-400'">
              Guides
            </router-link>
            <router-link to="/help/marketing" class="text-sm font-medium transition-colors" :class="$route.path === '/help/marketing' ? 'text-blue-500' : 'text-slate-600 hover:text-blue-500 dark:text-makoclaw-text-secondary dark:hover:text-blue-400'">
              Marketing
            </router-link>
          </div>

          <!-- Right actions -->
          <div class="flex items-center gap-2 sm:gap-4">
            <router-link to="/login" class="px-3 sm:px-4 py-2 text-slate-600 hover:text-blue-500 dark:text-makoclaw-text dark:hover:text-blue-500 transition-colors font-medium text-sm whitespace-nowrap">
              Sign In
            </router-link>
            <router-link to="/signup" class="px-3 sm:px-6 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition-all shadow-lg shadow-blue-500/20 font-medium text-sm whitespace-nowrap">
              <span class="sm:hidden">Start</span>
              <span class="hidden sm:inline">Get Started</span>
            </router-link>
          </div>
        </div>
      </div>
    </nav>

    <!-- Content -->
    <main>
      <router-view />
    </main>

    <!-- Footer -->
    <footer class="border-t border-slate-200/80 dark:border-makoclaw-border mt-20">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div class="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div class="flex items-center gap-2">
            <div class="w-6 h-6 rounded-md bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white text-xs">
              🦈
            </div>
            <span class="text-sm text-slate-500 dark:text-makoclaw-text-secondary">MakoClaw Documentation</span>
          </div>
          <div class="flex items-center gap-6">
            <router-link to="/help" class="text-sm text-slate-500 hover:text-blue-500 dark:text-makoclaw-text-secondary transition-colors">All Guides</router-link>
            <router-link to="/landing" class="text-sm text-slate-500 hover:text-blue-500 dark:text-makoclaw-text-secondary transition-colors">Home</router-link>
            <router-link to="/signup" class="text-sm text-slate-500 hover:text-blue-500 dark:text-makoclaw-text-secondary transition-colors">Get Started</router-link>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup>
</script>
