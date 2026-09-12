<template>
  <!-- 用量页"用户排行"tab 内容：无卡片外观，依赖父级统一卡片；筛选/时间范围复用页面级筛选栏 -->
  <div>
    <!-- Toolbar -->
    <div class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-100 px-4 py-3 dark:border-dark-700/50 sm:px-6">
      <p class="text-xs text-gray-400 dark:text-gray-500">{{ t('admin.usage.tokenRanking.subtitle') }}</p>
      <div class="flex items-center gap-3">
        <span v-if="!loading && items.length > 0" class="text-xs text-gray-400 dark:text-gray-500">
          {{ t('admin.usage.tokenRanking.userCount', { count: items.length }) }}
        </span>
        <div class="w-28">
          <Select v-model="limit" :options="limitOptions" @change="load" />
        </div>
      </div>
    </div>

    <!-- Table -->
    <div class="overflow-x-auto">
      <ElTable :data="loading ? [] : items" row-key="user_id" class="element-data-table cursor-pointer" @row-click="item => $emit('select-user', item.user_id, item.email)"><ElTableColumn :min-width="120" :width="64" align="left"><template #header>#</template><template #default="{ $index: index }"><div class="px-4 py-3 sm:px-6"><span
                v-if="index < 3"
                class="inline-flex h-6 w-6 items-center justify-center rounded-full text-xs font-semibold"
                :class="RANK_BADGE_CLASSES[index]"
              >{{ index + 1 }}</span><span v-else class="inline-block w-6 text-center text-sm tabular-nums text-gray-400">{{ index + 1 }}</span></div></template></ElTableColumn><ElTableColumn :min-width="120"  align="left"><template #header>{{ t('admin.usage.tokenRanking.columns.user') }}</template><template #default="{ row: item }"><div class="max-w-[260px] truncate px-4 py-3 text-sm font-medium text-gray-700 dark:text-gray-200" :title="item.email">{{ item.email || `User #${item.user_id}` }}
              <span class="ml-1 font-normal text-gray-400 dark:text-gray-500">#{{ item.user_id }}</span></div></template></ElTableColumn><ElTableColumn :min-width="120"  align="right"><template #header><ElButton text size="small" @click="setSort(sortableColumns[0].key)">{{ t(sortableColumns[0].label) }}<span v-if="sortBy === sortableColumns[0].key">↓</span></ElButton></template><template #default="{ row: item }"><div class="whitespace-nowrap px-4 py-3 text-right text-sm tabular-nums text-gray-500 dark:text-gray-400">{{ item.requests.toLocaleString() }}</div></template></ElTableColumn><ElTableColumn :min-width="120"  align="right"><template #header><ElButton text size="small" @click="setSort(sortableColumns[1].key)">{{ t(sortableColumns[1].label) }}<span v-if="sortBy === sortableColumns[1].key">↓</span></ElButton></template><template #default="{ row: item }"><div class="whitespace-nowrap px-4 py-3 text-right text-sm tabular-nums text-gray-500 dark:text-gray-400">{{ fmtTokens(item.input_tokens) }}</div></template></ElTableColumn><ElTableColumn :min-width="120"  align="right"><template #header><ElButton text size="small" @click="setSort(sortableColumns[2].key)">{{ t(sortableColumns[2].label) }}<span v-if="sortBy === sortableColumns[2].key">↓</span></ElButton></template><template #default="{ row: item }"><div class="whitespace-nowrap px-4 py-3 text-right text-sm tabular-nums text-gray-500 dark:text-gray-400">{{ fmtTokens(item.output_tokens) }}</div></template></ElTableColumn><ElTableColumn :min-width="120"  align="right"><template #header><ElButton text size="small" @click="setSort(sortableColumns[3].key)">{{ t(sortableColumns[3].label) }}<span v-if="sortBy === sortableColumns[3].key">↓</span></ElButton></template><template #default="{ row: item }"><div class="whitespace-nowrap px-4 py-3 text-right text-sm tabular-nums text-gray-500 dark:text-gray-400">{{ fmtTokens(item.cache_tokens) }}</div></template></ElTableColumn><ElTableColumn :min-width="120"  align="right"><template #header><ElButton text size="small" @click="setSort(sortableColumns[4].key)">{{ t(sortableColumns[4].label) }}<span v-if="sortBy === sortableColumns[4].key">↓</span></ElButton></template><template #default="{ row: item }"><div class="whitespace-nowrap px-4 py-3 text-right text-sm font-medium tabular-nums text-gray-900 dark:text-gray-100">{{ fmtTokens(item.total_tokens) }}</div></template></ElTableColumn><ElTableColumn :min-width="120"  align="right"><template #header><ElButton text size="small" @click="setSort(sortableColumns[5].key)">{{ t(sortableColumns[5].label) }}<span v-if="sortBy === sortableColumns[5].key">↓</span></ElButton></template><template #default="{ row: item }"><div class="whitespace-nowrap px-4 py-3 text-right text-sm font-medium tabular-nums text-green-600 dark:text-green-400">${{ fmtCost(item.actual_cost) }}</div></template></ElTableColumn><template #empty><LoadingSpinner v-if="loading" /><span v-else>{{ t('admin.dashboard.noDataAvailable') }}</span></template></ElTable>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getUserBreakdown, type UserBreakdownParams } from '@/api/admin/dashboard'
import { formatCompactNumber, formatCostFixed } from '@/utils/format'
import type { UserBreakdownItem } from '@/types'
import Select from '@/components/common/Select.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const props = defineProps<{
  startDate: string
  endDate: string
  filters: Record<string, unknown>
  model?: string
}>()

defineEmits<{ (e: 'select-user', userId: number, email: string): void }>()

const { t } = useI18n()

type SortKey = NonNullable<UserBreakdownParams['sort_by']>
const sortableColumns: { key: SortKey; label: string }[] = [
  { key: 'requests', label: 'admin.usage.tokenRanking.columns.requests' },
  { key: 'input_tokens', label: 'admin.usage.tokenRanking.columns.inputTokens' },
  { key: 'output_tokens', label: 'admin.usage.tokenRanking.columns.outputTokens' },
  { key: 'cache_tokens', label: 'admin.usage.tokenRanking.columns.cacheTokens' },
  { key: 'total_tokens', label: 'admin.usage.tokenRanking.columns.totalTokens' },
  { key: 'actual_cost', label: 'admin.usage.tokenRanking.columns.cost' },
]

const limitOptions = [
  { value: 20, label: 'Top 20' },
  { value: 50, label: 'Top 50' },
  { value: 100, label: 'Top 100' },
  { value: 200, label: 'Top 200' },
]

// 前三名金/银/铜徽章
const RANK_BADGE_CLASSES = [
  'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-400',
  'bg-gray-200 text-gray-600 dark:bg-gray-500/20 dark:text-gray-300',
  'bg-orange-100 text-orange-700 dark:bg-orange-500/20 dark:text-orange-400',
]

const items = ref<UserBreakdownItem[]>([])
const loading = ref(false)
const sortBy = ref<SortKey>('total_tokens')
const limit = ref(50)
let reqSeq = 0

const fmtTokens = (v: number) => formatCompactNumber(v)
const fmtCost = (v: number) => formatCostFixed(v, 4)

const setSort = (key: SortKey) => {
  if (sortBy.value === key) return
  sortBy.value = key
  load()
}

const load = async () => {
  const seq = ++reqSeq
  loading.value = true
  try {
    const params: UserBreakdownParams = {
      ...props.filters,
      start_date: props.startDate,
      end_date: props.endDate,
      sort_by: sortBy.value,
      limit: limit.value,
    }
    if (props.model) params.model = props.model
    const res = await getUserBreakdown(params)
    if (seq !== reqSeq) return
    items.value = res.users || []
  } catch {
    if (seq !== reqSeq) return
    items.value = []
  } finally {
    if (seq === reqSeq) loading.value = false
  }
}

// Reload when the shared filters / date range / model change.
watch(
  () => [props.startDate, props.endDate, props.model, JSON.stringify(props.filters)],
  () => load(),
  { immediate: true }
)

defineExpose({ reload: load })
</script>
