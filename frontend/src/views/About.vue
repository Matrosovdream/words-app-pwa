<script setup>
const stack = ['Go (net/http)', 'JSON API', 'Vue 3', 'Vue Router', 'Vite', 'Docker']
</script>

<template>
  <section class="about">
    <h1>About</h1>
    <p>
      This is a plain <strong>Vue 3 SPA</strong> talking to a
      <strong>Go JSON API</strong>. Vue Router handles navigation,
      <code>fetch()</code> pulls data, and CSS animations do the rest.
    </p>

    <h2>Stack</h2>
    <ul class="stack">
      <li
        v-for="(item, idx) in stack"
        :key="item"
        :style="{ '--d': idx * 80 + 'ms' }"
      >
        <span class="dot"></span>{{ item }}
      </li>
    </ul>

    <h2>How navigation works</h2>
    <ol class="steps">
      <li>You click a <code>&lt;RouterLink&gt;</code></li>
      <li>Vue Router swaps the matched view component</li>
      <li>The view's <code>onMounted</code> calls <code>fetch('/api/…')</code></li>
      <li>Go returns JSON, Vue updates refs, transitions play</li>
    </ol>
  </section>
</template>

<style scoped>
.about { max-width: 640px; margin: 0 auto; }
h1 { font-size: 2.2rem; margin: 0 0 1rem; }
h2 { font-size: 1.2rem; margin: 2rem 0 0.75rem; color: #38bdf8; }
p { line-height: 1.7; color: #cbd5e1; }
strong { color: #f1f5f9; }

code {
  background: rgba(148, 163, 184, 0.12);
  padding: 0.15em 0.45em;
  border-radius: 4px;
  font-size: 0.88em;
  color: #a78bfa;
}

.stack { list-style: none; padding: 0; margin: 0; }
.stack li {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 0.85rem;
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 8px;
  margin-bottom: 0.5rem;
  opacity: 0;
  transform: translateX(-20px);
  animation: slideIn 0.5s cubic-bezier(0.22, 1, 0.36, 1) forwards;
  animation-delay: var(--d);
  transition: transform 0.2s ease, border-color 0.2s ease;
}
.stack li:hover {
  transform: translateX(4px);
  border-color: rgba(56, 189, 248, 0.4);
}
.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: linear-gradient(135deg, #38bdf8, #a78bfa);
  box-shadow: 0 0 12px rgba(56, 189, 248, 0.5);
}
@keyframes slideIn {
  to { opacity: 1; transform: translateX(0); }
}

.steps { color: #cbd5e1; line-height: 1.8; padding-left: 1.25rem; }
.steps li { margin-bottom: 0.35rem; }
</style>
