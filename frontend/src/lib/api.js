// Shared API client — attaches JWT bearer token, throws on error,
// returns unwrapped `data` from the `{ data: ... }` envelope.

const TOKEN_KEY = 'words.token'
const USER_KEY = 'words.user'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token) {
  if (token) localStorage.setItem(TOKEN_KEY, token)
  else localStorage.removeItem(TOKEN_KEY)
}

export function getUser() {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) return null
  try { return JSON.parse(raw) } catch { return null }
}

export function setUser(user) {
  if (user) localStorage.setItem(USER_KEY, JSON.stringify(user))
  else localStorage.removeItem(USER_KEY)
}

export function clearAuth() {
  setToken('')
  setUser(null)
}

export async function api(path, { method = 'GET', body, signal } = {}) {
  const headers = { 'Content-Type': 'application/json' }
  const token = getToken()
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`/api${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
    signal,
  })

  if (res.status === 401) {
    clearAuth()
    // soft redirect
    if (!location.pathname.startsWith('/login')) {
      location.href = '/login'
    }
    throw new Error('unauthorized')
  }

  let payload = null
  try { payload = await res.json() } catch { /* ignore */ }

  if (!res.ok) {
    const msg = payload?.errors || payload?.error || `HTTP ${res.status}`
    throw new Error(msg)
  }
  return payload?.data
}
