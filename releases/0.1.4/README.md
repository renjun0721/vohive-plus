# VoHive Plus 0.1.4-personal-classic

2026-10-10。现有个人版在 VoWiFi 重连、密钥更新和部分运营商认证时缺少上游修复；本版本以 0.1.3 完整源码为基线，选择性纳入 8 项兼容更新。

## 改动

- AIS EAP-AKA 挑战响应补全。
- CHILD_SA 重协商和删除使用正确的入站/出站 SPI 方向。
- 重连时清理快速重认证身份，避免错误复用。
- 增加 SWu 会话失败原因及密钥重协商完成日志。
- Lebara UK 的 IKE SA 更新周期，以及 Spark NZ 的设备身份与拒绝重协商处理。
- DEVICE_IDENTITY 请求响应和运营商覆盖配置传递。

完整提交和 SHA 见 [selected-upstream.json](selected-upstream.json)。保留 EC20/EC25、设备无限制、eSIM AID 重试、USSD 会话渠道、经典中文界面、QQ/TG、LuCI 卡片开关/重启和 Tailscale 网络定制。没有新增模组直拨框架、整体同步 VoCat 或更换数据库。

## 验证

- SWu、runtimecore、policy、runtimehostcarrier 和 runtimehost 子包测试通过。
- API、device、esim、phone、notify、vowifihost 测试通过。
- 前端 269 项测试、类型检查和构建通过。
- USSI 与可运行的 IMS 测试通过。6 项 TCP MSS 测试在当前内核环境失败，0.1.3 基线出现完全相同的失败；另跳过之前已确认会卡住的 `TestProtectedUDPShutdownWhileRegisterLockHeld`。不宣称 IMS 全套通过。
- 运行容器内只读诊断：QMI、XFRM/IPsec、AMR/AMR-WB/MP3 库正常。
- 部署后健康检查与 /ping 正常；三个工作进程恢复，与升级前数量一致。配置文件 SHA-256 未改变。
- 本地 HTTP、Tailscale HTTPS 正常。Chromium 使用新版网页和模拟设备 API 验证经典页面、未插入设备显示、受信任 HTTPS，无 JavaScript 异常。

未执行真实运营商切卡、USSD、发短信、QQ/TG 测试消息或拨号。

## 构建与恢复

完整源码位于 `source/vohive-plus`，使用 Go 1.27.1、Node.js 24 和 pnpm 11.25.0。运行程序 SHA-256：`5684a750fbe68f0f4b301133b3f77e03d149e630ec07755c79db66390a148fe7`。

Release 提供 `vohive-plus_0.1.4_linux_amd64.gz` 和 `SHA256SUMS`。此程序应在包含所需语音库的 Docker 运行环境使用，不直接在 musl iStoreOS 中运行。

在现有安装升级时，先备份原程序、配置与数据库，再停止 `/etc/init.d/vohive`，将解压后的程序复制到 `vohive-plus:/usr/local/bin/vohive-plus` 并启动服务。已有容器保留原配置、数据、设备和网络挂载。新建容器可按 `source/deployment/Personal.Dockerfile` 构建 `vohive-plus-personal:0.1.4`。

本次路由器回滚快照：`/mnt/sata2-4/vohive-recovery-backups/20261010-before-014/`，含旧程序、原配置和通过完整性检查的 SQLite 在线备份。需要回滚时仅替换旧程序即可，避免覆盖升级后收到的短信。
