# 完整源码与个人版恢复

当前源码为 **VoHive Plus 0.1.20-personal-classic（2026-11-28）**。基于完整 0.1.19，仅为活跃经典短信页面加入换行保留和长词折行；后端、第三方运行代码、按 SIM 漫游白名单、实时身份检查、无限设备、QQ/TG、eSIM/USSD、LuCI 和网络定制保留。实际私有配置不在公开源码中，历史恢复附件保持原样。

`source/vohive-plus/` 为完整 Go/Vue/第三方源码、锁文件和构建脚本；`source/router-ui/` 为 LuCI 页面；`source/https/`、`source/router-network/`、`source/deployment/` 为现有非敏感部署资料。`source-manifest.json` 列出全部当前源码大小、SHA-256 和代码提交，`releases/0.1.20/` 保存构建与验证说明。

## 构建

Linux 使用 Go 1.27.1、Node 24、pnpm 11.25.0：

```sh
cd source/vohive-plus
sh build-personal.sh
cd ../deployment
cp ../vohive-plus/personal-dist/vohive-plus_linux_amd64 .
docker build -f Personal.Dockerfile -t vohive-plus-personal:0.1.20 .
```

程序需在含原有语音库的容器运行；数据库沿用 vohive-plus-final.db，使用现有配置、数据、日志和设备挂载，不要同时运行两个模组管理实例。升级前使用 SQLite 在线一致备份。当前程序 SHA-256：0088dc43e25a5247893dd2eb21cb49212ded9319a7b69e062760f8175392fab6。原容器和挂载保留；新 vohive-plus-personal:0.1.20 镜像供以后重建使用。

## 本轮验证

269 项现有前端测试、类型检查、Vite 构建、后端及引擎、关键 race 和 vet 通过。离线 Chromium 对照真实活跃气泡类和产物 CSS，旧版复现、候选四组通过。首次候选 phone AMR 用例及旧版同机 PCMU 对照均出现既有三秒 WebRTC ICE 超时，候选完整后端原命令复测通过；原日志保留且未修改断言。完整 imscore 未运行。

实机仅只读健康验证；本轮未执行真实短信、通话、USSD 或通知送达验收。选取 VoCat 3db888b7ad381274b816c4b77b48c6a434a99992（Tardisyuan），保留原作者与许可，详见 source/VoCat-SMS-NOTICE.md 及 releases/0.1.20/README.md。通信栈和硬件暂缓项保持。

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

## 0.1.14 每会话公网域名刷新

新媒体会话使用当前公网 DNS 地址，既有会话保留地址；DNS 失败不回退旧 IP，继承调用者取消/截止时间。旧版两项 Create 回归复现，候选 DNS 回归 10 轮 race、相关后端/引擎/关键 race/vet 通过。前端原样复用；完整 IMS 既知限制未解决且未重跑。只读实机健康验证不代表真实短信/通话/通知验收。详见 [0.1.14 更新说明](releases/0.1.14/README.md)。

## 0.1.15 QMI 冷却保护与 IMS 订阅退避

四项旧版回归复现，候选回归、相关后端/引擎和关键 race/vet 通过。五轮重复媒体套件在旧版和候选均复现原有三秒 ICE 超时，未修改或跳过断言；35 个媒体源码/测试文件完全相同。经典前端原样复用。仅只读实机验证，未执行真实短信、通话或通知验收。详见 [0.1.15 更新说明](releases/0.1.15/README.md)。

## 0.1.16 并发订阅确认隔离

旧版本四类回归复现，候选相关后端、引擎、关键 race 与 vet 全部通过。前端和内嵌资产逐字节不变。沿用个人版合并注册/退避存储，未引入按卡 MWI 重检/设置或 SA/运营商变更；完整 runtime 停止再启动的退避延续仍待独立适配。实机只读健康验证，未执行真实短信、通话或通知验收。详见 [0.1.16 更新说明](releases/0.1.16/README.md)。

## 0.1.17 IMS TCP 对端地址快照

IPv4/IPv6 拨出、接受和接受前复位回归在旧版复现，候选 netstack 全包三轮 race、相关后端、关键 IMS/host/device race 和 vet 通过。前端和内嵌资产不变。实机仅只读验证，未执行真实短信、通话或通知验收。旧 SA 生命周期及既有运营商/射频策略变更继续暂缓。详见 [0.1.17 更新说明](releases/0.1.17/README.md)。

## 0.1.18 跨短信中心长短信与串口缓存

跨短信中心合并及节点复用缓存回归在旧版复现；候选短信分片三轮 race、modem、后端、关键 IMS/host/device race 和 vet 通过。个人版 DeviceID/IMSI 存储模型及旧未完成分组保留，没有 ICCID schema 迁移。前端和内嵌资产不变。实机仅只读验证，未执行真实短信、通话或通知验收。旧 SA 生命周期及既有运营商/射频策略变更继续暂缓。详见 [0.1.18 更新说明](releases/0.1.18/README.md)。

## 0.1.19 WAP 短信分类

识别 SI/SL/彩信内容后再使用 2948 共用端口的 OMA 回退；已识别类型优先于冲突的内容特征。保留原始数据、OMA 与 SIM OTA 行为。旧版复现分类失败，候选协议回归三轮 race、后端、引擎、关键 IMS/host/device race 和 vet 通过。前端和内嵌资产不变；实机仅只读验证，未执行真实短信、通话或通知验收。详见 [0.1.19 更新说明](releases/0.1.19/README.md)。
