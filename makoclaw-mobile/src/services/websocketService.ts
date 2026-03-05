import { useAuthStore } from '../stores/authStore'
import { useConfigStore } from '../stores/configStore'

type EventCallback = (data?: any) => void

class ChatWebSocket {
    private ws: WebSocket | null = null
    private reconnectAttempts = 0
    private maxReconnectAttempts = 5
    private reconnectDelay = 1000
    private listeners: Record<string, EventCallback[]> = {}

    private getWebSocketURL(path: string): string {
        const configStore = useConfigStore()
        const serverUrl = configStore.serverUrl
        // Convert http(s) to ws(s)
        const wsUrl = serverUrl.replace(/^http/, 'ws')
        return `${wsUrl}${path}`
    }

    connect(): Promise<void> {
        return new Promise((resolve, reject) => {
            try {
                const authStore = useAuthStore()
                const token = authStore.token
                const url = this.getWebSocketURL('/ws/chat')

                const urlWithToken = token ? `${url}?token=${encodeURIComponent(token)}` : url
                this.ws = new WebSocket(urlWithToken)

                this.ws.onopen = () => {
                    this.reconnectAttempts = 0
                    this.emit('connected')
                    resolve()
                }

                this.ws.onmessage = (event: MessageEvent) => {
                    try {
                        const message = JSON.parse(event.data)
                        this.emit('message', message)
                    } catch (e) {
                        console.error('Failed to parse WebSocket message:', e)
                    }
                }

                this.ws.onerror = (error: Event) => {
                    console.error('WebSocket error:', error)
                    this.emit('error', error)
                    reject(error)
                }

                this.ws.onclose = () => {
                    this.emit('disconnected')
                    this.attemptReconnect()
                }
            } catch (error) {
                reject(error)
            }
        })
    }

    private attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++
            const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)
            setTimeout(() => this.connect(), delay)
        }
    }

    send(data: Record<string, any>) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(data))
        } else {
            console.warn('WebSocket not connected')
        }
    }

    on(event: string, callback: EventCallback) {
        if (!this.listeners[event]) {
            this.listeners[event] = []
        }
        this.listeners[event].push(callback)
    }

    off(event: string, callback: EventCallback) {
        if (this.listeners[event]) {
            this.listeners[event] = this.listeners[event].filter(cb => cb !== callback)
        }
    }

    private emit(event: string, data?: any) {
        if (this.listeners[event]) {
            this.listeners[event].forEach(cb => cb(data))
        }
    }

    disconnect() {
        if (this.ws) {
            this.ws.close()
            this.ws = null
        }
    }

    isConnected(): boolean {
        return this.ws !== null && this.ws.readyState === WebSocket.OPEN
    }
}

// Singleton
let chatWebSocketInstance: ChatWebSocket | null = null

export function getChatWebSocket(): ChatWebSocket {
    if (!chatWebSocketInstance) {
        chatWebSocketInstance = new ChatWebSocket()
    }
    return chatWebSocketInstance
}

export { ChatWebSocket }
