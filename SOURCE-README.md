# 完整源码与个人版恢复

当前源码为 **VoHive Plus 0.1.3-personal-classic（2026-10-03）**，与本次安装到路由器的程序对应。此次将原来只保存在本地的未插入设备显示和智能 HTTPS 改动纳入源码，同时修复 eSIM 兼容重试和 USSD 会话渠道判断。

| 路径 | 内容 |
| --- | --- |
| `source/vohive-plus/` | 完整 Go 后端、Vue 前端、锁文件、第三方源码与许可、个人版构建脚本 |
| `source/router-ui/` | 当前 LuCI 美化、中文日志和智能 HTTPS 入口 |
| `source/https/` | 公网反代、局域网/Tailscale DNS、HAProxy、隧道及证书同步脚本；不包含实际证书、私钥或隧道口令 |
| `source/deployment/` | 0.1.3 Dockerfile、iStoreOS 启动器和 procd 服务 |
| `source/router-network/` | 奇游/PassWall 兼容及 Tailscale 对端路由保护；2026-10-04 补齐出口客户端规则 |
| `source-manifest.json` | 源码文件大小、SHA-256 及本次程序信息 |
| `releases/0.1.3/` | 更新说明、构建与验证记录 |

后端基线保留 HiDeck `3fd3d6aaf55924c338c2cf3f092e1af1f588d152`，前端布局基线保留 VoHive `f240894e763cf7f0cf74c88562bb9b55f0d573b1`。原有许可和作者声明保留；本次没有整体同步 VoCat，也没有更换通信核心。

运行程序 SHA-256：`9a4df904a903a68fe0e54816c52712e29dc7abd2dc63d920cd023ce9fc530631`。现存容器沿用原挂载目录和容器文件层，程序已替换为 0.1.3；本地另外构建 `vohive-plus-personal:0.1.3` 镜像供后续重建使用。原 2026-10-01 的恢复附件仍是历史快照。

## 构建

Linux 环境使用 Go 1.27.1、Node.js 24、pnpm 11.25.0：

```sh
cd source/vohive-plus
sh build-personal.sh
cd ../deployment
cp ../vohive-plus/personal-dist/vohive-plus_linux_amd64 .
docker build -f Personal.Dockerfile -t vohive-plus-personal:0.1.3 .
```

数据库名称沿用 `vohive-plus-final.db`，容器需要原配置、数据和日志挂载以及设备访问权限。不要同时启动两个管理同一模组的实例。升级前使用 SQLite 在线备份或停止服务后备份数据。若使用 Release 中的编译产物，先 `gzip -d vohive-plus_0.1.3_linux_amd64.gz` 并核对 SHA-256，再按 Dockerfile 构建。

已有安装可停止 `/etc/init.d/vohive`，将新程序复制到 `vohive-plus:/usr/local/bin/vohive-plus` 后启动服务；原程序、配置及数据库快照应先保存。仅替换容器程序时，重建容器需选择新镜像。

## 本地改动与测试

未插入设备判定已迁入 `web/src/utils/devicePresence.ts`；每 15 秒只读查询硬件发现，保留初始化、近期在线和重启缓冲，检测失败/过期时回到原有状态。后端重连扫描保留。

LuCI 可用 `sh source/router-ui/install.sh` 安装；中文日志验证运行 `lua source/router-ui/check-logs.lua`。HTTPS 和网络脚本部署前核对实际接口、域名和证书路径；参见各目录说明。

本次验证记录见 [0.1.3 更新说明](releases/0.1.3/README.md)。实际 SIM 切换、运营商 USSD 及拨号未执行。
