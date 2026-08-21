<script setup lang="ts">
defineProps<{
  colorClass: string
  tooltip?: string
  current: string | number
  max: string | number
  suffix?: string
}>()
</script>

<template>
  <span
    :class="[
      'inline-flex items-center gap-1.5 rounded-lg px-2 py-1 text-[11px] font-medium leading-tight transition-all',
      colorClass
    ]"
    :title="tooltip"
  >
    <slot />
    <span class="font-mono font-semibold">{{ current }}</span>
    <span class="opacity-40">/</span>
    <span class="font-mono">{{ max }}</span>
    <span v-if="suffix" class="text-[10px] opacity-60 ml-0.5">{{ suffix }}</span>
  </span>
</template>
