<script setup>
import { ref, onMounted, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { api } from '../lib/api.js'

const categories = ref([])
const items = ref([])
const selectedCategory = ref('')
const sort = ref('newest')
const loading = ref(false)
const editingCat = ref(null)

async function loadCategories() {
  categories.value = await api('/learn/categories') || []
}

async function loadItems() {
  loading.value = true
  try {
    const q = new URLSearchParams({ sort: sort.value })
    if (selectedCategory.value) q.set('category_id', selectedCategory.value)
    items.value = await api(`/learn/items?${q}`) || []
  } finally { loading.value = false }
}

async function done(item) {
  await api(`/learn/items/${item.id}/done`, { method: 'POST' })
  items.value = items.value.filter(i => i.id !== item.id)
  await loadCategories() // counts changed
}

function newCategory() {
  editingCat.value = { name: '', color: '#38bdf8', sort_order: 0 }
}
function editCategory(c) {
  editingCat.value = { id: c.id, name: c.name, color: c.color, sort_order: c.sort_order }
}
async function saveCategory() {
  const c = editingCat.value
  if (c.id) {
    await api(`/learn/categories/${c.id}`, { method: 'PUT', body: c })
  } else {
    await api('/learn/categories', { method: 'POST', body: c })
  }
  editingCat.value = null
  await loadCategories()
}
async function removeCategory(c) {
  if (!confirm(`Delete category "${c.name}"? Words will remain uncategorized.`)) return
  await api(`/learn/categories/${c.id}`, { method: 'DELETE' })
  if (selectedCategory.value === c.id) selectedCategory.value = ''
  await loadCategories()
  await loadItems()
}

watch([selectedCategory, sort], loadItems)
onMounted(async () => { await loadCategories(); await loadItems() })
</script>

<template>
  <section>
    <header class="head">
      <h1>Learn</h1>
      <RouterLink to="/learn/archive" class="btn small">Archive</RouterLink>
    </header>

    <div class="filters">
      <div class="cats">
        <button
          class="cat-btn"
          :class="{ active: selectedCategory === '' }"
          @click="selectedCategory = ''"
        >All</button>
        <button
          v-for="c in categories" :key="c.id"
          class="cat-btn"
          :class="{ active: selectedCategory === c.id }"
          :style="c.color ? `--c: ${c.color}` : ''"
          @click="selectedCategory = c.id"
        >
          {{ c.name }} <span class="count">{{ c.item_count }}</span>
        </button>
        <button class="cat-btn add" @click="newCategory">+</button>
      </div>

      <div class="sort">
        <label>Sort:</label>
        <select v-model="sort">
          <option value="newest">Newest</option>
          <option value="oldest">Oldest</option>
          <option value="mastery">Mastery</option>
        </select>
      </div>
    </div>

    <div v-if="selectedCategory" class="cat-bar">
      <button class="btn xs" @click="editCategory(categories.find(c => c.id === selectedCategory))">Edit</button>
      <button class="btn xs danger" @click="removeCategory(categories.find(c => c.id === selectedCategory))">Delete</button>
    </div>

    <p v-if="loading" class="muted">Loading…</p>
    <p v-else-if="!items.length" class="muted">No words here yet.</p>

    <ul v-else class="list">
      <li v-for="item in items" :key="item.id" class="card">
        <RouterLink :to="`/words/${item.word_id}`" class="link">
          <div class="word">
            <h3>{{ item.lemma }}</h3>
            <span
              v-for="cat in item.categories" :key="cat.id"
              class="tag"
              :style="cat.color ? `background: ${cat.color}22; color: ${cat.color}` : ''"
            >{{ cat.name }}</span>
          </div>
          <span class="mastery">★ {{ item.mastery_level }}/5</span>
        </RouterLink>
        <button class="btn done" @click="done(item)" title="Mark as done (archive)">Done</button>
      </li>
    </ul>

    <!-- Category modal -->
    <div v-if="editingCat" class="modal" @click.self="editingCat = null">
      <form class="modal-card" @submit.prevent="saveCategory">
        <h2>{{ editingCat.id ? 'Edit category' : 'New category' }}</h2>
        <label>Name<input v-model="editingCat.name" required /></label>
        <label>Color<input type="color" v-model="editingCat.color" /></label>
        <label>Sort order<input type="number" v-model.number="editingCat.sort_order" /></label>
        <div class="actions">
          <button type="button" class="btn" @click="editingCat = null">Cancel</button>
          <button type="submit" class="btn primary">Save</button>
        </div>
      </form>
    </div>
  </section>
</template>

<style scoped>
.head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.25rem; }
.head h1 { margin: 0; font-size: 1.8rem; }
.muted { color: #94a3b8; }

.filters { display: flex; justify-content: space-between; align-items: flex-start; gap: 1rem; margin-bottom: 1rem; flex-wrap: wrap; }
.cats { display: flex; flex-wrap: wrap; gap: 0.4rem; flex: 1; }
.cat-btn {
  padding: 0.35rem 0.8rem; border-radius: 999px;
  background: rgba(148, 163, 184, 0.1); border: 1px solid rgba(148, 163, 184, 0.2);
  color: #cbd5e1; font-size: 0.85rem; cursor: pointer;
  --c: #38bdf8;
}
.cat-btn:hover { border-color: var(--c); color: #f1f5f9; }
.cat-btn.active { background: var(--c); color: #0f172a; border-color: var(--c); font-weight: 600; }
.cat-btn.add { padding: 0.35rem 0.7rem; }
.count { opacity: 0.7; font-size: 0.75rem; margin-left: 0.25rem; }

.sort { display: flex; align-items: center; gap: 0.5rem; font-size: 0.85rem; color: #94a3b8; }
.sort select {
  padding: 0.3rem 0.55rem; border-radius: 6px; border: 1px solid rgba(148, 163, 184, 0.2);
  background: rgba(15, 23, 42, 0.6); color: #f1f5f9; font-size: 0.85rem;
}

.cat-bar { display: flex; gap: 0.4rem; margin-bottom: 1rem; }

.list { list-style: none; padding: 0; margin: 0; display: grid; gap: 0.5rem; }
.card {
  background: rgba(30, 41, 59, 0.55);
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 10px;
  padding: 0.75rem 1rem;
  display: flex; align-items: center; gap: 0.75rem;
}
.link { flex: 1; display: flex; align-items: center; justify-content: space-between; color: inherit; }
.link:hover h3 { color: #38bdf8; }
.word { display: flex; align-items: baseline; gap: 0.75rem; }
.word h3 { margin: 0; font-size: 1.1rem; transition: color 0.2s; }
.tag { font-size: 0.7rem; padding: 0.1rem 0.5rem; border-radius: 999px; background: rgba(56, 189, 248, 0.2); color: #38bdf8; }
.mastery { font-size: 0.8rem; color: #94a3b8; }

.btn {
  padding: 0.5rem 0.9rem; border-radius: 8px; border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(148, 163, 184, 0.08); color: #e2e8f0;
  font-weight: 500; font-size: 0.85rem; cursor: pointer;
}
.btn.small { padding: 0.4rem 0.8rem; font-size: 0.8rem; }
.btn.xs { padding: 0.25rem 0.55rem; font-size: 0.75rem; }
.btn.primary { background: linear-gradient(135deg, #38bdf8, #a78bfa); color: #0f172a; border: 0; }
.btn.danger { color: #f87171; border-color: rgba(248, 113, 113, 0.4); }
.btn.done {
  background: rgba(34, 197, 94, 0.15);
  border-color: rgba(34, 197, 94, 0.4);
  color: #4ade80;
}
.btn.done:hover { background: rgba(34, 197, 94, 0.25); }

.modal { position: fixed; inset: 0; background: rgba(15, 23, 42, 0.8); display: flex; align-items: center; justify-content: center; z-index: 100; padding: 1rem; }
.modal-card { background: #1e293b; border-radius: 14px; padding: 1.75rem; width: 100%; max-width: 400px; }
.modal-card h2 { margin: 0 0 1rem; font-size: 1.2rem; }
.modal-card label { display: block; margin-bottom: 0.75rem; font-size: 0.85rem; color: #cbd5e1; }
.modal-card input { width: 100%; padding: 0.5rem 0.7rem; margin-top: 0.3rem; border-radius: 6px; border: 1px solid rgba(148, 163, 184, 0.2); background: rgba(15, 23, 42, 0.6); color: #f1f5f9; font-size: 0.9rem; }
.modal-card input[type=color] { height: 2.25rem; padding: 0.2rem; }
.modal-card .actions { display: flex; justify-content: flex-end; gap: 0.5rem; margin-top: 1rem; }
</style>
