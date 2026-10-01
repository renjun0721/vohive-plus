#!/bin/sh
# Bypass PassWall only for LAN clients currently owned by Qiyou's active rules.
mkdir -p /var/lock
exec 8>/var/lock/qiyou-passwall-compat.lock
flock -x 8
tag=qiyou-passwall-compat-v1
macs=$(
    for table in $(nft list tables 2>/dev/null | awk '$2 == "ip" && $3 ~ /^HOST_HOOK[0-9]+$/ {print $3}'); do
        nft list chain ip "$table" ACC_LOAD 2>/dev/null
    done | sed -n 's/.*iifname "br-lan" ether saddr \([0-9a-fA-F:]*\).*/\1/p' | tr 'A-F' 'a-f' | sort -u
)
batch=$(mktemp /tmp/qiyou-passwall-compat.XXXXXX) || exit 1
trap 'rm -f "$batch"' EXIT
changed=0
for chain in mangle_prerouting dstnat; do
    rules=$(nft -a list chain inet passwall "$chain" 2>/dev/null) || continue
    first_rules=$(printf '%s\n' "$rules" | awk '/# handle/ && !/chain / { if ($0 !~ /comment "qiyou-passwall-compat-v1-/) exit; print }')
    present=$(printf '%s\n' "$first_rules" | sed -n 's/.*comment "qiyou-passwall-compat-v1-\([0-9a-f:]*\)".*/\1/p' | sort -u)
    managed=$(printf '%s\n' "$rules" | grep -c 'comment "qiyou-passwall-compat-v1-' || true)
    expected=$(printf '%s\n' "$macs" | awk 'NF {n++} END {print n+0}')
    [ "$present" = "$macs" ] && [ "$managed" = "$expected" ] && continue
    for handle in $(printf '%s\n' "$rules" | sed -n '/comment "qiyou-passwall-compat-v1-/s/.*# handle \([0-9]*\).*/\1/p'); do
        printf 'delete rule inet passwall %s handle %s\n' "$chain" "$handle" >> "$batch"
    done
    for mac in $macs; do
        printf '%s\n' "$mac" | grep -Eq '^([0-9a-f]{2}:){5}[0-9a-f]{2}$' || continue
        printf 'insert rule inet passwall %s iifname "br-lan" ether saddr %s counter return comment "%s-%s"\n' "$chain" "$mac" "$tag" "$mac" >> "$batch"
    done
    changed=1
done
if [ "$changed" = 1 ]; then
    nft -f "$batch" || exit 1
    logger -t qiyou-passwall-compat "Updated active Qiyou client exclusions: ${macs:-none}"
fi
exit 0
