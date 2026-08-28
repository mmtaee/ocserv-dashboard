import { requireAuthorizationHeader } from '@/api/auth-token'
import { api } from '@/api/client'

function authorization() {
  return { authorization: requireAuthorizationHeader() }
}

export async function getDashboardOverview() {
  const response = await api.home.homeGet(authorization())
  return response.data
}

export async function getContainerStats() {
  const response = await api.home.homeContainerStatsGet(authorization())
  return response.data
}

export async function getOcservStats() {
  const response = await api.home.homeOcservStatsGet(authorization())
  return response.data
}

export async function getSystemStats() {
  const response = await api.home.homeSystemStatsGet(authorization())
  return response.data
}
