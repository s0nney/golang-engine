---
title: Order Tracking Hub
difficulty: hard
after: coordination-patterns
hints:
  - 'Give each subscription a buffered channel of size `buffer` and keep them in a `map[string]map[*Subscription]struct{}` keyed by order ID, all guarded by one mutex. `Publish` tries a non-blocking send (`select` with `default`) to each subscriber of that order; if the buffer is full, that subscriber is disconnected.'
  - 'Write one helper, called with the mutex held, that ends a subscription: if it already has an error, do nothing; otherwise record the error, remove it from the map and close its channel. Every way a subscription can end (too slow, ctx done, hub closed) goes through it, so a channel is closed exactly once and never sent on afterwards.'
  - '`context.AfterFunc(ctx, f)` runs `f` in its own goroutine once ctx is done, and returns a `stop` func. Use it to end the subscription with `context.Cause(ctx)` (locking the mutex inside `f`), and call `stop()` when the subscription ends for another reason so the context doesn''t keep it alive.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    )

    var (
    	ErrTooSlow   = errors.New("subscriber fell too far behind")
    	ErrHubClosed = errors.New("tracking hub closed")
    )

    // Event is a status update for one order.
    type Event struct {
    	OrderID string
    	Status  string
    }

    // Hub fans order status updates out to subscribers.
    type Hub struct {
    }

    // Subscription receives the events for one order on C. C is closed when
    // the subscription ends; Err then says why.
    type Subscription struct {
    	C <-chan Event
    }

    // NewHub returns a hub whose subscribers may fall up to buffer events behind.
    func NewHub(buffer int) *Hub {
    	return &Hub{}
    }

    // Subscribe starts receiving events for orderID until ctx is done.
    func (h *Hub) Subscribe(ctx context.Context, orderID string) (*Subscription, error) {
    	return &Subscription{C: make(chan Event)}, nil
    }

    // Publish delivers e to every subscriber of e.OrderID without blocking.
    func (h *Hub) Publish(e Event) error {
    	return nil
    }

    // Close ends every subscription. Publish and Subscribe then fail.
    func (h *Hub) Close() {
    }

    // Err returns nil while the subscription is active, then why it ended.
    func (s *Subscription) Err() error {
    	return nil
    }

    func main() {
    	hub := NewHub(2)
    	customer, _ := hub.Subscribe(context.Background(), "A1")
    	for _, status := range []string{"accepted", "cooking", "picked up"} {
    		fmt.Println("publish", status, "->", hub.Publish(Event{"A1", status}))
    	}
    	for {
    		select {
    		case e, ok := <-customer.C:
    			if !ok {
    				fmt.Println("subscription ended:", customer.Err())
    				return
    			}
    			fmt.Println("customer sees:", e.Status)
    		default:
    			fmt.Println("nothing more waiting; subscription error:", customer.Err())
    			return
    		}
    	}
    	// want: the customer sees "accepted" and "cooking", then the subscription
    	// ends with ErrTooSlow, because "picked up" didn't fit in the buffer of 2
    }
  solution: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"sync"
    )

    var (
    	ErrTooSlow   = errors.New("subscriber fell too far behind")
    	ErrHubClosed = errors.New("tracking hub closed")
    )

    // Event is a status update for one order.
    type Event struct {
    	OrderID string
    	Status  string
    }

    // Hub fans order status updates out to subscribers.
    type Hub struct {
    	buffer int

    	mu     sync.Mutex
    	closed bool
    	subs   map[string]map[*Subscription]struct{}
    }

    // Subscription receives the events for one order on C. C is closed when
    // the subscription ends; Err then says why.
    type Subscription struct {
    	C <-chan Event

    	hub     *Hub
    	orderID string
    	ch      chan Event
    	err     error       // guarded by hub.mu
    	stop    func() bool // unregisters the context.AfterFunc
    }

    // NewHub returns a hub whose subscribers may fall up to buffer events behind.
    func NewHub(buffer int) *Hub {
    	return &Hub{buffer: buffer, subs: map[string]map[*Subscription]struct{}{}}
    }

    // Subscribe starts receiving events for orderID until ctx is done.
    func (h *Hub) Subscribe(ctx context.Context, orderID string) (*Subscription, error) {
    	h.mu.Lock()
    	defer h.mu.Unlock()
    	if h.closed {
    		return nil, ErrHubClosed
    	}
    	if ctx.Err() != nil {
    		return nil, context.Cause(ctx)
    	}
    	ch := make(chan Event, h.buffer)
    	s := &Subscription{C: ch, hub: h, orderID: orderID, ch: ch}
    	if h.subs[orderID] == nil {
    		h.subs[orderID] = map[*Subscription]struct{}{}
    	}
    	h.subs[orderID][s] = struct{}{}
    	// The callback needs h.mu, which we hold, so it can't run before
    	// s.stop is set.
    	s.stop = context.AfterFunc(ctx, func() {
    		h.mu.Lock()
    		defer h.mu.Unlock()
    		h.end(s, context.Cause(ctx))
    	})
    	return s, nil
    }

    // end ends s with err, once. h.mu must be held.
    func (h *Hub) end(s *Subscription, err error) {
    	if s.err != nil {
    		return
    	}
    	s.err = err
    	delete(h.subs[s.orderID], s)
    	if len(h.subs[s.orderID]) == 0 {
    		delete(h.subs, s.orderID)
    	}
    	close(s.ch)
    	s.stop()
    }

    // Publish delivers e to every subscriber of e.OrderID without blocking.
    func (h *Hub) Publish(e Event) error {
    	h.mu.Lock()
    	defer h.mu.Unlock()
    	if h.closed {
    		return ErrHubClosed
    	}
    	for s := range h.subs[e.OrderID] {
    		select {
    		case s.ch <- e:
    		default:
    			h.end(s, ErrTooSlow)
    		}
    	}
    	return nil
    }

    // Close ends every subscription. Publish and Subscribe then fail.
    func (h *Hub) Close() {
    	h.mu.Lock()
    	defer h.mu.Unlock()
    	if h.closed {
    		return
    	}
    	h.closed = true
    	for _, subs := range h.subs {
    		for s := range subs {
    			h.end(s, ErrHubClosed)
    		}
    	}
    }

    // Err returns nil while the subscription is active, then why it ended.
    func (s *Subscription) Err() error {
    	s.hub.mu.Lock()
    	defer s.hub.mu.Unlock()
    	return s.err
    }

    func main() {
    	hub := NewHub(2)
    	customer, _ := hub.Subscribe(context.Background(), "A1")
    	for _, status := range []string{"accepted", "cooking", "picked up"} {
    		fmt.Println("publish", status, "->", hub.Publish(Event{"A1", status}))
    	}
    	for {
    		select {
    		case e, ok := <-customer.C:
    			if !ok {
    				fmt.Println("subscription ended:", customer.Err())
    				return
    			}
    			fmt.Println("customer sees:", e.Status)
    		default:
    			fmt.Println("nothing more waiting; subscription error:", customer.Err())
    			return
    		}
    	}
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"slices"
    	"sync"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // publish calls h.Publish and fails if it blocks. If it does, it reads
    // from the rescue subscriptions to unblock it, so the test can end.
    func publish(t *testing.T, h *Hub, orderID, status string, rescue ...*Subscription) error {
    	t.Helper()
    	done := make(chan error, 1)
    	go func() { done <- h.Publish(Event{orderID, status}) }()
    	select {
    	case err := <-done:
    		return err
    	case <-time.After(time.Hour):
    		for _, s := range rescue {
    			go func() {
    				for range s.C {
    				}
    			}()
    		}
    		t.Fatalf("Publish(%s, %s) blocked for an hour: it must never wait for a subscriber", orderID, status)
    		return nil
    	}
    }

    // subscribe calls h.Subscribe and fails the test on an error.
    func subscribe(t *testing.T, h *Hub, ctx context.Context, orderID string) *Subscription {
    	t.Helper()
    	s, err := h.Subscribe(ctx, orderID)
    	if err != nil || s == nil || s.C == nil {
    		t.Fatalf("Subscribe(ctx, %q) = %v, %v; want a subscription with a channel", orderID, s, err)
    	}
    	return s
    }

    // drain reads what's waiting on s.C without blocking. closed reports
    // whether the channel was closed.
    func drain(s *Subscription) (statuses []string, closed bool) {
    	for {
    		select {
    		case e, ok := <-s.C:
    			if !ok {
    				return statuses, true
    			}
    			statuses = append(statuses, e.Status)
    		default:
    			return statuses, false
    		}
    	}
    }

    func TestEventsGoToTheirOrdersSubscribers(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		h := NewHub(10)
    		ana := subscribe(t, h, t.Context(), "A1")
    		ben := subscribe(t, h, t.Context(), "A1")
    		cy := subscribe(t, h, t.Context(), "B2")

    		for _, e := range []Event{{"A1", "accepted"}, {"B2", "accepted"}, {"A1", "cooking"}, {"C3", "accepted"}, {"A1", "picked up"}} {
    			if err := publish(t, h, e.OrderID, e.Status); err != nil {
    				t.Errorf("Publish(%v) = %v, want nil", e, err)
    			}
    		}
    		want := []string{"accepted", "cooking", "picked up"}
    		for name, s := range map[string]*Subscription{"first A1 subscriber": ana, "second A1 subscriber": ben} {
    			if got, closed := drain(s); !slices.Equal(got, want) || closed {
    				t.Errorf("%s got %q (closed: %v), want %q, still open", name, got, closed, want)
    			}
    		}
    		if got, closed := drain(cy); !slices.Equal(got, []string{"accepted"}) || closed {
    			t.Errorf("B2 subscriber got %q (closed: %v), want [accepted], still open", got, closed)
    		}
    		if err := ana.Err(); err != nil {
    			t.Errorf("Err() on an active subscription = %v, want nil", err)
    		}
    	})
    }

    func TestOnlyEventsAfterSubscribing(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		h := NewHub(10)
    		publish(t, h, "A1", "accepted")
    		s := subscribe(t, h, t.Context(), "A1")
    		publish(t, h, "A1", "cooking")
    		if got, _ := drain(s); !slices.Equal(got, []string{"cooking"}) {
    			t.Errorf("subscribed after accepted, before cooking: got %q, want [cooking]", got)
    		}
    	})
    }

    func TestSlowSubscriberIsDisconnected(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		h := NewHub(3)
    		slow := subscribe(t, h, t.Context(), "A1") // never reads until the end
    		fast := subscribe(t, h, t.Context(), "A1")

    		var fastGot []string
    		for _, status := range []string{"accepted", "cooking", "ready", "picked up", "delivered"} {
    			if err := publish(t, h, "A1", status, slow, fast); err != nil {
    				t.Errorf("Publish(%s) = %v, want nil (a slow subscriber is not the publisher's problem)", status, err)
    			}
    			got, _ := drain(fast)
    			fastGot = append(fastGot, got...)
    		}

    		got, closed := drain(slow)
    		if want := []string{"accepted", "cooking", "ready"}; !slices.Equal(got, want) || !closed {
    			t.Errorf("buffer 3, slow subscriber never read while 5 events were published: it got %q (closed: %v), want %q and then its channel closed", got, closed, want)
    		}
    		if err := slow.Err(); !errors.Is(err, ErrTooSlow) {
    			t.Errorf("slow subscriber's Err() = %v, want ErrTooSlow", err)
    		}
    		if want := []string{"accepted", "cooking", "ready", "picked up", "delivered"}; !slices.Equal(fastGot, want) {
    			t.Errorf("fast subscriber got %q, want %q", fastGot, want)
    		}
    		if err := fast.Err(); err != nil {
    			t.Errorf("fast subscriber's Err() = %v, want nil", err)
    		}
    	})
    }

    func TestCancelEndsSubscription(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errLeft := errors.New("customer closed the app")
    		h := NewHub(5)
    		ctx, cancel := context.WithCancelCause(t.Context())
    		s := subscribe(t, h, ctx, "A1")
    		other := subscribe(t, h, t.Context(), "A1")
    		publish(t, h, "A1", "accepted")

    		cancel(errLeft)
    		synctest.Wait()
    		if got, closed := drain(s); !slices.Equal(got, []string{"accepted"}) || !closed {
    			t.Errorf("after ctx was cancelled: got %q (closed: %v), want [accepted] and then the channel closed", got, closed)
    		}
    		if err := s.Err(); !errors.Is(err, errLeft) {
    			t.Errorf("Err() after ctx was cancelled = %v, want the cause %q", err, errLeft)
    		}

    		// Publishing after the cancel must neither panic nor reach s.
    		if err := publish(t, h, "A1", "cooking"); err != nil {
    			t.Errorf("Publish after a subscriber left = %v, want nil", err)
    		}
    		if got, _ := drain(other); !slices.Equal(got, []string{"accepted", "cooking"}) {
    			t.Errorf("remaining subscriber got %q, want [accepted cooking]", got)
    		}
    	})
    }

    func TestSubscribeWithCancelledContext(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		errGone := errors.New("gone")
    		ctx, cancel := context.WithCancelCause(t.Context())
    		cancel(errGone)
    		s, err := NewHub(5).Subscribe(ctx, "A1")
    		if !errors.Is(err, errGone) || s != nil {
    			t.Errorf("Subscribe with a cancelled ctx = %v, %v; want nil, the cause %q", s, err, errGone)
    		}
    	})
    }

    func TestClose(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		h := NewHub(5)
    		ctx, cancel := context.WithCancel(t.Context())
    		a := subscribe(t, h, ctx, "A1")
    		b := subscribe(t, h, t.Context(), "B2")
    		publish(t, h, "A1", "accepted")

    		h.Close()
    		if got, closed := drain(a); !slices.Equal(got, []string{"accepted"}) || !closed {
    			t.Errorf("after Close: A1 subscriber got %q (closed: %v), want [accepted] and then closed", got, closed)
    		}
    		if _, closed := drain(b); !closed {
    			t.Errorf("after Close: B2 subscriber's channel is still open")
    		}
    		for name, s := range map[string]*Subscription{"A1": a, "B2": b} {
    			if err := s.Err(); !errors.Is(err, ErrHubClosed) {
    				t.Errorf("after Close: %s subscriber's Err() = %v, want ErrHubClosed", name, err)
    			}
    		}
    		if err := publish(t, h, "A1", "cooking"); !errors.Is(err, ErrHubClosed) {
    			t.Errorf("Publish after Close = %v, want ErrHubClosed", err)
    		}
    		if s, err := h.Subscribe(t.Context(), "A1"); !errors.Is(err, ErrHubClosed) || s != nil {
    			t.Errorf("Subscribe after Close = %v, %v; want nil, ErrHubClosed", s, err)
    		}

    		h.Close() // a second Close is harmless
    		cancel()  // cancelling an ended subscription is harmless
    		synctest.Wait()
    		if err := a.Err(); !errors.Is(err, ErrHubClosed) {
    			t.Errorf("cancelling ctx after Close changed Err() to %v, want it to stay ErrHubClosed", err)
    		}
    	})
    }

    // TestCancelRacesClose cancels a subscription's ctx and immediately closes
    // the hub, so both try to end the same subscription at about the same time.
    func TestCancelRacesClose(t *testing.T) {
    	for range 100 {
    		synctest.Test(t, func(t *testing.T) {
    			h := NewHub(1)
    			ctx, cancel := context.WithCancel(t.Context())
    			s := subscribe(t, h, ctx, "A1")
    			cancel()
    			h.Close()
    			synctest.Wait()
    			if _, closed := drain(s); !closed {
    				t.Fatalf("after cancel and Close, the subscription's channel is still open")
    			}
    			if err := s.Err(); !errors.Is(err, context.Canceled) && !errors.Is(err, ErrHubClosed) {
    				t.Fatalf("after cancel and Close, Err() = %v, want context.Canceled or ErrHubClosed", err)
    			}
    		})
    	}
    }

    func TestConcurrentUse(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		const publishers, events = 8, 200
    		h := NewHub(publishers * events)
    		orders := []string{"A1", "B2", "C3"}

    		type sub struct {
    			s      *Subscription
    			got    []Event
    			cancel context.CancelFunc
    		}
    		var subs []*sub
    		for i := range 12 {
    			ctx, cancel := context.WithCancel(t.Context())
    			subs = append(subs, &sub{s: subscribe(t, h, ctx, orders[i%3]), cancel: cancel})
    		}

    		var wg sync.WaitGroup
    		for _, sb := range subs {
    			wg.Go(func() {
    				for e := range sb.s.C {
    					sb.got = append(sb.got, e)
    				}
    			})
    		}
    		var pubs sync.WaitGroup
    		for p := range publishers {
    			pubs.Go(func() {
    				for i := range events {
    					h.Publish(Event{orders[i%3], fmt.Sprint(p, ":", i)})
    					if p == 0 && i%50 == 49 {
    						subs[i/50].cancel() // some customers leave mid-stream
    					}
    				}
    			})
    		}
    		pubs.Wait()
    		h.Close()
    		wg.Wait()

    		for i, sb := range subs {
    			last := map[int]int{}
    			for _, e := range sb.got {
    				if e.OrderID != orders[i%3] {
    					t.Fatalf("subscriber %d of %s received an event for %s", i, orders[i%3], e.OrderID)
    				}
    				var p, n int
    				fmt.Sscanf(e.Status, "%d:%d", &p, &n)
    				if prev, ok := last[p]; ok && n != prev+3 {
    					t.Fatalf("subscriber %d received publisher %d's event %d right after %d: events must arrive in order, with none missing", i, p, n, prev)
    				}
    				last[p] = n
    			}
    			want := 0 // events for this subscriber's order
    			for n := range events {
    				if n%3 == i%3 {
    					want += publishers
    				}
    			}
    			if i >= 4 && len(sb.got) != want { // subscribers 0-3 left early
    				t.Errorf("subscriber %d stayed until Close but got %d events, want %d", i, len(sb.got), want)
    			}
    			if err := sb.s.Err(); i >= 4 && !errors.Is(err, ErrHubClosed) {
    				t.Errorf("subscriber %d stayed until Close: Err() = %v, want ErrHubClosed", i, err)
    			}
    		}
    	})
    }
---

Customers love watching their order move: *accepted*, *cooking*, *picked up*,
*delivered*. Dispatchly's tracking hub fans each order's status updates out to
everyone watching it: the customer's phone, the restaurant tablet, the support
dashboard. Some of those are on terrible connections, and one slow phone must
never hold up anyone else.

Implement the `Hub`:

- `NewHub(buffer)` creates a hub. Each subscriber may fall up to `buffer`
  events behind.
- `Subscribe(ctx, orderID)` returns a `Subscription` whose channel `C`
  receives every event for `orderID` published from now on, in order. If the
  hub is closed, return `ErrHubClosed`; if `ctx` is already done, return
  `context.Cause(ctx)`.
- `Publish(e)` delivers `e` to every subscriber of `e.OrderID` and **never
  blocks**. A subscriber whose buffer is full is **disconnected** instead:
  its channel is closed (after the events already buffered) and its `Err()`
  becomes `ErrTooSlow`. So a subscriber either sees every event, or knows it
  missed some. After `Close`, `Publish` returns `ErrHubClosed`.
- When a subscriber's `ctx` is done, its subscription ends: `C` is closed and
  `Err()` returns `context.Cause(ctx)`.
- `Close()` ends every subscription with `ErrHubClosed`. Calling it again does
  nothing.
- `Err()` returns `nil` while the subscription is active, then the reason it
  ended. A subscription ends only once: later events don't change its reason.

## Example

```go
hub := NewHub(2)
sub, _ := hub.Subscribe(ctx, "A1")
hub.Publish(Event{"A1", "accepted"})   // buffered
hub.Publish(Event{"A1", "cooking"})    // buffered
hub.Publish(Event{"A1", "picked up"})  // buffer full: sub is disconnected
// sub.C yields accepted, cooking, then is closed; sub.Err() == ErrTooSlow
```

## Constraints

- Many goroutines publish, subscribe, cancel and close at once. Never send on
  a closed channel, and never close one twice.
- The hub needs no goroutines of its own (`context.AfterFunc` runs its
  callback in a goroutine that ends when the callback returns).
- The tests run in a `synctest` bubble: a `Publish` that blocks is reported,
  and any goroutine left blocked fails the test.
