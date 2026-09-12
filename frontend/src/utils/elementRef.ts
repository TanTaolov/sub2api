import type { ComponentPublicInstance } from 'vue'

/** 将 Element Plus 控件引用转换为定位浮层所需的真实 DOM 元素。 */
export function resolveControlElement(value: Element | ComponentPublicInstance | null): HTMLElement | null {
  const element = value && '$el' in value ? value.$el : value
  return element instanceof HTMLElement ? element : null
}
