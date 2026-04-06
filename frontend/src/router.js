import { createRouter, createWebHistory } from 'vue-router'
import { auth } from './stores/auth.js'

import Login from './views/Login.vue'
import Review from './views/Review.vue'
import Learn from './views/Learn.vue'
import LearnArchive from './views/LearnArchive.vue'
import WordDetail from './views/WordDetail.vue'
import AdminSites from './views/AdminSites.vue'
import Settings from './views/Settings.vue'

const routes = [
  { path: '/', redirect: '/review' },
  { path: '/login', name: 'login', component: Login, meta: { public: true } },
  { path: '/review', name: 'review', component: Review },
  { path: '/learn', name: 'learn', component: Learn },
  { path: '/learn/archive', name: 'learn-archive', component: LearnArchive },
  { path: '/words/:id', name: 'word', component: WordDetail },
  { path: '/admin/sites', name: 'admin-sites', component: AdminSites },
  { path: '/settings', name: 'settings', component: Settings },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() { return { top: 0 } },
})

router.beforeEach((to) => {
  if (to.meta.public) {
    if (to.name === 'login' && auth.isAuthenticated) return { path: '/review' }
    return true
  }
  if (!auth.isAuthenticated) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  return true
})

export default router
