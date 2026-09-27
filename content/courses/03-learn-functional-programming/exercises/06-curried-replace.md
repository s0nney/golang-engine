---
title: Curried Replace
difficulty: easy
after: currying
hints:
  - 'Read the return type from the outside in: `Curry3` returns a `func(A)` that returns a `func(B)` that returns a `func(C) R`. So you need three nested `return func(...) ...` literals.'
  - 'The innermost function has all three arguments in scope (the closures captured `a` and `b`), so it can finally call `f(a, b, c)`.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // Curry3 turns a three-argument function into a chain of one-argument
    // functions: Curry3(f)(a)(b)(c) == f(a, b, c).
    func Curry3[A, B, C, R any](f func(A, B, C) R) func(A) func(B) func(C) R {
    	return func(a A) func(B) func(C) R {
    		// Return a function that takes b, which returns a function that
    		// takes c and finally calls f.
    		return nil
    	}
    }

    // replace swaps every from for to in doc. Note the argument order:
    // the settings come first and the document comes last.
    func replace(from, to, doc string) string {
    	return strings.ReplaceAll(doc, from, to)
    }

    func main() {
    	curried := Curry3(replace)
    	smartQuotes := curried(`"`)("”")
    	if smartQuotes == nil {
    		fmt.Println("Curry3 isn't finished yet")
    		return
    	}
    	fmt.Println(smartQuotes(`say "hi"`)) // want: say ”hi”
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func Curry3[A, B, C, R any](f func(A, B, C) R) func(A) func(B) func(C) R {
    	return func(a A) func(B) func(C) R {
    		return func(b B) func(C) R {
    			return func(c C) R {
    				return f(a, b, c)
    			}
    		}
    	}
    }

    func replace(from, to, doc string) string {
    	return strings.ReplaceAll(doc, from, to)
    }

    func main() {
    	curried := Curry3(replace)
    	smartQuotes := curried(`"`)("”")
    	if smartQuotes == nil {
    		fmt.Println("Curry3 isn't finished yet")
    		return
    	}
    	fmt.Println(smartQuotes(`say "hi"`))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    )

    func TestCurry3Replace(t *testing.T) {
    	step := Curry3(replace)("--")
    	if step == nil {
    		t.Fatal("Curry3(replace)(\"--\") returned nil, want a function")
    	}
    	dash := step("—")
    	if dash == nil {
    		t.Fatal("Curry3(replace)(\"--\")(\"—\") returned nil, want a function")
    	}
    	for _, tt := range []struct{ in, want string }{
    		{"wait -- what", "wait — what"},
    		{"no dashes", "no dashes"},
    		{"a--b--c", "a—b—c"},
    		{"", ""},
    	} {
    		if got := dash(tt.in); got != tt.want {
    			t.Errorf("Curry3(replace)(\"--\")(\"—\")(%q) = %q, want %q", tt.in, got, tt.want)
    		}
    	}
    }

    func TestCurry3Order(t *testing.T) {
    	format := func(a int, b string, c bool) string { return fmt.Sprintf("%d-%s-%v", a, b, c) }
    	fa := Curry3(format)(7)
    	if fa == nil {
    		t.Fatal("Curry3(format)(7) returned nil, want a function")
    	}
    	fb := fa("x")
    	if fb == nil {
    		t.Fatal("Curry3(format)(7)(\"x\") returned nil, want a function")
    	}
    	if got := fb(true); got != "7-x-true" {
    		t.Errorf("Curry3(format)(7)(\"x\")(true) = %q, want %q: pass the arguments to f in order", got, "7-x-true")
    	}
    }

    func TestCurry3PartialsAreIndependent(t *testing.T) {
    	fromCat := Curry3(replace)("cat")
    	if fromCat == nil {
    		t.Fatal("Curry3(replace)(\"cat\") returned nil, want a function")
    	}
    	toDog, toCow := fromCat("dog"), fromCat("cow")
    	if toDog == nil || toCow == nil {
    		t.Fatal("Curry3(replace)(\"cat\")(...) returned nil, want a function")
    	}
    	if got := toDog("cat cat"); got != "dog dog" {
    		t.Errorf("toDog(%q) = %q, want %q", "cat cat", got, "dog dog")
    	}
    	if got := toCow("cat"); got != "cow" {
    		t.Errorf("toCow(%q) = %q, want %q (built from the same partial as toDog)", "cat", got, "cow")
    	}
    	if got := toDog("cat"); got != "dog" {
    		t.Errorf("toDog(%q) = %q after building toCow, want %q", "cat", got, "dog")
    	}
    }

    func TestCurry3CallsFLazily(t *testing.T) {
    	calls := 0
    	f := func(a, b, c int) int { calls++; return a + b + c }
    	partial := Curry3(f)(1)
    	if partial == nil {
    		t.Fatal("Curry3(f)(1) returned nil, want a function")
    	}
    	last := partial(2)
    	if calls != 0 {
    		t.Errorf("f was called %d times before the last argument arrived, want 0", calls)
    	}
    	if last == nil {
    		t.Fatal("Curry3(f)(1)(2) returned nil, want a function")
    	}
    	if got := last(3); got != 6 || calls != 1 {
    		t.Errorf("Curry3(sum)(1)(2)(3) = %d with %d call(s) of f, want 6 with 1 call", got, calls)
    	}
    }
---

Doc2Doc's typography pass is a list of replacements: `--` becomes `—`, straight
quotes become curly ones, `(c)` becomes `©`. Each is `replace(from, to, doc)`
with the first two arguments fixed. Currying lets you fix them one at a time
and end up with a ready-to-use `func(doc string) string`.

Complete the generic `Curry3(f)`. It turns a function of three arguments into a
chain of three one-argument functions, so that

```go
Curry3(f)(a)(b)(c) == f(a, b, c)
```

## Example

```go
curried := Curry3(replace)
dash := curried("--")("—")   // a func(string) string
dash("wait -- what")         // "wait — what"

toDog := curried("cat")("dog")
toDog("cat and cat")         // "dog and dog"
```

## Constraints

- Call `f` only once all three arguments have arrived, and exactly once per
  full chain of calls.
- Every partial function can be reused: `curried("cat")` can build both
  `toDog` and `toCow` without them interfering.
