<template>
  <section class="space-y-3 rounded-lg border p-3 dark:border-dark-600" data-testid="account-channel-binding">
    <button type="button" class="btn btn-secondary" :disabled="!valid || busy" data-testid="open-channel-bindings" @click="load(false)">{{ t('upstreamEntities.accountChannel') }} · {{ accountIds.length }} · {{ t('upstreamEntities.load') }}</button>
    <p class="text-xs text-gray-500">{{ t('upstreamEntities.history') }} {{ t('upstreamEntities.bounded') }}</p>
    <template v-if="opened">
      <UpstreamChannelSelector v-model="target" allow-keep :selected="bindings.find(b => b.channel_id === target)?.channel" @update:model-value="confirmed = false" />
      <ul class="max-h-48 overflow-auto text-sm"><li v-for="b in bindings" :key="b.account_id">#{{ b.account_id }} {{ b.account_name }} → {{ b.channel ? `${b.channel.name} · ${b.channel.code}` : t('upstreamEntities.unbound') }}</li></ul>
      <label class="flex items-start gap-2 text-sm"><input v-model="confirmed" type="checkbox" data-testid="confirm-channel-binding" :disabled="busy" />{{ t('upstreamEntities.confirmBinding') }} {{ bindings.length }} {{ t('upstreamEntities.accountsUnit') }} #{{ target }}</label>
      <div class="flex flex-wrap gap-2"><button type="button" class="btn btn-primary" data-testid="save-channel-binding" :disabled="busy || !confirmed || target < 0 || needsRead || bindings.length !== accountIds.length" @click="save">{{ t('profitOps.remarkSave') }}</button><button type="button" class="btn btn-secondary" :disabled="busy" @click="cancel">{{ t('profitOps.remarkCancel') }}</button><button type="button" class="btn btn-secondary" :disabled="busy" data-testid="reload-channel-bindings" @click="load(true)">{{ t('upstreamEntities.readLatest') }}</button></div>
    </template>
    <p v-if="error" role="alert" class="text-red-600" data-testid="binding-error">{{ error }}</p><p v-if="saved" role="status">{{ t('upstreamEntities.saved') }}</p>
  </section>
</template>
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { bindChannelAccounts, readChannelBindings, type ChannelBinding } from '@/api/admin/upstreamChannels'
import UpstreamChannelSelector from './UpstreamChannelSelector.vue'
const props = defineProps<{ accountIds: number[] }>()
const emit = defineEmits<{ saved: [] }>()
const { t } = useI18n()
const bindings = ref<ChannelBinding[]>([]), target = ref(-1), confirmed = ref(false), opened = ref(false), busy = ref(false), error = ref(''), saved = ref(false), needsRead = ref(false)
const valid = computed(() => props.accountIds.length > 0 && props.accountIds.length <= 100 && new Set(props.accountIds).size === props.accountIds.length)
watch(() => props.accountIds.join(','), () => { opened.value = false; bindings.value = []; target.value = -1; confirmed.value = false; saved.value = false; needsRead.value = false })
async function load(preserve: boolean) {
 if (!valid.value || busy.value) return
 busy.value = true; error.value = ''; saved.value = false
 try { const next = await readChannelBindings([...props.accountIds]); if (next.length !== props.accountIds.length) throw new Error('incomplete'); bindings.value = next; opened.value = true; confirmed.value = false; needsRead.value = false; if (!preserve) target.value = -1 }
 catch { error.value = t('upstreamEntities.failed') } finally { busy.value = false }
}
function cancel() { opened.value = false; target.value = -1; confirmed.value = false; error.value = ''; saved.value = false }
async function save() {
 if (!confirmed.value || target.value < 0 || busy.value || needsRead.value) return
 busy.value = true; error.value = ''; saved.value = false
 try {
  const written = await bindChannelAccounts(target.value, bindings.value)
  needsRead.value = true
  const readback = await readChannelBindings([...props.accountIds])
  if (readback.length !== bindings.value.length || readback.some(b => b.channel_id !== target.value || b.revision !== written.find(w => w.account_id === b.account_id)?.revision)) throw new Error('readback')
  bindings.value = readback; saved.value = true; needsRead.value = false; confirmed.value = false; emit('saved')
 } catch (cause) { const e = cause as { status?: number; response?: { status?: number } }; needsRead.value = true; error.value = t((e.status ?? e.response?.status) === 409 ? 'upstreamEntities.conflict' : 'upstreamEntities.readbackFailed') }
 finally { busy.value = false }
}
</script>
