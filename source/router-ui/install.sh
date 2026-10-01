#!/bin/sh
set -eu
task_ui_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
lua - "$task_ui_dir/logs.lua" "$task_ui_dir/vohive-luci-info" <<'LUA'
assert(loadfile(arg[1]))
assert(loadfile(arg[2]))
LUA

install_file() {
    source_path=$1
    target_path=$2
    target_mode=$3
    mkdir -p "$(dirname -- "$target_path")"
    cp "$source_path" "$target_path.new"
    chmod "$target_mode" "$target_path.new"
    mv -f "$target_path.new" "$target_path"
}

install_file "$task_ui_dir/logs.lua" /usr/lib/lua/vohive/logs.lua 644
install_file "$task_ui_dir/vohive-luci-info" /usr/libexec/vohive-luci-info 755
install_file "$task_ui_dir/vohive.css" /www/luci-static/resources/view/services/vohive.css 644
install_file "$task_ui_dir/vohive.js" /www/luci-static/resources/view/services/vohive.js 644
printf 'VoHive Plus 页面已更新，请在浏览器中按 Ctrl+F5 刷新。\n'
