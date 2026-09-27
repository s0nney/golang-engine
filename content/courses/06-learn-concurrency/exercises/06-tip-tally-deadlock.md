---
title: Tip Tally Deadlock
difficulty: easy
after: race-conditions-and-deadlocks
hints:
  - '`results` is unbuffered, so each send blocks until someone receives. Who receives? The `for r := range results` loop. When does that loop start?'
  - '`wg.Wait()` waits for the senders, but the senders are waiting for the receiver, which only starts after `wg.Wait()` returns. Everyone waits for everyone: a deadlock.'
  - 'Move `wg.Wait()` and `close(results)` into their own goroutine, so the range loop can start receiving straight away and still ends once every sender is done.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"sync"
    	"time"
    )

    // tallyTips looks up every courier's tips concurrently with tipsFor and
    // returns each courier's total. A courier listed twice gets both lookups
    // added together.
    //
    // BUG: it never returns.
    func tallyTips(couriers []string, tipsFor func(courier string) int) map[string]int {
    	type result struct {
    		courier string
    		tips    int
    	}
    	results := make(chan result)

    	var wg sync.WaitGroup
    	for _, c := range couriers {
    		wg.Go(func() {
    			results <- result{c, tipsFor(c)}
    		})
    	}
    	wg.Wait()
    	close(results)

    	totals := map[string]int{}
    	for r := range results {
    		totals[r.courier] += r.tips
    	}
    	return totals
    }

    func main() {
    	tipsFor := func(courier string) int {
    		time.Sleep(50 * time.Millisecond) // a slow payments API
    		return len(courier) * 100
    	}
    	fmt.Println("tallying tips...")
    	fmt.Println(tallyTips([]string{"ana", "ben", "cyrus"}, tipsFor))
    	// want: map[ana:300 ben:300 cyrus:500]
    }
  solution: |
    package main

    import (
    	"fmt"
    	"sync"
    	"time"
    )

    // tallyTips looks up every courier's tips concurrently with tipsFor and
    // returns each courier's total. A courier listed twice gets both lookups
    // added together.
    func tallyTips(couriers []string, tipsFor func(courier string) int) map[string]int {
    	type result struct {
    		courier string
    		tips    int
    	}
    	results := make(chan result)

    	var wg sync.WaitGroup
    	for _, c := range couriers {
    		wg.Go(func() {
    			results <- result{c, tipsFor(c)}
    		})
    	}
    	go func() {
    		wg.Wait()
    		close(results)
    	}()

    	totals := map[string]int{}
    	for r := range results {
    		totals[r.courier] += r.tips
    	}
    	return totals
    }

    func main() {
    	tipsFor := func(courier string) int {
    		time.Sleep(50 * time.Millisecond) // a slow payments API
    		return len(courier) * 100
    	}
    	fmt.Println("tallying tips...")
    	fmt.Println(tallyTips([]string{"ana", "ben", "cyrus"}, tipsFor))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"maps"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // slowTips takes a second (of fake time) and pays 100 per letter.
    func slowTips(courier string) int {
    	time.Sleep(time.Second)
    	return len(courier) * 100
    }

    // tally runs tallyTips and fails if it hasn't returned within an hour.
    func tally(t *testing.T, couriers []string) (map[string]int, time.Duration) {
    	t.Helper()
    	done := make(chan map[string]int, 1)
    	start := time.Now()
    	go func() { done <- tallyTips(couriers, slowTips) }()
    	select {
    	case got := <-done:
    		return got, time.Since(start)
    	case <-time.After(time.Hour):
    		t.Fatalf("tallyTips(%q) still hadn't returned after an hour: it's deadlocked", couriers)
    		return nil, 0
    	}
    }

    func TestTallyTips(t *testing.T) {
    	tests := []struct {
    		couriers []string
    		want     map[string]int
    	}{
    		{[]string{"ana", "ben", "cyrus"}, map[string]int{"ana": 300, "ben": 300, "cyrus": 500}},
    		{[]string{"dee"}, map[string]int{"dee": 300}},
    		{[]string{"ana", "ana", "bo"}, map[string]int{"ana": 600, "bo": 200}},
    		{nil, map[string]int{}},
    	}
    	for _, tt := range tests {
    		synctest.Test(t, func(t *testing.T) {
    			got, took := tally(t, tt.couriers)
    			if !maps.Equal(got, tt.want) {
    				t.Errorf("tallyTips(%q) = %v, want %v", tt.couriers, got, tt.want)
    			}
    			if len(tt.couriers) > 0 && took != time.Second {
    				t.Errorf("tallyTips(%q) with 1s lookups took %v, want 1s: keep the lookups concurrent", tt.couriers, took)
    			}
    		})
    	}
    }

    func TestTallyTipsMany(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		var couriers []string
    		want := map[string]int{}
    		for i := range 300 {
    			c := fmt.Sprint("courier", i%100)
    			couriers = append(couriers, c)
    			want[c] += len(c) * 100
    		}
    		got, took := tally(t, couriers)
    		if !maps.Equal(got, want) {
    			t.Errorf("tallyTips(300 lookups for 100 couriers) = %v, want %v", got, want)
    		}
    		if took != time.Second {
    			t.Errorf("tallyTips(300 lookups) took %v, want 1s: keep the lookups concurrent", took)
    		}
    	})
    }
---

At the end of each shift, Dispatchly tallies every courier's tips. The
payments API is slow, so `tallyTips` looks them up concurrently and collects
the results over a channel.

Press **Run**: it never finishes. Go spots that every goroutine is stuck and
stops the program with `fatal error: all goroutines are asleep - deadlock!`.

**Fix `tallyTips`** so it returns every courier's total. Keep the lookups
concurrent: with lookups that take 1s each, tallying any number of couriers
should take 1s.

## Example

```go
tallyTips([]string{"ana", "ben", "cyrus", "ana"}, tipsFor)
// map[ana:600 ben:300 cyrus:500] (ana is listed twice, so her lookups add up)
```

## Constraints

- `couriers` may be empty (return an empty map) or list a courier more than
  once.
- The tests run in a `synctest` bubble with fake time. If `tallyTips` is still
  stuck an hour of fake time later, they report the deadlock at once instead
  of hanging.
