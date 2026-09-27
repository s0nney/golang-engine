---
title: Heartbeat Monitor
difficulty: medium
after: testing-concurrent-code
hints:
  - 'Keep a `map[string]time.Time` with each online courier''s deadline: the time of their last beat plus `timeout`. A beat sets `deadline[courier] = time.Now().Add(timeout)`.'
  - 'You only ever need **one** timer: the one for the earliest deadline. Each time round the loop, find it, and `select` on the beats channel, that timer''s channel and `ctx.Done()`. With no couriers online, use a nil channel so that case never fires.'
  - 'When the timer fires, call `offline` for every courier whose deadline is not after `time.Now()` and delete them from the map, so each silence is reported exactly once.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"time"
    )

    // watch reads courier names from beats until beats is closed or ctx is
    // done. A courier who hasn't sent a beat for timeout is reported with
    // offline(courier), once per silence.
    func watch(ctx context.Context, beats <-chan string, timeout time.Duration, offline func(courier string)) {
    }

    func main() {
    	beats := make(chan string)
    	start := time.Now()
    	done := make(chan struct{})
    	go func() {
    		defer close(done)
    		watch(context.Background(), beats, 50*time.Millisecond, func(c string) {
    			fmt.Printf("%v: %s went offline\n", time.Since(start).Round(10*time.Millisecond), c)
    		})
    	}()

    	send := func(c string) {
    		select {
    		case beats <- c:
    		case <-time.After(time.Second):
    			fmt.Println("watch isn't reading beats")
    		}
    	}
    	send("ana")
    	send("ben")
    	for range 4 {
    		time.Sleep(20 * time.Millisecond)
    		send("ana") // ana keeps pinging; ben goes quiet
    	}
    	time.Sleep(100 * time.Millisecond)
    	close(beats)
    	<-done
    	// want: ben went offline at 50ms, ana at 130ms
    }
  solution: |
    package main

    import (
    	"context"
    	"fmt"
    	"time"
    )

    // watch reads courier names from beats until beats is closed or ctx is
    // done. A courier who hasn't sent a beat for timeout is reported with
    // offline(courier), once per silence.
    func watch(ctx context.Context, beats <-chan string, timeout time.Duration, offline func(courier string)) {
    	deadlines := map[string]time.Time{}
    	for {
    		var timer *time.Timer
    		var expired <-chan time.Time // nil: nobody is online
    		if len(deadlines) > 0 {
    			var next time.Time
    			for _, d := range deadlines {
    				if next.IsZero() || d.Before(next) {
    					next = d
    				}
    			}
    			timer = time.NewTimer(time.Until(next))
    			expired = timer.C
    		}

    		select {
    		case c, ok := <-beats:
    			if !ok {
    				return
    			}
    			deadlines[c] = time.Now().Add(timeout)
    		case now := <-expired:
    			for c, d := range deadlines {
    				if !d.After(now) {
    					delete(deadlines, c)
    					offline(c)
    				}
    			}
    		case <-ctx.Done():
    			return
    		}
    		if timer != nil {
    			timer.Stop()
    		}
    	}
    }

    func main() {
    	beats := make(chan string)
    	start := time.Now()
    	done := make(chan struct{})
    	go func() {
    		defer close(done)
    		watch(context.Background(), beats, 50*time.Millisecond, func(c string) {
    			fmt.Printf("%v: %s went offline\n", time.Since(start).Round(10*time.Millisecond), c)
    		})
    	}()

    	send := func(c string) {
    		select {
    		case beats <- c:
    		case <-time.After(time.Second):
    			fmt.Println("watch isn't reading beats")
    		}
    	}
    	send("ana")
    	send("ben")
    	for range 4 {
    		time.Sleep(20 * time.Millisecond)
    		send("ana") // ana keeps pinging; ben goes quiet
    	}
    	time.Sleep(100 * time.Millisecond)
    	close(beats)
    	<-done
    }
  tests: |
    package main

    import (
    	"context"
    	"fmt"
    	"slices"
    	"strings"
    	"sync"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // monitor runs watch in the background of a synctest bubble.
    type monitor struct {
    	t      *testing.T
    	start  time.Time
    	beats  chan string
    	cancel context.CancelFunc
    	done   chan struct{}

    	mu  sync.Mutex
    	log []string // "35s ben"
    }

    func startMonitor(t *testing.T, timeout time.Duration) *monitor {
    	ctx, cancel := context.WithCancel(t.Context())
    	m := &monitor{t: t, start: time.Now(), beats: make(chan string), cancel: cancel, done: make(chan struct{})}
    	go func() {
    		defer close(m.done)
    		watch(ctx, m.beats, timeout, func(c string) {
    			m.mu.Lock()
    			defer m.mu.Unlock()
    			m.log = append(m.log, fmt.Sprint(time.Since(m.start), " ", c))
    		})
    	}()
    	return m
    }

    // at sleeps until d after the start, then sends a beat from each courier.
    func (m *monitor) at(d time.Duration, couriers ...string) {
    	m.t.Helper()
    	if wait := time.Until(m.start.Add(d)); wait > 0 {
    		synctest.Sleep(wait)
    	}
    	for _, c := range couriers {
    		select {
    		case m.beats <- c:
    		case <-time.After(time.Hour):
    			m.t.Fatalf("at %v: watch didn't take %q's beat within an hour: it must keep reading beats", d, c)
    		}
    	}
    }

    // offlineLog returns what was reported so far, sorted within each instant.
    func (m *monitor) offlineLog() string {
    	m.mu.Lock()
    	defer m.mu.Unlock()
    	log := slices.Clone(m.log)
    	slices.SortStableFunc(log, func(a, b string) int {
    		da, _ := time.ParseDuration(strings.Fields(a)[0])
    		db, _ := time.ParseDuration(strings.Fields(b)[0])
    		if da != db {
    			return int(da - db)
    		}
    		return strings.Compare(a, b)
    	})
    	return strings.Join(log, ", ")
    }

    // stop cancels watch and checks that it returns promptly.
    func (m *monitor) stop() {
    	m.t.Helper()
    	m.cancel()
    	select {
    	case <-m.done:
    	case <-time.After(time.Hour):
    		m.t.Fatalf("watch didn't return within an hour of ctx being cancelled")
    	}
    }

    func TestWatchReportsSilentCouriers(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		m := startMonitor(t, 30*time.Second)
    		defer m.stop()
    		m.at(0, "ana")
    		m.at(5*time.Second, "ben")
    		m.at(10*time.Second, "ana")
    		m.at(20*time.Second, "ana")
    		m.at(22500*time.Millisecond, "cy")
    		synctest.Sleep(time.Hour)
    		want := "35s ben, 50s ana, 52.5s cy"
    		if got := m.offlineLog(); got != want {
    			t.Errorf("timeout 30s; ana beats at 0s, 10s, 20s; ben at 5s; cy at 22.5s:\noffline calls: %s\nwant:          %s", got, want)
    		}
    	})
    }

    func TestWatchCourierComesBack(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		m := startMonitor(t, 10*time.Second)
    		defer m.stop()
    		m.at(0, "ana")
    		m.at(15*time.Second, "ana") // back online after going offline at 10s
    		m.at(20*time.Second, "ana")
    		synctest.Sleep(time.Hour)
    		want := "10s ana, 30s ana"
    		if got := m.offlineLog(); got != want {
    			t.Errorf("timeout 10s; ana beats at 0s, 15s, 20s:\noffline calls: %s\nwant:          %s (once per silence)", got, want)
    		}
    	})
    }

    func TestWatchSameInstant(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		m := startMonitor(t, time.Minute)
    		defer m.stop()
    		m.at(0, "ana", "ben", "cy", "ana")
    		synctest.Sleep(time.Hour)
    		want := "1m0s ana, 1m0s ben, 1m0s cy"
    		if got := m.offlineLog(); got != want {
    			t.Errorf("timeout 1m; ana, ben, cy and ana again beat at 0s:\noffline calls: %s\nwant:          %s", got, want)
    		}
    	})
    }

    func TestWatchManyCouriers(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		m := startMonitor(t, 100*time.Second)
    		defer m.stop()
    		var want []string
    		for i := range 20 {
    			c := fmt.Sprint("c", i)
    			m.at(time.Duration(i)*time.Second, c)
    			if i%2 == 0 {
    				m.at(time.Duration(i)*time.Second+500*time.Millisecond, c)
    				want = append(want, fmt.Sprint(time.Duration(i)*time.Second+100500*time.Millisecond, " ", c))
    			} else {
    				want = append(want, fmt.Sprint(time.Duration(i+100)*time.Second, " ", c))
    			}
    		}
    		synctest.Sleep(time.Hour)
    		if got := m.offlineLog(); got != strings.Join(want, ", ") {
    			t.Errorf("20 couriers, timeout 100s:\noffline calls: %s\nwant:          %s", got, strings.Join(want, ", "))
    		}
    	})
    }

    func TestWatchStopsWhenBeatsClose(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		m := startMonitor(t, 30*time.Second)
    		m.at(0, "ana")
    		m.at(10 * time.Second)
    		close(m.beats)
    		select {
    		case <-m.done:
    		case <-time.After(time.Hour):
    			t.Fatalf("watch didn't return within an hour of beats being closed")
    		}
    		if took := time.Since(m.start); took != 10*time.Second {
    			t.Errorf("beats closed at 10s, watch returned at %v, want 10s", took)
    		}
    		synctest.Sleep(time.Hour)
    		if got := m.offlineLog(); got != "" {
    			t.Errorf("offline calls after watch returned: %s, want none", got)
    		}
    	})
    }

    func TestWatchStopsOnCancel(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		m := startMonitor(t, 30*time.Second)
    		m.at(0, "ana")
    		m.at(10 * time.Second)
    		m.stop()
    		if took := time.Since(m.start); took != 10*time.Second {
    			t.Errorf("ctx cancelled at 10s, watch returned at %v, want 10s", took)
    		}
    		synctest.Sleep(time.Hour)
    		if got := m.offlineLog(); got != "" {
    			t.Errorf("offline calls after watch returned: %s, want none", got)
    		}
    	})
    }
---

Couriers' phones send Dispatchly a **heartbeat** every few seconds. When a
phone goes quiet, perhaps a dead battery or a tunnel, dispatch must stop
offering that courier new orders.

Implement `watch(ctx, beats, timeout, offline)`. It runs until `beats` is
closed or `ctx` is done, then returns:

- Each value from `beats` is a courier's name. A beat puts that courier
  online, or keeps them online.
- When an online courier hasn't sent a beat for `timeout`, call
  `offline(courier)` at exactly that moment, and treat them as offline. Each
  silence is reported **once**; if the courier beats again later, they're back
  online and can go offline again.
- Keep reading `beats` promptly the whole time. Call `offline` from `watch`'s
  own goroutine, and never after `watch` has returned.

## Example

With `timeout = 30s`:

```
t=0s   beat ana
t=5s   beat ben
t=10s  beat ana
t=20s  beat ana
t=35s  offline("ben")   30s after ben's only beat
t=50s  offline("ana")   30s after ana's last beat
```

## Constraints

- Timeouts must be exact, even with beats at fractions of a second: polling
  on a ticker isn't precise enough. Every courier has their own deadline.
- The tests use `synctest` with fake time and compare the exact time of every
  `offline` call. When several couriers go offline at the same instant, any
  order is fine.
