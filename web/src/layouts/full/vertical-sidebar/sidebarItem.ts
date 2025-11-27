import { useI18n } from 'vue-i18n';
import { useProfileStore } from '@/stores/profile';

export interface Menu {
    header?: string;
    title?: string;
    icon?: any;
    to?: string;
    chip?: string;
    chipColor?: string;
    chipBgColor?: string;
    chipVariant?: string;
    chipIcon?: string;
    children?: Menu[];
    disabled?: boolean;
    type?: string;
    subCaption?: string;
    external?: boolean;
}

export function getSidebarItems(): Menu[] {
    const { t } = useI18n();
    const profileStore = useProfileStore();

    let defaultSidebarItems: Menu[] = [
        { header: t('HOME') },
        {
            title: t('DASHBOARD'),
            icon: 'mdi-monitor-dashboard',
            to: '/'
        },
        { header: 'OCSERV' }
    ];

    if (profileStore.isAdmin) {
        defaultSidebarItems.push({
            title: t('GROUP_DEFAULTS'),
            icon: 'mdi-router',
            to: '/ocserv/management/groups/defaults'
        });
    }

    // These two always visible
    defaultSidebarItems.push(
        {
            title: t('GROUPS'),
            icon: 'mdi-router-network',
            to: '/ocserv/management/groups'
        },
        {
            title: t('USERS'),
            icon: 'mdi-account-network',
            to: '/ocserv/management/users'
        },
        {
            title: `${t('SYNC')} Ocpasswd`,
            // icon: 'mdi-account-convert-outline',
            icon: 'mdi-file-sync-outline',
            to: '/ocserv/management/users/sync'
        },
        {
            title: 'OCCTL',
            icon: 'mdi-console',
            to: '/ocserv/occtl'
        }
    );

    // Admin-only extra sections
    if (profileStore.isAdmin) {
        defaultSidebarItems.push(
            { header: t('STATISTICS') },
            {
                title: t('STATISTICS'),
                icon: 'mdi-chart-bar-stacked',
                to: '/statistics'
            },
            {
                title: t('BANDWIDTHS'),
                icon: 'mdi-speedometer',
                to: '/bandwidths'
            },
            { header: t('LOGS') },
            {
                title: t('SERVER'),
                icon: 'mdi-server-network',
                to: '/logs/server'
            },
            { header: t('STAFFS') },
            {
                title: t('STAFFS'),
                icon: 'mdi-account-tie-hat-outline',
                to: '/staffs'
            },
            {
                title: t('ACTIVITIES'),
                icon: 'mdi-history',
                to: '/staffs/activities'
            }
        );
    }

    return defaultSidebarItems;
}
