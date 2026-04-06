<script setup>
import { ref, onMounted, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { auth } from './stores/auth.js'
import { api } from './lib/api.js'

const links = [
  { to: '/review', label: 'Review' },
  { to: '/learn', label: 'Learn' },
  { to: '/admin/sites', label: 'Sites' },
  { to: '/settings', label: 'Settings' },
]
const reviewCount = ref(0)
const route = useRoute()
const router = useRouter()

async function refreshCount() {
  if (!auth.isAuthenticated) { reviewCount.value = 0; return }
  try { reviewCount.value = await api('/review/count') || 0 } catch { /* ignore */ }
}

function logout() {
  auth.logout()
  router.push('/login')
}

onMounted(refreshCount)
watch(() => route.fullPath, refreshCount)
</script>

<template>
  <div class="app">
    <header v-if="auth.isAuthenticated" class="nav">
      <RouterLink to="/review" class="brand">✨ Words</RouterLink>
      <nav>
        <RouterLink
          v-for="l in links"
          :key="l.to"
          :to="l.to"
          class="nav-link"
          active-class="active"
        >
          {{ l.label }}
          <span v-if="l.to === '/review' && reviewCount > 0" class="badge">{{ reviewCount }}</span>
        </RouterLink>
      </nav>
      <button class="logout" @click="logout" :title="auth.user?.email">Sign out</button>
    </header>

    <RouterView v-slot="{ Component, route }">
      <Transition name="page" mode="out-in">
        <component :is="Component" :key="route.fullPath" />
      </Transition>
    </RouterView>
  </div>
</template>

<style>
* { box-sizing: border-box; }
html, body { margin: 0; padding: 0; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  background: radial-gradient(ellipse at top, #1e293b 0%, #0f172a 60%);
  background-attachment: fixed;
  color: #e2e8f0;
  min-height: 100vh;
}
a { color: inherit; text-decoration: none; }

.app {
  max-width: 920px;
  margin: 0 auto;
  padding:
    calc(1.5rem + env(safe-area-inset-top))
    calc(1.25rem + env(safe-area-inset-right))
    calc(4rem + env(safe-area-inset-bottom))
    calc(1.25rem + env(safe-area-inset-left));
}

.nav {
  display: flex; align-items: center; gap: 1rem;
  padding: 0.85rem 1.25rem;
  margin-bottom: 2rem;
  background: rgba(30, 41, 59, 0.6);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 14px;
}
.brand { font-weight: 700; font-size: 1rem; color: #38bdf8; }
.nav nav { display: flex; gap: 0.35rem; flex: 1; flex-wrap: wrap; }
.nav-link {
  position: relative;
  padding: 0.45rem 0.8rem; font-size: 0.9rem;
  color: #cbd5e1; border-radius: 8px;
  transition: color 0.2s, background 0.2s;
}
.nav-link:hover { color: #f1f5f9; background: rgba(148, 163, 184, 0.1); }
.nav-link.active { color: #f1f5f9; background: rgba(56, 189, 248, 0.15); }
.badge {
  display: inline-block; margin-left: 0.35rem;
  padding: 0.05rem 0.4rem; font-size: 0.7rem;
  background: #38bdf8; color: #0f172a; border-radius: 999px; font-weight: 600;
}
.logout {
  padding: 0.45rem 0.8rem; border-radius: 8px; font-size: 0.85rem;
  background: rgba(148, 163, 184, 0.08); border: 1px solid rgba(148, 163, 184, 0.2);
  color: #cbd5e1; cursor: pointer;
}
.logout:hover { color: #f87171; border-color: rgba(248, 113, 113, 0.4); }

.page-enter-active, .page-leave-active { transition: opacity 0.25s ease, transform 0.3s cubic-bezier(0.22, 1, 0.36, 1); }
.page-enter-from { opacity: 0; transform: translateY(10px); }
.page-leave-to { opacity: 0; transform: translateY(-6px); }
</style>
