import { apiRequest } from '@/lib/api/client'

export interface HealthData {
  status: string
  service: string
}

/**
 * 检查 Vue 前端能否正常连接 Go API。
 */
export function getHealth(): Promise<HealthData> {
  return apiRequest<HealthData>('/health')
}