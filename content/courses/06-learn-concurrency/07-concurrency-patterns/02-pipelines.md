---
title: Pipelines
quiz:
  - question: |
      In this pipeline stage, the receive from `in` is a plain `range`, not a `select` on `ctx.Done()`. Why is that usually fine?

      ```go
      for l := range in {
      	select {
      	case out <- strings.ToUpper(l):
      	case <-ctx.Done():
      		return
      	}
      }
      ```
    options:
      - text: Because receiving from a channel can never block
      - text: Because when ctx is cancelled the upstream stage returns and closes `in`, which ends the range loop
        correct: true
      - text: Because `range` checks the context automatically
      - text: It isn't fine, it always leaks
    explanation: |
      Each stage closes its output when it returns. Cancelling the context
      makes the upstream stage stop and close `in`, so this loop ends too.
      The sends are what need guarding, since the stage *downstream* may
      have stopped reading.
  - question: A pipeline has three stages that each take 10ms per item, one goroutine per stage. Roughly how long does it take to push 100 items through?
    options:
      - text: About 3 seconds (100 × 30ms)
      - text: About 1 second (the stages work on different items at the same time)
        correct: true
      - text: About 30ms
      - text: About 10ms
    explanation: |
      Like an assembly line, stage 1 works on item 3 while stage 2 works on
      item 2 and stage 3 on item 1. Once the pipeline is full, an item comes
      out every 10ms: about 100 × 10ms plus a little to fill it.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"runtime"
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

    // source emits orders 1, 2, 3, ... (every third one is empty) until ctx
    // is cancelled. It's already correct.
    func source(ctx context.Context) <-chan Order {
    	out := make(chan Order)
    	go func() {
    		defer close(out)
    		for id := 1; ; id++ {
    			select {
    			case out <- Order{ID: id, Items: id % 3}:
    			case <-ctx.Done():
    				return
    			}
    		}
    	}()
    	return out
    }

    // validate passes on orders that have at least one item and drops the rest.
    func validate(ctx context.Context, in <-chan Order) <-chan Order {
    	out := make(chan Order)
    	go func() {
    		defer close(out)
    		for o := range in {
    			if o.Items > 0 {
    				out <- o
    			}
    		}
    	}()
    	return out
    }

    // price turns each order into a quote: 199 cents plus 450 per item.
    func price(ctx context.Context, in <-chan Order) <-chan Quote {
    	out := make(chan Quote)
    	go func() {
    		defer close(out)
    		for o := range in {
    			out <- Quote{OrderID: o.ID, Cents: 199 + 450*o.Items}
    		}
    	}()
    	return out
    }

    func main() {
    	before := runtime.NumGoroutine()
    	ctx, cancel := context.WithCancel(context.Background())

    	quotes := price(ctx, validate(ctx, source(ctx)))
    	for range 4 {
    		fmt.Println(<-quotes)
    	}
    	time.Sleep(10 * time.Millisecond) // meanwhile, the stages fill up with more work
    	cancel()                          // we have enough quotes: stop the pipeline

    	time.Sleep(50 * time.Millisecond)
    	fmt.Println("goroutines left behind:", runtime.NumGoroutine()-before)
    }
  solution: |
    package main

    import (
    	"context"
    	"fmt"
    	"runtime"
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

    // source emits orders 1, 2, 3, ... (every third one is empty) until ctx
    // is cancelled. It's already correct.
    func source(ctx context.Context) <-chan Order {
    	out := make(chan Order)
    	go func() {
    		defer close(out)
    		for id := 1; ; id++ {
    			select {
    			case out <- Order{ID: id, Items: id % 3}:
    			case <-ctx.Done():
    				return
    			}
    		}
    	}()
    	return out
    }

    // validate passes on orders that have at least one item and drops the rest.
    func validate(ctx context.Context, in <-chan Order) <-chan Order {
    	out := make(chan Order)
    	go func() {
    		defer close(out)
    		for o := range in {
    			if o.Items == 0 {
    				continue
    			}
    			select {
    			case out <- o:
    			case <-ctx.Done():
    				return
    			}
    		}
    	}()
    	return out
    }

    // price turns each order into a quote: 199 cents plus 450 per item.
    func price(ctx context.Context, in <-chan Order) <-chan Quote {
    	out := make(chan Quote)
    	go func() {
    		defer close(out)
    		for o := range in {
    			select {
    			case out <- Quote{OrderID: o.ID, Cents: 199 + 450*o.Items}:
    			case <-ctx.Done():
    				return
    			}
    		}
    	}()
    	return out
    }

    func main() {
    	before := runtime.NumGoroutine()
    	ctx, cancel := context.WithCancel(context.Background())

    	quotes := price(ctx, validate(ctx, source(ctx)))
    	for range 4 {
    		fmt.Println(<-quotes)
    	}
    	time.Sleep(10 * time.Millisecond) // meanwhile, the stages fill up with more work
    	cancel()                          // we have enough quotes: stop the pipeline

    	time.Sleep(50 * time.Millisecond)
    	fmt.Println("goroutines left behind:", runtime.NumGoroutine()-before)
    }
  tests: |
    package main

    import (
    	"context"
    	"runtime"
    	"slices"
    	"testing"
    	"time"
    )

    func from(orders ...Order) <-chan Order {
    	ch := make(chan Order, len(orders))
    	for _, o := range orders {
    		ch <- o
    	}
    	close(ch)
    	return ch
    }

    func TestPipelineResults(t *testing.T) {
    	ctx, cancel := context.WithCancel(context.Background())
    	defer cancel()

    	in := from(Order{1, 2}, Order{2, 0}, Order{3, 1}, Order{4, 0}, Order{5, 3})
    	var got []Quote
    	for q := range price(ctx, validate(ctx, in)) {
    		got = append(got, q)
    	}
    	want := []Quote{{1, 1099}, {3, 649}, {5, 1549}}
    	if !slices.Equal(got, want) {
    		t.Errorf("pipeline produced %v, want %v", got, want)
    	}
    }

    // settled waits up to a second for the goroutine count to drop to want.
    func settled(want int) int {
    	n := runtime.NumGoroutine()
    	for deadline := time.Now().Add(time.Second); n > want && time.Now().Before(deadline); {
    		time.Sleep(5 * time.Millisecond)
    		n = runtime.NumGoroutine()
    	}
    	return n
    }

    func TestPipelineStopsWhenCancelled(t *testing.T) {
    	before := runtime.NumGoroutine()
    	ctx, cancel := context.WithCancel(context.Background())

    	quotes := price(ctx, validate(ctx, source(ctx)))
    	for range 3 {
    		<-quotes
    	}
    	time.Sleep(10 * time.Millisecond) // let every stage block on a send
    	cancel()                          // the consumer walks away without draining quotes

    	if n := settled(before); n > before {
    		t.Errorf("after cancelling, %d pipeline goroutine(s) are still running: a stage is stuck sending to a consumer that left (watch ctx.Done() when sending)", n-before)
    	}
    }

    func TestCancelledBeforeReading(t *testing.T) {
    	before := runtime.NumGoroutine()
    	ctx, cancel := context.WithCancel(context.Background())
    	_ = price(ctx, validate(ctx, source(ctx))) // nobody ever reads
    	time.Sleep(10 * time.Millisecond)
    	cancel()

    	if n := settled(before); n > before {
    		t.Errorf("a pipeline that was cancelled before anyone read from it left %d goroutine(s) running", n-before)
    	}
    }
---

A **pipeline** is a series of **stages** connected by channels. Each stage is a goroutine (or several) that receives values from upstream, does one job, and sends results downstream. It's an assembly line: while one stage is pricing order 1, the one before it is already validating order 2.

## Each stage has the same shape

A stage is just a generator whose input is another channel:

```go
func stage(ctx context.Context, in <-chan A) <-chan B {
	out := make(chan B)
	go func() {
		defer close(out)
		for a := range in {
			select {
			case out <- transform(a):
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
```

Because every stage takes a `<-chan` and returns a `<-chan`, stages snap together like Lego:

```go
package main

import (
	"context"
	"fmt"
	"strings"
)

type Quote struct {
	Order string
	Cents int
}

// Stage 1: emit raw order lines.
func lines(ctx context.Context, raw []string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, l := range raw {
			select {
			case out <- l:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Stage 2: drop blank lines and normalise the rest.
func clean(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for l := range in {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			select {
			case out <- strings.ToUpper(l):
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Stage 3: price each order.
func price(ctx context.Context, in <-chan string) <-chan Quote {
	out := make(chan Quote)
	go func() {
		defer close(out)
		for id := range in {
			select {
			case out <- Quote{Order: id, Cents: 299 + 100*len(id)}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	raw := []string{" a1 ", "", "b22", "  ", "c333"}
	for q := range price(ctx, clean(ctx, lines(ctx, raw))) {
		fmt.Printf("%s: $%d.%02d\n", q.Order, q.Cents/100, q.Cents%100)
	}
}
```

```text
A1: $4.99
B22: $5.99
C333: $6.99
```

## The two rules that make it safe

**1. Close your output when you're done.** `defer close(out)` in every stage. When the source runs out, it closes its channel, which ends the next stage's `range`, which closes *its* output, and so on. The shutdown ripples down the line until the consumer's `range` ends.

**2. Every send watches `ctx.Done()`.** If the consumer stops early (it found what it wanted, or it hit an error), it cancels the context. A stage blocked on `out <- ...` would otherwise wait forever, and so would every stage behind it. With the `select`, each stage returns, closes its output, and the whole line drains.

Receives usually don't need a `select`: when the context is cancelled, the upstream stage returns and closes the channel, which ends the `range`. Only the *first* stage, which has no upstream, needs to watch the context on its own.

## Pipelines vs one big loop

Why not write one loop that cleans and prices each line? For cheap, CPU-only steps, you should: a pipeline adds a channel handoff per item per stage. Pipelines pay off when stages are **slow in different ways**, for example one waits on a database while another waits on a maps API. Then the stages overlap their waiting, and you can give a slow stage more goroutines (next lesson) without touching the others.

## Your turn

Dispatchly's quoting pipeline has three stages: `source` emits orders, `validate` drops empty ones, and `price` turns orders into quotes. `source` is already correct. `validate` and `price` work fine when the pipeline runs to the end, but the consumer in `main` takes four quotes and then cancels. Press **Run**: goroutines are left behind, blocked forever on a send.

Fix `validate` and `price` so that every stage exits promptly once `ctx` is cancelled, even if nobody ever reads from its output again. Their results must stay the same.
