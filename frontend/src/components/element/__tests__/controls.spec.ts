import { afterEach, describe, expect, it } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { defineComponent, ref } from 'vue'
import ElementInput from '../ElementInput.vue'
import ElementCheckbox from '../ElementCheckbox.vue'
import ElementRadio from '../ElementRadio.vue'
import ElementSelect from '../ElementSelect.vue'
import { ElSelect } from 'element-plus'

const wrappers: VueWrapper[] = []
afterEach(() => { wrappers.forEach(wrapper => wrapper.unmount()); wrappers.length = 0 })

describe('Element Plus 原生事件兼容', () => {
  it('无 v-model 的输入框保留连续输入，并转发实际输入元素', async () => {
    const wrapper = mount(ElementInput)
    wrappers.push(wrapper)
    const input = wrapper.get('input')
    await input.setValue('12')
    await input.setValue('123')
    expect(input.element.value).toBe('123')
    expect((wrapper.emitted('input')?.[1]?.[0] as Event).target).toBe(input.element)
  })

  it('数字输入保持 number 语义，清空后不错误转换为 0', async () => {
    const wrapper = mount(defineComponent({
      components: { ElementInput },
      setup: () => ({ amount: ref<number | string>(2) }),
      template: '<ElementInput v-model.number="amount" type="number" />'
    }))
    wrappers.push(wrapper)
    await wrapper.get('input').setValue('12.5')
    expect(wrapper.vm.amount).toBe(12.5)
    await wrapper.get('input').setValue('')
    expect(wrapper.vm.amount).toBe('')
  })

  it('lazy 输入在 change 之前保持编辑值，并在 change 时提交一次', async () => {
    const wrapper = mount(ElementInput, { props: { modelValue: 'old', modelModifiers: { lazy: true } } })
    wrappers.push(wrapper)
    const input = wrapper.get('input')
    input.element.value = 'draft'
    await input.trigger('input')
    expect(input.element.value).toBe('draft')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    await input.trigger('change')
    expect(wrapper.emitted('update:modelValue')).toEqual([['draft']])
  })

  it('输入法组词期间不提交未完成的文本', async () => {
    const wrapper = mount(ElementInput)
    wrappers.push(wrapper)
    const input = wrapper.get('input')
    input.element.value = '中文'
    input.element.dispatchEvent(new InputEvent('input', { bubbles: true, isComposing: true }))
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    await input.trigger('compositionend')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['中文'])
  })

  it('数组复选框保留已有选择，取消时仅移除当前值', async () => {
    const wrapper = mount(ElementCheckbox, { props: { modelValue: [7], value: 9 } })
    wrappers.push(wrapper)
    await wrapper.get('input').setValue(true)
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([[7, 9]])
    await wrapper.setProps({ modelValue: [7, 9] })
    await wrapper.get('input').setValue(false)
    expect(wrapper.emitted('update:modelValue')?.[1]).toEqual([[7]])
    expect((wrapper.emitted('change')?.[1]?.[0] as Event).target).toBe(wrapper.get('input').element)
  })

  it('受控 radio 的 checked 与原生 change 保持兼容', async () => {
    const wrapper = mount(ElementRadio, { props: { checked: false }, attrs: { name: 'scope' } })
    wrappers.push(wrapper)
    await wrapper.get('input').setValue(true)
    expect(wrapper.emitted('change')).toHaveLength(1)
    expect((wrapper.emitted('change')?.[0]?.[0] as Event).target).toBe(wrapper.get('input').element)
  })

  it('数字下拉值保留类型，同时向旧处理器提供 select.value', () => {
    const wrapper = mount(ElementSelect, { props: { modelValue: 1 } })
    wrappers.push(wrapper)
    wrapper.getComponent(ElSelect).vm.$emit('change', 9)
    expect(wrapper.emitted('update:modelValue')).toEqual([[9]])
    const event = wrapper.emitted('change')?.[0]?.[0] as Event
    expect(event.target).toBeInstanceOf(HTMLSelectElement)
    expect((event.target as HTMLSelectElement).value).toBe('9')
  })
})
