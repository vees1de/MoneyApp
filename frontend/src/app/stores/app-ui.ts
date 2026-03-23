import { ref } from 'vue'
import { defineStore } from 'pinia'

import { createId } from '@/shared/lib/id'

interface ToastMessage {
  actionLabel?: string
  id: string
  onAction?: (() => void) | undefined
  tone: 'info' | 'success' | 'warning'
  text: string
}

export const useAppUiStore = defineStore('app-ui', () => {
  const isOnline = ref(typeof navigator === 'undefined' ? true : navigator.onLine)
  const toasts = ref<ToastMessage[]>([])
  const listenersBound = ref(false)
  const toastTimers = new Map<string, number>()

  function bootstrap() {
    if (typeof window === 'undefined' || listenersBound.value) {
      return
    }

    const updateStatus = () => {
      isOnline.value = window.navigator.onLine
    }

    window.addEventListener('online', updateStatus)
    window.addEventListener('offline', updateStatus)

    listenersBound.value = true
  }

  function dismissToast(toastId: string) {
    const timerId = toastTimers.get(toastId)
    if (timerId) {
      window.clearTimeout(timerId)
      toastTimers.delete(toastId)
    }

    toasts.value = toasts.value.filter((item) => item.id !== toastId)
  }

  function runToastAction(toastId: string) {
    const toast = toasts.value.find((item) => item.id === toastId)
    dismissToast(toastId)
    toast?.onAction?.()
  }

  function pushToast(
    text: string,
    tone: ToastMessage['tone'] = 'info',
    options?: {
      actionLabel?: string
      duration?: number
      onAction?: () => void
    },
  ) {
    const toast = {
      id: createId('toast'),
      actionLabel: options?.actionLabel,
      onAction: options?.onAction,
      text,
      tone,
    } satisfies ToastMessage

    toasts.value = [...toasts.value, toast]

    const timerId = window.setTimeout(() => {
      dismissToast(toast.id)
    }, options?.duration ?? 2600)

    toastTimers.set(toast.id, timerId)
  }

  function clearToasts() {
    for (const timerId of toastTimers.values()) {
      window.clearTimeout(timerId)
    }

    toastTimers.clear()
    toasts.value = []
  }

  return {
    bootstrap,
    clearToasts,
    dismissToast,
    isOnline,
    pushToast,
    runToastAction,
    toasts,
  }
})
