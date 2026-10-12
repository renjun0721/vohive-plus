# VoHive Plus 0.1.5-personal-classic

2026-10-12。以 0.1.4 个人版完整源码为基线，选择性合入 HiDeck v2.1.30 的短信身份与 QMI LTE Band 修复。

## 改动

- 已知 IMEI 时按当前物理模组查询 SIM 身份，避免同一设备名称曾对应多个模组时混用历史记录。保留换卡过渡、真实身份冲突检查及旧短信。
- 将 QMI NAS Active Band 原始枚举转换为真实 LTE 频段号；例如 141 显示为 B39。同步修正 FDD/TDD 判断，未知枚举明确显示未识别。

完整提交见 [selected-upstream.json](selected-upstream.json)。保留此前 VoWiFi 修复、多设备无限制、eSIM AID 重试、USSD 渠道、QQ/TG、经典中文界面、LuCI 和 Tailscale 定制。本轮不引入 VoCat AT Session 或按设备通知过滤，不调整射频配置。

## 验证

- 数据库、device、API、backend、esim、notify 测试通过；quectel-qmi-go 全部包测试通过。
- 短信身份和 QMI Band 相关 race 专项检查，以及 db/device/API/backend 的 vet 检查通过。
- 前端 269 项测试、类型检查、构建通过。
- 短信回归覆盖：同名设备更换物理模组、历史别名冲突、离线身份选择、SIM 切换过渡及历史短信保留。
- 在实际运行应用的只读 API 中确认 0.1.5，3 个 Worker 与无限设备配额保持正常；三台模组短信历史接口均可读，联系人数量与升级前一致。
- 两台实际模组从原始 122 正确显示为 B3/FDD；另一台从 136 正确显示为 B34/TDD。信道和注册状态保持不变。
- 配置文件校验值不变，容器健康，QMI/IPsec/语音库诊断通过；本地管理、Tailscale 和 VPS HTTPS 正常。
- Chromium 使用新版网页和模拟设备 API 检查经典页面与未插入显示，无 JavaScript 异常。

实际历史短信读取及频段状态已验证；自动化用例不等同于真实换卡或新短信收发验收。没有执行真实切卡、拨号或主动发送 QQ/TG 测试消息。

## 构建与恢复

Go 1.27.1、Node.js 24、pnpm 11.25.0；完整源码在 `source/vohive-plus`。

程序 SHA-256：`54f13e6600fbc6eb473819daf874ef77f4717a2b43e335f555862202c8c8ff0e`。

Release 提供 `vohive-plus_0.1.5_linux_amd64.gz`、SHA256SUMS、构建元数据及上游提交清单。程序运行于包含语音库的 Docker 环境，不直接在 musl iStoreOS 运行。已有容器升级时保留原配置、数据、网络与设备挂载；重建容器可按 `source/deployment/Personal.Dockerfile` 构建 `vohive-plus-personal:0.1.5`。

本次回滚快照：`/mnt/sata2-4/vohive-recovery-backups/20261012-before-015/`，含 0.1.4 程序、配置和通过完整性检查的 SQLite 在线备份。回滚时停止 `/etc/init.d/vohive`，复制快照中的旧程序回容器 `/usr/local/bin/vohive-plus`，再启动；避免覆盖收到的新短信数据库。
