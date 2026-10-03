# VoHive Plus HTTPS：局域网、Tailscale 与公网

访问地址：**https://xjp.721609.xyz/**。使用原 VoHive Plus 网页账号登录，并允许浏览器使用麦克风。

iStoreOS 的 VoHive Plus 服务管理页提供“智能 HTTPS”和“通话中心”入口，均使用这个域名，不再从入口打开 HTTP 页面。

## 按网络选择连接

浏览器始终访问相同的 HTTPS 域名和 443 端口，因此登录状态可以沿用：

- 使用 iStoreOS 的 LAN DNS（192.168.100.1）时，域名解析到路由器的 100.71.142.28，由本地 HTTPS 服务直接访问 VoHive Plus。
- 连接 Tailscale 并使用下述 Split DNS 规则时，同样解析到 100.71.142.28，经 Tailscale TCP 转发访问本地 HTTPS 服务。
- 使用其他网络的公共 DNS 时，域名仍指向新加坡 VPS，沿用公网反代。

路由器自身保留域名的公网解析，主 DNS 配置中的 `dhcp.vohive_tg_vps.ip` 仍为 129.150.57.185。客户端的私网解析由独立 DNS 服务提供，避免更改路由器自身的代理节点和 TG 连接。路由器不接受 Tailscale 的 DNS 设置，以避免自身成为该域名 DNS 上游时产生循环；此偏好保存在 Tailscale 状态文件。

**Tailscale 账号后台需要完成一次设置：**在 https://login.tailscale.com/admin/dns 的 Nameservers 中添加 Custom，服务器填 `100.71.142.28`，启用 Restrict to domain / Split DNS，域名填 `xjp.721609.xyz`。手机需要启用使用 Tailscale DNS。此次会话未持有 Tailscale 账号后台权限；截至部署验证时该规则尚未出现在网络配置中。

手机若缓存了旧解析，重新连接 Wi-Fi 或 Tailscale 后再打开页面。Tailscale 连接故障时，关闭 Tailscale 可恢复公共 DNS 线路。

### 本地服务

- `/etc/init.d/vohive-direct-https`：独立 HAProxy，8443 根据 TLS 域名路由；VoHive 在本地 8444 终止 TLS，再访问 HTTP 7575。其他 TLS 域名继续进入原 LuCI HTTPS 服务。
- `/etc/init.d/vohive-private-dns`：独立 DNS，监听 5354。目标域名返回 100.71.142.28，其他域名转发到原 DNS。
- `/etc/nftables.d/30-vohive-https.nft`：将私网地址的 HTTPS、DNS，以及 br-lan 到 192.168.100.1 的 DNS 请求送入上述服务。没有改变原 LuCI 443 监听设置。
- `tailscale serve --bg --tcp=443 tcp://127.0.0.1:8443`：私网 TCP 转发，设置持久化。
- `/etc/vohive-https/direct.pem`：与 VPS 域名相同的受信任证书及私钥，0600。
- `/etc/vohive-https/tls-sync.key`：受限 SSH 密钥，0600；服务器仅允许执行证书导出命令。
- `/usr/local/bin/vohive-sync-tls`：每日 05:23 同步 VPS 证书，验证域名、有效期和密钥匹配后更新本地服务。
- VPS `/usr/local/sbin/vohive-export-tls`：导出该域名的证书。新增 SSH 授权行标记为 `vohive-plus-tls-sync`，不允许端口转发或任意命令。

所有本地服务已设为开机启动。证书、SSH 私钥和带有密钥的配置需要纳入加密备份。

### 速度验证

2026-10-02 同一 JavaScript 文件的单次实测：公网原文件约 1.16 MB、10 秒；开启 gzip 与 HTTP/2 后约 437 KB、1.75 秒；本地 HTTPS gzip 约 457 KB、0.05 秒。实际手机速度取决于网络。

本地 HTTPS 首页约 0.03 秒，证书验证成功。私网 DNS 正确返回 100.71.142.28，其他域名（github.com）仍可正常转发解析。Tailscale 到当前在线手机的探测为直连、约 12 毫秒。

WebRTC 保留各接口的私网候选地址及 VPS 公网候选地址；客户端可通过 ICE 选择可达的语音路径。当前仍未进行实际拨号或已登录的 WebRTC 会话测试。

### 本次优化的回退

备份位于 `/mnt/sata2-4/vohive-recovery-backups/20261002-before-smart-https/`。停止并禁用 `vohive-direct-https`、`vohive-private-dns`，撤销 Tailscale TCP 443 转发，删除本次 `30-vohive-https.nft` 后检查并 reload 防火墙。撤销证书同步 cron 行，并按备份恢复此前的 Tailscale DNS 接受偏好（此前为 true）。如不再使用，删除本次 TLS 同步公钥的授权行，保留其他授权密钥。

VPS 公网性能配置备份为 `/etc/nginx/sites-available/tgapi.before-smart-https-20261002`。如需回退，恢复该文件，通过 `nginx -t` 后 reload。此前的公网 HTTPS 及语音转发配置仍然可独立使用。

## 连接方式

- VPS Nginx 使用现有受信任的证书，将网页、API、事件流及 WebSocket 转发到 VPS 的 `127.0.0.1:17575`。
- iStoreOS 通过 SSH 反向转发，将该端口连接到本地 `127.0.0.1:7575`。SSH 专用密钥仅允许指定端口转发，不能执行远程命令。
- WebRTC 使用 UDP **61580**。VPS 的 FRP 服务将其转发到路由器的同名 UDP 端口；FRP 控制连接通过 SSH 本地转发进入 VPS 回环地址。
- VoHive Plus 配置中的 `server.webrtc_public_host` 为 VPS 公网地址 `129.150.57.185`，`server.webrtc_udp_address` 为 `:61580`。
- 原 TG `/bot<token>/...` 和 `/file/bot<token>/...` 请求继续转发到 Telegram API。

原 VPS 的 UDP 端口跳跃规则覆盖 `3204–57345`，因此不使用默认语音端口 7580。使用 61580 不需要修改原端口跳跃规则。

## 安装位置

iStoreOS：

- `/etc/init.d/vohive-vps-tunnel`：SSH 通道，已设为开机启动及断线重连。
- `/etc/init.d/vohive-media-relay`：FRP 客户端，已设为开机启动。
- `/etc/vohive-https/ssh-tunnel.key` 和 `/etc/vohive-https/frpc.toml`：包含敏感信息，权限为 0600；备份时应加密。
- `/usr/local/bin/vohive-frpc`：官方 FRP 0.71.0 Linux amd64 二进制。
- `/root/.ssh/known_hosts` 与 `/.ssh/known_hosts`：相同的 VPS 主机公钥。procd 启动环境的主目录为 `/`，Dropbear 会读取后者；客户端保持主机密钥校验。

VPS：

- `/etc/nginx/sites-available/tgapi`、`/etc/nginx/conf.d/vohive-upgrade.conf`：HTTPS 转发。
- `/etc/systemd/system/vohive-media-relay.service`：FRP 服务，已设为开机启动，使用 `vohive-relay` 系统用户。
- `/etc/vohive-https/frps.toml`：包含敏感信息，root/vohive-relay 0640。
- `/usr/local/bin/vohive-frps`：官方 FRP 0.71.0 Linux arm64 二进制。
- `ubuntu` 用户的 `authorized_keys` 新增 `vohive-plus-vps-tunnel` 专用公钥。

服务端 FRP 仅允许代理 UDP 61580，控制端口仅监听 VPS 回环地址。证书沿用原站点及已有 Certbot 定时任务。

## 验证记录

2026-10-02 已确认：

- HTTPS 首页返回 200，TLS 证书验证成功，标题为 VoHive Plus。
- Chromium 无需跳过证书校验即可识别安全上下文，并成功获取模拟麦克风。
- 未登录访问电话设备 API 返回 401。
- HTTPS `/ping` 正常返回 pong。
- 原 TG 反代 `getMe` 返回 200、`ok=true`，没有发送额外通知。
- 外部 UDP 测试包经 VPS 的 UDP 61580 进入路由器回环接口，抵达 VoHive Plus 的 UDP 61580。
- 两端服务已启用开机启动。

尚未验证：当前 VoHive Plus 登录密码与配置文件旧密码不一致，等待用户提供当前凭据，才能测试已登录的 WebRTC ICE/DTLS 连接。没有拨打真实电话；手机流量下的实际双向通话还需验证。设备自身的 WiFi calling/IMS 状态也必须正常。

## 回退

路由器先停止并禁用 `vohive-media-relay` 和 `vohive-vps-tunnel`，然后将 `/mnt/sata2-4/vohive-recovery-backups/20261002-before-https/config.yaml` 恢复到 VoHive Plus 的配置位置，再重启 VoHive 服务。

如需撤销管理页入口，将同一备份目录的 `vohive.js` 恢复到本工作区 `ui/vohive.js` 及 `/www/luci-static/resources/view/services/vohive.js`。

VPS 将 `/etc/nginx/sites-available/tgapi.before-vohive-https-20261002` 恢复到 `tgapi`，删除本次新增的 `vohive-upgrade.conf`，通过 `nginx -t` 后 reload。停止并禁用 `vohive-media-relay`。如撤销 SSH 密钥，仅删除带有本次专用公钥的授权行，保留其他授权密钥。

## 官方来源

- [FRP UDP 转发](https://gofrp.org/en/docs/features/tcp-udp/)
- [FRP 0.71.0](https://github.com/fatedier/frp/releases/tag/v0.71.0)，下载校验匹配 GitHub Release 的 SHA-256：amd64 `84f27e39f11169f7adcef8e8b70c9329de17747b1f14dad9fb95eef5682ea716`；arm64 `f33c293c275d8fc68c654b6fba8f10b2551d6463d09a9fc9cffb7227eae82266`。
- [Nginx WebSocket 转发](https://nginx.org/en/docs/http/websocket.html)
