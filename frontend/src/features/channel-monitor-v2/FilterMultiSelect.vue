<script setup lang="ts">
import { ElSelect, ElOption, ElButton } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { monitorIntlLocale } from '@/features/channel-monitor-v2/monitorFormat'
withDefaults(defineProps<{
  label: string; allLabel: string; modelValue: string[]; options: Array<{ value: string; label: string; count?: number }>; compact?: boolean
}>(), { compact: false })
const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>()
const { t } = useI18n()
</script>
<template>
  <ElSelect :model-value="modelValue" multiple collapse-tags collapse-tags-tooltip clearable
    :placeholder="t('channelMonitorV2.filters.labelValue', { label, value: allLabel })" :aria-label="label"
    :size="compact ? 'small' : 'default'" :class="compact ? 'min-w-[7.25rem]' : 'min-w-[160px]'"
    @update:model-value="emit('update:modelValue', $event)">
    <template #header><ElButton text size="small" @click="emit('update:modelValue', [])">{{ allLabel }}</ElButton></template>
    <ElOption v-for="option in options" :key="option.value" :label="option.label" :value="option.value">
      <div class="flex items-center justify-between gap-4"><span>{{ option.label }}</span><small v-if="option.count != null" class="text-gray-400">{{ option.count.toLocaleString(monitorIntlLocale()) }}</small></div>
    </ElOption>
  </ElSelect>
</template>
