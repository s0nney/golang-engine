---
title: Always Close the Body
quiz:
  - question: |
      What's the bug?

      ```go
      resp, err := http.Get(url)
      defer resp.Body.Close()
      if err != nil {
          return err
      }
      ```
    options:
      - text: There's no bug
      - text: If `err` is non-nil, `resp` is nil, so the deferred `resp.Body.Close()` panics
        correct: true
      - text: '`defer` runs too early and closes the body before you read it'
      - text: You should call `resp.Close()`, not `resp.Body.Close()`
    explanation: |
      When `http.Get` fails, `resp` is nil. The `defer` evaluates `resp.Body` right away,
      which dereferences a nil pointer. Check `err` *first*, then defer the close.
  - question: Your function returns an error for a 500 status *before* the `defer resp.Body.Close()` line. What happens?
    options:
      - text: Nothing bad, since error responses have no body
      - text: Go closes it for you when `resp` goes out of scope
      - text: The body is never closed, so that connection can't go back into the pool for reuse
        correct: true
    explanation: |
      Go has no destructors. An unclosed body keeps its connection busy, so the next
      request has to open a brand new one. Put the `defer` straight after the error check,
      before any early returns.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    // fetchMOTD fetches baseURL + "/motd" (the Trackr "message of the day").
    // It returns the body with surrounding whitespace trimmed, or an error if
    // the request fails or the status isn't 200 OK.
    func fetchMOTD(baseURL string) (string, error) {
    	// ?
    	return "", nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		if r.URL.Path != "/motd" {
    			http.NotFound(w, r)
    			return
    		}
    		fmt.Fprintln(w, "  Sprint 42 ends Friday. Ship it!  ")
    	}))
    	defer srv.Close()

    	motd, err := fetchMOTD(srv.URL)
    	fmt.Printf("motd: %q, err: %v\n", motd, err)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    )

    func fetchMOTD(baseURL string) (string, error) {
    	resp, err := http.Get(baseURL + "/motd")
    	if err != nil {
    		return "", err
    	}
    	defer resp.Body.Close()

    	if resp.StatusCode != http.StatusOK {
    		return "", fmt.Errorf("fetching motd: unexpected status %s", resp.Status)
    	}

    	body, err := io.ReadAll(resp.Body)
    	if err != nil {
    		return "", err
    	}
    	return strings.TrimSpace(string(body)), nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		if r.URL.Path != "/motd" {
    			http.NotFound(w, r)
    			return
    		}
    		fmt.Fprintln(w, "  Sprint 42 ends Friday. Ship it!  ")
    	}))
    	defer srv.Close()

    	motd, err := fetchMOTD(srv.URL)
    	fmt.Printf("motd: %q, err: %v\n", motd, err)
    }
  tests: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"testing"
    )

    func TestFetchMOTD(t *testing.T) {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		if r.Method != http.MethodGet {
    			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    			return
    		}
    		if r.URL.Path != "/motd" {
    			http.NotFound(w, r)
    			return
    		}
    		fmt.Fprint(w, "\n\t Standup moved to 10:30 \n")
    	}))
    	defer srv.Close()

    	got, err := fetchMOTD(srv.URL)
    	if err != nil {
    		t.Fatalf("fetchMOTD(server) returned error %v, want nil", err)
    	}
    	if want := "Standup moved to 10:30"; got != want {
    		t.Errorf("fetchMOTD(server) = %q, want %q (did you trim the whitespace?)", got, want)
    	}
    }

    func TestFetchMOTDBadStatus(t *testing.T) {
    	for _, code := range []int{http.StatusNotFound, http.StatusInternalServerError, http.StatusNoContent} {
    		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    			w.WriteHeader(code)
    			fmt.Fprint(w, "nope")
    		}))
    		got, err := fetchMOTD(srv.URL)
    		srv.Close()
    		if err == nil {
    			t.Errorf("server replied %d: fetchMOTD returned (%q, nil), want an error for any status other than 200", code, got)
    		}
    	}
    }

    func TestFetchMOTDNetworkError(t *testing.T) {
    	srv := httptest.NewServer(http.NotFoundHandler())
    	url := srv.URL
    	srv.Close() // nothing is listening any more

    	if _, err := fetchMOTD(url); err == nil {
    		t.Error("fetchMOTD on a closed server returned a nil error, want the error from http.Get")
    	}
    }
---

In the last lesson you wrote `defer resp.Body.Close()`. That line matters more than it
looks.

## What the body really is

`resp.Body` isn't a string sitting in memory. It's a stream reading from a live
network connection. Until you close it, that connection is **tied up**. It can't be
reused for your next request, and its buffers and goroutine stay alive.

If `trackr` fetches 1,000 issues one by one and forgets to close each body, it can
end up holding 1,000 open connections. Eventually you hit the operating system's file
limit and get errors like `socket: too many open files`.

## The rule

> Whenever `err == nil`, you own `resp.Body`, and you must close it, **even if you
> don't read it and even if the status is an error**.

The idiomatic pattern is:

```go
resp, err := http.Get(url)
if err != nil {
	return err // no response, nothing to close
}
defer resp.Body.Close() // right after the check, before any other return

if resp.StatusCode != http.StatusOK {
	return fmt.Errorf("unexpected status %s", resp.Status) // body still gets closed
}
```

Two details to notice:

1. **Check `err` first.** When there's an error, `resp` is nil, and `resp.Body` would panic.
2. **Defer immediately.** Every `return` below that line, including the error returns,
   now closes the body for you.

## Reading and closing are different jobs

`io.ReadAll` reads the body to the end but does **not** close it. `Close` releases the
connection but doesn't read anything. You usually want both: read what you need, and
always close.

```go
body, err := io.ReadAll(resp.Body)
```

## Closed means reusable

When a body is read to the end *and* closed, `net/http` puts the connection back in a
pool, and your next request to the same host skips the whole connection setup. That's a
big speed-up. You'll measure it in chapter 6.

## Your turn

The Trackr API has a "message of the day" at `/motd` that `trackr` prints at startup.
Complete `fetchMOTD` so that it:

1. sends a `GET` to `baseURL + "/motd"` with `http.Get`;
2. returns the `http.Get` error if there is one;
3. closes the body on every path after a successful `http.Get`;
4. returns an error if the status is anything other than `200 OK`;
5. otherwise returns the body as a string with leading and trailing whitespace removed
   (`strings.TrimSpace`).

The hidden tests start their own fake API and pass you its URL, so don't hard-code an address.
