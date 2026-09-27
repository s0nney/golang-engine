---
title: Conditional Steps
difficulty: easy
after: function-transformations
hints:
  - 'Both functions **return a function**. Start each body with `return func(v T) T { ... }` and fill in the inside.'
  - '`When`: inside the returned function, call `pred(v)` and return either `f(v)` or `v` unchanged.'
  - '`Repeat`: inside the returned function, loop `for range n { v = f(v) }` and return `v`. A loop over a negative `n` runs zero times, which is exactly what you want.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // When returns a step that applies f to its input only if pred(input)
    // is true, and otherwise returns the input unchanged.
    func When[T any](pred func(T) bool, f func(T) T) func(T) T {
    	// Return a new function that checks pred before calling f.
    	return f
    }

    // Repeat returns a step that applies f to its input n times in a row:
    // Repeat(f, 3)(x) == f(f(f(x))). For n <= 0 it returns the input unchanged.
    func Repeat[T any](f func(T) T, n int) func(T) T {
    	// Return a new function that calls f n times.
    	return f
    }

    func indent(line string) string { return "  " + line }

    func isBullet(line string) bool { return strings.HasPrefix(line, "- ") }

    func main() {
    	nest := When(isBullet, Repeat(indent, 2))
    	fmt.Printf("%q\n", nest("- item")) // want: "    - item"
    	fmt.Printf("%q\n", nest("text"))   // want: "text"
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func When[T any](pred func(T) bool, f func(T) T) func(T) T {
    	return func(v T) T {
    		if pred(v) {
    			return f(v)
    		}
    		return v
    	}
    }

    func Repeat[T any](f func(T) T, n int) func(T) T {
    	return func(v T) T {
    		for range n {
    			v = f(v)
    		}
    		return v
    	}
    }

    func indent(line string) string { return "  " + line }

    func isBullet(line string) bool { return strings.HasPrefix(line, "- ") }

    func main() {
    	nest := When(isBullet, Repeat(indent, 2))
    	fmt.Printf("%q\n", nest("- item"))
    	fmt.Printf("%q\n", nest("text"))
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    )

    func TestWhen(t *testing.T) {
    	upperIfShort := When(func(s string) bool { return len(s) < 5 }, strings.ToUpper)
    	tests := []struct{ in, want string }{
    		{"abc", "ABC"},
    		{"", ""},
    		{"hello", "hello"},
    		{"longer text", "longer text"},
    	}
    	for _, tt := range tests {
    		if got := upperIfShort(tt.in); got != tt.want {
    			t.Errorf("When(len < 5, ToUpper)(%q) = %q, want %q", tt.in, got, tt.want)
    		}
    	}
    }

    func TestWhenSkipsF(t *testing.T) {
    	calls := 0
    	f := func(n int) int { calls++; return n * 10 }
    	evenTimes10 := When(func(n int) bool { return n%2 == 0 }, f)
    	if got := evenTimes10(4); got != 40 {
    		t.Errorf("When(even, *10)(4) = %d, want 40", got)
    	}
    	if got := evenTimes10(3); got != 3 {
    		t.Errorf("When(even, *10)(3) = %d, want 3", got)
    	}
    	if calls != 1 {
    		t.Errorf("f was called %d times for inputs 4 and 3, want 1: don't call f when pred is false", calls)
    	}
    }

    func TestRepeat(t *testing.T) {
    	double := func(n int) int { return n * 2 }
    	tests := []struct{ n, in, want int }{
    		{0, 5, 5},
    		{1, 5, 10},
    		{3, 1, 8},
    		{10, 1, 1024},
    		{-2, 7, 7},
    	}
    	for _, tt := range tests {
    		if got := Repeat(double, tt.n)(tt.in); got != tt.want {
    			t.Errorf("Repeat(double, %d)(%d) = %d, want %d", tt.n, tt.in, got, tt.want)
    		}
    	}
    }

    func TestRepeatIsReusable(t *testing.T) {
    	quote := Repeat(func(s string) string { return "> " + s }, 2)
    	a, b := quote("hi"), quote("yo")
    	if a != "> > hi" || b != "> > yo" {
    		t.Errorf("calling the same Repeat(quote, 2) step twice gave %q and %q, want %q and %q", a, b, "> > hi", "> > yo")
    	}
    }

    func TestWhenRepeatTogether(t *testing.T) {
    	nest := When(isBullet, Repeat(indent, 2))
    	for _, tt := range []struct{ in, want string }{
    		{"- item", "    - item"},
    		{"text", "text"},
    		{"-nospace", "-nospace"},
    	} {
    		if got := nest(tt.in); got != tt.want {
    			t.Errorf("When(isBullet, Repeat(indent, 2))(%q) = %q, want %q", tt.in, got, tt.want)
    		}
    	}
    }
---

Doc2Doc builds its formatting pipelines out of small steps of type
`func(T) T`. Two steps come up all the time: "only do this to *some* lines" and
"do this *n* times". Instead of writing a new function for every combination,
write two **function transformers** that build steps from other steps.

Complete both generic functions:

- `When(pred, f)` returns a step that calls `f` on its input when `pred(input)` is
  true, and otherwise returns the input unchanged (without calling `f`).
- `Repeat(f, n)` returns a step that applies `f` to its input `n` times in a row.
  For `n <= 0`, the step returns its input unchanged.

## Example

```go
indent := func(line string) string { return "  " + line }
isBullet := func(line string) bool { return strings.HasPrefix(line, "- ") }

Repeat(indent, 3)("x")                   // "      x"
nest := When(isBullet, Repeat(indent, 2))
nest("- item")                           // "    - item"
nest("text")                             // "text"
```

## Constraints

- Neither function calls `f` or `pred` itself: they only build and return a
  new function. The work happens when that function is called.
- The returned steps must be reusable: calling one twice gives independent results.
