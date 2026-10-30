# VoHive Plus 0.1.12-personal-classic

基于实际部署的 0.1.10，独立适配 HiDeck 的登录失败锁定修复。0.1.11 是未部署的本地短信候选，此版不包含其改动。

正确登录与正确的 rotate 密码认证不再消耗失败限额；同一真实连接 IP 两分钟内达到十次失败后锁定十分钟，返回 Retry-After。并发请求先预留名额，成功仅归还自身名额；旧计数窗口的迟到归还不能影响新窗口，过期 IP 记录定期清理。保留现有不信任转发头的设置和会话凭据模型；经过反代的连接仍按实际连接对端计数。

选取 BFNOC 的 HiDeck 8845a93045f71c858feab7d5bb066056115a3509 与 087afaf0107c047cf46c45a697b6c3f6ba649edb，保留作者、许可和第三方声明。其余 IMS/SA、运营商策略、SOCKS5 NAT-T、VoCat MCP/API Token、eSIM/QMI 架构及通知配置变化暂缓。设备无限制、经典 UI、未插入显示、QQ/TG、eSIM AID fallback、USSD、LuCI、按 SIM 漫游白名单和现有网络配置保持。

旧版在相同 HTTP/认证回归中真实失败，候选通过。相关后端全包、API/主程序/VoLTE/phone 两轮 race、漫游/注册两轮 race、runtimehost/logger/SWu 测试及 vet 通过。扩大 API race 首轮暴露旧测试对 ResponseRecorder 并发读写、重复运行复用内存数据库的问题，旧版同机复现；改为通过真实 loopback HTTP 读取 SSE、每次测试使用独立数据库并关闭连接，保留原断言。专项二十轮与完整最终复跑通过，初轮失败日志单独保留。

未修改 VoWiFi 引擎 884 个 Go 文件。前轮完整 imscore 的 MSS 环境失败及 UDP 停止并发问题仍未解决，本轮未重跑完整 imscore，也未声称其通过。经典前端源码和嵌入产物逐字节复用，没有重跑前端测试、类型检查或构建。

受控部署必须以实际 deployment-receipt.json 为准。程序备份、主配置、私有漫游策略和 SQLite 在线一致快照由既有部署入口生成；不回写旧数据库。无设备、无网络 doctor 仅核对镜像的音频库/qmicli/qmi-proxy，不是实卡验收。

单元/模拟/loopback 测试与只读实机验证分开记录。未发送真实短信、QQ/TG/Bot 消息，未拨号、录音、切卡、复位模组、操作余额任务或进行运营商验收。未修改用户配置、网络、路由、防火墙、DNS、证书或定时任务。
