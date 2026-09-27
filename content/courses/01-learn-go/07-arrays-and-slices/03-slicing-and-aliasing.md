---
title: Slicing and Aliasing
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	ids := []int{10, 20, 30, 40, 50}
      	part := ids[1:3]
      	part[0] = 99
      	fmt.Println(part, ids)
      }
      ```
    options:
      - text: '`[99 30] [10 20 30 40 50]`'
      - text: '`[99 30] [10 99 30 40 50]`'
        correct: true
      - text: '`[99 30 40] [10 99 30 40 50]`'
      - text: '`[20 30] [10 99 30 40 50]`'
    explanation: |
      `ids[1:3]` holds the elements at indexes 1 and 2 (the end is
      exclusive). Slicing doesn't copy: `part` shares the same underlying
      array as `ids`, so setting `part[0]` also changes `ids[1]`.
  - question: How do you make a fully independent copy of a slice `s`?
    options:
      - text: '`c := s`'
      - text: '`c := s[:]`'
      - text: '`c := slices.Clone(s)`'
        correct: true
    explanation: |
      Both `c := s` and `c := s[:]` create a new slice header that still
      points at the same array, so changes show up in both. `slices.Clone`
      allocates a new array and copies the elements.
---

You can take a smaller piece of a slice (or array) with the **slice expression** `s[low:high]`. It's handy, but it hides one of Go's most important gotchas.

## Slicing

`s[low:high]` gives you the elements from index `low` up to, **but not including**, `high`:

```go
package main

import "fmt"

func main() {
	queue := []string{"a", "b", "c", "d", "e"}

	fmt.Println(queue[1:3])
	fmt.Println(queue[:2])
	fmt.Println(queue[3:])
	fmt.Println(queue[:])
}
```

```text
[b c]
[a b]
[d e]
[a b c d e]
```

- Leave out `low` and it defaults to `0`.
- Leave out `high` and it defaults to `len(s)`.
- The result has `high - low` elements.

Textio uses this to send messages in batches: `queue[:100]` is the first batch, `queue[100:]` is everything else.

## Slices share memory

Here's the key idea. A slice doesn't hold its elements itself. It's a small descriptor with three fields: a **pointer** to an underlying array, a **length** and a **capacity**. Slicing creates a new descriptor pointing into the **same array**. Nothing is copied.

So changes through one slice are visible through the other. This is called **aliasing**:

```go
package main

import "fmt"

func main() {
	messages := []string{"hi", "hello", "hey"}
	firstTwo := messages[:2]

	firstTwo[0] = "HI!"
	fmt.Println(messages)
}
```

```text
[HI! hello hey]
```

The same happens when you pass a slice to a function. The function gets a copy of the descriptor, but it points at the same array, so the function can change your elements:

```go
package main

import "fmt"

func redact(msgs []string) {
	for i := range msgs {
		msgs[i] = "[redacted]"
	}
}

func main() {
	log := []string{"code 4821", "code 1234"}
	redact(log)
	fmt.Println(log)
}
```

```text
[[redacted] [redacted]]
```

## The `append` aliasing trap

Aliasing gets truly sneaky with `append`. If a slice has spare capacity, `append` writes into the **shared** array instead of allocating a new one:

```go
package main

import "fmt"

func main() {
	all := []int{1, 2, 3, 4}
	firstTwo := all[:2] // len 2, but cap 4!

	firstTwo = append(firstTwo, 99)
	fmt.Println(all)
}
```

```text
[1 2 99 4]
```

`firstTwo` had room to grow (its capacity reaches to the end of `all`), so `append` overwrote `all[2]`. If `firstTwo` had been full, `append` would have copied to a new array and `all` would be untouched. Code whose behaviour depends on capacity like this is a bug waiting to happen.

It's also why you must always write `s = append(s, x)`: when `append` allocates a new array, the old slice variable still points at the old one.

## Making a real copy

When you need an independent slice, copy it. The simplest way is `slices.Clone` from the `slices` package:

```go
package main

import (
	"fmt"
	"slices"
)

func main() {
	original := []string{"hi", "hello"}
	backup := slices.Clone(original)
	backup[0] = "changed"
	fmt.Println(original, backup)
}
```

```text
[hi hello] [changed hello]
```

There's also the built-in `copy(dst, src)`, which copies as many elements as fit into an existing slice:

```go
dst := make([]string, len(src))
copy(dst, src)
```

## Rules of thumb

- Slicing and passing slices to functions **share** data.
- Assume a sub-slice can be modified through the original, and vice versa.
- If you need to keep or modify a piece independently, `slices.Clone` it.

## Further reading

- [A Tour of Go: Slices are like references to arrays](https://go.dev/tour/moretypes/8)
- [The Go Blog: Go Slices, usage and internals](https://go.dev/blog/slices-intro)
