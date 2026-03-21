import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import { bootstrapApp } from './app/boot/bootstrap'
import { createAppRouter } from './app/router'
import './shared/styles/main.css'

async function mountApp() {
  const app = createApp(App)
  const pinia = createPinia()

  app.use(pinia)
  await bootstrapApp(pinia)

  const router = createAppRouter(pinia)
  app.use(router)
  await router.isReady()

  app.mount('#app')
}

void mountApp()
