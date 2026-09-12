<script setup lang="ts">
import { ElSwitch } from 'element-plus'
defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()
</script>
<template>
  <ElSwitch :model-value="modelValue" @update:model-value="emit('update:modelValue', Boolean($event))" />
</template>
