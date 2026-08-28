import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/CustomerLayout.vue'),
      children: [
        {
          path: '',
          name: 'customer-home',
          component: () => import('@/pages/CustomerHomePage.vue'),
          meta: { titleKey: 'customer.home.title' },
        },
      ],
    },
  ],
  scrollBehavior: () => ({ top: 0 }),
})

export default router
