# Log Collector
This project aims at providing a REST API to read log files in `/var/log`. The events in the log files are
returned in inversed chronological order i.e. from the most recent to the oldest event.

Here are the assumptions and limitations for this service:
- events are separated by new line
- user knows the file name in /var/log
- file is a text file, there is no check if it is a binary file
- only file directly located in /var/log i.e. file located in /var/log/subdir cannot be accessed
- event cannot be bigger than 4 KB. If no event separator after 4 KB, the content of 4 KB is returned as is
- the maximum number of events that can be returned is limited to 10,000
- most recent events are located at the end of file
- modifying a file while reading it is not supported and will return an error

## How to test
To launch the tests, you can use:
```shell
make gotest
```