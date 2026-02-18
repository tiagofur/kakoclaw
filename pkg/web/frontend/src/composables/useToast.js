import { reactive } from 'vue'

const toasts = reactive([])
let nextId = 0

export function useToast() {
  const add = (message, type = 'info', duration = 4000) => {
    const id = ++nextId
    toasts.push({ id, message, type, visible: true })

    if (duration > 0) {
      setTimeout(() => dismiss(id), duration)
    }
    return id
  }

  const dismiss = (id) => {
    const idx = toasts.findIndex(t => t.id === id)
    if (idx !== -1) {
      toasts[idx].visible = false
      setTimeout(() => {
        const removeIdx = toasts.findIndex(t => t.id === id)
        if (removeIdx !== -1) toasts.splice(removeIdx, 1)
      }, 300) // Allow exit animation
    }
  }

  const success = (message, duration) => add(message, 'success', duration)
  const error = (message, duration) => add(message, 'error', duration ?? 6000)
  const warning = (message, duration) => add(message, 'warning', duration)
  const info = (message, duration) => add(message, 'info', duration)

  return { toasts, add, dismiss, success, error, warning, info }
}
