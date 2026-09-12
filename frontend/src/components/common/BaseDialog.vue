<script setup lang="ts">
import { computed } from 'vue'
import { ElDialog } from 'element-plus'
const props = withDefaults(defineProps<{
  show: boolean
  title: string
  width?: 'narrow' | 'normal' | 'wide' | 'extra-wide' | 'full'
  closeOnEscape?: boolean
  closeOnClickOutside?: boolean
  showCloseButton?: boolean
  zIndex?: number
}>(), { width: 'normal', closeOnEscape: true, closeOnClickOutside: false, showCloseButton: true })
const emit = defineEmits<{ close: [] }>()
const dialogWidth = computed(() => ({ narrow: '448px', normal: '512px', wide: '768px', 'extra-wide': '1024px', full: '1280px' })[props.width])
function handleVisibilityChange(visible: boolean) {
  if (!visible) emit('close')
}
</script>
<template>
  <ElDialog
    :model-value="show" :title="title" :width="dialogWidth"
    :close-on-click-modal="closeOnClickOutside" :close-on-press-escape="closeOnEscape"
    :show-close="showCloseButton" :z-index="zIndex" align-center append-to-body destroy-on-close
    class="element-dialog" @update:model-value="handleVisibilityChange"
  >
    <slot />
    <template v-if="$slots.footer" #footer><slot name="footer" /></template>
  </ElDialog>
</template>
