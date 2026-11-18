# 0.1.18 跨短信中心长短信与串口身份缓存

基于个人版 0.1.17，选择性适配 HiDeck 267496fc8dbd1c659e3925c1bfab33e20b22f507、157e75885d62e1938699cde42e59a3c17d95c5e8（作者 杨伟杰，PR #116；合并 d5dcae677fc68e20843145b0101742a6f765fdff），以及 yibaiba 的 5e30475b3ee44ceb5b73621db1a02f4cee4115fa 中仅 IMEI 缓存部分。保留原作者、许可证和所有第三方声明。

长短信分组身份去掉 SMSC（传输元数据），保留发送者、收件人、引用号与位宽；持久化继续按个人版既有 DeviceID/IMSI 限定（未引入上游 ICCID/跨设备迁移 schema）。保留分片碰撞、重复检测、TTL 和通知规则。升级前含 sc 的旧未完成分组保留隔离，既不迁移合并，也不清空数据库；旧未完成组可能按原 TTL 过期。回归全部使用模拟内容，不是真实收信验收。

串口 IMEI 缓存绑定 os.FileInfo 身份与修改时间；节点消失、替换或迟到探测不会污染新节点。仅增加缓存验证，保留个人版 ProbeIMEIContext 取消和关闭语义。没有导入重扫描/恢复时始终重开 AT 探测的改动：需要先验证与当前 Worker 读写互斥及热插拔实际行为。原生 PCM、modemvoice 等待呼叫/CLCC 和 outbound early media 也暂缓，未改变通话控制、音频或 USB 拓扑。

VoCat d745a950 源码未变；此前 WWAN helper、TUN、QQ 绑定体系及 IMS/SA/RF/运营商暂缓项保留。经典 Vue UI、无限设备、漫游白名单、AT/QMI、eSIM、USSD、QQ/TG 与既有网络配置保留。
