import { reactive } from 'vue'
import { api, getToken, getUser, setToken, setUser, clearAuth } from '../lib/api.js'

export const auth = reactive({
  user: getUser(),
  token: getToken(),
  get isAuthenticated() {
    return !!this.token
  },
  async login(email, password) {
    const data = await api('/auth/login', { method: 'POST', body: { email, password } })
    setToken(data.token)
    setUser(data.user)
    this.token = data.token
    this.user = data.user
    return data
  },
  async fetchMe() {
    try {
      const data = await api('/auth/me')
      setUser(data)
      this.user = data
      return data
    } catch (e) {
      this.logout()
      throw e
    }
  },
  logout() {
    clearAuth()
    this.token = ''
    this.user = null
  },
})
