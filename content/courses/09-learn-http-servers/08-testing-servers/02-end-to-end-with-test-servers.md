---
title: End-to-End with Test Servers
quiz:
  - question: What does a test gain by using a real test server instead of calling `ServeHTTP` with a recorder?
    options:
      - text: Nothing; the two are identical
      - text: The request goes through a real HTTP client and server, so it exercises things like header parsing, body streaming, redirects, timeouts and cancellation
        correct: true
      - text: The test no longer needs a handler
      - text: It runs faster
    explanation: |
      A recorder calls your handler like a function. A test server speaks real HTTP, which
      catches problems that only appear on the wire. It's a bit slower, so use recorders for
      most unit tests and servers for end-to-end checks.
  - question: |
      Which statement about `httptest.NewTestServer(t, handler)` (Go 1.27) is **false**?
    options:
      - text: It uses an in-memory network by default
      - text: It shuts itself down when the test ends
      - text: Its client sends requests to the server whatever hostname you use
      - text: You must call `srv.Close()` yourself, as with `httptest.NewServer`
        correct: true
    explanation: |
      `NewTestServer` registers a cleanup with `t`, so it's closed automatically. It also
      fails the test if the handler panics. `httptest.NewServer` servers still need
      `defer srv.Close()`.
---

Recorder tests call handlers like functions. That's fast, but it skips the actual HTTP
layer: parsing, headers on the wire, streaming bodies, connection handling. For
end-to-end tests you want a real client talking to a real server. `httptest` has two
ways to do that.

## httptest.NewServer: a real server on loopback

`httptest.NewServer(handler)` starts a real `http.Server` on a random port on
`127.0.0.1`, in the background, and gives you its URL. You already used it in the
routing chapter to watch a 307 redirect in action:

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})
	mux.HandleFunc("POST /api/squeaks", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "stored %d bytes from %s", len(body), r.UserAgent())
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	fmt.Println("listening on loopback:", strings.HasPrefix(srv.URL, "http://127.0.0.1:"))

	client := srv.Client()
	resp, err := client.Get(srv.URL + "/api/healthz")
	if err != nil {
		panic(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println(resp.StatusCode, string(b))

	resp, err = client.Post(srv.URL+"/api/squeaks", "application/json", strings.NewReader(`{"body":"hi"}`))
	if err != nil {
		panic(err)
	}
	b, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println(resp.StatusCode, string(b))
}
```

```
listening on loopback: true
200 OK
201 stored 13 bytes from Go-http-client/1.1
```

`srv.Client()` returns an `*http.Client` configured for that server (it also trusts the
test certificate if you use `NewTLSServer`). Always `Close` the server, or it leaks a
goroutine and a port.

Using the loopback network is fine for most suites. On busy CI machines, thousands of
tests opening ports can hit port exhaustion or firewall prompts, and nothing about
time is under your control.

## httptest.NewTestServer: in memory (Go 1.27)

Go 1.27 adds a newer constructor designed for tests:

```go
func TestHealthz(t *testing.T) {
	srv := httptest.NewTestServer(t, newRouter(newTestConfig(t)))

	resp, err := srv.Client().Get("http://squeak.test/api/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/healthz status = %d, want 200", resp.StatusCode)
	}
}
```

What's different:

- **It uses an in-memory network.** No ports, no loopback, no firewall. Requests and
  responses still travel as real HTTP, just through in-process pipes.
- **Any URL works.** Its client sends every request to the test server whatever the
  host, so you can write readable URLs like `http://squeak.test/...`. (`srv.URL` is
  `http://example.com`.)
- **It cleans up after itself.** It registers `t.Cleanup` to close the server.
- **Handler panics fail the test** (except `http.ErrAbortHandler`), instead of being
  swallowed and logged by the server.
- **It works with `testing/synctest`.** Because there's no real network, a test server
  created inside a synctest bubble runs entirely on the bubble's fake clock. That's the
  subject of this chapter's exercise.

If you ever need the loopback network for a `NewTestServer`, calling `srv.Start()`
switches it over and sets `srv.Listener`.

## Configuring the server under test

Both kinds expose `srv.Config`, the underlying `*http.Server`, which you can tweak before
the first request. That lets you test that timeouts actually protect you:

```go
srv := httptest.NewTestServer(t, handler)
srv.Config.ReadHeaderTimeout = 50 * time.Millisecond
```

## What to test end to end

You don't need end-to-end tests for every branch. The recorder tests cover those. Good
candidates for a real server:

- **The full router with all middleware**, as the app actually runs it, for a few key
  flows: register, log in, post a squeak with the token, delete it.
- **Anything about the wire:** redirects, `Content-Length`, compression, keep-alive.
- **Timeouts and cancellation:** a client that gives up, a slow handler, graceful
  shutdown.

Make sure the router you test is built by the same function `main` uses (`newRouter(cfg)`
or similar). If tests assemble their own mux, they can pass while production is broken.
