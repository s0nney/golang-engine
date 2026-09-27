---
title: Parallel Delivery
difficulty: hard
after: concurrency-basics
hints:
  - 'Use a fixed pool of workers rather than one goroutine per recipient. Put every **index** `0, 1, 2, ...` into a channel, `close` it, and start `workers` goroutines with `wg.Go`, each ranging over the channel and calling `send` for the recipients it receives.'
  - 'To keep `Failed` in input order no matter which worker finishes first, give each result a home: `errs := make([]error, len(recipients))`. Each index is handled by exactly one worker, so `errs[i] = send(recipients[i])` needs no mutex.'
  - 'After `wg.Wait()`, walk `errs` in order: count the `nil` ones as sent and append the others'' recipients to `Failed`. Don''t forget to treat `workers < 1` as `1`.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strings"
    	"time"
    )

    type Report struct {
    	Sent   int
    	Failed []string
    }

    func deliverAll(recipients []string, workers int, send func(to string) error) Report {
    	return Report{}
    }

    func main() {
    	send := func(to string) error {
    		time.Sleep(100 * time.Millisecond) // talking to the carrier is slow
    		if strings.HasSuffix(to, "0000") {
    			return errors.New("carrier rejected " + to)
    		}
    		return nil
    	}
    	recipients := []string{"555-0101", "555-0000", "555-0102", "555-0103", "555-0104", "555-0105"}

    	start := time.Now()
    	report := deliverAll(recipients, 3, send)
    	fmt.Printf("%+v in %v\n", report, time.Since(start).Round(100*time.Millisecond))
    	// want: {Sent:5 Failed:[555-0000]} in 200ms
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strings"
    	"sync"
    	"time"
    )

    type Report struct {
    	Sent   int
    	Failed []string
    }

    func deliverAll(recipients []string, workers int, send func(to string) error) Report {
    	workers = max(workers, 1)
    	jobs := make(chan int, len(recipients))
    	for i := range recipients {
    		jobs <- i
    	}
    	close(jobs)

    	errs := make([]error, len(recipients))
    	var wg sync.WaitGroup
    	for range workers {
    		wg.Go(func() {
    			for i := range jobs {
    				errs[i] = send(recipients[i])
    			}
    		})
    	}
    	wg.Wait()

    	var report Report
    	for i, err := range errs {
    		if err != nil {
    			report.Failed = append(report.Failed, recipients[i])
    		} else {
    			report.Sent++
    		}
    	}
    	return report
    }

    func main() {
    	send := func(to string) error {
    		time.Sleep(100 * time.Millisecond) // talking to the carrier is slow
    		if strings.HasSuffix(to, "0000") {
    			return errors.New("carrier rejected " + to)
    		}
    		return nil
    	}
    	recipients := []string{"555-0101", "555-0000", "555-0102", "555-0103", "555-0104", "555-0105"}

    	start := time.Now()
    	report := deliverAll(recipients, 3, send)
    	fmt.Printf("%+v in %v\n", report, time.Since(start).Round(100*time.Millisecond))
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"slices"
    	"strings"
    	"sync"
    	"testing"
    	"time"
    )

    // carrier is a fake carrier that records how it was used.
    type carrier struct {
    	mu       sync.Mutex
    	delay    time.Duration
    	calls    map[string]int
    	inFlight int
    	peak     int
    }

    func newCarrier(delay time.Duration) *carrier {
    	return &carrier{delay: delay, calls: map[string]int{}}
    }

    func (c *carrier) send(to string) error {
    	c.mu.Lock()
    	c.calls[to]++
    	c.inFlight++
    	c.peak = max(c.peak, c.inFlight)
    	c.mu.Unlock()

    	time.Sleep(c.delay)

    	c.mu.Lock()
    	c.inFlight--
    	c.mu.Unlock()
    	if strings.Contains(to, "bad") {
    		return errors.New("rejected " + to)
    	}
    	return nil
    }

    func run(t *testing.T, recipients []string, workers int, c *carrier) Report {
    	t.Helper()
    	done := make(chan Report, 1)
    	go func() { done <- deliverAll(recipients, workers, c.send) }()
    	select {
    	case r := <-done:
    		return r
    	case <-time.After(3 * time.Second):
    		t.Fatalf("deliverAll(%d recipients, %d workers) didn't return within 3 seconds: is a goroutine stuck waiting on a channel?", len(recipients), workers)
    		return Report{}
    	}
    }

    func TestDeliverAllResults(t *testing.T) {
    	tests := []struct {
    		name       string
    		recipients []string
    		workers    int
    		want       Report
    	}{
    		{"none", nil, 4, Report{0, nil}},
    		{"all good", []string{"a", "b", "c"}, 2, Report{3, nil}},
    		{"failures keep input order", []string{"bad-3", "ok-1", "bad-1", "ok-2", "bad-2"}, 5, Report{2, []string{"bad-3", "bad-1", "bad-2"}}},
    		{"all fail", []string{"bad-a", "bad-b"}, 1, Report{0, []string{"bad-a", "bad-b"}}},
    		{"more workers than recipients", []string{"x", "bad-y"}, 50, Report{1, []string{"bad-y"}}},
    		{"zero workers means one", []string{"x", "bad-y", "z"}, 0, Report{2, []string{"bad-y"}}},
    		{"negative workers means one", []string{"x"}, -3, Report{1, nil}},
    		{"duplicates are sent twice", []string{"bad-dup", "ok", "bad-dup"}, 3, Report{1, []string{"bad-dup", "bad-dup"}}},
    		{"unicode", []string{"Zoë 🐝", "bad-🎉"}, 2, Report{1, []string{"bad-🎉"}}},
    	}
    	for _, tt := range tests {
    		c := newCarrier(time.Millisecond)
    		got := run(t, tt.recipients, tt.workers, c)
    		if got.Sent != tt.want.Sent || !slices.Equal(got.Failed, tt.want.Failed) {
    			t.Errorf("%s: deliverAll(%q, %d) = %+v, want %+v", tt.name, tt.recipients, tt.workers, got, tt.want)
    		}
    		wantCalls := map[string]int{}
    		for _, r := range tt.recipients {
    			wantCalls[r]++
    		}
    		for r, n := range wantCalls {
    			if c.calls[r] != n {
    				t.Errorf("%s: send(%q) was called %d times, want %d (once per entry in recipients)", tt.name, r, c.calls[r], n)
    			}
    		}
    	}
    }

    func TestDeliverAllRespectsWorkerLimit(t *testing.T) {
    	for _, workers := range []int{1, 3, 8} {
    		recipients := make([]string, 40)
    		for i := range recipients {
    			recipients[i] = fmt.Sprintf("555-%04d", i)
    		}
    		c := newCarrier(10 * time.Millisecond)
    		run(t, recipients, workers, c)
    		if c.peak > workers {
    			t.Errorf("deliverAll(40 recipients, %d workers) had %d sends running at once, want at most %d", workers, c.peak, workers)
    		}
    		if workers > 1 && c.peak < 2 {
    			t.Errorf("deliverAll(40 recipients, %d workers) only ever ran %d send at a time: the workers should send in parallel", workers, c.peak)
    		}
    	}
    }

    func TestDeliverAllIsParallel(t *testing.T) {
    	recipients := make([]string, 40)
    	for i := range recipients {
    		recipients[i] = fmt.Sprintf("555-%04d", i)
    	}
    	c := newCarrier(50 * time.Millisecond)
    	start := time.Now()
    	got := run(t, recipients, 10, c)
    	if d := time.Since(start); d > time.Second {
    		t.Errorf("deliverAll(40 recipients taking 50ms each, 10 workers) took %v, want about 200ms: send in parallel", d)
    	}
    	if got.Sent != 40 {
    		t.Errorf("deliverAll(40 good recipients, 10 workers).Sent = %d, want 40", got.Sent)
    	}
    }
---

Sending one message means a round trip to a phone carrier, which is slow.
Sending a 10,000-person campaign one message at a time would take forever,
but firing off 10,000 requests at once would get Textio blocked. The answer
is a **worker pool**: a fixed number of goroutines sharing the work.

Write `deliverAll(recipients, workers, send) Report`:

- Call `send(to)` exactly once for **every entry** in `recipients`
  (a number listed twice is sent twice).
- Never have more than `workers` calls to `send` running at the same time,
  but do run them in parallel. If `workers` is less than 1, use 1.
- Return a `Report` with the number of sends that returned a `nil` error in
  `Sent`, and the recipients whose send failed in `Failed`, in the **same
  order as in `recipients`**, however the goroutines happened to finish.

`send` is safe to call from many goroutines at once.

## Example

```go
recipients := []string{"555-0101", "555-0000", "555-0102", "555-0103", "555-0104", "555-0105"}
deliverAll(recipients, 3, send)
// {Sent:5 Failed:[555-0000]}
```

If each send takes 100ms, three workers finish the six sends in about
200ms instead of 600ms.

## Constraints

- One test sends to 40 recipients whose sends take 50ms each, with 10 workers,
  and allows at most one second (one at a time would take 2 seconds).
- The tests count how many sends run at once, so starting one goroutine per
  recipient isn't allowed.
- When nothing fails, `Failed` may be `nil` or empty.
