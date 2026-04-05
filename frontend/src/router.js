import { createRouter, createWebHistory } from 'vue-router'
import Home from './views/Home.vue'
import Words from './views/Words.vue'
import Word from './views/Word.vue'
import About from './views/About.vue'

const routes = [
  { path: '/', name: 'home', component: Home },
  { path: '/words', name: 'words', component: Words },
  { path: '/words/:slug', name: 'word', component: Word, props: true },
  { path: '/about', name: 'about', component: About },
]

export default createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})
