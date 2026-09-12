<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElSelect, type SelectInstance } from 'element-plus'

defineOptions({ inheritAttrs: false })
const props = defineProps<{
  modelValue?: string | number | boolean | string[] | number[] | null
  value?: string | number | boolean
  modelModifiers?: { number?: boolean }
}>()
const emit = defineEmits<{
  'update:modelValue': [value: any]
  change: [event: Event]
}>()
const control = ref<SelectInstance>()
const currentValue = computed(() => props.modelValue ?? props.value)
function handleChange(value: any) {
  const nextValue = props.modelModifiers?.number && typeof value === 'string' && value !== ''
    ? Number(value) : value
  emit('update:modelValue', nextValue)
  // 旧页面的 change 处理器读取 HTMLSelectElement.value；在组件边界保持该契约。
  const target = document.createElement('select')
  target.add(new Option(String(value ?? ''), String(value ?? '')))
  const event = new Event('change', { bubbles: true })
  Object.defineProperty(event, 'target', { value: target })
  emit('change', event)
}
defineExpose({ focus: () => control.value?.focus(), blur: () => control.value?.blur() })
</script>
<template>
  <ElSelect ref="control" v-bind="$attrs" :model-value="currentValue" :empty-values="[undefined, null]" class="element-select" @change="handleChange">
    <slot />
  </ElSelect>
</template>
