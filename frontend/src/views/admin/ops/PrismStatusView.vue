<template>
  <AppLayout>
    <div class="prism-status">
      <SmartOpsNav />
      <header class="page-heading">
        <div>
          <p class="eyebrow">{{ t('accountOps.smartTitle') }}</p>
          <h2 class="flex items-center gap-2"><Icon name="cpu" size="lg" class="shrink-0" aria-hidden="true" />{{ t('prismStatus.title') }}</h2>
          <p class="subtitle">{{ t('prismStatus.description') }}</p>
        </div>
        <button class="btn btn-secondary inline-flex items-center gap-2" :disabled="loading" @click="load">
          <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />{{ t('qualityOps.refresh') }}
        </button>
      </header>

      <p v-if="error" role="alert" class="error-banner">{{ error }}</p>

      <section v-if="!status" class="summary-grid" role="status" :aria-busy="loading" data-testid="prism-skeleton">
        <span v-for="n in 4" :key="n" class="summary-card skeleton" />
      </section>

      <template v-else>
        <section class="state-banner" :class="`state-${status.state}`" role="status" data-testid="prism-state">
          <Icon :name="stateIcon" size="md" class="mt-0.5 shrink-0" />
          <div>
            <strong>{{ t(`prismStatus.states.${status.state}`) }}</strong>
            <p>{{ t(`prismStatus.stateHints.${status.state}`) }}</p>
          </div>
        </section>

        <section class="summary-grid">
          <article class="summary-card" data-testid="prism-enabled">
            <span>{{ t('prismStatus.enabled') }}</span>
            <strong>{{ t(status.enabled ? 'prismStatus.on' : 'prismStatus.off') }}</strong>
            <small>{{ t('prismStatus.enabledHint') }}</small>
          </article>
          <article class="summary-card" data-testid="prism-base-url">
            <span>{{ t('prismStatus.baseUrl') }}</span>
            <strong class="value">{{ status.base_url || t('prismStatus.notSet') }}</strong>
            <small>{{ t('prismStatus.baseUrlHint') }}</small>
          </article>
          <article class="summary-card" data-testid="prism-endpoint">
            <span>{{ t('prismStatus.endpoint') }}</span>
            <strong class="value">{{ status.endpoint || t('prismStatus.endpointUnavailable') }}</strong>
            <small>{{ t('prismStatus.endpointHint') }}</small>
          </article>
          <article class="summary-card" data-testid="prism-api-key">
            <span>{{ t('prismStatus.apiKey') }}</span>
            <strong>{{ t(status.api_key_configured ? 'prismStatus.apiKeyConfigured' : 'prismStatus.apiKeyMissing') }}</strong>
            <small>{{ t('prismStatus.apiKeyHint') }}</small>
          </article>
        </section>

        <section class="models-card">
          <header>
            <div>
              <h3>{{ t('prismStatus.models') }}</h3>
              <p>{{ t('prismStatus.modelsHint') }}</p>
            </div>
            <span class="count">{{ status.models.length }}</span>
          </header>
          <ul class="model-chips" data-testid="prism-models">
            <li v-for="model in status.models" :key="model">{{ model }}</li>
          </ul>
        </section>
      </template>

      <aside class="scope-note"><Icon name="infoCircle" size="sm" /><p>{{ t('prismStatus.scopeNote') }}</p></aside>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import SmartOpsNav from '@/components/admin/operations/SmartOpsNav.vue'
import Icon from '@/components/icons/Icon.vue'
import { getPrismStatus, type PrismStatus } from '@/api/admin/prismStatus'
const { t } = useI18n()
const status = ref<PrismStatus | null>(null)
const loading = ref(false), error = ref('')
type PrismStateIcon = 'checkCircle' | 'infoCircle' | 'exclamationCircle'
const stateIcon = computed<PrismStateIcon>(() => (status.value?.state === 'ready' ? 'checkCircle' : status.value?.state === 'disabled' ? 'infoCircle' : 'exclamationCircle'))
async function load() {
  if (loading.value) return
  loading.value = true; error.value = ''
  try { status.value = await getPrismStatus() }
  catch (e) { error.value = (e as { message?: string })?.message || t('qualityOps.error') }
  finally { loading.value = false }
}
onMounted(() => { void load() })
</script>

<style scoped>
.prism-status { @apply w-full min-w-0 text-gray-900 dark:text-gray-100; }
.page-heading { @apply mb-6 flex flex-wrap items-center justify-between gap-4; }
.eyebrow { @apply mb-1 text-[11px] font-semibold tracking-widest text-primary-600; }
.page-heading h2 { @apply text-2xl font-semibold tracking-tight; }
.subtitle { @apply mt-2 max-w-3xl text-sm leading-6 text-gray-500 dark:text-gray-400; }
.summary-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(220px,1fr)); gap:20px; @apply mb-5; }
.summary-card,.models-card { @apply min-w-0 overflow-hidden rounded-2xl border border-gray-200/80 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-900; }
.summary-card span { @apply block text-xs font-medium text-gray-400; }
.summary-card strong { @apply mt-2 block text-2xl font-semibold tracking-tight; }
.summary-card strong.value { @apply text-sm font-medium break-all; }
.summary-card small { @apply mt-2 block text-xs leading-relaxed text-gray-400; }
.skeleton { @apply h-28 animate-pulse bg-gray-100 dark:bg-dark-800; }
.state-banner { @apply mb-5 flex items-start gap-3 rounded-2xl border p-4 text-sm; }
.state-banner strong { @apply block text-sm font-medium; }
.state-banner p { @apply mt-1 text-xs leading-relaxed; }
.state-ready { @apply border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900 dark:bg-emerald-950/30 dark:text-emerald-300; }
.state-disabled { @apply border-gray-200 bg-gray-50 text-gray-600 dark:border-dark-700 dark:bg-dark-800/60 dark:text-gray-300; }
.state-endpoint_invalid,.state-key_missing { @apply border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-300; }
.models-card header { @apply flex flex-wrap items-start justify-between gap-3; }
.models-card h3 { @apply text-base font-semibold; }
.models-card p { @apply mt-1 text-xs leading-relaxed text-gray-400; }
.models-card .count { @apply rounded-md bg-gray-100 px-2 py-0.5 text-xs font-normal tabular-nums text-gray-500 dark:bg-dark-800; }
.model-chips { @apply mt-4 flex flex-wrap gap-2; }
.model-chips li { @apply rounded-lg border border-gray-200 px-3 py-1.5 font-mono text-xs text-gray-700 dark:border-dark-600 dark:text-gray-200; }
.scope-note { @apply mt-5 flex items-start gap-2 text-xs leading-6 text-gray-400; }
.scope-note svg { @apply mt-1 shrink-0; }
.error-banner { @apply mb-4 rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-900 dark:bg-red-950/30 dark:text-red-300; }
button:disabled { @apply cursor-not-allowed opacity-40; }
</style>
