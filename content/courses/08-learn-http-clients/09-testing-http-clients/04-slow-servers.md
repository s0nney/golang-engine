---
title: Slow Servers and Timeouts
quiz:
  - question: |
      This test hangs until the test binary's timeout kills it. Why?

      ```go
      block := make(chan struct{})
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          <-block
      }))
      defer srv.Close()
      // ... client call that times out after 50ms ...
      ```
    options:
      - text: The client timeout doesn't work against httptest servers
      - text: '`srv.Close()` waits for running handlers, and this one waits forever on `block`, which nobody closes'
        correct: true
      - text: Channels can't be used in handlers
    explanation: |
      The client gives up, but the handler is still blocked, and `Close` waits for it.
      Wait on `r.Context().Done()` instead (it fires when the client disconnects), or
      `defer close(block)` *after* `defer srv.Close()` so it runs first.
  - question: Inside a `synctest.Test` bubble, a handler calls `time.Sleep(time.Minute)` and the client has a 30-second timeout. How long does the test take in real time?
    options:
      - text: About 30 seconds
      - text: About a minute
      - text: Almost no time, because the bubble's fake clock jumps ahead whenever every goroutine in it is blocked
        correct: true
    explanation: |
      In a bubble, time only advances when every goroutine is durably blocked, and then
      it jumps straight to the next timer. The test sees exactly 30 seconds pass, in a
      few microseconds of real time.
---

`trackr`'s timeouts are some of its most important code, and the easiest to leave
untested, because a test that waits 30 seconds for a timeout is a test nobody runs.
There are two good ways to make these tests fast.

## Option 1: tiny real timeouts

Make the fake server hang, and give the client a very short deadline:

```go
func TestGetIssueRespectsDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // hang until the client gives up
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, HTTPClient: srv.Client()}
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := c.GetIssue(ctx, 1)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("GetIssue error = %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("GetIssue took %v; it should give up after about 50ms", elapsed)
	}
}
```

Two details make it robust:

- **The handler waits on `r.Context().Done()`**, which fires when the client
  disconnects. Don't use `time.Sleep(10 * time.Second)` in a handler. `srv.Close()`
  waits for handlers to finish, so the test would take 10 seconds anyway.
- **The upper bound is generous.** Checking `elapsed < 60ms` would fail on a busy CI
  machine. Check that it gave up "reasonably soon" (well under the server's hang), not
  that it hit an exact time.

This works, but every such test still costs real milliseconds, and the upper bound is
fuzzy by necessity.

## Option 2: fake time with synctest

`testing/synctest` (Go 1.25) runs a function in a **bubble** with its own fake clock.
Inside the bubble, time stands still while any goroutine can make progress. When
**every** goroutine in the bubble is blocked (on a sleep, a timer, a channel), the
clock jumps straight to the next timer that would fire.

Go 1.27's `httptest.NewTestServer` runs its in-memory network inside the bubble too,
so a whole HTTP exchange can run on fake time:

```go
func TestClientTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Minute) // a very slow server, in fake time
		}))
		client := srv.Client()
		client.Timeout = 30 * time.Second
		c := &Client{BaseURL: "https://api.trackr.dev", HTTPClient: client}

		start := time.Now()
		_, err := c.GetIssue(t.Context(), 1)
		if err == nil {
			t.Fatal("GetIssue succeeded against a server that never answers")
		}
		if got := time.Since(start); got != 30*time.Second {
			t.Errorf("gave up after %v, want exactly 30s", got)
		}
	})
}
```

This test checks for **exactly** 30 seconds, and it finishes in about 4 ms of real
time. The handler's minute-long sleep is fine here, because the fake clock skips
through it once everything else is blocked.

A few rules of the bubble:

- Use `synctest.Test(t, func(t *testing.T) { ... })` and the inner `t`.
- Create the server, the client and any contexts **inside** the bubble. A
  `NewServer` listening on a real port involves real network I/O, which the bubble
  can't wait on, so use `NewTestServer`.
- `synctest.Wait()` blocks until every other goroutine in the bubble is blocked.
  That's handy for asserting "nothing happened yet" before advancing time with
  `time.Sleep`, or with `synctest.Sleep` (Go 1.27), which does both in one call.

## Which to use?

Use real short timeouts when the code under test uses a real network or other things
outside Go's control. Use synctest whenever you can. It's exact, deterministic and
instant, and in the next lesson it makes testing retries with backoff almost boring.

## Further reading

- The Go blog, "Testing concurrent code with testing/synctest": https://go.dev/blog/synctest
