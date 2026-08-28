import { getAccessToken } from '@/api/auth-token'
import { api } from '@/api/client'
import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemInitResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemPatchSystemUpdateData,
} from '@/api/generated'

export type SystemInitConfig =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemInitResponse

export type SystemUpdateData =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemPatchSystemUpdateData

export async function getSystemInit(): Promise<SystemInitConfig | null> {
  const response = await api.system.systemInitGet()
  return response.data ?? null
}

export async function updateSystemConfig(request: SystemUpdateData) {
  const token = getAccessToken()
  const authorization = token ? (token.startsWith('Bearer ') ? token : `Bearer ${token}`) : ''
  const response = await api.system.systemPatch({ authorization, request })
  return response.data
}
