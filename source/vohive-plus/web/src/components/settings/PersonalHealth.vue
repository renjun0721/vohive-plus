<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '../../stores/auth'

type Check = { key: string; ready: boolean; detail: string }
type Report = { platform: string; package_manager: string; serial_port_count: number; qmi_control_count: number; checks: Check[] }
const report = ref<Report | null>(null)
const loading = ref(false)
const error = ref('')
const labels: Record<string, string> = {
  at_ports: 'AT 串口', qmi_controls: 'QMI 控制口', qmi_proxy: 'QMI 共享服务',
  qmicli: 'QMI 诊断工具', xfrm_ipsec: 'VoWiFi / IPsec 内核接口',
  codec_amr: 'AMR 语音编解码', codec_amrwb: 'AMR-WB 高清语音', codec_mp3: 'MP3 录音编码'
}
async function refresh() {
  loading.value = true
  error.value = ''
  try { report.value = (await api.get<Report>('/personal/health')).data }
  catch { error.value = '环境诊断读取失败，请检查服务连接。' }
  finally { loading.value = false }
}
onMounted(refresh)
</script>

<template>
  <section class="personal-health p-5 rounded-2xl border border-[var(--ui-border)] bg-[var(--ui-surface)]">
    <div class="flex items-center justify-between gap-3 mb-3">
      <div>
        <h3 class="text-lg font-bold">个人版环境诊断</h3>
        <p class="text-sm text-[var(--ui-text-muted)]">只读检查，不安装软件、不修改设备。基于 HiDeck，参考 VoCat 的诊断方式。</p>
      </div>
      <el-button :loading="loading" @click="refresh">重新检查</el-button>
    </div>
    <p v-if="error" role="alert">{{ error }}</p>
    <template v-if="report">
      <p class="text-sm mb-3">{{ report.platform }} · {{ report.package_manager }} · {{ report.serial_port_count }} 个串口 / {{ report.qmi_control_count }} 个控制口</p>
      <dl class="space-y-2">
        <div v-for="check in report.checks" :key="check.key" class="flex flex-wrap gap-2 items-center">
          <dt class="min-w-48">{{ labels[check.key] || check.key }}</dt>
          <dd><el-tag :type="check.ready ? 'success' : 'warning'">{{ check.ready ? '可用' : '需要检查' }}</el-tag></dd>
          <dd class="text-sm break-all text-[var(--ui-text-muted)]">{{ check.detail }}</dd>
        </div>
      </dl>
      <p class="text-sm mt-3 text-[var(--ui-text-muted)]">接口和编解码库可用不代表已验证运营商注册或真实通话。电话功能请从 HTTPS 页面打开。</p>
    </template>
  </section>
</template>
