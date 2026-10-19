# VoHive Plus 个人版

本版本由用户授权，为其个人 iStoreOS 环境定制。版本：0.1.7-personal-classic，2026-10-19。

基础源码来自 [HiDeck](https://github.com/yibaiba/hideck)，固定提交 `3fd3d6aaf55924c338c2cf3f092e1af1f588d152`（v2.1.23 之后的源码快照）。HiDeck 继承 VoHive，项目原有作者声明、LICENSE 与 THIRD_PARTY_NOTICES.md 均保留；third_party 中的组件仍适用各自的许可，包括 vowifi-go 的 AGPL-3.0。

[VoCat](https://github.com/MengMengCode/VoCat) 的环境诊断思路用于设计个人版只读检查；本次没有复制 VoCat 的代码，也没有把两个项目的数据库或通信核心直接混合。

个人版以原版 VoHive UI 与布局为基础增加功能。前端基线为 VoHive 提交 `f240894e763cf7f0cf74c88562bb9b55f0d573b1`；保留原侧栏、首页统计与设备卡片、设备列表和详情页、短信三栏、代理页、登录和系统设置布局，并保留升级后的后台与数据。新增通话、自动任务、指令与余额入口，以及 Telegram/QQ 编辑器、只读环境诊断。个人版新增：无周期确认弹窗的控制台启动、关闭网页自卸载、独立 SQLite 副本迁移、登录保护的环境诊断、AMR/AMR-WB/MP3 库检测，以及避免官方镜像覆盖个人定制版的更新提示。短信、eSIM、电话、代理、通知、任务等主功能来自完整 HiDeck 源码。按用户选择，设置页只展示 Telegram 和 QQ Bot 提醒；其余提醒关闭，相关源码仍保留以便后续维护。

本说明不替代或变更任何原有许可。分发源码时应一并保留上游许可与作者声明；本次交付只用于用户的个人环境。

0.1.2：个人版设备配额设为无限制（API device_limit=0），统一解除添加、启动、重扫、重连中的设备数量检查；前端按原有逻辑隐藏配额标签。重复设备和硬件有效性检查继续保留。

0.1.3：参考 VoCat 的 eSIM 兼容修复（3cbb6c1）和 USSI 多轮会话修复（a1273f0），按个人版现有通信接口适配：ICCID 启用/禁用失败时，针对指定兼容错误查询目标 Profile 的 ISD-P AID 并重试，保留刷新和 CAT Busy 处理；USSD 后续输入与取消按原会话渠道执行，IMS 断开后不误发蜂窝 AT 指令。现有 vowifi-go 已支持网络 Session ID、SIP INFO 续接和 BYE 取消，保留该实现。

本地 2026-10-02 的未插入设备显示已迁入 TypeScript 源码：只读查询硬件发现，确认未接入才显示灰色“未插入设备”，保留初始化、重启缓冲及自动接管。智能 HTTPS、中文 LuCI 页面和网络部署脚本同步到仓库，个人配置与数据库不放入源码目录。

0.1.4：以现有个人版完整源码为基线，选择性同步 HiDeck 的 AIS EAP-AKA 挑战响应、CHILD_SA 重协商/删除 SPI 方向、重连时的快速重认证身份清理、SWu 失败与重协商日志，以及 Lebara UK / Spark NZ 的密钥更新和 DEVICE_IDENTITY 兼容配置。保留现有 eSIM AID 重试、USSD 会话渠道、设备无限制、QQ/TG、经典中文 UI 和未插入设备显示。没有引入整套模组直拨、重做界面或替换数据库。

0.1.5：选择性纳入 HiDeck v2.1.30 的短信物理 IMEI 身份匹配和 QMI LTE Band 枚举转换修复。同名设备更换模组后不再混用旧模组 SIM 身份，保留换卡过渡与旧短信；LTE 原始枚举转换为真实频段号并修正 FDD/TDD，未知枚举显示未识别。保留已有个人定制，不引入 VoCat AT Session 或按设备通知过滤。

0.1.6：定点移植 Worker 实例/代际与 VoWiFi 生命周期隔离，取消旧 SIM/恢复定时器，拦截旧 QMI/MBIM/AT 回调及迟到的 IP 查询；按现有个人版接口适配，不引入 MBN/外发额度等新框架。同步 NAT-T ESP/IKE 分流、IKE 响应认证/存活计时及旧 SA 清理、IPv6 CP 空值请求、认证后设备身份响应、IMS 短信 RP 回执连接回退、REGISTER 自身 Contact/有效期和 IPv6 地址匹配；增加 Globe PH ePDG 预设。保留现有 DEVICE_IDENTITY 默认策略、通知和经典界面，不启用通知过滤或自动 MBN 写入。

0.1.7：定点适配 HiDeck v2.1.32 后主分支的超长 REGISTER TCP/UDP 有界回退、SIP 拒绝保护、Globe PH 安全端口冲突恢复、DITO 零时区短信时间修正、单条短信持久化失败隔离及浏览器 RTP 换流/序号连续性修复。新增日志使用中文；保留经典界面、无限设备、QQ/TG 和原运行配置。未整体合并自动应答、WebDAV、切卡事务或 VoCat 的独立 AT/数据库架构。
