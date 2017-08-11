httpmq
======
[![Build Status](https://travis-ci.org/hnlq715/httpmq.svg?branch=master)](https://travis-ci.org/hnlq715/httpmq)
[![Docker Pulls](https://img.shields.io/docker/pulls/sophos/httpmq.svg)](https://hub.docker.com/r/sophos/httpmq/)
[![Go Report Card](https://goreportcard.com/badge/github.com/hnlq715/httpmq)](https://goreportcard.com/report/github.com/hnlq715/httpmq)


Httpmq is a simple HTTP message queue written in Go with goleveldb, just like httpsqs wriiten in C with Tokyo Cabinet.

Feature
======

* Very simple, less than 300 lines Go code.
* Very fast, more than 10000 requests/sec.
* High concurrency, support the tens of thousands of concurrent connections.
* Multiple queue.
* Low memory consumption, mass data storage, storage dozens of GB of data takes less than 100MB of physical memory buffer.
* Convenient to change the maximum queue length of per-queue.
* Queue status view.
* Be able to view the contents of the specified queue ID.
* Multi-Character encoding support.

Usage
======
  ```
Usage of ./httpmq:
  -auth string
    	auth password to access httpmq
  -buffer int
    	write buffer(MB) (default 32)
  -cache int
    	cache size(MB) (default 64)
  -cpu int
    	cpu number for httpmq (default 4)
  -db string
    	database path (default "level.db")
  -ip string
    	ip address to listen on (default "0.0.0.0")
  -k int
    	keepalive timeout for httpmq (default 60)
  -maxqueue int
    	the max queue length (default 1000000)
  -port string
    	port to listen on (default "1218")
  ```

1. PUT text message into a queue

  HTTP GET protocol (Using curl for example):
  ```
  curl "http://host:port/?name=your_queue_name&opt=put&data=url_encoded_text_message&auth=mypass123"
  ```
  HTTP POST protocol (Using curl for example):
  ```
  curl -d "url_encoded_text_message" "http://host:port/?name=your_queue_name&opt=put&auth=mypass123"
  ```

2. GET text message from a queue

  HTTP GET protocol (Using curl for example):
  ```
  curl "http://host:port/?charset=utf-8&name=your_queue_name&opt=get&auth=mypass123"
  ```

3. View queue status

  HTTP GET protocol (Using curl for example):
  ```
  curl "http://host:port/?name=your_queue_name&opt=status&auth=mypass123"
  ```
4. View queue details

  HTTP GET protocol (Using curl for example):
  ```
  curl "http://host:port/?name=your_queue_name&opt=view&pos=1&auth=mypass123"
  ```
5. Reset queue

  HTTP GET protocol (Using curl for example):
  ```
  curl "http://host:port/?name=your_queue_name&opt=reset&pos=1&auth=mypass123"
  ```

Benchmark
========

Test machine(Mac Pro):
  ```
  2.7 GHz Intel Core i5
  8 GB 1867 MHz DDR3
  ```

net/http
--------
### PUT queue:
```
wrk -c 10 -t 2 -d 10s "http://127.0.0.1:1218/?name=xoyo&opt=put&data=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
Running 10s test @ http://127.0.0.1:1218/?name=xoyo&opt=put&data=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.12ms    4.36ms  78.68ms   98.01%
    Req/Sec     9.05k     1.46k   12.38k    75.00%
  180109 requests in 10.01s, 30.30MB read
Requests/sec:  18000.03
Transfer/sec:      3.03MB
```
### GET queue:
```
wrk -c 10 -t 2 -d 10s "http://127.0.0.1:1218/?name=xoyo&opt=get"
Running 10s test @ http://127.0.0.1:1218/?name=xoyo&opt=get
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   758.39us    1.27ms  33.49ms   96.82%
    Req/Sec     8.03k     3.42k   14.66k    62.50%
  159807 requests in 10.01s, 103.07MB read
Requests/sec:  15970.14
Transfer/sec:     10.30MB
```


fasthttp
--------

### PUT queue:
```
wrk -c 100 -t 2 -d 10s "http://127.0.0.1:1218/?name=xoyo&opt=get"
Running 10s test @ http://127.0.0.1:1218/?name=xoyo&opt=get
  2 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     5.15ms    5.13ms  87.09ms   95.58%
    Req/Sec    10.85k     2.23k   13.96k    79.00%
  216088 requests in 10.02s, 143.14MB read
Requests/sec:  21572.72
Transfer/sec:     14.29MB
```

### GET queue:
```
wrk -c 10 -t 2 -d 10s "http://127.0.0.1:1218/?name=xoyo&opt=put&data=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
Running 10s test @ http://127.0.0.1:1218/?name=xoyo&opt=put&data=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   749.65us    2.56ms  44.05ms   98.00%
    Req/Sec    11.21k     1.99k   13.93k    68.00%
  223151 requests in 10.01s, 41.39MB read
Requests/sec:  22293.74
Transfer/sec:      4.14MB
```
