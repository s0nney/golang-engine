---
title: Graceful Shutdown
quiz:
  - question: |
      Why does the shutdown code build its deadline from `context.Background()` instead
      of from the signal context `ctx`?

      ```go
      <-ctx.Done() // SIGTERM arrived
      grace, cancel := context.WithTimeout(context.Background(), 10*time.Second)
      defer cancel()
      err := srv.Shutdown(grace)
      ```
    options:
      - text: '`Shutdown` refuses contexts that have a parent'
      - text: '`ctx` is already cancelled, so a timeout derived from it would be over before draining even started'
        correct: true
      - text: '`context.Background()` cancels every in-flight request for you'
      - text: It doesn't matter; both behave the same
    explanation: |
      A child context is cancelled as soon as its parent is. The signal context is done
      by the time you get here, so `context.WithTimeout(ctx, ...)` would give requests
      zero time to finish. Start a fresh deadline from `context.Background()`.
  - question: |
      `main` calls `srv.Shutdown(grace)` in one goroutine while another goroutine is
      blocked in `srv.ListenAndServe()`. What happens to `ListenAndServe`?
    options:
      - text: It keeps running until every request has finished, then returns `nil`
      - text: It returns `http.ErrServerClosed` immediately, while `Shutdown` is still waiting for requests to drain
        correct: true
      - text: It panics
      - text: It returns `context.DeadlineExceeded`
    explanation: |
      The docs are explicit: `Serve` and `ListenAndServe` return `ErrServerClosed` as
      soon as `Shutdown` is called. The draining happens inside `Shutdown`, so the
      program must wait for `Shutdown` to return. If `main` exits when `ListenAndServe`
      returns, in-flight requests die anyway.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"io"
    	"net"
    	"net/http"
    	"time"
    )

    // run serves handler on ln until ctx is cancelled, then shuts down
    // gracefully, giving in-flight requests up to drain to finish.
    func run(ctx context.Context, ln net.Listener, handler http.Handler, drain time.Duration) error {
    	// This version stops abruptly and ignores Serve's errors. Fix it.
    	srv := &http.Server{Handler: handler}
    	go srv.Serve(ln)
    	<-ctx.Done()
    	_ = errors.Join // delete this line once you use the errors package
    	return srv.Close()
    }

    func main() {
    	ln, err := net.Listen("tcp", "127.0.0.1:0")
    	if err != nil {
    		fmt.Println("listen:", err)
    		return
    	}
    	started := make(chan struct{})
    	mux := http.NewServeMux()
    	mux.HandleFunc("POST /api/squeaks", func(w http.ResponseWriter, r *http.Request) {
    		close(started)
    		time.Sleep(300 * time.Millisecond) // a slow database write
    		w.WriteHeader(http.StatusCreated)
    		io.WriteString(w, "squeak saved")
    	})

    	ctx, stop := context.WithCancel(context.Background())
    	result := make(chan error, 1)
    	go func() { result <- run(ctx, ln, mux, 2*time.Second) }()

    	reply := make(chan string, 1)
    	go func() {
    		resp, err := http.Post("http://"+ln.Addr().String()+"/api/squeaks", "application/json", nil)
    		if err != nil {
    			reply <- "request failed: the squeak was lost"
    			return
    		}
    		defer resp.Body.Close()
    		body, _ := io.ReadAll(resp.Body)
    		reply <- fmt.Sprintf("%d %s", resp.StatusCode, body)
    	}()

    	<-started
    	fmt.Println("Pip is posting a squeak... and the deploy sends SIGTERM")
    	stop()
    	fmt.Println("client got:", <-reply)
    	fmt.Println("run returned:", <-result)
    }
  solution: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"io"
    	"net"
    	"net/http"
    	"time"
    )

    // run serves handler on ln until ctx is cancelled, then shuts down
    // gracefully, giving in-flight requests up to drain to finish.
    func run(ctx context.Context, ln net.Listener, handler http.Handler, drain time.Duration) error {
    	srv := &http.Server{
    		Handler:           handler,
    		ReadHeaderTimeout: 5 * time.Second,
    	}
    	done := make(chan error, 1)
    	go func() { done <- srv.Serve(ln) }()

    	select {
    	case err := <-done:
    		return err // Serve failed before we were asked to stop
    	case <-ctx.Done():
    	}

    	grace, cancel := context.WithTimeout(context.Background(), drain)
    	defer cancel()
    	err := srv.Shutdown(grace)
    	if err != nil {
    		srv.Close() // out of time: cut off whatever is left
    	}
    	if serveErr := <-done; !errors.Is(serveErr, http.ErrServerClosed) {
    		return errors.Join(err, serveErr)
    	}
    	return err
    }

    func main() {
    	ln, err := net.Listen("tcp", "127.0.0.1:0")
    	if err != nil {
    		fmt.Println("listen:", err)
    		return
    	}
    	started := make(chan struct{})
    	mux := http.NewServeMux()
    	mux.HandleFunc("POST /api/squeaks", func(w http.ResponseWriter, r *http.Request) {
    		close(started)
    		time.Sleep(300 * time.Millisecond) // a slow database write
    		w.WriteHeader(http.StatusCreated)
    		io.WriteString(w, "squeak saved")
    	})

    	ctx, stop := context.WithCancel(context.Background())
    	result := make(chan error, 1)
    	go func() { result <- run(ctx, ln, mux, 2*time.Second) }()

    	reply := make(chan string, 1)
    	go func() {
    		resp, err := http.Post("http://"+ln.Addr().String()+"/api/squeaks", "application/json", nil)
    		if err != nil {
    			reply <- "request failed: the squeak was lost"
    			return
    		}
    		defer resp.Body.Close()
    		body, _ := io.ReadAll(resp.Body)
    		reply <- fmt.Sprintf("%d %s", resp.StatusCode, body)
    	}()

    	<-started
    	fmt.Println("Pip is posting a squeak... and the deploy sends SIGTERM")
    	stop()
    	fmt.Println("client got:", <-reply)
    	fmt.Println("run returned:", <-result)
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"io"
    	"net"
    	"net/http"
    	"testing"
    	"time"
    )

    func listen(t *testing.T) net.Listener {
    	t.Helper()
    	ln, err := net.Listen("tcp", "127.0.0.1:0")
    	if err != nil {
    		t.Fatalf("listen: %v", err)
    	}
    	return ln
    }

    type result struct {
    	code int
    	body string
    	err  error
    }

    func get(url string) <-chan result {
    	ch := make(chan result, 1)
    	go func() {
    		client := &http.Client{Timeout: 3 * time.Second}
    		resp, err := client.Get(url)
    		if err != nil {
    			ch <- result{err: err}
    			return
    		}
    		defer resp.Body.Close()
    		body, _ := io.ReadAll(resp.Body)
    		ch <- result{code: resp.StatusCode, body: string(body)}
    	}()
    	return ch
    }

    func TestServesUntilCancelled(t *testing.T) {
    	ln := listen(t)
    	ctx, cancel := context.WithCancel(t.Context())
    	defer cancel()
    	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		io.WriteString(w, "squeak!")
    	})
    	done := make(chan error, 1)
    	go func() { done <- run(ctx, ln, handler, time.Second) }()

    	res := <-get("http://" + ln.Addr().String() + "/")
    	if res.err != nil || res.body != "squeak!" {
    		t.Fatalf("GET before shutdown: body %q, error %v; want \"squeak!\"", res.body, res.err)
    	}
    	select {
    	case err := <-done:
    		t.Fatalf("run returned %v before ctx was cancelled; it should keep serving", err)
    	default:
    	}
    	cancel()
    	select {
    	case err := <-done:
    		if err != nil {
    			t.Errorf("run after a clean shutdown returned %v, want nil", err)
    		}
    	case <-time.After(2 * time.Second):
    		t.Fatal("run didn't return within 2s of ctx being cancelled")
    	}
    }

    func TestDrainsInFlightRequest(t *testing.T) {
    	ln := listen(t)
    	ctx, cancel := context.WithCancel(t.Context())
    	defer cancel()
    	started, release := make(chan struct{}), make(chan struct{})
    	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		close(started)
    		<-release
    		io.WriteString(w, "squeak saved")
    	})
    	done := make(chan error, 1)
    	go func() { done <- run(ctx, ln, handler, 2*time.Second) }()

    	reply := get("http://" + ln.Addr().String() + "/")
    	<-started
    	cancel() // shutdown begins while the request is still in the handler

    	select {
    	case err := <-done:
    		close(release)
    		t.Fatalf("run returned %v while a request was still in flight; use srv.Shutdown so it can finish", err)
    	case <-time.After(200 * time.Millisecond):
    	}
    	close(release)

    	res := <-reply
    	if res.err != nil {
    		t.Fatalf("in-flight request failed during shutdown: %v; it should have been allowed to finish", res.err)
    	}
    	if res.code != 200 || res.body != "squeak saved" {
    		t.Errorf("in-flight request got %d %q, want 200 \"squeak saved\"", res.code, res.body)
    	}
    	select {
    	case err := <-done:
    		if err != nil {
    			t.Errorf("run after draining returned %v, want nil", err)
    		}
    	case <-time.After(2 * time.Second):
    		t.Fatal("run didn't return after the last request finished")
    	}
    }

    func TestDrainDeadline(t *testing.T) {
    	ln := listen(t)
    	ctx, cancel := context.WithCancel(t.Context())
    	defer cancel()
    	started := make(chan struct{})
    	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		close(started)
    		<-r.Context().Done() // a handler that never finishes on its own
    	})
    	done := make(chan error, 1)
    	go func() { done <- run(ctx, ln, handler, 100*time.Millisecond) }()

    	get("http://" + ln.Addr().String() + "/")
    	<-started
    	cancel()
    	select {
    	case err := <-done:
    		if !errors.Is(err, context.DeadlineExceeded) {
    			t.Errorf("run with a stuck handler returned %v, want an error wrapping context.DeadlineExceeded", err)
    		}
    	case <-time.After(2 * time.Second):
    		t.Fatal("run with a 100ms drain didn't return within 2s; use a timeout context for Shutdown and Close when it expires")
    	}
    }

    func TestServeError(t *testing.T) {
    	ln := listen(t)
    	ln.Close() // Serve fails straight away on a closed listener
    	ctx, cancel := context.WithCancel(t.Context())
    	defer cancel()
    	done := make(chan error, 1)
    	go func() { done <- run(ctx, ln, http.NotFoundHandler(), time.Second) }()
    	select {
    	case err := <-done:
    		if err == nil || errors.Is(err, http.ErrServerClosed) {
    			t.Errorf("run with a closed listener returned %v, want Serve's error", err)
    		}
    	case <-time.After(time.Second):
    		cancel()
    		<-done
    		t.Fatal("run didn't return when Serve failed; it should report Serve's error without waiting for ctx")
    	}
    }
---

Pip posts a squeak just as you deploy a new version of Squeak. The deploy system sends
the old process a `SIGTERM`, the process exits on the spot, and Pip's request dies
halfway through. Was the squeak saved? Nobody knows, least of all Pip.

A **graceful shutdown** fixes that: stop accepting *new* connections, give the requests
already in progress a bounded amount of time to finish, and only then exit.

## Own the server

`http.ListenAndServe(addr, handler)` gives you no way to stop it. To shut down, you need
your own `http.Server` value, which you met in the timeouts lesson:

```go
srv := &http.Server{
	Addr:              ":8080",
	Handler:           handler,
	ReadHeaderTimeout: 5 * time.Second,
}
```

It has two ways to stop:

- **`srv.Shutdown(ctx)`** closes the listeners, closes idle connections, then *waits*
  for active requests to finish. If `ctx` expires first, it gives up and returns the
  context's error, leaving those connections open.
- **`srv.Close()`** slams everything shut immediately, in-flight requests included.

The usual recipe is "`Shutdown` with a deadline, then `Close` if the deadline passes".

## The shape of a graceful server

```go
func serve(ctx context.Context, srv *http.Server) error {
	done := make(chan error, 1)
	go func() { done <- srv.ListenAndServe() }()

	select {
	case err := <-done:
		return err // it never started, e.g. the port is taken
	case <-ctx.Done():
	}

	grace, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := srv.Shutdown(grace)
	if err != nil {
		srv.Close() // out of time: cut off whatever is left
	}
	if serveErr := <-done; !errors.Is(serveErr, http.ErrServerClosed) {
		return errors.Join(err, serveErr)
	}
	return err
}
```

Walk through it:

1. **Serve in a goroutine**, so the function can wait for two things at once. The
   channel is buffered, so the goroutine can always send its result and exit, even if
   nobody is listening any more.
2. **`select` on both.** If `ListenAndServe` returns first, something is wrong (a
   typical cause is `address already in use`), so return that error right away.
3. **Otherwise `ctx` was cancelled**: time to stop. Build a *fresh* deadline from
   `context.Background()`, because `ctx` is already done.
4. **`Shutdown`, then `Close` if it ran out of time.**
5. **Collect the serve goroutine's result.** After `Shutdown` it's always
   `http.ErrServerClosed`, which means "stopped on purpose", not a failure.

A gotcha worth repeating: `ListenAndServe` returns `ErrServerClosed` the *moment*
`Shutdown` is called, long before draining finishes. A `main` that exits as soon as
`ListenAndServe` returns kills the requests you were trying to save. The waiting happens
in `Shutdown`.

## Connecting signals

Who cancels `ctx`? In a real deployment, the operating system. `signal.NotifyContext`
from `os/signal` turns signals into context cancellation:

```go
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: ":8080", Handler: newRouter(cfg)}
	if err := serve(ctx, srv); err != nil {
		log.Fatal(err)
	}
}
```

`os.Interrupt` is Ctrl+C in a terminal, and `SIGTERM` is what Docker, Kubernetes and
systemd send to ask a process to stop. Calling `stop` restores the default behaviour, so
a *second* Ctrl+C kills the process the old-fashioned way once shutdown has begun.

Keep the signal wiring in `main`. The `serve` function only knows about a context, so
tests can cancel it directly instead of sending signals to the test process.

## What Shutdown doesn't do

- It doesn't wait for **goroutines you started yourself**, like a background job
  sending emails. Give those their own cancellation and a `sync.WaitGroup`.
- It doesn't stop a handler that's stuck. A handler blocked on a slow database should
  pass `r.Context()` along, so the query stops when the connection is finally closed.
- It doesn't touch **hijacked connections** such as WebSockets.
  `srv.RegisterOnShutdown` lets you notify them.
- Your deploy system's patience must be **longer** than your drain deadline. Kubernetes
  waits 30 seconds by default before sending `SIGKILL`. A 60-second drain there would
  never finish.

## Your task

Write `run(ctx, ln, handler, drain)`. It serves on an existing `net.Listener` using
`srv.Serve(ln)` (the tests pass one listening on a random local port), and:

1. builds an `http.Server` with `handler` and a `ReadHeaderTimeout`,
2. returns Serve's error straight away if `Serve` fails before `ctx` is cancelled,
3. when `ctx` is cancelled, calls `Shutdown` with a timeout of `drain` built from
   `context.Background()`,
4. calls `Close` if `Shutdown` returns an error, and returns that error (it wraps
   `context.DeadlineExceeded`),
5. waits for the `Serve` goroutine before returning, and returns `nil` after a clean
   shutdown.

The starter's version calls `Close` straight away. **Run** shows what that does to Pip's
squeak.

Further reading: [http.Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown).
