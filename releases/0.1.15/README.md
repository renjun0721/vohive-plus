# VoHive Plus 0.1.15：QMI 冷却保护与 IMS 订阅退避

基于已部署 0.1.14，独立适配 BFNOC 的四项 HiDeck 修复：USB 复位冷却按物理 USB 路径跨 Worker 保留；QMI 超时日志在清空窗口前记录服务数和名称；运营商覆盖对象深复制 handover 指针；IMS SUBSCRIBE 失败按身份与事件包退避，不再将 Expires 当失败重试间隔。

reg 退避 30 秒、1/2/4/5 分钟，MWI 5/10/20/30 分钟；匹配 active/pending NOTIFY 清除本地失败计数，网络 Retry-After 下限保留，旧订阅 NOTIFY 不清除新失败，Service 重建保留内存退避，新的初始注册 reg 仍可订阅一次。没有引入新 MWI 用户设置、数据库表、Vodafone 预设或 AT 关射频降级。

VoCat 新增的 5ber AID、命中停止及无 AID 切换归属查找在当前个人版已有等价实现，未重复移植；eSTK 双 SE 例外保留。选取提交及作者见 selected-upstream.json，保留所有原作者、许可证和 VoCat SOCKS5 许可说明。

四项故障在当前旧源码真实回归复现。候选订阅/策略三轮 race、QMI manager 两轮 race、相关后端全包、ipsec/swu/runtimehost、API/媒体两轮 race、漫游/USB 三轮 race 和相应 vet 通过。前端 288 文件和 112 内嵌资产原样复用，未运行前端测试/类型检查/构建。测试证明绑定实际源码提交、程序哈希与私有真实日志。

初轮缺少忽略跟踪的内嵌网页产物导致主程序测试初始化失败，复制并逐字节核对原资产后重跑通过。初轮媒体测试出现 ICE 三秒超时；额外五轮 phone 套件在当前旧版和候选均复现 AMR-WB 同一超时，35 个媒体源码/测试文件完全相同，未修改或跳过断言。最终后端单轮及 host race 两轮通过，不宣称额外重复套件全部通过。完整 imscore 未重跑，原 MSS ENOPROTOOPT/UDP-stop 并发限制仍在；本版针对订阅路径运行 race 回归。

部署由既有 maintenance/deploy-tested-upgrade.sh 完成，以实际 deployment-receipt.json 为依据，保存旧程序、主配置、私有漫游白名单和 SQLite 在线一致备份；回滚不能以旧数据库覆盖新短信。原生产容器及挂载保留，含语音库的干净 0.1.15 镜像另供重建，不代表现存容器镜像标签切换。无网络、无设备、只读 doctor 检查可用 codec/QMI 工具，不等于实卡验收。

保留无限设备、经典中文 UI、未插入显示、LuCI、QQ/TG、EC20/EC25 身份/USB 拓扑、eSIM/USSD、既有 VoWiFi 运营商策略，以及联通优先/移动备用/排除电信的按 SIM 漫游白名单与换卡隔离。有风险的射频/MWI 策略及此前架构变化继续暂缓。

本次实机仅只读管理 API 和本地/Tailscale/VPS HTTPS 检查。未进行真实短信、通话、录音、通知发送、USSD、切卡、余额任务、复位或拔插验收；未修改用户配置、网络、DNS、证书、VPS、SIM/MBN/射频或 cron。
