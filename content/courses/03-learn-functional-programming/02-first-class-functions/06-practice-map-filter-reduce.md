---
title: 'Practice: Map, Filter and Reduce'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // Map returns a new slice holding f applied to every element of s.
    func Map[T, U any](s []T, f func(T) U) []U {
    	// ?
    	return nil
    }

    // Filter returns a new slice holding the elements of s for which keep
    // returns true, in their original order.
    func Filter[T any](s []T, keep func(T) bool) []T {
    	// ?
    	return nil
    }

    // Reduce folds s into a single value, starting from init and combining
    // each element into the accumulator with f, left to right.
    func Reduce[T, A any](s []T, init A, f func(A, T) A) A {
    	// ?
    	return init
    }

    func main() {
    	words := strings.Fields("Doc2Doc turns plain text into lovely documents")

    	long := Filter(words, func(w string) bool { return len(w) > 5 })
    	lengths := Map(long, func(w string) int { return len(w) })
    	total := Reduce(lengths, 0, func(sum, n int) int { return sum + n })

    	fmt.Println(long)    // [Doc2Doc lovely documents]
    	fmt.Println(lengths) // [7 6 9]
    	fmt.Println(total)   // 22
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // Map returns a new slice holding f applied to every element of s.
    func Map[T, U any](s []T, f func(T) U) []U {
    	out := make([]U, 0, len(s))
    	for _, v := range s {
    		out = append(out, f(v))
    	}
    	return out
    }

    // Filter returns a new slice holding the elements of s for which keep
    // returns true, in their original order.
    func Filter[T any](s []T, keep func(T) bool) []T {
    	var out []T
    	for _, v := range s {
    		if keep(v) {
    			out = append(out, v)
    		}
    	}
    	return out
    }

    // Reduce folds s into a single value, starting from init and combining
    // each element into the accumulator with f, left to right.
    func Reduce[T, A any](s []T, init A, f func(A, T) A) A {
    	acc := init
    	for _, v := range s {
    		acc = f(acc, v)
    	}
    	return acc
    }

    func main() {
    	words := strings.Fields("Doc2Doc turns plain text into lovely documents")

    	long := Filter(words, func(w string) bool { return len(w) > 5 })
    	lengths := Map(long, func(w string) int { return len(w) })
    	total := Reduce(lengths, 0, func(sum, n int) int { return sum + n })

    	fmt.Println(long)    // [Doc2Doc lovely documents]
    	fmt.Println(lengths) // [7 6 9]
    	fmt.Println(total)   // 22
    }
  tests: |
    package main

    import (
    	"slices"
    	"strings"
    	"testing"
    )

    func TestMap(t *testing.T) {
    	words := []string{"go", "doc", "html"}
    	if got, want := Map(words, strings.ToUpper), []string{"GO", "DOC", "HTML"}; !slices.Equal(got, want) {
    		t.Errorf("Map(%q, strings.ToUpper) = %q, want %q", words, got, want)
    	}
    	lens := Map(words, func(s string) int { return len(s) })
    	if want := []int{2, 3, 4}; !slices.Equal(lens, want) {
    		t.Errorf("Map(%q, len) = %v, want %v", words, lens, want)
    	}
    	if got := Map([]int{}, func(n int) int { return n }); len(got) != 0 {
    		t.Errorf("Map on an empty slice = %v, want an empty slice", got)
    	}
    }

    func TestFilter(t *testing.T) {
    	files := []string{"a.md", "b.txt", "c.md", "d.html"}
    	isMd := func(s string) bool { return strings.HasSuffix(s, ".md") }
    	if got, want := Filter(files, isMd), []string{"a.md", "c.md"}; !slices.Equal(got, want) {
    		t.Errorf("Filter(%q, isMd) = %q, want %q", files, got, want)
    	}
    	if got := Filter(files, func(string) bool { return false }); len(got) != 0 {
    		t.Errorf("Filter keeping nothing = %q, want an empty slice", got)
    	}
    	if want := []string{"a.md", "b.txt", "c.md", "d.html"}; !slices.Equal(files, want) {
    		t.Errorf("Filter modified its input: now %q, want %q", files, want)
    	}
    }

    func TestReduce(t *testing.T) {
    	if got := Reduce([]int{1, 2, 3, 4}, 0, func(a, n int) int { return a + n }); got != 10 {
    		t.Errorf("Reduce sum of [1 2 3 4] = %d, want 10", got)
    	}
    	joined := Reduce([]string{"a", "b", "c"}, ">", func(acc, s string) string { return acc + s })
    	if joined != ">abc" {
    		t.Errorf("Reduce concatenation = %q, want %q (did you go left to right from init?)", joined, ">abc")
    	}
    	counts := Reduce(strings.Fields("go is go"), map[string]int{}, func(m map[string]int, w string) map[string]int {
    		m[w]++
    		return m
    	})
    	if counts["go"] != 2 || counts["is"] != 1 {
    		t.Errorf("Reduce word counts = %v, want map[go:2 is:1]", counts)
    	}
    	if got := Reduce([]int(nil), 42, func(a, n int) int { return a + n }); got != 42 {
    		t.Errorf("Reduce on an empty slice = %d, want the initial value 42", got)
    	}
    }
---

Doc2Doc's analysis features all boil down to three operations: transform every item,
keep some items, and combine items into one answer. Time to write the generic helpers
yourself.

## Your task

Implement the three generic functions in the editor:

- `Map[T, U any](s []T, f func(T) U) []U` returns a **new** slice with `f` applied to
  every element.
- `Filter[T any](s []T, keep func(T) bool) []T` returns a **new** slice holding only
  the elements where `keep` returns true, in their original order.
- `Reduce[T, A any](s []T, init A, f func(A, T) A) A` starts from `init` and folds each
  element in, left to right.

```go
Reduce([]string{"a", "b"}, ">", func(acc, s string) string { return acc + s })
// ">ab"
```

## Rules

- Don't modify the input slice. The tests check it afterwards.
- `Reduce` on an empty slice must return `init` unchanged.
- The accumulator type `A` can be anything, even a map. The tests use `Reduce` to count
  words into a `map[string]int`.

Once it passes, look at `main`: three tiny anonymous functions plus your three helpers
answer "how many letters are in the long words?" without a single visible loop.
