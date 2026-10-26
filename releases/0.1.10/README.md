# VoHive Plus 0.1.10：SWu/IKE 诊断日志修复

基于完整 0.1.9 个人版，选择性适配 BFNOC 的 HiDeck 55c46b7 和 f619660。SWu/IKE 引擎日志进入既有 app.log 与界面 SSE，遵守主程序日志等级并保留正确调用位置；IKE 重协商响应被拒绝时附加载荷类型、算法编号和提议概况，不记录密钥材料。

格式化依赖仅提取 6e4f7b1 的 create_child_sa_offer.go，不引入该提交的 CHILD_SA 协商或生命周期变更。完整提交和暂缓项见 selected-upstream.json。保留许可和作者声明、经典 Vue UI、未插入设备、无限设备、QQ/TG、eSIM AID fallback、USSD、按 SIM 漫游白名单和实时身份保护。未改运营商策略、MCP、短信确认路径、网络或用户配置。

## 构建与测试

- 本地源码提交：336ea9628e8a935ebbe049e06f20b94147445249；独立 codex/scheduled-010-engine-diagnostics 分支。
- 构建时间：2026-10-26T02:04:36Z，Go 1.27.1，linux/amd64，CGO_ENABLED=0。
- 程序 SHA-256：c41b4707d71e7f08a29168210e63ca849a963e627d99399302471fbf8bfd4677。
- 压缩包 SHA-256：ae18b7d2ce34d7ee1137f8704b8bf977123d5b97945ecb6e5ee83c0464a4312d，27,788,909 字节。
- 主程序日志集成及 backend/api/device/db/notify/esim/vowifihost/phone/volte/qmi/personal 全包通过；独立 engine/logger 与 engine/swu 全包通过。
- logger/swu、主程序日志/phone/volte 和漫游/注册相关 race 各两轮通过；相关 vet 无诊断。实际命令和原始日志同目录保存。
- 旧 0.1.9 加入两项新诊断断言后真实失败，候选通过。格式化测试检查提议、空载荷和不泄露 nonce；集成测试检查临时文件、SSE、调用位置、等级过滤。没有删除或跳过现有断言。
- 前端源码 288 文件及嵌入产物 112 文件逐项哈希不变，复用旧产物；没有重跑前端测试、类型检查或构建。
- 以干净 0.1.9 镜像加候选程序构建 vohive-plus-personal:0.1.10；无设备、无网络 doctor 确认 MP3/AMR/AMR-WB 和 QMI 工具可用。不将生产挂载数据放入镜像。

## 部署与实机检查

仅通过 maintenance/deploy-tested-upgrade.sh 备份并部署，收据 status=deployed，版本/哈希符合实际程序。旧程序、主配置、私有漫游策略和 SQLite 在线一致快照保存在设备私有恢复目录；数据库完整性检查通过，没有用旧数据库覆盖新短信。

部署前后四台设备 Worker、healthy、control_online、physical_present 均正常，注册依次为 5/5/1/5；短信联系人 14/11/0/1 保持。设备上限 0、活动通话 0；主配置、私有漫游策略和维护入口哈希保持。通过本地 HTTP、Tailscale HTTPS 和公网 VPS HTTPS 的首页与 /ping，证书验证开启；这些是路由器发起的检查，不代表异地手机验收。

本轮未发送真实短信、QQ/TG/Bot 消息、拨号或进行录音、实卡重协商验收；未切卡、复位、改变射频或用户选网。单元/模拟测试、只读实机验证与真实业务验收分别记录。

## 安装与恢复

完整可构建源码位于 source/vohive-plus，构建脚本 build-personal.sh；运行时 Dockerfile 位于 source/deployment。Release 提供程序压缩包、build-info.json、selected-upstream.json 和 SHA256SUMS。iStoreOS 需要含语音库的 Debian 容器，不能直接用 musl 运行程序。保留现有挂载，避免同时启动两个控制同一模组的实例；回滚只恢复程序，保留升级后的短信数据库。
