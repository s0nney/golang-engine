---
title: Rate Limiting
quiz:
  - question: 'A limiter allows 10 requests per second with a burst of 3. `trackr` fires 6 requests at once, starting with a full bucket. Roughly when does the 6th one go out?'
    options:
      - text: Immediately
      - text: After about 300ms
        correct: true
      - text: After about 600ms
      - text: After 1 second
    explanation: |
      Three tokens are waiting, so requests 1 to 3 go at once. After that a new token
      arrives every 100ms, so requests 4, 5 and 6 go at about 100, 200 and 300ms.
  - question: What's the difference between limiting **rate** and limiting **concurrency**?
    options:
      - text: They're the same thing
      - text: Rate limits how many requests *start* per second, and concurrency limits how many are *in flight* at once
        correct: true
      - text: Rate limiting is done by the server, and concurrency limiting by the client
    explanation: |
      Ten slow requests per second can still pile up to hundreds in flight. APIs often
      limit both, so clients sometimes need both: a limiter and a semaphore.
---

Every serious API limits how fast you can call it, say 10 requests per second or 5,000
per hour per token. Go over and you get `429 Too Many Requests`. Retrying 429s works,
but it's better not to trip the limit in the first place. That's **client-side rate
limiting**.

## Read the budget

Many APIs tell you where you stand on every response. The names vary, but they
typically look like this:

```
X-Ratelimit-Limit: 5000
X-Ratelimit-Remaining: 4987
X-Ratelimit-Reset: 1790503200
```

`Remaining` is how many calls you have left in the window, and `Reset` is when the
window restarts (here as Unix seconds). A client doing a big import can watch
`Remaining` and slow down before it hits zero:

```go
if rem, err := strconv.Atoi(resp.Header.Get("X-Ratelimit-Remaining")); err == nil && rem < 10 {
	log.Printf("only %d API calls left in this window; slowing down", rem)
}
```

## The token bucket

The classic algorithm is a **token bucket**. The bucket holds up to *burst* tokens and
gets a new one every *interval*. Each request takes a token, waiting if the bucket is
empty. You get short bursts, and a steady rate over time. Go's channels make a tidy
bucket: a buffered channel holds the tokens, and a ticker refills it.

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

// Limiter allows one event per interval, with bursts of up to burst events.
type Limiter struct {
	tokens chan struct{}
	ticker *time.Ticker
}

func NewLimiter(interval time.Duration, burst int) *Limiter {
	l := &Limiter{
		tokens: make(chan struct{}, burst),
		ticker: time.NewTicker(interval),
	}
	for range burst {
		l.tokens <- struct{}{} // start with a full bucket
	}
	go func() {
		for range l.ticker.C {
			select {
			case l.tokens <- struct{}{}: // add a token
			default: // bucket full: drop it
			}
		}
	}()
	return l
}

// Wait blocks until a token is available or ctx is done.
func (l *Limiter) Wait(ctx context.Context) error {
	select {
	case <-l.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	lim := NewLimiter(100*time.Millisecond, 3) // 10 per second, bursts of 3
	ctx := context.Background()
	start := time.Now()
	for i := range 6 {
		if err := lim.Wait(ctx); err != nil {
			fmt.Println(err)
			return
		}
		resp, err := srv.Client().Get(srv.URL)
		if err != nil {
			fmt.Println(err)
			return
		}
		resp.Body.Close()
		fmt.Printf("request %d at %v\n", i+1, time.Since(start).Round(100*time.Millisecond))
	}
}
```

Output:

```
request 1 at 0s
request 2 at 0s
request 3 at 0s
request 4 at 100ms
request 5 at 200ms
request 6 at 300ms
```

The first three spend the burst, and then it's one every 100ms. Notice that `Wait`
takes a context, so a cancelled command doesn't sit waiting for a token.

This toy version never stops its ticker goroutine. A real one needs a `Stop` method
(stop the ticker and signal the goroutine to exit through a `done` channel). In
production code most people use `golang.org/x/time/rate`, a well-tested limiter from
the Go team's extended libraries. Its `rate.NewLimiter(10, 3)` and
`lim.Wait(ctx)` work just like the one above. It isn't part of the standard library,
so you'd add it with `go get`.

## Where the limiter goes

Put it in **one** place that every request passes through, the same place that adds
the auth header, so no call can skip it:

```go
func (c *Client) do(req *http.Request) (*http.Response, error) {
	if err := c.limiter.Wait(req.Context()); err != nil {
		return nil, err
	}
	return c.http.Do(req)
}
```

Share one limiter across all goroutines. A limiter per goroutine would multiply your
rate.

## Limiting concurrency

Some APIs also cap how many requests may be **in flight** at once. A buffered channel
used as a semaphore handles that:

```go
sem := make(chan struct{}, 4) // at most 4 at a time

var wg sync.WaitGroup
for _, id := range issueIDs {
	wg.Go(func() {
		sem <- struct{}{}        // acquire
		defer func() { <-sem }() // release
		fetchIssue(ctx, id)
	})
}
wg.Wait()
```

`WaitGroup.Go` (Go 1.25) starts the goroutine and does the `Add`/`Done` bookkeeping for
you.

## Rate limits and retries together

The complete, polite request path for `trackr` is now:

1. wait for the rate limiter;
2. send the request;
3. on 429 or 503, wait for `Retry-After` (or backoff with jitter) and go back to 1;
4. give up after `maxAttempts` or when the context ends.

Each piece is small. Together they make a client that behaves well even when the API
is having a bad day.
