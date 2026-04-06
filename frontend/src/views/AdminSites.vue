<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { api } from '../lib/api.js'

const sites = ref([])
const loading = ref(true)
const editing = ref(null)
const expanded = ref(new Set())
const enqueueUrl = ref({})
const categoriesOf = ref({})
const activeJobs = ref({}) // siteId → count
const learnCategories = ref([])

const blank = () => ({
  name: '', base_url: '', is_active: true,
  daily_hit_limit: 100, hit_delay_ms: 2000,
  reparse_enabled: false, reparse_after_days: 30,
  user_agent: '',
})

async function load() {
  loading.value = true
  try {
    const [s, lc] = await Promise.all([
      api('/admin/sites'),
      api('/learn/categories'),
    ])
    sites.value = s || []
    learnCategories.value = lc || []
  } finally { loading.value = false }
}

async function pollProgress() {
  if (!expanded.value.size) return
  try {
    const stats = await api('/admin/parser/stats')
    if (!stats) return
    // build per-site active job count from stats.sites
    const map = {}
    for (const s of (stats.sites || [])) {
      // stats gives hits_today; for running+pending we use global counts but break by site
      map[s.id] = s.hits_today // rough proxy — used only as "something is running" indicator
    }
    // use running count directly
    for (const s of sites.value) {
      if (expanded.value.has(s.id)) {
        const n = await api(`/admin/sites/${s.id}/active-jobs`).catch(() => null)
        if (n !== null) activeJobs.value[s.id] = n
      }
    }
  } catch (_) {}
}

function startCreate() { editing.value = blank() }
function startEdit(s) { editing.value = { ...s } }
function cancelEdit() { editing.value = null }

async function save() {
  const payload = { ...editing.value }
  if (payload.id) {
    await api(`/admin/sites/${payload.id}`, { method: 'PUT', body: payload })
  } else {
    await api('/admin/sites', { method: 'POST', body: payload })
  }
  editing.value = null
  await load()
}

async function remove(s) {
  if (!confirm(`Delete site "${s.name}"?`)) return
  await api(`/admin/sites/${s.id}`, { method: 'DELETE' })
  await load()
}

async function toggleExpand(s) {
  if (expanded.value.has(s.id)) {
    expanded.value.delete(s.id)
  } else {
    expanded.value.add(s.id)
    categoriesOf.value[s.id] = await api(`/admin/sites/${s.id}/categories`) || []
  }
  expanded.value = new Set(expanded.value)
}

async function enqueue(s) {
  const url = enqueueUrl.value[s.id]
  if (!url) return
  await api(`/admin/sites/${s.id}/enqueue`, { method: 'POST', body: { url } })
  enqueueUrl.value[s.id] = ''
  alert('Queued for parsing')
}

async function runAll(s) {
  await api(`/admin/sites/${s.id}/run`, { method: 'POST' })
  alert(`Queued all categories for "${s.name}". Parser will pick them up shortly.`)
}

// category management (inline)
const editingCat = ref(null)
function blankCat(siteId) {
  return {
    site_id: siteId, name: '', start_url: '',
    url_pattern: '', selector_title: 'title',
    selector_body: 'article, main', source_language: 'en',
    is_active: true, learn_category_id: '',
  }
}
function startCreateCat(siteId) { editingCat.value = blankCat(siteId) }
function startEditCat(cat) { editingCat.value = { ...cat, learn_category_id: cat.learn_category_id || '' } }
function cancelCat() { editingCat.value = null }
async function saveCat() {
  const c = { ...editingCat.value }
  if (!c.learn_category_id) delete c.learn_category_id
  if (c.id) {
    await api(`/admin/sites/${c.site_id}/categories/${c.id}`, { method: 'PUT', body: c })
  } else {
    await api(`/admin/sites/${c.site_id}/categories`, { method: 'POST', body: c })
  }
  const siteId = editingCat.value.site_id
  editingCat.value = null
  categoriesOf.value[siteId] = await api(`/admin/sites/${siteId}/categories`) || []
}
async function removeCat(cat) {
  if (!confirm(`Delete category "${cat.name}"?`)) return
  await api(`/admin/sites/${cat.site_id}/categories/${cat.id}`, { method: 'DELETE' })
  categoriesOf.value[cat.site_id] = await api(`/admin/sites/${cat.site_id}/categories`) || []
}

async function crawlCat(cat) {
  await api(`/admin/sites/${cat.site_id}/categories/${cat.id}/crawl`, { method: 'POST' })
}

function learnCatName(id) {
  return learnCategories.value.find(c => c.id === id)?.name || ''
}

// poll stats every 5s while any site is expanded
let pollTimer = null
onMounted(async () => {
  await load()
  pollTimer = setInterval(async () => {
    if (!expanded.value.size) return
    try {
      const stats = await api('/admin/parser/stats')
      if (!stats) return
      const next = {}
      for (const row of (stats.sites || [])) {
        next[row.id] = row.hits_today
      }
      // store running+pending from global counts as a badge on each expanded site
      activeJobs.value = { _pending: stats.jobs?.pending || 0, _running: stats.jobs?.running || 0 }
    } catch (_) {}
  }, 5000)
})
onUnmounted(() => clearInterval(pollTimer))
</script>

<template>
  <section>
    <header class="head">
      <h1>Sites</h1>
      <button class="btn primary" @click="startCreate">+ New site</button>
    </header>

    <p v-if="loading" class="loading">Loading…</p>

    <!-- global running indicator -->
    <div v-if="(activeJobs._running || 0) > 0 || (activeJobs._pending || 0) > 0" class="progress-bar">
      <span class="pulse"></span>
      {{ activeJobs._running || 0 }} running · {{ activeJobs._pending || 0 }} pending jobs
    </div>

    <ul class="list" v-else-if="!loading">
      <li v-for="s in sites" :key="s.id" class="card">
        <div class="row">
          <div class="grow">
            <h3>
              {{ s.name }}
              <span v-if="!s.is_active" class="pill muted">paused</span>
              <span v-if="s.reparse_enabled" class="pill">reparse</span>
            </h3>
            <p class="url">{{ s.base_url }}</p>
            <p class="meta">
              {{ s.daily_hit_limit }} hits/day · {{ s.hit_delay_ms }}ms delay
            </p>
          </div>
          <div class="actions">
            <button class="btn small primary" @click="runAll(s)" title="Run all categories">▶ Run all</button>
            <button class="btn small" @click="toggleExpand(s)">
              {{ expanded.has(s.id) ? '▲' : '▼' }} categories
            </button>
            <button class="btn small" @click="startEdit(s)">Edit</button>
            <button class="btn small danger" @click="remove(s)">Delete</button>
          </div>
        </div>

        <div v-if="expanded.has(s.id)" class="detail">
          <div class="enqueue">
            <input v-model="enqueueUrl[s.id]" placeholder="https://example.com/article" />
            <button class="btn small primary" @click="enqueue(s)">Queue URL</button>
          </div>

          <h4>
            Categories
            <button class="btn small" @click="startCreateCat(s.id)">+ Add</button>
          </h4>
          <ul class="cats">
            <li v-for="c in categoriesOf[s.id]" :key="c.id">
              <div class="cat-info">
                <strong>{{ c.name }}</strong>
                <span class="muted"> — {{ c.start_url }}</span>
                <span v-if="c.learn_category_id" class="tag">{{ learnCatName(c.learn_category_id) }}</span>
              </div>
              <div class="cat-actions">
                <button class="btn xs primary" @click="crawlCat(c)" title="Re-enqueue start URL">Crawl</button>
                <button class="btn xs" @click="startEditCat(c)">Edit</button>
                <button class="btn xs danger" @click="removeCat(c)">×</button>
              </div>
            </li>
            <li v-if="!(categoriesOf[s.id] && categoriesOf[s.id].length)" class="muted">
              No categories yet. Add one with a Start URL to begin crawling.
            </li>
          </ul>
        </div>
      </li>
      <li v-if="!sites.length" class="muted">No sites yet.</li>
    </ul>

    <!-- Site edit modal -->
    <div v-if="editing" class="modal" @click.self="cancelEdit">
      <form class="modal-card" @submit.prevent="save">
        <h2>{{ editing.id ? 'Edit site' : 'New site' }}</h2>
        <label>Name<input v-model="editing.name" required /></label>
        <label>Base URL<input v-model="editing.base_url" type="url" required /></label>
        <div class="grid2">
          <label>Daily hit limit<input type="number" v-model.number="editing.daily_hit_limit" /></label>
          <label>Delay (ms)<input type="number" v-model.number="editing.hit_delay_ms" /></label>
        </div>
        <div class="grid2">
          <label class="check"><input type="checkbox" v-model="editing.is_active" /> Active</label>
          <label class="check"><input type="checkbox" v-model="editing.reparse_enabled" /> Re-parse old pages</label>
        </div>
        <label>Re-parse after (days)<input type="number" v-model.number="editing.reparse_after_days" /></label>
        <label>User agent<input v-model="editing.user_agent" placeholder="optional" /></label>
        <div class="actions">
          <button type="button" class="btn" @click="cancelEdit">Cancel</button>
          <button type="submit" class="btn primary">Save</button>
        </div>
      </form>
    </div>

    <!-- Category edit modal -->
    <div v-if="editingCat" class="modal" @click.self="cancelCat">
      <form class="modal-card" @submit.prevent="saveCat">
        <h2>{{ editingCat.id ? 'Edit category' : 'New category' }}</h2>
        <label>Name<input v-model="editingCat.name" required /></label>
        <label>Start URL<input v-model="editingCat.start_url" type="url" required /></label>
        <label>URL pattern (regex, links to follow)<input v-model="editingCat.url_pattern" placeholder="e.g. /business/" /></label>
        <div class="grid2">
          <label>Title selector<input v-model="editingCat.selector_title" /></label>
          <label>Body selector<input v-model="editingCat.selector_body" required /></label>
        </div>
        <div class="grid2">
          <label>Language<input v-model="editingCat.source_language" placeholder="en" /></label>
          <label class="check"><input type="checkbox" v-model="editingCat.is_active" /> Active</label>
        </div>
        <label>
          Learn category (auto-assign words)
          <select v-model="editingCat.learn_category_id">
            <option value="">— none —</option>
            <option v-for="c in learnCategories" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </label>
        <div class="actions">
          <button type="button" class="btn" @click="cancelCat">Cancel</button>
          <button type="submit" class="btn primary">Save</button>
        </div>
      </form>
    </div>
  </section>
</template>

<style scoped>
.head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
.head h1 { margin: 0; font-size: 1.8rem; }
.loading, .muted { color: #94a3b8; }

.progress-bar {
  display: flex; align-items: center; gap: 0.6rem;
  padding: 0.5rem 0.85rem; margin-bottom: 1rem;
  background: rgba(56, 189, 248, 0.1); border: 1px solid rgba(56, 189, 248, 0.3);
  border-radius: 8px; font-size: 0.85rem; color: #38bdf8;
}
.pulse {
  width: 8px; height: 8px; border-radius: 50%;
  background: #38bdf8; animation: pulse 1.2s ease-in-out infinite;
}
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.3; } }

.list { list-style: none; padding: 0; margin: 0; display: grid; gap: 0.75rem; }
.card {
  background: rgba(30, 41, 59, 0.55);
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 12px;
  padding: 1rem 1.25rem;
}
.row { display: flex; align-items: flex-start; gap: 1rem; }
.grow { flex: 1; min-width: 0; }
.card h3 { margin: 0 0 0.25rem; font-size: 1.05rem; display: flex; align-items: center; gap: 0.5rem; }
.url { margin: 0 0 0.25rem; font-size: 0.85rem; color: #94a3b8; overflow: hidden; text-overflow: ellipsis; }
.meta { margin: 0; font-size: 0.8rem; color: #64748b; }
.pill {
  font-size: 0.7rem; padding: 0.1rem 0.45rem; border-radius: 999px;
  background: rgba(56, 189, 248, 0.2); color: #38bdf8;
}
.pill.muted { background: rgba(148, 163, 184, 0.2); color: #94a3b8; }
.tag { font-size: 0.7rem; padding: 0.1rem 0.45rem; border-radius: 999px; background: rgba(167, 139, 250, 0.2); color: #a78bfa; margin-left: 0.4rem; }

.actions { display: flex; gap: 0.35rem; flex-wrap: wrap; }

.detail { margin-top: 1rem; padding-top: 1rem; border-top: 1px solid rgba(148, 163, 184, 0.15); }
.detail h4 { margin: 1rem 0 0.5rem; display: flex; align-items: center; gap: 0.75rem; font-size: 0.9rem; }
.enqueue { display: flex; gap: 0.5rem; }
.enqueue input { flex: 1; padding: 0.5rem 0.7rem; border-radius: 6px; border: 1px solid rgba(148, 163, 184, 0.2); background: rgba(15, 23, 42, 0.6); color: #f1f5f9; }
.cats { list-style: none; padding: 0; margin: 0.5rem 0 0; display: grid; gap: 0.35rem; }
.cats li { display: flex; justify-content: space-between; align-items: center; padding: 0.4rem 0.6rem; background: rgba(15, 23, 42, 0.5); border-radius: 6px; font-size: 0.85rem; }
.cat-info { display: flex; align-items: center; flex-wrap: wrap; gap: 0.25rem; }
.cat-actions { display: flex; gap: 0.25rem; }

.btn {
  padding: 0.5rem 0.9rem; border-radius: 8px; border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(148, 163, 184, 0.08); color: #e2e8f0;
  font-size: 0.85rem; font-weight: 500; cursor: pointer;
}
.btn:hover { background: rgba(148, 163, 184, 0.15); }
.btn.primary { background: linear-gradient(135deg, #38bdf8, #a78bfa); color: #0f172a; border: 0; }
.btn.danger { color: #f87171; border-color: rgba(248, 113, 113, 0.4); }
.btn.small { padding: 0.35rem 0.65rem; font-size: 0.8rem; }
.btn.xs { padding: 0.2rem 0.45rem; font-size: 0.75rem; }

.modal {
  position: fixed; inset: 0; background: rgba(15, 23, 42, 0.8);
  display: flex; align-items: center; justify-content: center; z-index: 100; padding: 1rem;
}
.modal-card {
  background: #1e293b; border-radius: 14px; padding: 1.75rem;
  width: 100%; max-width: 500px; max-height: 90vh; overflow: auto;
}
.modal-card h2 { margin: 0 0 1.25rem; font-size: 1.25rem; }
.modal-card label { display: block; margin-bottom: 0.9rem; font-size: 0.85rem; color: #cbd5e1; }
.modal-card input, .modal-card select {
  width: 100%; padding: 0.55rem 0.7rem; margin-top: 0.3rem; border-radius: 6px;
  border: 1px solid rgba(148, 163, 184, 0.2); background: rgba(15, 23, 42, 0.6);
  color: #f1f5f9; font-size: 0.9rem;
}
.modal-card .check { display: flex; align-items: center; gap: 0.5rem; margin-top: 1.5rem; }
.modal-card .check input { width: auto; margin: 0; }
.modal-card .grid2 { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.modal-card .actions { display: flex; gap: 0.5rem; justify-content: flex-end; margin-top: 1rem; }
</style>
