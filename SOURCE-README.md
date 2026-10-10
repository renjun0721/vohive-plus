# 完整源码与个人版恢复

当前源码为 **VoHive Plus 0.1.4-personal-classic（2026-10-10）**，与本次安装到路由器的程序对应。此次从 0.1.3 个人版完整源码选择性同步 8 项 HiDeck VoWiFi 修复，保留原有 eSIM/USSD、设备无限制、QQ/TG、经典 UI、LuCI 和网络定制。

| 路径 | 内容 |
| --- | --- |
| `source/vohive-plus/` | 完整 Go 后端、Vue 前端、锁文件、第三方源码与许可、个人版构建脚本 |
| `source/router-ui/` | 当前 LuCI 美化、中文日志、本地管理 / HTTPS 通话入口和服务状态开关与重启 |
| `source/https/` | 公网反代、局域网/Tailscale DNS、HAProxy、隧道及证书同步脚本；不包含实际证书、私钥或隧道口令 |
| `source/deployment/` | 0.1.4 Dockerfile、iStoreOS 启动器和 procd 服务 |
| `source/router-network/` | 奇游/PassWall 兼容及 Tailscale 对端路由保护；2026-10-04 补齐出口客户端规则 |
| `source-manifest.json` | 源码文件大小、SHA-256 及本次程序信息 |
| `releases/0.1.4/` | 更新说明、构建与验证记录 |

后端基线保留 HiDeck `3fd3d6aaf55924c338c2cf3f092e1af1f588d152`，前端布局基线保留 VoHive `f240894e763cf7f0cf74c88562bb9b55f0d573b1`。原有许可和作者声明保留；本次没有整体同步 VoCat，也没有更换通信核心。

运行程序 SHA-256：`5684a750fbe68f0f4b301133b3f77e03d149e630ec07755c79db66390a148fe7`。现存容器沿用原挂载目录和容器文件层，程序已替换为 0.1.4；本地另外构建 `vohive-plus-personal:0.1.4` 镜像供后续重建使用。原 2026-10-01 的恢复附件仍是历史快照。

## 构建

Linux 环境使用 Go 1.27.1、Node.js 24、pnpm 11.25.0：

```sh
cd source/vohive-plus
sh build-personal.sh
cd ../deployment
cp ../vohive-plus/personal-dist/vohive-plus_linux_amd64 .
docker build -f Personal.Dockerfile -t vohive-plus-personal:0.1.4 .
```

数据库名称沿用 `vohive-plus-final.db`，容器需要原配置、数据和日志挂载以及设备访问权限。不要同时启动两个管理同一模组的实例。升级前使用 SQLite 在线备份或停止服务后备份数据。若使用 Release 中的编译产物，先 `gzip -d vohive-plus_0.1.4_linux_amd64.gz` 并核对 SHA-256，再按 Dockerfile 构建。

已有安装可停止 `/etc/init.d/vohive`，将新程序复制到 `vohive-plus:/usr/local/bin/vohive-plus` 后启动服务；原程序、配置及数据库快照应先保存。仅替换容器程序时，重建容器需选择新镜像。

## 本地改动与测试

未插入设备判定已迁入 `web/src/utils/devicePresence.ts`；每 15 秒只读查询硬件发现，保留初始化、近期在线和重启缓冲，检测失败/过期时回到原有状态。后端重连扫描保留。

LuCI 可用 `sh source/router-ui/install.sh` 安装；中文日志验证运行 `lua source/router-ui/check-logs.lua`。HTTPS 和网络脚本部署前核对实际接口、域名和证书路径；参见各目录说明。

本次验证记录见 [0.1.4 更新说明](releases/0.1.4/README.md)。实际 SIM 切换、运营商 USSD 及拨号未执行。

## 2026-10-05 LuCI 管理入口与电源操作

“本地管理”按当前 iStoreOS 访问地址打开 HTTP 7575，支持局域网、Tailscale 与 IPv6；“通话中心”继续使用受信任 HTTPS。服务状态卡片右侧仅保留开关和重启两个按钮，运行时开关呈绿色、停止时呈灰色；复用现有 rc.init 接口，显示执行进度并避免重复点击，遵循只读权限和未知/重启状态保护。独立服务管理栏已移除，手动刷新入口放到日志标题右侧。开关与重启直接绑定点击事件，由页面管理禁用状态，避免 LuCI 点击包装器预先禁用按钮导致开关提前返回。浏览器测试使用当前安装的 LuCI 点击处理器和模拟 RPC；实际 rc.init 停止、启动验证通过，服务已恢复运行。

## 0.1.4 选择性上游修复

8 项提交和完整 SHA 见 `releases/0.1.4/selected-upstream.json`。本版本更新现有 VoWiFi 核心，不引入完整模组直拨框架，不整体同步 VoCat；新增运营商策略按现有配置生效。测试与环境限制见该版本说明。
