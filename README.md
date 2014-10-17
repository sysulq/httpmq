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

To be continued...
