<template>
  <div class="element-data-table">
    <el-table
      :data="data"
      :row-key="rowKey"
      :default-sort="defaultSort"
      :empty-text="emptyText"
      v-loading="loading"
      class="w-full"
      @sort-change="handleSortChange"
    >
      <el-table-column
        v-for="column in columns"
        :key="column.key"
        :prop="column.key"
        :label="column.label"
        :min-width="getColumnMinWidth(column)"
        :sortable="column.sortable ? 'custom' : false"
        :sort-orders="sortOrders"
        :class-name="column.class"
        show-overflow-tooltip
      >
        <template #default="{ row }">
          <slot :name="`cell-${column.key}`" :row="row" :value="row[column.key]">
            {{ column.formatter ? column.formatter(row[column.key], row) : row[column.key] ?? '-' }}
          </slot>
        </template>
      </el-table-column>

      <template #empty>
        <slot name="empty">
          <span>{{ emptyText }}</span>
        </slot>
      </template>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Column } from './types'

type SortOrder = 'ascending' | 'descending'
type PersistedSortState = {
  key: string
  order: 'asc' | 'desc'
}

interface Props {
  columns: Column[]
  data: any[]
  loading?: boolean
  rowKey?: string | ((row: any) => string | number)
  defaultSortKey?: string
  defaultSortOrder?: 'asc' | 'desc'
  sortStorageKey?: string
  emptyText?: string
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  rowKey: 'id',
  defaultSortOrder: 'asc',
  emptyText: ''
})

const emit = defineEmits<{
  sort: [key: string, order: 'asc' | 'desc']
}>()

const { t } = useI18n()
const restoredSort = ref<PersistedSortState | null>(readPersistedSortState())
const sortOrders: SortOrder[] = ['ascending', 'descending']

const emptyText = computed(() => props.emptyText || t('empty.noData'))

const defaultSort = computed(() => {
  const state = restoredSort.value
  const key = state?.key || props.defaultSortKey
  if (!key) return undefined

  const order = state?.order || props.defaultSortOrder
  return {
    prop: key,
    order: order === 'desc' ? 'descending' : 'ascending'
  }
})

function isSortableColumn(key: string): boolean {
  return props.columns.some((column) => column.key === key && column.sortable)
}

function getColumnMinWidth(column: Column): number {
  if (column.key.includes('created_at')) return 172
  if (column.key.includes('amount') || column.key.includes('quota')) return 132
  return 150
}

function readPersistedSortState(): PersistedSortState | null {
  if (!props.sortStorageKey) return null

  try {
    const value = localStorage.getItem(props.sortStorageKey)
    if (!value) return null
    const parsed = JSON.parse(value) as Partial<PersistedSortState>
    if (
      typeof parsed.key !== 'string' ||
      !isSortableColumn(parsed.key) ||
      (parsed.order !== 'asc' && parsed.order !== 'desc')
    ) {
      return null
    }
    return { key: parsed.key, order: parsed.order }
  } catch {
    return null
  }
}

function persistSortState(state: PersistedSortState) {
  if (!props.sortStorageKey) return

  try {
    localStorage.setItem(props.sortStorageKey, JSON.stringify(state))
  } catch {
    // Sorting remains functional when storage is unavailable.
  }
}

function handleSortChange(event: { prop: string | null; order: SortOrder | null }) {
  if (!event.prop || !event.order || !isSortableColumn(event.prop)) return

  const state: PersistedSortState = {
    key: event.prop,
    order: event.order === 'descending' ? 'desc' : 'asc'
  }
  restoredSort.value = state
  persistSortState(state)
  emit('sort', state.key, state.order)
}

</script>

<style scoped lang="scss">
.element-data-table {
  min-width: 0;

  :deep(.el-table__inner-wrapper::before) {
    display: none;
  }

  :deep(.el-table__header-wrapper th.el-table__cell) {
    height: 3.25rem;
    font-size: 0.75rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  :deep(.el-table__cell) {
    padding: 0.875rem 0;
  }
}

@media (max-width: 767px) {
  .element-data-table {
    overflow-x: auto;

    :deep(.el-table) {
      min-width: 46rem;
    }
  }
}
</style>
