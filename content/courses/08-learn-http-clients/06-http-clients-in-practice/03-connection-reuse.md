---
title: Connection Reuse
quiz:
  - question: |
      In Go 1.27, what happens to the connection here? The body is a small JSON error
      of about 100 bytes.

      ```go
      resp, err := client.Do(req)
      if err != nil {
          return err
      }
      defer resp.Body.Close()
      if resp.StatusCode != http.StatusOK {
          return fmt.Errorf("status %s", resp.Status) // body never read
      }
      ```
    options:
      - text: The connection is always thrown away, because the body wasn't read
      - text: '`Close` reads the small leftover body in the background, and the connection goes back into the pool'
        correct: true
      - text: The connection leaks until the program exits
    explanation: |
      Closing an unread body now drains it asynchronously, up to a conservative limit
      (256 KiB and 50 ms in the current source). Small error bodies no longer cost you
      the connection. You still have to call `Close`.
  - question: Which mistake stops **every** later request from reusing a connection?
    options:
      - text: Not reading a small body before closing it
      - text: Never calling `resp.Body.Close()`
        correct: true
      - text: Calling `io.ReadAll` on the body
    explanation: |
      An unclosed body keeps its connection busy forever, so the pool can't hand it out
      again. Nothing drains it for you until you call `Close`.
---

Opening a connection is expensive: a DNS lookup, a TCP handshake, and for HTTPS a TLS
handshake on top. That's several network round trips before your request even starts.
So `net/http` keeps connections open after a response and **reuses** them.

## Transports and the pool

The `http.Client` decides *what* to send (redirects, cookies, timeouts). The
**`Transport`** does the sending, and it owns the connection pool. With HTTP/1.1, a
connection carries one request at a time. After a response is finished, the connection
goes back into the pool as **idle**, ready for the next request to the same host.

A few `http.Transport` fields shape the pool:

| Field | Default | Meaning |
| --- | --- | --- |
| `MaxIdleConns` | 100 (in `DefaultTransport`) | idle connections kept in total |
| `MaxIdleConnsPerHost` | 2 | idle connections kept **per host** |
| `MaxConnsPerHost` | 0 (no limit) | all connections, busy or idle, per host |
| `IdleConnTimeout` | 90s | when an idle connection is closed |

`MaxIdleConnsPerHost` = 2 is low for a client that fires 20 concurrent requests at one
API: 18 of those connections get closed afterwards instead of reused. `trackr` bumps it:

```go
t := http.DefaultTransport.(*http.Transport).Clone()
t.MaxIdleConnsPerHost = 20
client := &http.Client{Timeout: 10 * time.Second, Transport: t}
```

(HTTP/2, which Go uses automatically for HTTPS servers that support it, sends many
requests over **one** connection at once, so these limits matter less there.)

## When can a connection go back?

Only when its response body is **finished** and **closed**. Until then, the rest of the
body might still be on the wire, and the next response can't start. Here it is in
action. `httptrace` reports whether each request got a reused connection:

```go
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"strings"
	"time"
)

func main() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/export" {
			fmt.Fprint(w, strings.Repeat("x", 1<<20)) // a 1 MiB body
			return
		}
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	}))
	defer srv.Close()
	client := srv.Client()

	// get sends one request and reports whether it got a reused connection.
	get := func(path, mode string) {
		var reused bool
		trace := &httptrace.ClientTrace{
			GotConn: func(info httptrace.GotConnInfo) { reused = info.Reused },
		}
		ctx := httptrace.WithClientTrace(context.Background(), trace)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+path, nil)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		switch mode {
		case "read+close":
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		case "close":
			resp.Body.Close()
		case "leak":
			// forgot to close!
		}
		fmt.Printf("%-8s %-10s reused=%v\n", path, mode, reused)
		time.Sleep(20 * time.Millisecond) // let the connection settle back into the pool
	}

	get("/issues", "read+close") // new connection
	get("/issues", "read+close") // reuses it
	get("/issues", "close")      // reuses it, then closes without reading
	get("/export", "close")      // still reused! Then closes 1 MiB unread
	get("/issues", "leak")       // new: the big unread body killed the old one
	get("/issues", "read+close") // new: the leaked connection is still busy
}
```

Output:

```
/issues  read+close reused=false
/issues  read+close reused=true
/issues  close      reused=true
/export  close      reused=true
/issues  leak       reused=false
/issues  read+close reused=false
```

Reading it line by line:

1. The first request has to open a connection.
2. Read and closed, so the next request reuses it.
3. Closed **without reading**, and the next request (`/export`) *still* reused it.
4. `/export` closed 1 MiB unread, and the connection was dropped.
5. So the leaking request got a new one, and never closed it...
6. ...which left the last request to open yet another.

## Go 1.27 drains small bodies for you

Line 3 is the new part. In Go 1.27, closing an HTTP/1 response body that you didn't
finish reading makes the transport **read the rest in the background**, up to a
conservative limit (in the current source: 256 KiB, for at most 50 ms). If it gets to
the end, the connection goes back into the pool. Before that, the common advice was to
drain every body yourself:

```go
defer func() {
	io.Copy(io.Discard, resp.Body) // no longer needed for small bodies
	resp.Body.Close()
}()
```

Now a plain `defer resp.Body.Close()` does the right thing for typical API responses,
including error bodies you never read. For big bodies (line 4), Go gives up and closes
the connection rather than downloading megabytes you don't want, which is usually
the right trade.

## The rules, updated

1. **Always close the body.** Nothing replaces this. A leaked body (line 5) is a
   leaked connection.
2. **Read what you need, then close.** Small leftovers are drained for you.
3. **Share one client** (and so one transport) across requests.
4. **Raise `MaxIdleConnsPerHost`** if you make many concurrent requests to one host.

`client.CloseIdleConnections()` closes idle pooled connections, which is handy at the
end of a long-running program or a test.
