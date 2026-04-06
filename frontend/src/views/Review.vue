<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../lib/api.js'

const items = ref([])
const categories = ref([])
const loading = ref(true)
const pickedCategory = ref({})

async function load() {
  loading.value = true
  try {
    const [it, cats] = await Promise.all([
      api('/review?limit=100'),
      api('/learn/categories'),
    ])
    items.value = it || []
    categories.value = cats || []
  } finally { loading.value = false }
}

async function add(item) {
  const catId = pickedCategory.value[item.id] || ''
  await api(`/review/${item.id}/add`, {
    method: 'POST',
    body: catId ? { learn_category_id: catId } : {},
  })
  items.value = items.value.filter(i => i.id !== item.id)
}

async function deny(item) {
  await api(`/review/${item.id}/deny`, { method: 'POST' })
  items.value = items.value.filter(i => i.id !== item.id)
}

onMounted(load)
</script>

<template>
  <section>
    <header class="head">
      <h1>Review</h1>
      <p class="muted">{{ items.length }} pending</p>
    </header>

    <p v-if="loading" class="muted">Loading…</p>
    <p v-else-if="!items.length" class="muted">No pending words. Parser will fill this as pages come in.</p>

    <ul class="list" v-else>
      <li v-for="item in items" :key="item.id" class="card">
        <div class="top">
          <h3>{{ item.lemma }}</h3>
          <span class="lang">{{ item.language }}</span>
          <span v-if="item.site_category_name" class="source-cat">{{ item.site_category_name }}</span>
        </div>
        <p v-if="item.sample_sentence" class="sample">"{{ item.sample_sentence }}"</p>
        <p v-if="item.source_url" class="source">
          <a :href="item.source_url" target="_blank" rel="noopener">source ↗</a>
        </p>
        <div class="actions">
          <select v-model="pickedCategory[item.id]">
            <option value="">— no category —</option>
            <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
          <button class="btn primary" @click="add(item)">✓ Add</button>
          <button class="btn danger" @click="deny(item)">✕ Deny</button>
        </div>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.head { display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 1.5rem; }
.head h1 { margin: 0; font-size: 1.8rem; }
.muted { color: #94a3b8; }

.list { list-style: none; padding: 0; margin: 0; display: grid; gap: 0.75rem; }
.card {
  background: rgba(30, 41, 59, 0.55);
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 12px;
  padding: 1rem 1.25rem;
}
.top { display: flex; align-items: baseline; gap: 0.75rem; }
.top h3 { margin: 0; font-size: 1.4rem; color: #f1f5f9; }
.lang { font-size: 0.75rem; color: #64748b; text-transform: uppercase; }
.source-cat { font-size: 0.7rem; padding: 0.1rem 0.5rem; border-radius: 999px; background: rgba(167, 139, 250, 0.2); color: #a78bfa; }
.sample { margin: 0.5rem 0 0.25rem; color: #cbd5e1; font-style: italic; }
.source { margin: 0.25rem 0 0; font-size: 0.8rem; }
.source a { color: #38bdf8; }

.actions { display: flex; gap: 0.5rem; margin-top: 0.75rem; flex-wrap: wrap; }
.actions select {
  padding: 0.5rem 0.7rem; border-radius: 6px; border: 1px solid rgba(148, 163, 184, 0.2);
  background: rgba(15, 23, 42, 0.6); color: #f1f5f9; font-size: 0.85rem;
}
.btn {
  padding: 0.5rem 1rem; border-radius: 8px; border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(148, 163, 184, 0.08); color: #e2e8f0;
  font-weight: 500; font-size: 0.9rem; cursor: pointer;
}
.btn.primary { background: linear-gradient(135deg, #38bdf8, #a78bfa); color: #0f172a; border: 0; }
.btn.danger { color: #f87171; border-color: rgba(248, 113, 113, 0.4); }
</style>
