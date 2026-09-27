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

## Pitfall 4: unrealistic inputs

A benchmark with 10 transactions tells you nothing about a 10-million-row import. Tiny inputs fit in the CPU cache, and algorithmic problems (an accidental O(n²)) don't show. Benchmark with sizes like the real ones, and several sizes where you can, as with the sub-benchmarks in the first lesson.

## Pitfall 5: optimising what doesn't matter

A benchmark tells you how fast one function is, not whether it matters. Making `Cents.String` nine times faster is pointless if the import spends 95% of its time reading the disk. Profile the real workload first (chapter 9 covers `pprof`), find the hot spot, *then* benchmark and optimise that.

## Custom metrics

`b.ReportMetric(value, unit)` adds your own column, which is great for domain-specific numbers:

```go
b.ReportMetric(float64(rows)*float64(b.N)/b.Elapsed().Seconds(), "rows/s")
```

After `b.Loop` returns `false`, `b.N` holds the total number of iterations, so you can use it to compute per-iteration metrics like this one.
