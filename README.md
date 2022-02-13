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
- application is not secured by authZ

## How to test
To launch the tests, you can use:
```shell
make gotest
```

## How to run the application
The run target starts the application with a default configuration located in the config folder.
It starts the http server on the port 8888 and the targeted log folder is the log folder in the project.
```shell
make run
```
After running the make target, the http server should be started.
You can open a browser at http://localhost:8888/log?file=access_combined.log&filter=HEAD&limit=10 to verify that the
application processes the file as expected i.e. keeps only the most recent 10 logs that contain HEAD keyword

