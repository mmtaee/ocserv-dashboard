import type {
    HomeDockerService,
    HomeGetHomeUser,
    ModelsDailyTraffic,
    ModelsIPBanPoints,
    ModelsOcservGroup,
    ModelsOcservGroupConfig,
    ModelsOcservUser,
    ModelsOcservUserTrafficTypeEnum,
    ModelsOnlineUserSession,
    RepositoryTopBandwidthUsers,
    RepositoryTotalBandwidths
} from '@/api';

const dummyTrafficData = <ModelsDailyTraffic[]>[
    { date: '2025-06-17', rx: 1.2, tx: 2.5 },
    { date: '2025-06-18', rx: 1.2, tx: 2.5 },
    { date: '2025-06-19', rx: 0.9, tx: 1.1 },
    { date: '2025-06-21', rx: 0.7, tx: 0.8 },
    { date: '2025-06-22', rx: 1.0, tx: 1.3 },
    { date: '2025-06-23', rx: 0.5, tx: 0.6 },
    { date: '2025-06-25', rx: 0.3, tx: 0.4 }
    // { date: '2025-06-26', rx: 1.5, tx: 2.0 },
    // { date: '2025-06-27', rx: 2.1, tx: 3.2 },
    // { date: '2025-06-28', rx: 10, tx: 4.0 },
    // { date: '2025-06-22', rx: 1.0, tx: 1.3 },
    // { date: '2025-06-23', rx: 0.5, tx: 0.6 },
    // { date: '2025-06-25', rx: 0.3, tx: 0.4 },
    // { date: '2025-06-26', rx: 1.5, tx: 2.0 },
    // { date: '2025-06-27', rx: 2.1, tx: 3.2 },
    // { date: '2025-06-28', rx: 10, tx: 4.0 },
    // { date: '2025-06-28', rx: 10, tx: 4.0 },
    // { date: '2025-06-22', rx: 1.0, tx: 1.3 },
    // { date: '2025-06-23', rx: 0.5, tx: 0.6 },
    // { date: '2025-06-25', rx: 0.3, tx: 0.4 },
    // { date: '2025-06-26', rx: 1.5, tx: 2.0 },
    // { date: '2025-06-27', rx: 2.1, tx: 3.2 },
    // { date: '2025-06-28', rx: 10, tx: 4.0 }
];

const dummyOnlineUsers = <Array<ModelsOnlineUserSession>>[
    {
        Username: 'masoud1',
        Groupname: '(none)',
        'Average RX': '12.3 kB/s',
        'Average TX': '1.2 kB/s',
        '_Last connected at': '20s'
    },
    {
        Username: 'jane_doe',
        Groupname: 'group_test',
        'Average RX': '34.6 kB/s',
        'Average TX': '5.7 kB/s',
        '_Last connected at': '65m:20s'
    },
    {
        Username: 'admin',
        Groupname: 'group_test2',
        'Average RX': '98.1 kB/s',
        'Average TX': '22.4 kB/s',
        '_Last connected at': '1h:30m:40s'
    },
    {
        Username: 'Tester 1',
        Groupname: 'group_test2',
        'Average RX': '98.1 kB/s',
        'Average TX': '22.4 kB/s',
        '_Last connected at': '1h:30m:40s'
    },
    {
        Username: 'Tester 2',
        Groupname: 'group_test2',
        'Average RX': '98.1 kB/s',
        'Average TX': '22.4 kB/s',
        '_Last connected at': '1h:30m:40s'
    }
];

const dummyBanIPs = <Array<ModelsIPBanPoints>>[
    {
        IP: '172.17.0.1',
        Since: '2025-06-28 18:26',
        _Since: ' 4m:55s',
        Score: 80
    },
    {
        IP: '172.17.0.2',
        Since: '2025-06-28 18:26',
        _Since: ' 9m:55s',
        Score: 120
    },
    {
        IP: '172.17.0.3',
        Since: '2025-06-28 19:26',
        _Since: ' 10m:55s',
        Score: 160
    },
    {
        IP: '172.17.0.4',
        Since: '2025-06-29 23:26',
        _Since: ' 1h:10m:55s',
        Score: 220
    },
    {
        IP: '172.17.0.5',
        Since: '2025-06-31 23:26',
        _Since: ' 1h:10m:55s',
        Score: 32
    },
    {
        IP: '172.17.0.6',
        Since: '2025-06-29 23:26',
        _Since: ' 1h:10m:55s',
        Score: 190
    }
];

const dummyGroupConfig: ModelsOcservGroupConfig = {
    dns: ['8.8.8.8', '1.1.1.1'],
    nbns: '192.168.1.1',
    'ipv4-network': '192.168.1.0/24',
    'rx-data-per-sec': 100000,
    'tx-data-per-sec': 200000,
    'explicit-ipv4': '192.168.100.10',
    cgroup: 'cpuset,cpu:test',
    iroute: '10.0.0.0/8',
    route: ['0.0.0.0/0', '10.10.0.0/16'],
    'no-route': ['192.168.0.0/16', '10.0.0.0/8'],
    'net-priority': 1,
    'deny-roaming': true,
    'no-udp': false,
    keepalive: 60,
    dpd: 90,
    'mobile-dpd': 300,
    'max-same-clients': 2,
    'tunnel-all-dns': true,
    'stats-report-time': 300,
    mtu: 1400,
    'idle-timeout': 600,
    'mobile-idle-timeout': 900,
    'restrict-user-to-routes': true,
    'restrict-user-to-ports': 'tcp(443),tcp(80),udp(53)',
    'split-dns': ['example.com', 'internal.company.com'],
    'session-timeout': 3600
};

const dummyGroupList: ModelsOcservGroup[] = [
    { id: 1, name: 'Anc 1234', config: { mtu: 1330 }, owner: 'masoud' },
    { id: 2, name: 'Anc 4568', owner: 'masoud' },
    { id: 3, name: 'Anc 1248', owner: 'masoud' },
    { id: 4, name: 'Anc 1298', owner: 'masoud' }
];

const dummyHomeDockerService: HomeDockerService = {
    log_stream: {
        name: 'log_stream',
        cpu: {
            avg_percent: 12.5,
            total: 2000,
            used_units: 250
        },
        ram: {
            total: 4096,
            used: 1024,
            used_percent: 25
        }
    },
    ocserv: {
        name: 'ocserv',
        cpu: {
            avg_percent: 8.3,
            total: 2000,
            used_units: 166
        },
        ram: {
            total: 2048,
            used: 512,
            used_percent: 25
        }
    },
    postgres: {
        name: 'postgres',
        cpu: {
            avg_percent: 15.7,
            total: 4000,
            used_units: 628
        },
        ram: {
            total: 8192,
            used: 4096,
            used_percent: 50
        }
    },
    user_expiry: {
        name: 'user_expiry',
        cpu: {
            avg_percent: 2.1,
            total: 1000,
            used_units: 21
        },
        ram: {
            total: 1024,
            used: 256,
            used_percent: 25
        }
    },
    web: {
        name: 'web',
        cpu: {
            avg_percent: 25.0,
            total: 4000,
            used_units: 1000
        },
        ram: {
            total: 4096,
            used: 2048,
            used_percent: 50
        }
    }
};


const dummyRepositoryTotalBandwidths: RepositoryTotalBandwidths = {
    rx: 1250000000,
    tx: 875000000
};

const dummyRepositoryTopBandwidthUsers: RepositoryTopBandwidthUsers =     {
        top_rx: [
            {
                uid: 'u1',
                username: 'alice',
                password: 'pass123',
                group: 'premium',
                owner: 'system',
                is_locked: false,
                is_online: true,
                rx: 5200000000,
                tx: 1200000000,
                traffic_size: 10000000000,
                traffic_type: 'monthly' as ModelsOcservUserTrafficTypeEnum,
                created_at: '2026-01-10T10:00:00Z',
                updated_at: '2026-06-01T12:00:00Z',
                certificate_available: true,
                certificate_enabled: true,
                online_sessions: []
            },
            {
                uid: 'u2',
                username: 'bob',
                password: 'pass123',
                group: 'basic',
                owner: 'system',
                is_locked: false,
                is_online: false,
                rx: 3100000000,
                tx: 900000000,
                traffic_size: 8000000000,
                traffic_type: 'monthly' as ModelsOcservUserTrafficTypeEnum,
                created_at: '2026-02-15T09:00:00Z',
                online_sessions: []
            }
        ],
        top_tx: [
            {
                uid: 'u3',
                username: 'charlie',
                password: 'pass123',
                group: 'premium',
                owner: 'system',
                is_locked: false,
                is_online: true,
                rx: 2000000000,
                tx: 7800000000,
                traffic_size: 12000000000,
                traffic_type: 'monthly' as ModelsOcservUserTrafficTypeEnum,
                created_at: '2026-03-01T08:00:00Z',
                certificate_available: true,
                certificate_enabled: true,
                online_sessions: []
            },
            {
                uid: 'u4',
                username: 'david',
                password: 'pass123',
                group: 'basic',
                owner: 'system',
                is_locked: false,
                is_online: true,
                rx: 1500000000,
                tx: 5400000000,
                traffic_size: 9000000000,
                traffic_type: 'monthly' as ModelsOcservUserTrafficTypeEnum,
                created_at: '2026-04-20T11:30:00Z',
                online_sessions: []
            }
        ]
};

const dummyHomeGetHomeUser: HomeGetHomeUser = {
    total: 4,
    online_users_session: [
        {
            ID: 1,
            Username: 'alice',
            Groupname: 'premium',
            Device: 'Windows 11 - Chrome',
            IPv4: '192.168.1.10',
            vhost: 'vpn-1',
            'Session started at': '2026-06-06T10:15:00Z',
            '_Last connected at': '2026-06-06T12:40:00Z',
            'Average RX': '12.5 MB/s',
            'Average TX': '3.2 MB/s'
        },
        {
            ID: 2,
            Username: 'bob',
            Groupname: 'basic',
            Device: 'Android - OpenConnect',
            IPv4: '192.168.1.11',
            vhost: 'vpn-1',
            'Session started at': '2026-06-06T09:05:00Z',
            '_Last connected at': '2026-06-06T12:30:00Z',
            'Average RX': '8.1 MB/s',
            'Average TX': '2.4 MB/s'
        },
        {
            ID: 3,
            Username: 'charlie',
            Groupname: 'premium',
            Device: 'macOS - Safari',
            IPv4: '192.168.1.12',
            vhost: 'vpn-2',
            'Session started at': '2026-06-06T08:45:00Z',
            '_Last connected at': '2026-06-06T12:25:00Z',
            'Average RX': '15.0 MB/s',
            'Average TX': '4.0 MB/s'
        },
        {
            ID: 4,
            Username: 'david',
            Groupname: 'basic',
            Device: 'Linux - Firefox',
            IPv4: '192.168.1.13',
            vhost: 'vpn-2',
            'Session started at': '2026-06-06T07:30:00Z',
            '_Last connected at': '2026-06-06T12:10:00Z',
            'Average RX': '5.2 MB/s',
            'Average TX': '1.1 MB/s'
        }
    ]
};

export {
    dummyTrafficData,
    dummyOnlineUsers,
    dummyBanIPs,
    dummyGroupConfig,
    dummyGroupList,
    dummyHomeDockerService,
    dummyRepositoryTotalBandwidths,
    dummyRepositoryTopBandwidthUsers,
    dummyHomeGetHomeUser
};
