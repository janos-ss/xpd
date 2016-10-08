Cross-Post Detector
===================

Analyze a configurable collection of feeds to detect cross-posts and react to them.

> **cross-post**
>
> *verb*:
> post (a message, link, image, etc.) to more than one online location, such as a blog, social media website, or forum.
> *"the app is set up so that you can easily cross-post your item on Craigslist too"*

Download and run
----------------

See the [Releases](https://github.com/xpd-org/xpd/releases) tab on GitHub for binaries.

Create a configuration like [this](https://github.com/xpd-org/xpd/blob/master/xpd.yml.example), and save it in a file named `xpd.yml`:

    feeds:
      - id: so-sonarqube
        url: http://stackoverflow.com/feeds/tag?tagnames=sonarqube&sort=newest
      - id: gg-sonarqube
        url: https://groups.google.com/forum/feed/sonarqube/msgs/rss.xml?num=15
    detectors:
      - sameBodyDetector
      - similarWordCountDetector

Edit the list of `feeds`:

- The `id` can be arbitrary, it is only used for display purposes.
- The URL should be an RSS feed.

Edit the list of `detectors`:

- These are algorithms that try to match new posts to existing posts.
- The algorithms are tried in their defined order.
  Note that when a detector finds suspected cross-posts, the remaining detectors are skipped. This may be improved in a future release.
- The currently supported algorithms:
    - `sameBodyDetector` matches posts with the exact same text body
    - `similarWordCountDetector` matches posts with similar count of the same words (&plusmn;10% of total word count)

Develop
-------

Download dependencies:

    go get

Create configuration: copy `xpd.yml.example` to `xpd.yml` and customize.

Run using the default configuration file (`xpd.yml`):

    ./run.sh

To generate test coverage report, run:

    ./coverage.sh

To rebuild the binaries for multiple platforms, run:

    ./rebuild-all.sh
