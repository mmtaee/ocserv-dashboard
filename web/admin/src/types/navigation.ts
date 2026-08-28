import type { RouteLocationRaw } from 'vue-router'

export interface AdminNavigationItem {
  titleKey: string
  icon: string
  to: RouteLocationRaw
}
