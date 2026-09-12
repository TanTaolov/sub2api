<template>
  <div class="bg-gray-50/50 dark:bg-dark-700/30">
    <div v-if="loading" class="flex items-center justify-center py-3">
      <LoadingSpinner />
    </div>
    <div v-else-if="items.length === 0" class="py-2 text-center text-xs text-gray-400">
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
    <ElTable v-else row-key="user_id" row-class-name="border-t border-gray-100/50 dark:border-dark-700/50" :data="items" :show-header="false" table-layout="auto" class="element-data-table">
  <ElTableColumn :min-width="120" align="left">
    <template #default="{ row: user, $index: rowIndex }"><div class="max-w-[120px] truncate py-1 pl-6 text-gray-600 dark:text-gray-300" :title="user.email" >{{ user.email || `User #${user.user_id}` }}</div></template>
  </ElTableColumn>
  <ElTableColumn :min-width="120" align="right">
    <template #default="{ row: user, $index: rowIndex }"><div class="py-1 text-right text-gray-500 dark:text-gray-400" >{{ user.requests.toLocaleString() }}</div></template>
  </ElTableColumn>
  <ElTableColumn :min-width="120" align="right">
    <template #default="{ row: user, $index: rowIndex }"><div class="py-1 text-right text-gray-500 dark:text-gray-400" >{{ formatTokens(user.total_tokens) }}</div></template>
  </ElTableColumn>
  <ElTableColumn :min-width="120" align="right">
    <template #default="{ row: user, $index: rowIndex }"><div class="py-1 text-right text-green-600 dark:text-green-400" >
            ${{ formatCost(user.actual_cost) }}</div></template>
  </ElTableColumn>
  <ElTableColumn v-if="showAccountCost" :min-width="120" align="right">
    <template #default="{ row: user, $index: rowIndex }"><div class="py-1 text-right text-orange-500 dark:text-orange-400" >
            ${{ formatCost(user.account_cost) }}</div></template>
  </ElTableColumn>
  <ElTableColumn :min-width="120" align="right">
    <template #default="{ row: user, $index: rowIndex }"><div class="py-1 pr-1 text-right text-gray-400 dark:text-gray-500" >
            ${{ formatCost(user.cost) }}</div></template>
  </ElTableColumn>
</ElTable>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { UserBreakdownItem } from '@/types'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  items: UserBreakdownItem[]
  loading?: boolean
  showAccountCost?: boolean
}>(), {
  loading: false,
  showAccountCost: true,
})

const showAccountCost = computed(() => props.showAccountCost)

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(2)}M`
  if (value >= 1_000) return `${(value / 1_000).toFixed(2)}K`
  return value.toLocaleString()
}

const formatCost = (value: number | undefined | null): string => {
  if (value == null) return '0.0000'
  if (value >= 1000) return (value / 1000).toFixed(2) + 'K'
  if (value >= 1) return value.toFixed(2)
  if (value >= 0.01) return value.toFixed(3)
  return value.toFixed(4)
}
</script>
