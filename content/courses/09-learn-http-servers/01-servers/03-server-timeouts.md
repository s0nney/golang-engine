---
title: http.Server and Timeouts
quiz:
  - question: Which `http.Server` field defends against a client that sends its request headers extremely slowly?
    options:
      - text: '`WriteTimeout`'
      - text: '`IdleTimeout`'
      - text: '`ReadHeaderTimeout`'
        correct: true
      - text: '`MaxHeaderBytes`'
    explanation: |
      `ReadHeaderTimeout` limits how long the server waits for the request line and
      headers. A "slow loris" client that dribbles headers gets disconnected once it
      expires. `MaxHeaderBytes` limits their *size*, not their speed.
  - question: |
      A handler streams a large export that takes 45 seconds to write. The server has
      `WriteTimeout: 10 * time.Second`. What happens?
    options:
      - text: The server waits, because `WriteTimeout` only applies to headers
      - text: Writes start failing after about 10 seconds and the client gets a truncated response
        correct: true
      - text: The client gets a `408 Request Timeout` status
      - text: The handler goroutine is killed after 10 seconds
    explanation: |
      `WriteTimeout` sets a deadline on the connection covering the response write.
      Once it passes, writes fail and the connection is closed, so the client sees a cut-off
      response. Nothing stops your handler goroutine, though: it keeps running until it
      returns. Long-running endpoints need a larger timeout or `http.ResponseController`
      to extend the deadline.
---

`http.ListenAndServe` is a convenience wrapper. Under the hood it builds an
`http.Server` with every field at its zero value, and for timeouts zero means
**"wait forever"**. For anything facing the internet, build the server yourself:

```go
srv := &http.Server{
	Addr:              ":8080",
	Handler:           mux,
	ReadHeaderTimeout: 5 * time.Second,
	ReadTimeout:       10 * time.Second,
	WriteTimeout:      15 * time.Second,
	IdleTimeout:       60 * time.Second,
	MaxHeaderBytes:    1 << 20, // 1 MiB
}
log.Fatal(srv.ListenAndServe())
```

## What each timeout covers

A connection's life has phases, and each timeout guards one of them:

| Field | Clock starts | Clock stops | Guards against |
|---|---|---|---|
| `ReadHeaderTimeout` | connection accepted | headers read | clients that send headers slowly |
| `ReadTimeout` | connection accepted | whole body read | slow uploads |
| `WriteTimeout` | headers read | response written | clients that read the response slowly |
| `IdleTimeout` | response finished | next request starts | keep-alive connections that just sit there |

If `ReadHeaderTimeout` is zero, the server falls back to `ReadTimeout`. If
`IdleTimeout` is zero, it falls back to `ReadTimeout` too. `ReadHeaderTimeout` is the
one you should *never* leave out: static analysis tools such as `gosec` flag servers
that don't set it.

## Watching a timeout fire

This program starts a real server on a random loopback port (`127.0.0.1:0` lets the OS
pick a free one), then plays a *slow loris* client: it starts a request and never
finishes the headers.

```go
package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

func main() {
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "squeak")
		}),
		ReadHeaderTimeout: 200 * time.Millisecond,
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	go srv.Serve(ln)
	defer srv.Close()

	// A "slow loris" client: starts a request and never finishes the headers.
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Fprint(conn, "GET / HTTP/1.1\r\nHost: squeak\r\n")

	start := time.Now()
	reply, _ := io.ReadAll(conn) // returns when the server hangs up
	fmt.Printf("server hung up after ~%v\n", time.Since(start).Round(100*time.Millisecond))
	fmt.Printf("reply: %q\n", reply)
}
```

It prints something like:

```
server hung up after ~200ms
reply: ""
```

Without `ReadHeaderTimeout` that `io.ReadAll` would wait forever, and so would the
server's goroutine and socket. Multiply by ten thousand attackers and you're out of
file descriptors.

Notice `srv.Serve(ln)`: instead of `ListenAndServe`, you can create the listener
yourself and hand it over. That's how tests and the later lessons run a server
without blocking `main`.

## Timeouts don't stop your handler

A server timeout closes the *connection*. It doesn't magically stop the handler
function, which keeps running until it returns. To make handler work stop early, you
watch `r.Context()`, which is cancelled when the client goes away or the server shuts
down. You'll use that in the testing and production chapters.

## Picking numbers

There's no universal answer. A JSON API like Squeak answers in milliseconds, so a few
seconds for headers and 10 to 30 seconds for writes is generous. Endpoints that stream
big files need longer writes. Some teams set a short server-wide `WriteTimeout` and
extend it per request with `http.NewResponseController(w).SetWriteDeadline(...)` where
needed.

## Other fields worth knowing

- `MaxHeaderBytes` caps header size (default 1 MB).
- `ErrorLog` is where the server logs connection-level problems.
- `BaseContext` and `ConnContext` let you seed every request's context.

From now on, Squeak always runs on an `http.Server` you built yourself.
