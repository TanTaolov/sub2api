<template>
  <el-dialog
    :model-value="show"
    :title="t('admin.users.createUser')"
    width="min(92vw, 34rem)"
    append-to-body
    destroy-on-close
    @update:model-value="handleDialogVisibilityChange"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="formRules"
      label-position="top"
      @submit.prevent="submit"
    >
      <el-form-item :label="t('admin.users.email')" prop="email">
        <el-input
          v-model="form.email"
          type="email"
          autocomplete="email"
          :placeholder="t('admin.users.enterEmail')"
        />
      </el-form-item>

      <el-form-item :label="t('admin.users.password')" prop="password">
        <div class="flex w-full gap-2">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            class="flex-1"
            autocomplete="new-password"
            :placeholder="t('admin.users.enterPassword')"
          />
          <el-button :aria-label="t('common.refresh')" @click="generateRandomPassword">
            <Icon name="refresh" size="sm" />
          </el-button>
        </div>
      </el-form-item>

      <el-form-item :label="t('admin.users.username')">
        <el-input v-model="form.username" :placeholder="t('admin.users.enterUsername')" />
      </el-form-item>

      <el-form-item :label="t('admin.users.form.roleLabel')" prop="role">
        <el-select v-model="form.role" class="w-full">
          <el-option value="user" :label="t('admin.users.roles.user')" />
          <el-option value="admin" :label="t('admin.users.roles.admin')" />
        </el-select>
      </el-form-item>

      <div class="grid grid-cols-1 gap-x-4 sm:grid-cols-2">
        <el-form-item :label="t('admin.users.columns.balance')">
          <el-input-number v-model="form.balance" :precision="2" :step="1" class="!w-full" />
        </el-form-item>
        <el-form-item :label="t('admin.users.columns.concurrency')" prop="concurrency">
          <el-input-number v-model="form.concurrency" :min="0" :step="1" class="!w-full" />
        </el-form-item>
      </div>

      <el-form-item :label="t('admin.users.form.rpmLimit')" prop="rpm_limit">
        <el-input-number v-model="form.rpm_limit" :min="0" :step="1" class="!w-full" />
        <div class="mt-1 text-xs text-gray-500 dark:text-dark-400">
          {{ t('admin.users.form.rpmLimitHint') }}
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <el-button @click="emit('close')">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="loading" @click="submit">
          {{ loading ? t('admin.users.creating') : t('common.create') }}
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
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import Icon from '@/components/icons/Icon.vue'
import { useStepUp, isStepUpBlocked, isStepUpCancelled, stepUpBlockReason } from '@/composables/useStepUp'
import TotpStepUpDialog from '@/components/auth/TotpStepUpDialog.vue'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{
  close: []
  success: []
}>()
const { t } = useI18n()
const appStore = useAppStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const stepUp = useStepUp()

const form = reactive({
  email: '',
  password: '',
  username: '',
  notes: '',
  role: 'user' as 'user' | 'admin',
  balance: undefined as number | undefined,
  concurrency: 1,
  rpm_limit: 0
})

const formRules = computed<FormRules>(() => ({
  email: [
    { required: true, message: t('admin.users.emailRequired'), trigger: 'blur' },
    { type: 'email', message: t('admin.users.enterEmail'), trigger: 'blur' }
  ],
  password: [{ required: true, message: t('admin.users.enterPassword'), trigger: 'blur' }],
  concurrency: [{ required: true, type: 'number', min: 0, message: t('admin.users.concurrencyNonNegative'), trigger: 'change' }],
  rpm_limit: [{ required: true, type: 'number', min: 0, message: t('admin.users.rpmLimitNonNegative'), trigger: 'change' }]
}))

function handleDialogVisibilityChange(visible: boolean) {
  if (!visible) emit('close')
}

async function submit() {
  if (loading.value) return

  let valid = false
  try {
    valid = await formRef.value?.validate() ?? false
  } catch {
    valid = false
  }
  if (!valid) return

  if (!isNonNegativeInteger(form.concurrency)) {
    appStore.showError(t('admin.users.concurrencyNonNegative'))
    return
  }
  if (!isNonNegativeInteger(form.rpm_limit)) {
    appStore.showError(t('admin.users.rpmLimitNonNegative'))
    return
  }
  if (form.balance !== undefined && !Number.isFinite(form.balance)) {
    appStore.showError(t('admin.users.amountRequired'))
    return
  }

  loading.value = true
  try {
    const { balance, ...rest } = { ...form }
    const payload: typeof rest & { balance?: number } = { ...rest }
    if (balance !== undefined) {
      payload.balance = balance
    }

    await stepUp.run(() => adminAPI.users.create(payload))
    appStore.showSuccess(t('admin.users.userCreated'))
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

    const message = error instanceof Error ? error.message : t('admin.users.failedToCreate')
    appStore.showError(message || t('admin.users.failedToCreate'))
  } finally {
    loading.value = false
  }
}

function isNonNegativeInteger(value: number | undefined): value is number {
  return Number.isInteger(value) && value >= 0
}

function generateRandomPassword() {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789!@#$%^&*'
  let password = ''
  for (let index = 0; index < 16; index += 1) {
    password += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  form.password = password
  void formRef.value?.validateField('password')
}

watch(
  () => props.show,
  (visible) => {
    if (!visible) return
    Object.assign(form, {
      email: '',
      password: '',
      username: '',
      notes: '',
      role: 'user',
      balance: undefined,
      concurrency: 1,
      rpm_limit: 0
    })
    formRef.value?.clearValidate()
  }
)
</script>
