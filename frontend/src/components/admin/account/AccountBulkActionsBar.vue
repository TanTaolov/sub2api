<template>
  <div class="mb-4 flex items-center justify-between rounded-lg bg-primary-50 p-3 dark:bg-primary-900/20">
    <div class="flex flex-wrap items-center gap-2">
      <span v-if="allResultsSelected" class="text-sm font-medium text-primary-900 dark:text-primary-100">
        {{ t('admin.accounts.bulkActions.selectedAll', { count: selectedIds.length }) }}
      </span>
      <span v-else-if="selectedIds.length > 0" class="text-sm font-medium text-primary-900 dark:text-primary-100">
        {{ t('admin.accounts.bulkActions.selected', { count: selectedIds.length }) }}
      </span>
      <span v-else class="text-sm font-medium text-primary-900 dark:text-primary-100">
        {{ t('admin.accounts.bulkEdit.title') }}
      </span>
      <template v-if="selectedIds.length > 0">
        <ElButton text
          @click="$emit('select-page')"
          class="text-xs font-medium text-primary-700 hover:text-primary-800 dark:text-primary-300 dark:hover:text-primary-200"
        >
          {{ t('admin.accounts.bulkActions.selectCurrentPage') }}
        </ElButton>
      </template>
      <template v-if="!allResultsSelected && totalResults > selectedIds.length">
        <span v-if="selectedIds.length > 0" class="text-gray-300 dark:text-primary-800">•</span>
        <ElButton text
          :disabled="selectingAll"
          @click="$emit('select-all-results')"
          class="text-xs font-medium text-primary-700 hover:text-primary-800 disabled:cursor-not-allowed disabled:opacity-60 dark:text-primary-300 dark:hover:text-primary-200"
        >
          {{
            selectingAll
              ? t('admin.accounts.bulkActions.selectingAll')
              : t('admin.accounts.bulkActions.selectAllResults', { count: totalResults })
          }}
        </ElButton>
      </template>
      <template v-if="selectedIds.length > 0">
        <span class="text-gray-300 dark:text-primary-800">•</span>
        <ElButton text
          @click="$emit('clear')"
          class="text-xs font-medium text-primary-700 hover:text-primary-800 dark:text-primary-300 dark:hover:text-primary-200"
        >
          {{ t('admin.accounts.bulkActions.clear') }}
        </ElButton>
      </template>
    </div>
    <div class="flex gap-2">
      <template v-if="selectedIds.length > 0">
        <ElButton type="danger" size="small" @click="$emit('delete')" class="">{{ t('admin.accounts.bulkActions.delete') }}</ElButton>
        <ElButton size="small" @click="$emit('reset-status')" class="">{{ t('admin.accounts.bulkActions.resetStatus') }}</ElButton>
        <ElButton size="small" @click="$emit('refresh-token')" class="">{{ t('admin.accounts.bulkActions.refreshToken') }}</ElButton>
        <ElButton size="small" @click="$emit('probe-upstream-billing')" class="">{{ t('admin.accounts.bulkActions.probeUpstreamBilling') }}</ElButton>
        <ElButton size="small" @click="$emit('toggle-schedulable', true)" class="btn-success">{{ t('admin.accounts.bulkActions.enableScheduling') }}</ElButton>
        <ElButton size="small" @click="$emit('toggle-schedulable', false)" class="btn-warning">{{ t('admin.accounts.bulkActions.disableScheduling') }}</ElButton>
        <ElButton type="primary" size="small" @click="$emit('edit-selected')" class="">{{ t('admin.accounts.bulkActions.edit') }}</ElButton>
      </template>
      <ElButton type="primary" size="small" @click="$emit('edit-filtered')" class="">
        {{ t('admin.accounts.bulkEdit.submit') }}
      </ElButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
  selectedIds: number[]
  totalResults: number
  selectingAll: boolean
  allResultsSelected: boolean
}>()

defineEmits([
  'delete',
  'edit-selected',
  'edit-filtered',
  'clear',
  'select-page',
  'select-all-results',
  'toggle-schedulable',
  'reset-status',
  'refresh-token',
  'probe-upstream-billing'
])

const { t } = useI18n()
</script>
