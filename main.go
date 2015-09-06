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
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strconv"
	"time"
)

// httpmq version
const VERSION = "0.4"

var db *leveldb.DB
var default_maxqueue, keepalive, cpu, cacheSize, writeBuffer *int
var ip, port, default_auth, dbpath *string
var verbose *bool

// httpmq read metadata api
// retrieve from leveldb
// name.maxqueue - maxqueue
// name.putpos - putpos
// name.getpos - getpos
func httpmq_read_metadata(name string) []string {
	maxqueue := name + ".maxqueue"
	data1, _ := db.Get([]byte(maxqueue), nil)
	if len(data1) == 0 {
		data1 = []byte(strconv.Itoa(*default_maxqueue))
	}
	putpos := name + ".putpos"
	data2, _ := db.Get([]byte(putpos), nil)
	getpos := name + ".getpos"
	data3, _ := db.Get([]byte(getpos), nil)
	return []string{string(data1), string(data2), string(data3)}
}

// httpmq now getpos api
// get the current getpos of httpmq for request
func httpmq_now_getpos(name string) string {
	metadata := httpmq_read_metadata(name)
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
func httpmq_now_putpos(name string) string {
	metadata := httpmq_read_metadata(name)
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

func main() {
	default_maxqueue = flag.Int("maxqueue", 1000000, "the max queue length")
	ip = flag.String("ip", "0.0.0.0", "ip address to listen on")
	port = flag.String("port", "1218", "port to listen on")
	default_auth = flag.String("auth", "", "auth password to access httpmq")
	dbpath = flag.String("db", "level.db", "database path")
	cacheSize = flag.Int("cache", 64, "cache size(MB)")
	writeBuffer = flag.Int("buffer", 32, "write buffer(MB)")
	cpu = flag.Int("cpu", runtime.NumCPU(), "cpu number for httpmq")
	keepalive = flag.Int("k", 60, "keepalive timeout for httpmq")
	flag.Parse()

	var err error
	db, err = leveldb.OpenFile(*dbpath, &opt.Options{BlockCacheCapacity: *cacheSize,
		WriteBuffer: *writeBuffer * 1024 * 1024})
	if err != nil {
		log.Fatalln("db.Get(), err:", err)
	}

	runtime.GOMAXPROCS(*cpu)
	sync := &opt.WriteOptions{Sync: true}

	putnamechan := make(chan string, 100)
	putposchan := make(chan string, 100)
	getnamechan := make(chan string, 100)
	getposchan := make(chan string, 100)
	putipchan := make(chan string, 100)

	go func(chan string, chan string) {
		for {
			name := <-putnamechan
			putpos := httpmq_now_putpos(name)
			putposchan <- putpos
		}
	}(putnamechan, putposchan)

	go func(chan string, chan string) {
		for {
			name := <-getnamechan
			getpos := httpmq_now_getpos(name)
			getposchan <- getpos
		}
	}(getnamechan, getposchan)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var data string
		var buf []byte
		auth := r.FormValue("auth")
		name := r.FormValue("name")
		opt := r.FormValue("opt")
		pos := r.FormValue("pos")
		num := r.FormValue("num")
		charset := r.FormValue("charset")

		if *default_auth != "" && *default_auth != auth {
			w.Write([]byte("HTTPMQ_AUTH_FAILED"))
			return
		}

		if r.Method == "GET" {
			data = r.FormValue("data")
		} else if r.Method == "POST" {
			if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
				data = r.PostFormValue("data")
			} else {
				buf, _ = ioutil.ReadAll(r.Body)
				r.Body.Close()
			}
		}

		if len(name) == 0 || len(opt) == 0 {
			w.Write([]byte("HTTPMQ_ERROR"))
			return
		}

		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-type", "text/plain")
		if len(charset) > 0 {
			w.Header().Set("Content-type", "text/plain; charset="+charset)
		}

		if opt == "put" {
			if len(data) == 0 && len(buf) == 0 {
				w.Write([]byte("HTTPMQ_PUT_ERROR"))
				return
			}

			putnamechan <- name
			putpos := <-putposchan

			if putpos != "0" {
				queue_name := name + putpos
				if data != "" {
					db.Put([]byte(queue_name), []byte(data), nil)
				} else if len(buf) > 0 {
					db.Put([]byte(queue_name), buf, nil)
				}
				ip, _, _ := net.SplitHostPort(r.RemoteAddr)
				putipchan <- ip
				w.Header().Set("Pos", putpos)
				w.Write([]byte("HTTPMQ_PUT_OK"))
			} else {
				w.Write([]byte("HTTPMQ_PUT_END"))
			}
		} else if opt == "get" {
			getnamechan <- name
			getpos := <-getposchan

			if getpos == "0" {
				w.Write([]byte("HTTPMQ_GET_END"))
			} else {
				queue_name := name + getpos
				v, err := db.Get([]byte(queue_name), nil)
				if err == nil {
					w.Header().Set("Pos", getpos)
					w.Write(v)
				} else {
					w.Write([]byte("HTTPMQ_GET_ERROR"))
				}
			}
		} else if opt == "status" {
			metadata := httpmq_read_metadata(name)
			maxqueue, _ := strconv.Atoi(metadata[0])
			putpos, _ := strconv.Atoi(metadata[1])
			getpos, _ := strconv.Atoi(metadata[2])

			var ungetnum float64
			var put_times, get_times string
			if putpos >= getpos {
				ungetnum = math.Abs(float64(putpos - getpos))
				put_times = "1st lap"
				get_times = "1st lap"
			} else if putpos < getpos {
				ungetnum = math.Abs(float64(maxqueue - getpos + putpos))
				put_times = "2nd lap"
				get_times = "1st lap"
			}

			buf := fmt.Sprintf("HTTP Simple Queue Service v%s\n", VERSION)
			buf += fmt.Sprintf("------------------------------\n")
			buf += fmt.Sprintf("Queue Name: %s\n", name)
			buf += fmt.Sprintf("Maximum number of queues: %d\n", maxqueue)
			buf += fmt.Sprintf("Put position of queue (%s): %d\n", put_times, putpos)
			buf += fmt.Sprintf("Get position of queue (%s): %d\n", get_times, getpos)
			buf += fmt.Sprintf("Number of unread queue: %g\n\n", ungetnum)

			w.Write([]byte(buf))
		} else if opt == "view" {
			v, err := db.Get([]byte(name+pos), nil)
			if err == nil {
				w.Write([]byte(v))
			} else {
				w.Write([]byte("HTTPMQ_VIEW_ERROR"))
			}
		} else if opt == "reset" {
			maxqueue := strconv.Itoa(*default_maxqueue)
			db.Put([]byte(name+".maxqueue"), []byte(maxqueue), sync)
			db.Put([]byte(name+".putpos"), []byte("0"), sync)
			db.Put([]byte(name+".getpos"), []byte("0"), sync)
			w.Write([]byte("HTTPMQ_RESET_OK"))
		} else if opt == "maxqueue" {
			maxqueue, _ := strconv.Atoi(num)
			if maxqueue > 0 && maxqueue <= 10000000 {
				db.Put([]byte(name+".maxqueue"), []byte(num), sync)
				w.Write([]byte("HTTPMQ_MAXQUEUE_OK"))
			} else {
				w.Write([]byte("HTTPMQ_MAXQUEUE_CANCLE"))
			}
		}
	})

	s := &http.Server{
		Addr:           *ip + ":" + *port,
		ReadTimeout:    time.Duration(*keepalive) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
