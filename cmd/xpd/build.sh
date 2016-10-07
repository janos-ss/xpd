#!/usr/bin/env bash

cd $(dirname "$0")

v=$1
test "$v" || v=v0

set -ex

arch=amd64

mkdir -p target
GOARCH=$arch GOOS=darwin go build && mv xpd target/xpd-$v-osx_$arch
GOARCH=$arch GOOS=linux go build && mv xpd target/xpd-$v-linux_$arch
GOARCH=$arch GOOS=windows go build && mv xpd.exe target/xpd-$v-windows_$arch.exe
