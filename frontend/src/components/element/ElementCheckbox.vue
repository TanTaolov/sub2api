<script setup lang="ts" generic="Value = boolean">
import { computed, useAttrs } from 'vue'
import { ElCheckbox } from 'element-plus'

defineOptions({ inheritAttrs: false })
const props = defineProps<{
  modelValue?: Value
  value?: unknown
  checked?: boolean
  trueValue?: string | number | boolean
  falseValue?: string | number | boolean
}>()
const emit = defineEmits<{
  'update:modelValue': [value: Value]
  change: [event: Event]
  input: [event: Event]
}>()
const attrs = useAttrs()
const checkedValue = computed(() => {
  if (Array.isArray(props.modelValue)) return props.modelValue.includes(props.value)
  if (props.modelValue instanceof Set) return props.modelValue.has(props.value)
  if (props.modelValue !== undefined) return props.modelValue === (props.trueValue ?? true)
  return props.checked ?? false
})

function handleChange(checked: string | number | boolean, event: Event) {
  let nextValue: unknown = checked ? props.trueValue ?? true : props.falseValue ?? false
  if (Array.isArray(props.modelValue)) {
    nextValue = checked
      ? [...new Set([...props.modelValue, props.value])]
      : props.modelValue.filter((value) => value !== props.value)
  } else if (props.modelValue instanceof Set) {
    const selection = new Set(props.modelValue)
    if (checked) selection.add(props.value)
    else selection.delete(props.value)
    nextValue = selection
  }
  emit('update:modelValue', nextValue as Value)
  emit('input', event)
  emit('change', event)
}
</script>

<template>
  <ElCheckbox v-bind="attrs" :model-value="checkedValue" class="element-checkbox" @change="handleChange">
    <slot />
  </ElCheckbox>
</template>
