const AuthRoutes = {
    path: '/',
    component: () => import('@/layouts/blank/BlankLayout.vue'),
    meta: {
        requiresAuth: false
    },
    children: [
        {
            name: 'Login',
            path: '/login',
            component: () => import('@/views/auth/Login.vue')
        },
        {
            name: 'Admin Login',
            path: '/login/admin',
            component: () => import('@/views/auth/AdminLogin.vue')
        }
    ]
};

export default AuthRoutes;
