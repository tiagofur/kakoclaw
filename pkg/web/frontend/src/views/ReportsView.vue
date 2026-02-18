<template>
  <div class="h-full flex flex-col max-w-4xl mx-auto w-full p-4 md:p-8">
    <div class="flex-none mb-8">
      <h2 class="text-3xl font-bold bg-gradient-to-r from-picoclaw-accent to-purple-500 bg-clip-text text-transparent mb-2">Reports</h2>
      <p class="text-picoclaw-text-secondary">Generate and send reports via email to Google Docs/Sheets integrations.</p>
    </div>

    <div class="bg-picoclaw-surface border border-picoclaw-border rounded-xl shadow-sm p-6 space-y-6">
        <!-- Report Form -->
        <div class="space-y-4">
            <div>
                <label class="block text-sm font-medium mb-1">To (Email)</label>
                <input 
                    v-model="to" 
                    type="email" 
                    placeholder="Defaults to configured recipient"
                    class="w-full bg-picoclaw-bg border border-picoclaw-border rounded-lg px-4 py-2 outline-none focus:border-picoclaw-accent transition-colors"
                />
            </div>
            
            <div>
                <label class="block text-sm font-medium mb-1">Subject</label>
                <input 
                    v-model="subject" 
                    type="text" 
                    placeholder="Weekly Summary / Project Update"
                    class="w-full bg-picoclaw-bg border border-picoclaw-border rounded-lg px-4 py-2 outline-none focus:border-picoclaw-accent transition-colors"
                />
            </div>

            <div>
                <label class="block text-sm font-medium mb-1">Content (Markdown)</label>
                <textarea 
                    v-model="body" 
                    rows="10"
                    placeholder="Write your report here..."
                    class="w-full bg-picoclaw-bg border border-picoclaw-border rounded-lg px-4 py-2 outline-none focus:border-picoclaw-accent transition-colors resize-none font-mono text-sm"
                ></textarea>
            </div>
        </div>

        <div class="flex justify-end pt-4 border-t border-picoclaw-border">
            <button 
                @click="sendReport" 
                :disabled="sending || !subject || !body"
                class="flex items-center gap-2 px-6 py-2.5 bg-picoclaw-accent text-white rounded-lg hover:bg-picoclaw-accent/90 transition-all font-medium disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-picoclaw-accent/20"
            >
                <div v-if="sending" class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                <span v-else>Send Report</span>
                <svg v-if="!sending" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" /></svg>
            </button>
        </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useChatStore } from '../stores/chatStore'
import { ChatWebSocket } from '../services/websocketService'
import { useRouter } from 'vue-router'
import { useToast } from '../composables/useToast'

const chatStore = useChatStore()
const router = useRouter()
const toast = useToast()

const to = ref('')
const subject = ref('')
const body = ref('')
const sending = ref(false)

const sendReport = async () => {
    if (!subject.value || !body.value) return

    sending.value = true
    
    // We send a command to the agent to use the email tool
    const prompt = `Please send an email report using the 'send_email_report' tool.
Subject: ${subject.value}
To: ${to.value}
Body:
${body.value}`

    try {
        // Try using existing chatStore WebSocket first
        if (chatStore.sendMessage(prompt)) {
            toast.success('Report sent to agent')
            router.push('/chat')
            return
        }

        // If chatStore WS not connected, create a temporary one
        const tempWs = new ChatWebSocket()
        await tempWs.connect()
        tempWs.send({
            type: 'message',
            content: prompt,
            session_id: 'web:chat:report:' + Date.now().toString(36)
        })
        tempWs.disconnect()
        toast.success('Report sent to agent')
        router.push('/chat')
    } catch (err) {
        console.error("Failed to send report command:", err)
        toast.error('Failed to send command to agent. Make sure the chat is connected.')
    } finally {
        sending.value = false
    }
}
</script>
