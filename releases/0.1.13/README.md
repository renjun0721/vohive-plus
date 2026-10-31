# VoHive Plus 0.1.13：SOCKS5 解析失败日志脱敏

以已部署的 0.1.12-personal-classic 为基线，选择性适配 MengMengCode 的 VoCat 提交 d1968d1c3fb88258d86a6e2e69f9c3304d160711。SOCKS5 代理可以在返回的 relay 域名中夹带敏感文本；旧版解析失败会把该域名和 DNS 原始错误链写入失败信息。此版在个人版现有的 ASSOCIATE/CONNECT 回复和 UDP 数据报两条路径中移除这些内容，保留取消、超时和 DNS 安全分类。

运行代码只修改现存 socks5_protocol.go 的失败返回及一个纯错误转换函数；不改成功解析、DNS 设置、代理连接生命周期、运营商策略或设备操作。保留按 SIM 漫游白名单、Worker/实时 ICCID、换卡隔离、经典中文 UI、未插入显示、设备无限制、QQ/TG、eSIM AID fallback、USSD、LuCI、iStoreOS Docker/procd 和原网络定制。原作者、LICENSE 和第三方声明保留。

旧 0.1.12 只加兼容接口的离线回归时，两条路径均复现泄露。候选安全/错误分类回归 20 轮 race 通过；相关后端/API/device/db/notify/eSIM/vowifihost/phone/VoLTE/QMI/personal 全包、引擎 ipsec/swu/runtimehost、关键 race 和 vet 的最终检查通过。新 worktree 缺少被忽略的前端嵌入产物，初轮 backend/vet 无法完成；保留原失败日志，复制当前版原样产物后完成最终检查，未删除断言或跳过失败。

前端源码 288 文件、嵌入产物 112 文件逐字节相同，复用产物，未重跑前端测试/类型检查/构建。完整 imscore 套件未重跑；既有 MSS ENOPROTOOPT 与 UDP 停止并发失败仍未解决，不能据此版本声称已经修复。

VoCat 同期的 ZXIC/未知 USB/typed AT profile、端口租用及 APDU/eSIM/ISIM 架构改动暂缓：个人版 Pool/Worker/QMI 及按卡策略模型不同，无目标硬件验收依据。既有 IMS/SA/运营商及其他架构暂缓项保留。

仅通过 maintenance/deploy-tested-upgrade.sh 受控部署并保留部署收据、旧程序、配置和 SQLite 在线一致备份；实机只读健康验证与单元测试分别记录。不进行真实短信、通话、录音、QQ/TG/通知发送、切卡、复位模组、余额任务或运营商验收。

使用 build-personal.sh 从完整源码重建，或按既有含语音库的个人 Dockerfile 安装 Release 程序；不支持直接在 iStoreOS musl 主机运行。构建时间、源码提交及 SHA-256 见 build-info.json 和 SHA256SUMS。
