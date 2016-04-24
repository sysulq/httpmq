// Copyright 2014 httpmq Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// httpmq is an open-source, lightweight and high-performance message queue.

package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	_ "net/http/pprof"
	"runtime"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/valyala/fasthttp"
)

// VERSION of httpmq
const VERSION = "0.5"

var db *leveldb.DB
var defaultMaxqueue, keepalive, cpu, cacheSize, writeBuffer *int
var ip, port, defaultAuth, dbPath *string
var verbose *bool

// httpmq read metadata api
// retrieve from leveldb
// name.maxqueue - maxqueue
// name.putpos - putpos
// name.getpos - getpos
func httpmqReadMetadata(name string) []string {
	maxqueue := name + ".maxqueue"
	data1, _ := db.Get([]byte(maxqueue), nil)
	if len(data1) == 0 {
		data1 = []byte(strconv.Itoa(*defaultMaxqueue))
	}
	putpos := name + ".putpos"
	data2, _ := db.Get([]byte(putpos), nil)
	getpos := name + ".getpos"
	data3, _ := db.Get([]byte(getpos), nil)
	return []string{string(data1), string(data2), string(data3)}
}

// httpmq now getpos api
// get the current getpos of httpmq for request
func httpmqNowGetpos(name string) string {
	metadata := httpmqReadMetadata(name)
	maxqueue, _ := strconv.Atoi(metadata[0])
	putpos, _ := strconv.Atoi(metadata[1])
	getpos, _ := strconv.Atoi(metadata[2])

	if getpos == 0 && putpos > 0 {
		getpos = 1 // first get operation, set getpos 1
	} else if getpos < putpos {
		getpos++ // 1nd lap, increase getpos
	} else if getpos > putpos && getpos < maxqueue {
		getpos++ // 2nd lap
	} else if getpos > putpos && getpos == maxqueue {
		getpos = 1 // 2nd first operation, set getpos 1
	} else {
		return "0" // all data in queue has been get
	}

	data := strconv.Itoa(getpos)
	db.Put([]byte(name+".getpos"), []byte(data), nil)
	return data
}

// httpmq now putpos api
// get the current putpos of httpmq for request
func httpmqNowPutpos(name string) string {
	metadata := httpmqReadMetadata(name)
	maxqueue, _ := strconv.Atoi(metadata[0])
	putpos, _ := strconv.Atoi(metadata[1])
	getpos, _ := strconv.Atoi(metadata[2])

	putpos++              // increase put queue pos
	if putpos == getpos { // queue is full
		return "0" // return 0 to reject put operation
	} else if getpos <= 1 && putpos > maxqueue { // get operation less than 1
		return "0" // and queue is full, just reject it
	} else if putpos > maxqueue { //  2nd lap
		metadata[1] = "1" // reset putpos as 1 and write to leveldb
	} else { // 1nd lap, convert int to string and write to leveldb
		metadata[1] = strconv.Itoa(putpos)
	}

	db.Put([]byte(name+".putpos"), []byte(metadata[1]), nil)

	return metadata[1]
}

func init() {
	defaultMaxqueue = flag.Int("maxqueue", 1000000, "the max queue length")
	ip = flag.String("ip", "0.0.0.0", "ip address to listen on")
	port = flag.String("port", "1218", "port to listen on")
	defaultAuth = flag.String("auth", "", "auth password to access httpmq")
	dbPath = flag.String("db", "level.db", "database path")
	cacheSize = flag.Int("cache", 64, "cache size(MB)")
	writeBuffer = flag.Int("buffer", 32, "write buffer(MB)")
	cpu = flag.Int("cpu", runtime.NumCPU(), "cpu number for httpmq")
	keepalive = flag.Int("k", 60, "keepalive timeout for httpmq")
	flag.Parse()

	var err error
	db, err = leveldb.OpenFile(*dbPath, &opt.Options{BlockCacheCapacity: *cacheSize,
		WriteBuffer: *writeBuffer * 1024 * 1024})
	if err != nil {
		log.Fatalln("db.Get(), err:", err)
	}
}

func main() {
	runtime.GOMAXPROCS(*cpu)

	sync := &opt.WriteOptions{Sync: true}

	putnamechan := make(chan string, 100)
	putposchan := make(chan string, 100)
	getnamechan := make(chan string, 100)
	getposchan := make(chan string, 100)

	go func(chan string, chan string) {
		for {
			name := <-putnamechan
			putpos := httpmqNowPutpos(name)
			putposchan <- putpos
		}
	}(putnamechan, putposchan)

	go func(chan string, chan string) {
		for {
			name := <-getnamechan
			getpos := httpmqNowGetpos(name)
			getposchan <- getpos
		}
	}(getnamechan, getposchan)

	m := func(ctx *fasthttp.RequestCtx) {
		var data string
		var buf []byte
		auth := string(ctx.FormValue("auth"))
		name := string(ctx.FormValue("name"))
		opt := string(ctx.FormValue("opt"))
		pos := string(ctx.FormValue("pos"))
		num := string(ctx.FormValue("num"))
		charset := string(ctx.FormValue("charset"))

		if *defaultAuth != "" && *defaultAuth != auth {
			ctx.Write([]byte("HTTPMQ_AUTH_FAILED"))
			return
		}

		method := string(ctx.Method())
		if method == "GET" {
			data = string(ctx.FormValue("data"))
		} else if method == "POST" {
			if string(ctx.Request.Header.ContentType()) == "application/x-www-form-urlencoded" {
				data = string(ctx.FormValue("data"))
			} else {
				buf = ctx.PostBody()
			}
		}

		if len(name) == 0 || len(opt) == 0 {
			ctx.Write([]byte("HTTPMQ_ERROR"))
			return
		}

		ctx.Response.Header.Set("Connection", "keep-alive")
		ctx.Response.Header.Set("Cache-Control", "no-cache")
		ctx.Response.Header.Set("Content-type", "text/plain")
		if len(charset) > 0 {
			ctx.Response.Header.Set("Content-type", "text/plain; charset="+charset)
		}

		if opt == "put" {
			if len(data) == 0 && len(buf) == 0 {
				ctx.Write([]byte("HTTPMQ_PUT_ERROR"))
				return
			}

			putnamechan <- name
			putpos := <-putposchan

			if putpos != "0" {
				queueName := name + putpos
				if data != "" {
					db.Put([]byte(queueName), []byte(data), nil)
				} else if len(buf) > 0 {
					db.Put([]byte(queueName), buf, nil)
				}
				ctx.Response.Header.Set("Pos", putpos)
				ctx.Write([]byte("HTTPMQ_PUT_OK"))
			} else {
				ctx.Write([]byte("HTTPMQ_PUT_END"))
			}
		} else if opt == "get" {
			getnamechan <- name
			getpos := <-getposchan

			if getpos == "0" {
				ctx.Write([]byte("HTTPMQ_GET_END"))
			} else {
				queueName := name + getpos
				v, err := db.Get([]byte(queueName), nil)
				if err == nil {
					ctx.Response.Header.Set("Pos", getpos)
					ctx.Write(v)
				} else {
					ctx.Write([]byte("HTTPMQ_GET_ERROR"))
				}
			}
		} else if opt == "status" {
			metadata := httpmqReadMetadata(name)
			maxqueue, _ := strconv.Atoi(metadata[0])
			putpos, _ := strconv.Atoi(metadata[1])
			getpos, _ := strconv.Atoi(metadata[2])

			var ungetnum float64
			var putTimes, getTimes string
			if putpos >= getpos {
				ungetnum = math.Abs(float64(putpos - getpos))
				putTimes = "1st lap"
				getTimes = "1st lap"
			} else if putpos < getpos {
				ungetnum = math.Abs(float64(maxqueue - getpos + putpos))
				putTimes = "2nd lap"
				getTimes = "1st lap"
			}

			buf := fmt.Sprintf("HTTP Simple Queue Service v%s\n", VERSION)
			buf += fmt.Sprintf("------------------------------\n")
			buf += fmt.Sprintf("Queue Name: %s\n", name)
			buf += fmt.Sprintf("Maximum number of queues: %d\n", maxqueue)
			buf += fmt.Sprintf("Put position of queue (%s): %d\n", putTimes, putpos)
			buf += fmt.Sprintf("Get position of queue (%s): %d\n", getTimes, getpos)
			buf += fmt.Sprintf("Number of unread queue: %g\n\n", ungetnum)

			ctx.Write([]byte(buf))
		} else if opt == "view" {
			v, err := db.Get([]byte(name+pos), nil)
			if err == nil {
				ctx.Write([]byte(v))
			} else {
				ctx.Write([]byte("HTTPMQ_VIEW_ERROR"))
			}
		} else if opt == "reset" {
			maxqueue := strconv.Itoa(*defaultMaxqueue)
			db.Put([]byte(name+".maxqueue"), []byte(maxqueue), sync)
			db.Put([]byte(name+".putpos"), []byte("0"), sync)
			db.Put([]byte(name+".getpos"), []byte("0"), sync)
			ctx.Write([]byte("HTTPMQ_RESET_OK"))
		} else if opt == "maxqueue" {
			maxqueue, _ := strconv.Atoi(num)
			if maxqueue > 0 && maxqueue <= 10000000 {
				db.Put([]byte(name+".maxqueue"), []byte(num), sync)
				ctx.Write([]byte("HTTPMQ_MAXQUEUE_OK"))
			} else {
				ctx.Write([]byte("HTTPMQ_MAXQUEUE_CANCLE"))
			}
		}
	}

	log.Fatal(fasthttp.ListenAndServe(*ip+":"+*port, m))
}
