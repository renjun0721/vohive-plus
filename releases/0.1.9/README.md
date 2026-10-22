# VoHive Plus 0.1.9：VoLTE 录音收尾

基于 0.1.8，仅适配 HiDeck `5a855d8e11f7129f92abd40a494c5dd6868bdbcf`（原作者 miaoguoqiang）。接通后的 VoLTE 通话在 QMI 结束、会话消失兜底和本地主动挂断路径补发 `CallFinalized`，触发已有 phone 层混音 WAV 转 MP3 与通话记录更新。未接通通话不补发，重复终态不重复收尾。

保留经典 Vue UI、无限设备、未插入显示、LuCI、QQ/TG、EC20/EC25 身份拓扑、eSIM AID fallback、USSD、VoWiFi 和 iStoreOS Docker/procd；保留 0.1.8 的按 SIM 漫游白名单、实时 ICCID/Worker 校验、换卡隔离、有界切换与自动驻网抑制。私有选网配置不进入源码或安装包。

## 验证

- 先仅向旧 0.1.8 添加上游测试，成功复现接通结束与本地挂断缺少收尾事件；失败日志单独保存为 recording-baseline-repro.log，不列为通过的测试证明。
- 候选 volte/phone 针对性测试通过；新增来电/去电、接通/未接通、QMI 结束/会话消失和重复报告的八种组合回归。
- backend/api/device/db/notify/esim/vowifihost/phone/volte/qmi/personal 全包通过。volte/phone 两轮完整 race、漫游/注册相关两轮 race 通过；相关 vet 无诊断。实际命令见 test-commands.txt。
- 经典前端源码及嵌入产物保持原样，本轮未重跑前端测试、类型检查或构建。
- 新镜像以不含生产挂载数据的 0.1.8 个人镜像与候选程序构建。隔离、无设备、无网络的 doctor 检查确认 MP3/AMR/AMR-WB、qmicli 与 qmi-proxy 可用；该检查不是实机模组验收。

所有协议/通话回归均使用模拟模组，未主动发短信、Bot 消息、拨号、切卡、重置模组或修改网络/用户配置。真实电话录音、真实短信与 QQ/TG 新消息送达本轮未验收，不沿用上一轮用户验收作为本轮结果。

## 本轮上游评估

实际比较 HiDeck `d5f64cfd...8ca988b1`（10 提交，含已评估内容）和 VoCat `162f7b67...06e94ed9`（5 提交）的源码；本地与公开 0.1.8 的 voice.go 和 roaming_operators.go 哈希一致。没有可访问的实时原聊天任务报告，直接读取 GitHub API/raw/patch。

Spark NZ 重协商、VOICE 订阅/射频恢复、按 SIM 本机号码、历史删除、VoCat SIM 上电/换卡恢复和运营商 Contact 策略暂缓，具体依赖和风险见 selected-upstream.json。仅选择可独立复现、无需改变卡/射频策略的录音事件修复。

## 构建与来源

版本 `0.1.9-personal-classic`；构建时间 `2026-10-22T02:02:12Z`；候选代码提交 `987c826`，分支 `codex/scheduled-019-volte-finalize`。Go 1.27.1，linux/amd64，CGO_ENABLED=0，with_utls/nomsgpack。完整构建脚本及锁文件随源码保留，第三方许可证和作者声明保留。

程序 SHA-256 `f79b0001d74cc5df37862e63810683510747e91a10834e495631baecc3b8310a`。
压缩包 SHA-256 `d4f413602431c224b114b5318ae7c0bf684929e55722fe9effe9d9327e08c785`。

## 部署结果

已由既有 deploy-tested-upgrade.sh 完成备份、空闲通话复查、切换和健康校验，部署收据状态 deployed。升级前后的三台健康 Worker、注册状态与短信联系人数量保持；黑色 EC20 原有离线状态仍存在。主 config.yaml 与私有 roaming-operators.json 哈希未变。本地 HTTP、Tailscale HTTPS、公网 VPS HTTPS 的 ping 和管理首页均通过。未执行真实短信、Bot 或电话录音验收。

干净镜像 `vohive-plus-personal:0.1.9`，ID 前缀 `e16accbe1fb4`。生产容器保留原挂载和配置，仅通过受控入口替换程序。旧程序、主/私有配置与 SQLite 在线一致备份保存在本机私有恢复目录；不上传运行数据。
