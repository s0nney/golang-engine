---
title: cmp.Ordered and Friends
quiz:
  - question: |
      What does this print?

      ```go
      nan := math.NaN()
      xs := []float64{3, nan, 1, 2}
      slices.Sort(xs)
      fmt.Println(xs, max(1.0, nan, 3.0))
      ```
    options:
      - text: '`[1 2 3 NaN] 3`'
      - text: '`[NaN 1 2 3] NaN`'
        correct: true
      - text: '`[3 NaN 1 2] 3`'
      - text: It panics
    explanation: |
      `slices.Sort` orders floats the way `cmp.Less` does: NaN before everything else.
      The built-in `max` propagates NaN: if any argument is NaN, the result is NaN.
  - question: Which types satisfy `cmp.Ordered`?
    options:
      - text: Every comparable type
      - text: Integers and floats only
      - text: Integers, floats, strings, and any type defined from them
        correct: true
      - text: Integers, floats, strings and complex numbers
    explanation: |
      `cmp.Ordered` is the union of `~int`, `~int8`, ... `~uintptr`, `~float32`,
      `~float64` and `~string`: exactly the types that support `<`. Complex numbers
      have `==` but no ordering.
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"math"
    )

    // Clamp returns v limited to the range [lo, hi].
    func Clamp[T cmp.Ordered](v, lo, hi T) T {
    	// ?
    	return v
    }

    // MinMax returns the smallest and largest values in xs, skipping NaNs.
    // ok is false if there are no (non-NaN) values at all.
    func MinMax[T cmp.Ordered](xs []T) (lo, hi T, ok bool) {
    	// ?
    	return lo, hi, false
    }

    func main() {
    	fmt.Println(Clamp(150, 0, 100), Clamp("m", "a", "k"))
    	fmt.Println(MinMax([]float64{3, math.NaN(), -1, 2}))
    	fmt.Println(MinMax([]string{"pear", "fig", "plum"}))
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"math"
    )

    // Clamp returns v limited to the range [lo, hi].
    func Clamp[T cmp.Ordered](v, lo, hi T) T {
    	return min(max(v, lo), hi)
    }

    // MinMax returns the smallest and largest values in xs, skipping NaNs.
    // ok is false if there are no (non-NaN) values at all.
    func MinMax[T cmp.Ordered](xs []T) (lo, hi T, ok bool) {
    	for _, x := range xs {
    		if x != x { // only true for NaN
    			continue
    		}
    		if !ok {
    			lo, hi, ok = x, x, true
    			continue
    		}
    		lo = min(lo, x)
    		hi = max(hi, x)
    	}
    	return lo, hi, ok
    }

    func main() {
    	fmt.Println(Clamp(150, 0, 100), Clamp("m", "a", "k"))
    	fmt.Println(MinMax([]float64{3, math.NaN(), -1, 2}))
    	fmt.Println(MinMax([]string{"pear", "fig", "plum"}))
    }
  tests: |
    package main

    import (
    	"math"
    	"testing"
    )

    type Priority int

    func TestClamp(t *testing.T) {
    	if got := Clamp(150, 0, 100); got != 100 {
    		t.Errorf("Clamp(150, 0, 100) = %d, want 100", got)
    	}
    	if got := Clamp(-5, 0, 100); got != 0 {
    		t.Errorf("Clamp(-5, 0, 100) = %d, want 0", got)
    	}
    	if got := Clamp(42, 0, 100); got != 42 {
    		t.Errorf("Clamp(42, 0, 100) = %d, want 42", got)
    	}
    	if got := Clamp("m", "a", "k"); got != "k" {
    		t.Errorf(`Clamp("m", "a", "k") = %q, want "k"`, got)
    	}
    	if got := Clamp(Priority(9), 1, 5); got != 5 {
    		t.Errorf("Clamp(Priority(9), 1, 5) = %d, want 5", got)
    	}
    }

    func TestMinMax(t *testing.T) {
    	lo, hi, ok := MinMax([]int{4, 9, -2, 7})
    	if !ok || lo != -2 || hi != 9 {
    		t.Errorf("MinMax([4 9 -2 7]) = (%d, %d, %v), want (-2, 9, true)", lo, hi, ok)
    	}
    	s1, s2, ok := MinMax([]string{"pear", "fig", "plum"})
    	if !ok || s1 != "fig" || s2 != "plum" {
    		t.Errorf("MinMax([pear fig plum]) = (%q, %q, %v), want (\"fig\", \"plum\", true)", s1, s2, ok)
    	}
    	one, same, ok := MinMax([]int{5})
    	if !ok || one != 5 || same != 5 {
    		t.Errorf("MinMax([5]) = (%d, %d, %v), want (5, 5, true)", one, same, ok)
    	}
    	if _, _, ok := MinMax([]int{}); ok {
    		t.Errorf("MinMax([]) returned ok=true, want false")
    	}
    }

    func TestMinMaxSkipsNaN(t *testing.T) {
    	nan := math.NaN()
    	lo, hi, ok := MinMax([]float64{nan, 3, nan, -1, 2})
    	if !ok || lo != -1 || hi != 3 {
    		t.Errorf("MinMax([NaN 3 NaN -1 2]) = (%v, %v, %v), want (-1, 3, true)", lo, hi, ok)
    	}
    	if lo, hi, ok := MinMax([]float64{nan, nan}); ok {
    		t.Errorf("MinMax([NaN NaN]) = (%v, %v, true), want ok=false", lo, hi)
    	}
    }
---

`cmp.Ordered` is the constraint you reach for whenever you need `<`. It's worth knowing exactly what's in it, what it *doesn't* promise, and what else lives in the `cmp` package.

## What cmp.Ordered is

Here's its definition from the standard library (reformatted):

```go
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}
```

No magic: it's a plain union with tildes, exactly the kind you wrote in the last lesson. Every type in it supports `<`, `<=`, `>`, `>=`, `==` and `!=`. Complex numbers are missing because they have no order. Because of the tildes, `type Priority int` and `type Key string` qualify.

## The NaN problem

Floats are ordered... mostly. `math.NaN()` isn't less than, greater than, or equal to *anything*, including itself:

```go
package main

import (
	"cmp"
	"fmt"
	"math"
)

func main() {
	nan := math.NaN()
	fmt.Println(nan < 1, nan >= 1, nan == nan)
	fmt.Println(cmp.Compare(nan, 1), cmp.Compare(nan, nan), cmp.Less(nan, math.Inf(-1)))
	fmt.Println(max(1.0, nan, 3.0))
}
```

```
false false false
-1 0 true
NaN
```

So a generic function that uses `<` on a `T cmp.Ordered` can give strange answers for float slices containing NaN. The `cmp` package fixes this with a *total* order:

- **`cmp.Compare(a, b)`** returns -1, 0 or +1, treating NaN as **less than every other value** and **equal to itself**.
- **`cmp.Less(a, b)`** is `cmp.Compare(a, b) < 0`.

`slices.Sort`, `slices.BinarySearch` and friends follow that order, which is why a sorted float slice has its NaNs at the front. The built-in `min` and `max`, on the other hand, **propagate** NaN.

There's also a neat generic NaN test: for any `T cmp.Ordered`, `x != x` is true **only** when `x` is a NaN. For ints and strings it's always false.

## cmp.Or

`cmp.Or` (Go 1.22) takes any `comparable` values and returns the first one that isn't the zero value. It's perfect for defaults:

```go
name := cmp.Or(user.Nickname, user.Name, "anonymous")
port := cmp.Or(cfg.Port, 8080)
```

It also makes multi-key sorting read nicely, since `cmp.Compare` returns 0 for "tie":

```go
slices.SortFunc(entries, func(a, b Entry) int {
	return cmp.Or(
		cmp.Compare(a.Priority, b.Priority),
		cmp.Compare(a.Key, b.Key),
	)
})
```

## Where's the constraints package?

The standard library has `cmp.Ordered` and `comparable`, and that's it. Constraints like `Integer`, `Signed` and `Float` live in `golang.org/x/exp/constraints`, outside the standard library. They're one line each, so most code just declares what it needs, as you did with `Measure`. Keep yours small and name them after what they *mean* in your domain.

## Your turn

Complete two helpers for Stash:

- `Clamp(v, lo, hi)` returns `v` limited to `[lo, hi]`. (The built-in `min` and `max` work on any `cmp.Ordered` type.)
- `MinMax(xs)` returns the smallest and largest values, **skipping NaNs**, with `ok == false` if nothing is left. Use the `x != x` trick, so the same code works for ints and strings.
