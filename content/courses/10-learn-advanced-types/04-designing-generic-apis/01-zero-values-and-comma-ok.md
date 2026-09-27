---
title: Zero Values and Comma-Ok
quiz:
  - question: |
      Why doesn't this compile?

      ```go
      func Find[T any](xs []T, pred func(T) bool) T {
      	for _, x := range xs {
      		if pred(x) {
      			return x
      		}
      	}
      	return nil
      }
      ```
    options:
      - text: '`pred` can''t take a `T`'
      - text: '`nil` isn''t a valid value for every `T` (an `int` can''t be `nil`), so it can''t be returned as a `T`'
        correct: true
      - text: Generic functions can't return early from a loop
      - text: It needs `T comparable`
    explanation: |
      `nil` only fits pointers, slices, maps, channels, functions and interfaces. For a
      general `T`, write `var zero T` and return that, usually alongside a `bool` so the
      caller can tell "found the zero value" from "found nothing".
  - question: Which constraint does `func IsZero[T ???](v T) bool { var zero T; return v == zero }` need?
    options:
      - text: '`any`'
      - text: '`comparable`'
        correct: true
      - text: '`cmp.Ordered`'
      - text: None; `==` works on every type
    explanation: |
      `==` needs `comparable`. `cmp.Ordered` would also compile but needlessly rules out
      structs, pointers and booleans. With `any`, you'd need reflection
      (`reflect.ValueOf(v).IsZero()`) to test for zero.
exercise:
  starter: |
    package main

    import "fmt"

    // FirstMatch returns the first element of s for which pred is true, and
    // true. If none matches, it returns the zero value of E and false.
    func FirstMatch[S ~[]E, E any](s S, pred func(E) bool) (E, bool) {
    	// ?
    	var zero E
    	return zero, false
    }

    // IsZero reports whether v is the zero value of its type.
    func IsZero[T comparable](v T) bool {
    	// ?
    	return false
    }

    // Deref returns *p, or the zero value of T if p is nil.
    func Deref[T any](p *T) T {
    	// ?
    	return *p
    }

    type Entry struct {
    	Key  string
    	Size int
    }

    func main() {
    	entries := []Entry{{"a", 10}, {"b", 0}, {"c", 30}}
    	e, ok := FirstMatch(entries, func(e Entry) bool { return e.Size == 0 })
    	fmt.Println(e, ok)
    	fmt.Println(IsZero(Entry{}), IsZero(""), IsZero(3))
    	var missing *Entry
    	fmt.Println(Deref(missing), Deref(&entries[2]))
    }
  solution: |
    package main

    import "fmt"

    // FirstMatch returns the first element of s for which pred is true, and
    // true. If none matches, it returns the zero value of E and false.
    func FirstMatch[S ~[]E, E any](s S, pred func(E) bool) (E, bool) {
    	for _, v := range s {
    		if pred(v) {
    			return v, true
    		}
    	}
    	var zero E
    	return zero, false
    }

    // IsZero reports whether v is the zero value of its type.
    func IsZero[T comparable](v T) bool {
    	var zero T
    	return v == zero
    }

    // Deref returns *p, or the zero value of T if p is nil.
    func Deref[T any](p *T) T {
    	if p == nil {
    		var zero T
    		return zero
    	}
    	return *p
    }

    type Entry struct {
    	Key  string
    	Size int
    }

    func main() {
    	entries := []Entry{{"a", 10}, {"b", 0}, {"c", 30}}
    	e, ok := FirstMatch(entries, func(e Entry) bool { return e.Size == 0 })
    	fmt.Println(e, ok)
    	fmt.Println(IsZero(Entry{}), IsZero(""), IsZero(3))
    	var missing *Entry
    	fmt.Println(Deref(missing), Deref(&entries[2]))
    }
  tests: |
    package main

    import "testing"

    func TestFirstMatch(t *testing.T) {
    	entries := []Entry{{"a", 10}, {"b", 0}, {"c", 0}}
    	e, ok := FirstMatch(entries, func(e Entry) bool { return e.Size == 0 })
    	if !ok || e.Key != "b" {
    		t.Errorf("FirstMatch(size==0) = (%v, %v), want ({b 0}, true)", e, ok)
    	}
    	n, ok := FirstMatch([]int{1, 3, 5}, func(n int) bool { return n%2 == 0 })
    	if ok || n != 0 {
    		t.Errorf("FirstMatch([1 3 5], even) = (%d, %v), want (0, false)", n, ok)
    	}
    	z, ok := FirstMatch([]int{7, 0, 9}, func(n int) bool { return n < 5 })
    	if !ok || z != 0 {
    		t.Errorf("FirstMatch([7 0 9], <5) = (%d, %v), want (0, true): a zero value can be a real match", z, ok)
    	}
    }

    func TestIsZero(t *testing.T) {
    	if !IsZero(Entry{}) || !IsZero("") || !IsZero(0) || !IsZero[*Entry](nil) {
    		t.Errorf("IsZero should be true for Entry{}, \"\", 0 and a nil pointer")
    	}
    	if IsZero(Entry{Key: "a"}) || IsZero("x") || IsZero(-1) || IsZero(&Entry{}) {
    		t.Errorf("IsZero should be false for Entry{Key: \"a\"}, \"x\", -1 and a non-nil pointer")
    	}
    }

    func TestDeref(t *testing.T) {
    	var missing *Entry
    	if got := Deref(missing); got != (Entry{}) {
    		t.Errorf("Deref(nil *Entry) = %v, want the zero Entry", got)
    	}
    	n := 42
    	if got := Deref(&n); got != 42 {
    		t.Errorf("Deref(&42) = %d, want 42", got)
    	}
    	var s *string
    	if got := Deref(s); got != "" {
    		t.Errorf("Deref(nil *string) = %q, want \"\"", got)
    	}
    }
---

This chapter is about writing generic code that's pleasant to *use*, which is the hard part. The first design question almost every generic function or container hits is: **what do I return when there's nothing to return?**

## Getting a zero value

For a concrete type you'd return `0`, `""` or `nil`. For a type parameter, none of those is valid for every `T`. The idiomatic answer is a zero-valued variable:

```go
var zero T
return zero, false
```

You'll also see the one-liner `*new(T)`, which allocates a zeroed `T` and dereferences it. It works, but `var zero T` is clearer (and the compiler optimises either to the same thing). Go has no built-in `zero` value, so don't go looking for one.

## Say *why* there's nothing: comma-ok

Returning a zero alone is ambiguous. Did `FirstMatch` find a `0`, or nothing at all? Follow the map-lookup convention and return a second value:

```go
func (s *Stack[T]) Pop() (T, bool) // like v, ok := m[k]
func Lookup[K comparable, V any](m map[K]V, k K) (V, bool)
```

Pick the second result by what the caller needs to know:

| Return | When |
|---|---|
| `(T, bool)` | Absence is normal and there's only one reason for it: not found, empty, closed. |
| `(T, error)` | There are several reasons, or the caller should report or wrap them: parse failures, I/O, validation. |
| `*T` | Rarely. Pointers to results force an allocation and make callers nil-check; use them only when `T` is big or shared on purpose. |
| `T` alone | When the zero value is a perfectly good answer, like `Sum` of nothing being `0`. |

## Testing for zero

`v == zero` needs `==`, so the function needs `T comparable`:

```go
func IsZero[T comparable](v T) bool {
	var zero T
	return v == zero
}
```

With just `any`, you can't compare at all. If you really need "is this the zero value?" for arbitrary types, `reflect.ValueOf(&v).Elem().IsZero()` does it (chapter 7), but first ask whether a `bool` result would remove the need.

## When T is a pointer or interface

If `T` is `*Entry`, the zero value is `nil`. If `T` is `error` or `any`, it's a nil interface. So a caller writing

```go
e, _ := FirstMatch(ptrs, pred)
e.Key // panics if nothing matched
```

hits a nil pointer. That's why the `bool` matters: it lets callers check before touching the value, whatever `T` turns out to be.

## Clearing a T

`clear` (Go 1.21) zeroes every element of a slice or deletes every key of a map, and it's generic-friendly: `clear(s.items)` inside a `Stack[T]` releases references held by old elements, so the garbage collector can reclaim them. Do the same for single slots when you pop:

```go
var zero T
s.items[len(s.items)-1] = zero // don't keep the popped value alive
s.items = s.items[:len(s.items)-1]
```

For a `Stack[int]` it makes no difference. For a `Stack[*BigThing]` it's the difference between freeing memory and leaking it until the slot is reused.

## Your turn

Write three small helpers for Stash:

- `FirstMatch(s, pred)` returns the first matching element and `true`, or the zero value and `false`. A zero value can be a real match!
- `IsZero(v)` reports whether `v` equals its type's zero value.
- `Deref(p)` returns `*p`, or the zero value if `p` is `nil`.
