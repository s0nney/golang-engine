---
title: Constraint Type Inference
quiz:
  - question: |
      With `type Tags []string`, what does this print?

      ```go
      func UpperA[E ~string](s []E) []E { /* uppercases each element */ }
      func UpperB[S ~[]E, E ~string](s S) S { /* same body */ }

      tags := Tags{"go"}
      fmt.Printf("%T %T\n", UpperA(tags), UpperB(tags))
      ```
    options:
      - text: '`main.Tags main.Tags`'
      - text: '`[]string main.Tags`'
        correct: true
      - text: '`[]string []string`'
      - text: '`UpperB` doesn''t compile, because `E` can''t be inferred'
    explanation: |
      `UpperA` takes and returns `[]E`, so the `Tags` argument is matched through its
      underlying type and the result is a plain `[]string`. `UpperB` infers `S = Tags`
      directly, then gets `E = string` from the constraint `~[]E`, and returns an `S`.
  - question: In `func Sum[S ~[]E, E cmp.Ordered](s S) E`, how is `E` inferred from `Sum(scores)`?
    options:
      - text: From the return type
      - text: 'From `S`''s constraint: once `S` is known, `~[]E` is matched against it to find `E`'
        correct: true
      - text: It can't be; you must write `Sum[[]int, int](scores)`
      - text: '`E` defaults to `any`'
    explanation: |
      `E` doesn't appear in any parameter type directly. Constraint type inference
      unifies the inferred `S` with the type term in its constraint, `~[]E`, which reveals `E`.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // Tags is a named slice with its own methods.
    type Tags []string

    // String joins the tags with commas.
    func (t Tags) String() string { return strings.Join(t, ",") }

    // Filter returns the elements of s for which keep returns true, in order.
    // It must return the same slice TYPE it was given: Filter(tags, ...) is a Tags.
    func Filter[S ~[]E, E any](s S, keep func(E) bool) S {
    	// ?
    	return nil
    }

    // Uniq returns s with duplicates removed, keeping the first occurrence
    // of each value, in order. It must not modify s.
    func Uniq[S ~[]E, E comparable](s S) S {
    	// ?
    	return nil
    }

    func main() {
    	tags := Tags{"go", "rust", "go", "zig", "generics"}
    	short := Filter(tags, func(t string) bool { return len(t) <= 4 })
    	fmt.Println(short.String()) // a Tags, so String is available
    	fmt.Println(Uniq(tags).String())
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // Tags is a named slice with its own methods.
    type Tags []string

    // String joins the tags with commas.
    func (t Tags) String() string { return strings.Join(t, ",") }

    // Filter returns the elements of s for which keep returns true, in order.
    func Filter[S ~[]E, E any](s S, keep func(E) bool) S {
    	var out S
    	for _, v := range s {
    		if keep(v) {
    			out = append(out, v)
    		}
    	}
    	return out
    }

    // Uniq returns s with duplicates removed, keeping the first occurrence
    // of each value, in order. It must not modify s.
    func Uniq[S ~[]E, E comparable](s S) S {
    	seen := make(map[E]bool, len(s))
    	var out S
    	for _, v := range s {
    		if !seen[v] {
    			seen[v] = true
    			out = append(out, v)
    		}
    	}
    	return out
    }

    func main() {
    	tags := Tags{"go", "rust", "go", "zig", "generics"}
    	short := Filter(tags, func(t string) bool { return len(t) <= 4 })
    	fmt.Println(short.String())
    	fmt.Println(Uniq(tags).String())
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    type IDs []int

    func TestFilterKeepsNamedType(t *testing.T) {
    	tags := Tags{"go", "rust", "go", "zig", "generics"}
    	got := Filter(tags, func(s string) bool { return len(s) <= 3 })
    	if s := got.String(); s != "go,go,zig" {
    		t.Errorf("Filter(tags, len<=3).String() = %q, want %q", s, "go,go,zig")
    	}
    	ids := Filter(IDs{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 })
    	if !slices.Equal(ids, IDs{2, 4, 6}) {
    		t.Errorf("Filter(IDs{1..6}, even) = %v, want [2 4 6]", ids)
    	}
    	if none := Filter(IDs{1, 3}, func(n int) bool { return n%2 == 0 }); len(none) != 0 {
    		t.Errorf("Filter(IDs{1, 3}, even) = %v, want an empty slice", none)
    	}
    }

    func TestUniq(t *testing.T) {
    	tags := Tags{"go", "rust", "go", "zig", "rust", "go"}
    	got := Uniq(tags)
    	if s := got.String(); s != "go,rust,zig" {
    		t.Errorf("Uniq(%v).String() = %q, want %q", []string(tags), s, "go,rust,zig")
    	}
    	if !slices.Equal(tags, Tags{"go", "rust", "go", "zig", "rust", "go"}) {
    		t.Errorf("Uniq modified its input: now %v", []string(tags))
    	}
    	if ids := Uniq(IDs{3, 3, 3, 1}); !slices.Equal(ids, IDs{3, 1}) {
    		t.Errorf("Uniq(IDs{3, 3, 3, 1}) = %v, want [3 1]", ids)
    	}
    }
---

In lesson 1, `First(tags)` worked by matching `Tags` against `[]T` through its underlying type. That's fine when you return an element, but when you return a *slice*, the named type is lost:

```go
func UpperA[E ~string](s []E) []E // UpperA(tags) returns a []string, not a Tags
```

The caller wanted a `Tags` back, with its methods. The fix is a pattern you've seen all over the `slices` package, and it relies on the second kind of inference: **constraint type inference**.

## Two type parameters: the slice and its element

```go
func UpperB[S ~[]E, E ~string](s S) S
```

Read it as: "`S` is any slice type whose elements are `E`s; `E` is any string-like type". The parameter is `S`, so the argument's type is matched against `S` **exactly**: `S = Tags`. And the result is an `S`, so a `Tags` comes back.

But how does the compiler find `E`? It doesn't appear in any parameter type. That's where the constraint comes in.

## Inference from constraints

After unifying the arguments, the compiler looks at each type parameter's constraint. If the constraint has a single underlying type term, like `~[]E`, it unifies the known type argument with it:

- `S = Tags`, whose underlying type is `[]string`
- unify `[]string` with `[]E`, so `E = string`

Then it checks `E = string` against `E`'s own constraint (`~string`), and everything's settled.

```go
package main

import (
	"fmt"
	"strings"
)

type Tags []string

func (t Tags) String() string { return strings.Join(t, ",") }

func UpperA[E ~string](s []E) []E {
	out := make([]E, len(s))
	for i, v := range s {
		out[i] = E(strings.ToUpper(string(v)))
	}
	return out
}

func UpperB[S ~[]E, E ~string](s S) S {
	out := make(S, len(s))
	for i, v := range s {
		out[i] = E(strings.ToUpper(string(v)))
	}
	return out
}

func main() {
	tags := Tags{"go", "types"}
	a := UpperA(tags)
	b := UpperB(tags)
	fmt.Printf("%T %v | %T %v\n", a, a, b, b)
}
```

```
[]string [GO TYPES] | main.Tags GO,TYPES
```

`b` is a `Tags`, so `fmt` even used its `String` method.

## You've been using this all along

The standard library is full of this shape:

```go
func slices.Clone[S ~[]E, E any](s S) S
func slices.Compact[S ~[]E, E comparable](s S) S
func maps.Clone[M ~map[K]V, K comparable, V any](m M) M
```

It works for maps and functions too. Any constraint with a single type literal term (`~map[K]V`, `~func(E) bool`, `~*T`) can feed inference. In chapter 4, the `interface{ *T; ... }` pattern uses exactly this.

**Rule of thumb:** if a generic function takes a slice or map and **returns** one of the same kind, use `S ~[]E` (or `M ~map[K]V`) so your callers keep their named types.

## Your turn

Stash's tag lists are `Tags`, with a `String` method. Implement two helpers that preserve the caller's slice type:

- `Filter(s, keep)` returns the elements for which `keep` returns `true`, in order.
- `Uniq(s)` removes duplicates, keeping the first occurrence of each value, in order, without modifying `s`. (A `map[E]bool` of seen values works nicely: `E` is `comparable`.)

Declare the result as `var out S` (or `make(S, 0, len(s))`) so it's the right type.
