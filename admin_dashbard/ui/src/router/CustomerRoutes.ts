const hasCustomerCiscoSetup = () => {
    const raw = sessionStorage.getItem('customerCiscoSetup');

    if (!raw) {
        return false;
    }

    try {
        JSON.parse(raw);
        return true;
    } catch {
        sessionStorage.removeItem('customerCiscoSetup');
        return false;
    }
};

const CustomerRoutes = {
    path: '/customers',
    component: () => import('@/layouts/blank/BlankLayout.vue'),
    meta: {
        requiresAuth: false
    },
    children: [
        {
            name: 'Customer Summary',
            path: '/summary',
            component: () => import('@/views/customer/Summary.vue')
        },
        {
            name: 'CustomerCiscoSetup',
            path: '/cisco-setup',
            beforeEnter: () => {
                if (hasCustomerCiscoSetup()) {
                    return true;
                }

                return { name: 'Customer Summary', replace: true };
            },
            component: () => import('@/views/customer/CiscoSetup.vue')
        }
    ]
};

export default CustomerRoutes;
