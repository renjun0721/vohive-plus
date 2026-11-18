# VoHive Plus 0.1.18-personal-classic

在 0.1.17 基础上选择性适配 HiDeck 267496fc、157e7588（杨伟杰，PR #116）以及 5e30475b（yibaiba）中仅 IMEI 缓存部分。跨短信中心的长短信按发送者、收件人和引用号合并；串口节点替换后丢弃旧 IMEI 缓存，迟到结果不能写入新节点。保留作者、许可证、经典 Vue UI、无限设备、漫游白名单及全部既有个人版功能。

保留个人版 DeviceID/IMSI 持久化模型，未导入上游 ICCID schema 或跨设备迁移。升级前含 sc 的旧未完成分组保持隔离，可能按原 TTL 过期；没有清空或迁移短信数据库。实时身份重探测、PCM、modemvoice 等待呼叫/CLCC 与 early media 涉及串口和媒体生命周期，继续暂缓。VoCat d745a950 源码未变；既有 WWAN/TUN/QQ、IMS-SA、RF 和运营商策略暂缓项保留。

旧版真实复现五种跨 SMSC 合并/重启恢复失败及串口复用缓存失败。候选短信分片三轮 race、完整 modem 两轮 race、后端/API/device/db/notify/eSIM/vowifihost/phone/VoLTE/QMI、engine/runtimehost/smsdelivery/netstack、关键 IMS/host/漫游恢复 race 和 vet 通过。未删除断言或跳过失败。经典前端及内嵌资产不变，未重跑前端检查；完整 imscore 未运行，既有 MSS/UDP-stop/媒体环境限制未宣称解决。

上述是本地模拟和单元测试；实机仅只读健康验证，未执行真实短信、通话或通知送达验收。生产只通过既有 deploy-tested-upgrade.sh 部署，旧程序、配置及 SQLite 在线一致备份和部署收据保存在私有维护目录。程序须在含语音库的容器运行，不支持 iStoreOS musl 原生执行。
