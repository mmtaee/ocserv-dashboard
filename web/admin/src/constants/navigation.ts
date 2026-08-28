import type { AdminNavigationItem } from '@/types/navigation'

export const adminNavigation: AdminNavigationItem[] = [
  {
    titleKey: 'admin.navigation.overview',
    icon: 'mdi-view-grid-outline',
    to: { name: 'admin-overview' },
  },
  {
    titleKey: 'admin.navigation.dashboard',
    icon: 'mdi-view-dashboard-outline',
    to: { name: 'admin-dashboard' },
  },
]
