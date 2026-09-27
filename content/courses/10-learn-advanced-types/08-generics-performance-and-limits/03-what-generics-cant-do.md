---
title: What Generics Can't Do
quiz:
  - question: |
      Which of these compiles?
    options:
      - text: '`var ss []fmt.Stringer = []Key{"a"}` (where `Key` has a `String` method)'
      - text: '`type Mapper interface { Map[U any](f func(int) U) []U }`'
      - text: '`func Upcast[To any, From To](xs []From) []To`'
      - text: '`func Upcast[To, From any](xs []From) []To`'
        correct: true
    explanation: |
      Slices aren't covariant, interface methods can't have type parameters, and a type
      parameter can't be used as a constraint. Declaring two independent type parameters
      is fine; the relationship between them just has to be checked some other way (at
      runtime, in the exercise).
  - question: You want a faster code path in `Sum[T Number]` when `T` is exactly `float64`. What's available?
    options:
      - text: Declare a second `func Sum(xs []float64) float64` alongside the generic one
      - text: 'Specialise with `func Sum[float64](...)`'
      - text: A runtime type switch on `any(xs)` inside the generic function, with the generic loop as the fallback
        correct: true
      - text: Nothing at all
    explanation: |
      Go has no overloading and no specialisation. A type switch such as
      `if fs, ok := any(xs).([]float64); ok { ... }` is the escape hatch, and the compiler
      can often resolve it statically for a given shape.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // Upcast converts each element of xs to the interface type To.
    // Go can't express "From implements To" as a constraint, so the check
    // happens at runtime: if any element isn't a To, Upcast returns nil and an
    // error naming the element's index and type.
    func Upcast[To, From any](xs []From) ([]To, error) {
    	// ?
    	return nil, nil
    }

    // typeName returns T's name, even when T is an interface type.
    func typeName[T any]() string {
    	return strings.TrimPrefix(fmt.Sprintf("%T", (*T)(nil)), "*")
    }

    // Key is a Stash key.
    type Key string

    func (k Key) String() string { return "key:" + string(k) }

    func main() {
    	keys := []Key{"a", "b"}
    	ss, err := Upcast[fmt.Stringer](keys)
    	fmt.Println(ss, err)
    	_, err = Upcast[fmt.Stringer]([]int{1, 2})
    	fmt.Println(err)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // Upcast converts each element of xs to the interface type To.
    // Go can't express "From implements To" as a constraint, so the check
    // happens at runtime: if any element isn't a To, Upcast returns nil and an
    // error naming the element's index and type.
    func Upcast[To, From any](xs []From) ([]To, error) {
    	out := make([]To, len(xs))
    	for i, x := range xs {
    		t, ok := any(x).(To)
    		if !ok {
    			return nil, fmt.Errorf("element %d: %T does not implement %s", i, x, typeName[To]())
    		}
    		out[i] = t
    	}
    	return out, nil
    }

    // typeName returns T's name, even when T is an interface type.
    func typeName[T any]() string {
    	return strings.TrimPrefix(fmt.Sprintf("%T", (*T)(nil)), "*")
    }

    // Key is a Stash key.
    type Key string

    func (k Key) String() string { return "key:" + string(k) }

    func main() {
    	keys := []Key{"a", "b"}
    	ss, err := Upcast[fmt.Stringer](keys)
    	fmt.Println(ss, err)
    	_, err = Upcast[fmt.Stringer]([]int{1, 2})
    	fmt.Println(err)
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strings"
    	"testing"
    )

    type code int

    func (c code) Error() string { return fmt.Sprint("code ", int(c)) }

    func TestUpcastOK(t *testing.T) {
    	ss, err := Upcast[fmt.Stringer]([]Key{"a", "b"})
    	if err != nil || len(ss) != 2 || ss[0].String() != "key:a" || ss[1].String() != "key:b" {
    		t.Fatalf("Upcast[fmt.Stringer]([]Key{a b}) = (%v, %v), want ([key:a key:b], nil)", ss, err)
    	}
    	errs, err := Upcast[error]([]code{404, 500})
    	if err != nil || len(errs) != 2 || errs[1].Error() != "code 500" {
    		t.Errorf("Upcast[error]([]code{404 500}) = (%v, %v), want two errors", errs, err)
    	}
    	if joined := errors.Join(errs...); joined == nil {
    		t.Errorf("the upcast values should be usable as []error")
    	}
    	anys, err := Upcast[any]([]int{1, 2, 3})
    	if err != nil || len(anys) != 3 || anys[2] != 3 {
    		t.Errorf("Upcast[any]([1 2 3]) = (%v, %v), want ([1 2 3], nil)", anys, err)
    	}
    	empty, err := Upcast[fmt.Stringer]([]Key{})
    	if err != nil || len(empty) != 0 {
    		t.Errorf("Upcast of an empty slice = (%v, %v), want ([], nil)", empty, err)
    	}
    }

    func TestUpcastFails(t *testing.T) {
    	mixed := []any{Key("a"), 7, Key("c")}
    	got, err := Upcast[fmt.Stringer](mixed)
    	if err == nil {
    		t.Fatalf("Upcast[fmt.Stringer]([a 7 c]) returned no error")
    	}
    	if got != nil {
    		t.Errorf("on failure Upcast should return nil, got %v", got)
    	}
    	want := "element 1: int does not implement fmt.Stringer"
    	if err.Error() != want {
    		t.Errorf("error = %q, want %q", err.Error(), want)
    	}
    	if _, err := Upcast[error]([]string{"x"}); err == nil || !strings.Contains(err.Error(), "string does not implement error") {
    		t.Errorf("Upcast[error]([]string) error = %v, want one mentioning \"string does not implement error\"", err)
    	}
    }
---

You've seen most of the corners already, scattered through the course. This lesson collects the limits of Go's generics in one place, with the exact error messages, and the usual workaround for each.

## 1. No type parameters on interface methods

```go
type Mapper interface {
	Map[U any](f func(int) U) []U // interface method must have no type parameters
}
```

And from Go 1.27's side: a generic method never satisfies an interface method. An interface is a fixed list of methods that a value must have at runtime; a generic method is a family of them. **Workaround:** make the operation a generic *function* over some shared representation (an `iter.Seq[T]`, say), or give the interface a non-generic method. See [Learn OOP](/courses/learn-oop/generics-and-oop/generic-methods-and-interfaces).

## 2. No specialisation or overloading

You can't provide a special version of `Sum[T]` for `float64`, and you can't declare two functions named `Sum`. Receiver brackets always *declare* type parameters, so `func (s Set[int]) Sum()` isn't specialisation either (chapter 4). **Workaround:** a type switch with a general fallback:

```go
func Sum[T Number](xs []T) T {
	if fs, ok := any(xs).([]float64); ok {
		return T(sumFloatsCarefully(fs)) // special path
	}
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}
```

## 3. No variance

A `[]Key` is not a `[]fmt.Stringer`, even though every `Key` is a `fmt.Stringer`. Same for `Box[Key]` and `Box[fmt.Stringer]`, and for `func(string)` and `func(any)`:

```go
var ss []fmt.Stringer = keys            // cannot use keys (variable of type []Key) as []fmt.Stringer value
var b Box[fmt.Stringer] = Box[Key]{"a"} // cannot use Box[Key]{…} ... as Box[fmt.Stringer] value
var f func(any) = func(s string) {}     // cannot use func(s string) {} ... as func(any) value
```

It's not stubbornness. A `[]Key` stores keys directly; a `[]fmt.Stringer` stores two-word interface values. They have different memory layouts, so "converting" one to the other means building a new slice. Go makes you write that loop so the cost is visible. And for mutable containers, variance would be unsound anyway: if a `[]Key` could be used as a `[]fmt.Stringer`, you could store some other `Stringer` into a slice of keys.

## 4. No "implements" constraint between type parameters

The natural signature for "convert a slice of Froms to a slice of Tos, where From implements To" would be:

```go
func Upcast[To any, From To](xs []From) []To // cannot use a type parameter as constraint
```

A constraint must be a real interface known at compile time, not another type parameter. **Workaround:** two independent type parameters, checked at runtime with an assertion to the type parameter, `any(x).(To)`. That's this lesson's exercise.

## 5. Things that aren't generic at all

- **No generic variables or constants**: only functions, methods and types take type parameters.
- **No embedding a type parameter** (`struct{ T }`), and **no type parameter as a whole type** (`type Wrapper[T any] T`), both from chapter 4.
- **No field access** through a type parameter, even if every type in the set has the field (chapter 2).
- **No operator overloading**: a constraint can't make `+` work on your struct. Use methods, like `Add(A) A` with a self-referential constraint.
- **No default type arguments, and no variadic type parameters** (`[Ts ...any]`).
- **No instantiation at runtime**: reflection can't create `Set[X]` for a type `X` it discovered, unless the program already contains that instantiation.

Most of these are deliberate: each would make the language, the compiler or the runtime considerably more complex. The Go team revisits them as real-world experience accumulates (generic methods took until 1.27), so check the release notes now and then.

## Your turn

Write `Upcast[To, From any](xs []From) ([]To, error)`, the variance workaround for Stash:

- convert each element with `any(x).(To)`, which succeeds when `x`'s type implements `To` (or is `To`);
- if an element fails, return `nil` and the error `element N: <type> does not implement <To>`, using `%T` for the element and the provided `typeName[To]()` for the target.

It makes one allocation for the new slice, exactly the cost Go wanted you to see.
