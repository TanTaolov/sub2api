<script setup lang="ts" generic="Value extends string | number | null | undefined = string">
import { computed, ref, useAttrs, watch } from 'vue'
import { ElInput, type InputInstance } from 'element-plus'

defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{
  modelValue?: Value
  value?: string | number | null
  type?: string
  modelModifiers?: { lazy?: boolean; trim?: boolean; number?: boolean }
}>(), { type: 'text' })
const emit = defineEmits<{
  'update:modelValue': [value: Value]
  input: [event: Event]
  change: [event: Event]
}>()
const attrs = useAttrs()
const control = ref<InputInstance>()
const localValue = ref<string | number>(props.modelValue ?? props.value ?? '')
const currentValue = computed(() => props.modelModifiers?.lazy
  ? localValue.value : props.modelValue ?? props.value ?? localValue.value)
watch(() => [props.modelValue, props.value], () => {
  localValue.value = props.modelValue ?? props.value ?? ''
})

function updateValue(value: string, phase: 'input' | 'change') {
  localValue.value = value
  if (props.modelModifiers?.lazy ? phase !== 'change' : phase !== 'input') return
  let nextValue: string | number = props.modelModifiers?.trim ? value.trim() : value
  if (props.type === 'number' || props.modelModifiers?.number) {
    const numericValue = Number.parseFloat(nextValue)
    if (!Number.isNaN(numericValue)) nextValue = numericValue
  }
  emit('update:modelValue', nextValue as Value)
}

// 保留现有业务处理器的原生事件契约，值更新仍由 Element Plus 控件负责。
function handleNativeEvent(event: Event, phase: 'input' | 'change') {
  const input = event.target as HTMLInputElement | HTMLTextAreaElement
  if (!(event instanceof InputEvent && event.isComposing)) {
    updateValue(input.value, phase)
  }
  emit(phase, event)
}

defineExpose({
  focus: () => control.value?.focus(),
  blur: () => control.value?.blur(),
  select: () => control.value?.select(),
  get input() { return control.value?.input ?? control.value?.textarea },
  get value() { return (control.value?.input ?? control.value?.textarea)?.value },
  set value(value: string | undefined) {
    localValue.value = value ?? ''
    const input = control.value?.input ?? control.value?.textarea
    if (input) input.value = value ?? ''
  },
  setSelectionRange: (start: number, end: number) =>
    (control.value?.input ?? control.value?.textarea)?.setSelectionRange(start, end)
})
</script>

<template>
  <ElInput
    ref="control"
    v-bind="attrs"
    :type="type"
    :model-value="currentValue"
    class="element-field"
    @input.capture="handleNativeEvent($event, 'input')"
    @change.capture="handleNativeEvent($event, 'change')"
    @compositionend="updateValue(($event.target as HTMLInputElement).value, 'input')"
  >
    <template v-for="(_, name) in $slots" #[name]="slotProps">
      <slot :name="name" v-bind="slotProps || {}" />
    </template>
  </ElInput>
</template>
