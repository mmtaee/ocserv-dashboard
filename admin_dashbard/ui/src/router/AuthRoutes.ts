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
        },
        {
            name: 'Admin Reset Password',
            path: '/login/admin/reset-password',
            component: () => import('@/views/auth/AdminResetPassword.vue')
        }
    ]
};

export default AuthRoutes;
