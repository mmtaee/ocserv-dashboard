import type { RouteLocationNormalized } from 'vue-router';

const AuthRoutes = {
    path: '/user',
    component: () => import('@/layouts/full/FullLayout.vue'),
    meta: {
        requiresAuth: true
    },
    children: [
        {
            name: 'Profile',
            path: '/profile',
            component: () => import('@/views/user/Profile.vue')
        },
        {
            name: 'Staffs',
            path: '/staffs',
            component: () => import('@/views/user/index.vue')
        },
        {
            name: 'Staff Create',
            path: '/staffs/create',
            component: () => import('@/views/user/StaffCreate.vue')
        },
        {
            name: 'Staff Activities',
            path: '/staffs/activities',
            component: () => import('@/views/user/StaffActivities.vue'),
            props: (route: RouteLocationNormalized) => ({
                uid: route.query.uid as string | undefined,
                username: route.query.username as string | undefined
            })
        }
    ]
};

export default AuthRoutes;
