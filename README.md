# Blackbox library
A cross-platform library to read CleanFlight flight logs.

Work in progress.

TODO:
* Improve logging
* Add more tests
* Manage to get the same output as the blackbox_decode Python implementation
  * Add option to discard whole stream when a corrupt P frame is found
* Add support for GPS frames
* Export Event frames
* Review decoder for simplifications in the bits operations
* Provide frames back as a channel to filter its content
* Make library more robust against stream corruption
* Implement multi-session support
