<script setup lang="ts" generic="Value = boolean">
import { computed, onBeforeUnmount, onMounted, ref, useAttrs } from 'vue'
import { ElCheckbox } from 'element-plus'

defineOptions({ inheritAttrs: false })
// 显式声明默认值：这两个 prop 含布尔类型，Vue 会把「未传值」直接归一成 false，
// 从而让下面的 trueValue ?? true 失效、勾选态判反，因此用 default: undefined 关掉该归一化。
const props = withDefaults(defineProps<{
  modelValue?: Value
  value?: unknown
  checked?: boolean
  trueValue?: string | number | boolean
  falseValue?: string | number | boolean
}>(), { trueValue: undefined, falseValue: undefined })
const emit = defineEmits<{
  'update:modelValue': [value: Value]
  change: [event: Event]
  input: [event: Event]
}>()
const attrs = useAttrs()
const control = ref<InstanceType<typeof ElCheckbox>>()
const nativeInput = ref<HTMLInputElement | null>(null)
const checkedValue = computed(() => {
  if (Array.isArray(props.modelValue)) return props.modelValue.includes(props.value)
  if (props.modelValue instanceof Set) return props.modelValue.has(props.value)
  if (props.modelValue !== undefined) return props.modelValue === (props.trueValue ?? true)
  return props.checked ?? false
})

/**
 * 旧页面（DataTable 等）与测试依赖 `$event.target.checked`，而 ElCheckbox 自身的 change 载荷只有值，
 * 因此这里直接订阅原生 change：既转发真实事件（target 就是那个 input），也从 DOM 取勾选态。
 */
function handleNativeChange(event: Event) {
  handleChange((event.target as HTMLInputElement).checked, event)
}

onMounted(() => {
  nativeInput.value = (control.value?.$el as ParentNode | undefined)?.querySelector?.('input') ?? null
  nativeInput.value?.addEventListener('change', handleNativeChange)
})

onBeforeUnmount(() => {
  nativeInput.value?.removeEventListener('change', handleNativeChange)
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
  <ElCheckbox ref="control" v-bind="attrs" :model-value="checkedValue" class="element-checkbox">
    <slot />
  </ElCheckbox>
</template>
