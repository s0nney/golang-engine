---
title: Slices in Memory
quiz:
  - question: |
      What does this print?

      ```go
      counts := []int{10, 20, 30, 40}
      top := counts[:2]
      top[0] = 99
      fmt.Println(counts[0])
      ```
    options:
      - text: '`10`'
      - text: '`99`'
        correct: true
      - text: It panics
    explanation: |
      Slicing doesn't copy. `top` and `counts` share the same backing array, so
      writing through `top` changes what `counts` sees.
  - question: |
      What does this print?

      ```go
      a := make([]int, 3, 10)
      b := append(a, 4)
      c := append(a, 5)
      fmt.Println(b[3], c[3])
      ```
    options:
      - text: '`4 5`'
      - text: '`5 5`'
        correct: true
      - text: '`4 4`'
    explanation: |
      `a` has spare capacity, so both `append` calls write into index 3 of the
      *same* backing array without reallocating. The second append overwrites the
      first, and `b` and `c` both see `5`. This is the classic append aliasing
      gotcha.
---

A **slice** is Go's answer to the fixed-size array. It feels like a growable
list, but under the hood it's a tiny struct that *points into* an array.

## The slice header

Every slice value is a three-word **header**:

```go
// Roughly what the runtime uses:
type sliceHeader struct {
	ptr *T  // pointer to the first element in the backing array
	len int // number of elements you can see
	cap int // number of elements available from ptr to the end of the array
}
```

The elements themselves live in a separate **backing array**. The header is
just 24 bytes on a 64-bit machine, however many elements there are, which is why
passing a slice to a function is cheap: you copy the header, not the data.

```
counts := make([]int, 3, 5)

header: { ptr ─┐, len: 3, cap: 5 }
               ▼
backing array: [ 0 | 0 | 0 | _ | _ ]
                 visible     spare
```

## Slicing shares memory

Re-slicing creates a new header pointing into the **same** backing array. No
elements are copied, so it's O(1):

```go
package main

import "fmt"

func main() {
	followers := []int{120, 340, 95, 410, 77}
	week2 := followers[1:3]
	fmt.Println(week2, len(week2), cap(week2))

	week2[0] = 1_000
	fmt.Println(followers)
}
```

```
[340 95] 2 4
[120 1000 95 410 77]
```

`week2` sees elements 1 and 2, and its capacity runs to the end of the backing
array (4 elements). Writing through `week2` changed `followers`, because they're
two windows onto the same memory. That's powerful (quick sort used it to sort
sub-slices in place) and dangerous (a function can modify data you thought was
private).

When you need an independent copy, say so explicitly:

```go
safe := slices.Clone(followers[1:3]) // new backing array
```

## What append really does

`append(s, x)` has two cases:

1. **There's spare capacity** (`len < cap`): write `x` into the next slot of
   the *existing* backing array and return a header with `len + 1`. O(1).
2. **It's full** (`len == cap`): allocate a *new, bigger* backing array, copy
   all the elements over, add `x`, and return a header pointing at the new array.
   O(n) for this one call.

That's why you must always use the result: `s = append(s, x)`. The returned
header might point somewhere completely new.

Case 1 is also behind the famous aliasing bug. If two slices share spare
capacity, appending to each writes to the *same* slot:

```go
package main

import "fmt"

func main() {
	base := make([]string, 2, 4)
	base[0], base[1] = "ava", "bo"

	campaignA := append(base, "cy")
	campaignB := append(base, "dee")

	fmt.Println(campaignA)
	fmt.Println(campaignB)
}
```

```
[ava bo dee]
[ava bo dee]
```

`cy` was silently overwritten! Both appends wrote into index 2 of `base`'s
backing array. The fix is to give each result its own memory, for example with
`slices.Clone(base)` before appending, or to use a full slice expression,
`base[:2:2]`, which caps the capacity so `append` is forced to reallocate.

## Nil vs empty

A `nil` slice has a nil pointer, length 0 and capacity 0. You can `len` it,
`range` over it and `append` to it, so you rarely need to care whether a slice
is nil or merely empty. One place it shows: `encoding/json` encodes a nil slice
as `null` but an empty one as `[]`.

## Further reading

- [Go Slices: usage and internals](https://go.dev/blog/slices-intro), the Go blog's tour of slice headers, re-slicing and growth.
