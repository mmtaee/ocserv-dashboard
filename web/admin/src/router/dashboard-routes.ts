import type { LucideIcon } from "@lucide/vue";
import {
  Activity,
  ArchiveRestore,
  ChartNoAxesCombined,
  DatabaseBackup,
  Gauge,
  History,
  LayoutDashboard,
  Network,
  Package,
  Radio,
  Router,
  Settings,
  SlidersHorizontal,
  Terminal,
  UserCog,
  Users,
  Wrench,
} from "@lucide/vue";

export interface DashboardRoute {
  path: string;
  alias?: string;
  name: string;
  titleKey: string;
  sectionKey: string;
  icon: LucideIcon;
  adminVisible: boolean;
  telegramOnly?: boolean;
}

export const dashboardRoutes: DashboardRoute[] = [
  {
    path: "/",
    name: "home",
    titleKey: "navigation.dashboard",
    sectionKey: "navigation.home",
    icon: LayoutDashboard,
    adminVisible: true,
  },
  {
    path: "/ocserv/management/runtime",
    name: "ocserv-runtime",
    titleKey: "navigation.runtime",
    sectionKey: "navigation.ocserv",
    icon: Wrench,
    adminVisible: false,
  },
  {
    path: "/ocserv/management/groups/defaults",
    name: "ocserv-group-defaults",
    titleKey: "navigation.groupDefaults",
    sectionKey: "navigation.ocserv",
    icon: SlidersHorizontal,
    adminVisible: false,
  },
  {
    path: "/ocserv/management/groups",
    name: "ocserv-groups",
    titleKey: "navigation.ocservGroups",
    sectionKey: "navigation.ocserv",
    icon: Network,
    adminVisible: true,
  },
  {
    path: "/ocserv/management/users",
    name: "ocserv-users",
    titleKey: "navigation.ocservUsers",
    sectionKey: "navigation.ocserv",
    icon: Users,
    adminVisible: true,
  },
  {
    path: "/ocserv/occtl",
    name: "occtl",
    titleKey: "navigation.occtl",
    sectionKey: "navigation.ocserv",
    icon: Terminal,
    adminVisible: false,
  },
  {
    path: "/ocserv/management/ocserv/sync",
    name: "ocserv-sync",
    titleKey: "navigation.sync",
    sectionKey: "navigation.ocserv",
    icon: ArchiveRestore,
    adminVisible: false,
  },
  {
    path: "/ocserv/management/statistics",
    name: "statistics",
    titleKey: "navigation.statistics",
    sectionKey: "navigation.statistics",
    icon: ChartNoAxesCombined,
    adminVisible: true,
  },
  {
    path: "/ocserv/management/bandwidths",
    name: "bandwidths",
    titleKey: "navigation.bandwidths",
    sectionKey: "navigation.statistics",
    icon: Gauge,
    adminVisible: true,
  },
  {
    path: "/ocserv/management/session_logs",
    name: "session-logs",
    titleKey: "navigation.sessionLogs",
    sectionKey: "navigation.statistics",
    icon: History,
    adminVisible: true,
  },
  {
    path: "/staffs",
    name: "staffs",
    titleKey: "navigation.staffs",
    sectionKey: "navigation.staffs",
    icon: UserCog,
    adminVisible: false,
  },
  {
    path: "/staffs/activities",
    name: "staff-activities",
    titleKey: "navigation.activities",
    sectionKey: "navigation.staffs",
    icon: Activity,
    adminVisible: false,
  },
  {
    path: "/telegram/requests",
    name: "telegram-requests",
    titleKey: "navigation.telegramRequests",
    sectionKey: "navigation.telegram",
    icon: Radio,
    adminVisible: false,
    telegramOnly: true,
  },
  {
    path: "/telegram/packages",
    name: "telegram-packages",
    titleKey: "navigation.telegramPackages",
    sectionKey: "navigation.telegram",
    icon: Package,
    adminVisible: false,
    telegramOnly: true,
  },
  {
    path: "/telegram/settings",
    name: "telegram-settings",
    titleKey: "navigation.telegramSettings",
    sectionKey: "navigation.telegram",
    icon: Settings,
    adminVisible: false,
    telegramOnly: true,
  },
  {
    path: "/system",
    alias: "/system-settings",
    name: "system-settings",
    titleKey: "navigation.settings",
    sectionKey: "navigation.system",
    icon: Router,
    adminVisible: false,
  },
  {
    path: "/backup",
    name: "backup",
    titleKey: "navigation.backup",
    sectionKey: "navigation.system",
    icon: DatabaseBackup,
    adminVisible: false,
  },
];
