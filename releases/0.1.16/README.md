# VoHive Plus 0.1.16：隔离并发订阅通知与失败退避

基于完整 0.1.15 个人版，选择性适配 HiDeck 的两项修复：旧 Service 的 active/pending NOTIFY 保留新 Service 刚记录的订阅失败退避；并发确认保留已有 MWI 停止状态和网络等待下限。SUBSCRIBE 实际发送时快照共享失败版本，已删除记录的新失败不会复用版本，排队中的刷新不确认另一 Service 的失败。

沿用个人版既有合并注册/退避存储，未引入上游新的按卡 MWI 开关、重检/重置 API 或存储生命周期架构。完整 runtime 停止再启动的退避延续仍待独立适配；此次只修复同一共享存储上的并发通知。选取提交、作者及适配边界见 selected-upstream.json 和源码 PERSONAL-UPSTREAM-016.md；所有原许可证与 VoCat SOCKS5 许可说明保留。

四类回归在当前旧版真实复现：旧 active/pending 通知清掉新失败、排队刷新误确认、三种 MWI 停止原因丢失、失败记录删除后旧通知再次误清。候选相关订阅三轮 race、后端/API/device/db/notify/eSIM/VoWiFi/phone 测试、引擎/runtimehost/runtimecore/policy 测试、关键 host 与漫游/USB race 及 vet 通过。协议测试使用本地模拟 SIP Service，不接触真实模组。

经典前端及内嵌网页资产与 0.1.15 逐字节一致，未重跑前端测试、类型检查或构建。完整 imscore 套件未重跑，既有 MSS ENOPROTOOPT/UDP-stop 并发及媒体 ICE 超时限制不宣称解决。测试证明绑定最终候选源码提交、程序哈希与本轮真实日志。

部署仅使用现有 maintenance/deploy-tested-upgrade.sh，以本轮 deployment-receipt.json 为依据，保存旧程序、主配置、私有漫游白名单及 SQLite 在线一致备份。保留原容器和挂载，新含语音库的干净镜像只供后续重建。无网络、无设备、只读 doctor 核验语音库和 QMI 工具，不作为实卡验收。

保留无限设备、经典中文 UI、未插入显示、LuCI、QQ/TG、EC20/EC25 USB/AT/QMI 身份、eSIM/USSD、既有 VoWiFi/运营商策略、按 SIM 漫游白名单、实时身份与换卡隔离。有风险的 SA/RF/运营商及 MWI 新策略继续暂缓。

实机验收仅只读管理 API、本地/Tailscale/公网 HTTPS 健康检查。未进行真实短信、通话、QQ/TG 通知、USSD、切卡、余额任务、复位或拔插验收；未修改用户配置、网络、证书、VPS、SIM/MBN/射频或 cron。
