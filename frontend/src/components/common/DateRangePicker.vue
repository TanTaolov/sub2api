<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElPopover, ElDatePicker, ElButton } from 'element-plus'
import Icon from '@/components/icons/Icon.vue'
interface DatePreset { labelKey: string; value: string; getRange: () => { start: string; end: string } }
const props = defineProps<{ startDate: string; endDate: string }>()
const emit = defineEmits<{
  'update:startDate': [value: string]
  'update:endDate': [value: string]
  change: [range: { startDate: string; endDate: string; preset: string | null }]
}>()
const { t, locale } = useI18n()
const isOpen = ref(false)
const localRange = ref<[string, string]>([props.startDate, props.endDate])
const activePreset = ref<string | null>('last24Hours')
const today = () => {
  // Use local timezone to avoid UTC timezone issues
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// Tomorrow's date - used for max date to handle timezone differences
// When user is in a timezone behind the server, "today" on server might be "tomorrow" locally
const tomorrow = () => {
  const d = new Date()
  d.setDate(d.getDate() + 1)
  return formatDateToString(d)
}

// Helper function to format date to YYYY-MM-DD using local timezone
const formatDateToString = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const presets: DatePreset[] = [
  {
    labelKey: 'dates.today',
    value: 'today',
    getRange: () => {
      const t = today()
      return { start: t, end: t }
    }
  },
  {
    labelKey: 'dates.yesterday',
    value: 'yesterday',
    getRange: () => {
      const d = new Date()
      d.setDate(d.getDate() - 1)
      const yesterday = formatDateToString(d)
      return { start: yesterday, end: yesterday }
    }
  },
  {
    labelKey: 'dates.last24Hours',
    value: 'last24Hours',
    getRange: () => {
      const end = new Date()
      const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
      return {
        start: formatDateToString(start),
        end: formatDateToString(end)
      }
    }
  },
  {
    labelKey: 'dates.last7Days',
    value: '7days',
    getRange: () => {
      const end = today()
      const d = new Date()
      d.setDate(d.getDate() - 6)
      const start = formatDateToString(d)
      return { start, end }
    }
  },
  {
    labelKey: 'dates.last14Days',
    value: '14days',
    getRange: () => {
      const end = today()
      const d = new Date()
      d.setDate(d.getDate() - 13)
      const start = formatDateToString(d)
      return { start, end }
    }
  },
  {
    labelKey: 'dates.last30Days',
    value: '30days',
    getRange: () => {
      const end = today()
      const d = new Date()
      d.setDate(d.getDate() - 29)
      const start = formatDateToString(d)
      return { start, end }
    }
  },
  {
    labelKey: 'dates.thisMonth',
    value: 'thisMonth',
    getRange: () => {
      const now = new Date()
      const start = formatDateToString(new Date(now.getFullYear(), now.getMonth(), 1))
      return { start, end: today() }
    }
  },
  {
    labelKey: 'dates.lastMonth',
    value: 'lastMonth',
    getRange: () => {
      const now = new Date()
      const start = formatDateToString(new Date(now.getFullYear(), now.getMonth() - 1, 1))
      const end = formatDateToString(new Date(now.getFullYear(), now.getMonth(), 0))
      return { start, end }
    }
  }
]


const displayValue = computed(() => {
  const preset = presets.find(preset => preset.value === activePreset.value)
  if (preset) return t(preset.labelKey)
  if (!props.startDate || !props.endDate) return t('dates.selectDateRange')
  const formatter = new Intl.DateTimeFormat(locale.value === 'zh' ? 'zh-CN' : 'en-US', { month: 'short', day: 'numeric' })
  return [props.startDate, props.endDate].map(date => formatter.format(new Date(date + 'T00:00:00'))).join(' – ')
})
function selectPreset(preset: DatePreset) {
  const range = preset.getRange()
  localRange.value = [range.start, range.end]
  activePreset.value = preset.value
}
function detectPreset() {
  activePreset.value = presets.find(preset => {
    const range = preset.getRange()
    return range.start === localRange.value[0] && range.end === localRange.value[1]
  })?.value ?? null
}
function apply() {
  const [startDate, endDate] = localRange.value
  if (!startDate || !endDate || startDate > endDate) return
  emit('update:startDate', startDate)
  emit('update:endDate', endDate)
  emit('change', { startDate, endDate, preset: activePreset.value })
  isOpen.value = false
}
watch(() => [props.startDate, props.endDate], ([start, end]) => { localRange.value = [start, end]; detectPreset() }, { immediate: true })
watch(isOpen, visible => {
  if (visible) {
    localRange.value = [props.startDate, props.endDate]
    detectPreset()
  }
})
</script>
<template>
  <ElPopover v-model:visible="isOpen" trigger="click" placement="bottom-start" :width="380">
    <template #reference>
      <ElButton :aria-expanded="isOpen"><Icon name="calendar" size="sm" class="mr-2" />{{ displayValue }}</ElButton>
    </template>
    <div class="flex flex-wrap gap-2 pb-4">
      <ElButton v-for="preset in presets" :key="preset.value" size="small" :type="activePreset === preset.value ? 'primary' : 'default'" plain @click="selectPreset(preset)">{{ t(preset.labelKey) }}</ElButton>
    </div>
    <ElDatePicker v-model="localRange" type="daterange" value-format="YYYY-MM-DD" :clearable="false"
      :start-placeholder="t('dates.startDate')" :end-placeholder="t('dates.endDate')" unlink-panels
      :disabled-date="(date: Date) => formatDateToString(date) > tomorrow()" style="width: 100%" @change="detectPreset" />
    <div class="mt-4 flex justify-end"><ElButton type="primary" @click="apply">{{ t('dates.apply') }}</ElButton></div>
  </ElPopover>
</template>
