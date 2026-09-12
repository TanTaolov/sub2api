<script setup lang="ts">
import { computed, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
import { ElEmpty, ElButton } from 'element-plus'
import Icon from '@/components/icons/Icon.vue'
const { t } = useI18n()
const props = withDefaults(defineProps<{
  icon?: Component | string; title?: string; description?: string; actionText?: string; actionTo?: string | object; actionIcon?: boolean; message?: string
}>(), { description: '', actionIcon: true })
const emit = defineEmits<{ action: [] }>()
const displayTitle = computed(() => props.title || props.message || t('common.noData'))
</script>
<template>
  <ElEmpty :image-size="96" class="empty-state">
    <template v-if="icon || $slots.icon" #image><slot name="icon"><component :is="icon" class="mx-auto h-16 w-16 text-gray-400" aria-hidden="true" /></slot></template>
    <template #description><h3 class="empty-state-title">{{ displayTitle }}</h3><p v-if="description" class="empty-state-description">{{ description }}</p></template>
    <slot name="action">
      <RouterLink v-if="actionText && actionTo" :to="actionTo" class="el-button el-button--primary"><Icon v-if="actionIcon" name="plus" size="md" class="mr-2" />{{ actionText }}</RouterLink>
      <ElButton v-else-if="actionText" type="primary" @click="emit('action')"><Icon v-if="actionIcon" name="plus" size="md" class="mr-2" />{{ actionText }}</ElButton>
    </slot>
  </ElEmpty>
</template>
