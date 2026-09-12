<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElPopover, ElButton } from 'element-plus'
import Icon from '@/components/icons/Icon.vue'
const props = withDefaults(defineProps<{ content?: string; trigger?: 'hover' | 'click'; widthClass?: string }>(), { trigger: 'hover', widthClass: 'w-64' })
const { t } = useI18n()
const visible = ref(false)
const width = computed(() => ({ 'w-48': 192, 'w-56': 224, 'w-64': 256, 'w-72': 288, 'w-80': 320, 'w-96': 384 })[props.widthClass] ?? 256)
</script>
<template>
  <ElPopover v-model:visible="visible" :trigger="trigger === 'hover' ? ['hover', 'focus'] : 'click'" :width="width" placement="top" :show-after="100" :hide-after="150">
    <template #reference>
      <span class="ml-1 inline-flex items-center align-middle" tabindex="0" :aria-label="content || t('common.more')" @keydown.esc="visible = false">
        <slot name="trigger"><Icon name="infoCircle" size="sm" class="cursor-help text-gray-400" /></slot>
      </span>
    </template>
    <div class="text-xs leading-relaxed" @keydown.esc="visible = false">
      <ElButton v-if="trigger === 'click'" text size="small" :aria-label="t('common.close')" class="float-right" @click="visible = false"><Icon name="x" size="sm" /></ElButton>
      <slot>{{ content }}</slot>
    </div>
  </ElPopover>
</template>
