import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { apiClient } from '@/api/client'
import UpstreamURLRemarkEditor from '@/components/admin/operations/UpstreamURLRemarkEditor.vue'
import zh from '@/i18n/locales/zh/profitOps'

vi.mock('@/i18n', () => ({ getLocale: () => 'zh' }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => (zh as Record<string, string>)[key.replace('profitOps.', '')] || key }) }))
const originalAdapter = apiClient.defaults.adapter
afterEach(() => { apiClient.defaults.adapter = originalAdapter })

describe('URL remark editor with the actual API module and client interceptors', () => {
 it('handles real normalized 409 errors, preserves the draft, reloads its version, then PUTs and GETs the exact identity', async () => {
  const upstream = 'https://relay.example:8443/Case/%2F'
  let persisted = { upstream, remark: '其他人的备注', revision: 2 }
  let conflict = true
  const bodies: unknown[] = []
  const adapter = vi.fn(async (config: InternalAxiosRequestConfig) => {
   expect(config.url).toBe('/admin/profit-operations/url-remark')
   const response = { status: 200, statusText: 'OK', headers: {}, config, data: { code: 0, data: persisted } }
   if (config.method === 'put') {
    const body = JSON.parse(config.data as string); bodies.push(body)
    if (conflict) {
     conflict = false
     throw new AxiosError('HTTP 409', AxiosError.ERR_BAD_REQUEST, config, undefined, {
      ...response, status: 409, data: { code: 409, reason: 'UPSTREAM_URL_REMARK_CONFLICT', message: 'conflict' }
     })
    }
    persisted = { ...body, revision: body.revision + 1 }
    response.data = { code: 0, data: persisted }
   } else {
    expect(config.method).toBe('get')
    expect(config.params).toEqual(expect.objectContaining({ upstream }))
   }
   return response
  })
  apiClient.defaults.adapter = adapter
  const w = mount(UpstreamURLRemarkEditor, { props: { value: { upstream, remark: '旧备注', revision: 1 } } })
  await w.get('[data-testid="edit-url-remark"]').trigger('click')
  await w.get('textarea').setValue('保留草稿')
  await w.get('[data-testid="save-url-remark"]').trigger('click'); await flushPromises()
  expect(w.get('[data-testid="url-remark-error"]').text()).toContain('其他管理员')
  expect(w.get('textarea').element.value).toBe('保留草稿')
  expect(w.get('[data-testid="save-url-remark"]').attributes('disabled')).toBeDefined()
  await w.get('[data-testid="reload-url-remark"]').trigger('click'); await flushPromises()
  expect(w.text()).toContain('其他人的备注'); expect(w.get('textarea').element.value).toBe('保留草稿')
  await w.get('[data-testid="save-url-remark"]').trigger('click'); await flushPromises()
  expect(bodies).toEqual([{ upstream, remark: '保留草稿', revision: 1 }, { upstream, remark: '保留草稿', revision: 2 }])
  expect(adapter.mock.calls.map(([config]) => config.method)).toEqual(['put', 'get', 'put', 'get'])
  expect(w.find('textarea').exists()).toBe(false)
  expect(w.emitted('saved')?.at(-1)).toEqual([{ upstream, remark: '保留草稿', revision: 3 }])
  w.unmount()
 })
})
