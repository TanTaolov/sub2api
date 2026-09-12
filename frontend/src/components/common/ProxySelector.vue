<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElSelect, ElOption, ElButton, ElTag } from 'element-plus'
import { adminAPI } from '@/api/admin'
import type { Proxy } from '@/types'
const props = withDefaults(defineProps<{ modelValue: number | null; proxies: Proxy[]; disabled?: boolean }>(), { disabled: false })
const emit = defineEmits<{ 'update:modelValue': [value: number | null] }>()
const { t } = useI18n()
interface ProxyTestResult { success: boolean; message: string; latency_ms?: number; ip_address?: string; city?: string; region?: string; country?: string }
const testResults = reactive<Record<number, ProxyTestResult>>({})
const testingProxyIds = reactive(new Set<number>())
const batchTesting = ref(false)
const searchQuery = ref('')
const handleTestProxy = async (proxy: Proxy) => {
  if (testingProxyIds.has(proxy.id)) return

  testingProxyIds.add(proxy.id)
  try {
    const result = await adminAPI.proxies.testProxy(proxy.id)
    testResults[proxy.id] = result
  } catch (error: any) {
    testResults[proxy.id] = {
      success: false,
      message: error.response?.data?.detail || 'Test failed'
    }
  } finally {
    testingProxyIds.delete(proxy.id)
  }
}

const handleBatchTest = async () => {
  if (batchTesting.value || props.proxies.length === 0) return

  batchTesting.value = true

  // Test all proxies in parallel
  const testPromises = props.proxies.map(async (proxy) => {
    testingProxyIds.add(proxy.id)
    try {
      const result = await adminAPI.proxies.testProxy(proxy.id)
      testResults[proxy.id] = result
    } catch (error: any) {
      testResults[proxy.id] = {
        success: false,
        message: error.response?.data?.detail || 'Test failed'
      }
    } finally {
      testingProxyIds.delete(proxy.id)
    }
  })

  await Promise.all(testPromises)
  batchTesting.value = false
}


</script>
<template>
  <ElSelect :model-value="modelValue ?? 'none'" :disabled="disabled" class="w-full" filterable
    :filter-method="(query: string) => searchQuery = query" :aria-label="t('admin.proxies.searchProxies')"
    @update:model-value="emit('update:modelValue', $event === 'none' ? null : Number($event))"
    @visible-change="(visible: boolean) => { if (!visible) searchQuery = '' }">
    <template #header><ElButton size="small" :loading="batchTesting" :disabled="proxies.length === 0" @click.stop="handleBatchTest">{{ t('admin.proxies.batchTest') }}</ElButton></template>
    <ElOption value="none" :label="t('admin.accounts.noProxy')" />
    <ElOption v-for="proxy in proxies.filter(proxy => (proxy.name + ' ' + proxy.host).toLocaleLowerCase().includes(searchQuery.toLocaleLowerCase()))"
      :key="proxy.id" :value="proxy.id" :label="proxy.name">
      <div class="flex items-center justify-between gap-3">
        <div class="min-w-0">
          <div class="flex items-center gap-2"><span class="truncate font-medium">{{ proxy.name }}</span><ElTag v-if="proxy.account_count !== undefined" size="small" type="info">{{ proxy.account_count }}</ElTag></div>
          <div class="truncate text-xs text-gray-500">{{ proxy.protocol }}://{{ proxy.host }}:{{ proxy.port }}</div>
          <ElTag v-if="testResults[proxy.id]" :type="testResults[proxy.id].success ? 'success' : 'danger'" size="small">
            {{ testResults[proxy.id].success ? [testResults[proxy.id].country, testResults[proxy.id].latency_ms != null ? testResults[proxy.id].latency_ms + 'ms' : ''].filter(Boolean).join(' ') : t('admin.proxies.testFailed') }}
          </ElTag>
        </div>
        <ElButton size="small" :loading="testingProxyIds.has(proxy.id)" @click.stop="handleTestProxy(proxy)">{{ t('admin.proxies.testConnection') }}</ElButton>
      </div>
    </ElOption>
  </ElSelect>
</template>
