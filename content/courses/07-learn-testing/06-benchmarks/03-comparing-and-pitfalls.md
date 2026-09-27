---
title: Comparing Results and Benchmark Pitfalls
quiz:
  - question: |
      You run a benchmark once before and once after a change. It went from
      `412 ns/op` to `398 ns/op`. What can you conclude?
    options:
      - text: The change made it 3.4% faster
      - text: Nothing yet; one run each can't separate a 3% change from normal noise
        correct: true
      - text: The change made it slower, because the first run warms the cache
      - text: The benchmark is broken
    explanation: |
      Benchmark timings vary from run to run by a few percent. Run each
      version many times (`-count=10`) and let `benchstat` decide whether
      the difference is statistically significant.
  - question: |
      What's wrong with this benchmark?

      ```go
      func BenchmarkSortTransactions(b *testing.B) {
          txns := loadSample(b)
          for b.Loop() {
              SortByDate(txns)
          }
      }
      ```
    options:
      - text: Nothing
      - text: '`SortByDate` sorts in place, so after the first iteration it''s always sorting already-sorted data'
        correct: true
      - text: '`loadSample` is timed'
      - text: '`b.Loop` can''t be used with slices'
    explanation: |
      Only the first iteration sees shuffled input. Every later one
      measures the (often much faster) already-sorted case. Copy fresh
      input each iteration, or measure the copy separately and subtract it.
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"flag"
    	"fmt"
    	"slices"
    	"testing"
    )

    // BenchmarkSortByDate should measure sortByDate on unsorted input, every
    // iteration. Fix it.
    func BenchmarkSortByDate(b *testing.B) {
    	for b.Loop() {
    		sortByDate(sample)
    	}
    }

    // ---- Ledgerly code ----

    // Txn is a transaction on a given day of the month.
    type Txn struct {
    	Day   int
    	Cents int64
    }

    // sample is shared by every benchmark in the package: 500 transactions
    // in no particular order. Benchmarks must not modify it.
    var sample = makeSample(500)

    func makeSample(n int) []Txn {
    	txns := make([]Txn, n)
    	for i := range txns {
    		txns[i] = Txn{Day: i * 7919 % n, Cents: int64(i)}
    	}
    	return txns
    }

    func byDay(a, b Txn) int { return cmp.Compare(a.Day, b.Day) }

    // sortByDate sorts txns in place by day. It's a variable so that tests
    // can wrap it with a spy.
    var sortByDate = func(txns []Txn) {
    	slices.SortFunc(txns, byDay)
    }

    func main() {
    	testing.Init()
    	flag.Set("test.benchtime", "2000x") // keep Run quick
    	r := testing.Benchmark(BenchmarkSortByDate)
    	fmt.Println("BenchmarkSortByDate:", r)
    	fmt.Println("sample still unsorted:", !slices.IsSortedFunc(sample, byDay))
    }
  solution: |
    package main

    import (
    	"cmp"
    	"flag"
    	"fmt"
    	"slices"
    	"testing"
    )

    // BenchmarkSortByDate measures sortByDate on unsorted input, every
    // iteration. Cloning is part of the measurement, which is fine for
    // comparing sort implementations, since they all pay for it.
    func BenchmarkSortByDate(b *testing.B) {
    	for b.Loop() {
    		txns := slices.Clone(sample)
    		sortByDate(txns)
    	}
    }

    // ---- Ledgerly code ----

    // Txn is a transaction on a given day of the month.
    type Txn struct {
    	Day   int
    	Cents int64
    }

    // sample is shared by every benchmark in the package: 500 transactions
    // in no particular order. Benchmarks must not modify it.
    var sample = makeSample(500)

    func makeSample(n int) []Txn {
    	txns := make([]Txn, n)
    	for i := range txns {
    		txns[i] = Txn{Day: i * 7919 % n, Cents: int64(i)}
    	}
    	return txns
    }

    func byDay(a, b Txn) int { return cmp.Compare(a.Day, b.Day) }

    // sortByDate sorts txns in place by day. It's a variable so that tests
    // can wrap it with a spy.
    var sortByDate = func(txns []Txn) {
    	slices.SortFunc(txns, byDay)
    }

    func main() {
    	testing.Init()
    	flag.Set("test.benchtime", "2000x") // keep Run quick
    	r := testing.Benchmark(BenchmarkSortByDate)
    	fmt.Println("BenchmarkSortByDate:", r)
    	fmt.Println("sample still unsorted:", !slices.IsSortedFunc(sample, byDay))
    }
  tests: |
    package main

    import (
    	"flag"
    	"slices"
    	"testing"
    )

    func runBenchmark(t *testing.T, iterations string) {
    	t.Helper()
    	if err := flag.Set("test.benchtime", iterations); err != nil {
    		t.Fatal(err)
    	}
    	testing.Benchmark(BenchmarkSortByDate)
    }

    func TestSampleUntouched(t *testing.T) {
    	orig := makeSample(500)
    	if !slices.Equal(sample, orig) {
    		t.Fatal("sample was already modified before the benchmark ran; don't change makeSample or sample")
    	}
    	runBenchmark(t, "50x")
    	if !slices.Equal(sample, orig) {
    		t.Error("after running BenchmarkSortByDate, sample has changed: the benchmark must not sort the shared sample in place (other benchmarks use it too)")
    	}
    	t.Cleanup(func() { sample = orig })
    }

    func TestEveryIterationSortsUnsortedInput(t *testing.T) {
    	sortFn := sortByDate
    	t.Cleanup(func() { sortByDate = sortFn })
    	calls, sortedInputs, wrongLen := 0, 0, 0
    	sortByDate = func(txns []Txn) {
    		calls++
    		if slices.IsSortedFunc(txns, byDay) {
    			sortedInputs++
    		}
    		if len(txns) != len(sample) {
    			wrongLen++
    		}
    		sortFn(txns)
    	}
    	sample = makeSample(500)
    	runBenchmark(t, "100x")
    	if calls < 100 {
    		t.Fatalf("with -benchtime=100x, sortByDate was called %d times, want at least 100: call it inside the b.Loop() loop", calls)
    	}
    	if sortedInputs > 0 {
    		t.Errorf("%d of %d calls to sortByDate got already-sorted input: every iteration needs a fresh unsorted copy of sample", sortedInputs, calls)
    	}
    	if wrongLen > 0 {
    		t.Errorf("%d of %d calls to sortByDate got a slice whose length isn't len(sample) = %d", wrongLen, calls, len(sample))
    	}
    }
---

Benchmarks are experiments, and experiments can be done badly. This lesson is about getting numbers you can trust.

## Noise, and why one run isn't enough

Run the same benchmark twice and you'll get different numbers. CPU frequency scaling, other processes, the garbage collector, even where the binary happened to be laid out in memory all shift results by a few percent. So a single "before" and "after" can't tell a 3% improvement from luck.

The fix is statistics:

```text
$ git stash
$ go test -run '^$' -bench ImportCSV -count 10 ./ledgerly > old.txt
$ git stash pop
$ go test -run '^$' -bench ImportCSV -count 10 ./ledgerly > new.txt
$ benchstat old.txt new.txt
```

`-count 10` runs each benchmark ten times. **benchstat** (install it with `go install golang.org/x/perf/cmd/benchstat@latest`) reads both files and prints something like:

```text
                 │   old.txt   │              new.txt               │
                 │   sec/op    │   sec/op     vs base               │
ImportCSV/10-12    2.412µ ± 1%   2.198µ ± 2%   -8.87% (p=0.000 n=10)
ImportCSV/1000-12  101.4µ ± 1%   100.9µ ± 3%        ~ (p=0.436 n=10)
```

For each benchmark you get the median, the variation (`± 1%`), the change, and a **p-value**. A `~` means "no statistically significant difference": here the change helped small files and did nothing measurable for big ones. That's the honest answer that two single runs would never have given you.

Tips for less noise:

- Close the browser and the video call. Plug the laptop in.
- Don't compare numbers from different machines, or from CI runners, which share hardware with other jobs.
- Run old and new close together in time, on the same machine.

## Pitfall 1: the compiler deleted your work

In an old-style `b.N` loop, if the result of a pure function is never used, the compiler is allowed to delete the call, and you end up timing an empty loop:

```go
func BenchmarkRound(b *testing.B) {
	for i := 0; i < b.N; i++ {
		roundToDollar(123456) // result unused, constant input
	}
}
```

A result under a nanosecond per operation is the tell-tale sign: that's a single CPU cycle. Old code worked around it by assigning to a package-level `sink` variable. `b.Loop` handles the result side for you, since results and arguments of calls inside the loop body are kept alive. Constant inputs can still be folded in at compile time, so feed the benchmark a variable declared before the loop.

## Pitfall 2: timing the setup

With `b.N` loops, expensive setup before the loop runs on every call to the benchmark function and is timed unless you call `b.ResetTimer()` after it. `b.Loop` resets the timer itself.

Setup that has to happen *inside* the loop is harder. You can bracket it with `b.StopTimer()` and `b.StartTimer()`, but those calls have overhead of their own and distort very short operations. Often it's cleaner to measure "setup + work" and "setup alone" as two benchmarks and compare.

## Pitfall 3: state that changes between iterations

The quiz shows the classic: an in-place sort benchmark that only sorts once. Watch for anything the loop body mutates: sorting, appending to a shared slice (which grows forever), filling a cache (every later iteration is a cache hit), consuming a reader (every later iteration reads nothing).

```go
for b.Loop() {
	txns := slices.Clone(sample) // fresh unsorted input
	SortByDate(txns)
}
```

The clone is now part of the measurement, which is often fine for comparing two sort implementations, since both pay the same cost.

## Your turn: fix a lying benchmark

`BenchmarkSortByDate` in the editor has the bug from the quiz. It sorts the shared `sample` in place, so only the first iteration sorts anything, and every other benchmark in the package now gets sorted data too.

Fix the benchmark so that **every iteration sorts a fresh, unsorted copy** and `sample` is never modified. Keep the `for b.Loop()` loop and the call to `sortByDate`.

The grader wraps `sortByDate` in a spy that records whether each call receives already-sorted input, and checks `sample` after the benchmark runs. Cloning once *before* the loop isn't enough: the second iteration would sort the already-sorted clone. **Run** prints the benchmark result and whether `sample` survived. Watch the ns/op jump once the benchmark measures real work.

## Pitfall 4: unrealistic inputs

A benchmark with 10 transactions tells you nothing about a 10-million-row import. Tiny inputs fit in the CPU cache, and algorithmic problems (an accidental O(n²)) don't show. Benchmark with sizes like the real ones, and several sizes where you can, as with the sub-benchmarks in the first lesson.

## Pitfall 5: optimising what doesn't matter

A benchmark tells you how fast one function is, not whether it matters. Making `Cents.String` nine times faster is pointless if the import spends 95% of its time reading the disk. Profile the real workload first (the Coverage and Tooling chapter covers `pprof`), find the hot spot, *then* benchmark and optimise that.

## Custom metrics

`b.ReportMetric(value, unit)` adds your own column, which is great for domain-specific numbers:

```go
b.ReportMetric(float64(rows)*float64(b.N)/b.Elapsed().Seconds(), "rows/s")
```

After `b.Loop` returns `false`, `b.N` holds the total number of iterations, so you can use it to compute per-iteration metrics like this one.
