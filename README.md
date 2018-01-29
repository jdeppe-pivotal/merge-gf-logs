### Merge Geode logs

Utility to merge and display multiple Geode log files sorted according to the logged timestamps on every line. For example:

    ./ml locator-1:./locatorgemfire1_9238/system.log locator-2:./locatorgemfire2_9292/system.log
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


Build it with:

    go build -o ml ./src/cmd/main.go