---
title: Iterator APIs for Collections
quiz:
  - question: |
      A concurrent Stash collection has this method. What happens if the loop body calls `c.Put(...)`?

      ```go
      func (c *SafeCache[K, V]) All() iter.Seq2[K, V] {
      	return func(yield func(K, V) bool) {
      		c.mu.Lock()
      		defer c.mu.Unlock()
      		for k, v := range c.items {
      			if !yield(k, v) {
      				return
      			}
      		}
      	}
      }
      ```
    options:
      - text: It works; `Put` waits until the loop finishes
      - text: 'It deadlocks: the loop body runs *inside* `yield`, while `All` still holds the mutex that `Put` needs'
        correct: true
      - text: It panics with "concurrent map writes"
      - text: '`Put` is silently ignored'
    explanation: |
      With range-over-func, the loop body *is* the `yield` call, so it runs with the lock
      held. `sync.Mutex` isn't reentrant, so `Put` blocks forever. Either document "don't
      modify during iteration" or copy a snapshot under the lock and yield from that.
  - question: When do you need `iter.Pull` instead of a plain `for range`?
    options:
      - text: Whenever an iterator returns pairs
      - text: When you must advance two (or more) sequences in step, such as merging or zipping them
        correct: true
      - text: To make an iterator run faster
      - text: To iterate a slice backwards
    explanation: |
      `range` drives one sequence at a time, and the iterator is in control. `iter.Pull`
      turns a sequence into a `next()` function you call when *you* want the next value,
      which is what walking two sequences side by side requires. Always call its `stop`.
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"iter"
    	"slices"
    )

    // Merge yields the values of two sorted sequences in sorted order.
    // On ties, the value from a comes first. It must stop pulling from a and b
    // as soon as the caller stops (breaks out of the loop).
    func Merge[T cmp.Ordered](a, b iter.Seq[T]) iter.Seq[T] {
    	return func(yield func(T) bool) {
    		// ? Use iter.Pull on both sequences, and defer both stop functions.
    	}
    }

    func main() {
    	hot := slices.Values([]string{"api", "db", "web"})
    	cold := slices.Values([]string{"archive", "backup", "logs"})
    	fmt.Println(slices.Collect(Merge(hot, cold)))
    	for v := range Merge(slices.Values([]int{1, 4, 9}), slices.Values([]int{2, 3, 10})) {
    		if v > 3 {
    			break
    		}
    		fmt.Print(v, " ")
    	}
    	fmt.Println()
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"iter"
    	"slices"
    )

    // Merge yields the values of two sorted sequences in sorted order.
    // On ties, the value from a comes first. It must stop pulling from a and b
    // as soon as the caller stops (breaks out of the loop).
    func Merge[T cmp.Ordered](a, b iter.Seq[T]) iter.Seq[T] {
    	return func(yield func(T) bool) {
    		nextA, stopA := iter.Pull(a)
    		defer stopA()
    		nextB, stopB := iter.Pull(b)
    		defer stopB()

    		va, okA := nextA()
    		vb, okB := nextB()
    		for okA || okB {
    			if okA && (!okB || !cmp.Less(vb, va)) {
    				if !yield(va) {
    					return
    				}
    				va, okA = nextA()
    			} else {
    				if !yield(vb) {
    					return
    				}
    				vb, okB = nextB()
    			}
    		}
    	}
    }

    func main() {
    	hot := slices.Values([]string{"api", "db", "web"})
    	cold := slices.Values([]string{"archive", "backup", "logs"})
    	fmt.Println(slices.Collect(Merge(hot, cold)))
    	for v := range Merge(slices.Values([]int{1, 4, 9}), slices.Values([]int{2, 3, 10})) {
    		if v > 3 {
    			break
    		}
    		fmt.Print(v, " ")
    	}
    	fmt.Println()
    }
  tests: |
    package main

    import (
    	"iter"
    	"slices"
    	"testing"
    )

    func TestMerge(t *testing.T) {
    	for _, tt := range []struct {
    		a, b, want []int
    	}{
    		{[]int{1, 4, 9}, []int{2, 3, 10, 11}, []int{1, 2, 3, 4, 9, 10, 11}},
    		{[]int{}, []int{5, 6}, []int{5, 6}},
    		{[]int{7}, nil, []int{7}},
    		{nil, nil, nil},
    		{[]int{1, 2, 2}, []int{2, 3}, []int{1, 2, 2, 2, 3}},
    	} {
    		got := slices.Collect(Merge(slices.Values(tt.a), slices.Values(tt.b)))
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("Merge(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
    		}
    	}
    }

    func TestMergeTiesPreferA(t *testing.T) {
    	a := slices.Values([]string{"b", "d"})
    	b := slices.Values([]string{"b", "c"})
    	got := slices.Collect(Merge(a, b))
    	if !slices.Equal(got, []string{"b", "b", "c", "d"}) {
    		t.Errorf("Merge([b d], [b c]) = %v, want [b b c d]", got)
    	}
    }

    // counting yields 0, 1, 2, ... up to n-1 and records how far it got
    // and whether it finished (its deferred cleanup ran).
    func counting(n int, produced *int, cleaned *bool) iter.Seq[int] {
    	return func(yield func(int) bool) {
    		defer func() { *cleaned = true }()
    		for i := range n {
    			*produced = i + 1
    			if !yield(i * 2) {
    				return
    			}
    		}
    	}
    }

    func TestMergeStopsEarly(t *testing.T) {
    	var pa, pb int
    	var ca, cb bool
    	var got []int
    	for v := range Merge(counting(1000, &pa, &ca), counting(1000, &pb, &cb)) {
    		got = append(got, v)
    		if len(got) == 3 {
    			break
    		}
    	}
    	if !slices.Equal(got, []int{0, 0, 2}) {
    		t.Errorf("first 3 merged values = %v, want [0 0 2]", got)
    	}
    	if pa > 5 || pb > 5 {
    		t.Errorf("after break, sources produced %d and %d values; Merge should stop pulling", pa, pb)
    	}
    	if !ca || !cb {
    		t.Errorf("source iterators weren't cleaned up after break (cleaned: %v, %v); defer the stop functions from iter.Pull", ca, cb)
    	}
    }
---

Every Stash collection exposes its contents as iterators. You learned to *write* iterators in [Learn Functional Programming](/courses/learn-functional-programming/iterators/writing-iterators); this lesson is about designing them as part of a collection's API.

## Follow the standard names

The standard library has settled on a vocabulary. Match it and your types feel native:

| Method | Type | Meaning |
|---|---|---|
| `All()` | `iter.Seq2[K, V]` or `iter.Seq[T]` | everything, in the collection's natural order |
| `Keys()` / `Values()` | `iter.Seq[K]` / `iter.Seq[V]` | one side of a keyed collection |
| `Backward()` | `iter.Seq2[int, E]` or similar | reverse order |

`slices.All` yields index-value pairs, `maps.All` key-value pairs, and so on. Stash's `Set.All()` yields values (a set has no keys), `OrderedMap.All()` yields key-value pairs in insertion order, and `LRU.Keys()` yields keys from newest to oldest.

Return the iterator from a **method**, not as a field, and create it fresh on each call. Callers can then `range` over `m.All()` as many times as they like.

## Decide what mutation means

What should happen if someone modifies a collection while ranging over it? There are two honest answers:

- **Live view**: the iterator reads the collection as it goes. Cheap, but changes during iteration may or may not be seen. Go's own maps work this way (with rules in the spec about which changes are visible).
- **Snapshot**: copy what you'll yield up front, then yield from the copy. Costs an allocation, but the loop sees a consistent picture.

Either is fine. What's not fine is leaving it undefined. Say which in the doc comment.

## Locks and yield don't mix

Remember what range-over-func really does: the loop body **is** the `yield` function. For a mutex-protected collection, that means code like the quiz's `All` runs the caller's loop body *while holding the lock*. If the body touches the collection, it deadlocks. If it's slow, every other goroutine waits.

The usual fix is a snapshot under the lock:

```go
func (c *SafeCache[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		c.mu.Lock()
		keys := slices.Collect(maps.Keys(c.items))
		c.mu.Unlock()
		for _, k := range keys { // lock released: the body can do anything
			if !yield(k) {
				return
			}
		}
	}
}
```

## Push and pull

A normal iterator is **push**-style: it calls `yield` and stays in control. That's perfect for one sequence at a time. But some operations need to advance **two** sequences at their own pace: merging two sorted streams, zipping, comparing element by element. For that, `iter.Pull` converts a push iterator into a **pull** function:

```go
next, stop := iter.Pull(seq)
defer stop()
v, ok := next() // ok is false once seq is exhausted
```

Behind the scenes it runs the iterator in a coroutine, handing you one value per `next()` call. The `stop` function matters: if you stop pulling before the sequence is exhausted, `stop` tells the iterator to finish, which runs any cleanup (deferred calls) it has. Always `defer stop()`.

## Your turn

Stash stores sorted runs of keys and needs to combine them. Write `Merge(a, b)`, which yields the values of two **sorted** sequences in sorted order:

- Pull from both with `iter.Pull`, and `defer` both `stop` functions.
- Keep the current value of each; yield the smaller one (on ties, `a`'s first) and advance only that side.
- When one side runs out, yield the rest of the other.
- If `yield` returns `false`, return immediately; the deferred `stop`s clean up the sources.
