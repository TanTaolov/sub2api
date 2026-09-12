<script setup lang="ts">
import { useDebounceFn } from '@vueuse/core'
import { ElInput } from 'element-plus'
import Icon from '@/components/icons/Icon.vue'
const props = withDefaults(defineProps<{ modelValue: string; placeholder?: string; debounceMs?: number }>(), { placeholder: 'Search…', debounceMs: 300 })
const emit = defineEmits<{ 'update:modelValue': [value: string]; search: [value: string] }>()
const debouncedEmitSearch = useDebounceFn((value: string) => emit('search', value), props.debounceMs)
function handleInput(value: string) { emit('update:modelValue', value); debouncedEmitSearch(value) }
</script>
<template>
  <ElInput :model-value="modelValue" :placeholder="placeholder" :aria-label="placeholder" type="search" class="element-field" clearable @update:model-value="handleInput">
    <template #prefix><Icon name="search" size="md" /></template>
  </ElInput>
</template>
