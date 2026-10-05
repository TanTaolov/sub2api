import { nextTick, ref, watch, type Ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { settingsLocation, type SettingsTab } from '@/utils/settingsSearch'

export function focusSettingsLocation(tab: unknown, hash: string) {
  const location = settingsLocation(tab, hash)
  if (!location.linked) return
  if (location.anchor) {
    const target = document.getElementById(location.anchor)
    target?.focus({ preventScroll: true })
    target?.scrollIntoView({ block: 'start' })
    return
  }
  // 纯 Tab 切换（无锚点）时不能对吸顶的 Tab 条调用 scrollIntoView：
  // sticky 元素已固定时浏览器会把页面滚到该元素的文档位置，
  // 导致吸顶导航脱离视口顶部。直接滚回页面顶部即可让导航回位。
  document.getElementById(`settings-tab-${location.tab}`)?.focus({ preventScroll: true })
  window.scrollTo({ top: 0 })
}

export function useSettingsNavigation(loading: Ref<boolean>, loadFailed: Ref<boolean>) {
  const route = useRoute()
  const router = useRouter()
  const activeTab = ref<SettingsTab>('general')

  watch(() => [route.query.tab, route.hash, loading.value] as const, async (_, __, onCleanup) => {
    const location = settingsLocation(route.query.tab, route.hash)
    activeTab.value = location.tab
    if (loading.value || loadFailed.value || !location.linked) return
    let cancelled = false
    onCleanup(() => { cancelled = true })
    await nextTick()
    if (cancelled) return
    focusSettingsLocation(route.query.tab, route.hash)
  }, { immediate: true, flush: 'post' })

  function selectSettingsTab(tab: SettingsTab) {
    activeTab.value = tab
    // Clear the old anchor so selecting the same search result works again.
    void router.replace({ query: { ...route.query, tab }, hash: '' })
  }

  return { activeTab, selectSettingsTab }
}
