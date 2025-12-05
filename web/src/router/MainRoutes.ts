import type { RouteLocationNormalized } from 'vue-router';

const MainRoutes = {
    path: '/main',
    meta: {
        requiresAuth: true
    },
    redirect: '/main',
    component: () => import('@/layouts/full/FullLayout.vue'),
    children: [
        {
            name: 'Dashboard',
            path: '/',
            component: () => import('@/views/dashboard/index.vue')
        },
        {
            name: 'Ocserv Group Defaults',
            path: '/ocserv/management/groups/defaults',
            component: () => import('@/views/ocserv_group/OcservGroupDefaults.vue')
        },
        {
            name: 'Ocserv Groups',
            path: '/ocserv/management/groups',
            component: () => import('@/views/ocserv_group/index.vue')
        },
        {
            name: 'Ocserv Group Create',
            path: '/ocserv/management/groups/create',
            component: () => import('@/views/ocserv_group/OcservGroupCreate.vue')
        },
        {
            name: 'Ocserv Group Detail',
            path: '/ocserv/management/groups/:id',
            component: () => import('@/views/ocserv_group/OcservGroupDetail.vue'),
            props: true
        },
        {
            name: 'Ocserv Group Update',
            path: '/ocserv/management/groups/:id/update',
            component: () => import('@/views/ocserv_group/OcservGroupUpdate.vue'),
            props: true
        },
        {
            name: 'Ocserv Users',
            path: '/ocserv/management/users',
            component: () => import('@/views/ocserv_user/index.vue')
        },
        {
            name: 'Ocserv User Create',
            path: '/ocserv/management/users/create',
            component: () => import('@/views/ocserv_user/OcservUserCreate.vue')
        },
        {
            name: 'Ocserv User Detail',
            path: '/ocserv/management/users/:uid',
            component: () => import('@/views/ocserv_user/OcservUserDetail.vue'),
            props: true
        },
        {
            name: 'Ocserv User Update',
            path: '/ocserv/management/users/:uid/update',
            component: () => import('@/views/ocserv_user/OcservUserUpdate.vue'),
            props: true
        },
        {
            name: 'Ocserv User Statistics',
            path: '/ocserv/management/users/:uid/statistics',
            component: () => import('@/views/ocserv_user/OcservUserStatistics.vue'),
            props: (route: RouteLocationNormalized) => ({
                uid: route.params.uid as string,
                username: route.query.username as string | undefined
            })
        },
        {
            name: 'Ocserv Sync',
            path: '/ocserv/management/ocserv/sync',
            component: () => import('@/views/ocserv/Sync.vue'),
            props: false
        },
        {
            name: 'OCCTL',
            path: '/ocserv/occtl',
            component: () => import('@/views/ocserv/Occtl.vue')
        },
        {
            name: 'Stats',
            path: '/statistics',
            component: () => import('@/views/statistic/Stats.vue')
        },
        {
            name: 'Bandwidths',
            path: '/bandwidths',
            component: () => import('@/views/statistic/Bandwidths.vue')
        },
        {
            name: 'Server Logs',
            path: '/logs/server',
            component: () => import('@/views/loggers/Server.vue')
        }
    ]
};

export default MainRoutes;
