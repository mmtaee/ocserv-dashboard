import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue'),
  },
  {
    path: '/users',
    name: 'Users',
    component: () => import('../views/Users.vue'),
  },
  {
    path: '/groups',
    name: 'Groups',
    component: () => import('../views/Groups.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
