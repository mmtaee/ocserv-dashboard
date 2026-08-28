import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/AdminLayout.vue'),
      children: [
        {
          path: '',
          name: 'admin-overview',
          component: () => import('@/pages/AdminOverviewPage.vue'),
          meta: { titleKey: 'admin.overview.title' },
        },
        {
          path: 'dashboard',
          name: 'admin-dashboard',
          component: () => import('@/pages/AdminDashboardPage.vue'),
          meta: { titleKey: 'admin.dashboard.title' },
        },
      ],
    },
  ],
  scrollBehavior: () => ({ top: 0 }),
})

export default router
