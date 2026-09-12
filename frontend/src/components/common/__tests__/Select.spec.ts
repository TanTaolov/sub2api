import { mount, type VueWrapper } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ElSelect, ElOption } from 'element-plus'
import Select from '../Select.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
const wrappers: VueWrapper[] = []
afterEach(() => { wrappers.forEach(wrapper => wrapper.unmount()); wrappers.length = 0; vi.useRealTimers() })
function createSelect(props: Record<string, unknown> = {}) {
  const wrapper = mount(Select, { props: { modelValue: null, options: [{ value: 1, label: 'Number' }, { value: '1', label: 'String' }, { value: null, label: 'None' }], ...props } })
  wrappers.push(wrapper)
  return wrapper
}

describe('Select', () => {
  it('区分数字、字符串与空值选项，并返回原始选项对象', () => {
    const wrapper = createSelect()
    wrapper.getComponent(ElSelect).vm.$emit('change', 'number:1')
    wrapper.getComponent(ElSelect).vm.$emit('change', 'string:1')
    wrapper.getComponent(ElSelect).vm.$emit('change', 'object:null')
    expect(wrapper.emitted('update:modelValue')).toEqual([[1], ['1'], [null]])
    expect(wrapper.emitted('change')?.[0]).toEqual([1, { value: 1, label: 'Number' }])
  })
  it('清空仅发出一次变更', () => {
    const wrapper = createSelect({ modelValue: 1, clearable: true })
    wrapper.getComponent(ElSelect).vm.$emit('change', undefined)
    wrapper.getComponent(ElSelect).vm.$emit('clear')
    expect(wrapper.emitted('update:modelValue')).toEqual([[null]])
    expect(wrapper.emitted('change')).toEqual([[null, null]])
  })
  it('使用自定义 valueKey 与 labelKey', () => {
    const wrapper = createSelect({ options: [{ id: 3, name: 'Three' }], valueKey: 'id', labelKey: 'name' })
    wrapper.getComponent(ElSelect).vm.$emit('change', 'number:3')
    expect(wrapper.emitted('change')?.[0]).toEqual([3, { id: 3, name: 'Three' }])
  })
  it('本地搜索过滤选项并保留原始模型值', async () => {
    const wrapper = createSelect({ searchable: true, modelValue: 1 })
    wrapper.getComponent(ElSelect).props('filterMethod')?.('String')
    await wrapper.vm.$nextTick()
    const options = wrapper.findAllComponents(ElOption).filter(option => option.attributes('hidden') === undefined)
    expect(options.map(option => option.props('label'))).toEqual(['String'])
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })
  it('远程搜索防抖，关闭后取消待发送请求', async () => {
    vi.useFakeTimers()
    const wrapper = createSelect({ remote: true })
    const select = wrapper.getComponent(ElSelect)
    select.vm.$emit('visible-change', true)
    select.props('filterMethod')?.('first')
    select.props('filterMethod')?.('latest')
    await vi.advanceTimersByTimeAsync(300)
    expect(wrapper.emitted('search')).toEqual([['latest']])
    select.props('filterMethod')?.('cancelled')
    select.vm.$emit('visible-change', false)
    await vi.advanceTimersByTimeAsync(300)
    expect(wrapper.emitted('search')).toHaveLength(1)
  })
  it('创建选项不重复添加已有值', async () => {
    const wrapper = createSelect({ creatable: true, options: [{ value: 'one', label: 'One' }] })
    wrapper.getComponent(ElSelect).props('filterMethod')?.('one')
    await wrapper.vm.$nextTick()
    expect(wrapper.findAllComponents(ElOption).filter(option => option.props('value') === 'string:one')).toHaveLength(1)
  })
})
