---
title: Ordered Quotes
difficulty: hard
after: coordination-patterns
hints:
  - 'Give every order its own result channel (buffered, size 1): a "future". One goroutine takes orders, starts a pricing goroutine for each that fills in its future, and queues the futures in arrival order on a channel. A second goroutine takes futures off that queue **in order**, waits for each result and sends the quote on the output.'
  - 'The window: a semaphore of size `workers`. Acquire a slot *before* taking an order and starting its pricing, and release it only *after* its quote has been sent on the output. Then at most `workers` orders are ever between "taken" and "sent", whether they''re still pricing or waiting their turn.'
  - 'For errors, derive `ctx, cancel := context.WithCancelCause(ctx)`; the first failure calls `cancel(err)`. Every blocking step (acquire, read an order, wait for a future, send a quote) selects on `ctx.Done()` too. Close the output when the sending goroutine stops, and send the error and close the error channel only after a `WaitGroup` says every goroutine has finished.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"time"
    )

    type Order struct {
    	ID    int
    	Items int
    }

    type Quote struct {
    	OrderID int
    	Cents   int
    }

    // quoteAll prices orders concurrently and sends the quotes in the same
    // order the orders arrived. At most workers orders are ever taken from
    // orders but not yet sent. The first error (or ctx being done) stops
    // everything; it's then sent on the error channel. Both returned channels
    // are closed when quoteAll is finished.
    func quoteAll(ctx context.Context, orders <-chan Order, workers int, price func(context.Context, Order) (Quote, error)) (<-chan Quote, <-chan error) {
    	quotes := make(chan Quote)
    	errc := make(chan error, 1)
    	go func() {
    		defer close(errc)
    		defer close(quotes)
    		for o := range orders {
    			q, err := price(ctx, o)
    			if err != nil {
    				errc <- err
    				return
    			}
    			quotes <- q
    		}
    	}()
    	return quotes, errc
    }

    func main() {
    	orders := make(chan Order, 6)
    	for id := range 6 {
    		orders <- Order{ID: id + 1, Items: 6 - id}
    	}
    	close(orders)

    	start := time.Now()
    	price := func(ctx context.Context, o Order) (Quote, error) {
    		time.Sleep(time.Duration(o.Items) * 10 * time.Millisecond) // big orders are slow to price
    		return Quote{o.ID, 199 + 450*o.Items}, nil
    	}
    	quotes, errc := quoteAll(context.Background(), orders, 3, price)
    	for q := range quotes {
    		fmt.Printf("%v: order %d costs %d cents\n", time.Since(start).Round(10*time.Millisecond), q.OrderID, q.Cents)
    	}
    	fmt.Println("error:", <-errc)
    	// want orders 1 to 6 in order, all done by about 90ms, then "error: <nil>"
    }
  solution: |
    package main

    import (
    	"context"
    	"fmt"
    	"sync"
    	"time"
    )

    type Order struct {
    	ID    int
    	Items int
    }

    type Quote struct {
    	OrderID int
    	Cents   int
    }

    type result struct {
    	q   Quote
    	err error
    }

    // quoteAll prices orders concurrently and sends the quotes in the same
    // order the orders arrived. At most workers orders are ever taken from
    // orders but not yet sent. The first error (or ctx being done) stops
    // everything; it's then sent on the error channel. Both returned channels
    // are closed when quoteAll is finished.
    func quoteAll(ctx context.Context, orders <-chan Order, workers int, price func(context.Context, Order) (Quote, error)) (<-chan Quote, <-chan error) {
    	ctx, cancel := context.WithCancelCause(ctx)
    	quotes := make(chan Quote)
    	errc := make(chan error, 1)

    	window := make(chan struct{}, workers)     // taken but not yet sent
    	futures := make(chan chan result, workers) // in arrival order
    	var wg sync.WaitGroup

    	// Take orders and start pricing them.
    	wg.Go(func() {
    		defer close(futures)
    		for {
    			select {
    			case window <- struct{}{}:
    			case <-ctx.Done():
    				return
    			}
    			var o Order
    			var ok bool
    			select {
    			case o, ok = <-orders:
    			case <-ctx.Done():
    			}
    			if !ok || ctx.Err() != nil {
    				return
    			}
    			f := make(chan result, 1)
    			futures <- f // never blocks: at most workers futures exist
    			wg.Go(func() {
    				q, err := price(ctx, o)
    				f <- result{q, err}
    			})
    		}
    	})

    	// Send the quotes in order.
    	wg.Go(func() {
    		defer close(quotes)
    		for f := range futures {
    			var r result
    			select {
    			case r = <-f:
    			case <-ctx.Done():
    				return
    			}
    			if r.err != nil {
    				cancel(r.err) // only the first cause sticks
    				return
    			}
    			select {
    			case quotes <- r.q:
    				<-window
    			case <-ctx.Done():
    				return
    			}
    		}
    	})

    	go func() {
    		wg.Wait()
    		if err := context.Cause(ctx); err != nil {
    			errc <- err
    		}
    		cancel(nil)
    		close(errc)
    	}()
    	return quotes, errc
    }

    func main() {
    	orders := make(chan Order, 6)
    	for id := range 6 {
    		orders <- Order{ID: id + 1, Items: 6 - id}
    	}
    	close(orders)

    	start := time.Now()
    	price := func(ctx context.Context, o Order) (Quote, error) {
    		time.Sleep(time.Duration(o.Items) * 10 * time.Millisecond) // big orders are slow to price
    		return Quote{o.ID, 199 + 450*o.Items}, nil
    	}
    	quotes, errc := quoteAll(context.Background(), orders, 3, price)
    	for q := range quotes {
    		fmt.Printf("%v: order %d costs %d cents\n", time.Since(start).Round(10*time.Millisecond), q.OrderID, q.Cents)
    	}
    	fmt.Println("error:", <-errc)
    }
  tests: |
    package main

    import (
    	"cmp"
    	"context"
    	"errors"
    	"fmt"
    	"slices"
    	"sync"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // pricer is a fake price func: order o takes o.Items seconds (unless
    // overridden) and costs 100*ID cents. It records starts and cancellations.
    type pricer struct {
    	start time.Time
    	took  map[int]time.Duration
    	fail  map[int]error

    	mu        sync.Mutex
    	starts    []start
    	running   int
    	peak      int
    	cancelled []string
    }

    type start struct {
    	at time.Duration
    	id int
    }

    func newPricer() *pricer {
    	return &pricer{start: time.Now(), took: map[int]time.Duration{}, fail: map[int]error{}}
    }

    func (p *pricer) price(ctx context.Context, o Order) (Quote, error) {
    	p.mu.Lock()
    	p.starts = append(p.starts, start{time.Since(p.start), o.ID})
    	p.running++
    	p.peak = max(p.peak, p.running)
    	d, ok := p.took[o.ID]
    	if !ok {
    		d = time.Duration(o.Items) * time.Second
    	}
    	p.mu.Unlock()
    	defer func() {
    		p.mu.Lock()
    		defer p.mu.Unlock()
    		p.running--
    	}()

    	select {
    	case <-time.After(d):
    		if err := p.fail[o.ID]; err != nil {
    			return Quote{}, err
    		}
    		return Quote{o.ID, 100 * o.ID}, nil
    	case <-ctx.Done():
    		p.mu.Lock()
    		p.cancelled = append(p.cancelled, fmt.Sprint(time.Since(p.start), " #", o.ID, ": ", context.Cause(ctx)))
    		p.mu.Unlock()
    		return Quote{}, ctx.Err()
    	}
    }

    // snapshot returns the starts as "1s #5", sorted by time and then by ID
    // (calls starting at the same instant may run in any order).
    func (p *pricer) snapshot() (starts []string, peak int, cancelled []string) {
    	p.mu.Lock()
    	defer p.mu.Unlock()
    	sorted := slices.Clone(p.starts)
    	slices.SortFunc(sorted, func(a, b start) int {
    		return cmp.Or(cmp.Compare(a.at, b.at), cmp.Compare(a.id, b.id))
    	})
    	for _, s := range sorted {
    		starts = append(starts, fmt.Sprint(s.at, " #", s.id))
    	}
    	cancelled = slices.Clone(p.cancelled)
    	slices.Sort(cancelled)
    	return starts, p.peak, cancelled
    }

    // queue returns a closed channel of orders with the given item counts;
    // IDs are 1, 2, 3, ...
    func queue(items ...int) chan Order {
    	ch := make(chan Order, len(items))
    	for i, n := range items {
    		ch <- Order{ID: i + 1, Items: n}
    	}
    	close(ch)
    	return ch
    }

    // pipe holds the two channels quoteAll returns.
    type pipe struct {
    	quotes <-chan Quote
    	errc   <-chan error
    }

    func both(quotes <-chan Quote, errc <-chan error) pipe { return pipe{quotes, errc} }

    // collect reads every quote and then the error channel.
    func collect(t *testing.T, p pipe) (ids []int, err error) {
    	t.Helper()
    	quotes, errc := p.quotes, p.errc
    	for {
    		select {
    		case q, ok := <-quotes:
    			if !ok {
    				return ids, finish(t, errc)
    			}
    			if q.Cents != 100*q.OrderID {
    				t.Errorf("quote %v, want Cents = %d", q, 100*q.OrderID)
    			}
    			ids = append(ids, q.OrderID)
    		case <-time.After(24 * time.Hour):
    			t.Fatalf("after quotes for %v, the quote channel wasn't closed within a day", ids)
    		}
    	}
    }

    // finish reads the error channel, which must deliver at most one error
    // and then be closed.
    func finish(t *testing.T, errc <-chan error) error {
    	t.Helper()
    	var got error
    	for {
    		select {
    		case err, ok := <-errc:
    			if !ok {
    				return got
    			}
    			if got != nil {
    				t.Errorf("error channel delivered a second error %v after %v, want at most one", err, got)
    			}
    			got = err
    		case <-time.After(time.Hour):
    			t.Fatalf("error channel still open an hour after the quote channel closed")
    		}
    	}
    }

    func TestQuotesInOrder(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		p := newPricer()
    		items := []int{5, 1, 4, 2, 3, 9, 1, 1, 6, 2, 2, 7}
    		ids, err := collect(t, both(quoteAll(t.Context(), queue(items...), 4, p.price)))
    		want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
    		if !slices.Equal(ids, want) {
    			t.Errorf("quotes came out for orders %v, want %v (the order they went in)", ids, want)
    		}
    		if err != nil {
    			t.Errorf("error channel delivered %v, want nothing (just closed)", err)
    		}
    		if _, peak, _ := p.snapshot(); peak != 4 {
    			t.Errorf("at most %d price calls ran at once, want 4 (workers = 4)", peak)
    		}
    	})
    }

    func TestWindowBoundsWorkInProgress(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		p := newPricer()
    		p.took[1] = 10 * time.Second // order 1 holds up everything behind it
    		ids, err := collect(t, both(quoteAll(t.Context(), queue(1, 1, 1, 1, 1, 1, 1, 1), 3, p.price)))
    		starts, _, _ := p.snapshot()
    		want := []string{"0s #1", "0s #2", "0s #3", "10s #4", "10s #5", "10s #6", "11s #7", "11s #8"}
    		if !slices.Equal(starts, want) {
    			t.Errorf("workers 3, order 1 takes 10s, the rest 1s:\nprice calls started %v\nwant                %v\n(at most 3 orders may be taken but not yet sent: #2 and #3 finish early but must wait for #1 to be sent before new orders start)", starts, want)
    		}
    		if !slices.Equal(ids, []int{1, 2, 3, 4, 5, 6, 7, 8}) || err != nil {
    			t.Errorf("got quotes for %v and error %v, want [1 ... 8] and nil", ids, err)
    		}
    	})
    }

    func TestEmptyInput(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		p := newPricer()
    		ids, err := collect(t, both(quoteAll(t.Context(), queue(), 3, p.price)))
    		if len(ids) != 0 || err != nil {
    			t.Errorf("no orders: got quotes %v and error %v, want none and nil", ids, err)
    		}
    	})
    }

    func TestFirstErrorStopsEverything(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errNoMenu := errors.New("restaurant has no menu")
    		p := newPricer()
    		p.took[3] = 1500 * time.Millisecond
    		p.fail[3] = errNoMenu
    		ids, err := collect(t, both(quoteAll(t.Context(), queue(1, 1, 5, 5, 9, 9, 9, 9, 9), 4, p.price)))
    		starts, _, cancelled := p.snapshot()

    		if !errors.Is(err, errNoMenu) {
    			t.Errorf("order 3 failed with %q: error channel delivered %v, want that error", errNoMenu, err)
    		}
    		if !slices.Equal(ids, []int{1, 2}) {
    			t.Errorf("order 3 failed: got quotes for %v, want [1 2] (the orders before it, and nothing after)", ids)
    		}
    		wantStarts := []string{"0s #1", "0s #2", "0s #3", "0s #4", "1s #5", "1s #6"}
    		if !slices.Equal(starts, wantStarts) {
    			t.Errorf("price calls started %v, want %v (none after the failure at 1.5s)", starts, wantStarts)
    		}
    		slices.Sort(cancelled)
    		wantCancelled := []string{"1.5s #4: restaurant has no menu", "1.5s #5: restaurant has no menu", "1.5s #6: restaurant has no menu"}
    		if !slices.Equal(cancelled, wantCancelled) {
    			t.Errorf("price calls cancelled: %v, want %v (cancel their ctx with the error as its cause)", cancelled, wantCancelled)
    		}
    	})
    }

    func TestCancelWithAbandonedConsumer(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errShutdown := errors.New("pricing service shutting down")
    		ctx, cancel := context.WithCancelCause(t.Context())
    		orders := make(chan Order, 20) // never closed
    		for id := range 20 {
    			orders <- Order{ID: id + 1, Items: 1}
    		}
    		p := newPricer()
    		quotes, errc := quoteAll(ctx, orders, 3, p.price)
    		for range 2 {
    			<-quotes
    		}
    		synctest.Sleep(time.Minute) // the consumer walks away...
    		cancel(errShutdown)         // ...and cancels
    		synctest.Wait()

    		ids, err := collect(t, pipe{quotes, errc})
    		if len(ids) != 0 {
    			t.Errorf("after cancel, got more quotes %v, want the quote channel closed", ids)
    		}
    		if !errors.Is(err, errShutdown) {
    			t.Errorf("cancelled with cause %q: error channel delivered %v, want the cause", errShutdown, err)
    		}
    		if starts, _, _ := p.snapshot(); len(starts) != 5 {
    			t.Errorf("consumer read 2 quotes, workers 3: %d price calls started, want 5 (never more than 3 taken but unsent)", len(starts))
    		}
    	})
    }

    func TestCancelWhileWaitingForOrders(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
    		defer cancel()
    		orders := make(chan Order) // nothing ever arrives
    		start := time.Now()
    		ids, err := collect(t, both(quoteAll(ctx, orders, 3, newPricer().price)))
    		if len(ids) != 0 || !errors.Is(err, context.DeadlineExceeded) {
    			t.Errorf("no orders, 1m timeout: got quotes %v and error %v, want none and context.DeadlineExceeded", ids, err)
    		}
    		if took := time.Since(start); took != time.Minute {
    			t.Errorf("quote channel closed after %v, want 1m", took)
    		}
    	})
    }
---

Pricing a Dispatchly order means asking the restaurant, the courier pool and
the surge model, so it's slow, and some orders are much slower than others.
The checkout screens need the quotes **in the order the orders arrived**,
and the pricing service must never be flooded.

Implement `quoteAll(ctx, orders, workers, price)`. It returns a quote channel
and an error channel, and works in the background:

- Price every order with `price(ctx, order)`, several at once.
- Send the quotes in **the same order** the orders arrived, as soon as each
  one and all the ones before it are ready.
- **Window:** at most `workers` orders may be taken from `orders` but not yet
  sent on the quote channel, whether they're still being priced or waiting
  for an earlier quote. (So a slow order holds up new work, and a slow
  consumer does too.)
- **First error stops everything:** when a `price` call fails, send no more
  quotes (not even for orders before it that aren't sent yet), take no more
  orders, cancel the ctx of the `price` calls still running with that error
  as the cause, and deliver that error on the error channel.
- If `ctx` is done, stop the same way and deliver `context.Cause(ctx)`, even
  if nobody is reading the quote channel and `orders` is never closed.
- Close the quote channel when done. Close the error channel after it (with
  one error, or none if everything succeeded), once every goroutine you
  started has finished.

## Example

`workers = 3`, 8 orders; order 1 takes 10s to price, the rest 1s:

```
t=0s   start pricing 1, 2, 3
t=1s   2 and 3 are ready but must wait for 1; the window is full
t=10s  send 1, 2, 3; start 4, 5, 6
t=11s  send 4, 5, 6; start 7, 8
t=12s  send 7, 8; close the quote channel, then the error channel
```

## Constraints

- `workers` ≥ 1. The consumer reads the quote channel until it's closed, then
  reads the error channel, unless it cancels `ctx` and walks away.
- The tests run in a `synctest` bubble. They check exact start times,
  quote order, which `price` calls were cancelled and why, and that no
  goroutine is left behind.
