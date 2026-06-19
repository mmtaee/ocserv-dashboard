import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue'),
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('../views/Profile.vue'),
  },
  {
    path: '/usage',
    name: 'Usage',
    component: () => import('../views/Usage.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
