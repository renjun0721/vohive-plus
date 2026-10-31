# 完整源码与个人版恢复

当前源码为 **VoHive Plus 0.1.13-personal-classic（2026-10-31）**。此次基于完整 0.1.12 个人版，独立适配 MengMengCode 的 VoCat d1968d1：SOCKS5 relay 域名解析失败脱敏；保留按 SIM 漫游白名单、有界切换、退避、换卡隔离和实时身份校验。原有 eSIM/USSD、无限设备、QQ/TG、经典 UI、LuCI 和网络定制保留。

| 路径 | 内容 |
| --- | --- |
| `source/vohive-plus/` | 完整 Go 后端、Vue 前端、锁文件、第三方源码与许可、个人版构建脚本 |
| `source/router-ui/` | 当前 LuCI 美化、中文日志、本地管理 / HTTPS 通话入口和服务状态开关与重启 |
| `source/https/` | 公网反代、局域网/Tailscale DNS、HAProxy、隧道及证书同步脚本；不包含实际证书、私钥或隧道口令 |
| `source/deployment/` | 0.1.13 Dockerfile、iStoreOS 启动器和 procd 服务 |
| `source/router-network/` | 奇游/PassWall 兼容及 Tailscale 对端路由保护；2026-10-04 补齐出口客户端规则 |
| `source-manifest.json` | 源码文件大小、SHA-256 及本次程序信息 |
| `releases/0.1.13/` | 更新说明、构建与验证记录 |

后端基线保留 HiDeck `3fd3d6aaf55924c338c2cf3f092e1af1f588d152`，前端布局基线保留 VoHive `f240894e763cf7f0cf74c88562bb9b55f0d573b1`。原有许可和作者声明保留；本次没有整体同步 VoCat，也没有更换通信核心。

运行程序 SHA-256：`7fc007719b02d41e4d508a5d6f104388c466567b605d674e69c174bd743b7e24`。现存容器沿用原挂载目录和容器文件层，程序已替换为 0.1.13；本地另外构建 `vohive-plus-personal:0.1.13` 镜像供后续重建使用。原 2026-10-01 的恢复附件仍是历史快照。

## 构建

Linux 环境使用 Go 1.27.1、Node.js 24、pnpm 11.25.0：

```sh
cd source/vohive-plus
sh build-personal.sh
cd ../deployment
cp ../vohive-plus/personal-dist/vohive-plus_linux_amd64 .
docker build -f Personal.Dockerfile -t vohive-plus-personal:0.1.13 .
```

数据库名称沿用 `vohive-plus-final.db`，容器需要原配置、数据和日志挂载以及设备访问权限。不要同时启动两个管理同一模组的实例。升级前使用 SQLite 在线备份或停止服务后备份数据。若使用 Release 中的编译产物，先 `gzip -d vohive-plus_0.1.13_linux_amd64.gz` 并核对 SHA-256，再按 Dockerfile 构建。

已有安装可停止 `/etc/init.d/vohive`，将新程序复制到 `vohive-plus:/usr/local/bin/vohive-plus` 后启动服务；原程序、配置及数据库快照应先保存。仅替换容器程序时，重建容器需选择新镜像。

## 本地改动与测试

未插入设备判定已迁入 `web/src/utils/devicePresence.ts`；每 15 秒只读查询硬件发现，保留初始化、近期在线和重启缓冲，检测失败/过期时回到原有状态。后端重连扫描保留。

LuCI 可用 `sh source/router-ui/install.sh` 安装；中文日志验证运行 `lua source/router-ui/check-logs.lua`。HTTPS 和网络脚本部署前核对实际接口、域名和证书路径；参见各目录说明。

本次验证记录见 [0.1.13 更新说明](releases/0.1.13/README.md)。本轮真实短信、Bot 送达、SIM 切换及拨号未执行。

## 2026-10-05 LuCI 管理入口与电源操作

“本地管理”按当前 iStoreOS 访问地址打开 HTTP 7575，支持局域网、Tailscale 与 IPv6；“通话中心”继续使用受信任 HTTPS。服务状态卡片右侧仅保留开关和重启两个按钮，运行时开关呈绿色、停止时呈灰色；复用现有 rc.init 接口，显示执行进度并避免重复点击，遵循只读权限和未知/重启状态保护。独立服务管理栏已移除，手动刷新入口放到日志标题右侧。开关与重启直接绑定点击事件，由页面管理禁用状态，避免 LuCI 点击包装器预先禁用按钮导致开关提前返回。浏览器测试使用当前安装的 LuCI 点击处理器和模拟 RPC；实际 rc.init 停止、启动验证通过，服务已恢复运行。

## 0.1.4 选择性上游修复

8 项提交和完整 SHA 见 `releases/0.1.4/selected-upstream.json`。本版本更新现有 VoWiFi 核心，不引入完整模组直拨框架，不整体同步 VoCat；新增运营商策略按现有配置生效。测试与环境限制见该版本说明。

## 0.1.5 选择性更新

本轮提交见 `releases/0.1.5/selected-upstream.json`。已在路由器使用只读 API 验证三个工作进程、每台模组短信历史可读及真实频段/双工状态。保留已有个人功能和网络设置，未引入 VoCat AT Session 或按设备通知过滤。

## 0.1.6 选择性更新

完整范围、适配和测试限制见 [0.1.6 更新说明](releases/0.1.6/README.md)。四台模组实机只读状态及历史短信已验证；实际运营商 VoWiFi、USB 拔插和 QQ/TG 新短信送达需后续验收。

## 0.1.7 选择性更新

范围和验证见 [0.1.7 更新说明](releases/0.1.7/README.md)。四个 Worker、注册及短信历史、本地与两条 HTTPS 入口已验证。新增日志为中文；真实收发短信、拨号、切卡和长期运行仍需实际使用验证。

## 0.1.8 按卡漫游白名单

配置格式和行为见 [ROAMING-OPERATORS.md](source/vohive-plus/ROAMING-OPERATORS.md)，实现与验证见 [0.1.8 更新说明](releases/0.1.8/README.md)。实际卡标识、roaming-operators.json 和短信数据库不放公开源码；只有虚构 ICCID 的示例。

## 0.1.9 VoLTE 录音收尾修复

接通后的通话在 QMI 终态、会话消失兜底和用户挂断路径补发一次 CallFinalized，触发已有混音录音转码和记录更新。保留未接通通话行为。上游原作者 miaoguoqiang（提交 5a855d8e11f7129f92abd40a494c5dd6868bdbcf）；补充来去电、重复终态及会话消失模拟回归。详见 [0.1.9 更新说明](releases/0.1.9/README.md)。没有进行真实电话录音验收。

## 0.1.10 SWu/IKE 日志修复

引擎复用主日志文件与 SSE 输出，重协商拒绝附加无密钥载荷摘要；保留运营商与 SIM 策略。作者 BFNOC，选择提交及诊断依赖见 [0.1.10 更新说明](releases/0.1.10/README.md)。相关单元、race、vet 与只读设备检查通过；未进行真实短信、通话或运营商重协商验收。

## 0.1.12 登录认证修复

正确认证只归还本次预留名额，同 IP 连续失败锁定十分钟并返回 Retry-After。未纳入未部署的本地 0.1.11。完整相关后端与 API/媒体 race、漫游 race 和 vet 通过，保留原始失败及测试夹具修正证据。前轮完整 IMS 套件 MSS/UDP-stop 限制仍未解决且本轮未重跑；未进行真实短信/通话验收。详见 [0.1.12 更新说明](releases/0.1.12/README.md)。

## 0.1.13 SOCKS5 错误脱敏

在现存两条解析路径去除代理可控域名、DNS 细节和原始错误链，保留安全分类。旧版两路径复现泄露，候选离线安全回归 20 轮 race 通过，相关后端和引擎、关键 race、vet 通过。前端复用原样产物；完整 IMS 套件既知 MSS/UDP-stop 限制未解决且未重跑。仅进行只读实机健康验证，未进行真实短信/通话/通知发送。详见 [0.1.13 更新说明](releases/0.1.13/README.md)。
