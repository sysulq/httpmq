httpmq
======

This program is a simple HTTP message queue written in Go with gdleveldb, just like httpsqs wriiten in C with Tokyo Cabinet.

Feature
======

* Very simple
* Very fast, more than 10000 requests/sec
* High concurrency, support the tens of thousands of concurrent connections.
* Multiple queue
* Low memory consumption, mass data storage, storage dozens of GB of data takes less than 100MB of physical memory buffer.
* Convenient to change the maximum queue length of per-queue.
* Queue status view
* Be able to view the contents of the specified queue ID.
* Multi-Character encoding support

Usage
======
```
Usage of ./httpmq:
  -auth="": auth password to access httpmq
  -db="level.db": database path
  -ip="0.0.0.0": ip address to listen on
  -maxqueue=100000: the max queue length
  -port="1218": port to listen on
  -readtimeout=15: read timeout for an http request
  -verbose=false: output log
  -writetimeout=15: write timeout for an http request
```

(1). PUT text message into a queue

HTTP GET protocol (Using curl for example):
```
curl "http://host:port/?name=your_queue_name&opt=put&data=url_encoded_text_message&auth=mypass123"
```
HTTP POST protocol (Using curl for example):
```
curl -d "url_encoded_text_message" "http://host:port/?name=your_queue_name&opt=put&auth=mypass123"
```

(2). GET text message from a queue

HTTP GET protocol (Using curl for example):
```
curl "http://host:port/?charset=utf-8&name=your_queue_name&opt=get&auth=mypass123"
```
```
curl "http://host:port/?charset=gb2312&name=your_queue_name&opt=get&auth=mypass123"
```
(3). View queue status

HTTP GET protocol (Using curl for example):
```
curl "http://host:port/?name=your_queue_name&opt=status&auth=mypass123"
```


Benchmark
========

Test machine:
```
2  AMD Athlon(tm) II X2 245 Processor
        Size: 2048 MB
        Locator: DIMM0
        Range Size: 2 GB
        Size: 2048 MB
        Locator: DIMM1
        Range Size: 2 GB
        Size: No Module Installed
        Locator: DIMM2
        Size: No Module Installed
        Locator: DIMM3
```


1. put queue:
```
ab -k -c 1000 -n 10000 "http://127.0.0.1:1218/?name=xoyo&opt=put&data=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
This is ApacheBench, Version 2.3 <$Revision: 655654 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient)
Completed 1000 requests
Completed 2000 requests
Completed 3000 requests
Completed 4000 requests
Completed 5000 requests
Completed 6000 requests
Completed 7000 requests
Completed 8000 requests
Completed 9000 requests
Completed 10000 requests
Finished 10000 requests


Server Software:        
Server Hostname:        127.0.0.1
Server Port:            1218

Document Path:          /?name=xoyo&opt=put&data=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
Document Length:        13 bytes

Concurrency Level:      1000
Time taken for tests:   0.771 seconds
Complete requests:      10000
Failed requests:        0
Write errors:           0
Keep-Alive requests:    10000
Total transferred:      1640000 bytes
HTML transferred:       130000 bytes
Requests per second:    12964.69 [#/sec] (mean)
Time per request:       77.133 [ms] (mean)
Time per request:       0.077 [ms] (mean, across all concurrent requests)
Transfer rate:          2076.38 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    2   7.7      0      41
Processing:     0   70  74.9     73     473
Waiting:        0   70  74.9     73     473
Total:          0   72  75.8     76     473

Percentage of the requests served within a certain time (ms)
  50%     76
  66%     91
  75%     98
  80%    110
  90%    183
  95%    216
  98%    272
  99%    310
 100%    473 (longest request)
```

2. GET queue:
```
ab -k -c 1000 -n 10000 "http://127.0.0.1:1218/?name=xoyo&opt=get"                                                                                                   [system]
This is ApacheBench, Version 2.3 <$Revision: 655654 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient)
Completed 1000 requests
Completed 2000 requests
Completed 3000 requests
Completed 4000 requests
Completed 5000 requests
Completed 6000 requests
Completed 7000 requests
Completed 8000 requests
Completed 9000 requests
Completed 10000 requests
Finished 10000 requests


Server Software:        
Server Hostname:        127.0.0.1
Server Port:            1218

Document Path:          /?name=xoyo&opt=get
Document Length:        512 bytes

Concurrency Level:      1000
Time taken for tests:   0.703 seconds
Complete requests:      10000
Failed requests:        0
Write errors:           0
Keep-Alive requests:    10000
Total transferred:      6640000 bytes
HTML transferred:       5120000 bytes
Requests per second:    14227.83 [#/sec] (mean)
Time per request:       70.285 [ms] (mean)
Time per request:       0.070 [ms] (mean, across all concurrent requests)
Transfer rate:          9225.86 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    1   5.3      0      33
Processing:     0   49  61.2     20     449
Waiting:        0   49  61.2     20     449
Total:          0   50  62.0     22     471

Percentage of the requests served within a certain time (ms)
  50%     22
  66%     67
  75%     87
  80%    105
  90%    128
  95%    161
  98%    224
  99%    240
 100%    471 (longest request)
```
