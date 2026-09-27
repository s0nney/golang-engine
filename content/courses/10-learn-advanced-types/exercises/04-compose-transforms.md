---
title: Compose Transforms
difficulty: easy
after: type-inference
hints:
  - '`Compose` returns a **function**. Build it with a function literal that captures `f` and `g`: `return func(a A) C { ... }`.'
  - 'Inside that literal, run `f` first and feed its result to `g`: `g(f(a))`. The types line up because `f` returns a `B` and `g` takes a `B`.'
  - '`Chain` is a loop over `fs` inside the returned function, feeding each result into the next. With no functions, the loop simply doesn''t run and you return `v` unchanged.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strconv"
    	"strings"
    )

    // Compose returns a function that applies f and then g: Compose(f, g)(a) == g(f(a)).
    func Compose[A, B, C any](f func(A) B, g func(B) C) func(A) C {
    	// Return a function literal that calls f, then g.
    	return func(a A) C {
    		var zero C
    		return zero
    	}
    }

    // Chain returns a function that applies every function in fs, in order,
    // each to the previous result. With no functions it returns its input.
    func Chain[T any](fs ...func(T) T) func(T) T {
    	// Loop over fs inside the returned function.
    	return func(v T) T {
    		var zero T
    		return zero
    	}
    }

    func main() {
    	clean := Compose(strings.TrimSpace, strings.ToLower)
    	fmt.Printf("%q\n", clean("  User:42 ")) // want "user:42"

    	size := Compose(func(s string) int { return len(s) }, strconv.Itoa)
    	fmt.Printf("%q\n", size("stash")) // want "5"

    	key := Chain(strings.TrimSpace, strings.ToLower, func(s string) string { return "key:" + s })
    	fmt.Printf("%q\n", key(" ABC ")) // want "key:abc"
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strconv"
    	"strings"
    )

    // Compose returns a function that applies f and then g: Compose(f, g)(a) == g(f(a)).
    func Compose[A, B, C any](f func(A) B, g func(B) C) func(A) C {
    	return func(a A) C {
    		return g(f(a))
    	}
    }

    // Chain returns a function that applies every function in fs, in order,
    // each to the previous result. With no functions it returns its input.
    func Chain[T any](fs ...func(T) T) func(T) T {
    	return func(v T) T {
    		for _, f := range fs {
    			v = f(v)
    		}
    		return v
    	}
    }

    func main() {
    	clean := Compose(strings.TrimSpace, strings.ToLower)
    	fmt.Printf("%q\n", clean("  User:42 "))

    	size := Compose(func(s string) int { return len(s) }, strconv.Itoa)
    	fmt.Printf("%q\n", size("stash"))

    	key := Chain(strings.TrimSpace, strings.ToLower, func(s string) string { return "key:" + s })
    	fmt.Printf("%q\n", key(" ABC "))
    }
  tests: |
    package main

    import (
    	"strconv"
    	"strings"
    	"testing"
    )

    type Key string

    type Rename func(string) string

    type Entry struct {
    	Key  Key
    	Hits int
    }

    func Double[T ~int | ~float64](v T) T { return v * 2 }

    func TestComposeSameTypes(t *testing.T) {
    	clean := Compose(strings.TrimSpace, strings.ToLower)
    	if got := clean("  User:42 "); got != "user:42" {
    		t.Errorf("Compose(TrimSpace, ToLower)(%q) = %q, want %q", "  User:42 ", got, "user:42")
    	}
    }

    func TestComposeOrder(t *testing.T) {
    	addOne := func(n int) int { return n + 1 }
    	times10 := func(n int) int { return n * 10 }
    	if got := Compose(addOne, times10)(2); got != 30 {
    		t.Errorf("Compose(addOne, times10)(2) = %d, want 30 (f runs first, then g)", got)
    	}
    	if got := Compose(times10, addOne)(2); got != 21 {
    		t.Errorf("Compose(times10, addOne)(2) = %d, want 21 (f runs first, then g)", got)
    	}
    }

    func TestComposeThreeTypes(t *testing.T) {
    	size := Compose(func(s string) int { return len(s) }, strconv.Itoa)
    	if got := size("stash"); got != "5" {
    		t.Errorf("Compose(len, Itoa)(%q) = %q, want %q", "stash", got, "5")
    	}
    	hits := Compose(func(e Entry) int { return e.Hits }, func(n int) bool { return n > 10 })
    	if !hits(Entry{"a", 11}) || hits(Entry{"b", 10}) {
    		t.Errorf("Compose(Entry.Hits, >10) gave %v for 11 hits and %v for 10, want true and false", hits(Entry{"a", 11}), hits(Entry{"b", 10}))
    	}
    	toKey := Compose(strings.ToUpper, func(s string) Key { return Key("k:" + s) })
    	if got := toKey("id"); got != Key("k:ID") {
    		t.Errorf("Compose(ToUpper, toKey)(%q) = %q, want %q", "id", got, "k:ID")
    	}
    }

    func TestComposeGenericFunctionValues(t *testing.T) {
    	quad := Compose(Double[int], Double[int])
    	if got := quad(3); got != 12 {
    		t.Errorf("Compose(Double[int], Double[int])(3) = %d, want 12", got)
    	}
    	half := Compose(Double[float64], func(f float64) string { return strconv.FormatFloat(f/4, 'f', 2, 64) })
    	if got := half(1); got != "0.50" {
    		t.Errorf("half(1) = %q, want %q", got, "0.50")
    	}
    }

    func TestComposeNamedFuncType(t *testing.T) {
    	var prefix Rename = func(s string) string { return "stash/" + s }
    	f := Compose(prefix, strings.ToUpper)
    	if got := f("a"); got != "STASH/A" {
    		t.Errorf("Compose(prefix, ToUpper)(%q) = %q, want %q", "a", got, "STASH/A")
    	}
    }

    func TestChain(t *testing.T) {
    	addOne := func(n int) int { return n + 1 }
    	times10 := func(n int) int { return n * 10 }
    	tests := []struct {
    		name string
    		fs   []func(int) int
    		in   int
    		want int
    	}{
    		{"no functions", nil, 7, 7},
    		{"one", []func(int) int{addOne}, 7, 8},
    		{"in order", []func(int) int{addOne, times10}, 2, 30},
    		{"reverse order", []func(int) int{times10, addOne}, 2, 21},
    		{"repeated", []func(int) int{addOne, addOne, addOne, times10}, 0, 30},
    	}
    	for _, tt := range tests {
    		if got := Chain(tt.fs...)(tt.in); got != tt.want {
    			t.Errorf("%s: Chain(...)(%d) = %d, want %d", tt.name, tt.in, got, tt.want)
    		}
    	}
    }

    func TestChainStringsAndNamedTypes(t *testing.T) {
    	key := Chain(strings.TrimSpace, strings.ToLower, func(s string) string { return "key:" + s })
    	if got := key(" ABC "); got != "key:abc" {
    		t.Errorf("Chain(TrimSpace, ToLower, prefix)(%q) = %q, want %q", " ABC ", got, "key:abc")
    	}
    	up := func(k Key) Key { return Key(strings.ToUpper(string(k))) }
    	bang := func(k Key) Key { return k + "!" }
    	if got := Chain(up, bang, bang)("go"); got != "GO!!" {
    		t.Errorf("Chain(up, bang, bang)(%q) = %q, want %q", "go", got, "GO!!")
    	}
    	if got := Chain[Key]()("same"); got != "same" {
    		t.Errorf("Chain[Key]()(%q) = %q, want it unchanged", "same", got)
    	}
    }

    func TestChainReusable(t *testing.T) {
    	calls := 0
    	count := func(n int) int { calls++; return n }
    	f := Chain(count, count)
    	f(1)
    	f(2)
    	if calls != 4 {
    		t.Errorf("calling a two-function Chain twice ran the functions %d times, want 4", calls)
    	}
    }
---

Stash cleans every key before storing it: trim spaces, lower-case, add a
prefix. Instead of hard-coding each pipeline, Stash builds them from small
functions.

Write two generic helpers:

- `Compose(f, g)` returns a function that runs `f` and then passes its result to
  `g`. The three types involved (`f`'s input, the value in the middle, `g`'s
  output) can all differ.
- `Chain(fs...)` takes any number of functions of the same type `func(T) T` and
  returns one function that applies them **in order**. With no functions it
  returns its input unchanged.

Neither helper runs anything when called: the work happens each time the
returned function is called.

## Examples

```go
clean := Compose(strings.TrimSpace, strings.ToLower)
clean("  User:42 ") // "user:42"

size := Compose(func(s string) int { return len(s) }, strconv.Itoa)
size("stash") // "5"   (string -> int -> string)

key := Chain(strings.TrimSpace, strings.ToLower, addPrefix)
key(" ABC ") // "key:abc"
```

Notice that callers never write type arguments: Go infers `A`, `B` and `C` from
the function arguments. The hidden tests also pass generic functions as values
(like `Double[int]`) and values of named function types.

## Constraints

- `Compose(f, g)(a)` must equal `g(f(a))`, never `f(g(a))`.
- The returned functions may be called many times.
