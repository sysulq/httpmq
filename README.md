httpmq
======
[![Build Status](https://travis-ci.org/sysulq/httpmq.svg?branch=master)](https://travis-ci.org/sysulq/httpmq)
[![Docker Pulls](https://img.shields.io/docker/pulls/sophos/httpmq.svg)](https://hub.docker.com/r/sophos/httpmq/)
[![Go Report Card](https://goreportcard.com/badge/github.com/sysulq/httpmq)](https://goreportcard.com/report/github.com/sysulq/httpmq)

> httpmq is powered by [Kod](https://github.com/go-kod/kod), which is a dependency injection framework for Go.  
> It is designed to be simple and easy to use, and to provide a consistent way to manage dependencies across your application.

Httpmq is a simple HTTP message queue written in Go with goleveldb, just like httpsqs wriiten in C with Tokyo Cabinet.

Also, you can refer to [github.com/sysulq/httpmq-rs#benchmark](https://github.com/sysulq/httpmq-rs#benchmark) for the benchmark of **Rust** implementation.

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
Docker
```
  docker run -d -it -p 1218:1218 sophos/httpmq
```

Binary
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

Test machine(Hackintosh):

```text
                    'c.          
                 ,xNMM.          ----------------------- 
               .OMMMMo           OS: macOS 11.6.1 20G224 x86_64 
               OMMM0,            Host: Hackintosh (SMBIOS: iMac20,1) 
     .;loddo:' loolloddol;.      Kernel: 20.6.0 
   cKMMMMMMMMMMNWMMMMMMMMMM0:    Uptime: 13 hours, 16 mins 
 .KMMMMMMMMMMMMMMMMMMMMMMMWd.    Packages: 45 (brew) 
 XMMMMMMMMMMMMMMMMMMMMMMMX.      Shell: zsh 5.8 
;MMMMMMMMMMMMMMMMMMMMMMMM:       Resolution: 1920x1080@2x 
:MMMMMMMMMMMMMMMMMMMMMMMM:       DE: Aqua 
.MMMMMMMMMMMMMMMMMMMMMMMMX.      WM: Quartz Compositor 
 kMMMMMMMMMMMMMMMMMMMMMMMMWd.    WM Theme: Blue (Dark) 
 .XMMMMMMMMMMMMMMMMMMMMMMMMMMk   Terminal: vscode 
  .XMMMMMMMMMMMMMMMMMMMMMMMMK.   CPU: Intel i5-10600K (12) @ 4.10GHz 
    kMMMMMMMMMMMMMMMMMMMMMMd     GPU: Radeon Pro W5500X 
     ;KMMMMMMMWXXWMMMMMMMk.      Memory: 17549MiB / 32768MiB 
       .cooc,.    .,coo:.
```

fasthttp
--------
### PUT queue:

```bash
wrk -c 10 -t 2 -d 10s "http://127.0.0.1:1218/?name=xoyo&opt=put&data=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
Running 10s test @ http://127.0.0.1:1218/?name=xoyo&opt=put&data=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   211.87us  169.58us   4.97ms   98.28%
    Req/Sec    24.36k     2.23k   27.87k    64.36%
  489524 requests in 10.10s, 82.64MB read
Requests/sec:  48459.16
Transfer/sec:      8.18MB
```

### GET queue:

```bash
wrk -c 10 -t 2 -d 10s "http://127.0.0.1:1218/?name=xoyo&opt=get"
Running 10s test @ http://127.0.0.1:1218/?name=xoyo&opt=get
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   218.26us  251.84us  11.56ms   99.22%
    Req/Sec    23.37k     1.06k   24.70k    90.59%
  469685 requests in 10.10s, 303.14MB read
Requests/sec:  46504.02
Transfer/sec:     30.01MB
```
