<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRoute, useRouter } from 'vue-router'
import { Person24Regular, LockClosed24Regular, ArrowRight24Regular } from '@vicons/fluent'
import { systemService } from '../services/system'
import { t } from '../i18n'
import type { PasswordCredentialStatus } from '../types/credentials'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const form = ref({ username: '', password: '' })
const loading = ref(false)
const passwordChangeOpen = ref(false)
const changingPassword = ref(false)
const passwordChangeError = ref('')
const newPasswordInput = ref<{ focus: () => void } | null>(null)
const passwordChangeForm = ref({ old_password: '', new_password: '', confirm_password: '' })

async function handleLogin() {
  const { ElMessage } = await import('element-plus')
  if (!form.value.username || !form.value.password) {
    ElMessage.warning(t('login.needCredentials'))
    return
  }

  loading.value = true
  const result = await auth.login(form.value.username, form.value.password)
  loading.value = false
  if (!result.ok) {
    ElMessage.error(t('login.failed'))
    return
  }

  if (await startPasswordRemediation(result.credential)) {
    return
  }
  await completeLogin(t('login.welcome'))
}

async function startPasswordRemediation(status: PasswordCredentialStatus): Promise<boolean> {
  if (!status.change_required) return false
  const { ElMessageBox } = await import('element-plus')
  if (status.management === 'environment') {
    const variable = status.environment_variable || 'PROXY_WEB_PASSWORD'
    await ElMessageBox.alert(
      t('login.weakPasswordEnv', { variable }),
      t('login.weakPasswordTitle'),
      { confirmButtonText: t('login.weakPasswordAck'), showClose: false, closeOnClickModal: false, closeOnPressEscape: false, type: 'warning' }
    )
    return false
  }
  passwordChangeForm.value = {
    old_password: form.value.password,
    new_password: '',
    confirm_password: ''
  }
  form.value.password = ''
  passwordChangeError.value = ''
  passwordChangeOpen.value = true
  return true
}

async function submitPasswordChange() {
  passwordChangeError.value = validatePasswordChange()
  if (passwordChangeError.value) return

  changingPassword.value = true
  const result = await systemService.changePassword(passwordChangeForm.value)
  changingPassword.value = false
  if (!result.ok) {
    passwordChangeError.value = result.error.message || '密码更新失败'
    return
  }
  auth.applyToken(result.data.token)
  await completeLogin(t('login.passwordUpdated'))
}

function validatePasswordChange(): string {
  if (!passwordChangeForm.value.old_password || !passwordChangeForm.value.new_password) {
    return '请填写当前密码和新密码'
  }
  if (passwordChangeForm.value.new_password !== passwordChangeForm.value.confirm_password) {
    return '两次输入的新密码不一致'
  }
  return ''
}

async function completeLogin(message: string) {
  const { ElMessage } = await import('element-plus')
  passwordChangeOpen.value = false
  passwordChangeForm.value = { old_password: '', new_password: '', confirm_password: '' }
  passwordChangeError.value = ''
  form.value.password = ''
  ElMessage.success(message)
  await redirectAfterLogin()
}

async function redirectAfterLogin() {
  const queryRedirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
  let redirect = queryRedirect ? decodeURIComponent(queryRedirect) : ''
  if (!redirect) {
    try {
      redirect = sessionStorage.getItem('post_login_redirect') || ''
    } catch {
      // Storage can be unavailable in hardened browser contexts.
    }
  }
  if (redirect) {
    try {
      sessionStorage.removeItem('post_login_redirect')
    } catch {
      // Storage can be unavailable in hardened browser contexts.
    }
    await router.push(redirect)
    return
  }
  await router.push('/')
}
</script>

<template>
  <div class="relative w-full h-full flex items-center justify-center overflow-hidden">
    <div class="absolute -top-32 -left-32 w-[520px] h-[520px] rounded-full bg-indigo-500/15 dark:bg-indigo-500/20 blur-[120px] animate-pulse-slow" />
    <div class="absolute -bottom-32 -right-32 w-[520px] h-[520px] rounded-full bg-purple-500/15 dark:bg-purple-500/20 blur-[120px] animate-pulse-slow" style="animation-delay: 2s" />

    <div class="relative w-full max-w-md p-1">
      <div class="relative bg-white/70 dark:bg-[#141418]/70 backdrop-blur-xl border border-gray-100 dark:border-white/10 rounded-2xl p-8 shadow-2xl overflow-hidden group">
        <div class="absolute inset-0 bg-gradient-to-br from-indigo-500/8 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none" />

        <div class="text-center mb-10 relative z-10">
          <div class="w-20 h-20 bg-gradient-to-tr from-indigo-500 to-purple-600 rounded-2xl mx-auto flex items-center justify-center text-white text-2xl font-bold shadow-lg shadow-indigo-500/20 mb-6 transform group-hover:scale-105 transition-transform duration-300">
            VH+
          </div>
          <h2 class="text-3xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-gray-900 to-gray-600 dark:from-white dark:to-gray-400">
            VoHive Plus
          </h2>
          <p class="text-gray-500 dark:text-gray-400 text-sm mt-3 tracking-wide">4G 模组管理后台</p>
        </div>

        <form @submit.prevent="handleLogin" class="space-y-6 relative z-10">
          <div class="space-y-2">
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-gray-400 dark:text-gray-500">
                <Person24Regular class="w-5 h-5" />
              </div>
              <input
                v-model="form.username"
                class="w-full bg-white/70 dark:bg-black/20 border border-gray-200 dark:border-white/10 rounded-lg py-3 pl-10 pr-4 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/25 focus:border-indigo-500/40 transition-all font-mono text-sm"
                placeholder="用户名"
                type="text" autocomplete="username"
              />
            </div>
          </div>

          <div class="space-y-2">
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-gray-400 dark:text-gray-500">
                <LockClosed24Regular class="w-5 h-5" />
              </div>
              <input
                v-model="form.password"
                class="w-full bg-white/70 dark:bg-black/20 border border-gray-200 dark:border-white/10 rounded-lg py-3 pl-10 pr-4 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/25 focus:border-indigo-500/40 transition-all font-mono text-sm"
                placeholder="密码"
                type="password" autocomplete="current-password"
              />
            </div>
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="w-full bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white font-bold py-3 px-4 rounded-lg shadow-lg shadow-indigo-600/30 flex items-center justify-center gap-2 transform active:scale-95 transition-all duration-200 disabled:opacity-70 disabled:cursor-not-allowed"
          >
            <span v-if="loading" class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
            <span v-else>登录</span>
            <ArrowRight24Regular v-if="!loading" class="w-5 h-5" />
          </button>
        </form>
      </div>

      <div class="text-center mt-6">
        <p class="text-gray-500 text-xs">VoHive Plus &copy; 2026</p>
      </div>
    </div>
    <el-dialog
      v-model="passwordChangeOpen"
      :title="t('login.changePassword')"
      width="min(440px, calc(100vw - 32px))"
      :show-close="false"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      @opened="newPasswordInput?.focus()"
    >
      <form class="password-change-form" @submit.prevent="submitPasswordChange">
        <p class="password-change-notice">当前密码仍是初始明文凭证或强度不足。建议立即修改。</p>

        <div class="space-y-1">
          <label for="login-current-password">当前密码</label>
          <el-input
            id="login-current-password"
            v-model="passwordChangeForm.old_password"
            type="password"
            show-password
            autocomplete="current-password"
          />
        </div>
        <div class="space-y-1">
          <label for="login-new-password">新密码</label>
          <el-input
            id="login-new-password"
            ref="newPasswordInput"
            v-model="passwordChangeForm.new_password"
            type="password"
            show-password
            autocomplete="new-password"
            placeholder="至少 8 位，建议 12 位以上"
          />
        </div>
        <div class="space-y-1">
          <label for="login-confirm-password">确认新密码</label>
          <el-input
            id="login-confirm-password"
            v-model="passwordChangeForm.confirm_password"
            type="password"
            show-password
            autocomplete="new-password"
          />
        </div>

        <p v-if="passwordChangeError" class="password-change-error" role="alert">
          {{ passwordChangeError }}
        </p>
        <div class="password-change-actions">
          <el-button native-type="button" :disabled="changingPassword" @click="completeLogin('欢迎回来')">
            稍后处理
          </el-button>
          <el-button type="primary" native-type="submit" :loading="changingPassword">
            更新密码并进入
          </el-button>
        </div>
      </form>
    </el-dialog>
  </div>
</template>
<style scoped>
.password-change-form { display: grid; gap: 16px; }
.password-change-actions { display: flex; gap: 8px; justify-content: flex-end; }
.password-change-error { color: var(--el-color-danger); }
</style>
