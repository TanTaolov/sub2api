import { afterEach, describe, expect, it } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { ElDialog } from 'element-plus'
import BaseDialog from '../BaseDialog.vue'
const wrappers: VueWrapper[] = []
afterEach(() => { wrappers.forEach(wrapper => wrapper.unmount()); wrappers.length = 0; document.body.innerHTML = '' })
describe('BaseDialog', () => {
  it('提供可访问的标题、内容与页脚', async () => {
    const wrapper = mount(BaseDialog, { attachTo: document.body, props: { show: true, title: 'Details' }, slots: { default: '<p>Dialog content</p>', footer: '<button>Save</button>' } })
    wrappers.push(wrapper)
    await wrapper.vm.$nextTick()
    expect(document.body.querySelector('[role="dialog"]')?.getAttribute('aria-label')).toBe('Details')
    expect(document.body.querySelector('.el-dialog__body')?.textContent).toContain('Dialog content')
    expect(document.body.querySelector('.el-dialog__footer')?.textContent).toContain('Save')
  })
  it('关闭请求交给父组件，保留禁止遮罩关闭的默认行为', () => {
    const wrapper = mount(BaseDialog, { props: { show: true, title: 'Details' } })
    wrappers.push(wrapper)
    const dialog = wrapper.getComponent(ElDialog)
    expect(dialog.props('closeOnClickModal')).toBe(false)
    dialog.vm.$emit('update:modelValue', false)
    expect(wrapper.emitted('close')).toEqual([[]])
    expect(wrapper.props('show')).toBe(true)
  })
})
