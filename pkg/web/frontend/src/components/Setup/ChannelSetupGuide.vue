<template>
  <div class="channel-setup-guide space-y-6">
    <!-- Platform-specific instructions -->
    <div class="space-y-4">
      <div v-if="channel === 'telegram'" class="space-y-4">
        <div class="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
          <h4 class="text-white font-bold mb-3 flex items-center">
            <span class="text-2xl mr-2">📱</span> Telegram Setup
          </h4>
          <ol class="space-y-3">
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">1.</span>
              <div>
                <p class="text-white font-medium">Open Telegram and find @BotFather</p>
                <p class="text-sm text-slate-400 mt-1">
                  <a href="https://t.me/botfather" target="_blank" rel="noopener" class="text-blue-400 hover:text-blue-300">
                    t.me/botfather →
                  </a>
                </p>
              </div>
            </li>
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">2.</span>
              <div>
                <p class="text-white font-medium">Send /newbot and follow the prompts</p>
                <p class="text-sm text-slate-400 mt-1">
                  Give your bot a name (e.g., "Mymakoclaw") and username
                </p>
              </div>
            </li>
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">3.</span>
              <div>
                <p class="text-white font-medium">Copy the HTTP API token</p>
                <p class="text-sm text-slate-400 mt-1">
                  Format: <code class="bg-slate-800 px-2 py-1 rounded">123456789:ABCdefGHIJKlmnOPQRstuvWXYZ</code>
                </p>
              </div>
            </li>
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">4.</span>
              <div>
                <p class="text-white font-medium">Get your chat ID (optional for group use)</p>
                <p class="text-sm text-slate-400 mt-1">
                  Open Telegram and message @userinfobot or @getidsbot
                </p>
              </div>
            </li>
          </ol>
        </div>

        <div class="bg-amber-900/20 border border-amber-500/30 rounded-lg p-4 text-sm text-amber-300">
          <p class="font-bold mb-2">💡 Pro tips:</p>
          <ul class="space-y-1">
            <li>• Bot commands: Send /setcommands to @BotFather to set command menu</li>
            <li>• Privacy: Send /setprivacy to @BotFather to allow group access</li>
            <li>• Inline: Enable inline mode in /setinline for richer responses</li>
          </ul>
        </div>
      </div>

      <div v-else-if="channel === 'discord'" class="space-y-4">
        <div class="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
          <h4 class="text-white font-bold mb-3 flex items-center">
            <span class="text-2xl mr-2">🎮</span> Discord Setup
          </h4>
          <ol class="space-y-3">
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">1.</span>
              <div>
                <p class="text-white font-medium">Go to Discord Developer Portal</p>
                <p class="text-sm text-slate-400 mt-1">
                  <a href="https://discord.com/developers/applications" target="_blank" rel="noopener" class="text-blue-400 hover:text-blue-300">
                    discord.com/developers →
                  </a>
                </p>
              </div>
            </li>
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">2.</span>
              <div>
                <p class="text-white font-medium">Click "New Application"</p>
                <p class="text-sm text-slate-400 mt-1">
                  Give it a name and click Create
                </p>
              </div>
            </li>
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">3.</span>
              <div>
                <p class="text-white font-medium">Navigate to "Bot" and click "Add Bot"</p>
                <p class="text-sm text-slate-400 mt-1">
                  This creates your bot account
                </p>
              </div>
            </li>
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">4.</span>
              <div>
                <p class="text-white font-medium">Copy the bot token</p>
                <p class="text-sm text-slate-400 mt-1">
                  Format: <code class="bg-slate-800 px-2 py-1 rounded">NzkyNzg3OTk1Njc0N...</code>
                </p>
              </div>
            </li>
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">5.</span>
              <div>
                <p class="text-white font-medium">Set bot permissions</p>
                <p class="text-sm text-slate-400 mt-1">
                  OAuth2 → URL Generator: Select "bot" scope and permissions (Send Messages, Read Messages, etc.)
                </p>
              </div>
            </li>
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">6.</span>
              <div>
                <p class="text-white font-medium">Invite bot to your server</p>
                <p class="text-sm text-slate-400 mt-1">
                  Use the generated OAuth2 URL from step 5
                </p>
              </div>
            </li>
          </ol>
        </div>

        <div class="bg-amber-900/20 border border-amber-500/30 rounded-lg p-4 text-sm text-amber-300">
          <p class="font-bold mb-2">💡 Pro tips:</p>
          <ul class="space-y-1">
            <li>• Permissions: Give bot "Send Messages", "Read Message History", "Embed Links"</li>
            <li>• Intents: Enable "Message Content Intent" in bot settings (required for message reading)</li>
            <li>• Channels: Assign bot to specific channels in server settings</li>
          </ul>
        </div>
      </div>

      <div v-else-if="channel === 'slack'" class="space-y-4">
        <div class="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
          <h4 class="text-white font-bold mb-3 flex items-center">
            <span class="text-2xl mr-2">💼</span> Slack Setup
          </h4>
          <ol class="space-y-3">
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">1.</span>
              <div>
                <p class="text-white font-medium">Go to Slack API</p>
                <p class="text-sm text-slate-400 mt-1">
                  <a href="https://api.slack.com/apps" target="_blank" rel="noopener" class="text-blue-400 hover:text-blue-300">
                    api.slack.com/apps →
                  </a>
                </p>
              </div>
            </li>
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">2.</span>
              <div>
                <p class="text-white font-medium">Click "Create New App"</p>
                <p class="text-sm text-slate-400 mt-1">
                  Select "From scratch" and choose your workspace
                </p>
              </div>
            </li>
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">3.</span>
              <div>
                <p class="text-white font-medium">Install app to workspace</p>
                <p class="text-sm text-slate-400 mt-1">
                  Navigate to "Install App" and click "Install to Workspace"
                </p>
              </div>
            </li>
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">4.</span>
              <div>
                <p class="text-white font-medium">Copy bot token</p>
                <p class="text-sm text-slate-400 mt-1">
                  Find "Bot User OAuth Token" starting with <code class="bg-slate-800 px-2 py-1 rounded">xoxb-</code>
                </p>
              </div>
            </li>
            <li class="flex gap-3">
              <span class="text-blue-400 font-bold min-w-fit">5.</span>
              <div>
                <p class="text-white font-medium">Set bot scopes</p>
                <p class="text-sm text-slate-400 mt-1">
                  OAuth & Permissions: Add scopes like chat:write, app_mentions:read, etc.
                </p>
              </div>
            </li>
          </ol>
        </div>

        <div class="bg-amber-900/20 border border-amber-500/30 rounded-lg p-4 text-sm text-amber-300">
          <p class="font-bold mb-2">💡 Pro tips:</p>
          <ul class="space-y-1">
            <li>• Scopes: chat:write, app_mentions:read, reactions:read, files:read</li>
            <li>• Socket Mode: Enable for webhooks and real-time events</li>
            <li>• Slash commands: Add /makoclaw command in Event Subscriptions</li>
          </ul>
        </div>
      </div>
    </div>

    <!-- Common security tips -->
    <div class="bg-blue-900/20 border border-blue-500/30 rounded-lg p-4">
      <p class="text-blue-300 font-bold mb-2 flex items-center">
        <span class="mr-2">🔐</span> Security Best Practices
      </p>
      <ul class="text-blue-300/80 text-sm space-y-1">
        <li>✓ Never share your bot token publicly</li>
        <li>✓ Regenerate tokens if accidentally exposed</li>
        <li>✓ Use separate bots for development and production</li>
        <li>✓ Regularly rotate API credentials</li>
        <li>✓ makoclaw never stores tokens in plain text</li>
      </ul>
    </div>
  </div>
</template>

<script setup>
defineProps({
  channel: {
    type: String,
    default: ''
  }
})
</script>

<style scoped>
code {
  font-family: 'Courier New', monospace;
  font-size: 0.85em;
}
</style>

