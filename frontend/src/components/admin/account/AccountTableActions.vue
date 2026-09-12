<template>
  <div class="flex flex-wrap items-center gap-3">
    <slot name="before"></slot>
    <ElButton @click="$emit('refresh')" :disabled="loading" class="">
      <Icon name="refresh" size="md" :class="[loading ? 'animate-spin' : '']" />
    </ElButton>
    <slot name="after"></slot>
    <slot name="beforeCreate"></slot>
    <ElButton type="primary" @click="$emit('create')" class="">{{ t('admin.accounts.createAccount') }}</ElButton>
    <slot name="afterCreate"></slot>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps(['loading'])
defineEmits(['refresh', 'create'])

const { t } = useI18n()
</script>
