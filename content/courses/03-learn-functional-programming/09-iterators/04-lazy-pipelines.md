---
title: Lazy Pipelines
quiz:
  - question: |
      Using the lazy `Map`, `Filter` and `Take` from this lesson, how many times does `expensive` run?

      ```go
      words := slices.Values([]string{"a", "bb", "ccc", "dddd", "eeeee"})
      for w := range Take(Map(words, expensive), 2) {
          fmt.Println(w)
      }
      ```
    options:
      - text: 5
      - text: 3
      - text: 0
      - text: 2
        correct: true
    explanation: |
      Lazy pipelines pull values through one at a time. After `Take` has yielded
      two values it stops asking, so `Map` never calls `expensive` on the rest.
  - question: What's the main advantage of a lazy `Filter` over the slice-based `Filter` from chapter 2?
    options:
      - text: It's always shorter to write
      - text: It sorts the results automatically
      - text: It doesn't build an intermediate slice, and it can stop early or work on infinite sequences
        correct: true
    explanation: |
      The slice version allocates a new slice at every stage and processes every
      element. The lazy version passes values straight through, one at a time, and
      does no more work than the final consumer asks for.
---

Back in chapter 2 you wrote `Map`, `Filter` and `Reduce` for slices. Each stage built
a whole new slice before the next stage started. With iterators you can build the
same pipeline **lazily**: values flow through every stage one at a time, and nothing
happens until someone consumes the result.

## Lazy Map, Filter and Take

```go
func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}

func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if keep(v) && !yield(v) {
				return
			}
		}
	}
}

func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		if n <= 0 {
			return
		}
		count := 0
		for v := range seq {
			if !yield(v) {
				return
			}
			count++
			if count == n {
				return
			}
		}
	}
}
```

Each function takes an iterator and returns a new one. Calling them does **no work
at all**. It just wraps one function in another. The work happens when a `for`
loop, or `slices.Collect`, finally ranges over the outermost iterator.

## Watching laziness happen

Doc2Doc wants the first two long headings in a document. Let's print inside each stage
to see the order things happen in:

```go
package main

import (
	"fmt"
	"iter"
	"strings"
)

func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}

func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if keep(v) && !yield(v) {
				return
			}
		}
	}
}

func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		if n <= 0 {
			return
		}
		count := 0
		for v := range seq {
			if !yield(v) {
				return
			}
			count++
			if count == n {
				return
			}
		}
	}
}

func main() {
	doc := "# Intro\n# Installing Doc2Doc\ntext\n# Converting files\n# Formats\n"

	headings := Filter(strings.Lines(doc), func(l string) bool {
		fmt.Println("  filter:", strings.TrimSpace(l))
		return strings.HasPrefix(l, "# ")
	})
	titles := Map(headings, func(l string) string {
		return strings.TrimSpace(strings.TrimPrefix(l, "# "))
	})
	long := Filter(titles, func(t string) bool { return len(t) > 8 })

	fmt.Println("pipeline built, nothing has run yet")
	for t := range Take(long, 2) {
		fmt.Println("GOT", t)
	}
}
```

```text
pipeline built, nothing has run yet
  filter: # Intro
  filter: # Installing Doc2Doc
GOT Installing Doc2Doc
  filter: text
  filter: # Converting files
GOT Converting files
```

Look closely:

- Nothing printed until the `for` loop started. Building the pipeline is free.
- Each line went through **all** the stages before the next line was read. There's
  no intermediate slice anywhere.
- After the second result, `Take` stopped, and `# Formats` was **never even
  examined**. On a 100 MB document that saves a lot of work.

## Collecting results

When you do want a slice at the end, `slices.Collect` drains an iterator:

```go
all := slices.Collect(Map(strings.Lines(doc), strings.ToUpper))
```

## Honest trade-offs

Lazy pipelines are great for big or infinite inputs and for stopping early. But:

- Every value passes through several function calls, so for small slices a plain
  loop is faster.
- Go has no method chaining for `iter.Seq` (it's a function type from another
  package, so you can't add methods to it). You read pipelines inside-out or build
  them step by step as above. Since Go 1.27 you *could* define your own
  `type Stream[T any] iter.Seq[T]` with generic methods like `Map[U]`, but that's
  non-standard, so check that your team wants it.
- Debugging is harder, because the stages interleave as the output above shows.

Use lazy pipelines where laziness buys you something, and plain loops elsewhere.
