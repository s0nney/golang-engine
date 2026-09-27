---
title: Amortized Append
quiz:
  - question: What does "append is amortized O(1)" mean?
    options:
      - text: Every single append takes exactly the same time
      - text: Some appends are O(n) because they copy, but averaged over many appends the cost per append is constant
        correct: true
      - text: Append is O(1) only if you preallocate
    explanation: |
      Occasionally append has to allocate and copy everything, which is O(n). But
      because capacity grows by a multiple each time, those copies are rare
      enough that the *total* work for n appends is O(n), or O(1) each on
      average.
  - question: You're about to append exactly 1,000,000 follower counts to an empty slice. What's the idiomatic way to avoid all the regrowth?
    options:
      - text: '`s := make([]int, 0, 1_000_000)` and then append'
        correct: true
      - text: '`s := make([]int, 1_000_000)` and then append'
      - text: '`var s [1_000_000]int`'
    explanation: |
      `make([]int, 0, n)` gives length 0 and capacity n, so all the appends fit
      without reallocating. `make([]int, n)` has *length* n already, so appending
      would add a million more after a million zeros.
---

Last lesson you saw that `append` sometimes copies the whole slice into a new
array. That sounds expensive. If appending one item is O(n) in the worst case,
isn't building a slice of `n` items O(n²)?

It isn't, thanks to a trick called **geometric growth**.

## Watching capacity grow

Let's append a thousand follower counts and print every time the capacity
changes:

```go
package main

import "fmt"

func main() {
	var counts []int
	lastCap := -1
	for i := range 1000 {
		counts = append(counts, i)
		if cap(counts) != lastCap {
			fmt.Printf("len=%-4d cap=%d\n", len(counts), cap(counts))
			lastCap = cap(counts)
		}
	}
}
```

With Go 1.27 on a 64-bit machine this prints:

```
len=1    cap=4
len=5    cap=8
len=9    cap=16
len=17   cap=32
len=33   cap=64
len=65   cap=128
len=129  cap=256
len=257  cap=512
len=513  cap=848
len=849  cap=1280
```

(The exact numbers depend on the Go version and the element size, because the
runtime rounds allocations up to fit its memory size classes.)

The pattern: while the slice is small, capacity **doubles** each time it fills
up. Past 256 elements, the growth factor eases off towards 1.25× so big slices
don't waste too much memory. Either way, it grows by a **multiple**, not by a
fixed amount. Only 10 allocations happened for 1,000 appends.

## Why doubling makes it O(1) on average

Suppose capacity doubles every time. To reach `n` elements, the copies along
the way were of size 1, 2, 4, 8, ..., n/2. Add them up:

```
1 + 2 + 4 + ... + n/2  <  n
```

So the *total* copying for all `n` appends is less than `n` element moves, plus
`n` writes for the appends themselves. That's O(n) total work, which is **O(1)
per append on average**. We say append is **amortized O(1)**: the occasional
expensive call is paid for by all the cheap ones around it, like a yearly
subscription averaged over 12 months.

If Go grew slices by a fixed amount instead (say, 10 more slots each time),
there'd be n/10 reallocations each copying up to n elements: O(n²) total. The
multiplier is what makes it work.

## Preallocate when you know the size

Amortized O(1) is good. Zero reallocations is better. If you know roughly how
many items you'll append, say so up front:

```go
func followerCounts(infs []Influencer) []int {
	counts := make([]int, 0, len(infs)) // len 0, cap n
	for _, inf := range infs {
		counts = append(counts, inf.Followers)
	}
	return counts
}
```

One allocation, no copying. If you already have a slice and are about to add
many more items, `slices.Grow(s, n)` ensures room for `n` more.

Don't confuse the two forms of `make`:

- `make([]int, 0, n)`: length 0, capacity n. Append fills it.
- `make([]int, n)`: length n (all zeros). Assign by index, `counts[i] = ...`.
  Appending to this adds *after* the zeros, which is a common bug.

## Amortized isn't worst case

For most code, amortized O(1) is all you need. But notice that any *single*
append can still take O(n) time. For something like a real-time game loop that
must never stutter, that one slow append could matter, and preallocating
removes the surprise.

This idea, "occasionally expensive, cheap on average", will come back when we
build stacks on top of slices in the next chapter.

## Further reading

- [Go Slices: usage and internals](https://go.dev/blog/slices-intro), which covers growing slices with `copy` and `append`.
