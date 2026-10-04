import { apiClient } from '../client'

// Gateway-side Prism adapter contract. The bridge key is reported as a boolean
// only: the adapter secret is never sent to the browser.
export type PrismState = 'disabled' | 'endpoint_invalid' | 'key_missing' | 'ready'

export interface PrismStatus {
  enabled: boolean
  base_url: string
  endpoint: string
  api_key_configured: boolean
  models: string[]
  state: PrismState
}

export async function getPrismStatus(): Promise<PrismStatus> {
  const { data } = await apiClient.get<PrismStatus>('/admin/account-ops/prism/status')
  return { ...data, models: data.models ?? [] }
}
