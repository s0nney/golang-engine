---
title: Merge Courier Feeds
difficulty: medium
after: concurrency-patterns
hints:
  - 'Start one forwarding goroutine per feed, all sending on the same output channel. A `sync.WaitGroup` counts them, and one extra goroutine does `wg.Wait()` then `close(out)`, so the output closes only after the last forwarder has stopped sending.'
  - 'A forwarder can get stuck in two places: waiting for its feed, and waiting for the consumer to take a value. Both waits need a `select` with `<-ctx.Done()`.'
  - 'Receive with `u, ok := <-feed` inside the `select`, and return when `ok` is false. Then send with another `select { case out <- u: case <-ctx.Done(): return }`.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"time"
    )

    // Update is a location ping from a courier's phone.
    type Update struct {
    	Courier string
    	Seq     int
    }

    // mergeFeeds forwards every Update from every feed onto the returned
    // channel, which is closed once all feeds are closed or ctx is done.
    func mergeFeeds(ctx context.Context, feeds ...<-chan Update) <-chan Update {
    	out := make(chan Update)
    	close(out)
    	return out
    }

    // phone sends n updates from courier, one every gap, then closes.
    func phone(courier string, n int, gap time.Duration) <-chan Update {
    	ch := make(chan Update)
    	go func() {
    		defer close(ch)
    		for i := range n {
    			time.Sleep(gap)
    			ch <- Update{courier, i + 1}
    		}
    	}()
    	return ch
    }

    func main() {
    	ctx := context.Background()
    	count := 0
    	for u := range mergeFeeds(ctx, phone("ana", 3, 10*time.Millisecond), phone("ben", 2, 15*time.Millisecond)) {
    		fmt.Printf("%s #%d\n", u.Courier, u.Seq)
    		count++
    	}
    	fmt.Println(count, "updates") // want: 5 updates, ana's and ben's interleaved
    }
  solution: |
    package main

    import (
    	"context"
    	"fmt"
    	"sync"
    	"time"
    )

    // Update is a location ping from a courier's phone.
    type Update struct {
    	Courier string
    	Seq     int
    }

    // mergeFeeds forwards every Update from every feed onto the returned
    // channel, which is closed once all feeds are closed or ctx is done.
    func mergeFeeds(ctx context.Context, feeds ...<-chan Update) <-chan Update {
    	out := make(chan Update)
    	var wg sync.WaitGroup
    	for _, feed := range feeds {
    		wg.Go(func() {
    			for {
    				select {
    				case u, ok := <-feed:
    					if !ok {
    						return
    					}
    					select {
    					case out <- u:
    					case <-ctx.Done():
    						return
    					}
    				case <-ctx.Done():
    					return
    				}
    			}
    		})
    	}
    	go func() {
    		wg.Wait()
    		close(out)
    	}()
    	return out
    }

    // phone sends n updates from courier, one every gap, then closes.
    func phone(courier string, n int, gap time.Duration) <-chan Update {
    	ch := make(chan Update)
    	go func() {
    		defer close(ch)
    		for i := range n {
    			time.Sleep(gap)
    			ch <- Update{courier, i + 1}
    		}
    	}()
    	return ch
    }

    func main() {
    	ctx := context.Background()
    	count := 0
    	for u := range mergeFeeds(ctx, phone("ana", 3, 10*time.Millisecond), phone("ben", 2, 15*time.Millisecond)) {
    		fmt.Printf("%s #%d\n", u.Courier, u.Seq)
    		count++
    	}
    	fmt.Println(count, "updates")
    }
  tests: |
    package main

    import (
    	"context"
    	"fmt"
    	"maps"
    	"slices"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // closedFeed returns a closed channel holding n updates from courier.
    func closedFeed(courier string, n int) <-chan Update {
    	ch := make(chan Update, n)
    	for i := range n {
    		ch <- Update{courier, i + 1}
    	}
    	close(ch)
    	return ch
    }

    // next receives one value from out, failing after an hour of fake time.
    func next(t *testing.T, out <-chan Update, what string) (Update, bool) {
    	t.Helper()
    	select {
    	case u, ok := <-out:
    		return u, ok
    	case <-time.After(time.Hour):
    		t.Fatalf("%s: nothing arrived on the output within an hour", what)
    		return Update{}, false
    	}
    }

    // drain reads out until it's closed, grouping updates per courier.
    func drain(t *testing.T, out <-chan Update, what string) map[string][]int {
    	t.Helper()
    	got := map[string][]int{}
    	for {
    		u, ok := next(t, out, what+" (is the output closed once every feed is done?)")
    		if !ok {
    			return got
    		}
    		got[u.Courier] = append(got[u.Courier], u.Seq)
    	}
    }

    func seqs(n int) []int {
    	var s []int
    	for i := range n {
    		s = append(s, i+1)
    	}
    	return s
    }

    func TestMergeFeedsForwardsEverything(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		out := mergeFeeds(t.Context(), closedFeed("ana", 3), closedFeed("ben", 1), closedFeed("cy", 0), closedFeed("dee", 50))
    		got := drain(t, out, "4 feeds")
    		want := map[string][]int{"ana": seqs(3), "ben": seqs(1), "dee": seqs(50)}
    		if !maps.EqualFunc(got, want, slices.Equal) {
    			t.Errorf("merging feeds of 3, 1, 0 and 50 updates: got %v, want %v (every update once, each feed in order)", got, want)
    		}
    	})
    }

    func TestMergeNoFeeds(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		if u, ok := next(t, mergeFeeds(t.Context()), "no feeds"); ok {
    			t.Errorf("mergeFeeds() with no feeds sent %v, want the output closed", u)
    		}
    	})
    }

    func TestMergeDoesNotWaitForSlowFeeds(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		quiet := make(chan Update) // sends nothing for a while
    		out := mergeFeeds(t.Context(), quiet, closedFeed("ben", 3))
    		start := time.Now()
    		for i := range 3 {
    			u, ok := next(t, out, "ben's updates while ana's feed is quiet")
    			if !ok || u != (Update{"ben", i + 1}) {
    				t.Fatalf("update %d = %v, %v, want {ben %d}, true", i+1, u, ok, i+1)
    			}
    		}
    		if took := time.Since(start); took != 0 {
    			t.Errorf("ben's 3 updates took %v to arrive while ana's feed was quiet, want 0s: forward every feed at the same time", took)
    		}
    		time.Sleep(time.Minute)
    		select {
    		case quiet <- Update{"ana", 1}:
    		case <-time.After(time.Hour):
    			t.Fatalf("mergeFeeds never read from ana's feed")
    		}
    		if u, _ := next(t, out, "ana's late update"); u != (Update{"ana", 1}) {
    			t.Errorf("after ana's quiet feed finally sent, got %v, want {ana 1}", u)
    		}
    		close(quiet)
    		if u, ok := next(t, out, "after every feed closed"); ok {
    			t.Errorf("got %v after every feed closed, want the output closed", u)
    		}
    	})
    }

    func TestMergeCancelStopsIdleFeeds(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		ctx, cancel := context.WithCancel(t.Context())
    		never := make(chan Update) // never sends, never closes
    		out := mergeFeeds(ctx, never, closedFeed("ben", 2))
    		for range 2 {
    			next(t, out, "ben's updates")
    		}
    		synctest.Sleep(time.Minute)
    		cancel()
    		if u, ok := next(t, out, "after cancel with a feed that never closes"); ok {
    			t.Errorf("after cancel, got %v, want the output closed", u)
    		}
    	})
    }

    func TestMergeCancelWithAbandonedConsumer(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		ctx, cancel := context.WithCancel(t.Context())
    		var feeds []<-chan Update
    		for i := range 5 {
    			feeds = append(feeds, closedFeed(fmt.Sprint("c", i), 10))
    		}
    		out := mergeFeeds(ctx, feeds...)
    		next(t, out, "first update")
    		synctest.Sleep(time.Minute) // the consumer walks away...
    		cancel()                    // ...and cancels
    		synctest.Wait()

    		// Every forwarder has seen the cancellation by now, so nothing more
    		// may be delivered: the output must simply be closed.
    		if u, ok := next(t, out, "after cancel with an abandoned consumer"); ok {
    			t.Errorf("after cancel, the output delivered %v, want it closed: forwarders blocked sending to a consumer that stopped reading must give up when ctx is done", u)
    		}
    	})
    }
---

Every courier's phone streams location **updates** on its own channel. The
live map wants a single stream, so Dispatchly fans them in.

Implement `mergeFeeds(ctx, feeds...)`. It returns a channel that:

- Carries every `Update` from every feed, as soon as it arrives. A quiet feed
  must never hold up the others, and each feed's updates keep their order.
- Is **closed** once every feed has been closed (straight away if there are no
  feeds).
- Is closed promptly when `ctx` is done, even if some feeds never close and
  even if nobody is reading the output any more. Once `ctx` is done, an
  update that hasn't been handed over yet is dropped, not delivered.

Every goroutine `mergeFeeds` starts must have exited by the time the output is
closed, so nothing is left running and nothing sends on a closed channel.

## Example

```go
out := mergeFeeds(ctx, anaFeed, benFeed)
for u := range out {
	fmt.Println(u.Courier, u.Seq) // ana 1, ben 1, ana 2, ... in arrival order
}
// the loop ends once both feeds are closed
```

## Constraints

- Up to hundreds of feeds, each with any number of updates.
- The tests run in a `synctest` bubble: a stuck forwarder shows up as an
  output that's never closed, and a goroutine left blocked fails the test.
