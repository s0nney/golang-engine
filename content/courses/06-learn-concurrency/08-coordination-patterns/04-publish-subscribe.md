---
title: Publishing to Multiple Subscribers
quiz:
- question: Two goroutines receive from the same event channel. Does each receive every event?
  options:
  - text: No; each send is received by one receiver
    correct: true
  - text: Yes; channels broadcast every value
  - text: Only when the channel is buffered
  explanation: Sharing one channel distributes work. Broadcasting data requires a separate delivery path for every subscriber.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"sync"
    )

    type Broker struct {
    	mu          sync.Mutex
    	subscribers map[chan string]struct{}
    }

    func (b *Broker) Subscribe() (<-chan string, func()) {
    	b.mu.Lock()
    	if b.subscribers == nil {
    		b.subscribers = make(map[chan string]struct{})
    	}
    	ch := make(chan string, 1)
    	b.subscribers[ch] = struct{}{}
    	b.mu.Unlock()
    	return ch, sync.OnceFunc(func() {
    		b.mu.Lock()
    		defer b.mu.Unlock()
    		delete(b.subscribers, ch)
    		close(ch)
    	})
    }
    func (b *Broker) Publish(event string) int {
    	// Offer the event to every subscriber without waiting for queue space.
    	return 0
    }

    func readNow(ch <-chan string) string {
    	select {
    	case value, ok := <-ch:
    		if !ok {
    			return "closed"
    		}
    		return value
    	default:
    		return "no update"
    	}
    }
    func main() {
    	var b Broker
    	dashboard, stopDashboard := b.Subscribe()
    	defer stopDashboard()
    	preview, stopPreview := b.Subscribe()
    	defer stopPreview()
    	fmt.Println("accepted:", b.Publish("A1 delivered"))
    	fmt.Println("dashboard:", readNow(dashboard))
    	fmt.Println("preview:", readNow(preview))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"sync"
    )

    type Broker struct {
    	mu          sync.Mutex
    	subscribers map[chan string]struct{}
    }

    func (b *Broker) Subscribe() (<-chan string, func()) {
    	b.mu.Lock()
    	if b.subscribers == nil {
    		b.subscribers = make(map[chan string]struct{})
    	}
    	ch := make(chan string, 1)
    	b.subscribers[ch] = struct{}{}
    	b.mu.Unlock()
    	return ch, sync.OnceFunc(func() {
    		b.mu.Lock()
    		defer b.mu.Unlock()
    		delete(b.subscribers, ch)
    		close(ch)
    	})
    }
    func (b *Broker) Publish(event string) int {
    	b.mu.Lock()
    	defer b.mu.Unlock()
    	delivered := 0
    	for ch := range b.subscribers {
    		select {
    		case ch <- event:
    			delivered++
    		default:
    		}
    	}
    	return delivered
    }

    func readNow(ch <-chan string) string {
    	select {
    	case value, ok := <-ch:
    		if !ok {
    			return "closed"
    		}
    		return value
    	default:
    		return "no update"
    	}
    }
    func main() {
    	var b Broker
    	dashboard, stopDashboard := b.Subscribe()
    	defer stopDashboard()
    	preview, stopPreview := b.Subscribe()
    	defer stopPreview()
    	fmt.Println("accepted:", b.Publish("A1 delivered"))
    	fmt.Println("dashboard:", readNow(dashboard))
    	fmt.Println("preview:", readNow(preview))
    }
  tests: |
    package main

    import (
    	"sync"
    	"sync/atomic"
    	"testing"
    )

    func take(t *testing.T, ch <-chan string, want string) {
    	t.Helper()
    	select {
    	case got, ok := <-ch:
    		if !ok || got != want {
    			t.Fatalf("receive = %q, %v; want %q, true", got, ok, want)
    		}
    	default:
    		t.Fatalf("missing event %q", want)
    	}
    }
    func TestIndependentSubscribers(t *testing.T) {
    	var b Broker
    	if got := b.Publish("nobody"); got != 0 {
    		t.Fatalf("empty broker accepted %d", got)
    	}
    	a, stopA := b.Subscribe()
    	defer stopA()
    	c, stopC := b.Subscribe()
    	defer stopC()
    	if got := b.Publish("first"); got != 2 {
    		t.Fatalf("accepted %d, want 2", got)
    	}
    	take(t, a, "first")
    	if got := b.Publish("second"); got != 1 {
    		t.Fatalf("accepted %d with one full subscriber, want 1", got)
    	}
    	take(t, a, "second")
    	take(t, c, "first")
    }
    func TestUnsubscribe(t *testing.T) {
    	var b Broker
    	a, stopA := b.Subscribe()
    	b.Publish("queued")
    	stopA()
    	stopA()
    	take(t, a, "queued")
    	select {
    	case _, ok := <-a:
    		if ok {
    			t.Fatal("subscription must be closed")
    		}
    	default:
    		t.Fatal("unsubscribe did not close channel")
    	}
    	if got := b.Publish("later"); got != 0 {
    		t.Fatalf("publish after unsubscribe = %d, want 0", got)
    	}
    }
    func TestConcurrentPublish(t *testing.T) {
    	var b Broker
    	a, stopA := b.Subscribe()
    	defer stopA()
    	c, stopC := b.Subscribe()
    	defer stopC()
    	var accepted atomic.Int64
    	var wg sync.WaitGroup
    	for range 40 {
    		wg.Go(func() { accepted.Add(int64(b.Publish("update"))) })
    	}
    	wg.Wait()
    	if got := accepted.Load(); got != 2 {
    		t.Fatalf("accepted %d, want two total with no readers", got)
    	}
    	take(t, a, "update")
    	take(t, c, "update")
    }
    func TestConcurrentMembership(t *testing.T) {
    	var b Broker
    	var wg sync.WaitGroup
    	for range 30 {
    		wg.Go(func() {
    			_, stop := b.Subscribe()
    			b.Publish("event")
    			stop()
    			stop()
    		})
    	}
    	wg.Wait()
    	if got := b.Publish("empty again"); got != 0 {
    		t.Fatalf("accepted %d after all unsubscribed", got)
    	}
    }

---

Dispatchly's dashboard and audit preview both want delivery updates. Putting them
on one channel divides the updates between consumers. A **publish/subscribe** system
gives each subscriber its own queue, so a publication can reach both.

## Define delivery before writing code

For this lesson, each subscription has a one-event buffer. If that buffer is full,
a new event is dropped for that subscriber. Other subscribers still receive it.
This is suitable for optional notifications; it isn't suitable for durable billing
or an audit log that promises to preserve every event.

The broker owns the subscription map, sends, and closes. Subscribers receive through
receive-only channels. Serialize publishing and unsubscribe with a mutex so a send
cannot race with closing its destination. Keep sends nonblocking while holding that
mutex: a slow subscriber mustn't freeze membership changes.

```go
select {
case sub <- event:
	// Queued for this subscriber.
default:
	// Queue full: drop this event for this subscriber.
}
```

This deliberately drops the **new** event, retaining the queued one. A latest-value
feed would use a different policy. Bounded buffering makes overload visible instead
of hiding an indefinitely growing queue in memory.

## Your task

Complete `Publish` in the editor's `Broker`:

- Lock the broker while inspecting subscriptions and sending.
- Attempt one nonblocking send to every subscription.
- Return how many subscriber queues accepted the event.

`Subscribe` and its returned unsubscribe function are provided. Unsubscribing removes
and closes that channel exactly once, even if called repeatedly. Previously buffered
values remain readable before a receive reports that the channel is closed.
The zero-value broker works: publishing before any subscription returns zero.

The tests cover two independent subscribers, a slow subscriber, repeated unsubscribe,
and concurrent publishers. **Run** shows the number of accepted deliveries and reads
both queues without blocking, so an unfinished implementation still gives useful output.

## Lifetime and ordering

The broker starts no goroutines of its own. Subscription owners must call unsubscribe
when finished, otherwise the broker retains their queues. A production API may also
need a whole-broker shutdown operation.

Concurrent calls to `Publish` are ordered by lock acquisition, not by the wall-clock
time the calls began. Every queued event follows that order, but individual subscribers
can observe gaps because of drops. If delivery must survive process failure, move to
a durable messaging design with acknowledgments and replay; adding a larger buffer
can't provide those guarantees.

Further reading: [Go specification: select](https://go.dev/ref/spec#Select_statements).
