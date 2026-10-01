#!/bin/sh
set -eu
umask 077

snapshot_dir=${1:?请提供解密后的 snapshot 目录}
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
task_app_dir=/mnt/sata2-4/vohive-plus

[ "$(id -u)" = 0 ] || { echo '请以 root 身份恢复。' >&2; exit 1; }
for task_cmd in docker lua sha256sum cp chmod; do
    command -v "$task_cmd" >/dev/null || { echo "缺少命令：$task_cmd" >&2; exit 1; }
done
lua -e 'require("lyaml"); require("luci.jsonc"); require("lsqlite3"); require("nixio")'
[ -d /www/luci-static/resources ] || { echo '请先安装 LuCI。' >&2; exit 1; }
[ ! -e "$task_app_dir" ] || { echo '应用目录已存在，请先完成现有安装的备份和迁移。' >&2; exit 1; }
if docker container inspect vohive-plus >/dev/null 2>&1; then
    echo 'vohive-plus 容器已存在，请先完成现有安装的备份和迁移。' >&2
    exit 1
fi
if [ -e /etc/init.d/vohive ] || [ -e /usr/local/bin/vohive-plus-run ]; then
    echo '已有 VoHive 启动文件，请先完成现有安装的备份和迁移。' >&2
    exit 1
fi
snapshot_dir=$(CDPATH= cd -- "$snapshot_dir" && pwd)
[ -f "$snapshot_dir/runtime-image.tar" ]
[ -f "$snapshot_dir/app/config/config.yaml" ]
[ -f "$snapshot_dir/app/data/vohive-plus-final.db" ]
(cd "$snapshot_dir" && sha256sum -c SHA256SUMS)

docker load -i "$snapshot_dir/runtime-image.tar"
task_image_id=$(docker image inspect vohive-plus-personal:0.1.1 --format '{{.Id}}')
[ "$task_image_id" = 'sha256:8c88341bf1742723baacb677d4a1e02c90358d146e0f6373c79c3ac88c1fc799' ] || {
    echo '镜像 ID 与本次备份不一致。' >&2; exit 1;
}
mkdir -p "$task_app_dir"
cp -rp "$snapshot_dir/app/." "$task_app_dir/"
chmod 600 "$task_app_dir/config/config.yaml" "$task_app_dir/data/vohive-plus-final.db"

docker create --name vohive-plus --network host --privileged \
    -v /dev:/dev \
    -v "$task_app_dir/config:/app/config" \
    -v "$task_app_dir/data:/app/data" \
    -v "$task_app_dir/logs:/app/logs" \
    vohive-plus-personal:0.1.1 \
    -c /app/config/config.yaml --database /app/data/vohive-plus-final.db

mkdir -p /etc/init.d /usr/local/bin /usr/lib/lua/luci/controller /usr/share/rpcd/acl.d
cp "$snapshot_dir/system/etc/init.d/vohive" /etc/init.d/vohive
cp "$snapshot_dir/system/usr/local/bin/vohive-plus-run" /usr/local/bin/vohive-plus-run
cp "$snapshot_dir/system/usr/lib/lua/luci/controller/vohive.lua" /usr/lib/lua/luci/controller/vohive.lua
cp "$snapshot_dir/system/usr/share/rpcd/acl.d/luci-app-vohive.json" /usr/share/rpcd/acl.d/luci-app-vohive.json
chmod 755 /etc/init.d/vohive /usr/local/bin/vohive-plus-run
chmod 644 /usr/lib/lua/luci/controller/vohive.lua /usr/share/rpcd/acl.d/luci-app-vohive.json
sh "$snapshot_dir/ui/install.sh"
echo '恢复文件与容器已准备完成。核对配置和 USB 连接后，执行 /etc/init.d/vohive enable 和 /etc/init.d/vohive start。'
