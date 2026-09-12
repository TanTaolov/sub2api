<script setup lang="ts" generic="Value extends string | number | boolean | null | undefined">
import { computed } from 'vue'
import { ElRadio } from 'element-plus'

defineOptions({ inheritAttrs: false })
const props = defineProps<{ modelValue?: Value; value?: Value; checked?: boolean }>()
const radioValue = computed(() => props.value ?? true)
const selectedValue = computed(() => props.modelValue ?? (props.checked ? radioValue.value : undefined))
const emit = defineEmits<{
  'update:modelValue': [value: Value]
  change: [event: Event]
}>()
function handleChange(event: Event) {
  emit('update:modelValue', props.value as Value)
  emit('change', event)
}
</script>

<template>
  <ElRadio v-bind="$attrs" :model-value="selectedValue" :value="radioValue" class="element-radio" @change.capture="handleChange($event as unknown as Event)">
    <slot />
  </ElRadio>
</template>
