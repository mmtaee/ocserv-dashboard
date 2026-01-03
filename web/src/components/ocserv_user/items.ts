import { reactive } from 'vue';
import { useI18n } from 'vue-i18n';
import { domainRule, ipOrRangeRule, ipRule, ipWithRangeRule } from '@/utils/rules';

function getFormFields() {
    const { t } = useI18n();
    const rules = {
        ip: (v: string) => ipRule(v, t),
        ipOrRange: (v: string) => ipOrRangeRule(v, t),
        domain: (v: string) => domainRule(v, t),
        ipWithRange: (v: string) => ipWithRangeRule(v, t)
    };

    const fields = [
        // Network Configuration
        {
            key: 'nbns',
            label: 'NBNS',
            type: 'text',
            hint: 'Net BIOS',
            example: '192.168.1.10',
            rules: [rules.ip]
        },
        {
            key: 'ipv4-network',
            label: 'IPv4 Network',
            type: 'text',
            hint: 'CIDR',
            example: '192.168.0.0/24',
            rules: [rules.ipWithRange]
        },
        {
            key: 'explicit-ipv4',
            label: 'Explicit IPv4',
            type: 'text',
            hint: t('SPECIFIC_IP_ADDRESS'),
            example: '192.168.1.5',
            rules: [rules.ip]
        },
        {
            key: 'iroute',
            label: 'Internal Route',
            type: 'text',
            hint: t('CUSTOM_INTERNAL_ROUTE'),
            example: ' 10.0.0.0/8 ',
            rules: [rules.ipOrRange]
        },
        {
            key: 'restrict-to-ports',
            label: 'Restrict User To Ports',
            type: 'text',
            hint: t('ALLOWED_PORTS'),
            example: '80,443'
        },

        // Performance and Session Settings
        { key: 'idle-timeout', label: 'Idle Timeout', type: 'number', hint: t('INACTIVITY_TIMEOUT_S') },
        {
            key: 'mobile-idle-timeout',
            label: 'Mobile Idle Timeout',
            type: 'number',
            hint: t('MOBILE_INACTIVITY_TIMEOUT_S')
        },
        { key: 'session-timeout', label: 'Session Timeout', type: 'number', hint: t('MAX_SESSION_DURATION_S') },
        {
            key: 'rekey-time',
            label: 'Rekey Time',
            type: 'number',
            hint: t('TRIGGERS_KEY_RENEGOTIATION'),
            example: '86400 for 24 hours'
        },

        // Access and Feature Controls
        {
            key: 'restrict-to-routes',
            label: 'Restrict User To Routes',
            type: 'switch',
            hint: t('ALLOW_CLIENT_ACCESS_ONLY_TO_DEFINED_ROUTES')
        }
    ];

    const textFields = [
        // Routes
        {
            key: 'route',
            label: 'Route',
            type: 'text',
            example: '10.0.0.0/8',
            hint: t('ROUTES_ASSIGNED_TO_CLIENT'),
            rules: [rules.ipOrRange]
        },
        {
            key: 'no-route',
            label: 'No Route',
            type: 'text',
            hint: t('NON_VPN_NETWORKS'),
            example: '172.16.0.0/12',
            rules: [rules.ipOrRange]
        },
        {
            key: 'dns',
            label: 'DNS',
            type: 'text',
            hint: t('DNS_SERVERS_LIST'),
            example: '8.8.8.8/example.com',
            rules: [rules.ip]
        },
        {
            key: 'split-dns',
            label: 'Split DNS',
            type: 'text',
            hint: t('DNS_SPECIFIC_DOMAINS'),
            example: 'example.com',
            rules: [rules.domain]
        }
    ];

    const chipInputs = reactive<Record<string, string>>({
        dns: '',
        route: '',
        'no-route': '',
        'split-dns': ''
    });

    return { fields, textFields, chipInputs };
}

export { getFormFields };
