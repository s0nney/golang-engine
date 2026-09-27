---
title: Measuring Generic Code
quiz:
  - question: |
      In the lesson's benchmark, generic `Sum[int]` ran as fast as `SumInts`, but `SumValues[Size]` (calling `x.Value()`) was about five times slower than summing directly. Why?
    options:
      - text: Generic functions are always slower when they have methods
      - text: '`+=` on the `int` shape is a plain add, but `x.Value()` is an indirect call through the dictionary that can''t be inlined'
        correct: true
      - text: '`SumValues` allocates on every call'
      - text: Because `Size` is a named type
    explanation: |
      The benchmark reported 0 allocations for all versions. The difference is the
      method call: in shared shape code it goes through a function pointer, like an
      interface call, so the compiler can't inline `Value` into the loop.
  - question: What's wrong with judging performance from a single `go test -bench` run?
    options:
      - text: Nothing; benchmarks are deterministic
      - text: Results vary between runs (CPU frequency, other processes, GC), so you should run several times (`-count`) and compare distributions, for example with `benchstat`
        correct: true
      - text: '`-bench` only runs each benchmark once'
      - text: Benchmarks don't work on generic code
    explanation: |
      The lesson's own first run showed `SumInts` 10% *slower* than on later runs. Noise
      is normal, which is why Learn Testing recommends `-count` and `benchstat`.
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"slices"
    )

    // Largest returns the largest value in xs, and false if xs is empty.
    // It's correct, but it converts every element to an interface, which
    // allocates. Rewrite it so it doesn't allocate at all.
    func Largest[T cmp.Ordered](xs []T) (T, bool) {
    	var zero T
    	if len(xs) == 0 {
    		return zero, false
    	}
    	boxed := make([]any, len(xs))
    	for i, x := range xs {
    		boxed[i] = x
    	}
    	best := boxed[0]
    	for _, b := range boxed[1:] {
    		if b.(T) > best.(T) {
    			best = b
    		}
    	}
    	return best.(T), true
    }

    // IndexOf returns the index of the first element equal to v, or -1.
    // Same problem: every element gets boxed. Fix it.
    func IndexOf[T comparable](xs []T, v T) int {
    	return slices.Index(toAny(xs), any(v))
    }

    func toAny[T any](xs []T) []any {
    	out := make([]any, len(xs))
    	for i, x := range xs {
    		out[i] = x
    	}
    	return out
    }

    func main() {
    	sizes := []int{512, 4096, 1024}
    	fmt.Println(Largest(sizes))
    	fmt.Println(Largest([]string{"pear", "fig", "plum"}))
    	fmt.Println(IndexOf(sizes, 1024), IndexOf(sizes, 7))
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    // Largest returns the largest value in xs, and false if xs is empty.
    func Largest[T cmp.Ordered](xs []T) (T, bool) {
    	if len(xs) == 0 {
    		var zero T
    		return zero, false
    	}
    	best := xs[0]
    	for _, x := range xs[1:] {
    		if x > best {
    			best = x
    		}
    	}
    	return best, true
    }

    // IndexOf returns the index of the first element equal to v, or -1.
    func IndexOf[T comparable](xs []T, v T) int {
    	for i, x := range xs {
    		if x == v {
    			return i
    		}
    	}
    	return -1
    }

    func main() {
    	sizes := []int{512, 4096, 1024}
    	fmt.Println(Largest(sizes))
    	fmt.Println(Largest([]string{"pear", "fig", "plum"}))
    	fmt.Println(IndexOf(sizes, 1024), IndexOf(sizes, 7))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    )

    func TestLargest(t *testing.T) {
    	if v, ok := Largest([]int{512, 4096, 1024}); !ok || v != 4096 {
    		t.Errorf("Largest([512 4096 1024]) = (%d, %v), want (4096, true)", v, ok)
    	}
    	if v, ok := Largest([]string{"pear", "fig", "plum"}); !ok || v != "plum" {
    		t.Errorf("Largest([pear fig plum]) = (%q, %v), want (\"plum\", true)", v, ok)
    	}
    	if v, ok := Largest([]float64{}); ok || v != 0 {
    		t.Errorf("Largest([]) = (%v, %v), want (0, false)", v, ok)
    	}
    }

    func TestIndexOf(t *testing.T) {
    	if got := IndexOf([]int{512, 4096, 1024, 4096}, 4096); got != 1 {
    		t.Errorf("IndexOf([512 4096 1024 4096], 4096) = %d, want 1", got)
    	}
    	if got := IndexOf([]string{"a", "m", "z"}, "q"); got != -1 {
    		t.Errorf(`IndexOf([a m z], "q") = %d, want -1`, got)
    	}
    }

    func TestNoAllocations(t *testing.T) {
    	ints := make([]int, 1000)
    	strs := make([]string, 1000)
    	for i := range ints {
    		ints[i] = 1000 + i*7919%1000
    		strs[i] = fmt.Sprint("key-", i*7919%1000)
    	}
    	if a := testing.AllocsPerRun(20, func() { Largest(ints) }); a != 0 {
    		t.Errorf("Largest([]int) allocated %v times per call, want 0", a)
    	}
    	if a := testing.AllocsPerRun(20, func() { Largest(strs) }); a != 0 {
    		t.Errorf("Largest([]string) allocated %v times per call, want 0", a)
    	}
    	if a := testing.AllocsPerRun(20, func() { IndexOf(strs, "missing") }); a != 0 {
    		t.Errorf("IndexOf([]string) allocated %v times per call, want 0", a)
    	}
    	if a := testing.AllocsPerRun(20, func() { IndexOf(ints, -1) }); a != 0 {
    		t.Errorf("IndexOf([]int) allocated %v times per call, want 0", a)
    	}
    }
---

Chapter 6 and the last lesson gave you models of what interfaces and generics cost. Models are for predicting; **benchmarks** are for deciding. You learned to write them with `b.Loop` in [Learn Testing](/courses/learn-testing/benchmarks/writing-benchmarks); here we point that at the generics question.

## Four ways to sum

```go
package sum

type Number interface {
	~int | ~int64 | ~float64
}

// SumInts is plain, concrete code.
func SumInts(xs []int) int {
	total := 0
	for _, x := range xs {
		total += x
	}
	return total
}

// Sum is generic over an operator.
func Sum[T Number](xs []T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

// Valuer is the interface-based alternative.
type Valuer interface{ Value() int }

type Size int

func (s Size) Value() int { return int(s) }

func SumValuers(xs []Valuer) int {
	total := 0
	for _, x := range xs {
		total += x.Value()
	}
	return total
}

// SumValues is generic over a method constraint.
func SumValues[T Valuer](xs []T) int {
	total := 0
	for _, x := range xs {
		total += x.Value()
	}
	return total
}
```

And a benchmark per version, each over 1,000 elements:

```go
func BenchmarkGeneric(b *testing.B) {
	xs := make([]int, 1000)
	for b.Loop() {
		Sum(xs)
	}
}

func BenchmarkGenericMethod(b *testing.B) {
	xs := make([]Size, 1000)
	for i := range xs {
		xs[i] = Size(i)
	}
	for b.Loop() {
		SumValues(xs)
	}
}

// ...and the same for SumInts and SumValuers
```

`b.Loop` keeps the setup out of the timing and stops the compiler from optimising the calls away, the two classic benchmark bugs.

## One machine's results

Running `go test -bench . -count 3` on one desktop machine gave (trimmed):

```
BenchmarkConcrete-12        263.7 ns/op
BenchmarkConcrete-12        236.9 ns/op
BenchmarkConcrete-12        236.7 ns/op
BenchmarkGeneric-12         236.5 ns/op
BenchmarkGeneric-12         236.7 ns/op
BenchmarkGeneric-12         236.3 ns/op
BenchmarkInterface-12        1408 ns/op
BenchmarkInterface-12        2350 ns/op
BenchmarkInterface-12        1887 ns/op
BenchmarkGenericMethod-12    1200 ns/op
BenchmarkGenericMethod-12    1200 ns/op
BenchmarkGenericMethod-12    1200 ns/op
```

Your numbers will differ. The *shape* is what matters, and it matches the model:

- **Generic operators = concrete code.** `Sum[int]` and `SumInts` are indistinguishable. The `int` shape compiles `+=` to the same add instruction.
- **Method calls are the cost.** Both method-calling versions are around five times slower than a direct sum, because each `Value()` is an indirect call that can't be inlined. The generic one is a bit faster and steadier here, since its data is a flat `[]Size` rather than a slice of two-word interfaces pointing elsewhere.
- **Noise is real.** Look at the first `Concrete` run, and the spread of the `Interface` runs. One run proves nothing; use `-count` and a tool like `benchstat` to compare.
- **All reported 0 allocations** with `-benchmem`. Converting a `Size` to `Valuer` happened once in setup, not in the loop.

## Allocations are the usual culprit

In real code, the biggest difference between `any`-based and generic code is rarely call overhead. It's **boxing**: every non-pointer value converted to an interface may be copied to the heap. A `[]any` of a thousand `int`s can mean a thousand allocations, and then work for the garbage collector. Generic code over `[]T` stores values directly.

`testing.AllocsPerRun(runs, f)` measures that deterministically, which makes it perfect for a test that locks in "this doesn't allocate":

```go
if a := testing.AllocsPerRun(20, func() { Largest(ints) }); a != 0 {
	t.Errorf("Largest allocated %v times per call, want 0", a)
}
```

## Your turn

Someone wrote two Stash helpers by converting everything to `any` first. They give the right answers, but the tests show they allocate on every element. Rewrite both so they allocate **nothing**:

- `Largest(xs)` returns the largest value and `true`, or the zero value and `false` for an empty slice.
- `IndexOf(xs, v)` returns the index of the first element equal to `v`, or -1.

Both type parameters already have the constraints you need (`cmp.Ordered` gives you `>`, `comparable` gives you `==`), so work on `T` values directly.
