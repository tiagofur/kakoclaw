import { useAuthStore } from '../stores/authStore'

class ChatWebSocket {
  constructor() {
    this.ws = null
    this.url = this.getWebSocketURL('/ws/chat')
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
    this.reconnectDelay = 1000
    this.listeners = {}
  }

  getWebSocketURL(path) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${window.location.host}${path}`
  }

  connect() {
    return new Promise((resolve, reject) => {
      try {
        const authStore = useAuthStore()
        const token = authStore.token

        // Append token to URL for handshake authorization
        const urlWithToken = token ? `${this.url}?token=${encodeURIComponent(token)}` : this.url
        this.ws = new WebSocket(urlWithToken)

        // Send auth token after connection
        this.ws.onopen = () => {
          this.reconnectAttempts = 0
          this.emit('connected')
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data)
            this.emit('message', message)
          } catch (e) {
            console.error('Failed to parse WebSocket message:', e)
          }
        }

        this.ws.onerror = (error) => {
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

  attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)
      setTimeout(() => this.connect(), delay)
    }
  }

  send(data) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    } else {
      console.warn('WebSocket not connected')
    }
  }

  on(event, callback) {
    if (!this.listeners[event]) {
      this.listeners[event] = []
    }
    this.listeners[event].push(callback)
  }

  off(event, callback) {
    if (this.listeners[event]) {
      this.listeners[event] = this.listeners[event].filter(cb => cb !== callback)
    }
  }

  emit(event, data) {
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

  isConnected() {
    return this.ws && this.ws.readyState === WebSocket.OPEN
  }
}

class TaskWebSocket {
  constructor() {
    this.ws = null
    this.url = this.getWebSocketURL('/ws/tasks')
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
    this.reconnectDelay = 1000
    this.listeners = {}
  }

  getWebSocketURL(path) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${window.location.host}${path}`
  }

  connect() {
    return new Promise((resolve, reject) => {
      try {
        const authStore = useAuthStore()
        const token = authStore.token

        // Append token to URL for handshake authorization
        const urlWithToken = token ? `${this.url}?token=${encodeURIComponent(token)}` : this.url
        this.ws = new WebSocket(urlWithToken)

        this.ws.onopen = () => {
          this.reconnectAttempts = 0
          this.emit('connected')
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data)
            this.emit('update', message)
          } catch (e) {
            console.error('Failed to parse WebSocket message:', e)
          }
        }

        this.ws.onerror = (error) => {
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

  attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)
      setTimeout(() => this.connect(), delay)
    }
  }

  send(data) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    } else {
      console.warn('WebSocket not connected')
    }
  }

  on(event, callback) {
    if (!this.listeners[event]) {
      this.listeners[event] = []
    }
    this.listeners[event].push(callback)
  }

  off(event, callback) {
    if (this.listeners[event]) {
      this.listeners[event] = this.listeners[event].filter(cb => cb !== callback)
    }
  }

  emit(event, data) {
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

  isConnected() {
    return this.ws && this.ws.readyState === WebSocket.OPEN
  }
}

export { ChatWebSocket, TaskWebSocket }
