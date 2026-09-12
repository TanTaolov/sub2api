<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElSelect, ElOption } from 'element-plus'
export interface SelectOption {
  value: string | number | boolean | null
  label: string
  disabled?: boolean
  [key: string]: unknown
}
type SelectValue = string | number | boolean | null
const props = withDefaults(defineProps<{
  modelValue: SelectValue | undefined
  options: SelectOption[] | Array<Record<string, unknown>>
  placeholder?: string
  disabled?: boolean
  error?: boolean
  searchable?: boolean | 'auto'
  searchPlaceholder?: string
  emptyText?: string
  valueKey?: string
  labelKey?: string
  creatable?: boolean
  creatablePrefix?: string
  clearable?: boolean
  id?: string
  ariaLabel?: string
  ariaDescribedby?: string
  remote?: boolean
  loading?: boolean
}>(), { disabled: false, error: false, searchable: 'auto', creatable: false, creatablePrefix: '', clearable: false, valueKey: 'value', labelKey: 'label', remote: false, loading: false })
const emit = defineEmits<{
  'update:modelValue': [value: SelectValue]
  change: [value: SelectValue, option: SelectOption | null]
  search: [query: string]
}>()
const { t } = useI18n()
const query = ref('')
const isOpen = ref(false)
let searchTimer: ReturnType<typeof setTimeout> | undefined
const isSearchable = computed(() => props.remote || props.creatable || props.searchable === true || (props.searchable === 'auto' && props.options.length > 10))
function getValue(option: Record<string, unknown>): SelectValue { return option[props.valueKey] as SelectValue }
function getLabel(option: Record<string, unknown>): string { return String(option[props.labelKey] ?? '') }
function encodeValue(value: SelectValue | undefined): string { return typeof value + ':' + String(value) }
const selectedOption = computed(() => props.options.find(option => getValue(option) === props.modelValue) ?? null)
const filteredOptions = computed(() => {
  if (!query.value || props.remote) return props.options
  const normalizedQuery = query.value.toLocaleLowerCase()
  const options = props.options.filter(option => getLabel(option).toLocaleLowerCase().includes(normalizedQuery) || String(option.description ?? '').toLocaleLowerCase().includes(normalizedQuery))
  if (props.creatable && query.value.trim() && !options.some(option => getValue(option) === query.value.trim())) {
    return [{ [props.valueKey]: query.value.trim(), [props.labelKey]: (props.creatablePrefix || t('common.search')) + ' “' + query.value.trim() + '”', _creatable: true }, ...options]
  }
  return options
})
function filterOptions(value: string) {
  query.value = value
  clearTimeout(searchTimer)
  if (props.remote && isOpen.value) searchTimer = setTimeout(() => emit('search', value), 300)
}
function handleVisibilityChange(visible: boolean) {
  isOpen.value = visible
  if (!visible) { clearTimeout(searchTimer); query.value = '' }
}
function handleSelection(encoded: string | undefined) {
  const option = filteredOptions.value.find(option => encodeValue(getValue(option)) === encoded)
  const value = option ? getValue(option) : null
  emit('update:modelValue', value)
  emit('change', value, (option as SelectOption | undefined) ?? null)
}
onBeforeUnmount(() => clearTimeout(searchTimer))
</script>
<template>
  <ElSelect
    :id="id" :model-value="modelValue === undefined || (modelValue === null && !selectedOption) ? undefined : encodeValue(modelValue)"
    :placeholder="placeholder ?? t('common.selectOption')" :disabled="disabled"
    :filterable="isSearchable" :filter-method="filterOptions" :clearable="clearable"
    :loading="loading" :loading-text="t('common.loading')" :no-data-text="emptyText ?? t('common.noOptionsFound')"
    :no-match-text="emptyText ?? t('common.noOptionsFound')" :aria-label="ariaLabel ?? searchPlaceholder ?? placeholder"
    :aria-describedby="ariaDescribedby" :class="['w-full', { 'element-select-error': error }]"
    popper-class="element-select-dropdown" :empty-values="[undefined]"
    @change="handleSelection" @visible-change="handleVisibilityChange"
  >
    <template v-if="$slots.selected" #label><slot name="selected" :option="selectedOption" /></template>
    <ElOption v-if="modelValue != null && !selectedOption && !filteredOptions.some(option => getValue(option) === modelValue)"
      :value="encodeValue(modelValue)" :label="String(modelValue)" hidden />
    <ElOption v-for="option in filteredOptions" :key="encodeValue(getValue(option))"
      :value="encodeValue(getValue(option))" :label="getLabel(option)"
      :disabled="Boolean(option.disabled) || option.kind === 'group'"
      :class="{ 'element-option-group': option.kind === 'group' }"
    >
      <slot name="option" :option="option" :selected="getValue(option) === modelValue">{{ getLabel(option) }}</slot>
    </ElOption>
  </ElSelect>
</template>
