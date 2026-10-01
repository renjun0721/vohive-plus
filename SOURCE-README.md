# 完整源码与个人版恢复

本仓库已补齐与实际运行程序对应的 **VoHive Plus 0.1.2-personal-classic** 完整源码，以及后续 LuCI 美化和中文日志修改。

## 文件位置

| 路径 | 内容 |
| --- | --- |
| `source/vohive-plus/` | 完整 Go 后端、原版布局的 Vue 前端、锁文件、第三方源码和许可、个人版构建脚本 |
| `source/router-ui/` | 路由器上实际安装的 LuCI CSS、JavaScript、Lua、安装脚本和界面预览 |
| `source/deployment/` | 0.1.2 Dockerfile、iStoreOS 启动器和 procd 服务 |
| `source/router-network/` | 奇游与 PassWall 自动兼容、Tailscale 与 PassWall 标记保护脚本及启动服务 |
| `source/VoHivePlus-unlimited-devices.patch` | 0.1.1 至 0.1.2 解除设备限额的补丁 |
| `source-manifest.json` | 每个源码文件的大小和 SHA-256，及程序对应信息 |

源码提交为 `f1c24573bdc0326cbf8800ba5feab544838e60d7`。Go 后端基线为 HiDeck `3fd3d6aaf55924c338c2cf3f092e1af1f588d152`；Vue 前端原版布局基线为 VoHive `f240894e763cf7f0cf74c88562bb9b55f0d573b1`。原有 LICENSE、作者声明和第三方许可均保留。VoCat 的诊断思路用于设计，并未合并 VoCat 源码。

实际运行程序 SHA-256 为 `1947be2cb394cdedb5c2c91ca963f397ed0d60238a5ece02b26328730d1bd137`，与源码构建交付记录一致。容器的基础镜像标签仍是 0.1.1，但其中运行的程序已更新为 0.1.2。LuCI 美化文件属于独立安装的覆盖层，不是另一次 Go 后端构建。

## 从源码构建

Linux 构建环境需 Go 1.27.1、Node.js 24、pnpm 11.25.0。按锁文件安装依赖：

```sh
cd source/vohive-plus
sh build-personal.sh
```

产物位于 `personal-dist/vohive-plus_linux_amd64`。程序依赖容器中的语音编解码运行库，应按 Dockerfile 构建运行环境，不要直接替换为 iStoreOS 原生程序。

```sh
cd ../deployment
cp ../vohive-plus/personal-dist/vohive-plus_linux_amd64 .
docker build -f Personal.Dockerfile -t vohive-plus-personal:0.1.2 .
```

没有容器和现存数据的新环境，可先按仓库首页的加密备份恢复步骤恢复。已有安装应先备份配置和数据库，确认原容器挂载目录后再计划升级；不要覆盖新产生的短信或同时启动两个管理同一模组的实例。当前数据库名称为 `vohive-plus-final.db`，创建容器时需要沿用对应参数。

## LuCI 与网络兼容文件

LuCI 页面可执行 `sh source/router-ui/install.sh` 安装，并用 `lua source/router-ui/check-logs.lua` 检查中文日志转换。它保留现有应用配置及数据。

网络脚本按 `source/router-network/` 下的绝对路径布局安装，赋予脚本和服务可执行权限，再启用对应服务。奇游兼容脚本自动读取正在加速的设备规则，只在加速期间让这些设备绕过 PassWall；其他设备继续使用 PassWall。不要把它当作通用固件安装脚本：迁移前核对接口名称 `br-lan`、奇游 `HOST_HOOK*` 规则和 PassWall nftables 表是否相同。

## 核验与备份范围

本次新增源码没有短信数据库、个人配置、Bot 凭据、路由器或 VPS SSH 私钥，也没有再次加入解密密钥。已检查实际 Telegram Token 和 QQ App Secret 未出现在新增源码中。

原先 Release 的加密包保持原样，里面包含运行环境和私人数据。源码现在位于仓库文件树中；不要把原加密包误认为已包含这次新加入的源码。

可用以下命令核对源码文件：

```sh
python3 - <<'PY'
import json, hashlib
from pathlib import Path
manifest = json.loads(Path('source-manifest.json').read_text())
for entry in manifest['files']:
    data = Path(entry['path']).read_bytes()
    assert len(data) == entry['size'], entry['path']
    assert hashlib.sha256(data).hexdigest() == entry['sha256'], entry['path']
print('全部源码文件校验通过')
PY
```
