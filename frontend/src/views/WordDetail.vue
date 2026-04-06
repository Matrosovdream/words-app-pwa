<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { api } from '../lib/api.js'

const route = useRoute()
const word = ref(null)
const loading = ref(true)
const error = ref('')
const targetLang = ref('ru')

async function load() {
  loading.value = true
  error.value = ''
  try {
    word.value = await api(`/words/${route.params.id}?lang=${targetLang.value}`)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function changeLang(lang) {
  targetLang.value = lang
  await load()
}

onMounted(load)
</script>

<template>
  <article class="detail">
    <RouterLink to="/learn" class="back">← Back</RouterLink>

    <p v-if="loading" class="muted">Loading…</p>
    <p v-else-if="error" class="err">{{ error }}</p>

    <template v-else-if="word">
      <div class="hero">
        <h1>{{ word.lemma }}</h1>
        <div class="meta">
          <span class="lang">{{ word.language }}</span>
          <span v-if="word.ipa">/{{ word.ipa }}/</span>
          <span v-if="word.pos">· {{ word.pos }}</span>
          <span v-if="word.frequency_rank">· rank #{{ word.frequency_rank }}</span>
        </div>
      </div>

      <section v-if="word.definition" class="block">
        <h2>Definition</h2>
        <p>{{ word.definition }}</p>
      </section>

      <section class="block">
        <div class="row">
          <h2>Translation</h2>
          <div class="lang-switch">
            <button
              v-for="l in ['ru','de','es','fr']" :key="l"
              class="lang-btn" :class="{ active: targetLang === l }"
              @click="changeLang(l)"
            >{{ l.toUpperCase() }}</button>
          </div>
        </div>
        <ul v-if="word.translations && word.translations.length" class="translations">
          <li v-for="(t, i) in word.translations" :key="i">
            <strong>{{ t.translation }}</strong>
            <span class="src">— {{ t.source }}</span>
          </li>
        </ul>
        <p v-else class="muted">No translation available. Configure DeepL key in Settings.</p>
      </section>

      <section v-if="word.synonyms && word.synonyms.length" class="block">
        <h2>Synonyms</h2>
        <div class="chips">
          <span v-for="(s, i) in word.synonyms" :key="i" class="chip">{{ s.text }}</span>
        </div>
      </section>

      <section v-if="word.antonyms && word.antonyms.length" class="block">
        <h2>Antonyms</h2>
        <div class="chips">
          <span v-for="(a, i) in word.antonyms" :key="i" class="chip chip-ant">{{ a.text }}</span>
        </div>
      </section>

      <section v-if="word.occurrences && word.occurrences.length" class="block">
        <h2>Seen in</h2>
        <ul class="occs">
          <li v-for="(o, i) in word.occurrences" :key="i">
            <p class="sample">"{{ o.sentence }}"</p>
            <a v-if="o.source_url" :href="o.source_url" target="_blank" rel="noopener" class="source">
              {{ o.page_title || o.source_url }} ↗
            </a>
          </li>
        </ul>
      </section>
    </template>
  </article>
</template>

<style scoped>
.detail { max-width: 680px; margin: 0 auto; }
.back { font-size: 0.9rem; color: #94a3b8; display: inline-block; margin-bottom: 1.5rem; }
.back:hover { color: #38bdf8; }
.muted { color: #94a3b8; }
.err { color: #f87171; }

.hero { text-align: center; margin-bottom: 2rem; }
.hero h1 {
  font-size: 3rem; margin: 0 0 0.5rem;
  background: linear-gradient(90deg, #38bdf8, #a78bfa); -webkit-background-clip: text; background-clip: text; color: transparent;
}
.meta { color: #64748b; font-size: 0.9rem; display: flex; gap: 0.5rem; justify-content: center; flex-wrap: wrap; }
.lang { text-transform: uppercase; font-weight: 600; }

.block {
  background: rgba(30, 41, 59, 0.55);
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 12px;
  padding: 1.25rem 1.5rem;
  margin-bottom: 1rem;
}
.block h2 { margin: 0 0 0.75rem; font-size: 1rem; color: #38bdf8; letter-spacing: 0.02em; text-transform: uppercase; }
.block p { margin: 0; color: #e2e8f0; line-height: 1.6; }

.row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.75rem; }
.row h2 { margin: 0; }
.lang-switch { display: flex; gap: 0.25rem; }
.lang-btn {
  padding: 0.2rem 0.55rem; border-radius: 6px; font-size: 0.7rem;
  background: rgba(148, 163, 184, 0.1); border: 1px solid rgba(148, 163, 184, 0.2);
  color: #94a3b8; cursor: pointer;
}
.lang-btn.active { background: #38bdf8; color: #0f172a; border-color: #38bdf8; font-weight: 600; }

.translations { list-style: none; padding: 0; margin: 0; }
.translations li { padding: 0.4rem 0; border-bottom: 1px solid rgba(148, 163, 184, 0.1); }
.translations li:last-child { border: 0; }
.translations strong { font-size: 1.1rem; color: #f1f5f9; }
.src { font-size: 0.75rem; color: #64748b; margin-left: 0.5rem; }

.chips { display: flex; flex-wrap: wrap; gap: 0.4rem; }
.chip {
  padding: 0.3rem 0.7rem; border-radius: 999px; font-size: 0.85rem;
  background: rgba(56, 189, 248, 0.15); color: #38bdf8;
}
.chip-ant { background: rgba(248, 113, 113, 0.15); color: #f87171; }

.occs { list-style: none; padding: 0; margin: 0; }
.occs li { padding: 0.5rem 0; border-bottom: 1px solid rgba(148, 163, 184, 0.1); }
.occs li:last-child { border: 0; }
.sample { font-style: italic; color: #cbd5e1; margin: 0 0 0.25rem; }
.source { font-size: 0.8rem; color: #38bdf8; }
</style>
