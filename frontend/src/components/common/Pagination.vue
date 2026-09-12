<template>
  <div class="element-pagination">
    <el-pagination
      background
      :small="isMobile"
      :current-page="page"
      :page-size="pageSize"
      :page-sizes="pageSizeSelectOptions"
      :total="total"
      :layout="paginationLayout"
      :pager-count="isMobile ? 5 : 7"
      @update:current-page="handlePageChange"
      @update:page-size="handlePageSizeChange"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { getConfiguredTablePageSizeOptions, normalizeTablePageSize } from '@/utils/tablePreferences'
import { setPersistedPageSize } from '@/composables/usePersistedPageSize'

interface Props {
  total: number
  page: number
  pageSize: number
  pageSizeOptions?: number[]
  showPageSizeSelector?: boolean
  showJump?: boolean
}

interface Emits {
  (e: 'update:page', page: number): void
  (e: 'update:pageSize', pageSize: number): void
}

const props = withDefaults(defineProps<Props>(), {
  pageSizeOptions: () => getConfiguredTablePageSizeOptions(),
  showPageSizeSelector: true,
  showJump: false
})

const emit = defineEmits<Emits>()
const isMobile = ref(false)

const pageSizeSelectOptions = computed(() => {
  return Array.from(
    new Set([
      ...props.pageSizeOptions,
      ...getConfiguredTablePageSizeOptions(),
      normalizeTablePageSize(props.pageSize)
    ])
  ).sort((a, b) => a - b)
})

const paginationLayout = computed(() => {
  if (isMobile.value) return 'prev, pager, next'

  const sections = ['total']
  if (props.showPageSizeSelector) sections.push('sizes')
  sections.push('prev', 'pager', 'next')
  if (props.showJump) sections.push('jumper')
  return sections.join(', ')
})

function updateViewportMode() {
  isMobile.value = window.innerWidth < 640
}

function handlePageChange(nextPage: number) {
  if (nextPage !== props.page) {
    emit('update:page', nextPage)
  }
}

function handlePageSizeChange(nextPageSize: number) {
  const normalizedPageSize = normalizeTablePageSize(nextPageSize)
  if (normalizedPageSize === props.pageSize) return

  setPersistedPageSize(normalizedPageSize)
  emit('update:pageSize', normalizedPageSize)
}

onMounted(() => {
  updateViewportMode()
  window.addEventListener('resize', updateViewportMode)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateViewportMode)
})
</script>

<style scoped lang="scss">
.element-pagination {
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid var(--el-border-color-light);
  padding: 0.875rem 1rem;

  :deep(.el-pagination) {
    justify-content: flex-end;
  }
}

@media (max-width: 639px) {
  .element-pagination {
    justify-content: center;
    padding: 0.75rem;

    :deep(.el-pagination) {
      justify-content: center;
    }
  }
}
</style>
