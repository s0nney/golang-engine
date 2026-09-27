---
title: 'Practice: Compose and Pipe'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // Compose returns a function that runs f, then passes f's result to g.
    func Compose[A, B, C any](f func(A) B, g func(B) C) func(A) C {
    	return func(a A) C {
    		// ?
    		var zero C
    		return zero
    	}
    }

    // Pipe returns a function that runs every step in order, left to right,
    // feeding each step's output into the next. With no steps it returns its
    // input unchanged.
    func Pipe[T any](steps ...func(T) T) func(T) T {
    	return func(v T) T {
    		// ?
    		return v
    	}
    }

    func addFooter(s string) string { return s + "\n-- Doc2Doc" }

    func main() {
    	wordCount := Compose(strings.Fields, func(ws []string) int { return len(ws) })
    	fmt.Println(wordCount("functional go is fun")) // 4

    	publish := Pipe(strings.TrimSpace, strings.ToUpper, addFooter)
    	fmt.Println(publish("  release notes  "))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // Compose returns a function that runs f, then passes f's result to g.
    func Compose[A, B, C any](f func(A) B, g func(B) C) func(A) C {
    	return func(a A) C {
    		return g(f(a))
    	}
    }

    // Pipe returns a function that runs every step in order, left to right,
    // feeding each step's output into the next. With no steps it returns its
    // input unchanged.
    func Pipe[T any](steps ...func(T) T) func(T) T {
    	return func(v T) T {
    		for _, step := range steps {
    			v = step(v)
    		}
    		return v
    	}
    }

    func addFooter(s string) string { return s + "\n-- Doc2Doc" }

    func main() {
    	wordCount := Compose(strings.Fields, func(ws []string) int { return len(ws) })
    	fmt.Println(wordCount("functional go is fun")) // 4

    	publish := Pipe(strings.TrimSpace, strings.ToUpper, addFooter)
    	fmt.Println(publish("  release notes  "))
    }
  tests: |
    package main

    import (
    	"strconv"
    	"strings"
    	"testing"
    )

    func TestCompose(t *testing.T) {
    	count := Compose(strings.Fields, func(ws []string) int { return len(ws) })
    	if got := count("a b c"); got != 3 {
    		t.Errorf("Compose(strings.Fields, len)(%q) = %d, want 3", "a b c", got)
    	}

    	label := Compose(func(n int) int { return n * 2 }, strconv.Itoa)
    	if got := label(21); got != "42" {
    		t.Errorf("Compose(double, strconv.Itoa)(21) = %q, want %q", got, "42")
    	}

    	order := Compose(func(s string) string { return s + "!" }, func(s string) string { return "<" + s + ">" })
    	if got := order("hi"); got != "<hi!>" {
    		t.Errorf("Compose(f, g)(%q) = %q, want %q (f must run first, then g)", "hi", got, "<hi!>")
    	}
    }

    func TestPipe(t *testing.T) {
    	p := Pipe(strings.TrimSpace, strings.ToUpper, func(s string) string { return s + "!" })
    	if got := p("  go  "); got != "GO!" {
    		t.Errorf("Pipe(TrimSpace, ToUpper, bang)(%q) = %q, want %q", "  go  ", got, "GO!")
    	}

    	order := Pipe(func(s string) string { return s + "a" }, func(s string) string { return s + "b" })
    	if got := order(""); got != "ab" {
    		t.Errorf("Pipe(addA, addB)(\"\") = %q, want %q (steps run left to right)", got, "ab")
    	}

    	if got := Pipe[int]()(7); got != 7 {
    		t.Errorf("Pipe()(7) = %d, want 7 (no steps means no change)", got)
    	}

    	double := func(n int) int { return n * 2 }
    	if got := Pipe(double, double, double)(1); got != 8 {
    		t.Errorf("Pipe(double, double, double)(1) = %d, want 8", got)
    	}
    }
---

Doc2Doc builds its conversion pipelines out of small functions. You need two helpers
to glue them together.

## Your task

- `Compose[A, B, C any](f func(A) B, g func(B) C) func(A) C` returns a function that
  runs `f` and then feeds its result to `g`. Types can change along the way:
  `Compose(strings.Fields, countWords)` goes from `string` to `[]string` to `int`.
- `Pipe[T any](steps ...func(T) T) func(T) T` returns a function that runs every step in
  order, **left to right**. With no steps, it returns its input unchanged.

```go
publish := Pipe(strings.TrimSpace, strings.ToUpper, addFooter)
publish("  release notes  ")
// "RELEASE NOTES\n-- Doc2Doc"
```

## Tips

- Both helpers *return* functions. The work happens inside the returned closure, which
  captures `f`, `g` or `steps`.
- Order matters. The tests check that `f` runs before `g`, and that `Pipe` runs its steps
  left to right.
- Type inference does the rest. Notice that `main` never writes out the type parameters.
