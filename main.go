package main

import (
	"flag"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

var db *leveldb.DB
var default_maxqueue, keepalive, readtimeout, writetimeout *int
var ip, port, default_auth, dbpath *string
var verbose *bool

func httpmq_read_maxqueue(name string) int {
	queue_name := name + ".maxqueue"
	data, _ := db.Get([]byte(queue_name), nil)

	maxqueue, _ := strconv.Atoi(string(data))
	if maxqueue == 0 {
		maxqueue = *default_maxqueue
	}
	log.Println("httpmq_read_maxqueue:", maxqueue)
	return maxqueue
}

func httpmq_read_putpos(name string) int {
	queue_name := name + ".putpos"
	data, _ := db.Get([]byte(queue_name), nil)

	putpos, _ := strconv.Atoi(string(data))
	log.Println("httpmq_read_putpos:", putpos)
	return putpos
}

func httpmq_read_getpos(name string) int {
	queue_name := name + ".getpos"
	data, _ := db.Get([]byte(queue_name), nil)

	getpos, _ := strconv.Atoi(string(data))
	log.Println("httpmq_read_getpos:", getpos)
	return getpos
}

func httpmq_now_getpos(name string) int {
	maxqueue := httpmq_read_maxqueue(name)
	putpos := httpmq_read_putpos(name)
	getpos := httpmq_read_getpos(name)

	queue_name := name + ".getpos"
	if getpos == 0 && putpos > 0 {
		getpos = 1
		db.Put([]byte(queue_name), []byte("1"), nil)
	} else if getpos < putpos {
		getpos++
		db.Put([]byte(queue_name), []byte(strconv.Itoa(getpos)), nil)
	} else if getpos > putpos && getpos < maxqueue {
		getpos++
		db.Put([]byte(queue_name), []byte(strconv.Itoa(getpos)), nil)
	} else if getpos > putpos && getpos == maxqueue {
		getpos = 1
		db.Put([]byte(queue_name), []byte("1"), nil)
	} else {
		getpos = 0
	}

	log.Println("httpmq_now_getpos:", getpos)

	return getpos
}

func httpmq_now_putpos(name string) int {
	maxqueue := httpmq_read_maxqueue(name)
	putpos := httpmq_read_putpos(name)
	getpos := httpmq_read_getpos(name)

	queue_name := name + ".putpos"

	putpos++
	if putpos == getpos {
		putpos = 0
	} else if getpos <= 1 && putpos > maxqueue {
		putpos = 1
		db.Put([]byte(queue_name), []byte("1"), nil)

	} else {
		db.Put([]byte(queue_name), []byte(strconv.Itoa(putpos)), nil)
	}
	log.Println("httpmq_now_putpos:", putpos)
	return putpos
}

func main() {
	default_maxqueue = flag.Int("maxqueue", 100000, "the max queue length")
	readtimeout = flag.Int("readtimeout", 15, "read timeout for an http request")
	writetimeout = flag.Int("writetimeout", 15, "write timeout for an http request")
	ip = flag.String("ip", "0.0.0.0", "ip address to listen on")
	port = flag.String("port", "1218", "port to listen on")
	default_auth = flag.String("auth", "", "auth password to access httpmq")
	dbpath = flag.String("db", "level.db", "database path")
	verbose = flag.Bool("verbose", false, "output log")
	flag.Parse()

	var err error
	db, err = leveldb.OpenFile(*dbpath, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if *verbose == false {
		log.SetOutput(ioutil.Discard)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		auth := r.FormValue("auth")
		name := r.FormValue("name")
		charset := r.FormValue("charset")
		opt := r.FormValue("opt")
		data := r.FormValue("data")
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
			putpos := httpmq_now_putpos(name)
			buf, _ := ioutil.ReadAll(r.Body)
			queue_name := name + strconv.Itoa(putpos)
			log.Println("put queue name:", queue_name)

			if len(buf) > 0 {
				log.Println("buf:", string(buf))

				if putpos > 0 {

					db.Put([]byte(queue_name), []byte(buf), nil)
					w.Write([]byte("HTTPMQ_PUT_OK"))
				} else {
					w.Write([]byte("HTTPMQ_PUT_END"))
				}
			} else {
				log.Println("data:", data)

				if putpos > 0 {
					db.Put([]byte(queue_name), []byte(data), nil)
					w.Write([]byte("HTTPMQ_PUT_OK"))
				} else {
					w.Write([]byte("HTTPMQ_PUT_END"))
				}
			}
		} else if opt == "get" {

			getpos := httpmq_now_getpos(name)
			if getpos == 0 {
				w.Write([]byte("HTTPMQ_GET_END"))
			} else {
				queue_name := name + strconv.Itoa(getpos)
				log.Println("get queue name:", queue_name)
				v, _ := db.Get([]byte(queue_name), nil)
				if v != nil {
					w.Write(v)
				} else {
					w.Write([]byte("HTTPMQ_GET_END"))
				}
			}

		} else if opt == "status" {
			getpos := httpmq_read_getpos(name)
			putpos := httpmq_read_putpos(name)
			maxqueue := httpmq_read_maxqueue(name)
			var ungetnum int
			var put_times, get_times string
			if putpos > getpos {
				ungetnum = putpos - getpos
				put_times = "1st lap"
				get_times = "1st lap"
			} else if putpos < getpos {
				ungetnum = maxqueue - getpos + putpos
				put_times = "2nd lap"
				get_times = "1st lap"
			}
			buf := "HTTP message queue\n"
			buf += "-------------------\n"
			buf += fmt.Sprintf("Queue Name: %s\n", name)
			buf += fmt.Sprintf("Maximun number of queues: %d\n", maxqueue)
			buf += fmt.Sprintf("Put position of queue (%s): %d\n", put_times, putpos)
			buf += fmt.Sprintf("Get position of queue (%s): %d\n", get_times, getpos)
			buf += fmt.Sprintf("Number of unread queue: %d\n\n", ungetnum)

			m := &runtime.MemStats{}
			runtime.ReadMemStats(m)

			buf += "Go runtime status\n"
			buf += "-------------------\n"
			buf += fmt.Sprintf("NumGoroutine: %d\n", runtime.NumGoroutine())
			buf += fmt.Sprintf("Memory Acquired: %d\n", m.Sys)
			buf += fmt.Sprintf("Memory Used: %d\n", m.Alloc)
			buf += fmt.Sprintf("EnableGc: %t\n", m.EnableGC)
			buf += fmt.Sprintf("NumGc: %d\n", m.NumGC)

			lastgc := time.Unix(0, int64(m.LastGC))

			buf += fmt.Sprintf("Pause Ns: %s\n", time.Nanosecond*time.Duration(m.PauseTotalNs))
			buf += fmt.Sprintf("Last Gc: %s\n\n", lastgc.Format("Mon Jan 2 15:04:05 -0700 MST 2006"))

			value, _ := db.GetProperty("leveldb.stats")
			buf += "Leveldb status\n"
			buf += value + "\n"

			w.Write([]byte(buf))
		} else if opt == "view" {
			v, _ := db.Get([]byte(name+pos), nil)
			if v != nil {
				w.Write([]byte(v))
			} else {
				w.Write([]byte("HTTPMQ_VIEW_ERROR"))
			}
		} else if opt == "reset" {
			db.Put([]byte(name+".putpos"), []byte(""), nil)
			db.Put([]byte(name+".getpos"), []byte(""), nil)
			db.Put([]byte(name+".maxqueue"), []byte(""), nil)
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
