#!/bin/sh -e
# 
# File: list.sh
# Purpose: show all running xpd screen sessions
#

screen -ls | grep -F .xpd-
