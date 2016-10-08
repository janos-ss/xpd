Cross-Post Detector
===================

Analyze a configurable collection of feeds to detect cross-posts and react to them.

> **cross-post**
>
> *verb*
>
> post (a message, link, image, etc.) to more than one online location, such as a blog, social media website, or forum.
>
> *"the app is set up so that you can easily cross-post your item on Craigslist too"*

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
