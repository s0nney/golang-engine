---
title: Config Fallbacks
difficulty: easy
after: designing-generic-apis
hints:
  - 'Every type has a zero value, and inside a generic function you get it with `var zero T`. Because `T` is `comparable`, you can then write `v != zero`.'
  - '`FirstSet` can''t compare `T` values at all (its constraint is `any`), and it doesn''t need to: a *pointer* is set when it isn''t `nil`. Return `*p, true` for the first non-nil `p`.'
  - 'A non-nil pointer to a zero value (like `new(0)`) still counts as set for `FirstSet`. That''s the whole point of the comma-ok result: `0, true` means "explicitly 0", `0, false` means "not set anywhere".'
exercise:
  starter: |
    package main

    import "fmt"

    // Coalesce returns the first value in vs that isn't the zero value of T.
    // If every value is zero (or vs is empty), it returns the zero value.
    func Coalesce[T comparable](vs ...T) T {
    	// Declare the zero value of T, then look for the first v that differs from it.
    	var zero T
    	return zero
    }

    // FirstSet returns the value behind the first non-nil pointer in ptrs,
    // and true. If every pointer is nil, it returns the zero value and false.
    func FirstSet[T any](ptrs ...*T) (T, bool) {
    	// A pointer is "set" when it isn't nil, even if it points at a zero value.
    	var zero T
    	return zero, false
    }

    func main() {
    	flagPort, envPort, defaultPort := 0, 8080, 6379
    	fmt.Println(Coalesce(flagPort, envPort, defaultPort)) // want 8080

    	var flagTTL *int
    	// The environment explicitly sets the TTL to 0.
    	envTTL := new(0)
    	fmt.Println(FirstSet(flagTTL, envTTL, new(60))) // want 0 true
    }
  solution: |
    package main

    import "fmt"

    // Coalesce returns the first value in vs that isn't the zero value of T.
    // If every value is zero (or vs is empty), it returns the zero value.
    func Coalesce[T comparable](vs ...T) T {
    	var zero T
    	for _, v := range vs {
    		if v != zero {
    			return v
    		}
    	}
    	return zero
    }

    // FirstSet returns the value behind the first non-nil pointer in ptrs,
    // and true. If every pointer is nil, it returns the zero value and false.
    func FirstSet[T any](ptrs ...*T) (T, bool) {
    	for _, p := range ptrs {
    		if p != nil {
    			return *p, true
    		}
    	}
    	var zero T
    	return zero, false
    }

    func main() {
    	flagPort, envPort, defaultPort := 0, 8080, 6379
    	fmt.Println(Coalesce(flagPort, envPort, defaultPort))

    	var flagTTL *int
    	envTTL := new(0)
    	fmt.Println(FirstSet(flagTTL, envTTL, new(60)))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    type Region string

    type Limits struct {
    	MaxKeys int
    	Evict   bool
    }

    func TestCoalesceInts(t *testing.T) {
    	tests := []struct {
    		in   []int
    		want int
    	}{
    		{nil, 0},
    		{[]int{0, 0, 0}, 0},
    		{[]int{0, 8080, 6379}, 8080},
    		{[]int{3000, 8080}, 3000},
    		{[]int{0, 0, -1}, -1},
    	}
    	for _, tt := range tests {
    		if got := Coalesce(tt.in...); got != tt.want {
    			t.Errorf("Coalesce(%v...) = %d, want %d", tt.in, got, tt.want)
    		}
    	}
    }

    func TestCoalesceOtherTypes(t *testing.T) {
    	if got := Coalesce("", "", "eu-west"); got != "eu-west" {
    		t.Errorf("Coalesce(\"\", \"\", \"eu-west\") = %q, want %q", got, "eu-west")
    	}
    	if got := Coalesce[Region]("", "us-east", "eu-west"); got != "us-east" {
    		t.Errorf("Coalesce[Region](\"\", \"us-east\", \"eu-west\") = %q, want %q", got, "us-east")
    	}
    	if got := Coalesce(Limits{}, Limits{MaxKeys: 0, Evict: true}, Limits{MaxKeys: 5}); got != (Limits{Evict: true}) {
    		t.Errorf("Coalesce(Limits...) = %+v, want {MaxKeys:0 Evict:true}: a struct is zero only when every field is zero", got)
    	}
    	if got := Coalesce(0.0, 2.5); got != 2.5 {
    		t.Errorf("Coalesce(0.0, 2.5) = %v, want 2.5", got)
    	}
    	if got := Coalesce(false, false); got {
    		t.Errorf("Coalesce(false, false) = true, want false")
    	}
    	x := 7
    	if got := Coalesce(nil, &x); got != &x {
    		t.Errorf("Coalesce(nil, &x) = %v, want &x", got)
    	}
    }

    func TestFirstSet(t *testing.T) {
    	if got, ok := FirstSet[int](); got != 0 || ok {
    		t.Errorf("FirstSet[int]() = %d, %v, want 0, false", got, ok)
    	}
    	if got, ok := FirstSet[int](nil, nil); got != 0 || ok {
    		t.Errorf("FirstSet[int](nil, nil) = %d, %v, want 0, false", got, ok)
    	}
    	if got, ok := FirstSet(nil, new(0), new(60)); got != 0 || !ok {
    		t.Errorf("FirstSet(nil, new(0), new(60)) = %d, %v, want 0, true: a pointer to 0 is still set", got, ok)
    	}
    	if got, ok := FirstSet(nil, new(30), new(60)); got != 30 || !ok {
    		t.Errorf("FirstSet(nil, new(30), new(60)) = %d, %v, want 30, true", got, ok)
    	}
    	if got, ok := FirstSet(new(Region("")), new(Region("us"))); got != "" || !ok {
    		t.Errorf("FirstSet(new(Region(\"\")), ...) = %q, %v, want \"\", true", got, ok)
    	}
    	if got, ok := FirstSet(nil, &Limits{MaxKeys: 9}); got != (Limits{MaxKeys: 9}) || !ok {
    		t.Errorf("FirstSet(nil, &Limits{MaxKeys: 9}) = %+v, %v, want {MaxKeys:9 Evict:false}, true", got, ok)
    	}
    }

    func TestFirstSetNonComparable(t *testing.T) {
    	// Slices aren't comparable, but FirstSet only compares pointers.
    	a := []string{"x", "y"}
    	got, ok := FirstSet(nil, &a)
    	if !ok || !slices.Equal(got, a) {
    		t.Errorf("FirstSet(nil, &%q) = %q, %v, want %q, true", a, got, ok, a)
    	}
    	var none *map[string]int
    	if m, ok := FirstSet(none); m != nil || ok {
    		t.Errorf("FirstSet(nil *map) = %v, %v, want nil map, false", m, ok)
    	}
    }
---

Stash reads each setting from several places: a command-line flag, an
environment variable, then a built-in default. It needs two small helpers to
pick the winner.

- `Coalesce(vs...)` returns the first value that isn't the **zero value** of its
  type (like SQL's `COALESCE`). If they're all zero, it returns the zero value.
- `FirstSet(ptrs...)` is for settings where the zero value is meaningful, like a
  TTL of `0`. It returns the value behind the first **non-nil pointer**, and
  `true`. If every pointer is `nil`, it returns the zero value and `false`.

## Examples

```go
Coalesce(0, 8080, 6379)                   // 8080
Coalesce("", "", "eu-west")               // "eu-west"
Coalesce(0, 0)                            // 0

FirstSet(nil, new(0), new(60))            // 0, true (explicitly zero)
FirstSet[int](nil, nil)                   // 0, false (not set anywhere)
```

## Constraints

- `Coalesce` works for any `comparable` type: numbers, strings, your own named
  types, structs, pointers.
- `FirstSet` works for **any** type, including ones that can't be compared,
  like slices and maps. Don't change its constraint.
