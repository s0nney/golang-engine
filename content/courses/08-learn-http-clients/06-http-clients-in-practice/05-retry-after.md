---
title: Respecting Retry-After
quiz:
  - question: 'A 503 response says `Retry-After: 30`. Your backoff would wait 200ms. What should `trackr` do?'
    options:
      - text: Wait 200ms, because the client's own backoff always wins
      - text: Wait 30 seconds (or give up if that's longer than it's willing to wait)
        correct: true
      - text: Retry immediately, since the header is only a hint
    explanation: |
      The server knows how long it needs. Retrying earlier just earns another 503 (or
      a ban). If 30 seconds is too long for the user, give up with a clear message
      instead of retrying early.
  - question: 'Which of these is a valid `Retry-After` value?'
    options:
      - text: '`30s`'
      - text: '`1.5`'
      - text: '`Sun, 27 Sep 2026 12:01:30 GMT`'
        correct: true
    explanation: |
      `Retry-After` is either a whole number of seconds (`30`) or an HTTP date. Units
      and fractions aren't allowed.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"time"
    )

    // retryAfter reads the Retry-After header. It returns how long to wait and
    // true, or 0 and false if the header is missing or invalid.
    //
    // The header is either a whole number of seconds ("120") or an HTTP date
    // ("Sun, 27 Sep 2026 12:01:30 GMT"). A date in the past means "retry now" (0).
    // The wait is never more than maxWait.
    func retryAfter(h http.Header, now time.Time, maxWait time.Duration) (time.Duration, bool) {
    	// ?
    	return 0, false
    }

    func main() {
    	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
    	for _, v := range []string{"", "3", "600", "soon", now.Add(90 * time.Second).Format(http.TimeFormat)} {
    		h := http.Header{}
    		if v != "" {
    			h.Set("Retry-After", v)
    		}
    		d, ok := retryAfter(h, now, 5*time.Minute)
    		fmt.Printf("Retry-After %-31q -> %v %v\n", v, d, ok)
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"strconv"
    	"time"
    )

    func retryAfter(h http.Header, now time.Time, maxWait time.Duration) (time.Duration, bool) {
    	v := h.Get("Retry-After")
    	if v == "" {
    		return 0, false
    	}

    	var d time.Duration
    	if secs, err := strconv.Atoi(v); err == nil {
    		if secs < 0 {
    			return 0, false
    		}
    		d = time.Duration(secs) * time.Second
    	} else if t, err := http.ParseTime(v); err == nil {
    		d = max(t.Sub(now), 0)
    	} else {
    		return 0, false
    	}
    	return min(d, maxWait), true
    }

    func main() {
    	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
    	for _, v := range []string{"", "3", "600", "soon", now.Add(90 * time.Second).Format(http.TimeFormat)} {
    		h := http.Header{}
    		if v != "" {
    			h.Set("Retry-After", v)
    		}
    		d, ok := retryAfter(h, now, 5*time.Minute)
    		fmt.Printf("Retry-After %-31q -> %v %v\n", v, d, ok)
    	}
    }
  tests: |
    package main

    import (
    	"net/http"
    	"testing"
    	"time"
    )

    func TestRetryAfter(t *testing.T) {
    	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
    	date := func(d time.Duration) string { return now.Add(d).Format(http.TimeFormat) }

    	tests := []struct {
    		header  string
    		maxWait time.Duration
    		want    time.Duration
    		wantOK  bool
    	}{
    		{"", time.Minute, 0, false},
    		{"0", time.Minute, 0, true},
    		{"7", time.Minute, 7 * time.Second, true},
    		{"120", time.Hour, 2 * time.Minute, true},
    		{"120", 30 * time.Second, 30 * time.Second, true},
    		{date(90 * time.Second), time.Hour, 90 * time.Second, true},
    		{date(2 * time.Hour), 10 * time.Minute, 10 * time.Minute, true},
    		{date(-time.Minute), time.Hour, 0, true},
    		{"-5", time.Minute, 0, false},
    		{"soon", time.Minute, 0, false},
    		{"1.5", time.Minute, 0, false},
    	}
    	for _, tt := range tests {
    		h := http.Header{}
    		if tt.header != "" {
    			h.Set("Retry-After", tt.header)
    		}
    		got, ok := retryAfter(h, now, tt.maxWait)
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("Retry-After: %q (maxWait %v, now %s): got (%v, %v), want (%v, %v)",
    				tt.header, tt.maxWait, now.Format(http.TimeFormat), got, ok, tt.want, tt.wantOK)
    		}
    	}
    }

    func TestRetryAfterFromServer(t *testing.T) {
    	// The header name is case-insensitive on the wire.
    	h := http.Header{}
    	h["Retry-After"] = []string{"4"}
    	if got, ok := retryAfter(h, time.Now(), time.Minute); !ok || got != 4*time.Second {
    		t.Errorf("retryAfter(Retry-After: 4) = (%v, %v), want (4s, true); use h.Get", got, ok)
    	}
    }
---

When a server is overloaded (`503`) or you're over your rate limit (`429`), it often
tells you exactly how long to back off, in a `Retry-After` header:

```
HTTP/1.1 429 Too Many Requests
Retry-After: 30
X-Ratelimit-Remaining: 0
```

A polite client listens. It's also the fastest route back to working requests: many
APIs extend or escalate bans for clients that ignore it.

## Two formats

`Retry-After` comes in one of two forms:

- **Seconds**, a non-negative whole number: `Retry-After: 120`.
- **An HTTP date**: `Retry-After: Sun, 27 Sep 2026 12:01:30 GMT`.

`net/http` can parse the date form with `http.ParseTime`, which accepts all three date
formats HTTP allows. `http.TimeFormat` is the layout for writing one:

```go
t, err := http.ParseTime("Sun, 27 Sep 2026 12:01:30 GMT")
wait := time.Until(t) // same as t.Sub(time.Now())
```

A date in the past just means "go ahead now". Clocks on different machines disagree
a little, so don't treat that as an error.

## Trust, but cap

A buggy or hostile server could send `Retry-After: 86400` (a day). A command-line tool
shouldn't silently hang for a day, so cap the wait. If the cap is lower than what the
server asked for, it's usually better to **give up** with a clear message than to
retry early:

```go
const maxWait = 5 * time.Minute

wait, ok := retryAfter(resp.Header, time.Now(), maxWait)
switch {
case !ok:
	wait = backoff(attempt, base) // no hint: use our own backoff
case wait == maxWait:
	return fmt.Errorf("Trackr asked us to wait at least %v; try again later", wait)
}
```

## Fitting it into the retry loop

`Retry-After` replaces your computed backoff for that one attempt:

```go
delay := backoff(attempt, base)
if resp != nil {
	if d, ok := retryAfter(resp.Header, time.Now(), maxWait); ok {
		delay = d
	}
}
```

You still wait in a `select` on `ctx.Done()`, and you still count the attempt.

## Why pass `now` in?

`retryAfter` takes the current time as a parameter instead of calling `time.Now()`
inside. That makes it a **pure function**: the same input always gives the same output,
so tests can pin `now` to a fixed moment and check exact durations. It's a small design
choice that pays off in chapter 9.

## Your turn

Complete `retryAfter(h, now, maxWait)`:

1. Read the header with `h.Get("Retry-After")`. If it's missing, return `0, false`.
2. If it's a whole number (`strconv.Atoi` succeeds), treat it as seconds. A negative
   number is invalid.
3. Otherwise try `http.ParseTime`. The wait is the time from `now` until then, or `0`
   if that's in the past.
4. Anything else is invalid: return `0, false`.
5. Never return more than `maxWait` (the `min` built-in helps).
