---
title: Kitchen Batches
difficulty: easy
after: channels-in-depth
hints:
  - '`for id := range in` reads until `in` is closed. Append each ID to the current batch, and send the batch once it reaches `size`.'
  - 'After the loop, send whatever is left over (if anything), then close `out`. `defer close(out)` at the top of the goroutine makes the closing hard to forget.'
  - 'After sending a batch, start the next one with a **fresh** slice (`batch = nil`), not `batch[:0]`. The receiver still holds the old slice, and `batch[:0]` would write the next IDs into the same backing array under its feet.'
exercise:
  starter: |
    package main

    import "fmt"

    // batches groups the order IDs from in into slices of size IDs each and
    // sends them on the returned channel. When in is closed, it sends the
    // final, smaller batch (if there is one) and closes the returned channel.
    func batches(in <-chan string, size int) <-chan []string {
    	out := make(chan []string)
    	// TODO: in a goroutine, read in, send full batches on out, then send
    	// the leftovers and close out. For now it just closes out.
    	close(out)
    	return out
    }

    func main() {
    	in := make(chan string)
    	go func() {
    		for _, id := range []string{"A1", "A2", "A3", "A4", "A5"} {
    			in <- id
    		}
    		close(in)
    	}()
    	n := 0
    	for b := range batches(in, 2) {
    		fmt.Println(b)
    		n++
    	}
    	fmt.Println("the kitchen got", n, "batches")
    	// want:
    	// [A1 A2]
    	// [A3 A4]
    	// [A5]
    	// the kitchen got 3 batches
    }
  solution: |
    package main

    import "fmt"

    // batches groups the order IDs from in into slices of size IDs each and
    // sends them on the returned channel. When in is closed, it sends the
    // final, smaller batch (if there is one) and closes the returned channel.
    func batches(in <-chan string, size int) <-chan []string {
    	out := make(chan []string)
    	go func() {
    		defer close(out)
    		var batch []string
    		for id := range in {
    			batch = append(batch, id)
    			if len(batch) == size {
    				out <- batch
    				batch = nil // a fresh slice: the receiver owns the old one
    			}
    		}
    		if len(batch) > 0 {
    			out <- batch
    		}
    	}()
    	return out
    }

    func main() {
    	in := make(chan string)
    	go func() {
    		for _, id := range []string{"A1", "A2", "A3", "A4", "A5"} {
    			in <- id
    		}
    		close(in)
    	}()
    	n := 0
    	for b := range batches(in, 2) {
    		fmt.Println(b)
    		n++
    	}
    	fmt.Println("the kitchen got", n, "batches")
    }
  tests: |
    package main

    import (
    	"fmt"
    	"slices"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // feed returns a closed channel holding ids.
    func feed(ids []string) <-chan string {
    	in := make(chan string, len(ids))
    	for _, id := range ids {
    		in <- id
    	}
    	close(in)
    	return in
    }

    // collect reads every batch, failing if out isn't closed within an hour
    // of fake time.
    func collect(t *testing.T, out <-chan []string) [][]string {
    	t.Helper()
    	var got [][]string
    	for {
    		select {
    		case b, ok := <-out:
    			if !ok {
    				return got
    			}
    			got = append(got, b)
    		case <-time.After(time.Hour):
    			t.Fatalf("after batches %q the output channel was never closed: close it once the input is closed", got)
    		}
    	}
    }

    func TestBatches(t *testing.T) {
    	tests := []struct {
    		ids  []string
    		size int
    		want [][]string
    	}{
    		{[]string{"A1", "A2", "A3", "A4", "A5"}, 2, [][]string{{"A1", "A2"}, {"A3", "A4"}, {"A5"}}},
    		{[]string{"A1", "A2", "A3", "A4"}, 2, [][]string{{"A1", "A2"}, {"A3", "A4"}}},
    		{[]string{"A1", "A2"}, 5, [][]string{{"A1", "A2"}}},
    		{[]string{"A1", "A2", "A3"}, 1, [][]string{{"A1"}, {"A2"}, {"A3"}}},
    		{[]string{"A1", "A2", "A3", "A4", "A5", "A6", "A7"}, 3, [][]string{{"A1", "A2", "A3"}, {"A4", "A5", "A6"}, {"A7"}}},
    		{nil, 3, nil},
    	}
    	for _, tt := range tests {
    		synctest.Test(t, func(t *testing.T) {
    			out := batches(feed(tt.ids), tt.size)
    			if out == nil {
    				t.Fatalf("batches returned a nil channel")
    			}
    			got := collect(t, out)
    			if !slices.EqualFunc(got, tt.want, slices.Equal) {
    				t.Errorf("batches(%q, %d) sent %q, want %q", tt.ids, tt.size, got, tt.want)
    			}
    		})
    	}
    }

    // TestBatchesAreIndependent holds on to every batch before comparing, so
    // batches that share a backing array get overwritten.
    func TestBatchesAreIndependent(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		var ids []string
    		var want [][]string
    		for i := range 12 {
    			ids = append(ids, fmt.Sprint("B", i))
    			if i%4 == 3 {
    				want = append(want, ids[i-3:i+1])
    			}
    		}
    		got := collect(t, batches(feed(ids), 4))
    		if !slices.EqualFunc(got, want, slices.Equal) {
    			t.Errorf("batches of 4 from 12 IDs, all read before checking: got %q, want %q (each batch needs its own slice; don't reuse the old one with batch[:0])", got, want)
    		}
    	})
    }
---

Dispatchly's kitchens don't want orders trickling in one at a time. They want
them in **batches**, so the cook can plan a few tickets together.

Complete `batches(in, size)`. It returns a channel and, in a goroutine, reads
the order IDs from `in`:

- Every time it has collected `size` IDs, it sends them on the returned
  channel as one `[]string` batch, in the order they arrived.
- When `in` is closed, it sends the final, smaller batch (only if it has at
  least one ID) and then **closes** the returned channel.

## Example

```
in:     A1 A2 A3 A4 A5 (then closed)
size:   2
output: [A1 A2] [A3 A4] [A5] (then closed)
```

## Constraints

- `size` ≥ 1. `in` may be closed without sending anything, in which case the
  output channel is closed without sending a batch.
- A receiver may keep every batch it gets, so each batch must be its own slice
  that you never write to again after sending it.
- The tests use a `synctest` bubble, so an output channel that's never closed
  is reported straight away instead of hanging.
