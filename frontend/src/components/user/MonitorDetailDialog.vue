<template>
  <BaseDialog
    :show="show"
    :title="title"
    width="wide"
    @close="$emit('close')"
  >
    <div v-if="loading" class="py-8 text-center text-sm text-gray-500">
      {{ t('common.loading') }}
    </div>
    <div v-else-if="!detail" class="py-8 text-center text-sm text-gray-500">
      {{ t('channelStatus.detailLoadError') }}
    </div>
    <div v-else class="overflow-x-auto">
      <ElTable  row-key="model" row-class-name="border-b border-gray-100 dark:border-dark-800" :data="detail.models" table-layout="auto" class="element-data-table">
  <ElTableColumn :min-width="120" align="left">
    <template #header><div class="py-2 pr-3">{{ t('channelStatus.detailColumns.model') }}</div></template>
    <template #default="{ row: m, $index: rowIndex }"><div class="py-2 pr-3 font-medium text-gray-900 dark:text-gray-100" >{{ formatMonitorModel(m.model) }}</div></template>
  </ElTableColumn>
  <ElTableColumn :min-width="120" align="left">
    <template #header><div class="py-2 pr-3">{{ t('channelStatus.detailColumns.latestStatus') }}</div></template>
    <template #default="{ row: m, $index: rowIndex }"><div class="py-2 pr-3" ><span
                class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px]"
                :class="statusBadgeClass(m.latest_status)"
              >
                {{ statusLabel(m.latest_status) }}
              </span></div></template>
  </ElTableColumn>
  <ElTableColumn :min-width="120" align="left">
    <template #header><div class="py-2 pr-3">{{ t('channelStatus.detailColumns.latestLatency') }}</div></template>
    <template #default="{ row: m, $index: rowIndex }"><div class="py-2 pr-3 text-gray-700 dark:text-gray-300" >{{ formatLatency(m.latest_latency_ms) }}</div></template>
  </ElTableColumn>
  <ElTableColumn :min-width="120" align="left">
    <template #header><div class="py-2 pr-3">{{ t('channelStatus.detailColumns.availability7d') }}</div></template>
    <template #default="{ row: m, $index: rowIndex }"><div class="py-2 pr-3 text-gray-700 dark:text-gray-300" >{{ formatPercent(m.availability_7d) }}</div></template>
  </ElTableColumn>
  <ElTableColumn :min-width="120" align="left">
    <template #header><div class="py-2 pr-3">{{ t('channelStatus.detailColumns.availability15d') }}</div></template>
    <template #default="{ row: m, $index: rowIndex }"><div class="py-2 pr-3 text-gray-700 dark:text-gray-300" >{{ formatPercent(m.availability_15d) }}</div></template>
  </ElTableColumn>
  <ElTableColumn :min-width="120" align="left">
    <template #header><div class="py-2 pr-3">{{ t('channelStatus.detailColumns.availability30d') }}</div></template>
    <template #default="{ row: m, $index: rowIndex }"><div class="py-2 pr-3 text-gray-700 dark:text-gray-300" >{{ formatPercent(m.availability_30d) }}</div></template>
  </ElTableColumn>
  <ElTableColumn :min-width="120" align="left">
    <template #header><div class="py-2 pr-3">{{ t('channelStatus.detailColumns.avgLatency7d') }}</div></template>
    <template #default="{ row: m, $index: rowIndex }"><div class="py-2 pr-3 text-gray-700 dark:text-gray-300" >{{ formatLatency(m.avg_latency_7d_ms) }}</div></template>
  </ElTableColumn>
</ElTable>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <ElButton @click="$emit('close')" class="">
          {{ t('channelStatus.closeDetail') }}
        </ElButton>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  status as fetchChannelMonitorDetail,
  type UserMonitorDetail,
} from '@/api/channelMonitor'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

const props = defineProps<{
  show: boolean
  monitorId: number | null
  title: string
}>()

defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const { statusLabel, statusBadgeClass, formatLatency, formatPercent, formatMonitorModel } = useChannelMonitorFormat()

const detail = ref<UserMonitorDetail | null>(null)
const loading = ref(false)

async function load(id: number) {
  detail.value = null
  loading.value = true
  try {
    detail.value = await fetchChannelMonitorDetail(id)
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.detailLoadError')))
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.show, props.monitorId] as const,
  ([show, id]) => {
    if (!show) {
      detail.value = null
      return
    }
    if (id != null) void load(id)
  },
  { immediate: true },
)
</script>
