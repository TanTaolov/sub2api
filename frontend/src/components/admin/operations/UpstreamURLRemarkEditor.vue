<template>
  <div class="space-y-2 border-t py-3 dark:border-dark-700" data-testid="url-remark-editor">
    <p class="break-all font-medium">{{ value.upstream }}</p>
    <p class="whitespace-pre-wrap break-words text-sm"><span v-if="editing" class="text-gray-500">{{ t('profitOps.remarkCurrent') }}：</span>{{ current.remark || t('profitOps.remarkEmpty') }}</p>
    <template v-if="editing">
      <label class="block text-sm">{{ t('profitOps.remarkLabel') }}
        <textarea v-model="draft" :disabled="busy" rows="3" class="input mt-1 w-full" :aria-invalid="tooLong" />
      </label>
      <p class="text-xs" :class="tooLong ? 'text-red-600' : 'text-gray-500'">{{ Array.from(draft.trim()).length }} / 500</p>
      <p v-if="tooLong" role="alert" class="text-sm text-red-600">{{ t('profitOps.remarkLength') }}</p>
      <div class="flex flex-wrap gap-2">
        <button type="button" class="btn btn-primary" data-testid="save-url-remark" :disabled="busy || tooLong || conflicted" @click="save">{{ t('profitOps.remarkSave') }}</button>
        <button type="button" class="btn btn-secondary" data-testid="cancel-url-remark" :disabled="busy" @click="cancel">{{ t('profitOps.remarkCancel') }}</button>
        <button type="button" class="btn btn-secondary" data-testid="reload-url-remark" :disabled="busy" @click="reload">{{ t('profitOps.remarkReload') }}</button>
      </div>
    </template>
    <button v-else type="button" class="btn btn-secondary" data-testid="edit-url-remark" @click="edit">{{ t('profitOps.remarkEdit') }}</button>
    <p v-if="error" role="alert" data-testid="url-remark-error" class="text-sm text-red-600">{{ error }}</p>
  </div>
</template>
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { readUpstreamURLRemark, saveUpstreamURLRemark, type UpstreamURLRemark } from '@/api/admin/profitOperations'
const props = defineProps<{ value: UpstreamURLRemark }>()
const emit = defineEmits<{ saved: [value: UpstreamURLRemark] }>()
const { t } = useI18n()
const current = ref({ ...props.value }), draft = ref(''), editing = ref(false), busy = ref(false), error = ref(''), conflicted = ref(false)
const tooLong = computed(() => Array.from(draft.value.trim()).length > 500)
watch(() => props.value, value => { if (!busy.value) current.value = { ...value } })
function edit() { draft.value = current.value.remark; error.value = ''; conflicted.value = false; editing.value = true }
function cancel() { editing.value = false; error.value = ''; draft.value = '' }
async function reload() {
 busy.value = true; error.value = ''
 try { current.value = await readUpstreamURLRemark(props.value.upstream); conflicted.value = false; emit('saved', current.value) }
 catch { error.value = t('profitOps.remarkReloadFailed') }
 finally { busy.value = false }
}
async function save() {
 if (busy.value || tooLong.value || conflicted.value) return
 busy.value = true; error.value = ''
 let written: UpstreamURLRemark
 try {
  written = await saveUpstreamURLRemark({ upstream: props.value.upstream, remark: draft.value.trim(), revision: current.value.revision })
 } catch (err) {
  conflicted.value = (err as { response?: { status?: number } }).response?.status === 409
  error.value = t(conflicted.value ? 'profitOps.remarkConflict' : 'profitOps.remarkFailed')
  busy.value = false; return
 }
 try {
  const persisted = await readUpstreamURLRemark(props.value.upstream)
  if (persisted.revision !== written.revision || persisted.remark !== draft.value.trim()) {
   conflicted.value = true; error.value = t('profitOps.remarkConflict')
  } else {
   current.value = persisted; emit('saved', persisted); editing.value = false
  }
 } catch { error.value = t('profitOps.remarkReadbackFailed'); conflicted.value = true }
 finally { busy.value = false }
}
</script>
