# Blackbox Library

[![travis-ci](https://travis-ci.org/maxlaverse/blackbox-library.svg?branch=master)](https://travis-ci.org/maxlaverse/blackbox-library)
[![codecov](https://codecov.io/gh/maxlaverse/blackbox-library/branch/master/graph/badge.svg)](https://codecov.io/gh/maxlaverse/blackbox-library)



A cross-platform library to read [Cleanflight]/[Betaflight] blackbox flight logs.

Work in progress.

## How to install
```
$ git clone https://github.com/maxlaverse/blackbox-library
$ cd blackbox-library && make
```

## blackbox_decode
The tool `blackbox_decode` converts flight log files from binary format into CSV format.
It's a proof of concept which is meant as a drop-in replacement for [Cleanflight/blackbox-tools] when being used with the [Plasmatree/PID-Analyzer].

```
Usage:
  blackbox_decode [options] <input logs> [flags]

Flags:
      --debug         Show extra debugging information
  -h, --help          help for blackbox_decode
      --raw           Don't apply predictions to fields (show raw field deltas)
  -v, --verbose int   Be verbose on log output
```

**Example:**
```
$ bin/blackbox_decode ~/examples/LOG00007.BFL
Log 1 of 1, start 00:55.158, end 01:14.536, duration 00:19.378

Header bloc size:    3177 bytes
Frame bloc size:     1150586 bytes
Valid frames:        38767 (1150586 bytes)
Corrupted data:      0 bytes
Corrupted frames:    0

Frame stats:
  E frames       5 valid   0 corrupt         0 desync         8.4 bytes avg        42 bytes total    4 sizes
  S frames       4 valid   0 corrupt         0 desync         6.0 bytes avg        24 bytes total    1 sizes
  I frames     606 valid   0 corrupt         0 desync        54.0 bytes avg     32743 bytes total   15 sizes
  P frames   38152 valid   0 corrupt         0 desync        29.3 bytes avg   1117777 bytes total   12 sizes
    Frames   38767 valid   0 corrupt   29.7 bytes avg   1150586 bytes total
Data rate	 2001 Hz	 59538 bytes/s

$ head ~/examples/LOG00007.01.csv
loopIteration, time (us), axisP[0], axisP[1], axisP[2], axisI[0], axisI[1], axisI[2], axisD[0], axisD[1], axisF[0], axisF[1], axisF[2], rcCommand[0], rcCommand[1], rcCommand[2], rcCommand[3], setpoint[0], setpoint[1], setpoint[2], setpoint[3], vbatLatest (V), amperageLatest (A), rssi, gyroADC[0], gyroADC[1], gyroADC[2], accSmooth[0], accSmooth[1], accSmooth[2], debug[0], debug[1], debug[2], debug[3], motor[0], motor[1], motor[2], motor[3], energyCumulative (mAh), flightModeFlags (flags), stateFlags (flags), failsafePhase (flags), rxSignalReceived, rxFlightChannelsValid
52992, 55158008,  -1,  -4,  -1,   5,  -2,  -1,  -1, -30,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,   0,   3,   3,  60,   8, 2232,  -5, -14,   1,   0, 521, 650, 519, 665, 0.000000, 0, 0, IDLE, 0, 0
52993, 55158507,   0,  -4,   0,   5,  -2,  -1,  -5, -35,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,   0,   4,   2,  60,   6, 2231,  -3,  -9,  -1,   0, 516, 666, 505, 669, 0.002107, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1
52994, 55159007,   0,  -3,  -1,   5,  -2,  -1,  -6, -31,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,  -1,   3,   3,  60,   6, 2231,  -1,  -2,   0,   0, 525, 658, 512, 661, 0.004218, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1
52995, 55159511,   2,   0,  -1,   5,  -2,  -1,  -2, -19,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,  -3,   1,   3,  60,   4, 2229,  -1,   4,   1,   0, 544, 616, 551, 645, 0.006346, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1
52996, 55160009,   1,   3,  -1,   5,  -2,  -1,   5,   1,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,  -2,  -2,   3,  60,   4, 2229,   0,  10,   2,   0, 576, 558, 611, 612, 0.008449, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1
52997, 55160510,   1,   5,   0,   5,  -2,  -1,   7,  22,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 15.557, 13.775, 785,  -2,  -3,   2,  61,   4, 2229,   2,  14,   3,   0, 616, 513, 657, 570, 0.010366, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1
```

## To be done
* Improve test coverage
* Improve logging
* Add support for GPS frames
* Export event alongside to CSVs
* Simplify bits operations
* Simplify types
* Implement multi-session support

[Cleanflight]: https://github.com/cleanflight/cleanflight
[Betaflight]: https://github.com/betaflight/betaflight
[Cleanflight/blackbox-tools]: https://github.com/cleanflight/blackbox-tools
[Plasmatree/PID-Analyzer]: https://github.com/Plasmatree/PID-Analyzer