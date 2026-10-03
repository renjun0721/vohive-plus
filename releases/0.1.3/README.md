# 0.1.3-personal-classic · 2026-10-03

本版以现有个人版为基础，合入两项兼容修复并同步本地改动。

- eSIM：参考 [VoCat 3cbb6c1](https://github.com/MengMengCode/VoCat/commit/3cbb6c1aa0e9bd77a1da98f6c4d60cd866b4e4c1)。ICCID 启用/禁用仅在标识符或指定兼容错误时，按目标 Profile 的 ISD-P AID 重试；保留 refresh 和 CAT Busy 处理，不对策略拒绝或传输错误盲目重试。
- USSD：参考 [VoCat a1273f0](https://github.com/MengMengCode/VoCat/commit/a1273f0281cf74b288bab329d734c83c5afc3f48)。现有 vowifi-go 已支持网络 Session ID、SIP INFO 续接和 BYE 取消；修复 API 路由，让后续输入与取消继续使用原会话渠道，IMS 断开后不会误发蜂窝 AT+CUSD。
- 未插入设备：将 2026-10-02 的显示补丁迁入完整 TypeScript 源码，保留只读检测和重启缓冲。
- 智能 HTTPS：补齐最新 LuCI 入口和公网/局域网/Tailscale 部署脚本。保留无限设备、QQ/TG、原前端布局和中文日志改动。

本版没有加入 registration-only 自动任务或短信离线导出。

## 验证

- 前端 269 项测试、TypeScript 检查与 Vite 构建通过。
- Go `internal/esim`、`internal/api`、`internal/device`、`internal/notify`、`internal/vowifihost` 测试通过。
- vowifi-go 的完整 `ussi` 测试及 imscore USSI/USSD/BYE 专项测试通过；完整 imscore 套件停在未修改的 `TestProtectedUDPShutdownWhileRegisterLockHeld`，捕获栈后终止，未宣称整套通过。
- Chromium 加载编译后的 HTTPS 前端，使用浏览器本地 API 测试数据：两台在线、两台未插入、只读发现、0 JS 异常；部署前后均通过。
- 中文日志检查通过；只读 doctor 检查可识别两个 QMI 模组和 AMR/AMR-WB/MP3 编解码库。
- 运行程序哈希与产物一致；私网和公网 HTTPS 首页响应正常。
- 未进行真实切卡、运营商 USSD、短信发送或拨号。

程序 SHA-256：`9a4df904a903a68fe0e54816c52712e29dc7abd2dc63d920cd023ce9fc530631`。

升级前本地快照：`/mnt/sata2-4/vohive-recovery-backups/20261003-before-013/`。SQLite 快照完整性检查通过。GitHub 只更新源码、部署脚本及版本产物，未上传新的真实配置或数据库。
