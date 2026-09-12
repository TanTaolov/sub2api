<template>
  <el-dialog
    :model-value="show"
    :title="t('admin.users.editUser')"
    width="min(92vw, 36rem)"
    append-to-body
    destroy-on-close
    @update:model-value="handleDialogVisibilityChange"
  >
    <el-form
      v-if="user"
      ref="formRef"
      :model="form"
      :rules="formRules"
      label-position="top"
      @submit.prevent="handleUpdateUser"
    >
      <el-form-item :label="t('admin.users.email')" prop="email">
        <el-input v-model="form.email" type="email" autocomplete="email" />
      </el-form-item>

      <el-form-item :label="t('admin.users.password')">
        <div class="flex w-full gap-2">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            class="flex-1"
            autocomplete="new-password"
            :placeholder="t('admin.users.enterNewPassword')"
          >
            <template v-if="form.password" #suffix>
              <el-button
                text
                :type="passwordCopied ? 'success' : 'info'"
                :aria-label="t('common.copy')"
                @click="copyPassword"
              >
                <svg v-if="passwordCopied" class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                </svg>
                <svg v-else class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" />
                </svg>
              </el-button>
            </template>
          </el-input>
          <el-button :aria-label="t('common.refresh')" @click="generatePassword">
            <Icon name="refresh" size="sm" />
          </el-button>
        </div>
      </el-form-item>

      <el-form-item :label="t('admin.users.username')">
        <el-input v-model="form.username" />
      </el-form-item>

      <el-form-item :label="t('admin.users.form.roleLabel')" prop="role">
        <el-select v-model="form.role" class="w-full">
          <el-option
            v-for="option in roleOptions"
            :key="option.value"
            :value="option.value"
            :label="option.label"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('admin.users.notes')">
        <el-input v-model="form.notes" type="textarea" :rows="3" />
      </el-form-item>

      <div class="grid grid-cols-1 gap-x-4 sm:grid-cols-2">
        <el-form-item :label="t('admin.users.columns.concurrency')" prop="concurrency">
          <el-input-number v-model="form.concurrency" :min="0" :step="1" class="!w-full" />
          <div class="mt-1 text-xs text-gray-500 dark:text-dark-400">
            {{ t('admin.users.form.concurrencyHint') }}
          </div>
        </el-form-item>

        <el-form-item :label="t('admin.users.form.rpmLimit')" prop="rpm_limit">
          <el-input-number v-model="form.rpm_limit" :min="0" :step="1" class="!w-full" />
          <div class="mt-1 text-xs text-gray-500 dark:text-dark-400">
            {{ t('admin.users.form.rpmLimitHint') }}
          </div>
        </el-form-item>
      </div>

      <el-form-item :label="t('admin.users.attributes.title')">
        <UserAttributeForm v-model="form.customAttributes" :user-id="user.id" />
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <el-button @click="emit('close')">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="handleUpdateUser">
          {{ submitting ? t('admin.users.updating') : t('common.update') }}
        </el-button>
      </div>
    </template>
  </el-dialog>

  <TotpStepUpDialog :controller="stepUp" />
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { adminAPI } from '@/api/admin'
import type { AdminUser, UserAttributeValuesMap } from '@/types'
import UserAttributeForm from '@/components/user/UserAttributeForm.vue'
import Icon from '@/components/icons/Icon.vue'
import { useStepUp, isStepUpBlocked, isStepUpCancelled, stepUpBlockReason } from '@/composables/useStepUp'
import TotpStepUpDialog from '@/components/auth/TotpStepUpDialog.vue'

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
const emit = defineEmits<{
  close: []
  success: []
}>()
const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const formRef = ref<FormInstance>()
const submitting = ref(false)
const passwordCopied = ref(false)
const stepUp = useStepUp()

const roleOptions = computed(() => [
  { value: 'user', label: t('admin.users.roles.user') },
  { value: 'admin', label: t('admin.users.roles.admin') }
])

const form = reactive({
  email: '',
  password: '',
  username: '',
  notes: '',
  role: 'user' as AdminUser['role'],
  concurrency: 1,
  rpm_limit: 0,
  customAttributes: {} as UserAttributeValuesMap
})

const formRules = computed<FormRules>(() => ({
  email: [
    { required: true, message: t('admin.users.emailRequired'), trigger: 'blur' },
    { type: 'email', message: t('admin.users.enterEmail'), trigger: 'blur' }
  ],
  concurrency: [{ required: true, type: 'number', min: 0, message: t('admin.users.concurrencyNonNegative'), trigger: 'change' }],
  rpm_limit: [{ required: true, type: 'number', min: 0, message: t('admin.users.rpmLimitNonNegative'), trigger: 'change' }]
}))

function handleDialogVisibilityChange(visible: boolean) {
  if (!visible) emit('close')
}

function generatePassword() {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789!@#$%^&*'
  let password = ''
  for (let index = 0; index < 16; index += 1) {
    password += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  form.password = password
}

async function copyPassword() {
  if (form.password && await copyToClipboard(form.password, t('admin.users.passwordCopied'))) {
    passwordCopied.value = true
    window.setTimeout(() => {
      passwordCopied.value = false
    }, 2000)
  }
}

async function validateForm(): Promise<boolean> {
  try {
    return await formRef.value?.validate() ?? false
  } catch {
    return false
  }
}

function isNonNegativeInteger(value: number): boolean {
  return Number.isInteger(value) && value >= 0
}

async function handleUpdateUser() {
  if (!props.user || submitting.value) return
  if (!await validateForm()) return

  // 0 表示不限流，与网关和批量配额接口保持一致。
  if (!isNonNegativeInteger(form.concurrency)) {
    appStore.showError(t('admin.users.concurrencyNonNegative'))
    return
  }
  if (!isNonNegativeInteger(form.rpm_limit)) {
    appStore.showError(t('admin.users.rpmLimitNonNegative'))
    return
  }

  const userId = props.user.id
  submitting.value = true
  try {
    const data: any = {
      email: form.email,
      username: form.username,
      notes: form.notes,
      role: form.role,
      concurrency: form.concurrency,
      rpm_limit: form.rpm_limit
    }
    if (form.password.trim()) {
      data.password = form.password.trim()
    }

    await stepUp.run(() => adminAPI.users.update(userId, data))
    if (Object.keys(form.customAttributes).length > 0) {
      await adminAPI.userAttributes.updateUserAttributeValues(userId, form.customAttributes)
    }
    appStore.showSuccess(t('admin.users.userUpdated'))
    emit('success')
    emit('close')
  } catch (error: unknown) {
    if (isStepUpCancelled(error)) {
      return
    }
    if (isStepUpBlocked(error)) {
      appStore.showError(
        stepUpBlockReason(error) === 'STEP_UP_ADMIN_API_KEY_FORBIDDEN'
          ? t('stepUp.adminApiKeyForbidden')
          : t('stepUp.notEnabled')
      )
      return
    }

    const message = error instanceof Error ? error.message : t('admin.users.failedToUpdate')
    appStore.showError(message || t('admin.users.failedToUpdate'))
  } finally {
    submitting.value = false
  }
}

function resetForm(user: AdminUser) {
  Object.assign(form, {
    email: user.email,
    password: '',
    username: user.username || '',
    notes: user.notes || '',
    role: user.role || 'user',
    concurrency: user.concurrency,
    rpm_limit: user.rpm_limit ?? 0,
    customAttributes: {}
  })
  passwordCopied.value = false
  formRef.value?.clearValidate()
}

watch(
  [() => props.user, () => props.show],
  ([user, visible]) => {
    if (user && visible) {
      resetForm(user)
    }
  },
  { immediate: true }
)
</script>
