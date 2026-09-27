---
title: Timeouts
quiz:
  - question: '`client.Timeout` is 5 seconds. The headers arrive after 1 second, and the body takes 10 seconds to download. What happens?'
    options:
      - text: It succeeds, because the timeout only covers waiting for headers
      - text: Reading the body fails after about 4 more seconds, because `Client.Timeout` covers the whole exchange including the body
        correct: true
      - text: '`Do` returns an error after 5 seconds'
    explanation: |
      `Do` returns after 1 second with the headers. The timer keeps running while you
      read, and the read fails once 5 seconds in total have passed. Large downloads
      need a bigger (or no) client timeout, plus a context.
  - question: 'A client has `Timeout: 10*time.Second`, and the request''s context has a 2-second deadline. When does it give up?'
    options:
      - text: After 2 seconds
        correct: true
      - text: After 10 seconds
      - text: After 12 seconds
    explanation: |
      Both limits apply, and whichever runs out first wins. Contexts are how callers
      tighten the limit for a single call.
---

A request with no time limit is a bug waiting for a bad network day. Go gives you
several layers of timeouts. You'll use two of them all the time.

## Client.Timeout: a ceiling for everything

`Client.Timeout` limits the **entire** exchange: connecting, sending, waiting for
headers **and reading the body**. The timer starts when you call `Do` and keeps
running after `Do` returns, until you've finished reading the body.

## Context deadlines: per-call limits

A context on the request limits that one call. It also covers the whole exchange,
including the body, and can be cancelled early (Ctrl+C, a parent operation that gave
up, and so on).

When both are set, **the earlier one wins**. A good pattern is a generous
`Client.Timeout` as a safety net (so nothing hangs forever) plus contexts for
per-command deadlines and cancellation.

## Seeing them fire

This fake API sends headers at once and then trickles the body out over 250 ms. The
`/hang` path waits 300 ms before sending anything.

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	// A fake API that sends headers straight away, then dribbles the body.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/hang" {
			time.Sleep(300 * time.Millisecond)
		}
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		for range 5 {
			fmt.Fprintln(w, "chunk")
			w.(http.Flusher).Flush()
			time.Sleep(50 * time.Millisecond)
		}
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 100 * time.Millisecond}
	resp, err := client.Get(srv.URL + "/slow-body")
	if err != nil {
		fmt.Println("get:", err)
		return
	}
	_, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println("read:", err)

	_, err = client.Get(srv.URL + "/hang")
	fmt.Println("hang:", err)
	if netErr, ok := errors.AsType[net.Error](err); ok {
		fmt.Println("timeout?", netErr.Timeout())
	}
	fmt.Println("deadline?", errors.Is(err, context.DeadlineExceeded))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/hang", nil)
	_, err = http.DefaultClient.Do(req)
	fmt.Println("ctx:", err)
}
```

Output (the port changes every run):

```
read: context deadline exceeded (Client.Timeout or context cancellation while reading body)
hang: Get "http://127.0.0.1:36261/hang": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
timeout? true
deadline? true
ctx: Get "http://127.0.0.1:36261/hang": context deadline exceeded
```

Three things to notice:

1. The first `Get` **succeeded**. The timeout hit while reading the body, so the error
   came from `io.ReadAll`. Always check read errors.
2. Timeout errors satisfy `net.Error` with `Timeout() == true`, and they also match
   `context.DeadlineExceeded` with `errors.Is`. Either check works.
3. The context version gives the same kind of error, but it's controlled by the
   caller.

## Telling the user

```go
_, err := client.Do(req)
if errors.Is(err, context.DeadlineExceeded) {
	return fmt.Errorf("Trackr didn't respond in time (is api.trackr.dev down?): %w", err)
}
if errors.Is(err, context.Canceled) {
	return err // the user pressed Ctrl+C: stay quiet
}
```

## Finer-grained transport timeouts

The `Transport` has timeouts for individual phases. `http.DefaultTransport` already
sets some of them:

| Setting | Default in `DefaultTransport` | Limits |
| --- | --- | --- |
| `net.Dialer{Timeout}` | 30s | opening the TCP connection |
| `TLSHandshakeTimeout` | 10s | the TLS handshake |
| `ResponseHeaderTimeout` | none | waiting for headers after sending the request |
| `IdleConnTimeout` | 90s | how long an idle pooled connection is kept |

`ResponseHeaderTimeout` is the useful one for downloads: "the server must *start*
answering within 5 seconds", without limiting how long a big body may take. To set it,
clone the default transport (so you keep its other good defaults) and change one field:

```go
t := http.DefaultTransport.(*http.Transport).Clone()
t.ResponseHeaderTimeout = 5 * time.Second
client := &http.Client{Transport: t} // no overall Timeout: bodies may be large
```

## What value?

There's no universal answer. For `trackr`'s small JSON calls, 10 to 30 seconds overall
is reasonable. The important thing is that there **is** a limit. And if you retry
(next lessons), remember each attempt gets its own `Client.Timeout`, so a context
should cap the total.
