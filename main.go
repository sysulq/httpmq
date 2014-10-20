package main

import (
	"flag"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/cache"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const VERSION = "0.1"

var db *leveldb.DB
var default_maxqueue, cpu, cacheSize, writeBuffer, keepalive, readtimeout, writetimeout *int
var ip, port, default_auth, dbpath *string
var verbose *bool
var mu sync.Mutex

func httpmq_read_metadata(name string) []string {
	queue_name := name + ".metadata"
	data, _ := db.Get([]byte(queue_name), nil)
	metadata := strings.Split(string(data), ",")
	if len(metadata) == 1 {
		metadata = []string{strconv.Itoa(*default_maxqueue), "0", "0"}
	}

	return metadata
}

func httpmq_write_metadata(name string, metadata []string) {
	queue_name := name + ".metadata"
	db.Put([]byte(queue_name), []byte(strings.Join(metadata, ",")), nil)
}

func httpmq_now_getpos(name string) string {
	metadata := httpmq_read_metadata(name)

	maxqueue, _ := strconv.Atoi(metadata[0])
	putpos, _ := strconv.Atoi(metadata[1])
	getpos, _ := strconv.Atoi(metadata[2])

	if getpos == 0 && putpos > 0 {
		getpos = 1
	} else if getpos < putpos {
		getpos++
	} else if getpos > putpos && getpos < maxqueue {
		getpos++
	} else if getpos > putpos && getpos == maxqueue {
		getpos = 1
	} else {
		return "0"
	}

	metadata[2] = strconv.Itoa(getpos)
	httpmq_write_metadata(name, metadata)

	return metadata[2]
}

func httpmq_now_putpos(name string) string {
	metadata := httpmq_read_metadata(name)

	maxqueue, _ := strconv.Atoi(metadata[0])
	putpos, _ := strconv.Atoi(metadata[1])
	getpos, _ := strconv.Atoi(metadata[2])

	putpos++
	if putpos == getpos {
		return "0"
	} else if getpos <= 1 && putpos > maxqueue {
		return "0"
	} else if putpos > maxqueue {
		putpos = 1
	}

	metadata[1] = strconv.Itoa(putpos)
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
	verbose = flag.Bool("verbose", true, "output log")
	flag.Parse()

	var err error
	ca := cache.NewLRUCache(*cacheSize * 1024 * 1024)
	db, err = leveldb.OpenFile(*dbpath, &opt.Options{BlockCache: ca,
		WriteBuffer: *writeBuffer * 1024 * 1024})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if *verbose == false {
		log.SetOutput(ioutil.Discard)
	}

	runtime.GOMAXPROCS(*cpu)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var data string
		var buf []byte
		auth := r.FormValue("auth")
		name := r.FormValue("name")
		charset := r.FormValue("charset")
		opt := r.FormValue("opt")
		if r.Method == "GET" {
			data = r.FormValue("data")
		} else if r.Method == "POST" {
			if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
				data = r.PostFormValue("data")
			} else {
				buf, err = ioutil.ReadAll(r.Body)
				r.Body.Close()
			}
		}

		pos := r.FormValue("pos")

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
			mu.Lock()
			putpos := httpmq_now_putpos(name)
			mu.Unlock()

			queue_name := name + putpos
			if r.Method == "POST" {

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
			} else if r.Method == "GET" {

				if putpos != "0" {
					if data != "" {
						db.Put([]byte(queue_name), []byte(data), nil)
					} else {
						w.Write([]byte("HTTPMQ_PUT_ERROR"))
						return
					}
					w.Header().Set("Pos", putpos)
					w.Write([]byte("HTTPMQ_PUT_OK"))
				} else {
					w.Write([]byte("HTTPMQ_PUT_END"))
				}
			}
		} else if opt == "get" {
			mu.Lock()
			getpos := httpmq_now_getpos(name)
			mu.Unlock()

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

			var ungetnum int
			var put_times, get_times string
			if putpos >= getpos {
				ungetnum = putpos - getpos
				put_times = "1st lap"
				get_times = "1st lap"
			} else if putpos < getpos {
				ungetnum = maxqueue - getpos + putpos
				put_times = "2nd lap"
				get_times = "1st lap"
			}

			buf := fmt.Sprintf("HTTP message queue v%s\n", VERSION)
			buf += fmt.Sprintf("-----------------------\n")
			buf += fmt.Sprintf("Queue Name: %s\n", name)
			buf += fmt.Sprintf("Maximun number of queues: %d\n", maxqueue)
			buf += fmt.Sprintf("Put position of queue (%s): %d\n", put_times, putpos)
			buf += fmt.Sprintf("Get position of queue (%s): %d\n", get_times, getpos)
			buf += fmt.Sprintf("Number of unread queue: %d\n\n", ungetnum)

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
