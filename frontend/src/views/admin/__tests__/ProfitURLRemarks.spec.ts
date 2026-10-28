vi.mock('@/api/admin/upstreamChannels', () => ({ listUpstreamChannels: vi.fn(), getUpstreamChannel: vi.fn(), saveUpstreamChannel: vi.fn(), deleteUpstreamChannel: vi.fn(), getChannelMembers: vi.fn(), bindChannelAccounts: vi.fn(), readChannelBindings: vi.fn() }))
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProfitOperationsView from '../ProfitOperationsView.vue'
import { readProfitOperations, readUpstreamURLRemark, saveUpstreamURLRemark, type OperationsSnapshot } from '@/api/admin/profitOperations'
import zh from '@/i18n/locales/zh/profitOps'
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<main><slot /></main>' } }))
vi.mock('@/components/admin/operations/SmartOpsNav.vue', () => ({ default: { template: '<nav />' } }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => (zh as Record<string, string>)[key.replace('profitOps.', '')] || key }) }))
vi.mock('@/api/admin/profitOperations', () => ({ readProfitOperations: vi.fn(), readUpstreamURLRemark: vi.fn(), saveUpstreamURLRemark: vi.fn() }))
const upstream = 'https://relay.example:8443/tenant/v1'
const value = (remark = '', revision = 0) => ({ upstream, remark, revision })
const fixture = (remark = '', revision = 0) => ({
 scope: { dimension: 'upstream' }, queried_at: '', currency: 'USD',
 url_remarks: { [upstream]: value(remark, revision) },
 usage: { summary: {}, rows: [{ key: upstream, name: upstream }], total: 1, accounts: [], catalog: [{ key: upstream, name: upstream, accounts: 2 }], catalog_total: 1 },
 quality: { status: 'ready', data: { summary: {}, rows: [] } }, errors: { status: 'ready', data: { rows: [] } }
}) as unknown as OperationsSnapshot
beforeEach(() => {
 vi.resetAllMocks()
 vi.mocked(readProfitOperations).mockResolvedValue(fixture('共享备注', 2))
 vi.mocked(saveUpstreamURLRemark).mockResolvedValue(value('新备注', 3))
 vi.mocked(readUpstreamURLRemark).mockResolvedValue(value('新备注', 3))
})
describe('profit-operations URL channel remarks', () => {
 it('edits a shared URL in the exact profit page, cancels, saves through API and reads back before success, then survives remount', async () => {
  let w = mount(ProfitOperationsView); await flushPromises()
  expect(w.text()).toContain('URL 渠道备注')
  expect(w.findAll('[data-testid="url-remark-editor"]')).toHaveLength(1)
  expect(w.text()).toContain('共享备注')
  await w.get('[data-testid="edit-url-remark"]').trigger('click')
  await w.get('textarea').setValue('取消的草稿')
  await w.get('[data-testid="cancel-url-remark"]').trigger('click')
  expect(saveUpstreamURLRemark).not.toHaveBeenCalled()
  await w.get('[data-testid="edit-url-remark"]').trigger('click')
  expect(w.get('textarea').element.value).toBe('共享备注')
  await w.get('textarea').setValue('新备注')
  await w.get('[data-testid="save-url-remark"]').trigger('click'); await flushPromises()
  expect(saveUpstreamURLRemark).toHaveBeenCalledWith({ upstream, remark: '新备注', revision: 2 })
  expect(readUpstreamURLRemark).toHaveBeenCalledWith(upstream)
  expect(w.find('textarea').exists()).toBe(false)
  vi.mocked(readProfitOperations).mockResolvedValue(fixture('新备注', 3))
  await w.get('[data-testid="refresh-operations"]').trigger('click'); await flushPromises()
  expect(w.text()).toContain('新备注'); w.unmount()
  w = mount(ProfitOperationsView); await flushPromises(); expect(w.text()).toContain('新备注'); w.unmount()
 })
 it('preserves drafts on write failure and conflict; refreshes the version explicitly without replacing draft', async () => {
  const w = mount(ProfitOperationsView); await flushPromises()
  await w.get('[data-testid="edit-url-remark"]').trigger('click')
  await w.get('textarea').setValue('保留的草稿')
  vi.mocked(saveUpstreamURLRemark).mockRejectedValueOnce({ status: 409, reason: 'UPSTREAM_URL_REMARK_CONFLICT' })
  await w.get('[data-testid="save-url-remark"]').trigger('click'); await flushPromises()
  expect(w.get('textarea').element.value).toBe('保留的草稿')
  expect(w.get('[data-testid="url-remark-error"]').text()).toContain('其他管理员')
  vi.mocked(readUpstreamURLRemark).mockResolvedValue(value('其他人的备注', 4))
  await w.get('[data-testid="reload-url-remark"]').trigger('click'); await flushPromises()
  expect(w.get('textarea').element.value).toBe('保留的草稿')
  expect(w.text()).toContain('其他人的备注')
  vi.mocked(saveUpstreamURLRemark).mockRejectedValueOnce(new Error('private SQL secret'))
  await w.get('[data-testid="save-url-remark"]').trigger('click'); await flushPromises()
  expect(saveUpstreamURLRemark).toHaveBeenLastCalledWith({ upstream, remark: '保留的草稿', revision: 4 })
  expect(w.get('textarea').element.value).toBe('保留的草稿')
  expect(w.text()).not.toContain('private SQL secret'); w.unmount()
 })
 it('requires real GET readback, retaining a draft when verification fails', async () => {
  const w = mount(ProfitOperationsView); await flushPromises()
  await w.get('[data-testid="edit-url-remark"]').trigger('click'); await w.get('textarea').setValue('新备注')
  vi.mocked(readUpstreamURLRemark).mockRejectedValueOnce(new Error('offline'))
  await w.get('[data-testid="save-url-remark"]').trigger('click'); await flushPromises()
  expect(w.get('textarea').element.value).toBe('新备注')
  expect(w.get('[data-testid="url-remark-error"]').text()).toContain('读回')
  w.unmount()
 })
 it('validates Unicode length and clears by an explicit empty save', async () => {
  const w = mount(ProfitOperationsView); await flushPromises()
  await w.get('[data-testid="edit-url-remark"]').trigger('click')
  await w.get('textarea').setValue('中'.repeat(501))
  expect(w.get('[data-testid="save-url-remark"]').attributes('disabled')).toBeDefined()
  await w.get('textarea').setValue('😀'.repeat(500))
  expect(w.get('[data-testid="save-url-remark"]').attributes('disabled')).toBeUndefined()
  await w.get('textarea').setValue('  ')
  vi.mocked(saveUpstreamURLRemark).mockResolvedValue(value('', 3)); vi.mocked(readUpstreamURLRemark).mockResolvedValue(value('', 3))
  await w.get('[data-testid="save-url-remark"]').trigger('click'); await flushPromises()
  expect(saveUpstreamURLRemark).toHaveBeenCalledWith({ upstream, remark: '', revision: 2 })
  expect(w.text()).toContain('未填写备注'); w.unmount()
 })
 it('does not offer editing for unknown attribution buckets', async () => {
  const f = fixture(); f.url_remarks = {}; f.usage.catalog[0]!.key = 'unknown:invalid_url'; f.usage.rows[0]!.key = 'unknown:invalid_url'
  vi.mocked(readProfitOperations).mockResolvedValue(f)
  const w = mount(ProfitOperationsView); await flushPromises()
  expect(w.find('[data-testid="edit-url-remark"]').exists()).toBe(false); w.unmount()
 })
})
