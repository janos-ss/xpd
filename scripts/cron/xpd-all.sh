#!/bin/sh -e
# 
# File: xpd-all.sh
# Purpose: run xpd for each configuration, if not already running
#

cd "$(dirname "$0")"
. ./common.sh

match_session() {
    ./list.sh | grep -F .xpd-$1
}

for name in $(names); do
    config=conf/$name.yml

    # stop running session, unless it matches "Attached" or "Detached"
    if ! match_session $name | awk '$0 !~ /tached/ { exit 1 }'; then
        ./stop.sh $name
    fi

    if ! match_session $name >/dev/null; then
        echo \* starting $name ...
        screen -d -m -S xpd-$name ./xpd-single.sh $name
        echo \* done.
    fi
done
