package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	//"io"
	"io/ioutil"
	"log"
	"net/http"
	//"net/url"
	"os"
	"strconv"
	"time"
)

var db *bolt.DB

func httpmq_read_maxqueue(name string) int {
	queue_name := name + ".maxqueue"
	var tmp string
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		if b != nil {
			data := b.Get([]byte(queue_name))
			tmp = string(data)
		}
		return nil
	})

	maxqueue, _ := strconv.Atoi(tmp)
	if maxqueue == 0 {
		maxqueue = 10000
	}
	log.Println("httpmq_read_maxqueue:", maxqueue)
	return maxqueue
}

func httpmq_read_putpos(name string) int {
	queue_name := name + ".putpos"
	var tmp string
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		data := b.Get([]byte(queue_name))
		tmp = string(data)
		return nil
	})

	putpos, _ := strconv.Atoi(tmp)
	log.Println("httpmq_read_putpos:", putpos)
	return putpos
}

func httpmq_read_getpos(name string) int {
	queue_name := name + ".getpos"
	var tmp string
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		data := b.Get([]byte(queue_name))
		tmp = string(data)
		return nil
	})

	getpos, _ := strconv.Atoi(tmp)
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
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(name))
			err := b.Put([]byte(queue_name), []byte("1"))
			return err
		})
	} else if getpos < putpos {
		getpos++
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(name))
			err := b.Put([]byte(queue_name), []byte(strconv.Itoa(getpos)))
			return err
		})
	} else if getpos > putpos && getpos < maxqueue {
		getpos++
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(name))
			err := b.Put([]byte(queue_name), []byte(strconv.Itoa(getpos)))
			return err
		})
	} else if getpos > putpos && getpos == maxqueue {
		getpos = 1
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(name))
			err := b.Put([]byte(queue_name), []byte("1"))
			return err
		})
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
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(name))
			err := b.Put([]byte(queue_name), []byte("1"))
			return err
		})
	} else {
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(name))
			err := b.Put([]byte(queue_name), []byte(strconv.Itoa(putpos)))
			return err
		})
	}
	log.Println("httpmq_now_putpos:", putpos)
	return putpos
}

func main() {
	var err error
	db, err = bolt.Open("test.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	go func() {
		// Grab the initial stats.
		prev := db.Stats()

		for {
			// Wait for 10s.
			time.Sleep(100 * time.Second)

			// Grab the current stats and diff them.
			stats := db.Stats()
			diff := stats.Sub(&prev)

			// Encode stats to JSON and print to STDERR.
			json.NewEncoder(os.Stderr).Encode(diff)

			// Save stats for the next loop.
			prev = stats
		}
	}()

	log.SetOutput(ioutil.Discard)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		//auth := r.FormValue("auth")
		name := r.FormValue("name")
		charset := r.FormValue("charset")
		opt := r.FormValue("opt")
		data := r.FormValue("data")

		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Cache-Control", "no-cache")

		if len(charset) > 0 {
			w.Header().Set("Content-type", "text/plain; charset="+charset)
		} else {
			w.Header().Set("Content-type", "text/plain")
		}

		if len(name) == 0 && len(opt) == 0 {
			w.Write([]byte("HTTPMQ_ERROR"))
			return
		}

		db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				log.Println("create bucket: ", err)
				return err
			}

			return nil
		})

		if opt == "put" {
			putpos := httpmq_now_putpos(name)
			buf, _ := ioutil.ReadAll(r.Body)
			queue_name := name + strconv.Itoa(putpos)
			log.Println("put queue name:", queue_name)

			if len(buf) > 0 {
				log.Println("buf:", string(buf))

				if putpos > 0 {

					db.Update(func(tx *bolt.Tx) error {
						b := tx.Bucket([]byte(name))

						err = b.Put([]byte(queue_name), []byte(buf))
						w.Write([]byte("HTTPMQ_PUT_OK"))
						return nil
					})
				} else {
					w.Write([]byte("HTTPMQ_PUT_END"))
					return
				}
			} else {
				log.Println("data:", data)

				if putpos > 0 {
					db.Update(func(tx *bolt.Tx) error {
						b := tx.Bucket([]byte(name))

						err = b.Put([]byte(queue_name), []byte(data))
						w.Write([]byte("HTTPMQ_PUT_OK"))
						return nil
					})
				} else {
					w.Write([]byte("HTTPMQ_PUT_END"))
					return
				}
			}

			return
		} else if opt == "get" {

			getpos := httpmq_now_getpos(name)
			if getpos == 0 {
				w.Write([]byte("HTTPMQ_GET_END"))
			} else {
				queue_name := name + strconv.Itoa(getpos)
				log.Println("get queue name:", queue_name)
				db.View(func(tx *bolt.Tx) error {

					b := tx.Bucket([]byte(name))
					v := b.Get([]byte(queue_name))
					if v != nil {
						w.Write(v)
					} else {
						w.Write([]byte("HTTPMQ_GET_END"))
					}
					return nil
				})
			}

		}

	})

	s := &http.Server{
		Addr:           ":9001",
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())

}
