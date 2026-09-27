---
title: What You Can Do With a T
quiz:
  - question: |
      Does this compile?

      ```go
      type User struct{ Name string }
      type Admin struct{ Name string }

      func Names[T User | Admin](xs []T) []string {
      	var out []string
      	for _, x := range xs {
      		out = append(out, x.Name)
      	}
      	return out
      }
      ```
    options:
      - text: Yes, because both types have a `Name` field
      - text: 'No: you can''t access fields through a type parameter, even if every type in the set has that field'
        correct: true
      - text: No, because unions can't contain struct types
      - text: Yes, but only since Go 1.25
    explanation: |
      Field access on type parameters isn't supported at all:
      `x.Name undefined (type T has no field or method Name)`. Give both types a
      `GetName() string` method and use a method constraint instead.
  - question: |
      With `T ~string | ~[]byte`, which operation is **not** allowed on `s T`?
    options:
      - text: '`len(s)`'
      - text: '`s[0]`'
      - text: '`for _, c := range s`'
        correct: true
      - text: '`string(s)`'
    explanation: |
      Ranging over a string yields runes, and ranging over a `[]byte` yields bytes. They
      don't agree, so the compiler refuses: `string and []byte have different underlying
      types`. `len`, indexing (a byte either way), slicing and conversion to `string` all
      behave the same for both, so they're allowed.
exercise:
  starter: |
    package main

    import "fmt"

    // Text is anything made of bytes: strings, byte slices, and named
    // types built from them.
    type Text interface {
    	~string | ~[]byte
    }

    // Key is a Stash key.
    type Key string

    // CountByte reports how many times b appears in s.
    func CountByte[T Text](s T, b byte) int {
    	// ?
    	return 0
    }

    // CutPrefix returns s without prefix and true if s starts with prefix.
    // Otherwise it returns s unchanged and false.
    func CutPrefix[T Text](s, prefix T) (T, bool) {
    	// ?
    	return s, false
    }

    func main() {
    	fmt.Println(CountByte("user:42:profile", ':'))
    	rest, ok := CutPrefix(Key("user:42"), "user:")
    	fmt.Printf("%T %q %v\n", rest, rest, ok)
    	b, ok := CutPrefix([]byte("v1/items"), []byte("v1/"))
    	fmt.Println(string(b), ok)
    }
  solution: |
    package main

    import "fmt"

    // Text is anything made of bytes: strings, byte slices, and named
    // types built from them.
    type Text interface {
    	~string | ~[]byte
    }

    // Key is a Stash key.
    type Key string

    // CountByte reports how many times b appears in s.
    func CountByte[T Text](s T, b byte) int {
    	n := 0
    	for i := range len(s) {
    		if s[i] == b {
    			n++
    		}
    	}
    	return n
    }

    // CutPrefix returns s without prefix and true if s starts with prefix.
    // Otherwise it returns s unchanged and false.
    func CutPrefix[T Text](s, prefix T) (T, bool) {
    	if len(s) >= len(prefix) && string(s[:len(prefix)]) == string(prefix) {
    		return s[len(prefix):], true
    	}
    	return s, false
    }

    func main() {
    	fmt.Println(CountByte("user:42:profile", ':'))
    	rest, ok := CutPrefix(Key("user:42"), "user:")
    	fmt.Printf("%T %q %v\n", rest, rest, ok)
    	b, ok := CutPrefix([]byte("v1/items"), []byte("v1/"))
    	fmt.Println(string(b), ok)
    }
  tests: |
    package main

    import "testing"

    func TestCountByte(t *testing.T) {
    	if got := CountByte("user:42:profile", ':'); got != 2 {
    		t.Errorf(`CountByte("user:42:profile", ':') = %d, want 2`, got)
    	}
    	if got := CountByte([]byte("banana"), 'a'); got != 3 {
    		t.Errorf(`CountByte([]byte("banana"), 'a') = %d, want 3`, got)
    	}
    	if got := CountByte(Key(""), 'x'); got != 0 {
    		t.Errorf(`CountByte(Key(""), 'x') = %d, want 0`, got)
    	}
    	if got := CountByte("héllo", 'l'); got != 2 {
    		t.Errorf(`CountByte("héllo", 'l') = %d, want 2`, got)
    	}
    }

    func TestCutPrefix(t *testing.T) {
    	rest, ok := CutPrefix(Key("user:42"), "user:")
    	if !ok || rest != "42" {
    		t.Errorf(`CutPrefix(Key("user:42"), "user:") = (%q, %v), want ("42", true)`, rest, ok)
    	}
    	rest, ok = CutPrefix(Key("order:7"), "user:")
    	if ok || rest != "order:7" {
    		t.Errorf(`CutPrefix(Key("order:7"), "user:") = (%q, %v), want ("order:7", false)`, rest, ok)
    	}
    	rest, ok = CutPrefix(Key("us"), "user:")
    	if ok || rest != "us" {
    		t.Errorf(`CutPrefix(Key("us"), "user:") = (%q, %v), want ("us", false)`, rest, ok)
    	}
    	b, ok := CutPrefix([]byte("v1/items"), []byte("v1/"))
    	if !ok || string(b) != "items" {
    		t.Errorf(`CutPrefix([]byte("v1/items"), []byte("v1/")) = (%q, %v), want ("items", true)`, b, ok)
    	}
    	all, ok := CutPrefix("same", "same")
    	if !ok || all != "" {
    		t.Errorf(`CutPrefix("same", "same") = (%q, %v), want ("", true)`, all, ok)
    	}
    }
---

A constraint is a contract in two directions. Callers may pass any type in the type set. In return, the body may only do things that work for **every** type in the set. This lesson makes "things that work" precise.

## The rule

An operation on a value of type parameter type `T` is allowed if it's valid for **all types in `T`'s type set**, with the same meaning for each. Walking through the constraint you pick:

| Constraint | What the body can do |
|---|---|
| `any` | declare, assign, pass and return values; `var zero T`; take `&v`; convert to an interface (`any(v)`) and type-assert from one |
| `comparable` | everything above, plus `==` and `!=`, and use `T` as a map key |
| methods (`interface{ String() string }`) | call those methods |
| union (`~int \| ~float64`) | operators, built-ins and conversions that every type in the union supports |

## Operators and conversions

With `T ~int | ~float64`, all of `+ - * / < ==` work, and so does converting to and from other numeric types, because each term supports it. Constants are converted per type too:

```go
func Half[T ~int | ~float64](v T) T { return v / T(2) }   // fine: 2 fits both
func Grow[T ~int | ~float64](v T) T { return v * T(1.5) } // error
```

The second fails with `cannot convert 1.5 (untyped float constant) to type int (in T)`: `1.5` isn't a valid `int`, so it isn't valid for every type in the set.

## Built-ins, indexing and range

For `len`, `cap`, indexing, slicing, `range`, `make` and composite literals, the types in the set must agree on the **structure** as well:

- `func Mk[S ~[]E, E any](n int) S { return make(S, n) }` is fine: every type in `S`'s set is a slice of `E`.
- `T ~string | ~[]byte` allows `len(s)`, `s[i]` (a `byte` either way), `s[i:j]` (a `T`), and `string(s)`. But **not** `range s`: a string ranges over runes and a byte slice over bytes. The error is `cannot range over s (variable of type T constrained by ~string | ~[]byte): string and []byte have different underlying types`.
- `T ~[]int | ~[]string` can't be ranged over or indexed usefully either, because the element types differ.

Until Go 1.25 the spec explained these rules with a concept called the **core type** (roughly: the single underlying type shared by the whole type set, with a special case for `string | []byte`). Go 1.25 removed core types from the spec and describes each operation directly. The behaviour you get is essentially the same, but you'll still see "core type" in older blog posts and error messages.

## Things you can never do

- **Access fields.** Even if every type in the set is a struct with a `Name` field, `x.Name` is an error. Use a method constraint.
- **Declare typed constants** of type `T`, like `const limit T = 10`. Constants need a concrete type. Use a variable or a conversion: `T(10)`.
- **Type-switch on `v` directly.** `switch v.(type)` needs an interface. Convert first: `switch any(v).(type)`. Chapter 6 covers this.
- **Call methods that aren't in the constraint.** If `T`'s constraint is `any`, `v.String()` doesn't compile, even if you only ever use it with Stringers.

## Your turn

Stash handles keys that might be strings, `[]byte`, or named types like `Key`. Using the `Text` constraint (`~string | ~[]byte`):

- `CountByte(s, b)` counts occurrences of byte `b` in `s`. You can't `range` over `s`, so loop over indexes with `for i := range len(s)`.
- `CutPrefix(s, prefix)` returns `s` without the prefix and `true`, or `s` unchanged and `false`. Slicing `s[a:b]` gives you a `T`; compare contents with `string(...)`, because `==` isn't defined for `[]byte`.
