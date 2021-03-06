#!/bin/sh -e
# 
# File: single.sh
# Purpose: run xpd for named config
#

cd "$(dirname "$0")"
. ./common.sh

test $# = 1 || fatal usage: $0 name

name=$1
config=conf/$name.yml

mkdir -p logs

./xpd -config $config 2>&1 | tee -a logs/$name.log
