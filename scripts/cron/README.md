Scripts
-------

`xpd-all.sh`:
wrapper script to run `xpd` for each configuration file in an independent `screen` session.
It reuses an existing session or else it creates a new one.

`crontab.sh`:
helper script to add a line like this in `crontab` to periodically run `autossh.sh`:

    0 * * * * $PWD/xpd-all.sh

How to install
--------------

Simply run `./setup.sh` and follow the steps. This script does not
do anything. It only tells you the configuration it detected and
gives you the steps you need to follow to complete the configuration.


How to uninstall
----------------

- Remove any `cron` jobs running `xpd-all.sh`
- `./stop.sh` to stop any running `xpd` and `screen` instances

