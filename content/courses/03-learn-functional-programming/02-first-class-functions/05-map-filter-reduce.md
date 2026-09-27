---
title: Map, Filter and Reduce
quiz:
  - question: |
      Using the `Map`, `Filter` and `Reduce` from this lesson, what does this print?

      ```go
      nums := []int{1, 2, 3, 4, 5}
      evens := Filter(nums, func(n int) bool { return n%2 == 0 })
      squares := Map(evens, func(n int) int { return n * n })
      total := Reduce(squares, 0, func(acc, n int) int { return acc + n })
      fmt.Println(total)
      ```
    options:
      - text: '`55`'
      - text: '`6`'
      - text: '`30`'
      - text: '`20`'
        correct: true
    explanation: |
      `Filter` keeps `[2 4]`, `Map` squares them into `[4 16]`, and `Reduce` sums
      them starting from 0, giving 20.
  - question: 'Why does `Map` have two type parameters, `Map[T, U any]`?'
    options:
      - text: So the output element type can differ from the input element type, like turning strings into ints
        correct: true
      - text: Go requires at least two type parameters on every generic function
      - text: One is for the slice and one is for its length
    explanation: |
      `T` is the input element type and `U` is the output element type. With a
      single parameter you couldn't map `[]string` to `[]int` word lengths.
---

Three higher-order functions form the backbone of functional programming:

- **Map** transforms every element: `[a b c]` becomes `[f(a) f(b) f(c)]`.
- **Filter** keeps the elements that pass a test.
- **Reduce** (also called *fold*) combines all elements into a single value.

Many languages ship these built in. Go's standard library doesn't have them for
slices. Thanks to generics, each one takes only a few lines to write yourself.

## The implementations

```go
func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, 0, len(s))
	for _, v := range s {
		out = append(out, f(v))
	}
	return out
}

func Filter[T any](s []T, keep func(T) bool) []T {
	var out []T
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

func Reduce[T, A any](s []T, init A, f func(A, T) A) A {
	acc := init
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}
```

Each one returns a **new** value and never modifies `s`. Notice that the loops are
still there. FP in Go doesn't get rid of loops, it just writes each one once.

## Doc2Doc: analysing a document

Let's use them to find the total length of the long words in a document:

```go
package main

import (
	"fmt"
	"strings"
)

func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, 0, len(s))
	for _, v := range s {
		out = append(out, f(v))
	}
	return out
}

func Filter[T any](s []T, keep func(T) bool) []T {
	var out []T
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

func Reduce[T, A any](s []T, init A, f func(A, T) A) A {
	acc := init
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}

func main() {
	doc := "Functional code is often shorter and easier to test"
	words := strings.Fields(doc)

	long := Filter(words, func(w string) bool { return len(w) > 5 })
	lengths := Map(long, func(w string) int { return len(w) })
	total := Reduce(lengths, 0, func(sum, n int) int { return sum + n })

	fmt.Println(long)
	fmt.Println(lengths)
	fmt.Println(total)
}
```

```text
[Functional shorter easier]
[10 7 6]
23
```

`Reduce` is the most general of the three. Its accumulator type `A` can be
anything, including a map. This counts how often each word appears:

```go
counts := Reduce(words, map[string]int{}, func(m map[string]int, w string) map[string]int {
	m[strings.ToLower(w)]++
	return m
})
```

(That one mutates the map as it goes, which is fine because the map was created just
for this call and nobody else can see it.)

## Should you use them?

In Go, it depends. A chain of `Filter`, `Map` and `Reduce` builds a new slice at every
step, and each anonymous function is wordy. For a one-off calculation, a single
`for` loop that filters, transforms and sums in one pass is often shorter *and*
faster. Use these helpers when they make the intent clearer, for example when the
transformation functions already have good names:

```go
names := Map(paths, filepath.Base)
upper := Map(names, strings.ToUpper)
```

In the Iterators chapter you'll build lazy versions that don't allocate a slice at
every step.

## Further reading

- [Learn Go with Tests: Revisiting arrays and slices with generics](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/revisiting-arrays-and-slices-with-generics)
