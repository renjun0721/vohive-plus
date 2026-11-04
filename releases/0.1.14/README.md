# VoHive Plus 0.1.14：每会话刷新 WebRTC 公网域名

基于已部署的 0.1.13，适配 HiDeck 的最终 DNS 刷新修复。公网域名地址变化后，新媒体会话使用新 IP，已有会话保持原地址；DNS 失败明确返回，不沿用旧地址。解析遵循调用者取消和截止时间，保留启动校验及地址变化中文日志。空配置及 IP 字面量不增加 DNS 查询。

上游作者 yibaiba、BFNOC；具体提交见 selected-upstream.json。保留全部原作者、LICENSE、THIRD_PARTY_NOTICES，以及既有 VoCat SOCKS5 许可和出处声明。个人版 Create 保持三参数接口，没有引入上游 AccessHost 或其他媒体数据面。

独立候选分支 codex/scheduled-0.1.14-media-dns，实际候选提交 3e78a77dee80b7390b39a05e8c93eba1f9ff0b9e；实际公开运行代码提交与后续元数据提交分别记录于 build-info.json 和发布证据。构建时间 2026-11-04T02:06:05Z，Go 1.27.1，CGO_ENABLED=0，linux/amd64。程序 SHA-256 为 0e32729ebeb5add2ebfea636018cd5d84482296a335615564900691f791f2002；gzip SHA-256 为 1edd6cec72d6f1efbb38f627043da4d42082371b20428295d23baa400cffc3af。

同一兼容回归通过实际 MediaManager.Create 和本进程回环 DNS fixture，在旧 0.1.13 复现旧 IP 残留与解析失败后仍创建会话；候选 DNS/取消/截止时间回归 10 轮 race 通过。相关 backend/API/device/db/notify/eSIM/vowifihost/phone/VoLTE/QMI/personal/主程序测试、引擎 ipsec/swu/runtimehost 测试、引擎与 API/媒体 race 两轮、漫游/注册 race 两轮及相应 vet 均通过。pkg/logger 无测试文件。测试证明绑定候选程序哈希、源码提交及真实日志；日志保存在本机私有运行目录。

仅通过 maintenance/deploy-tested-upgrade.sh 预检、SQLite 在线一致备份、受控部署和健康验证。现有容器及挂载保留，程序更新；干净的含语音库 0.1.14 镜像另供重建，不代表现存容器镜像标签切换。无网络、无设备、只读的 doctor 确认 AMR/AMR-WB/MP3 与 QMI 工具可用。回滚仅恢复旧程序，不能以旧数据库覆盖新短信。

保留设备无限制、经典中文 UI、未插入显示、QQ/TG、LuCI、eSIM AID fallback、USSD 渠道、既有 VoWiFi/运营商策略和按卡漫游白名单（联通优先、移动备用、排除电信）。未修改实际用户配置、网络、路由、DNS、证书、VPS、SIM/MBN/射频或 cron。

限制：经典前端 288 个源码文件和 112 个嵌入文件原样复用，未跑前端测试、类型检查或构建。完整 imscore 未重跑，前轮 MSS ENOPROTOOPT/UDP-stop 并发限制仍未解决。初轮未用 import 编译失败和回归夹具 loopback-only mux 缺少 ICE 候选的失败日志均保留；修正适配/夹具后重跑，没有删除断言或跳过失败。

本次实机只读检查覆盖 API 健康摘要、本地 HTTP、Tailscale 和 VPS 公网 HTTPS 管理入口。未进行真实短信、通话、录音、通知发送、USSD、切卡、余额任务或运营商业务验收；回环 WebRTC 测试与只读健康检查不等于真实业务验收。HiDeck PCM 媒体重绑定/teardown 架构继续暂缓，VoCat 新 SSO/JOSE 路径不适用于现有个人版。
