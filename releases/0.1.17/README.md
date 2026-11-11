# VoHive Plus 0.1.17-personal-classic

在 0.1.16 基础上选择性适配 BFNOC 的 HiDeck aa5b7ce334cfc0fd48cea37d716805a0011f0599（PR #109）。IMS TCP 在拨出和接受时保存对端地址，连接复位后仍能返回，覆盖 IPv4/IPv6 与接受前复位；保留错误分类、MSS、keepalive 和全部个人版功能。

旧版六种情况真实复现；候选 netstack 全包三轮 race、后端/API/device/db/notify/eSIM/phone/VoLTE/QMI/personal、engine/runtimehost、关键 IMS/host/漫游/USB race 和 vet 通过。前端及内嵌资产逐字节不变，未重跑前端检查；完整 imscore 未运行，既有 MSS/UDP-stop/媒体环境限制保留。协议测试是本地模拟，实机仅只读健康验证，未进行真实短信、通话或通知验收。

55a0eda 的旧 SA timeout 修复依赖尚未适配的 saRetired/refresh_security_switch，继续暂缓；既有运营商、RF、MWI 与其他架构暂缓项保留。VoCat 9919d14a 源码未变，既有 5ber/eSTK/ICCID 逻辑保留。

生产仅通过既有 deploy-tested-upgrade.sh 部署并保存旧程序、配置及 SQLite 在线一致备份；本地维护记录和部署收据为实际结果依据。保留原作者、许可证及第三方声明。安装包须在含语音库的容器内运行，不能作为 iStoreOS musl 原生程序使用。
