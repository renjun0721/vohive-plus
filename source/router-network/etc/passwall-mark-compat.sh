#!/bin/sh
# Protect PassWall's full mark from legacy connmark rules in mangle tables.
mark=$(sed -n 's/^FWMARK="\(0x[0-9A-Fa-f]*\)"/\1/p' /usr/share/passwall/nftables.sh | head -n 1)
[ -n "$mark" ] || exit 0
mkdir -p /var/lock
exec 9>/var/lock/passwall-mark-compat.lock
flock -x 9
comment="passwall-mark-compat-v3-$mark"
batch=$(mktemp /tmp/passwall-mark-compat.XXXXXX) || exit 1
trap 'rm -f "$batch"' EXIT
changed=0
for family in ip ip6; do
    for chain in PREROUTING OUTPUT; do
        rules=$(nft -a list chain "$family" mangle "$chain" 2>/dev/null) || continue
        printf '%s\n' "$rules" | grep -Eq '0x0000ff00|0x00ff0000' || continue
        first_two=$(printf '%s\n' "$rules" | awk '/# handle/ && !/chain / {print; n++; if (n == 2) exit}')
        [ "$(printf '%s\n' "$first_two" | grep -c "$comment")" = 2 ] && continue
        for handle in $(printf '%s\n' "$rules" | sed -n '/comment "passwall-mark-compat-/s/.*# handle \([0-9]*\).*/\1/p'); do
            printf 'delete rule %s mangle %s handle %s\n' "$family" "$chain" "$handle" >>"$batch"
        done
        printf 'insert rule %s mangle %s ct mark %s counter return comment "%s"\n' "$family" "$chain" "$mark" "$comment" >>"$batch"
        printf 'insert rule %s mangle %s meta mark %s counter return comment "%s"\n' "$family" "$chain" "$mark" "$comment" >>"$batch"
        changed=1
    done
done
if [ "$changed" = 1 ]; then
    nft -f "$batch" || exit 1
    logger -t passwall-mark-compat 'Restored PassWall mark guards after firewall rule creation or reload'
fi
mark_mask=0xffffff00
mark_signature=$(printf '0x%08x' "$((mark & mark_mask))")
for subnet in $(ip -4 route show table main dev br-lan scope link | awk '$1 ~ /\// { print $1 }'); do
    ip -4 rule show | grep -F "to $subnet " | grep -F "fwmark $mark_signature/$mark_mask " | grep -q 'lookup main' && continue
    ip -4 rule add pref 998 to "$subnet" fwmark "$mark_signature/$mark_mask" lookup main
done

# Reapply active Qiyou exclusions immediately after PassWall startup/reload.
[ ! -x /etc/qiyou-passwall-compat.sh ] || /etc/qiyou-passwall-compat.sh
