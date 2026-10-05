# VoHive Plus 服务管理页面

面向当前 iStoreOS 的 LuCI 页面：紫蓝色主题、服务状态卡片、操作按钮、可筛选的中文日志、折叠诊断和手机布局。

源文件保存在本目录，对应安装位置如下：

| 源文件 | 安装位置 |
| --- | --- |
| `vohive.js` | `/www/luci-static/resources/view/services/vohive.js` |
| `vohive.css` | `/www/luci-static/resources/view/services/vohive.css` |
| `vohive-luci-info` | `/usr/libexec/vohive-luci-info` |
| `logs.lua` | `/usr/lib/lua/vohive/logs.lua` |

状态读取实际的 `vohive-plus` Docker 容器，诊断使用 `/mnt/sata2-4/vohive-plus` 下的配置和数据目录。页面每 10 秒读取一次状态，可以暂停自动刷新。最近 200 条系统日志按时间倒序显示，默认隐藏服务心跳。

“本地管理”打开当前 iStoreOS 访问地址的 HTTP 7575 端口，适用于局域网 IP、Tailscale IP 和 IPv6 地址。“通话中心”打开 `https://xjp.721609.xyz/#/phone`，使用受信任 HTTPS 申请麦克风权限。

服务状态卡片的电源按钮根据实际状态执行停止或启动，与下方服务管理按钮共用 LuCI `rc.init` 接口。执行期间显示中文进度并禁止重复操作；只读权限、状态未知或容器重启时禁用。

请求动作、日志等级、时间、常见错误和工作进程字段使用中文展示。接口地址、设备编号和源代码位置等技术值保留，源代码位置及请求路径可悬停查看。未识别的消息保留原文，避免改变错误含义；应用写入磁盘的原始日志不变。

在本机重新安装：

```sh
sh /root/vohive/ui/install.sh
```

安装完成后用 Ctrl+F5 刷新 LuCI 页面。安装仅替换管理页面和状态读取文件，不重启 VoHive Plus。现有 RPC 权限无需增加。

改动前的文件保存在 `backups/20261001-before-redesign/`。恢复旧页面：

```sh
cp -p /root/vohive/ui/backups/20261001-before-redesign/vohive.js /www/luci-static/resources/view/services/vohive.js
cp -p /root/vohive/ui/backups/20261001-before-redesign/vohive-luci-info /usr/libexec/vohive-luci-info
```

日志解析检查：在 `/root/vohive` 中执行 `lua ui/check-logs.lua`。`check-browser.cjs` 是临时容器内的浏览器检查入口，使用真实状态快照和模拟 LuCI API，不会执行实际的启动、停止或重启。浏览器检查包括桌面和手机布局、筛选、复制、断线恢复、只读权限和日志文字注入。
