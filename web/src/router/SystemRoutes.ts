import type { RouteRecordRaw } from 'vue-router';

const SystemRoutes: RouteRecordRaw[] = [
    // Setup route with BlankLayout
    {
        path: '/setup',
        meta: { requiresAuth: true },
        component: () => import('@/layouts/blank/BlankLayout.vue'),
        children: [
            {
                path: '', // empty for default child
                name: 'System Setup',
                component: () => import('@/views/system/Setup.vue')
            }
        ]
    },

    // System Update route with FullLayout
    {
        path: '/system',
        meta: { requiresAuth: true },
        component: () => import('@/layouts/full/FullLayout.vue'),
        children: [
            {
                path: '', // NO leading slash
                name: 'System',
                component: () => import('@/views/system/System.vue')
            }
        ]
    },
    {
        path: '/backup',
        meta: { requiresAuth: true },
        component: () => import('@/layouts/full/FullLayout.vue'),
        children: [
            {
                path: '', // NO leading slash
                name: 'Backup',
                component: () => import('@/views/system/Backup.vue')
            }
        ]
    }
];

export default SystemRoutes;
