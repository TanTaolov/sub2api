<script setup lang="ts">
import { computed } from 'vue'
import { ElTag } from 'element-plus'
const props = defineProps<{ status: string; label: string }>()
const type = computed(() => {
  if (['active', 'success'].includes(props.status)) return 'success'
  if (['disabled', 'inactive', 'warning'].includes(props.status)) return 'warning'
  if (['error', 'danger'].includes(props.status)) return 'danger'
  return 'info'
})
</script>
<template><ElTag :type="type" effect="light" size="small">{{ label }}</ElTag></template>
