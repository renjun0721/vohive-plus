# VoHive Plus 0.1.6-personal-classic

2026-10-15。以 0.1.5 完整个人版为基线，按上游报告选择性移植设备恢复隔离和 VoWiFi 协议修复。完整 SHA、适配范围和暂缓项见 [selected-upstream.json](selected-upstream.json)。

## 改动

- 自动 VoWiFi 启动绑定原 Worker 实例、代际和 VoWiFi 生命周期；设备移除时取消 SIM/恢复定时器，忽略旧 QMI/MBIM/AT 回调和迟到的 IP 查询。按个人版接口适配，不引入 MBN 或外发额度框架。
- NAT-T 模式下正确区分 ESP 与 IKE；只有已认证报文刷新存活状态，验证 IKE 响应后再完成事务，并清理陈旧的 IKE SA。保留上游对本地密钥错误及已认证语法错误的严格处理。
- IMS 短信 RP 回执连接无响应时回退到注册连接，并记住可用路径。额外修复上游边界：独立入站连接出现 EPIPE 等致命写入错误时，不再连带拆除正常的注册连接；新增真实错误路径回归测试。
- IKE_AUTH 内层 IPv6 地址使用空值请求；REGISTER 按本机 Contact 取有效期，IPv6 不同文本表示按地址匹配。补齐个人版缺少的 Contact 校验前置修复。
- Globe PH 预设增加 weconnect.globe.com.ph，不覆盖其他运营商或运行配置。保留 DEVICE_IDENTITY 全局默认策略。
- 为个人版的生命周期互斥增加初始化等待保护：等待 Worker 启动完成时释放锁，避免启动发布与停止操作相互等待。

保留设备数量无限制、eSIM AID fallback、USSD 会话渠道、QQ/TG、经典界面、LuCI 中文日志与服务按钮、USB 拓扑及网络定制。通知过滤、自动 MBN、上游后续整套设备/SIM 事务框架未纳入此版。

## 验证

- device、db、vowifihost、API、backend、esim、notify 全包测试通过；前端 269 项测试、类型检查和构建通过。
- VoWiFi engine 全包测试通过；IMS、voice、runtimehost 相关测试通过，除以下已确认的环境限制。
- Worker 替换、定时器、初始化等待、短信身份、RP 回执、Contact 和 IKE/NAT-T 的关键 race 检查通过；相关 vet 检查通过。
- 构建容器有六项 TCP MSS/IPsec 内核测试报 protocol not available，同机原 0.1.5 对照也失败。另一个既知锁测试不在受支持测试集中。未修改断言来掩盖环境问题。
- 初轮宽泛 race 匹配命中一个使用 20ms 时限的旧短信测试，在并行编译负载下失败；降低负载后单独连续五轮通过。新增回执和注册匹配 race 专项通过。
- 实际只读 API 确认版本 0.1.6、4 个健康 Worker、device_limit=0；四台模组的注册、频段、控制在线状态和历史短信联系人数量与升级前一致。
- 配置校验值不变；QMI、IPsec、AMR/AMR-WB/MP3 只读诊断通过；容器健康，本地 HTTP、Tailscale HTTPS、公网 HTTPS 均正常。

没有主动发送短信、QQ/TG 消息、拨号、切卡或拔插 USB。自动化模拟不等于 One NZ/Globe 真实 VoWiFi、短信通知送达或长时间重连验收，这些仍需使用实际 SIM 验证。

## 构建与回滚

Go 1.27.1、Node.js 24、pnpm 11.25.0。完整源码在 source/vohive-plus，Docker 环境见 source/deployment/Personal.Dockerfile。

程序 SHA-256：`45e545e804baa25617df31517b1f7038cc566a03c16a952ab1a16e648afb4634`。Release 提供压缩程序、SHA256SUMS、构建信息与上游清单。现存容器保留原挂载，只替换程序；新镜像为 vohive-plus-personal:0.1.6。

本地回滚快照：`/mnt/sata2-4/vohive-recovery-backups/20261015-before-016/`，含旧 0.1.5 程序、配置和完整性检查通过的 SQLite 在线备份。停止服务，复制快照旧程序回容器 /usr/local/bin/vohive-plus，再启动即可回滚；避免覆盖升级后收到的新短信数据库。
