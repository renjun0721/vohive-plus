# VoHive Plus 个人版

本版本由用户授权，为其个人 iStoreOS 环境定制。版本：0.1.3-personal-classic，2026-10-03。

基础源码来自 [HiDeck](https://github.com/yibaiba/hideck)，固定提交 `3fd3d6aaf55924c338c2cf3f092e1af1f588d152`（v2.1.23 之后的源码快照）。HiDeck 继承 VoHive，项目原有作者声明、LICENSE 与 THIRD_PARTY_NOTICES.md 均保留；third_party 中的组件仍适用各自的许可，包括 vowifi-go 的 AGPL-3.0。

[VoCat](https://github.com/MengMengCode/VoCat) 的环境诊断思路用于设计个人版只读检查；本次没有复制 VoCat 的代码，也没有把两个项目的数据库或通信核心直接混合。

个人版以原版 VoHive UI 与布局为基础增加功能。前端基线为 VoHive 提交 `f240894e763cf7f0cf74c88562bb9b55f0d573b1`；保留原侧栏、首页统计与设备卡片、设备列表和详情页、短信三栏、代理页、登录和系统设置布局，并保留升级后的后台与数据。新增通话、自动任务、指令与余额入口，以及 Telegram/QQ 编辑器、只读环境诊断。个人版新增：无周期确认弹窗的控制台启动、关闭网页自卸载、独立 SQLite 副本迁移、登录保护的环境诊断、AMR/AMR-WB/MP3 库检测，以及避免官方镜像覆盖个人定制版的更新提示。短信、eSIM、电话、代理、通知、任务等主功能来自完整 HiDeck 源码。按用户选择，设置页只展示 Telegram 和 QQ Bot 提醒；其余提醒关闭，相关源码仍保留以便后续维护。

本说明不替代或变更任何原有许可。分发源码时应一并保留上游许可与作者声明；本次交付只用于用户的个人环境。

0.1.2：个人版设备配额设为无限制（API device_limit=0），统一解除添加、启动、重扫、重连中的设备数量检查；前端按原有逻辑隐藏配额标签。重复设备和硬件有效性检查继续保留。

0.1.3：参考 VoCat 的 eSIM 兼容修复（3cbb6c1）和 USSI 多轮会话修复（a1273f0），按个人版现有通信接口适配：ICCID 启用/禁用失败时，针对指定兼容错误查询目标 Profile 的 ISD-P AID 并重试，保留刷新和 CAT Busy 处理；USSD 后续输入与取消按原会话渠道执行，IMS 断开后不误发蜂窝 AT 指令。现有 vowifi-go 已支持网络 Session ID、SIP INFO 续接和 BYE 取消，保留该实现。

本地 2026-10-02 的未插入设备显示已迁入 TypeScript 源码：只读查询硬件发现，确认未接入才显示灰色“未插入设备”，保留初始化、重启缓冲及自动接管。智能 HTTPS、中文 LuCI 页面和网络部署脚本同步到仓库，个人配置与数据库不放入源码目录。
