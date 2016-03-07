package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatus(t *testing.T) {
	n := setupMux()
	ts := httptest.NewServer(n)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/?name=TestStatus&opt=status")
	if err != nil {
		t.Fatal(err)
	}
	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	expected :=
		`HTTP Simple Queue Service v0.4
------------------------------
Queue Name: test
Maximum number of queues: 0
Put position of queue (1st lap): 0
Get position of queue (1st lap): 0
Number of unread queue: 0

`
	if string(greeting) == expected {
		t.Fatal(string(greeting))
	}
}

func TestPutGet(t *testing.T) {
	n := setupMux()
	ts := httptest.NewServer(n)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/?name=TestPutGet&opt=put&data=testdata")
	if err != nil {
		t.Fatal(err)
	}

	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	expected := `HTTPMQ_PUT_OK`

	if string(greeting) != expected {
		t.Fatal(string(greeting))
	}

	res, err = http.Get(ts.URL + "/?name=TestPutGet&opt=get")
	if err != nil {
		t.Fatal(err)
	}
	greeting, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	expected = `testdata`

	if string(greeting) != expected {
		t.Fatal(string(greeting))
	}
}

func BenchmarkPut(b *testing.B) {
	n := setupMux()
	ts := httptest.NewServer(n)
	defer ts.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := http.Get(ts.URL + "/?BenchmarkPut=test&opt=put&data=testdata")
		if err != nil {
			b.Fatal(err)
		}
		greeting, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			b.Fatal(greeting, err)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	n := setupMux()
	ts := httptest.NewServer(n)
	defer ts.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := http.Get(ts.URL + "/?name=BenchmarkGet&opt=get")
		if err != nil {
			b.Fatal(err)
		}
		greeting, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			b.Fatal(greeting, err)
		}
	}
}
