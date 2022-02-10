# Log Collector
This project aims at providing a REST API to read log files in `/var/log`. The events in the log files are
returned in inversed chronological order i.e. from the most recent to the oldest event.

Here are the assumptions and limitations for this service:
- events are separated by new line
- event cannot be bigger than 4 KB. If no event separator after 4 KB, an error is returned
- encoding of the file: ascii?
- the maximum number of events that can be returned is limited to 10,000
- most recent events are located at the end of file
- modifying a file while it is read is not supported and will return an error

## How to test
To launch the tests, you can use:
```shell
make gotest
```