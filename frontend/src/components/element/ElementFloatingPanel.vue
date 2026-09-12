<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { ElPopover } from 'element-plus'

const props = withDefaults(defineProps<{
  visible: boolean
  anchor?: HTMLElement | { getBoundingClientRect: () => DOMRect } | null
  position?: { top?: number | string; bottom?: number | string; left?: number | string; x?: number; y?: number } | null
  width?: number | string
  placement?: 'bottom-start' | 'bottom-end' | 'top' | 'right'
  interactive?: boolean
  fitReference?: boolean
}>(), { width: 256, placement: 'bottom-start', interactive: true })
const emit = defineEmits<{ close: [] }>()
const content = ref<HTMLElement>()
const referenceContainer = ref<HTMLElement>()
let previousFocus: HTMLElement | null = null
const referenceElement = computed(() => props.anchor ?? referenceContainer.value?.firstElementChild as HTMLElement | undefined)
const panelWidth = computed(() => props.fitReference && props.visible && referenceElement.value
  ? Math.max(192, referenceElement.value.getBoundingClientRect().width) : props.width)
const virtualAnchor = computed(() => referenceElement.value ?? {
  getBoundingClientRect: () => {
    const position = props.position
    const left = Number.parseFloat(String(position?.left ?? position?.x ?? 0)) || 0
    const top = position?.bottom != null
      ? window.innerHeight - Number.parseFloat(String(position.bottom))
      : Number.parseFloat(String(position?.top ?? position?.y ?? 0)) || 0
    return new DOMRect(left, top, 0, 0)
  }
})
function closeOutside(event: PointerEvent) {
  if (!props.interactive) return
  const target = event.target as Node
  if (content.value?.contains(target)) return
  if (referenceElement.value instanceof HTMLElement && referenceElement.value.contains(target)) return
  emit('close')
}
function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    emit('close')
    previousFocus?.focus()
    return
  }
  if (!props.interactive || !['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key)) return
  if (event.target instanceof HTMLInputElement || event.target instanceof HTMLTextAreaElement) return
  if (document.activeElement !== previousFocus && !content.value?.contains(document.activeElement)) return
  const controls = Array.from(content.value?.querySelectorAll<HTMLElement>('button:not(:disabled), a[href], [tabindex="0"]') ?? []).filter(element => element.getClientRects().length)
  if (!controls.length) return
  event.preventDefault()
  const current = controls.indexOf(document.activeElement as HTMLElement)
  const next = event.key === 'Home' ? 0 : event.key === 'End' ? controls.length - 1 : (current + (event.key === 'ArrowUp' ? -1 : 1) + controls.length) % controls.length
  controls[next]?.focus()
}
function removeListeners() {
  document.removeEventListener('pointerdown', closeOutside)
  document.removeEventListener('keydown', handleKeydown, true)
}
watch(() => props.visible, visible => {
  removeListeners()
  if (visible) {
    previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null
    document.addEventListener('pointerdown', closeOutside)
    document.addEventListener('keydown', handleKeydown, true)
  }
}, { immediate: true })
onBeforeUnmount(removeListeners)
</script>

<template>
  <div v-if="$slots.reference" ref="referenceContainer" class="contents"><slot name="reference" /></div>
  <ElPopover :visible="visible" :virtual-ref="virtualAnchor" virtual-triggering
    :width="panelWidth" :placement="placement" :show-arrow="false" :offset="referenceElement ? 8 : 0"
    :persistent="false" :enterable="interactive" popper-class="element-floating-panel">
    <div ref="content" @click.stop><slot /></div>
  </ElPopover>
</template>
