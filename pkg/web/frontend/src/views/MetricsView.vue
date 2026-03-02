<template>
  <div class="h-full flex flex-col bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-sky-500/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-blue-500/20 via-transparent to-transparent" />
    </div>

    <!-- Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
        <div class="flex items-center gap-3">
          <!-- Icon Container -->
          <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-sky-500/20 to-blue-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-sky-500/10">
            <svg
              class="w-5 h-5 sm:w-6 sm:h-6 text-sky-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M16 8v8m-4-5v5m-4-2v2m-2 4h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
          </div>

          <div class="flex-1 min-w-0">
            <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-sky-400 bg-clip-text text-transparent">
              Metrics
            </h1>
            <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5 hidden sm:block">
              In-process observability for LLM and tool usage
            </p>
          </div>

          <!-- Refresh Button -->
          <button
            :disabled="loading"
            class="px-4 sm:px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-sky-500 to-blue-500 hover:from-sky-600 hover:to-blue-600 text-white rounded-xl transition-all shadow-lg shadow-sky-500/25 hover:shadow-sky-500/40 text-sm font-bold flex items-center gap-2 active:scale-95 flex-shrink-0 disabled:opacity-50"
            @click="loadMetrics"
          >
            <div
              v-if="loading"
              class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
            />
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
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            <span>Refresh</span>
          </button>
        </div>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto p-4 md:p-6 space-y-8 relative z-10 custom-scrollbar">
      <!-- Info Header -->
      <div class="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <p class="text-makoclaw-text-secondary opacity-80 max-w-2xl">
            In-process observability for LLM calls, tool executions, and agent runs.
          </p>
        </div>
        
        <!-- Uptime Indicator -->
        <div
          v-if="metrics"
          class="flex items-center gap-3 px-4 py-2 bg-makoclaw-bg/30 border border-makoclaw-border/50 rounded-full backdrop-blur-md"
        >
          <div class="flex items-center gap-2">
            <span class="relative flex h-2 w-2">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75" />
              <span class="relative inline-flex rounded-full h-2 w-2 bg-green-500" />
            </span>
            <span class="text-[10px] uppercase font-bold tracking-widest text-makoclaw-text-secondary">
              Uptime
            </span>
          </div>
          <div class="h-4 w-px bg-makoclaw-border/50" />
          <span class="text-sm font-mono text-makoclaw-text">
            {{ formatUptime(metrics.uptime_seconds) }}
          </span>
          <div class="h-4 w-px bg-makoclaw-border/50" />
          <span class="text-xs text-makoclaw-text-secondary opacity-60">
            Started {{ formatDate(metrics.started_at) }}
          </span>
        </div>
      </div>

      <!-- Error state -->
      <div
        v-if="error"
        class="bg-red-500/10 border border-red-500/30 rounded-xl p-4 text-red-400"
      >
        {{ error }}
      </div>

      <!-- Summary Cards -->
      <div
        v-if="metrics"
        class="grid grid-cols-1 md:grid-cols-3 gap-6"
      >
        <!-- LLM Card -->
        <div class="glass-panel p-6 rounded-2xl relative overflow-hidden group/card hover:-translate-y-1 transition-all duration-300 shadow-xl shadow-makoclaw-accent/5">
          <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover/card:w-full transition-all duration-500 opacity-50" />
          
          <div class="flex items-center gap-4 mb-6">
            <div class="p-3 bg-makoclaw-accent/10 rounded-xl text-makoclaw-accent relative">
              <div class="absolute inset-0 bg-makoclaw-accent/20 blur-lg rounded-full opacity-0 group-hover/card:opacity-100 transition-opacity duration-500" />
              <svg 
                class="w-6 h-6 relative z-10" 
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                />
              </svg>
            </div>
            <h3 class="text-lg font-bold text-makoclaw-text tracking-tight">
              LLM Calls
            </h3>
          </div>
          <div class="space-y-3">
            <div class="flex justify-between items-center group/item p-2 hover:bg-makoclaw-accent/5 rounded-lg transition-colors">
              <span class="text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider opacity-60">
                Total calls
              </span>
              <span class="font-mono text-base font-bold text-makoclaw-text">
                {{ metrics.llm_calls }}
              </span>
            </div>
            <div class="flex justify-between items-center group/item p-2 hover:bg-makoclaw-accent/5 rounded-lg transition-colors">
              <span class="text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider opacity-60">
                Errors
              </span>
              <span
                class="font-mono text-base font-bold"
                :class="metrics.llm_errors > 0 ? 'text-makoclaw-error' : 'text-makoclaw-text'"
              >
                {{ metrics.llm_errors }}
              </span>
            </div>
            <div class="flex justify-between items-center group/item p-2 hover:bg-makoclaw-accent/5 rounded-lg transition-colors">
              <span class="text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider opacity-60">
                Avg latency
              </span>
              <span class="font-mono text-base font-bold text-makoclaw-accent">
                {{ formatMs(metrics.llm_avg_ms) }}
              </span>
            </div>
            <div class="pt-2 mt-2 border-t border-makoclaw-border/30 grid grid-cols-2 gap-2">
              <div class="p-2 bg-makoclaw-bg/20 rounded-lg">
                <p class="text-[10px] font-bold text-makoclaw-text-secondary uppercase opacity-50 mb-0.5">
                  Tokens In
                </p>
                <p class="font-mono text-sm font-bold text-makoclaw-text">
                  {{ formatNumber(metrics.llm_tokens_in) }}
                </p>
              </div>
              <div class="p-2 bg-makoclaw-bg/20 rounded-lg">
                <p class="text-[10px] font-bold text-makoclaw-text-secondary uppercase opacity-50 mb-0.5">
                  Tokens Out
                </p>
                <p class="font-mono text-sm font-bold text-makoclaw-text">
                  {{ formatNumber(metrics.llm_tokens_out) }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Tool Card -->
        <div class="glass-panel p-6 rounded-2xl relative overflow-hidden group/card hover:-translate-y-1 transition-all duration-300 shadow-xl shadow-makoclaw-accent/5">
          <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-green-500 to-emerald-400 group-hover/card:w-full transition-all duration-500 opacity-50" />
          
          <div class="flex items-center gap-4 mb-6">
            <div class="p-3 bg-green-500/10 rounded-xl text-green-500 relative">
              <div class="absolute inset-0 bg-green-500/20 blur-lg rounded-full opacity-0 group-hover/card:opacity-100 transition-opacity duration-500" />
              <svg 
                class="w-6 h-6 relative z-10" 
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
            <h3 class="text-lg font-bold text-makoclaw-text tracking-tight">
              Tool Calls
            </h3>
          </div>
          <div class="space-y-3">
            <div class="flex justify-between items-center group/item p-2 hover:bg-green-500/5 rounded-lg transition-colors">
              <span class="text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider opacity-60">
                Total calls
              </span>
              <span class="font-mono text-base font-bold text-makoclaw-text">
                {{ metrics.tool_calls }}
              </span>
            </div>
            <div class="flex justify-between items-center group/item p-2 hover:bg-green-500/5 rounded-lg transition-colors">
              <span class="text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider opacity-60">
                Errors
              </span>
              <span
                class="font-mono text-base font-bold"
                :class="metrics.tool_errors > 0 ? 'text-makoclaw-error' : 'text-makoclaw-text'"
              >
                {{ metrics.tool_errors }}
              </span>
            </div>
            <div class="flex justify-between items-center group/item p-2 hover:bg-green-500/5 rounded-lg transition-colors">
              <span class="text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider opacity-60">
                Avg latency
              </span>
              <span class="font-mono text-base font-bold text-green-500">
                {{ formatMs(metrics.tool_avg_ms) }}
              </span>
            </div>
            <div class="p-3 bg-makoclaw-bg/20 rounded-xl border border-makoclaw-border/30 mt-4">
              <p class="text-[10px] font-bold text-makoclaw-text-secondary uppercase opacity-50 mb-1">
                Cumulative Execution Time
              </p>
              <p class="font-mono text-lg font-bold text-makoclaw-text">
                {{ formatMs(metrics.tool_total_ms) }}
              </p>
            </div>
          </div>
        </div>

        <!-- Agent Card -->
        <div class="glass-panel p-6 rounded-2xl relative overflow-hidden group/card hover:-translate-y-1 transition-all duration-300 shadow-xl shadow-makoclaw-accent/5">
          <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-blue-400 to-indigo-500 group-hover/card:w-full transition-all duration-500 opacity-50" />
          
          <div class="flex items-center gap-4 mb-6">
            <div class="p-3 bg-blue-500/10 rounded-xl text-blue-400 relative">
              <div class="absolute inset-0 bg-blue-500/20 blur-lg rounded-full opacity-0 group-hover/card:opacity-100 transition-opacity duration-500" />
              <svg 
                class="w-6 h-6 relative z-10" 
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
            <h3 class="text-lg font-bold text-makoclaw-text tracking-tight">
              Agent Runs
            </h3>
          </div>
          <div class="space-y-3">
            <div class="flex justify-between items-center group/item p-2 hover:bg-blue-500/5 rounded-lg transition-colors">
              <span class="text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider opacity-60">
                Total runs
              </span>
              <span class="font-mono text-base font-bold text-makoclaw-text">
                {{ metrics.agent_runs }}
              </span>
            </div>
            <div class="flex justify-between items-center group/item p-2 hover:bg-blue-500/5 rounded-lg transition-colors">
              <span class="text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider opacity-60">
                Avg latency
              </span>
              <span class="font-mono text-base font-bold text-blue-400">
                {{ formatMs(metrics.agent_avg_ms) }}
              </span>
            </div>
            <div class="flex justify-between items-center group/item p-2 hover:bg-blue-500/5 rounded-lg transition-colors">
              <span class="text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider opacity-60">
                Avg iterations
              </span>
              <span class="font-mono text-base font-bold text-makoclaw-text">
                {{ metrics.agent_avg_iterations?.toFixed(1) || '0' }}
              </span>
            </div>
            <div class="p-3 bg-makoclaw-bg/20 rounded-xl border border-makoclaw-border/30 mt-4">
              <div class="flex justify-between items-center">
                <div>
                  <p class="text-[10px] font-bold text-makoclaw-text-secondary uppercase opacity-50 mb-1">
                    Total Iterations
                  </p>
                  <p class="font-mono text-lg font-bold text-makoclaw-text">
                    {{ metrics.agent_iterations_total }}
                  </p>
                </div>
                <div class="text-right">
                  <p class="text-[10px] font-bold text-makoclaw-text-secondary uppercase opacity-50 mb-1">
                    Errors
                  </p>
                  <p
                    class="font-mono text-lg font-bold"
                    :class="metrics.agent_errors > 0 ? 'text-makoclaw-error' : 'text-makoclaw-text'"
                  >
                    {{ metrics.agent_errors }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>


      <!-- Per-model breakdown -->
      <div v-if="metrics && Object.keys(metrics.llm_by_model || {}).length > 0">
        <h3 class="text-xl font-bold text-makoclaw-text mb-4 px-1">
          LLM by Model
        </h3>
        <div class="glass-panel rounded-2xl border border-makoclaw-border/50 overflow-hidden shadow-xl shadow-makoclaw-accent/5">
          <table class="w-full text-sm">
            <thead class="bg-makoclaw-bg/50 text-makoclaw-text-secondary">
              <tr class="border-b border-makoclaw-border/50">
                <th class="text-left px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Model
                </th>
                <th class="text-right px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Calls
                </th>
                <th class="text-right px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Errors
                </th>
                <th class="text-right px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Avg ms
                </th>
                <th class="text-right px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Tokens In
                </th>
                <th class="text-right px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Tokens Out
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(m, name) in metrics.llm_by_model"
                :key="name"
                class="border-b border-makoclaw-border/30 hover:bg-makoclaw-accent/5 transition-colors group/row"
              >
                <td class="px-6 py-4 font-mono text-[11px] text-makoclaw-accent font-bold">
                  {{ name }}
                </td>

                <td class="px-6 py-4 text-right font-mono text-makoclaw-text font-bold">
                  {{ m.calls }}
                </td>
                <td
                  class="px-6 py-4 text-right font-mono font-bold"
                  :class="m.errors > 0 ? 'text-makoclaw-error' : 'text-makoclaw-text'"
                >
                  {{ m.errors }}
                </td>

                <td class="px-6 py-4 text-right font-mono text-makoclaw-accent font-bold">
                  {{ formatMs(m.avg_ms) }}
                </td>
                <td class="px-6 py-4 text-right font-mono text-makoclaw-text">
                  {{ formatNumber(m.tokens_in) }}
                </td>
                <td class="px-6 py-4 text-right font-mono text-makoclaw-text font-medium">
                  {{ formatNumber(m.tokens_out) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>


      <!-- Per-tool breakdown -->
      <div
        v-if="metrics && Object.keys(metrics.tool_by_name || {}).length > 0"
        class="mb-8"
      >
        <h3 class="text-xl font-bold text-makoclaw-text mb-4 px-1">
          Tools by Name
        </h3>
        <div class="glass-panel rounded-2xl border border-makoclaw-border/50 overflow-hidden shadow-xl shadow-makoclaw-accent/5">
          <table class="w-full text-sm">
            <thead class="bg-makoclaw-bg/50 text-makoclaw-text-secondary">
              <tr class="border-b border-makoclaw-border/50">
                <th class="text-left px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Tool
                </th>
                <th class="text-right px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Calls
                </th>
                <th class="text-right px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Errors
                </th>
                <th class="text-right px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Avg ms
                </th>
                <th class="text-right px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Total ms
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(t, name) in metrics.tool_by_name"
                :key="name"
                class="border-b border-makoclaw-border/30 hover:bg-makoclaw-accent/5 transition-colors group/row"
              >
                <td class="px-6 py-4 font-mono text-[11px] text-green-500 font-bold">
                  {{ name }}
                </td>
                <td class="px-6 py-4 text-right font-mono text-makoclaw-text font-bold">
                  {{ t.calls }}
                </td>
                <td
                  class="px-6 py-4 text-right font-mono font-bold"
                  :class="t.errors > 0 ? 'text-makoclaw-error' : 'text-makoclaw-text'"
                >
                  {{ t.errors }}
                </td>
                <td class="px-6 py-4 text-right font-mono text-green-500 font-bold">
                  {{ formatMs(t.avg_ms) }}
                </td>
                <td class="px-6 py-4 text-right font-mono text-makoclaw-text font-medium">
                  {{ formatMs(t.total_ms) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Recent Events -->
      <div
        v-if="metrics && (metrics.recent_events || []).length > 0"
        class="mb-8"
      >
        <h3 class="text-xl font-bold text-makoclaw-text mb-4 px-1">
          Recent Events (last {{ metrics.recent_events.length }})
        </h3>
        <div class="glass-panel rounded-2xl border border-makoclaw-border/50 overflow-hidden max-h-96 overflow-y-auto shadow-xl shadow-makoclaw-accent/5">
          <table class="w-full text-sm">
            <thead class="sticky top-0 bg-makoclaw-bg/50 text-makoclaw-text-secondary">
              <tr class="border-b border-makoclaw-border/50">
                <th class="text-left px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Time
                </th>
                <th class="text-left px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Type
                </th>
                <th class="text-left px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Detail
                </th>
                <th class="text-right px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Duration
                </th>
                <th class="text-left px-6 py-4 text-xs font-bold uppercase tracking-wider">
                  Error
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(evt, i) in [...metrics.recent_events].reverse()"
                :key="i"
                class="border-b border-makoclaw-border/30 hover:bg-makoclaw-accent/5 transition-colors group/row"
              >
                <td class="px-6 py-4 text-xs text-makoclaw-text-secondary whitespace-nowrap font-mono">
                  {{ formatTime(evt.timestamp) }}
                </td>
                <td class="px-6 py-4">
                  <span
                    class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium"
                    :class="{
                      'bg-blue-500/10 text-blue-400': evt.type === 'llm_call',
                      'bg-green-500/10 text-green-400': evt.type === 'tool_call',
                      'bg-blue-500/10 text-blue-400': evt.type === 'agent_run',
                      'bg-red-500/10 text-red-400': evt.type === 'error'
                    }"
                  >
                    {{ evt.type }}
                  </span>
                </td>
                <td class="px-6 py-4 font-mono text-xs text-makoclaw-text">
                  {{ evt.model || evt.tool || '-' }}
                </td>
                <td class="px-6 py-4 text-right font-mono text-xs text-makoclaw-accent font-bold">
                  {{ formatMs(evt.duration_ms) }}
                </td>
                <td class="px-6 py-4 text-xs text-makoclaw-error max-w-xs truncate font-mono">
                  {{ evt.error || '' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Empty state -->
      <div
        v-if="metrics && metrics.llm_calls === 0 && metrics.tool_calls === 0 && metrics.agent_runs === 0"
        class="glass-panel rounded-3xl p-16 text-center relative overflow-hidden group/empty"
      >
        <div class="absolute inset-0 bg-gradient-to-b from-makoclaw-accent/5 to-transparent opacity-0 group-hover/empty:opacity-100 transition-opacity duration-700" />
        
        <div class="relative w-24 h-24 mx-auto mb-8">
          <div class="absolute inset-0 bg-makoclaw-accent/20 blur-2xl rounded-full animate-pulse" />
          <div class="relative glass-panel w-full h-full rounded-2xl flex items-center justify-center text-makoclaw-accent shadow-xl shadow-makoclaw-accent/10">
            <svg
              class="w-10 h-10"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M16 8v8m-4-5v5m-4-2v2m-2 4h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
          </div>
        </div>
        
        <h3 class="text-xl font-bold text-makoclaw-text mb-2">
          No Metrics Yet
        </h3>
        <p class="text-makoclaw-text-secondary opacity-70 max-w-sm mx-auto">
          Start a chat or assign tasks to see LLM and tool usage data appearing here in real-time.
        </p>
      </div>

      <!-- Loading state -->
      <div
        v-if="!metrics && loading"
        class="flex items-center justify-center py-20"
      >
        <div class="w-12 h-12 border-2 border-makoclaw-accent/20 border-t-makoclaw-accent rounded-full animate-spin" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import advancedService from '../services/advancedService'

const metrics = ref(null)
const loading = ref(false)
const error = ref('')

async function loadMetrics() {
  loading.value = true
  error.value = ''
  try {
    metrics.value = await advancedService.fetchMetrics()
  } catch (e) {
    error.value = 'Failed to load metrics: ' + (e.response?.data?.error || e.message)
  } finally {
    loading.value = false
  }
}

function formatMs(ms) {
  if (!ms && ms !== 0) return '-'
  if (ms < 1000) return ms + 'ms'
  return (ms / 1000).toFixed(1) + 's'
}

function formatNumber(n) {
  if (!n && n !== 0) return '0'
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return String(n)
}

function formatUptime(seconds) {
  if (!seconds) return '0s'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  if (h > 0) return `${h}h ${m}m`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

function formatDate(iso) {
  if (!iso) return '-'
  return new Date(iso).toLocaleString()
}

function formatTime(iso) {
  if (!iso) return '-'
  return new Date(iso).toLocaleTimeString()
}

let refreshInterval = null

onMounted(() => {
  loadMetrics()
  refreshInterval = setInterval(loadMetrics, 30000)
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
})
</script>

<style scoped>
.glass-sticky {
  background: rgba(var(--makoclaw-surface-rgb), 0.7);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}
</style>
