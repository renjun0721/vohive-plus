# VoHive Plus 0.1.19-personal-classic

基于完整 0.1.18 个人版，仅适配 HiDeck 7b1c7413a77834c4bedc03d63dedd05920037b2e（ilazycat，2026-11-20）：先识别 WAP SI/SL 和 MMS 通知，再使用共用端口 2948 的 OMA 回退；已识别 content-type 优先于冲突的内容特征。修复彩信通知等被误标为“OMA CP 运营商配置短信/可能加密”的问题。保留原始数据、端口、OMA 配置短信、SIM OTA 行为、作者及许可证。

旧版在 tests-only worktree 复现端口分类、显式类型冲突及完整 8-bit SMS-DELIVER 解析失败。候选外层及内嵌 VoWiFi 完整 smscodec 各三轮 race、后端/API/device/db/notify/eSIM/vowifihost/phone/VoLTE/QMI、引擎、关键 IMS/host/漫游恢复 race 和 vet 通过。首次后端准备因 Git 忽略的内嵌 dist 不存在而失败，复制当前相同产物后完整重跑通过；保留失败日志，没有删除断言或跳过失败。

同一修复也应用于个人版中原本字节相同的 VoWiFi 内嵌分类器，两个解码入口均有完整 wire 回归。本轮不改变短信发送、收件箱/数据库、SIM/射频/运营商、网络、用户配置或前端。经典 Vue 源码及 112 个内嵌资产保持当前内容；未运行前端测试/类型检查/构建，未运行完整 imscore 套件。HiDeck 新 ML307/AT 收件箱/发送与选卡/语音/SIM 生命周期系列涉及个人版不存在的模型和跨入口行为，继续单独暂缓。VoCat 源码未变；此前 PC/SC、Bark 加密、IMS-SA、运营商及 RF 等暂缓项保留。

上述是本地模拟、协议和单元测试。实机仅只读健康验证，未执行真实短信、通话、QQ/TG/Bark 送达验收。生产只经现有 deploy-tested-upgrade.sh 校验测试证明、备份并受控部署；部署收据和旧程序、配置及 SQLite 在线一致快照保存在私有维护目录。程序需在含语音库的容器运行，不支持 iStoreOS musl 原生执行。
