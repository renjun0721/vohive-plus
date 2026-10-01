<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useSettingsStore } from '../stores/settings'
import { useAuthStore } from '../stores/auth'
import type { PasswordCredentialStatus } from '../types/credentials'
import PageHeader from '../components/PageHeader.vue'
import FieldRow from '../components/FieldRow.vue'
import PersonalHealth from '../components/settings/PersonalHealth.vue'
import QQNotificationTab from '../components/settings/QQNotificationTab.vue'
import TelegramNotificationTab from '../components/settings/TelegramNotificationTab.vue'
import WeComBotNotificationTab from '../components/settings/WeComBotNotificationTab.vue'
import WeixinNotificationTab from '../components/settings/WeixinNotificationTab.vue'
import FeishuNotificationTab from '../components/settings/FeishuNotificationTab.vue'
import {
  Key24Regular,
  Save24Regular,
  Server24Regular,
  Alert24Regular,
  Add20Regular,
  Delete20Regular,
  DocumentText24Regular
} from '@vicons/fluent'
import { formatDeviceDateTime } from '../utils/deviceTime'
import { useTheme } from '../composables/useTheme'
import { t, useLocale } from '../i18n'
import type { AppLocale } from '../utils/locale'

const settingsStore = useSettingsStore()
const authStore = useAuthStore()
const { isClassic, applyClassic, restoreNavyTheme } = useTheme()
const { locale, setLocale } = useLocale()
const { systemInfo, loadingNotifications, savingNotifications, testingWebhook, testingBark, testingEmail, testingWeCom, changingPassword, passwordForm, telegramForm, feishuForm, qqForm, weixinForm, weComBotForm, webhookSettings, barkSettings, emailForm, pushplusForm, weComSettings } = storeToRefs(settingsStore)
const activeNotifyTab = ref('telegram')
const openWRTDynamicInterfaces = ref(false)
const loadingSystemSettings = ref(false)
const savingSystemSettings = ref(false)
const passwordStatus = ref<PasswordCredentialStatus | null>(null)
const loadingPasswordStatus = ref(false)
const passwordManagedByEnvironment = computed(() => passwordStatus.value?.management === 'environment')

const enabledNotificationCount = computed(() => [
  telegramForm.value.enabled,
  qqForm.value.enabled
].filter(Boolean).length)



const hasValidWebhookURLs = computed(() => {
  if (!Array.isArray(webhookSettings.value.urls)) {
    return false
  }
  return webhookSettings.value.urls.some((u) => String(u || '').trim().length > 0)
})

const hasValidBarkURLs = computed(() => {
  if (!Array.isArray(barkSettings.value.urls)) {
    return false
  }
  return barkSettings.value.urls.some((u) => String(u || '').trim().length > 0)
})

const hasValidEmailConfig = computed(() => {
  return !!(
    emailForm.value.smtp_host &&
    emailForm.value.smtp_port &&
    emailForm.value.username &&
    emailForm.value.password &&
    emailForm.value.from_address &&
    emailForm.value.to_addresses
  )
})

const MAX_WECOM_URLS = 8
const hasValidWeComConfig = computed(() => {
  return Array.isArray(weComSettings.value.urls) &&
    weComSettings.value.urls.some(url => String(url || '').trim()) &&
    !!String(weComSettings.value.payload_template || '').trim()
})


async function changePassword() {
  if (passwordManagedByEnvironment.value) {
    ElMessage.warning(`当前密码由 ${passwordStatus.value?.environment_variable || 'PROXY_WEB_PASSWORD'} 管理，请修改部署环境并重启`)
    return
  }
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    ElMessage.error('两次输入的新密码不一致')
    return
  }

  const result = await settingsStore.changePasswordFromForm()
  if (!result.ok) {
    ElMessage.error(result.error.message || '密码更新失败')
    return
  }
  authStore.applyToken(result.data.token)
  passwordStatus.value = result.data.credential
  ElMessage.success('密码已更新')
  settingsStore.resetPasswordForm()
}

async function loadPasswordStatus() {
  loadingPasswordStatus.value = true
  const result = await systemService.getPasswordStatus()
  loadingPasswordStatus.value = false
  if (!result.ok) {
    passwordStatus.value = null
    ElMessage.error(result.error.message || '密码管理状态加载失败')
    return
  }
  passwordStatus.value = result.data
}

async function loadSystemInfo() {
  const result = await settingsStore.fetchSystemInfo()
  if (!result.ok) {
    console.error('系统信息读取失败', result.error)
  }
}

async function loadSystemSettings() {
  loadingSystemSettings.value = true
  const result = await systemService.getSystemSettings()
  loadingSystemSettings.value = false
  if (!result.ok) {
    ElMessage.error(result.error.message || '系统设置加载失败')
    return
  }
  openWRTDynamicInterfaces.value = !!result.data.openwrt_dynamic_interfaces
}

async function updateOpenWRTDynamicInterfaces(value: string | number | boolean) {
  const enabled = value === true
  const previous = !enabled
  if (enabled) {
    try {
      await ElMessageBox.confirm(
        '该设置仅适用于 OpenWrt。确认把当前拨号数据网卡交给 netifd 展示？',
        '启用 OpenWrt 接口映射',
        { confirmButtonText: '启用', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      openWRTDynamicInterfaces.value = previous
      return
    }
  }
  savingSystemSettings.value = true
  const result = await systemService.saveSystemSettings({ openwrt_dynamic_interfaces: enabled })
  savingSystemSettings.value = false
  if (!result.ok) {
    openWRTDynamicInterfaces.value = previous
    ElMessage.error(result.error.message || 'OpenWrt 接口映射更新失败')
    return
  }
  ElMessage.success(enabled ? 'OpenWrt 接口映射已启用' : 'OpenWrt 接口映射已关闭')
}

function applyClassicTheme(value: string | number | boolean) {
  if (value === true) {
    applyClassic()
    ElMessage.success('已应用经典主题')
    return
  }
  restoreNavyTheme('navy-night')
  ElMessage.success('已恢复海军主题')
}


async function loadNotifications() {
  try {
    const result = await settingsStore.fetchNotifications()
    if (!result.ok) throw new Error(result.error.message || '通知配置加载失败')
    syncWebhookHeaderRowsFromSettings()
  } catch {
    ElMessage.error('通知配置加载失败')
  }
}

function openAPIDocs() {
  const docsURL = String(systemInfo.value.docs?.swagger_ui || '').trim()
  if (!docsURL) {
    ElMessage.warning('API 文档入口暂不可用')
    return
  }
  window.open(docsURL, '_blank', 'noopener,noreferrer')
}

async function saveNotifications() {
  try {
    const result = await settingsStore.saveNotificationsFromForms()
    if (!result.ok) throw new Error(result.error.message || '通知配置保存失败')
    const applied = result.data.applied
    const warning = result.data.warning
    if (applied === false && warning) {
      ElMessage.warning(warning)
    } else {
      ElMessage.success('通知配置已保存（已写入 config.yaml）')
    }
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '通知配置保存失败')
  }
}

async function testWebhookNotification() {
  try {
    const result = await settingsStore.testWebhookFromForm()
    if (!result.ok) {
      throw new Error(result.error.message || 'Webhook 测试失败')
    }
    const data = result.data
    if (data.ok) {
      ElMessage.success(data.message || '测试通知已发送')
      return
    }
    if (Array.isArray(data.failed_urls) && data.failed_urls.length > 0) {
      ElMessage.error(`${data.message}\n失败 URL: ${data.failed_urls.join(', ')}`)
      return
    }
    ElMessage.error(data.message || 'Webhook 测试失败')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : 'Webhook 测试失败')
  }
}

function addWebhookUrl() {
  if (!webhookSettings.value.urls) {
     webhookSettings.value.urls = []
  }
  webhookSettings.value.urls.push('')
}

function removeWebhookUrl(index: number) {
  webhookSettings.value.urls.splice(index, 1)
}

// 自定义请求头以「行」形式编辑（rows 为唯一编辑源），保存时单向回写为 map。
// 受保护的系统头由后端强制覆盖。
const PROTECTED_WEBHOOK_HEADERS = new Set(['content-type', 'x-hideck-signature'])
// 常用请求头预设，下拉可选；filterable + allow-create 也允许自行输入其它名称
const COMMON_WEBHOOK_HEADERS = [
  'Authorization',
  'X-Api-Key',
  'X-Auth-Token',
  'X-Webhook-Token',
  'X-Signature',
  'X-Request-Id',
  'Accept',
  'User-Agent'
]
// 每行带稳定 id，避免用数组下标作 v-for key 时，删除中间行后 el-select 复用实例残留选项
let webhookHeaderUid = 0
const webhookHeaderRows = ref<{ id: number; key: string; value: string }[]>([])

// 加载完成后调用，把已保存的 headers map 转换为可编辑的行
function syncWebhookHeaderRowsFromSettings() {
  const headers = webhookSettings.value.headers || {}
  webhookHeaderRows.value = Object.entries(headers).map(([key, value]) => ({
    id: webhookHeaderUid++,
    key,
    value: String(value ?? '')
  }))
}

// 行变化时单向回写为 map（丢弃空 key 与受保护头）。无反向 watch，故不会回环。
watch(
  webhookHeaderRows,
  (rows) => {
    const map: Record<string, string> = {}
    for (const row of rows) {
      const key = String(row.key || '').trim()
      if (!key || PROTECTED_WEBHOOK_HEADERS.has(key.toLowerCase())) continue
      map[key] = String(row.value ?? '')
    }
    webhookSettings.value.headers = map
  },
  { deep: true }
)

function addWebhookHeader() {
  webhookHeaderRows.value.push({ id: webhookHeaderUid++, key: '', value: '' })
}

function removeWebhookHeader(index: number) {
  webhookHeaderRows.value.splice(index, 1)
}

async function testBarkNotification() {
  try {
    const result = await settingsStore.testBarkFromForm()
    if (!result.ok) {
      throw new Error(result.error.message || 'Bark 测试失败')
    }
    const data = result.data
    if (data.ok) {
      ElMessage.success(data.message || '测试通知已发送')
      return
    }
    if (Array.isArray(data.failed_urls) && data.failed_urls.length > 0) {
      ElMessage.error(`${data.message}\n失败 URL: ${data.failed_urls.join(', ')}`)
      return
    }
    ElMessage.error(data.message || 'Bark 测试失败')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : 'Bark 测试失败')
  }
}

async function testEmailNotification() {
  try {
    const result = await settingsStore.testEmailFromForm()
    if (!result.ok) {
      throw new Error(result.error.message || 'Email 测试失败')
    }
    const data = result.data
    if (data.ok) {
      ElMessage.success(data.message || '测试邮件已发送')
      return
    }
    ElMessage.error(data.message || 'Email 测试失败')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : 'Email 测试失败')
  }
}

async function testWeComNotification() {
  try {
    const result = await settingsStore.testWeComFromForm()
    if (!result.ok) {
      throw new Error(result.error.message || '企业微信测试失败')
    }
    const data = result.data
    if (data.ok) {
      ElMessage.success(data.message || '企业微信测试通知已发送')
      return
    }
    const failure = data.failed_count ? `（失败目标：${data.failed_count}）` : ''
    ElMessage.error(`${data.message || '企业微信测试失败'}${failure}`)
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '企业微信测试失败')
  }
}

function addWeComUrl() {
  if (weComSettings.value.urls.length >= MAX_WECOM_URLS) return
  weComSettings.value.urls.push('')
}

function removeWeComUrl(index: number) {
  weComSettings.value.urls.splice(index, 1)
}

function addBarkUrl() {
  if (!barkSettings.value.urls) {
     barkSettings.value.urls = []
  }
  barkSettings.value.urls.push('')
}

function removeBarkUrl(index: number) {
  barkSettings.value.urls.splice(index, 1)
}



watch(() => emailForm.value.smtp_port, (newPort) => {
  if (Number(newPort) === 465) {
    emailForm.value.use_ssl = true
  }
})



import { systemService, type UpdateInfo } from '../services/system'

const checkingUpdate = ref(false)
const updateInfo = ref<UpdateInfo | null>(null)

async function doCheckUpdate() {
  checkingUpdate.value = true
  try {
    const res = await systemService.checkUpdate()
    if (!res.ok) throw new Error(res.error.message || '检查更新失败')
    updateInfo.value = res.data
    if (!res.data.has_update) {
      ElMessage.info(res.data.release_note || '当前已是最新版本')
    }
  } catch (e: any) {
    ElMessage.error(e.message || '检查更新失败')
  } finally {
    checkingUpdate.value = false
  }
}

function showUpdateInstructions() {
  const info = updateInfo.value
  if (!info) return

  const instructions = info.is_docker
    ? 'docker compose pull\ndocker compose up -d'
    : '请按当前安装方式重新部署对应版本；本程序不会在运行中覆盖自身文件。'
  const content = h('div', { class: 'space-y-3 text-sm leading-6' }, [
    h('p', `当前版本：${info.current_version}，最新版本：${info.latest_version}`),
    h('p', info.release_note),
    h('pre', {
      class: 'overflow-x-auto rounded-[var(--ui-radius-lg)] border border-[var(--el-border-color)] bg-[var(--el-fill-color-light)] p-3 text-xs whitespace-pre-wrap'
    }, instructions)
  ])
  ElMessageBox.alert(content, info.is_docker ? 'Docker 更新方法' : '更新方法', {
    confirmButtonText: '知道了',
    type: 'warning'
  })
}

onMounted(() => {
  loadNotifications()
  loadSystemInfo()
  loadSystemSettings()
  loadPasswordStatus()
})

</script>

<template>
  <div class="max-w-5xl mx-auto">
    <PageHeader title="系统设置" subtitle="管理网关参数与运行信息" />

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
      <!-- Security Card -->
      <div class="ui-card p-8 relative overflow-hidden group">
         <div class="absolute top-0 right-0 w-40 h-40 bg-indigo-500/5 rounded-bl-full -mr-10 -mt-10 transition-transform group-hover:scale-110"></div>

         <div class="flex items-center gap-3 mb-6 relative z-10">
            <div class="w-12 h-12 rounded-xl bg-indigo-50 dark:bg-indigo-500/10 flex items-center justify-center text-indigo-600 dark:text-indigo-400">
               <el-icon size="24"><Key24Regular /></el-icon>
            </div>
            <div>
               <h3 class="text-lg font-bold text-gray-800 dark:text-gray-100">安全</h3>
               <p class="text-xs text-gray-500">更新访问凭证</p>
            </div>
         </div>

         <div class="space-y-4 relative z-10">
             <p v-if="passwordManagedByEnvironment" class="text-sm text-amber-600">密码由部署环境管理，请在部署配置中修改，控制台不会覆盖它。</p>
             <div class="space-y-1">
                <label class="text-xs font-bold text-gray-500 uppercase tracking-wider">当前密码</label>
                <el-input :disabled="!passwordStatus || passwordManagedByEnvironment" v-model="passwordForm.old_password" type="password" show-password placeholder="••••••••" size="large" />
             </div>
             <div class="space-y-1">
                <label class="text-xs font-bold text-gray-500 uppercase tracking-wider">新密码</label>
                <el-input :disabled="!passwordStatus || passwordManagedByEnvironment" v-model="passwordForm.new_password" type="password" show-password placeholder="••••••••" size="large" />
             </div>
             <div class="space-y-1">
                <label class="text-xs font-bold text-gray-500 uppercase tracking-wider">确认新密码</label>
                <el-input :disabled="!passwordStatus || passwordManagedByEnvironment" v-model="passwordForm.confirm_password" type="password" show-password placeholder="••••••••" size="large" />
             </div>

             <div class="pt-4">
                 <el-button type="primary" :disabled="!passwordStatus || passwordManagedByEnvironment" :loading="changingPassword" @click="changePassword" size="large" class="w-full !border-0">
                   <el-icon><Save24Regular /></el-icon>
                   更新凭证
                 </el-button>
             </div>
         </div>
      </div>

      <!-- System Info Card -->
      <div class="ui-card p-8 relative overflow-hidden group">
         <div class="absolute top-0 right-0 w-40 h-40 bg-green-500/5 rounded-bl-full -mr-10 -mt-10 transition-transform group-hover:scale-110"></div>

         <div class="flex items-center gap-3 mb-6 relative z-10">
            <div class="w-12 h-12 rounded-xl bg-green-50 dark:bg-green-500/10 flex items-center justify-center text-green-600 dark:text-green-400">
               <el-icon size="24"><Server24Regular /></el-icon>
            </div>
            <div>
               <h3 class="text-lg font-bold text-gray-800 dark:text-gray-100">系统信息</h3>
               <p class="text-xs text-gray-500">运行环境</p>
            </div>
         </div>

         <div class="space-y-4 text-sm relative z-10">
            <div class="p-3 bg-gray-50 dark:bg-white/5 rounded-lg">
              <FieldRow label="版本" :value="systemInfo.version" monospace>
                <div class="flex items-center justify-end gap-3">
                  <el-button size="small" type="primary" class="!border-0" :loading="checkingUpdate" @click.stop="doCheckUpdate">
                    检查更新
                  </el-button>
                  <span>{{ systemInfo.version || 'Unknown' }}</span>
                </div>
              </FieldRow>
            </div>

            <div v-if="updateInfo?.has_update" class="p-4 bg-amber-50 dark:bg-amber-500/10 rounded-lg border border-amber-200 dark:border-amber-500/20">
               <div class="flex items-center gap-2 text-amber-800 dark:text-amber-200 mb-2 font-bold text-[13px]">
                 <el-icon><Alert24Regular /></el-icon>发现新版本: {{ updateInfo.latest_version }}
               </div>
               <div class="text-xs text-amber-700 dark:text-amber-300/80 mb-4 whitespace-pre-wrap max-h-32 overflow-y-auto pr-2 custom-scrollbar">
                 {{ updateInfo.release_note || '暂无更新说明' }}
               </div>
               <el-button type="warning" @click="showUpdateInstructions" class="w-full !border-0">
                 {{ updateInfo.is_docker ? '查看 Docker 更新方法' : '查看更新方法' }}
               </el-button>
            </div>
            <div class="p-3 bg-gray-50 dark:bg-white/5 rounded-lg">
              <FieldRow label="构建时间" :value="systemInfo.build_time" monospace />
            </div>
            <div class="ui-panel-muted p-3 flex flex-wrap items-center justify-between gap-3">
              <div><div class="text-sm font-bold">OpenWrt 动态接口</div><div class="text-xs text-gray-500">在系统网络中映射设备接口</div></div>
              <el-switch v-model="openWRTDynamicInterfaces" :loading="savingSystemSettings || loadingSystemSettings" @change="updateOpenWRTDynamicInterfaces" />
            </div>
            <div class="p-3 bg-gray-50 dark:bg-white/5 rounded-lg">
              <FieldRow label="配置路径" :value="systemInfo.config" monospace copyable />
            </div>
            <div class="p-3 bg-gray-50 dark:bg-white/5 rounded-lg">
              <FieldRow label="交流群" value="https://t.me/vohive" monospace copyable />
            </div>
            <div class="ui-panel-muted px-4 py-4">
              <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div class="min-w-0">
                  <div class="flex items-center gap-3">
                    <div class="w-9 h-9 rounded-xl bg-blue-50 dark:bg-blue-500/10 flex items-center justify-center text-blue-600 dark:text-blue-400">
                      <el-icon size="18"><DocumentText24Regular /></el-icon>
                    </div>
                    <div>
                      <div class="text-sm font-bold text-gray-800 dark:text-gray-100">API 文档</div>
                      <div class="text-xs text-gray-500">打开后端直出的 OpenAPI 页面</div>
                    </div>
                  </div>

                </div>
                <el-button
                  type="primary"
                  class="self-start sm:self-center shrink-0 !border-0"
                  :disabled="!systemInfo.docs?.swagger_ui"
                  @click="openAPIDocs"
                >
                  <el-icon><DocumentText24Regular /></el-icon>
                  打开 API 文档
                </el-button>
              </div>
            </div>
         </div>
      </div>

      <div class="notify-card ui-card p-8 relative overflow-hidden group lg:col-span-2">
         <div class="absolute top-0 right-0 w-40 h-40 bg-purple-500/5 rounded-bl-full -mr-10 -mt-10 transition-transform group-hover:scale-110"></div>

         <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6 relative z-10">
            <div class="flex items-center gap-3">
               <div class="w-12 h-12 rounded-xl bg-purple-50 dark:bg-purple-500/10 flex items-center justify-center text-purple-600 dark:text-purple-400">
                  <el-icon size="24"><Alert24Regular /></el-icon>
               </div>
               <div>
                  <h3 class="text-lg font-bold text-gray-800 dark:text-gray-100">通知</h3>
                  <p class="text-xs text-gray-500">Telegram / 飞书 / QQ / Webhook</p>
               </div>
            </div>
            <el-button type="primary" :loading="savingNotifications" :disabled="loadingNotifications" @click="saveNotifications" class="!border-0">
              <el-icon><Save24Regular /></el-icon>
              保存通知配置
            </el-button>
         </div>

         <div v-if="loadingNotifications" class="p-6 text-sm text-gray-500 dark:text-gray-400">正在加载通知配置…</div>

         <div v-else class="relative z-10 w-full overflow-hidden">
            <el-tabs v-model="activeNotifyTab" class="notify-tabs">
              <el-tab-pane label="Telegram Bot" name="telegram" class="pt-2"><TelegramNotificationTab /></el-tab-pane>
              <el-tab-pane label="QQ Bot" name="qq" class="pt-2"><QQNotificationTab /></el-tab-pane>
            </el-tabs>
         </div>
      </div>
    </div>
    <PersonalHealth class="mt-6" />
  </div>
</template>

<style scoped>
:deep(.notify-card .el-input-number) {
  width: 100%;
}
:deep(.settings-notify-tabs) {
  border: none;
  background: transparent;
}
:deep(.settings-notify-tabs .el-tabs__header) {
  margin-bottom: 24px;
  background-color: var(--el-fill-color-light);
  border-radius: 12px;
  border-bottom: none;
  display: inline-flex;
  padding: 4px;
}
:deep(.settings-notify-tabs .el-tabs__nav-wrap::after) {
  display: none;
}
:deep(.settings-notify-tabs .el-tabs__active-bar) {
  display: none;
}
:deep(.settings-notify-tabs .el-tabs__item) {
  height: 38px;
  line-height: 38px;
  padding: 0 20px !important;
  border-radius: 8px;
  margin-right: 4px;
  color: var(--el-text-color-regular);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  font-weight: 500;
}
:deep(.settings-notify-tabs .el-tabs__item:last-child) {
  margin-right: 0;
}
:deep(.settings-notify-tabs .el-tabs__item:hover) {
  color: var(--el-color-primary);
}
:deep(.settings-notify-tabs .el-tabs__item.is-active) {
  background-color: var(--el-bg-color);
  color: var(--el-color-primary);
  font-weight: 600;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05), 0 2px 8px rgba(0, 0, 0, 0.03);
}
</style>
