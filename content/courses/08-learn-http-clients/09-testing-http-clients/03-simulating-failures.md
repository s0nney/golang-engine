---
title: Simulating Failures
quiz:
  - question: 'How can a test handler make the client see a dropped connection, with no HTTP response at all?'
    options:
      - text: '`w.WriteHeader(0)`'
      - text: '`panic(http.ErrAbortHandler)`'
        correct: true
      - text: '`return` without writing anything'
    explanation: |
      Returning without writing sends an empty `200 OK`. Panicking with
      `http.ErrAbortHandler` makes the server abort the connection (without logging a
      stack trace), so the client gets an error such as `EOF`.
  - question: 'Why does `Flaky` count requests with `atomic.Int64` instead of a plain `int`?'
    options:
      - text: '`int` can''t count past 32,767'
      - text: The server runs each request's handler in its own goroutine, so concurrent requests would race on a plain `int`
        correct: true
      - text: Atomics are required by `http.Handler`
    explanation: |
      `net/http` handles every connection on its own goroutine. A fake that's hit
      concurrently needs a mutex or an atomic, or it can miscount (and `go test -race`
      will flag it).
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"sync/atomic"
    )

    // Flaky is a fake API for tests. The first Failures requests get
    // 503 Service Unavailable with Content-Type application/json and the body
    // {"error": "try again"}. Every request after that gets 200 OK with Body.
    // It must be safe for concurrent use.
    type Flaky struct {
    	Failures int
    	Body     string

    	requests atomic.Int64
    }

    func (f *Flaky) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	// ?
    }

    // Requests reports how many requests f has handled so far.
    func (f *Flaky) Requests() int {
    	// ?
    	return 0
    }

    func main() {
    	api := &Flaky{Failures: 2, Body: `{"status":"ok"}`}
    	srv := httptest.NewServer(api)
    	defer srv.Close()

    	for i := range 4 {
    		resp, err := http.Get(srv.URL + "/health")
    		if err != nil {
    			fmt.Println(err)
    			return
    		}
    		body, _ := io.ReadAll(resp.Body)
    		resp.Body.Close()
    		fmt.Printf("request %d: %s %s\n", i+1, resp.Status, body)
    	}
    	fmt.Println("requests seen:", api.Requests())
    }
  solution: |
    package main

    import (
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"sync/atomic"
    )

    // Flaky is a fake API for tests. The first Failures requests get
    // 503 Service Unavailable with Content-Type application/json and the body
    // {"error": "try again"}. Every request after that gets 200 OK with Body.
    // It must be safe for concurrent use.
    type Flaky struct {
    	Failures int
    	Body     string

    	requests atomic.Int64
    }

    func (f *Flaky) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	n := f.requests.Add(1)
    	if n <= int64(f.Failures) {
    		w.Header().Set("Content-Type", "application/json")
    		w.WriteHeader(http.StatusServiceUnavailable)
    		io.WriteString(w, `{"error": "try again"}`)
    		return
    	}
    	io.WriteString(w, f.Body)
    }

    func (f *Flaky) Requests() int {
    	return int(f.requests.Load())
    }

    func main() {
    	api := &Flaky{Failures: 2, Body: `{"status":"ok"}`}
    	srv := httptest.NewServer(api)
    	defer srv.Close()

    	for i := range 4 {
    		resp, err := http.Get(srv.URL + "/health")
    		if err != nil {
    			fmt.Println(err)
    			return
    		}
    		body, _ := io.ReadAll(resp.Body)
    		resp.Body.Close()
    		fmt.Printf("request %d: %s %s\n", i+1, resp.Status, body)
    	}
    	fmt.Println("requests seen:", api.Requests())
    }
  tests: |
    package main

    import (
    	"encoding/json/v2"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"testing"
    )

    func get(t *testing.T, client *http.Client, url string) (int, string, string) {
    	t.Helper()
    	resp, err := client.Get(url)
    	if err != nil {
    		t.Fatalf("GET %s: %v", url, err)
    	}
    	defer resp.Body.Close()
    	b, _ := io.ReadAll(resp.Body)
    	return resp.StatusCode, resp.Header.Get("Content-Type"), string(b)
    }

    func TestFlakySequence(t *testing.T) {
    	api := &Flaky{Failures: 3, Body: "hello"}
    	srv := httptest.NewServer(api)
    	defer srv.Close()

    	for i := 1; i <= 5; i++ {
    		status, ct, body := get(t, srv.Client(), srv.URL+"/anything")
    		if i <= 3 {
    			if status != http.StatusServiceUnavailable {
    				t.Errorf("request %d: status %d, want 503 (Failures is 3)", i, status)
    			}
    			if !strings.HasPrefix(ct, "application/json") {
    				t.Errorf("request %d: Content-Type %q, want application/json", i, ct)
    			}
    			var e struct {
    				Error string `json:"error"`
    			}
    			if err := json.Unmarshal([]byte(body), &e); err != nil || e.Error != "try again" {
    				t.Errorf("request %d: body %q, want {\"error\": \"try again\"}", i, body)
    			}
    		} else {
    			if status != http.StatusOK || body != "hello" {
    				t.Errorf("request %d: got %d %q, want 200 \"hello\"", i, status, body)
    			}
    		}
    		if got := api.Requests(); got != i {
    			t.Errorf("after %d requests, Requests() = %d", i, got)
    		}
    	}
    }

    func TestFlakyNoFailures(t *testing.T) {
    	api := &Flaky{Body: `{"ok":true}`}
    	srv := httptest.NewServer(api)
    	defer srv.Close()
    	if status, _, body := get(t, srv.Client(), srv.URL); status != 200 || body != `{"ok":true}` {
    		t.Errorf("Failures 0: got %d %q, want 200 with the body", status, body)
    	}
    }

    func TestFlakyConcurrent(t *testing.T) {
    	api := &Flaky{Failures: 25, Body: "ok"}
    	srv := httptest.NewServer(api)
    	defer srv.Close()

    	var mu sync.Mutex
    	counts := map[int]int{}
    	var wg sync.WaitGroup
    	for range 100 {
    		wg.Go(func() {
    			resp, err := srv.Client().Get(srv.URL)
    			if err != nil {
    				t.Error(err)
    				return
    			}
    			resp.Body.Close()
    			mu.Lock()
    			counts[resp.StatusCode]++
    			mu.Unlock()
    		})
    	}
    	wg.Wait()
    	if counts[503] != 25 || counts[200] != 75 {
    		t.Errorf("100 concurrent requests with Failures 25: got %d x 503 and %d x 200, want 25 and 75 (use an atomic counter)", counts[503], counts[200])
    	}
    	if api.Requests() != 100 {
    		t.Errorf("Requests() = %d, want 100", api.Requests())
    	}
    }
---

Happy-path tests are the easy part. The bugs that wake people up at night live in the
unhappy paths: the 503 during a deploy, the proxy that returns HTML, the connection
that drops halfway through a body. With a fake server, each of those is a few lines.

## A catalogue of failures

```go
package main

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/bad-json", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id": 42, "title": "Login bro`)
	})
	mux.HandleFunc("/html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprint(w, "<html><body>502 Bad Gateway</body></html>")
	})
	mux.HandleFunc("/hang-up", func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler) // drop the connection without a response
	})
	mux.HandleFunc("/truncated", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		fmt.Fprint(w, `{"id": 42`) // promised 1000 bytes, sent 9
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	for _, path := range []string{"/bad-json", "/html", "/hang-up", "/truncated"} {
		resp, err := srv.Client().Get(srv.URL + path)
		if err != nil {
			fmt.Printf("%-10s request error: %v\n", path, err)
			continue
		}
		var v struct{ ID int }
		err = json.UnmarshalRead(resp.Body, &v)
		resp.Body.Close()
		fmt.Printf("%-10s %d, decode error: %v\n", path, resp.StatusCode, err)
	}
}
```

Output (with your own port):

```
/bad-json  200, decode error: jsontext: unexpected EOF within "/title" after offset 30
/html      502, decode error: jsontext: invalid character '<' at start of value
/hang-up   request error: Get "http://127.0.0.1:PORT/hang-up": EOF
/truncated 200, decode error: jsontext: read error: unexpected EOF
```

Each one targets a different path in your client:

| Fake | What it tests |
| --- | --- |
| **error status** (`w.WriteHeader(503)`) | status checks, `APIError`, retries |
| **HTML instead of JSON** | the status is checked before decoding, and the error message is sensible |
| **malformed JSON** | decode errors are returned, not ignored |
| **`panic(http.ErrAbortHandler)`** | transport errors: no response at all |
| **short body** (a `Content-Length` bigger than what's written) | errors while *reading* the body |
| **closed server** (`srv.Close()` first) | connection refused |

`http.ErrAbortHandler` is a special panic value: the server aborts the connection
without logging a stack trace. It's the cleanest way to simulate a server that dies
mid-request.

## Stateful fakes

Retries need a server that fails **and then recovers**. That means state: a counter
of requests so far. Remember that `net/http` runs each request's handler on its own
goroutine, so the counter must be safe for concurrent use:

```go
type Flaky struct {
	Failures int
	Body     string

	requests atomic.Int64
}
```

A struct with a `ServeHTTP` method *is* an `http.Handler`, so you can pass `&Flaky{...}`
straight to `httptest.NewServer`. Exporting a `Requests()` method lets tests assert
how many attempts the client made, which is how you'd check "retries exactly 3
times" or "doesn't retry a 404".

## Asserting on the requests, too

A fake can record what it received, so the test checks both sides:

```go
var mu sync.Mutex
var paths []string
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	paths = append(paths, r.URL.RequestURI())
	mu.Unlock()
	// ... respond ...
}))
```

With the mutex, the recorded list is safe to read after the client call returns. The
pagination exercise in chapter 7 checked its cursors exactly like this.

## Table-driven failure tests

Once you have fakes, a table of cases keeps the tests short:

```go
tests := []struct {
	name    string
	handler http.HandlerFunc
	wantErr bool
}{
	{"ok", ok(`{"id":1}`), false},
	{"503", status(503), true},
	{"html", html(502), true},
	{"bad json", ok(`{"id":`), true},
	{"hang up", func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) }, true},
}
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		srv := httptest.NewTestServer(t, tt.handler)
		c := &Client{BaseURL: "https://api.trackr.dev", HTTPClient: srv.Client()}
		_, err := c.GetIssue(t.Context(), 1)
		if (err != nil) != tt.wantErr {
			t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
		}
	})
}
```

(`ok`, `status` and `html` are tiny helper functions that return handlers.)

## Your turn

Build the reusable fake. Complete `Flaky` so that:

1. `ServeHTTP` counts every request. The first `Failures` requests get
   `503 Service Unavailable` with `Content-Type: application/json` and the body
   `{"error": "try again"}`. Set the header **before** `WriteHeader`.
2. Every later request gets `200 OK` with `Body` as the body.
3. `Requests` returns how many requests have been handled.
4. It works correctly when many requests arrive at once. Use the `atomic.Int64` field.
   Its `Add` method returns the new value, so each request gets its own number.
