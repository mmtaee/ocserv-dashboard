import { createRouter, createWebHistory } from 'vue-router';
import MainRoutes from './MainRoutes';
import AuthRoutes from '@/router/AuthRoutes';
import UserRoutes from '@/router/UserRoutes';
import CustomerRoutes from '@/router/CustomerRoutes';
import SystemRoutes from '@/router/SystemRoutes';

export const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/:pathMatch(.*)*',
            component: () => import('@/views/Error404.vue')
        },
        MainRoutes,
        ...SystemRoutes,
        UserRoutes,
        AuthRoutes,
        CustomerRoutes
    ]
});
