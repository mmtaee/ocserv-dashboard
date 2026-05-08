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
        if (import.meta.env.VITE_SYSTEMD == 'true') {
            defaultSidebarItems.push({
                title: t('TOOLS') + ' (systemd)',
                icon: 'mdi-tools',
                to: '/ocserv/management/systemd_tools'
            });
        }

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
        }
    );

    if (profileStore.isAdmin) {
        defaultSidebarItems.push(
            {
                title: 'OCCTL',
                icon: 'mdi-console',
                to: '/ocserv/occtl'
            },
            {
                title: t('SYNC'),
                // icon: 'mdi-account-convert-outline',
                icon: 'mdi-file-sync-outline',
                to: '/ocserv/management/ocserv/sync'
            }
        );
    }

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
            {
                title: t('SESSION_LOGS'),
                icon: 'mdi-timeline-text-outline',
                to: '/session_logs'
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
            },
            { header: t('TELEGRAM') },
            {
                title: t('TELEGRAM_REQUESTS'),
                icon: 'mdi-tray-full',
                to: '/telegram/requests'
            },
            {
                title: t('TELEGRAM_PACKAGES'),
                icon: 'mdi-package-variant',
                to: '/telegram/packages'
            },
            {
                title: t('TELEGRAM_SETTINGS'),
                icon: 'mdi-robot',
                to: '/telegram/settings'
            },
            { header: t('SYSTEM') },
            {
                title: t('SETTINGS'),
                icon: 'mdi-cog',
                to: '/system'
            },
            {
                title: t('BACKUP'),
                icon: 'mdi-database',
                to: '/backup'
            }
        );
    }

    return defaultSidebarItems;
}
