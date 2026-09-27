---
title: Constraints and cmp.Ordered
quiz:
  - question: |
      Why doesn't this compile?

      ```go
      func biggest[T any](a, b T) T {
      	if a > b {
      		return a
      	}
      	return b
      }
      ```
    options:
      - text: Generic functions can't use `if`
      - text: '`any` allows types that can''t be compared with `>`, such as structs and slices'
        correct: true
      - text: Type parameters must be called `K` or `V`
    explanation: |
      The body may only do what *every* type allowed by the constraint
      supports. `any` includes types with no ordering, so `>` is rejected.
      Use `cmp.Ordered` as the constraint instead.
  - question: 'Which constraint lets you use `==` on values of type `T`, but not `<`?'
    options:
      - text: '`any`'
      - text: '`comparable`'
        correct: true
      - text: '`cmp.Ordered`'
    explanation: |
      `comparable` is built in and covers every type that supports `==` and
      `!=`, which is exactly what map keys need. `cmp.Ordered` adds `<`,
      `<=`, `>` and `>=`, and `any` allows neither.
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    // largest returns the largest value and true, or the zero value
    // and false if values is empty.
    func largest[T cmp.Ordered](values []T) (T, bool) {
    	var zero T
    	// ?
    	return zero, false
    }

    func main() {
    	fmt.Println(largest([]int{42, 160, 7}))                 // want 160 true
    	fmt.Println(largest([]string{"bob", "carol", "alice"})) // want carol true
    	fmt.Println(largest([]float64{}))                       // want 0 false
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    func largest[T cmp.Ordered](values []T) (T, bool) {
    	var zero T
    	if len(values) == 0 {
    		return zero, false
    	}
    	best := values[0]
    	for _, v := range values[1:] {
    		if v > best {
    			best = v
    		}
    	}
    	return best, true
    }

    func main() {
    	fmt.Println(largest([]int{42, 160, 7}))
    	fmt.Println(largest([]string{"bob", "carol", "alice"}))
    	fmt.Println(largest([]float64{}))
    }
  tests: |
    package main

    import "testing"

    func TestLargestInts(t *testing.T) {
    	tests := []struct {
    		values []int
    		want   int
    		wantOK bool
    	}{
    		{[]int{42, 160, 7}, 160, true},
    		{[]int{-5, -2, -9}, -2, true},
    		{[]int{3}, 3, true},
    		{nil, 0, false},
    	}
    	for _, tt := range tests {
    		got, ok := largest(tt.values)
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("largest(%v) = %v, %v; want %v, %v", tt.values, got, ok, tt.want, tt.wantOK)
    		}
    	}
    }

    func TestLargestStrings(t *testing.T) {
    	got, ok := largest([]string{"bob", "carol", "alice"})
    	if got != "carol" || !ok {
    		t.Errorf(`largest(["bob" "carol" "alice"]) = %q, %v; want "carol", true`, got, ok)
    	}
    	got, ok = largest([]string{})
    	if got != "" || ok {
    		t.Errorf(`largest([]) = %q, %v; want "", false`, got, ok)
    	}
    }

    func TestLargestFloats(t *testing.T) {
    	got, ok := largest([]float64{0.01, 0.05, 0.02})
    	if got != 0.05 || !ok {
    		t.Errorf("largest([0.01 0.05 0.02]) = %v, %v; want 0.05, true", got, ok)
    	}
    }
---

A generic function can only do things that work for every type its constraint allows. With `any`, that's not much. **Constraints** narrow down the allowed types, which unlocks more operations.

## Constraints are interfaces

A constraint is written as an **interface**. Besides listing methods, an interface used as a constraint can list the **types** it allows, separated by `|`:

```go
package main

import "fmt"

type Number interface {
	~int | ~int64 | ~float64
}

func sum[T Number](values []T) T {
	var total T
	for _, v := range values {
		total += v
	}
	return total
}

func main() {
	fmt.Println(sum([]int{160, 42, 7}))
	fmt.Println(sum([]float64{0.01, 0.02}))
}
```

```text
209
0.03
```

Every type in `Number` supports `+`, so `total += v` compiles. `var total T` starts at the zero value of whatever `T` is.

The `~` means "this type, **or any type whose underlying type is** this type". So `~int` also allows your own types like `type credits int`. Without the `~`, `credits` wouldn't be allowed.

## `comparable`

`comparable` is a built-in constraint for types that support `==` and `!=`. You need it for map keys, or for searching:

```go
package main

import "fmt"

func contains[T comparable](items []T, target T) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func main() {
	optedOut := []string{"+1-555-0199", "+1-555-0142"}
	fmt.Println(contains(optedOut, "+1-555-0142"))
	fmt.Println(contains([]int{1, 2, 3}, 5))
}
```

```text
true
false
```

(That's essentially how `slices.Contains` is written.)

## `cmp.Ordered`

For `<` and `>`, the standard library's `cmp` package provides `cmp.Ordered`: every integer, float and string type. It's the constraint you'll use most often:

```go
package main

import (
	"cmp"
	"fmt"
)

func largest[T cmp.Ordered](values []T) T {
	best := values[0]
	for _, v := range values[1:] {
		if v > best {
			best = v
		}
	}
	return best
}

func main() {
	fmt.Println(largest([]int{42, 160, 7}))
	fmt.Println(largest([]string{"bob", "alice", "carol"}))
}
```

```text
160
carol
```

Strings are ordered alphabetically (more precisely, byte by byte), so `"carol"` is the largest.

## `cmp.Compare`

The `cmp` package also has `cmp.Compare(a, b)`, which returns `-1` if `a < b`, `0` if they're equal and `+1` if `a > b`. It's exactly what `slices.SortFunc` wants, which makes sorting structs by a field a one-liner:

```go
package main

import (
	"cmp"
	"fmt"
	"slices"
)

type customer struct {
	name string
	sent int
}

func main() {
	customers := []customer{
		{"alice", 120},
		{"bob", 45},
		{"carol", 300},
	}
	slices.SortFunc(customers, func(a, b customer) int {
		return cmp.Compare(b.sent, a.sent) // b first: biggest senders first
	})
	fmt.Println(customers)
}
```

```text
[{carol 300} {alice 120} {bob 45}]
```

## Choosing a constraint

| Constraint | Allows | You can use |
|------------|--------|-------------|
| `any` | Every type | Assignment, passing around |
| `comparable` | Types with `==` | `==`, `!=`, map keys |
| `cmp.Ordered` | Numbers and strings | `==`, `<`, `>` and friends |
| Your own interface | Whatever you list | Whatever all of them support |

Pick the **loosest** constraint that lets your code compile. The looser it is, the more types your function works with.

## Your turn

Complete the generic `largest` function. It should work for any ordered type (ints, floats, strings) and return two values: the largest element and `true`, or the zero value and `false` when the slice is empty.

Watch out: negative numbers! Starting your "best so far" at `0` gives the wrong answer for `[]int{-5, -2, -9}`. Start from the first element instead.

## Further reading

- [Go by Example: Generics](https://gobyexample.com/generics)
- [An Introduction To Generics (Go blog)](https://go.dev/blog/intro-generics)
