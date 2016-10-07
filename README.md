Cross-Post Detector
===================

Analyze a configurable collection of feeds to detect cross-posts and react to them.

Configure
---------

Copy `xpd.yml.example` to `xpd.yml` and edit the list of RSS feeds to monitor.

Run
---

    go run cmp/xpd/main.go

Build binaries for multiple platforms
-------------------------------------

Run:

    ./cmd/xpd/build.sh
