import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { ElPopover } from 'element-plus'
import HelpTooltip from '../HelpTooltip.vue'
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
const wrappers: VueWrapper[] = []
afterEach(() => { wrappers.forEach(wrapper => wrapper.unmount()); wrappers.length = 0; vi.useRealTimers() })
describe('HelpTooltip', () => {
  it('悬停提示同时支持键盘焦点，并允许移入提示复制内容', () => {
    const wrapper = mount(HelpTooltip, { props: { content: 'Details' } })
    wrappers.push(wrapper)
    expect(wrapper.get('[tabindex="0"]').attributes('aria-label')).toBe('Details')
    expect(wrapper.getComponent(ElPopover).props('trigger')).toEqual(['hover', 'focus'])
    expect(wrapper.getComponent(ElPopover).props('enterable')).toBe(true)
  })
  it('点击模式保留开关状态，Escape 关闭', async () => {
    const wrapper = mount(HelpTooltip, { props: { content: 'Details', trigger: 'click' } })
    wrappers.push(wrapper)
    const popover = wrapper.getComponent(ElPopover)
    expect(popover.props('trigger')).toBe('click')
    popover.vm.$emit('update:visible', true)
    await wrapper.vm.$nextTick()
    expect(popover.props('visible')).toBe(true)
    await wrapper.get('[tabindex="0"]').trigger('keydown', { key: 'Escape' })
    expect(popover.props('visible')).toBe(false)
  })
})
