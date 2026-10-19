# VoHive Plus 0.1.7-personal-classic

以个人版 0.1.6 为基线，代码级核对 HiDeck v2.1.32 及主分支 d5f64cf、VoCat 162f7b6，选择性移植 8 项 HiDeck 提交，不整体覆盖个人版。完整 SHA、作者与暂缓项见 [selected-upstream.json](selected-upstream.json)。

## 本轮修复

- 超过 1300 字节的自动传输 REGISTER 优先走 TCP；连接等待有界，保留 UDP 重试时间。TCP 连接失败可回退，已有 SIP 明确拒绝不再换协议绕过，UDP 不可用时也不重新打开已失败的 TCP。
- Globe PH 刷新注册遇到 IP/端口冲突时，使用新安全端口重新注册，预设优先 TCP。
- VoWiFi 单条短信入库失败返回临时错误，不再将整个运行实例标记为内存满；显式内存满及 RP-SMMA 行为保留。
- 对 DITO 515-66 的零时区 SMSC 时间，仅在改为菲律宾时间后接近接收时间时纠正；真实 UTC 和旧积压短信保持原值。蜂窝/VoWiFi 入库与通知使用相同时间。
- 浏览器通话 IMS RTP 换 SSRC 或序号跳变时重新排队，丢弃过期包，浏览器侧序号/时间戳持续向前，避免后续音频被当成重放。

新增运行日志为中文。保留经典 UI、无限设备、QQ/TG、eSIM AID fallback、USSD 渠道、LuCI 按钮/中文日志、USB 拓扑、网络、证书和反代配置。

## 验证与限制

- device/db/backend/vowifihost/API/esim/notify/phone 包测试通过；相关 race 和 vet 通过。
- 受支持 IMS 测试共 660 项通过；保留 7 项既知内核/锁环境限制，详见已有 0.1.6 说明，未删除其断言。runtimehost、运营商策略测试通过。
- 超长注册问题有旧版对照：0.1.6 同一回归场景超时，新版通过。首次 race 暴露上游测试服务器过早 RST 导致 Dial 时序不确定；改为读到完整 REGISTER 再断开，保留原断言，最终专项连续 5 轮通过。另移除旧 phone 测试辅助函数不可达的 return，使 vet 无诊断。
- 前端源码与 0.1.6 无变化，嵌入产物一致，沿用其已构建经典 UI；本轮没有重复声称重新跑前端 269 测试。
- 实际 API 确认 0.1.7、4 个 healthy/control_online/physical_present Worker、全部 reg_status=5、设备上限 0；短信历史联系人数量 13/10/10/1 与停服前一致，配置 SHA-256 不变。
- 本地 HTTP、Tailscale HTTPS、公网 VPS HTTPS 的 /ping 均通过，HTTPS 证书校验启用。
- 未主动发送短信、Bot 消息、拨号、切卡或拔插。真实运营商通话、QQ/TG 新短信送达及长期稳定性尚未验证。

大型自动应答/WebDAV、完整 SIM/多 IMSI 事务框架、VoLTE 原生 UAC/自适应音频恢复链暂缓；VoCat 的 AT Session、原始短信数据库、USSD 和通知调度使用另一架构，不直接替换本项目。当前版未做新 ICE 网卡裁剪，保留 Tailscale/蜂窝候选；最新上游也已撤回物理网卡过滤。

## 构建与恢复

Go 1.27.1、Node.js 24、pnpm 11.25.0；源码和构建脚本位于 source/vohive-plus，运行依赖见 source/deployment/Personal.Dockerfile。程序 SHA-256：`9b77ef393b983a2da0a5d5abfa1bfec78ee08f4544dc51b8895895806f78560e`。Release 提供 gzip 程序、校验和、构建信息和上游清单。

部署入口先检查无活动通话，再保存原程序、配置及 SQLite 在线一致备份；升级后只读核对通过。当前容器保留原挂载，新重建镜像 vohive-plus-personal:0.1.7。不将升级后的新短信数据库替换为旧快照；回滚只停服、恢复备份旧程序并启动。
