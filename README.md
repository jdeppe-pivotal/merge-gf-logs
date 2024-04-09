### Merge GemFire or Geode logs

Utility to merge and display multiple GemFire log files sorted according to the logged timestamps on every line.

Usage:

    ml [--grep regex] [--highlight regex] [<tag1>:]logfile1...

For example:

    ./ml locator-1:locatorgemfire1_9238/system.log locator-2:locatorgemfire2_9292/system.log
    [locator2] [info 2018/01/25 19:09:36.933 UTC locatorgemfire2_7f6dea03-dbdf-4d14-7891-987e4dbff0d8_9292 <unicast receiver,7f6dea03-dbdf-4d14-7891-987e4dbff0d8-10191> tid=0x25] Peer locator received new membership view: View[7f6dea03-dbdf-4d14-7891-987e4dbff0d8(locatorgemfire1_7f6dea03-dbdf-4d14-7891-987e4dbff0d8_9238:9238:locator)<ec><v0>:1024|1] members: [7f6dea03-dbdf-4d14-7891-987e4dbff0d8(locatorgemfire1_7f6dea03-dbdf-4d14-7891-987e4dbff0d8_9238:9238:locator)<ec><v0>:1024, 7f6dea03-dbdf-4d14-7891-987e4dbff0d8(locatorgemfire2_7f6dea03-dbdf-4d14-7891-987e4dbff0d8_9292:9292:locator)<ec><v1>:1025]
    [locator2]
    [locator1] [info 2018/01/25 19:09:36.945 UTC locatorgemfire1_7f6dea03-dbdf-4d14-7891-987e4dbff0d8_9238 <vm_7_thr_12_locator1_7f6dea03-dbdf-4d14-7891-987e4dbff0d8_9238> tid=0x15] Initializing region _monitoringRegion_10.254.0.154<v0>1024
    [locator1]
    [locator2] [info 2018/01/25 19:09:36.948 UTC locatorgemfire2_7f6dea03-dbdf-4d14-7891-987e4dbff0d8_9292 <vm_8_thr_13_locator2_7f6dea03-dbdf-4d14-7891-987e4dbff0d8_9292> tid=0x15] Finished joining (took 351ms).
    [locator2]
    [locator1] [info 2018/01/25 19:09:36.948 UTC locatorgemfire1_7f6dea03-dbdf-4d14-7891-987e4dbff0d8_9238 <Geode Membership View Creator> tid=0x2d] finished waiting for responses to view preparation
    [locator1]
    [locator2] [info 2018/01/25 19:09:36.949 UTC locatorgemfire2_7f6dea03-dbdf-4d14-7891-987e4dbff0d8_9292 <vm_8_thr_13_locator2_7f6dea03-dbdf-4d14-7891-987e4dbff0d8_9292> tid=0x15] Starting DistributionManager 7f6dea03-dbdf-4d14-7891-987e4dbff0d8(locatorgemfire2_7f6dea03-dbdf-4d14-7891-987e4dbff0d8_9292:9292:locator)<ec><v1>:1025.  (took 617 ms)
    [locator2]

Each file can be assigned a 'tag' which is used in the output to identify where the line originated.
By default, the filename is used as the tag.

Output is colored by default from a simple palette of 8 colors. Coloring can be controlled using
the `--color` switch. Options are `off`, `light` and `dark` (default).

Output can be limited by timestamp using the following options: (timestamps must be given in the
exact form output by GemFire; for example: `2018/01/25 19:09:36.949 UTC`)

- `--start`
- `--stop`
- `--duration` Duration is given in seconds and is relative to either _start_ or _stop_.
Thus _stop_ = _start_ + _duration_ or, conversely, _start_ = _stop_ - _duration_.
If both _start_ and _stop_ are provided then the duration is ignored.

Using the `--grep` option will only output lines matching the given regex.

The `--highlight` option highlights any text matching the given regex.

By default, the utility will attempt to match rolled log files with the same system in order to
keep the same coloring of log lines. This can be disabled with the `--no-roll` flag.

### Building

Simply:

    go build -o ml
