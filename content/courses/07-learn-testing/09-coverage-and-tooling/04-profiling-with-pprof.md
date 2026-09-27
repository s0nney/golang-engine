---
title: Profiling with pprof
quiz:
  - question: |
      In `go tool pprof -top` output, a function has `flat` 0.11s and
      `cum` 2.48s. What does that mean?
    options:
      - text: It ran 0.11s the first time and 2.48s the second
      - text: 0.11s was spent in the function's own code, and 2.48s in it plus everything it called
        correct: true
      - text: It was called 11 times out of 248
      - text: The profile is corrupted
    explanation: |
      `flat` counts samples where the function itself was executing. `cum`
      (cumulative) also counts samples where it was further up the call
      stack. A big `cum` with a small `flat` means the cost is in its
      callees.
  - question: Which command writes a CPU profile while running Ledgerly's import benchmark?
    options:
      - text: '`go tool pprof -bench ImportCSV`'
      - text: '`go test -run ''^$'' -bench ImportCSV -cpuprofile cpu.out ./ledgerly`'
        correct: true
      - text: '`go build -cpuprofile cpu.out`'
      - text: '`go test -cover -cpu ImportCSV`'
    explanation: |
      `go test` can write CPU (`-cpuprofile`), heap (`-memprofile`),
      blocking (`-blockprofile`) and mutex (`-mutexprofile`) profiles for
      whatever tests or benchmarks it runs. `go tool pprof` then reads the
      file.
---

Benchmarks tell you *how fast* something is. Profiles tell you *where the time goes*. The chapter on benchmarks warned against optimising the wrong thing, and a profile is how you find the right thing.

## Getting a profile from a benchmark

A benchmark is a ready-made workload, so it's the easiest thing to profile:

```text
$ go test -run '^$' -bench 'ImportCSV/1000' -cpuprofile cpu.out -memprofile mem.out ./ledgerly
BenchmarkImportCSV/1000-12    23203    102543 ns/op    233.17 MB/s
PASS
```

That writes two files: `cpu.out`, a **CPU profile** (every 10 milliseconds, the runtime records which function is running) and `mem.out`, a **heap profile** (a sample of allocations and where they happened). The test binary is saved next to them as `ledgerly.test`, because pprof needs it to turn addresses into function names.

## Reading it: top

```text
$ go tool pprof -top cpu.out
Duration: 2.38s, Total samples = 2920ms (122.66%)
      flat  flat%   sum%        cum   cum%
     440ms 15.07% 15.07%      440ms 15.07%  indexbytebody
     410ms 14.04% 29.11%     1790ms 61.30%  encoding/csv.(*Reader).readRecord
     190ms  6.51% 35.62%      220ms  7.53%  runtime.mallocgcSmallScanNoHeaderSC5
     130ms  4.45% 40.07%      220ms  7.53%  bufio.(*Reader).ReadSlice
     120ms  4.11% 44.18%      120ms  4.11%  internal/strconv.ParseUint
     110ms  3.77% 47.95%     2480ms 84.93%  ledgerly.ImportCSV
     ...
      60ms  2.05% 72.95%      400ms 13.70%  ledgerly.ParseAmount
```

- **flat**: time spent *in* that function's own code.
- **cum**: time spent in it *and everything it called*.

Sort by cumulative time with `-top -cum` to see the big picture. Here, of the 2.48 seconds inside `ImportCSV`, 1.79 went to `encoding/csv` reading records and 0.40 to `ParseAmount`. The `runtime.mallocgc...` lines are allocation. So if you wanted a faster import, rewriting `ParseAmount` would save at most about a sixth. The CSV reading and the allocations are where the time goes.

## Zooming in: list

`-list` shows a function's source with the time spent on each line:

```text
$ go tool pprof -list 'ledgerly.ParseAmount' cpu.out
      60ms      400ms (flat, cum) 13.70% of Total
         .       10ms     30:	digits, neg := strings.CutPrefix(s, "-")
         .      200ms     31:	whole, frac, hasDot := strings.Cut(digits, ".")
      10ms       10ms     32:	if whole == "" || (hasDot && (len(frac) < 1 || len(frac) > 2)) {
      10ms       40ms     38:	d, err1 := strconv.ParseUint(whole, 10, 63)
         .      100ms     39:	c, err2 := strconv.ParseUint(frac, 10, 63)
```

Half of `ParseAmount` is finding the `.`, which is useful to know before you guess.

## Memory profiles

The same commands work on `mem.out`. By default pprof shows memory **in use** at the time of the profile. For finding allocation-heavy code, look at everything allocated instead:

```text
$ go tool pprof -sample_index=alloc_space -top mem.out
      flat  flat%   sum%        cum   cum%
 1979.08MB 54.01% 54.01%  3658.97MB 99.85%  ledgerly.ImportCSV
 1592.56MB 43.46% 97.47%  1592.56MB 43.46%  encoding/csv.(*Reader).readRecord
   87.33MB  2.38% 99.85%    87.33MB  2.38%  bufio.NewReaderSize (inline)
```

(These are totals over all 23,203 benchmark iterations.) `ImportCSV`'s own allocations come from growing the `txns` slice. `readRecord` allocates a fresh `[]string` for every row. `encoding/csv` has an option for exactly that: set `ReuseRecord = true` on the reader and it reuses the slice between rows, which is safe here because `ImportCSV` copies the fields it needs. Use `alloc_objects` instead of `alloc_space` to count allocations rather than bytes.

## The web UI

```text
go tool pprof -http=localhost:8080 cpu.out
```

This opens an interactive view in your browser: a call graph, a **flame graph** (each bar is a function, its width is its share of the time, and callers sit above callees), and source listings. For anything bigger than a toy, it's much easier to explore than the text output.

## Profiling a real program

Benchmarks are the easy case. For a long-running program you can:

- call `pprof.StartCPUProfile(f)` and `pprof.StopCPUProfile()` from `runtime/pprof` around the work, or `pprof.WriteHeapProfile(f)` for memory;
- in a server, import `net/http/pprof` to expose live profiles under `/debug/pprof/`, then point `go tool pprof` at the URL.

## A workflow

1. Write a benchmark for the slow operation, with realistic input.
2. Profile it. Find the top few entries by `cum`.
3. Change *one* thing in the hot spot.
4. Re-run the benchmark with `-count 10` and compare with `benchstat`.
5. Keep the change only if the numbers say it helped, and your tests still pass.

That loop, measure, change, measure again, is the whole of performance work. The tools in this chapter and the last make each step a single command.
