import { mount, type VueWrapper } from '@vue/test-utils'
import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import DataTable from '../DataTable.vue'
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
const wrappers: VueWrapper[] = []
function viewport(desktop: boolean) {
  window.matchMedia = vi.fn(query => ({ matches: desktop, media: query, onchange: null, addEventListener: vi.fn(), removeEventListener: vi.fn(), addListener: vi.fn(), removeListener: vi.fn(), dispatchEvent: vi.fn() }))
}
beforeEach(() => { viewport(true); localStorage.clear() })
afterEach(() => { wrappers.forEach(wrapper => wrapper.unmount()); wrappers.length = 0 })
function createTable(props: Record<string, unknown> = {}, slots = {}) {
  const wrapper = mount(DataTable, {
    props: { columns: [{ key: 'name', label: 'Name', sortable: true }], data: [{ id: 1, name: 'Beta' }, { id: 2, name: 'Alpha' }], ...props },
    slots,
    global: { stubs: { ElAutoResizer: defineComponent({ setup: (_, { slots }) => () => h('div', slots.default?.({ width: 900, height: 500 })) }) }
  })
  wrappers.push(wrapper)
  return wrapper
}
const rowNames = (wrapper: VueWrapper) => wrapper.findAll('.el-table__body .el-table__row').map(row => row.text())

describe('DataTable', () => {
  it('小数据集完整显示，点击表头切换稳定排序', async () => {
    const wrapper = createTable({ defaultSortKey: 'name' }, { 'header-name': '<span>Name</span>' })
    await wrapper.vm.$nextTick()
    expect(rowNames(wrapper)).toEqual(['Alpha', 'Beta'])
    await wrapper.get('button[aria-label="Name"]').trigger('click')
    expect(rowNames(wrapper)).toEqual(['Beta', 'Alpha'])
  })
  it('服务端排序只发出事件，不重新排列服务端返回的数据', async () => {
    const wrapper = createTable({ serverSideSort: true })
    await wrapper.vm.$nextTick()
    await wrapper.get('button[aria-label="Name"]').trigger('click')
    expect(wrapper.emitted('sort')).toEqual([['name', 'asc']])
    expect(rowNames(wrapper)).toEqual(['Beta', 'Alpha'])
  })
  it('保存并恢复指定表格的排序偏好', async () => {
    const wrapper = createTable({ sortStorageKey: 'table-preference' })
    await wrapper.vm.$nextTick()
    await wrapper.get('button[aria-label="Name"]').trigger('click')
    await wrapper.get('button[aria-label="Name"]').trigger('click')
    const restored = createTable({ sortStorageKey: 'table-preference' })
    await restored.vm.$nextTick()
    expect(rowNames(restored)).toEqual(['Beta', 'Alpha'])
    expect(restored.get('button[aria-label="Name"]').attributes('aria-sort')).toBe('descending')
  })
  it('当前页全选保留其他页的选中键，取消仅移除本页键', async () => {
    const wrapper = createTable({ selectable: true, selectedKeys: [99] })
    await wrapper.vm.$nextTick()
    await wrapper.get('[data-test="select-all"] input').setValue(true)
    expect(wrapper.emitted('update:selectedKeys')?.[0]).toEqual([[99, 1, 2]])
    await wrapper.setProps({ selectedKeys: [99, 1, 2] })
    await wrapper.get('[data-test="select-all"] input').setValue(false)
    expect(wrapper.emitted('update:selectedKeys')?.[1]).toEqual([[99]])
  })
  it('选择复选框不会触发行点击', async () => {
    const wrapper = createTable({ selectable: true, clickableRows: true })
    await wrapper.vm.$nextTick()
    await wrapper.get('[data-test="select-row"] input').trigger('click')
    expect(wrapper.emitted('rowClick')).toBeUndefined()
    await wrapper.get('.el-table__row').trigger('click')
    expect(wrapper.emitted('rowClick')?.[0]).toEqual([{ id: 1, name: 'Beta' }])
  })
  it('大列表采用虚拟表格，分页替换后不残留上一页数据', async () => {
    const wrapper = createTable({ data: Array.from({ length: 200 }, (_, index) => ({ id: index, name: 'Old ' + index })) })
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.el-table-v2').exists()).toBe(true)
    expect(wrapper.findAll('.el-table-v2__row').length).toBeLessThan(200)
    await wrapper.setProps({ data: Array.from({ length: 200 }, (_, index) => ({ id: index + 200, name: 'New ' + index })) })
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).not.toContain('Old ')
    expect(wrapper.text()).toContain('New 0')
  })
  it('移动端卡片保留单元格插槽和当前页多选', async () => {
    viewport(false)
    const wrapper = createTable({ selectable: true, selectedKeys: [99] }, { 'cell-name': '<span data-test="mobile-name">{{ params.row.name }}</span>' })
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.el-table').exists()).toBe(false)
    expect(wrapper.findAll('[data-test="mobile-name"]')).toHaveLength(2)
    await wrapper.get('[data-test="select-all-mobile"] input').setValue(true)
    expect(wrapper.emitted('update:selectedKeys')?.[0]).toEqual([[99, 1, 2]])
  })
})
