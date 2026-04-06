<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { auth } from '../stores/auth.js'

const email = ref('admin@words.local')
const password = ref('admin')
const loading = ref(false)
const error = ref('')
const router = useRouter()
const route = useRoute()

async function submit() {
  loading.value = true
  error.value = ''
  try {
    await auth.login(email.value, password.value)
    const next = route.query.redirect || '/review'
    router.replace(next)
  } catch (e) {
    error.value = e.message === 'unauthorized' ? 'Invalid email or password' : e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <section class="login">
    <form class="card" @submit.prevent="submit">
      <h1>Sign in</h1>
      <label>
        <span>Email</span>
        <input type="email" v-model="email" required autofocus />
      </label>
      <label>
        <span>Password</span>
        <input type="password" v-model="password" required />
      </label>
      <button type="submit" :disabled="loading">
        {{ loading ? 'Signing in…' : 'Sign in' }}
      </button>
      <p v-if="error" class="err">{{ error }}</p>
    </form>
  </section>
</template>

<style scoped>
.login { max-width: 420px; margin: 4rem auto; }
.card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 14px;
  padding: 2rem;
}
h1 { margin: 0 0 1.25rem; font-size: 1.6rem; }
label { display: block; margin-bottom: 1rem; }
label span { display: block; font-size: 0.85rem; color: #94a3b8; margin-bottom: 0.35rem; }
input {
  width: 100%;
  padding: 0.65rem 0.8rem;
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  background: rgba(15, 23, 42, 0.6);
  color: #f1f5f9;
  font-size: 0.95rem;
}
input:focus { outline: none; border-color: #38bdf8; }
button {
  width: 100%;
  padding: 0.7rem 1rem;
  border-radius: 8px;
  border: 0;
  background: linear-gradient(135deg, #38bdf8, #a78bfa);
  color: #0f172a;
  font-weight: 600;
  cursor: pointer;
}
button:disabled { opacity: 0.6; cursor: not-allowed; }
.err { color: #f87171; font-size: 0.9rem; margin: 0.75rem 0 0; }
</style>
