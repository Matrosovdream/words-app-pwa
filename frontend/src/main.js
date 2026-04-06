import { createApp } from 'vue'
import App from './App.vue'
import router from './router.js'
import { registerSW } from 'virtual:pwa-register'

// Auto-update SW: re-registers every hour and refreshes on new version.
registerSW({ immediate: true })

createApp(App).use(router).mount('#app')
