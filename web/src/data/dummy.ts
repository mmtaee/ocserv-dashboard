import type {
    HomeDockerService,
    ModelsDailyTraffic,
    ModelsIPBanPoints,
    ModelsOcservGroup,
    ModelsOcservGroupConfig,
    ModelsOnlineUserSession
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
        '_Connected at': '20s'
    },
    {
        Username: 'jane_doe',
        Groupname: 'group_test',
        'Average RX': '34.6 kB/s',
        'Average TX': '5.7 kB/s',
        '_Connected at': '65m:20s'
    },
    {
        Username: 'admin',
        Groupname: 'group_test2',
        'Average RX': '98.1 kB/s',
        'Average TX': '22.4 kB/s',
        '_Connected at': '1h:30m:40s'
    },
    {
        Username: 'Tester 1',
        Groupname: 'group_test2',
        'Average RX': '98.1 kB/s',
        'Average TX': '22.4 kB/s',
        '_Connected at': '1h:30m:40s'
    },
    {
        Username: 'Tester 2',
        Groupname: 'group_test2',
        'Average RX': '98.1 kB/s',
        'Average TX': '22.4 kB/s',
        '_Connected at': '1h:30m:40s'
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

const mockHomeDockerService: HomeDockerService = {
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

export { dummyTrafficData, dummyOnlineUsers, dummyBanIPs, dummyGroupConfig, dummyGroupList, mockHomeDockerService };
