<script setup>
import { RouterLink, RouterView } from 'vue-router'

const links = [
  { to: '/', label: 'Home' },
  { to: '/words', label: 'Words' },
  { to: '/about', label: 'About' },
]
</script>

<template>
  <div class="app">
    <header class="nav">
      <RouterLink to="/" class="brand">✨ Words</RouterLink>
      <nav>
        <RouterLink
          v-for="l in links"
          :key="l.to"
          :to="l.to"
          class="nav-link"
          active-class="active"
          :exact-active-class="l.to === '/' ? 'active' : ''"
        >
          {{ l.label }}
          <span class="underline"></span>
        </RouterLink>
      </nav>
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
  max-width: 820px;
  margin: 0 auto;
  padding: 2rem 1.5rem 4rem;
}

.nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  margin-bottom: 2.5rem;
  background: rgba(30, 41, 59, 0.6);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 14px;
}
.brand {
  font-weight: 700;
  font-size: 1.1rem;
  color: #38bdf8;
  transition: transform 0.2s ease;
}
.brand:hover { transform: scale(1.05); }

.nav nav { display: flex; gap: 0.5rem; }
.nav-link {
  position: relative;
  padding: 0.5rem 0.9rem;
  font-size: 0.95rem;
  color: #cbd5e1;
  border-radius: 8px;
  transition: color 0.2s ease, background 0.2s ease;
}
.nav-link:hover { color: #f1f5f9; background: rgba(148, 163, 184, 0.1); }
.nav-link .underline {
  position: absolute;
  left: 0.9rem;
  right: 0.9rem;
  bottom: 0.35rem;
  height: 2px;
  background: linear-gradient(90deg, #38bdf8, #a78bfa);
  border-radius: 2px;
  transform: scaleX(0);
  transform-origin: left;
  transition: transform 0.3s cubic-bezier(0.65, 0, 0.35, 1);
}
.nav-link.active { color: #f1f5f9; }
.nav-link.active .underline { transform: scaleX(1); }

/* Page transitions */
.page-enter-active,
.page-leave-active {
  transition: opacity 0.35s ease, transform 0.45s cubic-bezier(0.22, 1, 0.36, 1), filter 0.35s ease;
}
.page-enter-from {
  opacity: 0;
  transform: translateY(16px) scale(0.98);
  filter: blur(6px);
}
.page-leave-to {
  opacity: 0;
  transform: translateY(-12px) scale(0.98);
  filter: blur(6px);
}
</style>
