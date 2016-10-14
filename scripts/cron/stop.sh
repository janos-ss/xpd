#!/bin/sh -e
# 
# File: stop.sh
# Purpose: stop xpd processes
#

cd "$(dirname "$0")"
. ./common.sh

test $# = 0 && set -- $(names)

for name; do
    info stopping $name ...
    screen -S xpd-$name -wipe >/dev/null || :
    screen -S xpd-$name -p 0 -X stuff 
done
