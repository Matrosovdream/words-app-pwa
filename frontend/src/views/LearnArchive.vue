<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { api } from '../lib/api.js'

const items = ref([])
const loading = ref(true)

async function load() {
  loading.value = true
  try { items.value = await api('/learn/archived?limit=200') || [] }
  finally { loading.value = false }
}

function fmt(t) {
  if (!t) return ''
  return new Date(t * 1000).toLocaleDateString()
}

onMounted(load)
</script>

<template>
  <section>
    <header class="head">
      <RouterLink to="/learn" class="back">← Learn</RouterLink>
      <h1>Archive</h1>
    </header>

    <p v-if="loading" class="muted">Loading…</p>
    <p v-else-if="!items.length" class="muted">Nothing archived yet.</p>

    <ul v-else class="list">
      <li v-for="item in items" :key="item.id" class="card">
        <RouterLink :to="`/words/${item.word_id}`" class="link">
          <h3>{{ item.lemma }}</h3>
          <span v-if="item.category_name" class="tag">{{ item.category_name }}</span>
          <span class="date">{{ fmt(item.archived_at) }}</span>
        </RouterLink>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.head { margin-bottom: 1.5rem; }
.head h1 { margin: 0.5rem 0 0; font-size: 1.8rem; }
.back { font-size: 0.9rem; color: #94a3b8; }
.back:hover { color: #38bdf8; }
.muted { color: #94a3b8; }

.list { list-style: none; padding: 0; margin: 0; display: grid; gap: 0.4rem; }
.card {
  background: rgba(30, 41, 59, 0.4);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 8px;
  padding: 0.65rem 0.95rem;
}
.link { display: flex; align-items: baseline; gap: 0.75rem; color: inherit; }
.link h3 { margin: 0; font-size: 1rem; flex: 1; }
.link:hover h3 { color: #38bdf8; }
.tag { font-size: 0.7rem; padding: 0.1rem 0.5rem; border-radius: 999px; background: rgba(56, 189, 248, 0.15); color: #38bdf8; }
.date { font-size: 0.75rem; color: #64748b; }
</style>
