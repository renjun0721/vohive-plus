# 路由器网络兼容脚本

## 2026-10-04：Tailscale 出口与 TG 连接修复

PassWall 的 TPROXY 给手机客户端流量加代理标记，优先进入本机 table 999。原脚本只为 LAN 地址添加优先路由，缺少 Tailscale 对端的保护，导致反向源地址验证将手机客户端判断为本机地址，新连接在 INPUT 前被丢弃。表现为手机 TG 反复 SYN、没有 SYN-ACK。

补充仅针对 PassWall 标记的 IPv4 `100.64.0.0/10` 和 IPv6 `fd7a:115c:a1e0::/48` 规则，在优先级 998 使用 Tailscale 路由表 52。仅在该表存在 tailscale0 路由时添加，重复执行不会重复添加。现有开机服务和每 10 秒兼容监控保留，重建规则后自动恢复。

修复后实际手机到 Telegram IP 的握手和双向应用数据均恢复；YouTube HTTPS、VoHive 私网/公网 HTTPS 与私网 DNS 查询正常。未切换代理节点或更改 TG 内置代理设置。
