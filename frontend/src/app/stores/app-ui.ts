import { ref } from 'vue'
import { defineStore } from 'pinia'

import { createId } from '@/shared/lib/id'

interface ToastMessage {
  id: string
  tone: 'info' | 'success' | 'warning'
  text: string
}

export const useAppUiStore = defineStore('app-ui', () => {
  const isOnline = ref(typeof navigator === 'undefined' ? true : navigator.onLine)
  const toasts = ref<ToastMessage[]>([])
  const listenersBound = ref(false)

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

  function pushToast(text: string, tone: ToastMessage['tone'] = 'info') {
    const toast = {
      id: createId('toast'),
      text,
      tone,
    } satisfies ToastMessage

    toasts.value = [...toasts.value, toast]

    window.setTimeout(() => {
      toasts.value = toasts.value.filter((item) => item.id !== toast.id)
    }, 2600)
  }

  function clearToasts() {
    toasts.value = []
  }

  return {
    bootstrap,
    clearToasts,
    isOnline,
    pushToast,
    toasts,
  }
})
