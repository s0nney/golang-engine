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
exercise:
  starter: |
    package main

    import "fmt"

    // Counts is a growable list of follower counts that manages its own
    // backing array, the way append does behind the scenes.
    // The zero value is an empty list.
    type Counts struct {
    	data   []int // backing array: len(data) is the capacity
    	n      int   // how many values are stored
    	Copies int   // total values copied while growing
    }

    func (c *Counts) Len() int      { return c.n }
    func (c *Counts) Cap() int      { return len(c.data) }
    func (c *Counts) Get(i int) int { return c.data[i] }

    // resize moves the values into a new backing array of the given capacity.
    func (c *Counts) resize(capacity int) {
    	bigger := make([]int, capacity)
    	c.Copies += copy(bigger, c.data[:c.n])
    	c.data = bigger
    }

    // Push adds v to the end. When the array is full it must grow
    // geometrically: double the capacity (or 1 if it was 0).
    func (c *Counts) Push(v int) {
    	if c.n == len(c.data) {
    		c.resize(len(c.data) + 1) // grows one slot at a time: O(n²) in total!
    	}
    	c.data[c.n] = v
    	c.n++
    }

    // Grow makes sure there is room for at least extra more values, growing
    // at most once. If there's already room, it does nothing.
    func (c *Counts) Grow(extra int) {
    	// ?
    }

    func main() {
    	var c Counts
    	for i := range 1000 {
    		c.Push(i)
    	}
    	fmt.Println("len", c.Len(), "cap", c.Cap(), "copies", c.Copies) // want: len 1000 cap 1024 copies 1023

    	var d Counts
    	d.Grow(500)
    	for i := range 500 {
    		d.Push(i)
    	}
    	fmt.Println("len", d.Len(), "cap", d.Cap(), "copies", d.Copies) // want: len 500 cap 500 copies 0
    }
  solution: |
    package main

    import "fmt"

    // Counts is a growable list of follower counts that manages its own
    // backing array, the way append does behind the scenes.
    // The zero value is an empty list.
    type Counts struct {
    	data   []int // backing array: len(data) is the capacity
    	n      int   // how many values are stored
    	Copies int   // total values copied while growing
    }

    func (c *Counts) Len() int      { return c.n }
    func (c *Counts) Cap() int      { return len(c.data) }
    func (c *Counts) Get(i int) int { return c.data[i] }

    // resize moves the values into a new backing array of the given capacity.
    func (c *Counts) resize(capacity int) {
    	bigger := make([]int, capacity)
    	c.Copies += copy(bigger, c.data[:c.n])
    	c.data = bigger
    }

    // Push adds v to the end. When the array is full it must grow
    // geometrically: double the capacity (or 1 if it was 0).
    func (c *Counts) Push(v int) {
    	if c.n == len(c.data) {
    		c.resize(max(1, 2*len(c.data)))
    	}
    	c.data[c.n] = v
    	c.n++
    }

    // Grow makes sure there is room for at least extra more values, growing
    // at most once. If there's already room, it does nothing.
    func (c *Counts) Grow(extra int) {
    	if c.n+extra > len(c.data) {
    		c.resize(c.n + extra)
    	}
    }

    func main() {
    	var c Counts
    	for i := range 1000 {
    		c.Push(i)
    	}
    	fmt.Println("len", c.Len(), "cap", c.Cap(), "copies", c.Copies)

    	var d Counts
    	d.Grow(500)
    	for i := range 500 {
    		d.Push(i)
    	}
    	fmt.Println("len", d.Len(), "cap", d.Cap(), "copies", d.Copies)
    }
  tests: |
    package main

    import "testing"

    func TestPushKeepsValues(t *testing.T) {
    	var c Counts
    	for i := range 100 {
    		c.Push(i * 10)
    	}
    	if c.Len() != 100 {
    		t.Fatalf("after 100 pushes Len() = %d, want 100", c.Len())
    	}
    	for i := range 100 {
    		if got := c.Get(i); got != i*10 {
    			t.Fatalf("Get(%d) = %d, want %d", i, got, i*10)
    		}
    	}
    }

    func TestPushDoubles(t *testing.T) {
    	var c Counts
    	wantCaps := map[int]int{1: 1, 2: 2, 3: 4, 5: 8, 9: 16, 17: 32, 1000: 1024}
    	for i := 1; i <= 1000; i++ {
    		c.Push(i)
    		if want, ok := wantCaps[i]; ok && c.Cap() != want {
    			t.Errorf("after %d pushes Cap() = %d, want %d (double when full, starting at 1)", i, c.Cap(), want)
    		}
    	}
    	if c.Copies != 1023 {
    		t.Errorf("1000 pushes made %d copies, want 1023 (1+2+4+...+512)", c.Copies)
    	}
    }

    func TestPushIsAmortizedConstant(t *testing.T) {
    	var c Counts
    	const n = 200_000
    	for i := range n {
    		c.Push(i)
    		if c.Copies >= 2*n {
    			break // already too many: stop early so the test stays fast
    		}
    	}
    	if c.Copies >= 2*n {
    		t.Errorf("%d pushes made %d copies, want fewer than %d: is growth geometric?", n, c.Copies, 2*n)
    	}
    }

    func TestGrow(t *testing.T) {
    	var c Counts
    	c.Grow(500)
    	if c.Cap() < 500 {
    		t.Fatalf("after Grow(500) on an empty list Cap() = %d, want at least 500", c.Cap())
    	}
    	for i := range 500 {
    		c.Push(i)
    	}
    	if c.Copies != 0 {
    		t.Errorf("500 pushes after Grow(500) made %d copies, want 0", c.Copies)
    	}

    	var d Counts
    	for i := range 3 {
    		d.Push(i)
    	}
    	before := d.Copies
    	d.Grow(1) // cap is 4, len 3: already room
    	if d.Cap() != 4 || d.Copies != before {
    		t.Errorf("Grow(1) with room to spare changed Cap to %d and Copies by %d, want no change", d.Cap(), d.Copies-before)
    	}
    	d.Grow(10)
    	if d.Cap() < 13 {
    		t.Errorf("Grow(10) with len 3 gave Cap() = %d, want at least 13", d.Cap())
    	}
    	for i := range 3 {
    		if d.Get(i) != i {
    			t.Errorf("after Grow, Get(%d) = %d, want %d", i, d.Get(i), i)
    		}
    	}
    	grown := d.Copies
    	for i := range 10 {
    		d.Push(i)
    	}
    	if d.Copies != grown {
    		t.Errorf("10 pushes after Grow(10) copied %d values, want 0", d.Copies-grown)
    	}
    }
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

## Your turn: grow it yourself

`append` hides all of this from you, so let's do it by hand. `Counts` in the
editor manages its own backing array: `len(c.data)` is the capacity and `c.n`
is how many values are stored. Its `resize` helper allocates a new array and
adds the number of values it copied to `c.Copies`.

1. `Push` currently grows the array **one slot at a time**, the fixed-increment
   strategy that costs O(n²). Fix it so a full array **doubles** its capacity
   (a zero-capacity array grows to 1). After 1,000 pushes you should see a
   capacity of 1,024 and exactly 1,023 copies: 1 + 2 + 4 + ... + 512.
2. Complete `Grow(extra)`, your own `slices.Grow`: if there isn't room for
   `extra` more values, resize **once** to exactly `c.n + extra`. If there's
   already room, do nothing. Pushing `extra` values afterwards must copy nothing.

## Further reading

- [Go Slices: usage and internals](https://go.dev/blog/slices-intro), which covers growing slices with `copy` and `append`.
