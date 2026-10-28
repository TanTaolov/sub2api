import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import ProfitOperationsView from '../ProfitOperationsView.vue'
import { readProfitOperations, type OperationsSnapshot } from '@/api/admin/profitOperations'
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<main><slot /></main>' } }))
vi.mock('@/components/admin/operations/SmartOpsNav.vue', () => ({ default: { template: '<nav />' } }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/api/admin/profitOperations', () => ({ readProfitOperations: vi.fn() }))
vi.mock('@/api/admin/upstreamChannels', () => ({ listUpstreamChannels: vi.fn().mockResolvedValue({ items: [], total: 0 }) }))
beforeEach(() => vi.mocked(readProfitOperations).mockResolvedValue({ usage: { catalog: [], rows: [{ key: 'channel:7', name: '渠道A', assignment: 'explicit', channel_id: 7, channel_remark: '统一采购' }], accounts: [], summary: {} }, scope: {}, quality: { status: 'ready', data: { summary: {}, rows: [] } }, errors: { status: 'ready', data: { rows: [] } } } as unknown as OperationsSnapshot))
it('offers actual entity management and marks explicit ownership in the exact profit page', async () => {
 const w = mount(ProfitOperationsView); await flushPromises()
 expect(w.find('[data-testid="upstream-channel-manager"]').exists()).toBe(true)
 expect(w.text()).toContain('upstreamEntities.explicit')
 expect(w.text()).toContain('统一采购')
 w.unmount()
})
