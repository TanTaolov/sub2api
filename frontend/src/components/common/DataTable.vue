<template>
<div v-if="!isDesktopViewport" class="space-y-3">
    <template v-if="loading">
      <div v-for="i in 5" :key="i" class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
        <div class="space-y-3">
          <div v-for="column in dataColumns" :key="column.key" class="flex justify-between">
            <div class="h-4 w-20 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
            <div class="h-4 w-32 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
          </div>
          <div v-if="hasActionsColumn" class="border-t border-gray-200 pt-3 dark:border-dark-700">
            <div class="h-8 w-full animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
          </div>
        </div>
      </div>
    </template>

    <template v-else-if="!data || data.length === 0">
      <div class="rounded-lg border border-gray-200 bg-white p-12 text-center dark:border-dark-700 dark:bg-dark-900">
        <slot name="empty">
          <div class="flex flex-col items-center">
            <Icon
              name="inbox"
              size="xl"
              class="mb-4 h-12 w-12 text-gray-400 dark:text-dark-500"
            />
            <p class="text-lg font-medium text-gray-900 dark:text-gray-100">
              {{ t('empty.noData') }}
            </p>
          </div>
        </slot>
      </div>
    </template>

    <template v-else>
      <div v-if="selectable" class="flex items-center justify-end gap-2 px-1">
        <ElementCheckbox :checked="allVisibleSelected" :indeterminate="someVisibleSelected" data-test="select-all-mobile" @change="toggleAllVisible(($event.target as HTMLInputElement).checked)" :class="[&quot;flex items-center gap-2 text-sm font-medium text-gray-600 dark:text-gray-300&quot;,&quot;&quot;]"><span>{{ t('common.selectAll') }}</span></ElementCheckbox>
      </div>
      <div
        v-for="(row, index) in sortedData"
        :key="resolveRowKey(row, index)"
        class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900"
        :class="{
          'cursor-pointer': clickableRows,
          'border-primary-300 bg-primary-50/40 dark:border-primary-700 dark:bg-primary-900/10': selectable && isRowSelected(row, index)
        }"
        @click="clickableRows && emit('rowClick', row)"
      >
        <div class="space-y-3">
          <div v-if="selectable" class="flex justify-end">
            <ElementCheckbox

              class=""
              :checked="isRowSelected(row, index)"
              :aria-label="getRowSelectionLabel(row, index)"
              data-test="select-row"
              @click.stop
              @change="toggleRowSelection(row, index, ($event.target as HTMLInputElement).checked)"
            />
          </div>
          <div
            v-for="column in dataColumns"
            :key="column.key"
            :data-field="column.key"
            class="flex min-w-0 items-start justify-between gap-4"
          >
            <span class="text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
              {{ column.label }}
            </span>
            <div class="min-w-0 max-w-full text-right text-sm text-gray-900 dark:text-gray-100">
              <slot :name="`cell-${column.key}`" :row="row" :value="row[column.key]" :expanded="actionsExpanded">
                {{ column.formatter ? column.formatter(row[column.key], row) : row[column.key] }}
              </slot>
            </div>
          </div>
          <div v-if="hasActionsColumn" class="border-t border-gray-200 pt-3 dark:border-dark-700">
            <slot name="cell-actions" :row="row" :value="row['actions']" :expanded="actionsExpanded"></slot>
          </div>
        </div>
      </div>
    </template>
  </div>
<div v-else ref="tableWrapperRef" class="table-wrapper element-table-wrapper" :class="{ 'actions-expanded': actionsExpanded }">
<ElTable v-if="!shouldVirtualize" :data="loading ? [] : sortedData" :row-key="(row: any) => resolveRowKey(row, sortedData.indexOf(row))"
      class="element-data-table" table-layout="auto" :class="{ 'cursor-pointer': clickableRows }"
      @row-click="(row: any) => clickableRows && emit('rowClick', row)">
      <ElTableColumn v-if="selectable" :width="48" fixed="left">
        <template #header>
          <ElementCheckbox :checked="allVisibleSelected" :indeterminate="someVisibleSelected"
            :aria-label="t('common.selectAll')" data-test="select-all"
            @change="toggleAllVisible(($event.target as HTMLInputElement).checked)" />
        </template>
        <template #default="{ row, $index }">
          <ElementCheckbox :checked="isRowSelected(row, $index)" :aria-label="getRowSelectionLabel(row, $index)"
            data-test="select-row" @click.stop @change="toggleRowSelection(row, $index, ($event.target as HTMLInputElement).checked)" />
        </template>
      </ElTableColumn>
      <ElTableColumn v-for="(column, columnIndex) in columns" :key="column.key" :prop="column.key" :label="column.label"
        :min-width="column.key === 'actions' ? 180 : 140" :class-name="column.class"
        :fixed="column.key === 'actions' && stickyActionsColumn ? 'right' : columnIndex === 0 && stickyFirstColumn ? 'left' : undefined">
        <template #header>
          <div class="flex items-center gap-1">
            <ElButton v-if="column.sortable" text size="small" :aria-label="column.label"
              :aria-sort="getColumnAriaSort(column.key)" @click="handleSort(column.key)">
              <slot :name="'header-' + column.key" :column="column" :sort-key="sortKey" :sort-order="sortOrder">{{ column.label }}</slot>
              <Icon v-if="sortKey === column.key" name="arrowUp" size="xs" :class="{ 'rotate-180': sortOrder === 'desc' }" />
            </ElButton>
            <slot v-else :name="'header-' + column.key" :column="column" :sort-key="sortKey" :sort-order="sortOrder">{{ column.label }}</slot>
            <ElButton v-if="column.key === 'actions' && expandableActions" text size="small" :aria-label="column.label"
              :aria-expanded="actionsExpanded" @click="actionsExpanded = !actionsExpanded">
              <Icon :name="actionsExpanded ? 'chevronLeft' : 'chevronRight'" size="xs" />
            </ElButton>
          </div>
        </template>
        <template #default="{ row }">
          <slot :name="'cell-' + column.key" :row="row" :value="row[column.key]" :expanded="actionsExpanded">
            {{ column.formatter ? column.formatter(row[column.key], row) : row[column.key] }}
          </slot>
        </template>
      </ElTableColumn>
      <template #empty>
        <ElSkeleton v-if="loading" :rows="5" animated class="p-6" />
        <slot v-else name="empty"><ElEmpty :description="t('empty.noData')" :image-size="80" /></slot>
      </template>
    </ElTable>
<div v-else class="element-virtual-table"><ElAutoResizer><template #default="{ width, height }"><ElTableV2 ref="virtualTable" :key="virtualTableRevision" :columns="virtualColumns" :data="virtualRows" row-key="key" :width="width" :height="height" :row-height="estimateRowHeight ?? 56" :estimated-row-height="estimateRowHeight ?? 56" :cache="overscan ?? 5" :row-event-handlers="{ onClick: ({ rowData }: { rowData: { data: any } }) => clickableRows && emit('rowClick', rowData.data) }" /></template></ElAutoResizer></div>
</div>
</template>
<script setup lang="ts">
import { computed, ref, onMounted, watch, h, useSlots } from 'vue'
import { useMediaQuery } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { ElTable, ElTableColumn, ElTableV2, ElAutoResizer, ElCheckbox, ElButton, ElEmpty, ElSkeleton, TableV2FixedDir, type Column as VirtualColumn } from 'element-plus'
import type { Column } from './types'
import Icon from '@/components/icons/Icon.vue'
const { t } = useI18n()
const slots = useSlots()
const isDesktopViewport = useMediaQuery('(min-width: 768px)')
const emit = defineEmits<{
  sort: [key: string, order: 'asc' | 'desc']; rowClick: [row: any];
  'update:selectedKeys': [keys: Array<string | number>]; selectionChange: [keys: Array<string | number>]
}>()
const tableWrapperRef = ref<HTMLElement | null>(null)
const virtualTable = ref<InstanceType<typeof ElTableV2>>()
const virtualTableRevision = ref(0)
interface Props {
  columns: Column[]
  data: any[]
  loading?: boolean
  stickyFirstColumn?: boolean
  stickyActionsColumn?: boolean
  expandableActions?: boolean
  actionsCount?: number // 操作按钮总数，用于判断是否需要展开功能
  rowKey?: string | ((row: any) => string | number)
  /**
   * Default sort configuration (only applied when there is no persisted sort state)
   */
  defaultSortKey?: string
  defaultSortOrder?: 'asc' | 'desc'
  /**
   * Persist sort state (key + order) to localStorage using this key.
   * If provided, DataTable will load the stored sort state on mount.
   */
  sortStorageKey?: string
  /**
   * Enable server-side sorting mode. When true, clicking sort headers
   * will emit 'sort' events instead of performing client-side sorting.
   */
  serverSideSort?: boolean
  /** Emit 'rowClick' on row/card click and show pointer cursor (interactive cells should @click.stop) */
  clickableRows?: boolean
  /** Estimated row height in px for the virtualizer (default 56) */
  estimateRowHeight?: number
  /** Number of rows to render beyond the visible area (default 5) */
  overscan?: number
  /**
   * Only virtualize when the row count exceeds this threshold (default 100).
   * Smaller lists render in full, avoiding the scroll-compensation jank caused by
   * estimated-vs-actual row heights when rows have variable height.
   */
  virtualizeThreshold?: number
  /** Enable controlled row selection. Stable row keys are strongly recommended. */
  selectable?: boolean
  /** Selected row keys. Keys outside the current data page are preserved. */
  selectedKeys?: Array<string | number>
  /** Accessible label for a row selection checkbox. */
  selectionLabel?: string | ((row: any) => string)
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  stickyFirstColumn: true,
  stickyActionsColumn: true,
  expandableActions: true,
  defaultSortOrder: 'asc',
  serverSideSort: false,
  selectable: false,
  selectedKeys: () => []
})


const sortKey = ref<string>('')
const sortOrder = ref<'asc' | 'desc'>('asc')
const actionsExpanded = ref(false)

type PersistedSortState = {
  key: string
  order: 'asc' | 'desc'
}

const collator = new Intl.Collator(undefined, {
  numeric: true,
  sensitivity: 'base'
})

const getSortableKeys = () => {
  const keys = new Set<string>()
  for (const col of props.columns) {
    if (col.sortable) keys.add(col.key)
  }
  return keys
}

const normalizeSortKey = (candidate: string) => {
  if (!candidate) return ''
  const sortableKeys = getSortableKeys()
  return sortableKeys.has(candidate) ? candidate : ''
}

const normalizeSortOrder = (candidate: any): 'asc' | 'desc' => {
  return candidate === 'desc' ? 'desc' : 'asc'
}

const readPersistedSortState = (): PersistedSortState | null => {
  if (!props.sortStorageKey) return null
  try {
    const raw = localStorage.getItem(props.sortStorageKey)
    if (!raw) return null
    const parsed = JSON.parse(raw) as Partial<PersistedSortState>
    const key = normalizeSortKey(typeof parsed.key === 'string' ? parsed.key : '')
    if (!key) return null
    return { key, order: normalizeSortOrder(parsed.order) }
  } catch (e) {
    console.error('[DataTable] Failed to read persisted sort state:', e)
    return null
  }
}

const writePersistedSortState = (state: PersistedSortState) => {
  if (!props.sortStorageKey) return
  try {
    localStorage.setItem(props.sortStorageKey, JSON.stringify(state))
  } catch (e) {
    console.error('[DataTable] Failed to persist sort state:', e)
  }
}

const resolveInitialSortState = (): PersistedSortState | null => {
  const persisted = readPersistedSortState()
  if (persisted) return persisted

  const key = normalizeSortKey(props.defaultSortKey || '')
  if (!key) return null
  return { key, order: normalizeSortOrder(props.defaultSortOrder) }
}

const applySortState = (state: PersistedSortState | null) => {
  if (!state) return
  sortKey.value = state.key
  sortOrder.value = state.order
}



const getColumnAriaSort = (key: string) => {
  if (sortKey.value !== key) return 'none'
  return sortOrder.value === 'asc' ? 'ascending' : 'descending'
}

const getHeaderContentAlignmentClass = (column: Column) => {
  const className = column.class || ''
  if (className.includes('text-center')) return 'justify-center'
  if (className.includes('text-right')) return 'justify-end'
  return 'justify-start'
}

const isNullishOrEmpty = (value: any) => value === null || value === undefined || value === ''

const toFiniteNumberOrNull = (value: any): number | null => {
  if (typeof value === 'number') return Number.isFinite(value) ? value : null
  if (typeof value === 'boolean') return value ? 1 : 0
  if (typeof value === 'string') {
    const trimmed = value.trim()
    if (!trimmed) return null
    const n = Number(trimmed)
    return Number.isFinite(n) ? n : null
  }
  return null
}

const toSortableString = (value: any): string => {
  if (value === null || value === undefined) return ''
  if (typeof value === 'string') return value
  if (typeof value === 'number' || typeof value === 'boolean') return String(value)
  if (value instanceof Date) return value.toISOString()
  try {
    return JSON.stringify(value)
  } catch {
    return String(value)
  }
}

const compareSortValues = (a: any, b: any): number => {
  const aEmpty = isNullishOrEmpty(a)
  const bEmpty = isNullishOrEmpty(b)
  if (aEmpty && bEmpty) return 0
  if (aEmpty) return 1
  if (bEmpty) return -1

  const aNum = toFiniteNumberOrNull(a)
  const bNum = toFiniteNumberOrNull(b)
  if (aNum !== null && bNum !== null) {
    if (aNum === bNum) return 0
    return aNum < bNum ? -1 : 1
  }

  const aStr = toSortableString(a)
  const bStr = toSortableString(b)
  const res = collator.compare(aStr, bStr)
  if (res === 0) return 0
  return res < 0 ? -1 : 1
}
const resolveStableRowKey = (row: any): string | number | undefined => {
  if (typeof props.rowKey === 'function') {
    const key = props.rowKey(row)
    return key ?? undefined
  }
  if (typeof props.rowKey === 'string' && props.rowKey) {
    const key = row?.[props.rowKey]
    return key ?? undefined
  }
  const key = row?.id
  return key ?? undefined
}

const resolveRowKey = (row: any, index: number) => resolveStableRowKey(row) ?? index

const dataColumns = computed(() => props.columns.filter((column) => column.key !== 'actions'))
const columnsSignature = computed(() =>
  props.columns.map((column) => `${column.key}:${column.sortable ? '1' : '0'}`).join('|')
)

const handleSort = (key: string) => {
  let newOrder: 'asc' | 'desc' = 'asc'
  if (sortKey.value === key) {
    newOrder = sortOrder.value === 'asc' ? 'desc' : 'asc'
  }

  if (props.serverSideSort) {
    // Server-side sort mode: emit event and update internal state for UI feedback
    sortKey.value = key
    sortOrder.value = newOrder
    emit('sort', key, newOrder)
  } else {
    // Client-side sort mode: just update internal state
    sortKey.value = key
    sortOrder.value = newOrder
  }
}

const sortedData = computed(() => {
  // Server-side sort mode: return data as-is (server handles sorting)
  if (props.serverSideSort || !sortKey.value || !props.data) return props.data

  const key = sortKey.value
  const order = sortOrder.value

  // Stable sort (tie-break with original index) to avoid jitter when values are equal.
  return props.data
    .map((row, index) => ({ row, index }))
    .sort((a, b) => {
      const cmp = compareSortValues(a.row?.[key], b.row?.[key])
      if (cmp !== 0) return order === 'asc' ? cmp : -cmp
      return a.index - b.index
    })
    .map(item => item.row)
})

const selectedKeySet = computed(() => new Set(props.selectedKeys))
const visibleRowKeys = computed(() =>
  (sortedData.value ?? []).map((row, index) => resolveRowKey(row, index))
)
const allVisibleSelected = computed(() =>
  visibleRowKeys.value.length > 0
  && visibleRowKeys.value.every((key) => selectedKeySet.value.has(key))
)
const someVisibleSelected = computed(() => {
  if (allVisibleSelected.value) return false
  return visibleRowKeys.value.some((key) => selectedKeySet.value.has(key))
})

const emitSelection = (next: Set<string | number>) => {
  const keys = Array.from(next)
  emit('update:selectedKeys', keys)
  emit('selectionChange', keys)
}

const isRowSelected = (row: any, index: number) =>
  selectedKeySet.value.has(resolveRowKey(row, index))

const getRowSelectionLabel = (row: any, index: number) => {
  if (typeof props.selectionLabel === 'function') return props.selectionLabel(row)
  if (props.selectionLabel) return props.selectionLabel
  return `${t('common.selectOption')} ${resolveRowKey(row, index)}`
}

const toggleRowSelection = (row: any, index: number, checked: boolean) => {
  const next = new Set(props.selectedKeys)
  const key = resolveRowKey(row, index)
  if (checked) next.add(key)
  else next.delete(key)
  emitSelection(next)
}

const toggleAllVisible = (checked: boolean) => {
  const next = new Set(props.selectedKeys)
  for (const key of visibleRowKeys.value) {
    if (checked) next.add(key)
    else next.delete(key)
  }
  emitSelection(next)
}


const shouldVirtualize = computed(() => isDesktopViewport.value && !props.loading && props.data.length > (props.virtualizeThreshold ?? 100))
const hasActionsColumn = computed(() => props.columns.some(column => column.key === 'actions'))
const virtualRows = computed(() => sortedData.value.map((data, index) => ({ data, index, key: resolveRowKey(data, index) })))
function getColumnWidth(column: Column) {
  const explicitWidth = column.class?.match(/(?:min-)?w-\[(\d+)px\]/)?.[1]
  if (explicitWidth) return Number(explicitWidth)
  if (column.key === 'actions') return actionsExpanded.value ? Math.max(220, (props.actionsCount ?? 4) * 38) : 180
  return 180
}
function renderHeader(column: Column) {
  const content = () => slots['header-' + column.key]?.({ column, sortKey: sortKey.value, sortOrder: sortOrder.value }) ?? column.label
  const children = [column.sortable
    ? h(ElButton, { text: true, size: 'small', 'aria-label': column.label, onClick: () => handleSort(column.key) }, () => [content(), sortKey.value === column.key ? (sortOrder.value === 'asc' ? ' ↑' : ' ↓') : ''])
    : content()]
  if (column.key === 'actions' && props.expandableActions) children.push(h(ElButton, { text: true, size: 'small', 'aria-label': column.label, 'aria-expanded': actionsExpanded.value, onClick: () => actionsExpanded.value = !actionsExpanded.value }, () => h(Icon, { name: actionsExpanded.value ? 'chevronLeft' : 'chevronRight', size: 'xs' })))
  return h('div', { class: ['flex items-center gap-1', getHeaderContentAlignmentClass(column)], 'aria-sort': column.sortable ? getColumnAriaSort(column.key) : undefined }, children)
}
const virtualColumns = computed<VirtualColumn[]>(() => {
  const columns: VirtualColumn[] = props.columns.map((column, index) => ({
    key: column.key, dataKey: column.key, title: column.label, width: getColumnWidth(column),
    fixed: column.key === 'actions' && props.stickyActionsColumn ? TableV2FixedDir.RIGHT : index === 0 && props.stickyFirstColumn ? TableV2FixedDir.LEFT : undefined,
    headerCellRenderer: () => renderHeader(column),
    cellRenderer: ({ rowData }: { rowData: { data: any; index: number } }) => h('div', { class: ['w-full min-w-0 py-2', column.class] }, slots['cell-' + column.key]?.({ row: rowData.data, value: rowData.data[column.key], expanded: actionsExpanded.value }) ?? String(column.formatter ? column.formatter(rowData.data[column.key], rowData.data) : rowData.data[column.key] ?? ''))
  }))
  if (!props.selectable) return columns
  return [{
    key: '__selection', dataKey: '__selection', title: '', width: 48, fixed: TableV2FixedDir.LEFT,
    headerCellRenderer: () => h(ElCheckbox, { modelValue: allVisibleSelected.value, indeterminate: someVisibleSelected.value, 'aria-label': t('common.selectAll'), 'data-test': 'select-all', onChange: (value: string | number | boolean) => toggleAllVisible(Boolean(value)) }),
    cellRenderer: ({ rowData }: { rowData: { data: any; index: number } }) => h(ElCheckbox, { modelValue: isRowSelected(rowData.data, rowData.index), 'aria-label': getRowSelectionLabel(rowData.data, rowData.index), 'data-test': 'select-row', onClick: (event: Event) => event.stopPropagation(), onChange: (value: string | number | boolean) => toggleRowSelection(rowData.data, rowData.index, Boolean(value)) })
  }, ...columns]
})
// Init + keep persisted sort state consistent with current columns
const didInitSort = ref(false)

onMounted(() => {
  const initial = resolveInitialSortState()
  applySortState(initial)
  didInitSort.value = true
})

watch(
  columnsSignature,
  () => {
    // If current sort key is no longer sortable/visible, fall back to default/persisted.
    const normalized = normalizeSortKey(sortKey.value)
    if (!sortKey.value) {
      const initial = resolveInitialSortState()
      applySortState(initial)
      return
    }

    if (!normalized) {
      const fallback = resolveInitialSortState()
      if (fallback) {
        applySortState(fallback)
      } else {
        sortKey.value = ''
        sortOrder.value = 'asc'
      }
    }
  },
  { flush: 'post' }
)

watch(
  [sortKey, sortOrder],
  ([nextKey, nextOrder]) => {
    if (!didInitSort.value) return
    if (!props.sortStorageKey) return
    const key = normalizeSortKey(nextKey)
    if (!key) return
    writePersistedSortState({ key, order: normalizeSortOrder(nextOrder) })
  },
  { flush: 'post' }
)


// 分页、排序和列宽变化时重建虚拟表，清除上一页的动态行高缓存。
watch([virtualRows, virtualColumns], () => { virtualTableRevision.value += 1 }, { flush: 'pre' })
defineExpose({ virtualizer: virtualTable, shouldVirtualize, sortedData, resolveRowKey, tableWrapperEl: tableWrapperRef })
</script>
<style scoped>
.element-table-wrapper { width: 100%; min-width: 0; overflow: auto; }
.element-virtual-table { height: clamp(320px, 65vh, 760px); width: 100%; min-width: 0; }
</style>
