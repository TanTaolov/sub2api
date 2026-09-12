<script setup lang="ts">
import { computed } from 'vue'
import { ElPagination } from 'element-plus'
import { useMediaQuery } from '@vueuse/core'
import { getConfiguredTablePageSizeOptions, normalizeTablePageSize } from '@/utils/tablePreferences'
import { setPersistedPageSize } from '@/composables/usePersistedPageSize'
const props = withDefaults(defineProps<{
  total: number; page: number; pageSize: number; pageSizeOptions?: number[]; showPageSizeSelector?: boolean; showJump?: boolean
}>(), { pageSizeOptions: () => getConfiguredTablePageSizeOptions(), showPageSizeSelector: true, showJump: false })
const emit = defineEmits<{ 'update:page': [page: number]; 'update:pageSize': [pageSize: number] }>()
const isMobile = useMediaQuery('(max-width: 639px)')
const layout = computed(() => isMobile.value ? 'prev, pager, next' : ['total', ...(props.showPageSizeSelector ? ['sizes'] : []), 'prev', 'pager', 'next', ...(props.showJump ? ['jumper'] : [])].join(', '))
function changePageSize(value: number) {
  const size = normalizeTablePageSize(value)
  setPersistedPageSize(size)
  emit('update:pageSize', size)
  emit('update:page', 1)
}
</script>
<template>
  <div class="flex flex-wrap items-center justify-end gap-3 border-t border-gray-200 px-4 py-3 dark:border-dark-700">
    <ElPagination :current-page="page" :page-size="pageSize" :total="total" :page-sizes="pageSizeOptions"
      :layout="layout" :pager-count="isMobile ? 5 : 7" :size="isMobile ? 'small' : 'default'" background
      @update:current-page="emit('update:page', $event)" @update:page-size="changePageSize" />
  </div>
</template>
