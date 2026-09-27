---
title: Benchmarking with testing.B
quiz:
  - question: |
      In Go 1.24 and later, what's the idiomatic shape of a benchmark loop?
    options:
      - text: '`for i := 0; i < b.N; i++ { ... }`'
      - text: '`for b.Loop() { ... }`'
        correct: true
      - text: '`for range b.N { b.ResetTimer() }`'
      - text: '`b.Run(func() { ... })`'
    explanation: |
      `b.Loop()` runs the body as many times as needed, excludes the setup before
      the loop from the timing automatically, and keeps the loop body's results
      alive so the compiler can't optimise the work away.
  - question: |
      A benchmark reports `2000 ns/op` for 1,000 items and `8000 ns/op` for 2,000
      items. Which complexity does that suggest?
    options:
      - text: O(n)
      - text: O(n log n)
      - text: O(n²)
        correct: true
    explanation: |
      Doubling the input quadrupled the time, which is the signature of
      quadratic growth. For O(n) it would roughly double; for O(n log n) it would
      be a little over double.
---

Big O tells you how an algorithm *should* scale. A **benchmark** tells you how
it *actually* performs on real hardware. You want both: analysis to choose an
approach, and measurement to check your reasoning and catch constant factors
that Big O deliberately ignores.

## Writing a benchmark

Benchmarks live in `_test.go` files next to your code, just like tests. A
benchmark function starts with `Benchmark` and takes a `*testing.B`. Since Go
1.24, the body uses `b.Loop()`:

```go
// dupes.go
package dupes

func HasDuplicatePairs(counts []int) bool {
	for i := range counts {
		for j := i + 1; j < len(counts); j++ {
			if counts[i] == counts[j] {
				return true
			}
		}
	}
	return false
}

func HasDuplicateSet(counts []int) bool {
	seen := make(map[int]struct{}, len(counts))
	for _, c := range counts {
		if _, ok := seen[c]; ok {
			return true
		}
		seen[c] = struct{}{}
	}
	return false
}
```

```go
// dupes_test.go
package dupes

import (
	"fmt"
	"testing"
)

func uniqueCounts(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i * 7
	}
	return s
}

func BenchmarkDuplicates(b *testing.B) {
	for _, n := range []int{1_000, 2_000, 4_000} {
		counts := uniqueCounts(n) // setup: not timed
		b.Run(fmt.Sprintf("pairs/n=%d", n), func(b *testing.B) {
			for b.Loop() {
				HasDuplicatePairs(counts)
			}
		})
		b.Run(fmt.Sprintf("set/n=%d", n), func(b *testing.B) {
			for b.Loop() {
				HasDuplicateSet(counts)
			}
		})
	}
}
```

We use all-unique counts on purpose: that's the **worst case**, where neither
function can return early. `b.Run` creates sub-benchmarks so we can compare
sizes side by side.

## Running it

```
go test -bench=. -benchmem
```

The output looks something like this (your numbers will differ). The `-12`
suffix is the number of CPUs Go used:

```
BenchmarkDuplicates/pairs/n=1000-12     10000     109139 ns/op        0 B/op    0 allocs/op
BenchmarkDuplicates/set/n=1000-12       77811      15834 ns/op    36944 B/op    5 allocs/op
BenchmarkDuplicates/pairs/n=2000-12      2758     416500 ns/op        0 B/op    0 allocs/op
BenchmarkDuplicates/set/n=2000-12       38622      30588 ns/op    73888 B/op    9 allocs/op
BenchmarkDuplicates/pairs/n=4000-12       708    1661048 ns/op        0 B/op    0 allocs/op
BenchmarkDuplicates/set/n=4000-12       19726      60844 ns/op   147776 B/op   17 allocs/op
```

Read the `ns/op` column as `n` doubles. The pairs version goes roughly ×4 each
time: that's O(n²). The set version roughly doubles: O(n). And `-benchmem`
shows the trade-off from the last lesson: pairs allocates nothing, the set
version allocates memory proportional to `n`.

## Why `b.Loop()`?

Older code writes `for i := 0; i < b.N; i++`. That still works, but `b.Loop()`
is better:

- **Setup is excluded automatically.** The timer resets the first time
  `b.Loop()` is called, so you don't need `b.ResetTimer()`.
- **The work can't be optimised away.** With `b.N` loops, the compiler may
  notice you ignore `HasDuplicateSet`'s result and delete the call, giving a
  fantastically fast and totally fake benchmark. Inside `for b.Loop() { ... }`,
  results and arguments are kept alive, so the call really happens.
- **The function body runs once.** With `b.N`, the benchmark function is
  called several times with growing `b.N`, re-running your setup each time.
  With `b.Loop()`, the ramp-up happens inside the loop.

## Benchmarking tips

- Benchmark the **worst case** input if that's what you'll face in production.
- Close other heavy programs. Noise is real.
- Run with `-count=10` and compare runs with the `benchstat` tool before
  trusting a small difference.
- A benchmark confirms Big O, it doesn't replace it. At `n = 4,000` the O(n²)
  version is "only" 1.7 ms. Big O is what tells you it'll take minutes at a
  million.

## Further reading

- [Learn Go with Tests: Iteration](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/iteration), which writes a first benchmark with `b.Loop()` and `-benchmem`.
