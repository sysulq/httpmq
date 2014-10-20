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
	"github.com/syndtr/goleveldb/leveldb/cache"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// httpmq version
const VERSION = "0.1"

var db *leveldb.DB
var default_maxqueue, cpu, cacheSize, writeBuffer, keepalive, readtimeout, writetimeout *int
var ip, port, default_auth, dbpath *string
var mu sync.Mutex

// httpmq read metadata api
// retrieve from leveldb and split with ","
// [name.metadata - maxqueue,putpos,getpos]
func httpmq_read_metadata(name string) []string {
	queue_name := name + ".metadata"
	data, _ := db.Get([]byte(queue_name), nil)

	metadata := strings.Split(string(data), ",")
	if len(metadata) == 1 {
		metadata = []string{strconv.Itoa(*default_maxqueue), "0", "0"}
	}

	return metadata
}

// httpmq write metadata api
// stored in leveldb and join with ","
// [name.metadata - maxqueue,putpos,getpos]
func httpmq_write_metadata(name string, metadata []string) {
	queue_name := name + ".metadata"
	db.Put([]byte(queue_name), []byte(strings.Join(metadata, ",")), nil)
}

// httpmq now getpos api
// get the current getpos of httpmq for request
// should be atomic with sync.Mutex
func httpmq_now_getpos(name string) string {
	mu.Lock()
	defer mu.Unlock()

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

	metadata[2] = strconv.Itoa(getpos)
	httpmq_write_metadata(name, metadata)
	return metadata[2]
}

// httpmq now putpos api
// get the current putpos of httpmq for request
// should be atomic with sync.Mutex
func httpmq_now_putpos(name string) string {
	mu.Lock()
	defer mu.Unlock()

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

	httpmq_write_metadata(name, metadata)

	return metadata[1]
}

func main() {
	default_maxqueue = flag.Int("maxqueue", 1000000, "the max queue length")
	readtimeout = flag.Int("readtimeout", 15, "read timeout for an http request")
	writetimeout = flag.Int("writetimeout", 15, "write timeout for an http request")
	ip = flag.String("ip", "0.0.0.0", "ip address to listen on")
	port = flag.String("port", "1218", "port to listen on")
	default_auth = flag.String("auth", "", "auth password to access httpmq")
	dbpath = flag.String("db", "level.db", "database path")
	cacheSize = flag.Int("cache", 8, "cache size(MB)")
	writeBuffer = flag.Int("buffer", 4, "write buffer(MB)")
	cpu = flag.Int("cpu", 1, "cpu number for httpmq")
	flag.Parse()

	var err error
	ca := cache.NewLRUCache(*cacheSize * 1024 * 1024)
	db, err = leveldb.OpenFile(*dbpath, &opt.Options{BlockCache: ca,
		WriteBuffer: *writeBuffer * 1024 * 1024})
	if err != nil {
		log.Fatalln("db.Get(), err:", err)
	}

	runtime.GOMAXPROCS(*cpu)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var data string
		var buf []byte
		auth := r.FormValue("auth")
		name := r.FormValue("name")
		opt := r.FormValue("opt")
		pos := r.FormValue("pos")
		num := r.FormValue("num")
		charset := r.FormValue("charset")

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

		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Cache-Control", "no-cache")

		if len(charset) > 0 {
			w.Header().Set("Content-type", "text/plain; charset="+charset)
		} else {
			w.Header().Set("Content-type", "text/plain")
		}

		if *default_auth != "" && *default_auth != auth {
			w.Write([]byte("HTTPMQ_AUTH_FAILED"))
			return
		}

		if len(name) == 0 && len(opt) == 0 {
			w.Write([]byte("HTTPMQ_ERROR"))
			return
		}

		if opt == "put" {
			putpos := httpmq_now_putpos(name)

			queue_name := name + putpos

			if putpos != "0" {
				if data != "" {
					db.Put([]byte(queue_name), []byte(data), nil)
				} else if len(buf) > 0 {
					db.Put([]byte(queue_name), buf, nil)
				} else {
					w.Write([]byte("HTTPMQ_PUT_ERROR"))
					return
				}
				w.Header().Set("Pos", putpos)
				w.Write([]byte("HTTPMQ_PUT_OK"))
			} else {
				w.Write([]byte("HTTPMQ_PUT_END"))
			}
		} else if opt == "get" {
			getpos := httpmq_now_getpos(name)

			if getpos == "0" {
				w.Write([]byte("HTTPMQ_GET_END"))
			} else {
				queue_name := name + getpos
				v, _ := db.Get([]byte(queue_name), nil)
				if v != nil {
					w.Header().Set("Pos", getpos)
					w.Write(v)
				} else {
					w.Write([]byte("HTTPMQ_GET_END"))
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

			buf := fmt.Sprintf("HTTP message queue v%s\n", VERSION)
			buf += fmt.Sprintf("-----------------------\n")
			buf += fmt.Sprintf("Queue Name: %s\n", name)
			buf += fmt.Sprintf("Maximun number of queues: %d\n", maxqueue)
			buf += fmt.Sprintf("Put position of queue (%s): %d\n", put_times, putpos)
			buf += fmt.Sprintf("Get position of queue (%s): %d\n", get_times, getpos)
			buf += fmt.Sprintf("Number of unread queue: %g\n\n", ungetnum)

			w.Write([]byte(buf))
		} else if opt == "view" {
			v, _ := db.Get([]byte(name+pos), nil)
			if v != nil {
				w.Write([]byte(v))
			} else {
				w.Write([]byte("HTTPMQ_VIEW_ERROR"))
			}
		} else if opt == "reset" {
			metadata := []string{strconv.Itoa(*default_maxqueue), "0", "0"}
			httpmq_write_metadata(name, metadata)
			w.Write([]byte("HTTPMQ_RESET_OK"))
		} else if opt == "maxqueue" {
			metadata := httpmq_read_metadata(name)
			maxqueue, _ := strconv.Atoi(num)
			if maxqueue > 0 && maxqueue <= 10000000 {
				metadata[0] = strconv.Itoa(maxqueue)
				httpmq_write_metadata(name, metadata)
				w.Write([]byte("HTTPMQ_MAXQUEUE_OK"))
			} else {
				w.Write([]byte("HTTPMQ_MAXQUEUE_CANCLE"))
			}
		}
	})

	s := &http.Server{
		Addr:           *ip + ":" + *port,
		ReadTimeout:    time.Duration(*readtimeout) * time.Second,
		WriteTimeout:   time.Duration(*writetimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
