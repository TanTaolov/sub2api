import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import PrismStatusView from '../PrismStatusView.vue'
import { getPrismStatus } from '@/api/admin/prismStatus'
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<main><slot /></main>' } }))
vi.mock('@/components/admin/operations/SmartOpsNav.vue', () => ({ default: { template: '<nav />' } }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/api/admin/prismStatus', () => ({ getPrismStatus: vi.fn() }))
const status = (patch: Record<string, unknown> = {}) => ({
  enabled: true,
  base_url: 'http://127.0.0.1:8319/v1',
  endpoint: 'http://127.0.0.1:8319/v1/responses',
  api_key_configured: true,
  models: ['gpt-6.1-sol', 'gpt-5.6-sol', 'gpt-5.6-terra', 'gpt-6-luna'],
  state: 'ready',
  ...patch
})
const create = () => mount(PrismStatusView, { global: { plugins: [createPinia()] } })
beforeEach(() => { vi.resetAllMocks(); vi.mocked(getPrismStatus).mockResolvedValue(status() as any) })
describe('prism gateway status', () => {
  it('shows the effective endpoint, bridge key state and model scope', async () => {
    const wrapper = create(); await flushPromises()
    expect(wrapper.get('[data-testid="prism-state"]').text()).toContain('prismStatus.states.ready')
    expect(wrapper.get('[data-testid="prism-endpoint"]').text()).toContain('http://127.0.0.1:8319/v1/responses')
    expect(wrapper.get('[data-testid="prism-api-key"]').text()).toContain('prismStatus.apiKeyConfigured')
    expect(wrapper.get('[data-testid="prism-models"]').text()).toContain('gpt-6.1-sol')
    wrapper.unmount()
  })
  it('flags a missing bridge key without inventing an endpoint', async () => {
    vi.mocked(getPrismStatus).mockResolvedValue(status({ api_key_configured: false, endpoint: '', state: 'key_missing' }) as any)
    const wrapper = create(); await flushPromises()
    expect(wrapper.get('[data-testid="prism-state"]').text()).toContain('prismStatus.states.key_missing')
    expect(wrapper.get('[data-testid="prism-endpoint"]').text()).toContain('prismStatus.endpointUnavailable')
    expect(wrapper.get('[data-testid="prism-api-key"]').text()).toContain('prismStatus.apiKeyMissing')
    wrapper.unmount()
  })
  it('reports a load failure instead of a false ready state', async () => {
    vi.mocked(getPrismStatus).mockRejectedValue(new Error('boom'))
    const wrapper = create(); await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('boom')
    expect(wrapper.find('[data-testid="prism-state"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="prism-models"]').exists()).toBe(false)
    wrapper.unmount()
  })
})
