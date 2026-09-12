import type ElementFloatingPanel from '@/components/element/ElementFloatingPanel.vue'
import type ElementInput from '@/components/element/ElementInput.vue'
import type ElementCheckbox from '@/components/element/ElementCheckbox.vue'
import type ElementRadio from '@/components/element/ElementRadio.vue'
import type ElementSelect from '@/components/element/ElementSelect.vue'

declare module 'vue' {
  export interface GlobalComponents {
    ElementFloatingPanel: typeof ElementFloatingPanel
    ElementInput: typeof ElementInput
    ElementCheckbox: typeof ElementCheckbox
    ElementRadio: typeof ElementRadio
    ElementSelect: typeof ElementSelect
    ElButton: typeof import('element-plus')['ElButton']
    ElOption: typeof import('element-plus')['ElOption']
    ElOptionGroup: typeof import('element-plus')['ElOptionGroup']
    ElForm: typeof import('element-plus')['ElForm']
    ElCard: typeof import('element-plus')['ElCard']
    ElTag: typeof import('element-plus')['ElTag']
    ElProgress: typeof import('element-plus')['ElProgress']
    ElEmpty: typeof import('element-plus')['ElEmpty']
    ElSkeleton: typeof import('element-plus')['ElSkeleton']
    ElSwitch: typeof import('element-plus')['ElSwitch']
    ElSlider: typeof import('element-plus')['ElSlider']
    ElTable: typeof import('element-plus')['ElTable']
    ElDialog: typeof import('element-plus')['ElDialog']
    ElCollapse: typeof import('element-plus')['ElCollapse']
    ElCollapseItem: typeof import('element-plus')['ElCollapseItem']
    ElImageViewer: typeof import('element-plus')['ElImageViewer']
    ElPopover: typeof import('element-plus')['ElPopover']
    ElRadio: typeof import('element-plus')['ElRadio']
    ElRadioGroup: typeof import('element-plus')['ElRadioGroup']
    ElCheckbox: typeof import('element-plus')['ElCheckbox']
    ElTabs: typeof import('element-plus')['ElTabs']
    ElTabPane: typeof import('element-plus')['ElTabPane']
    ElDropdown: typeof import('element-plus')['ElDropdown']
    ElDropdownMenu: typeof import('element-plus')['ElDropdownMenu']
    ElDropdownItem: typeof import('element-plus')['ElDropdownItem']
    ElTableColumn: typeof import('element-plus')['ElTableColumn']
  }
}
export {}
