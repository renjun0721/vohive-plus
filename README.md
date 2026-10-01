# VoHive Plus 私有备份

本仓库保存 iStoreOS 上已验证的 VoHive Plus 部署文件、LuCI 页面及恢复说明。完整运行镜像、应用配置、短信数据库、TLS 证书和日志放在 Release 的加密附件中。

## 本次备份

- 备份日期：2026-10-01，Asia/Shanghai。
- 平台：iStoreOS / Linux amd64。
- 运行镜像：`vohive-plus-personal:0.1.1`。
- 镜像 ID：`sha256:8c88341bf1742723baacb677d4a1e02c90358d146e0f6373c79c3ac88c1fc799`。
- 程序 SHA-256：`1947be2cb394cdedb5c2c91ca963f397ed0d60238a5ece02b26328730d1bd137`。
- 数据库：`vohive-plus-final.db`，通过 SQLite 在线备份接口生成快照，已通过完整性检查。
- 短信接收、QQ 和 Telegram 实际送达均已验证；本次备份未停止正在运行的服务。

已补齐与实际运行的 0.1.2-personal-classic 对应的完整后端、原版布局前端、第三方源码，以及后续 LuCI UI 和中文日志修改。详见 [完整源码与构建恢复说明](SOURCE-README.md)。原 Release 加密备份保持原样。

## 下载和解密

在 Release 下载 `vohive-plus-full-20261001.tar.gz.age` 和 `SHA256SUMS`。按仓库所有者要求，解密密钥也备份在本私有仓库的 `recovery-keys/vohive-plus-20261001.agekey`。下载 Release 附件与该密钥即可恢复。有仓库访问权限的人可以解密完整备份，仓库须保持私有。

```sh
sha256sum -c SHA256SUMS
age --decrypt --identity recovery-keys/vohive-plus-20261001.agekey \
  --output vohive-plus-full-20261001.tar.gz vohive-plus-full-20261001.tar.gz.age
mkdir restore-work
tar -xzf vohive-plus-full-20261001.tar.gz -C restore-work
cd restore-work/snapshot
sha256sum -c SHA256SUMS
```

仍可把密钥另存到电脑或其他安全介质，方便离线恢复。解密后的目录包含 Bot 密钥、短信和证书私钥，继续保留在加密附件中。

## 恢复到 iStoreOS

需要 Docker、LuCI、Lua 5.1，以及 `lyaml`、`luci.jsonc`、`lsqlite3` 和 `nixio`。默认数据目录为 `/mnt/sata2-4/vohive-plus`，HTTP / HTTPS 端口分别为 7575 / 7576。

先完成下载、解密和校验，在仓库目录执行：

```sh
sh deployment/restore.sh /完整路径/restore-work/snapshot
```

脚本只允许恢复到没有 `vohive-plus` 容器和没有现存应用目录的环境；遇到已安装的环境会退出。迁移现有安装时，先自行备份并停止服务，再处理原容器和目录。

恢复完成后脚本创建容器、安装启动服务与 LuCI 文件，但不会启动应用。核对配置、模组 USB 连接和端口后执行：

```sh
/etc/init.d/vohive enable
/etc/init.d/vohive start
```

然后刷新 LuCI 并用新短信验证 QQ / TG 通知。`snapshot/telegram-dns.txt` 保留本次 Telegram 反代的专用 DNS 条目；只有在域名、VPS IP 和路由器 DNS 设置仍相同的情况下才恢复该条目，不要直接覆盖整份 DHCP 配置。

VPS 的 Nginx 配置不属于这个路由器备份。加密包中的维护记录保存了本次反代修复方法，但没有 VPS SSH 私钥。

## 单独安装 LuCI 页面

```sh
sh ui/install.sh
lua ui/check-logs.lua
```

日志页面使用中文摘要、日志等级筛选、折叠诊断和手机布局。完整加密包中的 `ui/` 也保存同一套页面文件。
