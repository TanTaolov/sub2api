<script setup lang="ts">
import { computed, getCurrentInstance, ref } from 'vue'
import { ElInput, type InputInstance } from 'element-plus'
defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{
  modelValue: string | number | null | undefined; type?: string; label?: string; placeholder?: string;
  disabled?: boolean; required?: boolean; readonly?: boolean; error?: string; hint?: string; id?: string; autocomplete?: string
}>(), { type: 'text', disabled: false, required: false, readonly: false })
const emit = defineEmits<{
  'update:modelValue': [value: string]; change: [value: string]; blur: [event: FocusEvent]; focus: [event: FocusEvent]; enter: [event: KeyboardEvent]
}>()
const instanceId = getCurrentInstance()!.uid
const fieldId = computed(() => props.id || 'input-' + instanceId)
const inputRef = ref<InputInstance>()
defineExpose({ focus: () => inputRef.value?.focus(), select: () => inputRef.value?.select() })
</script>
<template>
  <div class="w-full">
    <label v-if="label" :for="fieldId" class="input-label mb-1.5 block">{{ label }}<span v-if="required" class="ml-1 text-red-500">*</span></label>
    <ElInput ref="inputRef" v-bind="$attrs" :id="fieldId" :model-value="modelValue ?? ''" :type="type"
      :disabled="disabled" :required="required" :readonly="readonly" :placeholder="placeholder ?? ''" :autocomplete="autocomplete"
      :aria-invalid="!!error" :aria-describedby="error || hint ? fieldId + '-description' : undefined"
      :class="['element-field', { 'input-error': error }]"
      @update:model-value="emit('update:modelValue', String($event))" @change="emit('change', String($event))"
      @blur="emit('blur', $event)" @focus="emit('focus', $event)" @keyup.enter="emit('enter', $event)">
      <template v-if="$slots.prefix" #prefix><slot name="prefix" /></template>
      <template v-if="$slots.suffix" #suffix><slot name="suffix" /></template>
    </ElInput>
    <p v-if="error || hint" :id="fieldId + '-description'" :class="[error ? 'input-error-text' : 'input-hint', 'mt-1.5']" :role="error ? 'alert' : undefined">{{ error || hint }}</p>
  </div>
</template>
