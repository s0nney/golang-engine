---
title: Type Switches in Generic Code
quiz:
  - question: |
      What does `KindBad[error]()` return?

      ```go
      func KindBad[T any]() string {
      	var zero T
      	switch any(zero).(type) {
      	case int:
      		return "int"
      	case error:
      		return "error"
      	case nil:
      		return "nil!"
      	}
      	return "other"
      }
      ```
    options:
      - text: '`"error"`'
      - text: '`"nil!"`'
        correct: true
      - text: '`"other"`'
      - text: It doesn't compile
    explanation: |
      The zero value of the interface type `error` is a nil interface, and converting it
      to `any` gives a nil `any`, with no dynamic type at all. So it matches `case nil`.
      Switching on `any((*T)(nil))` instead always carries the type: a `*error`.
  - question: |
      In a type switch, what's the type of `x` in the `case int, int64:` branch of `switch x := v.(type)`?
    options:
      - text: '`int`'
      - text: '`int64`'
      - text: The type of `v` (for example `any`), because a multi-type case can't pick one
        correct: true
      - text: '`interface{ int | int64 }`'
    explanation: |
      Only single-type cases give `x` that type. With several types in one case (and in
      `default`), `x` keeps the type of the switched expression, so you'd still need an
      assertion or a conversion to do arithmetic on it.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    // OfType returns the items whose dynamic type is T (or, if T is an
    // interface type, implements T), in their original order.
    func OfType[T any](items []any) []T {
    	// ?
    	return nil
    }

    // Classify describes v as one of:
    //
    //	"nil"      a nil interface
    //	"number"   an int, int64 or float64
    //	"text"     a string or []byte
    //	"error"    anything implementing error (checked before fmt.Stringer)
    //	"stringer" anything else implementing fmt.Stringer
    //	"other"    everything else
    func Classify(v any) string {
    	// ?
    	return "other"
    }

    type Key string

    func (k Key) String() string { return "key:" + string(k) }

    func main() {
    	items := []any{1, "go", Key("a"), errors.New("boom"), 2.5, nil, []byte("hi"), Key("b")}
    	fmt.Println(OfType[Key](items))
    	fmt.Println(len(OfType[fmt.Stringer](items)))
    	for _, it := range items {
    		fmt.Print(Classify(it), " ")
    	}
    	fmt.Println()
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    // OfType returns the items whose dynamic type is T (or, if T is an
    // interface type, implements T), in their original order.
    func OfType[T any](items []any) []T {
    	var out []T
    	for _, it := range items {
    		if t, ok := it.(T); ok {
    			out = append(out, t)
    		}
    	}
    	return out
    }

    // Classify describes v as one of:
    //
    //	"nil"      a nil interface
    //	"number"   an int, int64 or float64
    //	"text"     a string or []byte
    //	"error"    anything implementing error (checked before fmt.Stringer)
    //	"stringer" anything else implementing fmt.Stringer
    //	"other"    everything else
    func Classify(v any) string {
    	switch v.(type) {
    	case nil:
    		return "nil"
    	case int, int64, float64:
    		return "number"
    	case string, []byte:
    		return "text"
    	case error:
    		return "error"
    	case fmt.Stringer:
    		return "stringer"
    	default:
    		return "other"
    	}
    }

    type Key string

    func (k Key) String() string { return "key:" + string(k) }

    func main() {
    	items := []any{1, "go", Key("a"), errors.New("boom"), 2.5, nil, []byte("hi"), Key("b")}
    	fmt.Println(OfType[Key](items))
    	fmt.Println(len(OfType[fmt.Stringer](items)))
    	for _, it := range items {
    		fmt.Print(Classify(it), " ")
    	}
    	fmt.Println()
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"slices"
    	"testing"
    )

    // loud is both an error and a Stringer.
    type loud struct{}

    func (loud) Error() string  { return "loud error" }
    func (loud) String() string { return "LOUD" }

    func TestOfType(t *testing.T) {
    	items := []any{1, "go", Key("a"), 2, nil, Key("b"), int64(3)}
    	if got := OfType[Key](items); !slices.Equal(got, []Key{"a", "b"}) {
    		t.Errorf("OfType[Key] = %v, want [a b]", got)
    	}
    	if got := OfType[int](items); !slices.Equal(got, []int{1, 2}) {
    		t.Errorf("OfType[int] = %v, want [1 2] (int64(3) is not an int)", got)
    	}
    	if got := OfType[fmt.Stringer](items); len(got) != 2 || got[0].String() != "key:a" {
    		t.Errorf("OfType[fmt.Stringer] = %v, want the two Keys", got)
    	}
    	if got := OfType[bool](items); len(got) != 0 {
    		t.Errorf("OfType[bool] = %v, want none", got)
    	}
    }

    func TestClassify(t *testing.T) {
    	for _, tt := range []struct {
    		v    any
    		want string
    	}{
    		{nil, "nil"},
    		{42, "number"},
    		{int64(7), "number"},
    		{2.5, "number"},
    		{"go", "text"},
    		{[]byte("hi"), "text"},
    		{errors.New("boom"), "error"},
    		{loud{}, "error"},
    		{Key("a"), "stringer"},
    		{true, "other"},
    		{int32(1), "other"},
    		{[]int{1}, "other"},
    	} {
    		if got := Classify(tt.v); got != tt.want {
    			t.Errorf("Classify(%#v) = %q, want %q", tt.v, got, tt.want)
    		}
    	}
    }
---

You know type assertions and type switches from [Learn OOP](/courses/learn-oop/polymorphism/type-assertions-and-switches). Generic code adds new wrinkles: you can't switch on a type parameter value directly, you can switch on the *type* itself, and type parameters can appear in cases.

## Rules worth remembering

- **Single-type case**: the variable has that type. `case string:` gives you a `string`.
- **Multi-type case or `default`**: the variable keeps the switched expression's type (`any`). `case int, int64:` can't pick one for you.
- **`case nil`** matches only a nil interface value, one with no dynamic type.
- **First match wins**, and interface cases match anything implementing them. So order matters: put `case error:` before `case fmt.Stringer:` if a value might be both.

## Switching in generic code

A value of type `T` isn't an interface, so `switch v.(type)` is a compile error: `cannot use type switch on type parameter value v`. Convert to `any` first:

```go
switch x := any(v).(type) {
case string:
	// ...
}
```

That works on the **dynamic** value. Sometimes you want to switch on the **type** `T` itself, with no value around, or with a value that might be a nil interface. The trick is a nil *pointer* to `T`, which always carries its type:

```go
package main

import (
	"fmt"
)

func Kind[T any]() string {
	switch any((*T)(nil)).(type) {
	case *int, *int64:
		return "integer"
	case *string:
		return "string"
	case *error:
		return "error interface"
	}
	return "something else"
}

func KindBad[T any]() string {
	var zero T
	switch any(zero).(type) {
	case int:
		return "int"
	case error:
		return "error"
	case nil:
		return "nil!"
	}
	return "other"
}

func Match[T any](v any) string {
	switch x := v.(type) {
	case T:
		return fmt.Sprintf("a T: %v", x)
	case []T:
		return fmt.Sprintf("%d Ts", len(x))
	case int:
		return "an int"
	}
	return "no match"
}

func main() {
	fmt.Println(Kind[int](), Kind[string](), Kind[error](), Kind[bool]())
	fmt.Println(KindBad[int](), KindBad[error]())
	fmt.Println(Match[string]("hi"), Match[string]([]string{"a", "b"}), Match[int](3), Match[string](2.5))
}
```

```
integer string error interface something else
int nil!
a T: hi 2 Ts a T: 3 no match
```

Things to notice:

- **`Kind[error]()` works**: `(*error)(nil)` is a typed nil pointer, so `case *error` matches.
- **`KindBad[error]()` doesn't**: the zero value of `error` is a nil interface, and `any(nil interface)` has no type at all, so it lands in `case nil`. Watch for this whenever `T` could be an interface type.
- **Type parameters can appear in cases.** In `Match`, `case T:` and `case []T:` are resolved per instantiation. With `T = int`, `case T` and `case int` are duplicates, which is allowed in this situation; the first one wins, so `Match[int](3)` says "a T".

## Asserting to a type parameter

A plain type assertion to `T` works too: `x, ok := v.(T)`. If `T` is a concrete type it checks for exactly that type; if `T` is an interface type, it checks whether the dynamic type **implements** it. That's how the Learn OOP inventory's `First[T Item]()` generic method found items by type, and it's the core of this lesson's exercise.

## When a type switch is the right tool

As chapter 4 warned, don't use `any` plus a switch to fake a constraint. Type switches are right when:

- the input is **genuinely heterogeneous** (decoded JSON, event payloads, `[]any` from an API);
- you're adding **optional fast paths** with a correct general fallback (as `fmt` does for `Stringer`);
- you're checking for **optional interfaces** (last lesson).

## Your turn

Two helpers for Stash's heterogeneous data:

- `OfType[T](items)`: returns the items that are a `T` (or implement it, when `T` is an interface), in order. Use `it.(T)`.
- `Classify(v)`: returns `"nil"`, `"number"` (`int`, `int64`, `float64`), `"text"` (`string`, `[]byte`), `"error"`, `"stringer"`, or `"other"`. Something that is both an `error` and a `Stringer` is an `"error"`.
