#!/bin/sh
set -eu
umask 077
config_dir=/etc/vohive-https
stage=$(mktemp "$config_dir/direct.pem.XXXXXX")
trap 'rm "$stage"' EXIT
/usr/bin/dbclient -T -i "$config_dir/tls-sync.key" -o BatchMode=yes \
    -K 15 -I 60 ubuntu@129.150.57.185 export-tls > "$stage"
openssl x509 -in "$stage" -noout -checkend 86400 >/dev/null
openssl x509 -in "$stage" -noout -checkhost xjp.721609.xyz >/dev/null
openssl x509 -in "$stage" -pubkey -noout | openssl pkey -pubin -outform DER > "$stage.cert-public"
openssl pkey -in "$stage" -pubout -outform DER > "$stage.key-public"
if ! cmp -s "$stage.cert-public" "$stage.key-public"; then
    rm "$stage.cert-public" "$stage.key-public"
    logger -t vohive-https '证书更新失败：证书与私钥不匹配'
    exit 1
fi
rm "$stage.cert-public" "$stage.key-public"
if [ -f "$config_dir/direct.pem" ] && cmp -s "$stage" "$config_dir/direct.pem"; then
    exit 0
fi
if [ -f "$config_dir/direct.pem" ]; then
    cp "$config_dir/direct.pem" "$config_dir/direct.pem.previous"
fi
cp "$stage" "$config_dir/direct.pem"
chmod 600 "$config_dir/direct.pem"
if ! /usr/sbin/haproxy -c -f "$config_dir/direct-haproxy.cfg" >/dev/null 2>&1; then
    [ ! -f "$config_dir/direct.pem.previous" ] || cp "$config_dir/direct.pem.previous" "$config_dir/direct.pem"
    logger -t vohive-https '证书更新失败：HTTPS 配置检查未通过'
    exit 1
fi
/etc/init.d/vohive-direct-https restart
logger -t vohive-https 'HTTPS 证书已更新'
