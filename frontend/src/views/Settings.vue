<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { api } from '../lib/api.js'

const settings = ref(null)
const stats = ref(null)
const loading = ref(true)
const saving = ref(false)
const message = ref('')
const form = ref({
  deepl_api_key: '',
  deepl_endpoint: '',
  default_target_language: 'ru',
  relations_source: 'datamuse',
})
let pollTimer = null

async function loadSettings() {
  const s = await api('/admin/settings/translation')
  settings.value = s
  form.value.deepl_endpoint = s.deepl_endpoint
  form.value.default_target_language = s.default_target_language
  form.value.relations_source = s.relations_source
  form.value.deepl_api_key = ''
}

async function loadStats() {
  try { stats.value = await api('/admin/parser/stats') } catch { /* ignore */ }
}

async function save() {
  saving.value = true
  message.value = ''
  try {
    const payload = {}
    if (form.value.deepl_api_key) payload.deepl_api_key = form.value.deepl_api_key
    if (form.value.deepl_endpoint) payload.deepl_endpoint = form.value.deepl_endpoint
    if (form.value.default_target_language) payload.default_target_language = form.value.default_target_language
    if (form.value.relations_source) payload.relations_source = form.value.relations_source
    settings.value = await api('/admin/settings/translation', { method: 'PUT', body: payload })
    form.value.deepl_api_key = ''
    message.value = 'Saved.'
    setTimeout(() => message.value = '', 2500)
  } catch (e) {
    message.value = 'Error: ' + e.message
  } finally { saving.value = false }
}

function fmtTime(ts) {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  const now = new Date()
  const diff = (now - d) / 1000
  if (diff < 60) return `${Math.floor(diff)}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return d.toLocaleDateString()
}

function shortUrl(u) {
  if (!u) return ''
  return u.length > 60 ? u.slice(0, 57) + '…' : u
}

onMounted(async () => {
  loading.value = true
  try { await Promise.all([loadSettings(), loadStats()]) }
  finally { loading.value = false }
  pollTimer = setInterval(loadStats, 4000)
})
onUnmounted(() => { if (pollTimer) clearInterval(pollTimer) })
</script>

<template>
  <section class="settings">
    <h1>Settings</h1>

    <div v-if="loading" class="muted">Loading…</div>

    <template v-else>
      <!-- Parser progress -->
      <div v-if="stats" class="card progress">
        <div class="row">
          <h2>Parser</h2>
          <span class="status" :class="stats.worker.enabled ? 'on' : 'off'">
            {{ stats.worker.enabled ? '● Running' : '○ Disabled' }}
          </span>
        </div>
        <p class="hint">
          Polls every {{ stats.worker.poll_interval_seconds }}s ·
          common-word threshold: rank ≤ {{ stats.worker.common_word_threshold }}
        </p>

        <div class="grid-4">
          <div class="stat">
            <span class="n">{{ stats.jobs.pending }}</span>
            <span class="k">pending</span>
          </div>
          <div class="stat">
            <span class="n hot">{{ stats.jobs.running }}</span>
            <span class="k">running</span>
          </div>
          <div class="stat">
            <span class="n ok">{{ stats.jobs.done }}</span>
            <span class="k">done</span>
          </div>
          <div class="stat">
            <span class="n err">{{ stats.jobs.failed }}</span>
            <span class="k">failed</span>
          </div>
        </div>

        <div class="grid-4">
          <div class="stat sm">
            <span class="n">{{ stats.counts.dict_words }}</span><span class="k">dict words</span>
          </div>
          <div class="stat sm">
            <span class="n">{{ stats.counts.pending_review }}</span><span class="k">to review</span>
          </div>
          <div class="stat sm">
            <span class="n">{{ stats.counts.active_learn }}</span><span class="k">learning</span>
          </div>
          <div class="stat sm">
            <span class="n">{{ stats.counts.archived_learn }}</span><span class="k">archived</span>
          </div>
        </div>

        <div v-if="stats.sites && stats.sites.length" class="subsection">
          <h3>Per-site today</h3>
          <ul class="sites">
            <li v-for="s in stats.sites" :key="s.id" :class="{ off: !s.is_active }">
              <span class="name">{{ s.name }}</span>
              <span class="limit">
                {{ s.hits_today }} / {{ s.daily_hit_limit }}
                <span v-if="!s.is_active" class="tag">paused</span>
              </span>
              <span class="bar">
                <span class="fill" :style="{ width: Math.min(100, s.daily_hit_limit ? (s.hits_today / s.daily_hit_limit) * 100 : 0) + '%' }"></span>
              </span>
            </li>
          </ul>
        </div>

        <div v-if="stats.recent_jobs && stats.recent_jobs.length" class="subsection">
          <h3>Recent jobs</h3>
          <ul class="jobs">
            <li v-for="j in stats.recent_jobs" :key="j.id">
              <span class="status-dot" :class="j.status"></span>
              <span class="job-url" :title="j.url">{{ shortUrl(j.url) }}</span>
              <span class="job-site">{{ j.site_name }}</span>
              <span class="job-time">{{ fmtTime(j.finished_at || j.started_at || j.scheduled_at) }}</span>
            </li>
          </ul>
        </div>

        <div v-else-if="!stats.jobs.done && !stats.jobs.pending" class="muted mt">
          No jobs yet. Add a site and queue a URL to start parsing.
        </div>
      </div>

      <!-- Translation settings -->
      <form class="card" @submit.prevent="save">
        <h2>DeepL</h2>
        <p class="hint">
          Free tier: <code>https://api-free.deepl.com/v2/translate</code> —
          sign up at deepl.com/pro-api for an API Free key.
        </p>
        <label>
          <span>API Key <em v-if="settings.deepl_api_key_set">(set — leave blank to keep)</em></span>
          <input v-model="form.deepl_api_key" type="password" placeholder="DeepL API key" />
        </label>
        <label>
          <span>Endpoint</span>
          <input v-model="form.deepl_endpoint" type="url" />
        </label>

        <h2>Languages</h2>
        <label>
          <span>Default target language</span>
          <select v-model="form.default_target_language">
            <option value="ru">Russian (ru)</option>
            <option value="de">German (de)</option>
            <option value="es">Spanish (es)</option>
            <option value="fr">French (fr)</option>
            <option value="it">Italian (it)</option>
            <option value="ja">Japanese (ja)</option>
            <option value="zh">Chinese (zh)</option>
          </select>
        </label>

        <h2>Synonyms / Antonyms</h2>
        <label>
          <span>Relations source</span>
          <select v-model="form.relations_source">
            <option value="datamuse">Datamuse (free, English only)</option>
          </select>
        </label>

        <div class="actions">
          <button type="submit" :disabled="saving" class="btn primary">
            {{ saving ? 'Saving…' : 'Save' }}
          </button>
          <span v-if="message" class="msg">{{ message }}</span>
        </div>
      </form>
    </template>
  </section>
</template>

<style scoped>
.settings { max-width: 680px; margin: 0 auto; }
.settings h1 { margin: 0 0 1rem; font-size: 1.8rem; }
.muted { color: #94a3b8; }
.mt { margin-top: 0.5rem; }

.card {
  background: rgba(30, 41, 59, 0.55);
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 14px;
  padding: 1.5rem;
  margin-bottom: 1rem;
}

/* Progress panel */
.progress .row { display: flex; justify-content: space-between; align-items: center; }
.progress h2 { margin: 0; font-size: 0.9rem; color: #38bdf8; text-transform: uppercase; letter-spacing: 0.04em; }
.status { font-size: 0.8rem; font-weight: 600; }
.status.on { color: #4ade80; }
.status.off { color: #94a3b8; }
.hint { font-size: 0.8rem; color: #94a3b8; margin: 0.4rem 0 1rem; line-height: 1.5; }
.hint code { background: rgba(148, 163, 184, 0.1); padding: 0.1em 0.35em; border-radius: 3px; }

.grid-4 {
  display: grid; grid-template-columns: repeat(4, 1fr); gap: 0.5rem; margin-bottom: 0.5rem;
}
.stat {
  background: rgba(15, 23, 42, 0.5);
  padding: 0.75rem 0.85rem;
  border-radius: 8px;
  display: flex; flex-direction: column; gap: 0.1rem;
}
.stat .n { font-size: 1.4rem; font-weight: 700; color: #f1f5f9; font-variant-numeric: tabular-nums; }
.stat.sm .n { font-size: 1.1rem; }
.stat .k { font-size: 0.7rem; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.05em; }
.stat .n.hot { color: #fbbf24; }
.stat .n.ok { color: #4ade80; }
.stat .n.err { color: #f87171; }

.subsection { margin-top: 1.25rem; }
.subsection h3 {
  font-size: 0.75rem; color: #64748b; text-transform: uppercase;
  letter-spacing: 0.06em; margin: 0 0 0.5rem;
}

.sites { list-style: none; padding: 0; margin: 0; }
.sites li {
  display: grid; grid-template-columns: 1fr auto; grid-template-rows: auto auto; gap: 0.15rem 0.75rem;
  padding: 0.5rem 0; border-bottom: 1px solid rgba(148, 163, 184, 0.08);
  font-size: 0.85rem;
}
.sites li:last-child { border: 0; }
.sites li.off { opacity: 0.5; }
.name { color: #e2e8f0; font-weight: 500; }
.limit { color: #94a3b8; font-size: 0.8rem; font-variant-numeric: tabular-nums; }
.bar { grid-column: 1 / -1; height: 3px; background: rgba(148, 163, 184, 0.15); border-radius: 2px; overflow: hidden; }
.fill { display: block; height: 100%; background: linear-gradient(90deg, #38bdf8, #a78bfa); transition: width 0.3s; }
.tag { font-size: 0.65rem; padding: 0.05rem 0.4rem; border-radius: 999px; background: rgba(148, 163, 184, 0.2); color: #94a3b8; margin-left: 0.35rem; }

.jobs { list-style: none; padding: 0; margin: 0; }
.jobs li {
  display: flex; align-items: center; gap: 0.6rem;
  padding: 0.35rem 0; font-size: 0.8rem; border-bottom: 1px solid rgba(148, 163, 184, 0.08);
}
.jobs li:last-child { border: 0; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.status-dot.pending { background: #64748b; }
.status-dot.running { background: #fbbf24; box-shadow: 0 0 6px #fbbf24; animation: pulse 1.2s ease-in-out infinite; }
.status-dot.done { background: #4ade80; }
.status-dot.failed { background: #f87171; }
.status-dot.skipped { background: #a78bfa; }
@keyframes pulse { 50% { opacity: 0.4; } }
.job-url { flex: 1; color: #cbd5e1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-family: ui-monospace, SFMono-Regular, monospace; font-size: 0.75rem; }
.job-site { color: #64748b; font-size: 0.75rem; }
.job-time { color: #64748b; font-size: 0.75rem; font-variant-numeric: tabular-nums; }

/* Settings form */
h2 { margin: 1rem 0 0.75rem; font-size: 0.9rem; color: #38bdf8; text-transform: uppercase; letter-spacing: 0.04em; }
h2:first-child { margin-top: 0; }
label { display: block; margin-bottom: 0.9rem; }
label span { display: block; font-size: 0.85rem; color: #cbd5e1; margin-bottom: 0.3rem; }
label em { color: #64748b; font-style: normal; font-size: 0.8rem; }
input, select {
  width: 100%; padding: 0.55rem 0.75rem; border-radius: 6px;
  border: 1px solid rgba(148, 163, 184, 0.2); background: rgba(15, 23, 42, 0.6);
  color: #f1f5f9; font-size: 0.9rem;
}
input:focus, select:focus { outline: none; border-color: #38bdf8; }

.actions { display: flex; align-items: center; gap: 1rem; margin-top: 1rem; }
.btn {
  padding: 0.6rem 1.2rem; border-radius: 8px; border: 0;
  background: linear-gradient(135deg, #38bdf8, #a78bfa); color: #0f172a;
  font-weight: 600; cursor: pointer;
}
.btn:disabled { opacity: 0.6; cursor: not-allowed; }
.msg { font-size: 0.85rem; color: #4ade80; }

@media (max-width: 500px) {
  .grid-4 { grid-template-columns: repeat(2, 1fr); }
}
</style>
