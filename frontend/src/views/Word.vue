<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'

const props = defineProps({
  slug: String,
})

const word = ref(null)
const notFound = ref(false)

onMounted(async () => {
  const res = await fetch(`/api/words/${props.slug}`)
  if (res.status === 404) {
    notFound.value = true
    return
  }
  const { data } = await res.json()
  word.value = data
})
</script>

<template>
  <article class="detail">
    <RouterLink to="/words" class="back">← Back to words</RouterLink>

    <p v-if="notFound" class="missing">Word not found.</p>

    <div v-else-if="word" class="card">
      <div class="emoji-wrap">
        <span class="emoji">{{ word.emoji }}</span>
        <div class="ring"></div>
      </div>
      <h1>{{ word.word }}</h1>
      <p class="definition">{{ word.definition }}</p>
      <blockquote>
        <span class="quote-mark">“</span>
        {{ word.example }}
      </blockquote>
    </div>
  </article>
</template>

<style scoped>
.detail { max-width: 620px; margin: 0 auto; }
.missing { text-align: center; color: #94a3b8; }

.back {
  display: inline-block;
  font-size: 0.9rem;
  color: #94a3b8;
  margin-bottom: 1.5rem;
  transition: color 0.2s ease, transform 0.2s ease;
}
.back:hover { color: #38bdf8; transform: translateX(-4px); }

.card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 20px;
  padding: 3rem 2rem;
  text-align: center;
  position: relative;
  overflow: hidden;
}

.emoji-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 120px;
  height: 120px;
  margin-bottom: 1.5rem;
}
.emoji {
  font-size: 4rem;
  animation: bounce 0.8s cubic-bezier(0.34, 1.56, 0.64, 1);
  position: relative;
  z-index: 2;
}
.ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(56, 189, 248, 0.3), transparent 70%);
  animation: pulse 2.5s ease-in-out infinite;
}
@keyframes bounce {
  0%   { transform: scale(0) rotate(-180deg); opacity: 0; }
  60%  { transform: scale(1.15) rotate(10deg); opacity: 1; }
  100% { transform: scale(1) rotate(0); }
}
@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 0.6; }
  50%      { transform: scale(1.2); opacity: 0.3; }
}

h1 {
  font-size: 2.8rem;
  margin: 0 0 1rem;
  background: linear-gradient(90deg, #38bdf8, #a78bfa);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  animation: slideUp 0.5s 0.2s cubic-bezier(0.22, 1, 0.36, 1) both;
}
.definition {
  font-size: 1.15rem;
  color: #cbd5e1;
  max-width: 460px;
  margin: 0 auto 2rem;
  line-height: 1.6;
  animation: slideUp 0.5s 0.3s cubic-bezier(0.22, 1, 0.36, 1) both;
}
blockquote {
  margin: 0 auto;
  max-width: 460px;
  padding: 1.25rem 1.5rem;
  background: rgba(15, 23, 42, 0.5);
  border-left: 3px solid #a78bfa;
  border-radius: 8px;
  color: #e2e8f0;
  font-style: italic;
  text-align: left;
  position: relative;
  animation: slideUp 0.5s 0.4s cubic-bezier(0.22, 1, 0.36, 1) both;
}
.quote-mark {
  font-size: 2rem;
  color: #a78bfa;
  line-height: 0;
  vertical-align: -0.25em;
  margin-right: 0.2em;
}
@keyframes slideUp {
  from { opacity: 0; transform: translateY(16px); }
  to   { opacity: 1; transform: translateY(0); }
}
</style>
