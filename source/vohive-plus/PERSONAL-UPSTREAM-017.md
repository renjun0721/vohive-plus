# 0.1.17 IMS TCP 对端地址快照

基于个人版 0.1.16，仅适配 BFNOC 的 HiDeck aa5b7ce334cfc0fd48cea37d716805a0011f0599（PR #109；合并 9e017722）。拨出和接受时保存对端地址，使 IPv4/IPv6 连接复位后 RemoteAddr 仍可用于日志和连接身份判断；保留原错误归一化、MSS 和 keepalive。四个 netstack 文件与固定上游逐字节一致。

55a0eda 的旧 SA timeout 分类修正依赖未移植的 saRetired/refresh_security_switch，继续暂缓；没有引入该架构或运营商/射频策略。VoCat 9919d14a 未变，既有 5ber、eSTK 与 ICCID 查找保留。全部原作者、许可证和第三方说明保留。

回归覆盖 IPv4/IPv6 dialled、accepted、reset-before-accept；使用本地 gVisor loopback，不代表真实短信或通话验收。
