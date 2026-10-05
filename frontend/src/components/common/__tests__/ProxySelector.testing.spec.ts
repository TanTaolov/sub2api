import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import ProxySelector from '../ProxySelector.vue'
import type { Proxy } from '@/types'

const testProxy = vi.hoisted(() => vi.fn())
vi.mock('@/api/admin', () => ({ adminAPI: { proxies: { testProxy } } }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)
beforeEach(() => { vi.clearAllMocks() })

async function openSelector() {
  const wrapper = mount(ProxySelector, {
    props: { modelValue: null, proxies: [1, 2].map(id => ({
      id, name: `Proxy ${id}`, host: 'localhost', port: 8080, protocol: 'http'
    } as Proxy)) },
    global: { stubs: { Icon: true } }
  })
  await wrapper.get('.select-trigger').trigger('click')
  return wrapper
}

describe('proxy connection tests', () => {
  it('does not restart an individual test when a batch is started', async () => {
    let finish!: (result: object) => void
    testProxy.mockImplementation((id: number) => id === 1
      ? new Promise(resolve => { finish = resolve })
      : Promise.resolve({ success: true, country: 'GB' }))
    const wrapper = await openSelector()
    await wrapper.findAll('.test-btn')[0].trigger('click')
    await wrapper.get('.batch-test-btn').trigger('click')
    await flushPromises()
    expect(testProxy.mock.calls.map(([id]) => id)).toEqual([1, 2])
    expect(wrapper.findAll('.test-btn')[0].attributes('disabled')).toBeDefined()
    finish({ success: true, country: 'US' })
    await flushPromises()
    expect(wrapper.text()).toContain('US')
    expect(wrapper.findAll('.test-btn')[0].attributes('disabled')).toBeUndefined()
  })

  it('shows per-proxy outcomes and allows another batch after a failure', async () => {
    testProxy.mockImplementation((id: number) => id === 1
      ? Promise.reject(new Error('offline'))
      : Promise.resolve({ success: true, country: 'GB' }))
    const wrapper = await openSelector()
    await wrapper.get('.batch-test-btn').trigger('click')
    await flushPromises()
    expect(testProxy).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('admin.proxies.testFailed')
    expect(wrapper.text()).toContain('GB')
    expect(wrapper.get('.batch-test-btn').attributes('disabled')).toBeUndefined()
    await wrapper.get('.batch-test-btn').trigger('click')
    await flushPromises()
    expect(testProxy).toHaveBeenCalledTimes(4)
  })
})

describe('random static proxy selection', () => {
  const proxy = (id: number, status: Proxy['status'], expires_at: string | null = null) => ({
    id, name: `Proxy ${id}`, host: 'localhost', port: 8080, protocol: 'http', status, expires_at
  } as Proxy)

  it('keeps the virtual option first and emits a real eligible ID', async () => {
    const wrapper = mount(ProxySelector, {
      props: { modelValue: null, proxies: [
        proxy(1, 'inactive'), proxy(2, 'active', '2020-01-01T00:00:00Z'), proxy(3, 'active')
      ] },
      global: { stubs: { Icon: true } }
    })
    await wrapper.get('.select-trigger').trigger('click')
    expect(wrapper.get('.select-options').element.firstElementChild?.getAttribute('data-testid')).toBe('random-proxy-option')
    await wrapper.get('input.select-search-input').setValue('not a proxy')
    expect(wrapper.get('[data-testid="random-proxy-option"]').exists()).toBe(true)
    await wrapper.get('[data-testid="random-proxy-option"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([3])
    expect(wrapper.emitted('update:randomMode')?.[0]).toEqual([true])
    await wrapper.setProps({ modelValue: 3, randomMode: true })
    expect(wrapper.get('.select-trigger').text()).toContain('Proxy 3')
    await wrapper.get('.select-trigger').trigger('click')
    expect(wrapper.get('[data-testid="random-proxy-option"]').classes()).toContain('select-option-selected')
    expect(wrapper.findAll('.select-option-selected')).toHaveLength(1)
  })

  it('never turns an empty eligible pool into no proxy', async () => {
    const wrapper = mount(ProxySelector, {
      props: { modelValue: null, proxies: [proxy(1, 'expired')] },
      global: { stubs: { Icon: true } }
    })
    await wrapper.get('.select-trigger').trigger('click')
    await wrapper.get('[data-testid="random-proxy-option"]').trigger('click')
    expect(wrapper.get('[role="alert"]').text()).toBe('admin.accounts.randomProxyEmpty')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('defers bulk allocation without emitting a fake proxy ID', async () => {
    const wrapper = mount(ProxySelector, {
      props: { modelValue: 4, randomMode: false, deferRandom: true, proxies: [] },
      global: { stubs: { Icon: true } }
    })
    await wrapper.get('.select-trigger').trigger('click')
    await wrapper.get('[data-testid="random-proxy-option"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([null])
    expect(wrapper.emitted('update:randomMode')?.[0]).toEqual([true])
    await wrapper.setProps({ modelValue: null, randomMode: true })
    await wrapper.get('.select-trigger').trigger('click')
    expect(wrapper.findAll('.select-option-selected')).toHaveLength(1)
    await wrapper.findAll('.select-option')[1].trigger('click')
    expect(wrapper.emitted('update:randomMode')?.[1]).toEqual([false])
  })
})
