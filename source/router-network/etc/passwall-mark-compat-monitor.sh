#!/bin/sh
# Boot services can create legacy chains after firewall and interface hooks ran.
while :; do
    /etc/passwall-mark-compat.sh
    sleep 10
done
