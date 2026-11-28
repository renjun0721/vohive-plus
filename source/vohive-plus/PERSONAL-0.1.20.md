# VoHive Plus 0.1.20-personal-classic

基于 0.1.19 选择性适配 VoCat 3db888b7ad381274b816c4b77b48c6a434a99992（Tardisyuan，2026-11-27，PR #228）：经典短信页实际使用的内联气泡加入 whitespace-pre-wrap 和 break-words。接收和发送历史保留文本换行、连续空格，并折行长网址；不改变短信内容、存储、发送策略或通知。

未使用的 SmsMessageTimeline.vue 组件已有类似样式，但当前 Sms.vue 不渲染它，不能作为现有经典页已修复的证据。实际运行产物也确认未包含气泡换行类。本次只改变活跃气泡的一处 class、构建版本和本说明；保留经典布局及所有后端和第三方运行代码、原作者与许可证。

构建前执行现有前端测试、类型检查和 Vite 构建；候选执行后端/API/device/db/notify/eSIM/phone/VoWiFi、关键 race 和 vet。离线 Chromium 使用实际气泡类、产物 CSS 和虚构文本，对照旧版换行折叠及长词溢出，并验证候选修复。这些是本地模拟/布局/单元检查，实机只读健康验证不能替代真实短信、通话或通知送达验收；本轮不主动执行上述业务操作。

HiDeck 8a1c29d85fa7f5df2e801aa9627ef46dab92a31a 未变。既有 ML307、短信身份/存储迁移、SIM 生命周期，以及 VoCat 低 MTU/分片/IMS 联动等暂缓项继续保留，未整体更新通信栈、改选网或用户配置。

上游原始来源：https://github.com/MengMengCode/VoCat/commit/3db888b7ad381274b816c4b77b48c6a434a99992 。上游作者 Tardisyuan <66050966+Tardisyuan@users.noreply.github.com>；当前项目与第三方原有许可说明保持完整。
