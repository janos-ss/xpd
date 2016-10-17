#!/bin/sh -e
# 
# File: list.sh
# Purpose: show all running xpd screen sessions
#

cd "$(dirname "$0")"
. ./common.sh

screen -ls | grep -F .$prefix-
