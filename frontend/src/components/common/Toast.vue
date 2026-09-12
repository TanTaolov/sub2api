<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'
import { ElNotification, type NotificationHandle } from 'element-plus'
import { useAppStore } from '@/stores/app'
const appStore = useAppStore()
const notifications = new Map<string, NotificationHandle>()
watch(() => [...appStore.toasts], (toasts) => {
  const activeIds = new Set(toasts.map(toast => toast.id))
  for (const [id, notification] of notifications) {
    if (!activeIds.has(id)) { notifications.delete(id); notification.close() }
  }
  for (const toast of toasts) {
    if (notifications.has(toast.id)) continue
    notifications.set(toast.id, ElNotification({
      title: toast.title ?? '', message: toast.message, type: toast.type, duration: 0,
      position: 'top-right', onClose: () => { notifications.delete(toast.id); appStore.hideToast(toast.id) }
    }))
  }
}, { immediate: true })
onBeforeUnmount(() => { for (const notification of notifications.values()) notification.close(); notifications.clear() })
</script>
<template><span class="sr-only" aria-live="polite" aria-atomic="true">{{ appStore.toasts[appStore.toasts.length - 1]?.message }}</span></template>
