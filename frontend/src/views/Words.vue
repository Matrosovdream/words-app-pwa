<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'

const words = ref([])
const loading = ref(true)

onMounted(async () => {
  const res = await fetch('/api/words')
  const { data } = await res.json()
  words.value = data
  loading.value = false
})
</script>

<template>
  <section>
    <header class="head">
      <h1>Word Collection</h1>
      <p>Click any word to see its detail.</p>
    </header>

    <p v-if="loading" class="loading">Loading…</p>

    <ul v-else class="grid">
      <li
        v-for="(w, idx) in words"
        :key="w.slug"
        class="card"
        :style="{ '--d': idx * 70 + 'ms' }"
      >
        <RouterLink :to="`/words/${w.slug}`" class="card-link">
          <span class="emoji">{{ w.emoji }}</span>
          <div class="body">
            <h3>{{ w.word }}</h3>
            <p>{{ w.definition }}</p>
          </div>
          <span class="arrow">→</span>
        </RouterLink>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.head { text-align: center; margin-bottom: 2rem; }
.head h1 { font-size: 2.2rem; margin: 0 0 0.5rem; }
.head p { color: #94a3b8; margin: 0; }
.loading { text-align: center; color: #64748b; }

.grid {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 0.9rem;
}

.card {
  opacity: 0;
  transform: translateY(24px);
  animation: cardIn 0.55s cubic-bezier(0.22, 1, 0.36, 1) forwards;
  animation-delay: var(--d);
}
@keyframes cardIn {
  to { opacity: 1; transform: translateY(0); }
}

.card-link {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
  background: rgba(30, 41, 59, 0.55);
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 14px;
  transition:
    transform 0.25s cubic-bezier(0.22, 1, 0.36, 1),
    border-color 0.25s ease,
    background 0.25s ease,
    box-shadow 0.25s ease;
}
.card-link:hover {
  transform: translateY(-3px);
  border-color: rgba(56, 189, 248, 0.4);
  background: rgba(30, 41, 59, 0.85);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.35);
}
.card-link:hover .arrow { transform: translateX(6px); color: #38bdf8; }
.card-link:hover .emoji { transform: scale(1.15) rotate(-8deg); }

.emoji {
  font-size: 2rem;
  transition: transform 0.3s cubic-bezier(0.22, 1, 0.36, 1);
}
.body { flex: 1; min-width: 0; }
.body h3 { margin: 0 0 0.25rem; font-size: 1.1rem; color: #f1f5f9; }
.body p {
  margin: 0;
  color: #94a3b8;
  font-size: 0.9rem;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
}
.arrow {
  font-size: 1.3rem;
  color: #64748b;
  transition: transform 0.25s ease, color 0.25s ease;
}
</style>
