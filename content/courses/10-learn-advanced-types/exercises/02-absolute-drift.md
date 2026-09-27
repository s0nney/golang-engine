---
title: Absolute Drift
difficulty: easy
after: type-sets-and-constraints
hints:
  - 'A union like `int | int64` only admits those exact types. To also admit a defined type such as `type Drift int32`, put a tilde in front: `~int32` means "any type whose underlying type is `int32`".'
  - 'List every signed integer and float kind: `~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64`. Unsigned types are left out on purpose: they are never negative.'
  - 'Inside `Abs`, every type in the set supports `<` and unary `-`, and the untyped constant `0` converts to all of them, so `if v < 0 { return -v }` compiles.'
exercise:
  starter: |
    package main

    import "fmt"

    // Signed is the set of signed number types, including defined types
    // built on them (like Drift below).
    type Signed interface {
    	// Replace this with a union of every signed integer and float type.
    	// Remember the tilde!
    	int
    }

    // Abs returns the absolute value of v, with the same type as v.
    func Abs[T Signed](v T) T {
    	// ?
    	return v
    }

    // Drift is how far a replica lags behind, in milliseconds.
    type Drift int32

    func main() {
    	fmt.Println(Abs(-7)) // want 7
    	// Uncomment these once Signed admits floats and Drift:
    	// fmt.Println(Abs(2.5), Abs(-2.5)) // want 2.5 2.5
    	// fmt.Printf("%T %v\n", Abs(Drift(-40)), Abs(Drift(-40))) // want main.Drift 40
    }
  solution: |
    package main

    import "fmt"

    // Signed is the set of signed number types, including defined types
    // built on them (like Drift below).
    type Signed interface {
    	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
    }

    // Abs returns the absolute value of v, with the same type as v.
    func Abs[T Signed](v T) T {
    	if v < 0 {
    		return -v
    	}
    	return v
    }

    // Drift is how far a replica lags behind, in milliseconds.
    type Drift int32

    func main() {
    	fmt.Println(Abs(-7))
    	fmt.Println(Abs(2.5), Abs(-2.5))
    	fmt.Printf("%T %v\n", Abs(Drift(-40)), Abs(Drift(-40)))
    }
  tests: |
    package main

    import (
    	"math"
    	"testing"
    )

    type Offset int8
    type Latency float32
    type Skew int64
    type Balance float64

    func TestAbsBuiltinTypes(t *testing.T) {
    	if got := Abs(-7); got != 7 {
    		t.Errorf("Abs(-7) = %d, want 7", got)
    	}
    	if got := Abs(7); got != 7 {
    		t.Errorf("Abs(7) = %d, want 7", got)
    	}
    	if got := Abs(0); got != 0 {
    		t.Errorf("Abs(0) = %d, want 0", got)
    	}
    	if got := Abs(int16(-300)); got != 300 {
    		t.Errorf("Abs(int16(-300)) = %d, want 300", got)
    	}
    	if got := Abs(int64(math.MinInt64 + 1)); got != math.MaxInt64 {
    		t.Errorf("Abs(int64(MinInt64+1)) = %d, want %d", got, int64(math.MaxInt64))
    	}
    	if got := Abs(-2.5); got != 2.5 {
    		t.Errorf("Abs(-2.5) = %v, want 2.5", got)
    	}
    	if got := Abs(float32(-0.5)); got != 0.5 {
    		t.Errorf("Abs(float32(-0.5)) = %v, want 0.5", got)
    	}
    }

    func TestAbsDefinedTypes(t *testing.T) {
    	if got := Abs(Drift(-40)); got != 40 {
    		t.Errorf("Abs(Drift(-40)) = %v, want 40", got)
    	}
    	if got := Abs(Offset(-128 + 1)); got != 127 {
    		t.Errorf("Abs(Offset(-127)) = %v, want 127", got)
    	}
    	if got := Abs(Latency(-1.25)); got != 1.25 {
    		t.Errorf("Abs(Latency(-1.25)) = %v, want 1.25", got)
    	}
    	if got := Abs(Skew(99)); got != 99 {
    		t.Errorf("Abs(Skew(99)) = %v, want 99", got)
    	}
    	if got := Abs(Balance(-1e9)); got != 1e9 {
    		t.Errorf("Abs(Balance(-1e9)) = %v, want 1e9", got)
    	}
    	// The result keeps the argument's type.
    	var d Drift = Abs(Drift(-3))
    	if d != 3 {
    		t.Errorf("Abs(Drift(-3)) = %v, want 3", d)
    	}
    }
---

Stash's replication monitor tracks **drift**: how far each replica is ahead of
or behind the primary. The dashboard only cares about the size of the drift, so
it needs an absolute value that works for every signed number type, including
Stash's own defined types such as `Drift`.

1. Replace the body of the `Signed` constraint with a type set that holds every
   signed integer and floating-point type (`int`, `int8`, `int16`, `int32`,
   `int64`, `float32`, `float64`) **and every defined type built on them**.
2. Implement `Abs(v)`: return `-v` for negative values and `v` otherwise. The
   result has the same type as `v`.

## Examples

```go
Abs(-7)           // 7 (int)
Abs(-2.5)         // 2.5 (float64)
Abs(Drift(-40))   // 40, and its type is still Drift
```

## Constraints

- Unsigned types stay out of `Signed`: they can't be negative, so `Abs` makes
  no sense for them.
- The hidden tests call `Abs` with defined types of their own, such as
  `type Offset int8` and `type Latency float32`, so the constraint must accept
  any type whose **underlying** type is in the list.
- You don't need to handle the most negative integer (`math.MinInt64` and
  friends), whose absolute value doesn't fit.
