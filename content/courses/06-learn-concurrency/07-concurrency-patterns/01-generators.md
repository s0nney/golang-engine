---
title: Generators
quiz:
  - question: |
      What's wrong with this generator?

      ```go
      func couriersNearby(zone string) <-chan string {
      	out := make(chan string)
      	go func() {
      		for _, c := range scan(zone) {
      			out <- c
      		}
      	}()
      	return out
      }
      ```
    options:
      - text: It returns the channel before the goroutine has sent anything
      - text: It never closes `out`, so a caller ranging over it never finishes, and it can't be stopped early
        correct: true
      - text: The return type should be `chan string`
      - text: '`range` over a slice inside a goroutine is a data race'
    explanation: |
      Returning immediately is the whole point of a generator. But without
      `defer close(out)` a caller's `for range` never ends, and without a
      `ctx` (or done channel) a caller that stops reading early leaves the
      goroutine blocked forever on `out <- c`.
  - question: When is a range-over-func iterator (`iter.Seq`) a better choice than a channel-based generator?
    options:
      - text: When producing each value needs to happen in parallel with consuming it
      - text: When several goroutines produce values at once
      - text: When you just want to produce values lazily in sequence, with no concurrency needed
        correct: true
      - text: Never, channels are always preferred in Go
    explanation: |
      An iterator is a plain function call per value: no goroutine, no
      channel, no leak if the loop breaks early. Channels earn their cost
      when production is genuinely concurrent, such as a goroutine doing
      slow I/O while the consumer works.
---

You've built every piece of Go's concurrency toolkit. This chapter assembles them into **patterns**: shapes that come up again and again. The simplest is the **generator**: a function that starts a goroutine to produce values and returns a channel to receive them from.

## The shape

```go
package main

import (
	"context"
	"fmt"
)

// orderIDs generates order IDs starting at start, until ctx is cancelled.
func orderIDs(ctx context.Context, start int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for id := start; ; id++ {
			select {
			case <-ctx.Done():
				return
			case out <- id:
			}
		}
	}()
	return out
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	ids := orderIDs(ctx, 1001)

	fmt.Println(<-ids, <-ids, <-ids)

	cancel() // the generator's goroutine returns and closes ids
	for range ids {
		// drain anything it sent before it noticed
	}
	fmt.Println("generator stopped")
}
```

```text
1001 1002 1003
generator stopped
```

Every generator follows the same checklist, and each item fixes a bug you've already met:

1. **Takes a `ctx` first** so the caller can stop it at any time.
2. **Creates the channel, starts the goroutine, returns immediately.** The caller can start receiving straight away.
3. **Returns `<-chan T`**, so the caller can't send on it or close it.
4. **`defer close(out)`**, so a caller's `range` ends when the generator does.
5. **Every send is a `select` with `<-ctx.Done()`**, so a caller that walks away never leaves the goroutine stuck on a send.

## Why the drain loop?

After `cancel()`, `main` ranges over `ids` until it's closed. That isn't needed to avoid a leak: the generator's `select` will see `ctx.Done()`. But it's a handy way to *wait* until the goroutine has really finished. Once `ids` is closed, the deferred `close` has run, so the goroutine is on its way out. Tests use this to make sure nothing is left running.

## Wrapping slow I/O

Generators shine when producing each value is slow, because the goroutine can fetch the *next* value while the consumer works on the current one. Here's a generator that pages through a restaurant API:

```go
func menuItems(ctx context.Context, restaurant string) <-chan MenuItem {
	out := make(chan MenuItem)
	go func() {
		defer close(out)
		for page := 1; ; page++ {
			items, more := fetchPage(ctx, restaurant, page) // slow HTTP call
			for _, it := range items {
				select {
				case out <- it:
				case <-ctx.Done():
					return
				}
			}
			if !more {
				return
			}
		}
	}()
	return out
}
```

With a buffered `out`, the generator can even run a page ahead of the consumer.

## Errors

A channel carries one type, so how does a generator report that `fetchPage` failed? Common options:

- Send a result struct: `chan Result` where `type Result struct { Item MenuItem; Err error }`.
- Return a second value alongside the channel, like `func() error`, that the caller calls after the channel closes.
- Use the error group you'll build at the end of this chapter.

## Generators vs iterators

A channel generator costs a goroutine plus a channel operation per value, and needs a `ctx` so it can be stopped. Since Go 1.23, an `iter.Seq[T]` does *sequential* lazy generation with none of that: `break` in the caller's loop just makes `yield` return `false`. So:

- **Lazy sequence, no concurrency needed** (IDs, pages of a slice, tree nodes): use an iterator.
- **Values produced concurrently** (slow I/O in the background, several producers, feeding a pipeline): use a channel generator.

The rest of this chapter is about the second kind.

## Further reading

- [Go blog: Pipelines and cancellation](https://go.dev/blog/pipelines)
