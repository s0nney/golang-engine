---
title: Assignability and Conversion
quiz:
  - question: |
      Given `type Tags []string`, which line does **not** compile?
    options:
      - text: '`var t Tags = []string{"go"}`'
      - text: '`var s []string = Tags{"go"}`'
      - text: '`var t Tags = nil`'
      - text: '`var u []Key = []string{"go"}` (with `type Key string`)'
        correct: true
    explanation: |
      The first two work because `Tags` and `[]string` share an underlying type and one
      side is a type literal. `nil` is assignable to any slice type. But `[]Key` and
      `[]string` have *different* underlying types (their element types differ), so no
      assignment or conversion connects them. You need a loop.
  - question: |
      Two struct types differ only in their field tags:

      ```go
      type A struct {
      	Name string `json:"name"`
      }

      type B struct {
      	Name string `stash:"name"`
      }
      ```

      Can you write `B(a)` for `a` of type `A`?
    options:
      - text: Yes, conversion ignores struct tags
        correct: true
      - text: No, tags are part of the type, so the types aren't convertible
      - text: Only with `unsafe`
      - text: Only if both structs have the same tags
    explanation: |
      Tags *are* part of struct type identity, so `A` and `B` aren't assignable to each
      other. But the conversion rules explicitly ignore tags when comparing underlying
      types, so an explicit `B(a)` is fine.
exercise:
  starter: |
    package main

    import "fmt"

    // Key is a Stash key.
    type Key string

    // Keys is a list of keys.
    type Keys []Key

    // Strings returns the keys as plain strings, in the same order.
    // It returns a new slice; []string(ks) doesn't compile!
    func (ks Keys) Strings() []string {
    	// ?
    	return nil
    }

    // FromStrings converts plain strings into Keys, in the same order.
    func FromStrings(ss []string) Keys {
    	// ?
    	return nil
    }

    // Join joins the keys with sep, like strings.Join does for strings.
    func (ks Keys) Join(sep string) string {
    	// ?
    	return ""
    }

    func main() {
    	ks := FromStrings([]string{"user:1", "user:2"})
    	fmt.Println(len(ks), ks.Strings())
    	fmt.Println(ks.Join(", "))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // Key is a Stash key.
    type Key string

    // Keys is a list of keys.
    type Keys []Key

    // Strings returns the keys as plain strings, in the same order.
    func (ks Keys) Strings() []string {
    	out := make([]string, len(ks))
    	for i, k := range ks {
    		out[i] = string(k)
    	}
    	return out
    }

    // FromStrings converts plain strings into Keys, in the same order.
    func FromStrings(ss []string) Keys {
    	out := make(Keys, len(ss))
    	for i, s := range ss {
    		out[i] = Key(s)
    	}
    	return out
    }

    // Join joins the keys with sep, like strings.Join does for strings.
    func (ks Keys) Join(sep string) string {
    	return strings.Join(ks.Strings(), sep)
    }

    func main() {
    	ks := FromStrings([]string{"user:1", "user:2"})
    	fmt.Println(len(ks), ks.Strings())
    	fmt.Println(ks.Join(", "))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestFromStrings(t *testing.T) {
    	ks := FromStrings([]string{"a", "b", "c"})
    	want := Keys{"a", "b", "c"}
    	if !slices.Equal(ks, want) {
    		t.Fatalf("FromStrings([a b c]) = %q, want %q", ks, want)
    	}
    	if got := FromStrings(nil); len(got) != 0 {
    		t.Errorf("FromStrings(nil) = %q, want an empty Keys", got)
    	}
    }

    func TestStrings(t *testing.T) {
    	ks := Keys{"user:1", "user:2", "order:9"}
    	got := ks.Strings()
    	want := []string{"user:1", "user:2", "order:9"}
    	if !slices.Equal(got, want) {
    		t.Fatalf("Keys%q.Strings() = %q, want %q", ks, got, want)
    	}
    	got[0] = "changed"
    	if ks[0] != "user:1" {
    		t.Errorf("Strings() must return a new slice, but changing it changed the Keys")
    	}
    }

    func TestJoin(t *testing.T) {
    	for _, tt := range []struct {
    		ks   Keys
    		sep  string
    		want string
    	}{
    		{Keys{"a", "b"}, ", ", "a, b"},
    		{Keys{"solo"}, "|", "solo"},
    		{nil, ",", ""},
    	} {
    		if got := tt.ks.Join(tt.sep); got != tt.want {
    			t.Errorf("Keys%q.Join(%q) = %q, want %q", tt.ks, tt.sep, got, tt.want)
    		}
    	}
    }
---

Named types are safe because the compiler refuses to mix them. This lesson pins down exactly when it *does* let a value move from one type to another: implicitly (**assignability**) or with an explicit `T(x)` (**conversion**).

## Assignability: when no conversion is needed

A value `x` of type `V` can be assigned to a variable of type `T` (or passed as a `T` argument, or returned as a `T`) when one of these holds:

1. `V` and `T` are **identical**.
2. `V` and `T` have **identical underlying types** and **at least one of them is not a named type**.
3. `T` is an **interface** and `x` implements it.
4. `x` is **`nil`** and `T` is a pointer, function, slice, map, channel or interface type.
5. `x` is an **untyped constant** representable as a `T` (next lesson).
6. A channel rule you'll rarely need: a bidirectional `chan E` is assignable to an unnamed `<-chan E` or `chan<- E`.

Rule 2 is the interesting one. It's why you can pass a `[]string` literal to a function taking `Tags`, and a `Tags` to `strings.Join`:

```go
package main

import (
	"fmt"
	"strings"
)

type Tags []string

func describe(t Tags) string { return strings.Join(t, "+") } // Tags -> []string

func main() {
	fmt.Println(describe([]string{"go", "types"})) // []string -> Tags
}
```

```
go+types
```

But two *named* types never meet under rule 2, even with the same underlying type. That's why `UserID` and `OrderID` couldn't be mixed.

## Conversion: when you say so

When assignment isn't allowed, an explicit conversion `T(x)` may be. The main cases:

- anything assignable is also convertible;
- `V` and `T` have **identical underlying types, ignoring struct tags** (so `UserID(o)` works, and so does converting between two structs that differ only in tags);
- both are **numeric** types (`float64(n)`, `int8(x)`, which may truncate or wrap);
- **strings** to and from `[]byte` and `[]rune`;
- a **slice to an array** or array pointer, since Go 1.20: `[4]byte(s)` (panics if `s` is too short).

## The famous gap: `[]Key` is not `[]string`

Here's where people get stuck. With `type Key string`, you can convert one `Key` to a `string`, but not a whole slice:

```go
type Key string

keys := []Key{"a", "b"}
ss := []string(keys) // compile error: cannot convert keys (variable of type []Key) to type []string
```

The underlying type of `[]Key` is `[]Key` (it's already a type literal), and the underlying type of `[]string` is `[]string`. Their element types differ, so the underlying types differ, and no rule applies. Same story for `map[Key]int` vs `map[string]int` and `func(Key)` vs `func(string)`.

It's not the compiler being fussy. `Key` might have methods, might be used as a different kind of thing, and one day it might not even be a string. Converting element by element keeps every conversion visible.

## Your turn

Stash represents keys with `type Key string` and lists of them with `type Keys []Key`. Complete:

- `(Keys) Strings() []string`: a **new** slice with each key converted to `string`, same order.
- `FromStrings([]string) Keys`: the reverse.
- `(Keys) Join(sep string) string`: like `strings.Join`, for keys. (Hint: reuse `Strings`.)

## Further reading

- [The Go spec: Assignability](https://go.dev/ref/spec#Assignability) and [Conversions](https://go.dev/ref/spec#Conversions)
