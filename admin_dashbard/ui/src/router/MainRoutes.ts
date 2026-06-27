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
            name: 'Ocserv Tools',
            path: '/ocserv/management/systemd_tools',
            component: () => import('@/views/ocserv/Tools.vue')
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
            name: 'Statistics',
            path: '/statistics',
            component: () => import('@/views/statistic/Statistics.vue')
        },
        {
            name: 'Session Logs',
            path: '/session_logs',
            component: () => import('@/views/statistic/SessionLogs.vue')
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
        },
        {
            name: 'Telegram Settings',
            path: '/telegram/settings',
            component: () => import('@/views/telegram/Settings.vue')
        },
        {
            name: 'Telegram Packages',
            path: '/telegram/packages',
            component: () => import('@/views/telegram/Packages.vue')
        },
        {
            name: 'Telegram Requests',
            path: '/telegram/requests',
            component: () => import('@/views/telegram/Requests.vue')
        }
    ]
};

export default MainRoutes;
