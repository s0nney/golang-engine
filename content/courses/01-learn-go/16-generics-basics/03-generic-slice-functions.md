---
title: Generic Slice Functions
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func filter[T any](items []T, keep func(T) bool) []T {
      	var result []T
      	for _, item := range items {
      		if keep(item) {
      			result = append(result, item)
      		}
      	}
      	return result
      }

      func main() {
      	lengths := []int{12, 200, 160, 161}
      	long := filter(lengths, func(n int) bool { return n > 160 })
      	fmt.Println(long, len(lengths))
      }
      ```
    options:
      - text: '`[200 160 161] 4`'
      - text: '`[200 161] 2`'
      - text: '`[200 161] 4`'
        correct: true
      - text: '`[12 160] 4`'
    explanation: |
      `filter` keeps the elements for which `keep` returns `true`: only 200
      and 161 are more than 160. It builds a new slice, so the original
      `lengths` still has 4 elements.
  - question: 'Before writing your own generic helper, what should you check first?'
    options:
      - text: Whether the `slices`, `maps` or `cmp` packages already provide it
        correct: true
      - text: Whether it can be written without type parameters by using `any` everywhere
      - text: Whether Go 1.27 still supports generics
    explanation: |
      The standard library already has well-tested generic helpers such as
      `slices.Index`, `slices.ContainsFunc`, `slices.SortFunc` and
      `maps.Keys`. Reuse them rather than rewriting them.
---

Generics are at their most useful for working with **collections**. The `slices` and `maps` packages you've already used are written with generics. In this lesson you'll write a few helpers of your own, and see how they fit with the standard library.

## Filter

Textio often needs "the messages that match some rule". Write it once for any type:

```go
package main

import "fmt"

func filter[T any](items []T, keep func(T) bool) []T {
	var result []T
	for _, item := range items {
		if keep(item) {
			result = append(result, item)
		}
	}
	return result
}

type message struct {
	to        string
	delivered bool
}

func main() {
	msgs := []message{
		{"alice", true},
		{"bob", false},
		{"carol", false},
	}
	failed := filter(msgs, func(m message) bool { return !m.delivered })
	fmt.Println(failed)
}
```

```text
[{bob false} {carol false}]
```

`filter` builds a **new** slice, so the original is untouched. (The standard library's `slices.DeleteFunc` does the opposite job *in place*, reusing the original array.)

## Reduce

"Combine every element into one value" is another classic:

```go
package main

import "fmt"

func reduce[T, A any](items []T, start A, f func(A, T) A) A {
	acc := start
	for _, item := range items {
		acc = f(acc, item)
	}
	return acc
}

func main() {
	bodies := []string{"hi", "Your code is 4821", "ok"}
	totalChars := reduce(bodies, 0, func(total int, b string) int {
		return total + len(b)
	})
	fmt.Println(totalChars)
}
```

```text
21
```

`A` is the type of the accumulated result (here `int`), which can differ from the element type `T` (here `string`).

## Grouping into a map

Here's the "map of slices" pattern from the maps chapter, now generic. `K` must be `comparable`, because it's used as a map key:

```go
package main

import (
	"fmt"
	"maps"
	"slices"
)

func groupBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
	groups := make(map[K][]T)
	for _, item := range items {
		k := key(item)
		groups[k] = append(groups[k], item)
	}
	return groups
}

func main() {
	numbers := []string{"+1-555-0100", "+44-20-7946-0000", "+1-555-0199"}
	byCountry := groupBy(numbers, func(n string) string {
		if n[:3] == "+44" {
			return "GB"
		}
		return "US"
	})
	for _, c := range slices.Sorted(maps.Keys(byCountry)) {
		fmt.Println(c, byCountry[c])
	}
}
```

```text
GB [+44-20-7946-0000]
US [+1-555-0100 +1-555-0199]
```

## Use the standard library first

Before writing a helper, check `slices` and `maps`. Many are already there, generic and well tested:

| You want to... | Use |
|----------------|-----|
| Find an element | `slices.Index`, `slices.IndexFunc` |
| Check for a match | `slices.Contains`, `slices.ContainsFunc` |
| Sort by a field | `slices.SortFunc` with `cmp.Compare` |
| Remove matches in place | `slices.DeleteFunc` |
| Smallest or largest | `slices.Min`, `slices.Max`, `slices.MaxFunc` |
| Get map keys, sorted | `slices.Sorted(maps.Keys(m))` |

## Don't overdo it

Generics are a tool, not a goal. If a function only ever handles `[]message`, write it for `[]message`: it's simpler to read. Reach for type parameters when you actually have the same logic for several types. A good rule: write the concrete version first, and make it generic when you need the second copy.

## Further reading

- [Learn Go with Tests: Revisiting arrays and slices with generics](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/revisiting-arrays-and-slices-with-generics)
- [Package slices](https://pkg.go.dev/slices)
