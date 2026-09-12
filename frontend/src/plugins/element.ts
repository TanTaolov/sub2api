import ElementFloatingPanel from '@/components/element/ElementFloatingPanel.vue'
import type { App } from 'vue'
import { ElButton, ElOption, ElOptionGroup, ElForm, ElCard, ElTag, ElProgress, ElEmpty, ElSkeleton, ElSwitch, ElSlider, ElTable, ElTableColumn, ElDialog, ElCollapse, ElCollapseItem, ElImageViewer, ElPopover, ElRadio, ElRadioGroup, ElCheckbox, ElTabs, ElTabPane, ElDropdown, ElDropdownMenu, ElDropdownItem } from 'element-plus'
import ElementInput from '@/components/element/ElementInput.vue'
import ElementCheckbox from '@/components/element/ElementCheckbox.vue'
import ElementRadio from '@/components/element/ElementRadio.vue'
import ElementSelect from '@/components/element/ElementSelect.vue'

export const elementPlugin = {
  install(app: App) {
    for (const component of [ElButton, ElOption, ElOptionGroup, ElForm, ElCard, ElTag, ElProgress, ElEmpty, ElSkeleton, ElSwitch, ElSlider, ElTable, ElTableColumn, ElDialog, ElCollapse, ElCollapseItem, ElImageViewer, ElPopover, ElRadio, ElRadioGroup, ElCheckbox, ElTabs, ElTabPane, ElDropdown, ElDropdownMenu, ElDropdownItem]) {
      app.component(component.name!, component)
    }
    app.component('ElementFloatingPanel', ElementFloatingPanel)
    app.component('ElementInput', ElementInput)
    app.component('ElementCheckbox', ElementCheckbox)
    app.component('ElementRadio', ElementRadio)
    app.component('ElementSelect', ElementSelect)
  }
}
