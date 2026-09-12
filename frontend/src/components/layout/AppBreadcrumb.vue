<template>
  <el-breadcrumb v-if="items.length > 0" separator="/" class="app-breadcrumb">
    <el-breadcrumb-item
      v-for="(item, index) in items"
      :key="`${item.label}-${index}`"
      :to="item.to"
    >
      {{ item.label }}
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { resolveRouteMetaKeys } from '@/router/title'
import { useAppStore } from '@/stores/app'
import { resolveSiteBillingMode } from '@/utils/siteBillingMode'

type BreadcrumbItem = {
  label: string
  to?: string
}

const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()

const routeMetaKeys = computed(() => resolveRouteMetaKeys(route, {
  billingMode: resolveSiteBillingMode(appStore.cachedPublicSettings)
}))

const currentLabel = computed(() => {
  if (routeMetaKeys.value.titleKey) {
    return t(routeMetaKeys.value.titleKey)
  }
  return (route.meta.title as string) || ''
})

const items = computed<BreadcrumbItem[]>(() => {
  const configuredItems = route.meta.breadcrumbs ?? []
  const normalizedItems = configuredItems
    .map((item) => ({
      label: item.labelKey ? t(item.labelKey) : item.label || '',
      to: item.to
    }))
    .filter((item) => item.label.trim())

  const lastItem = normalizedItems.at(-1)
  if (currentLabel.value && lastItem?.label !== currentLabel.value) {
    normalizedItems.push({ label: currentLabel.value })
  }

  return normalizedItems
})
</script>

<style scoped lang="scss">
.app-breadcrumb {
  :deep(.el-breadcrumb__inner),
  :deep(.el-breadcrumb__inner a) {
    color: #6b7280;
    font-size: 0.75rem;
    font-weight: 400;
  }

  :deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
    color: #374151;
    font-weight: 500;
  }

  :deep(.el-breadcrumb__separator) {
    color: #9ca3af;
    margin: 0 0.5rem;
  }
}

:global(.dark) .app-breadcrumb {
  :deep(.el-breadcrumb__inner),
  :deep(.el-breadcrumb__inner a) {
    color: #9ca3af;
  }

  :deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
    color: #e5e7eb;
  }

  :deep(.el-breadcrumb__separator) {
    color: #6b7280;
  }
}
</style>
