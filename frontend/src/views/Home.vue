<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'

const message = ref('')
const time = ref('')

onMounted(async () => {
  const res = await fetch('/api/hello')
  const { data } = await res.json()
  message.value = data.message
  time.value = data.time
})
</script>

<template>
  <section class="hero">
    <div class="glow"></div>
    <h1>
      <span class="word-anim" style="--d:0ms">Welcome</span>
      <span class="word-anim" style="--d:80ms">to</span>
      <span class="word-anim word-accent" style="--d:160ms">Words</span>
    </h1>
    <p class="tagline word-anim" style="--d:300ms">{{ message || '\u00A0' }}</p>
    <p class="time word-anim" style="--d:380ms">Server time: {{ time || '…' }}</p>

    <div class="cta word-anim" style="--d:480ms">
      <RouterLink to="/words" class="btn primary">Explore words →</RouterLink>
      <RouterLink to="/about" class="btn ghost">About this app</RouterLink>
    </div>
  </section>
</template>

<style scoped>
.hero {
  position: relative;
  padding: 4rem 2rem;
  text-align: center;
  overflow: hidden;
}
.glow {
  position: absolute;
  inset: -20% 20% auto 20%;
  height: 400px;
  background: radial-gradient(circle, rgba(56, 189, 248, 0.25), transparent 60%);
  filter: blur(40px);
  z-index: -1;
  animation: float 6s ease-in-out infinite;
}
@keyframes float {
  0%, 100% { transform: translateY(0) scale(1); }
  50%      { transform: translateY(-20px) scale(1.05); }
}

h1 {
  font-size: clamp(2.5rem, 6vw, 4rem);
  margin: 0 0 1rem;
  font-weight: 800;
  letter-spacing: -0.02em;
}
.word-accent {
  background: linear-gradient(90deg, #38bdf8, #a78bfa, #f472b6);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.tagline { font-size: 1.2rem; color: #cbd5e1; margin: 0.5rem 0; min-height: 1.6em; }
.time { font-size: 0.85rem; color: #64748b; margin: 0 0 2rem; }

.word-anim {
  display: inline-block;
  opacity: 0;
  transform: translateY(20px);
  animation: rise 0.6s cubic-bezier(0.22, 1, 0.36, 1) forwards;
  animation-delay: var(--d, 0ms);
  margin-right: 0.3em;
}
@keyframes rise {
  to { opacity: 1; transform: translateY(0); }
}

.cta { display: flex; gap: 0.75rem; justify-content: center; flex-wrap: wrap; }
.btn {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  border-radius: 10px;
  font-weight: 600;
  font-size: 0.95rem;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}
.btn.primary {
  background: linear-gradient(135deg, #38bdf8, #a78bfa);
  color: #0f172a;
  box-shadow: 0 4px 20px rgba(56, 189, 248, 0.3);
}
.btn.primary:hover {
  transform: translateY(-2px) scale(1.02);
  box-shadow: 0 8px 28px rgba(56, 189, 248, 0.45);
}
.btn.ghost {
  background: rgba(148, 163, 184, 0.08);
  color: #e2e8f0;
  border: 1px solid rgba(148, 163, 184, 0.2);
}
.btn.ghost:hover { background: rgba(148, 163, 184, 0.15); transform: translateY(-2px); }
</style>
