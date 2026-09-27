---
title: The Race Detector
quiz:
  - question: '`go test -race ./...` passes with no warnings. What can you conclude?'
    options:
      - text: The code has no data races
      - text: No data race happened in the code paths the tests actually executed
        correct: true
      - text: The code has no race conditions of any kind
      - text: The code has no deadlocks
    explanation: |
      The race detector watches memory accesses as they happen at run time.
      It has no false positives, but it can only report races in code
      that actually ran, with goroutines that actually overlapped. Untested
      paths can still be racy, and logic-level race conditions are
      invisible to it.
  - question: Why don't teams simply build their production binaries with `-race`?
    options:
      - text: It's only available on Windows
      - text: It uses several times more memory and runs much slower, so it's normally reserved for tests, CI and staging
        correct: true
      - text: It changes the program's output
      - text: It disables goroutines
    explanation: |
      The instrumentation typically costs 5-10x memory and 2-20x CPU time.
      That's fine for tests, and some teams run a canary instance with it,
      but it's too heavy for normal production traffic.
---

Data races are nasty to find by reading code and nearly impossible to find by testing normally: the racy test usually passes. Go ships a tool that finds them for you: the **race detector**. It's built into the `go` command, and turning it on is one flag.

## Turning it on

```text
go test -race ./...
go run -race .
go build -race
```

`-race` compiles your program with extra instrumentation around every memory access and synchronization operation. While the program runs, the detector tracks which goroutine touched which memory and what synchronization happened in between. When two accesses conflict without any synchronization ordering them, it prints a report.

## Reading a report

Here's a package whose test passes normally:

```go
type Stats struct {
	delivered int
}

func (s *Stats) Delivered() { s.delivered++ }

func DeliverAll(n int) int {
	var s Stats
	var wg sync.WaitGroup
	for range n {
		wg.Go(s.Delivered)
	}
	wg.Wait()
	return s.delivered
}
```

```text
$ go test .
ok  	dispatch	0.001s
```

With `-race` (trimmed a little):

```text
$ go test -race .
==================
WARNING: DATA RACE
Read at 0x00c0000182f8 by goroutine 15:
  dispatch.(*Stats).Delivered()
      /home/you/dispatch/stats.go:9 +0x30

Previous write at 0x00c0000182f8 by goroutine 11:
  dispatch.(*Stats).Delivered()
      /home/you/dispatch/stats.go:9 +0x44

Goroutine 15 (running) created at:
  sync.(*WaitGroup).Go()
      /usr/lib/go/src/sync/waitgroup.go:238 +0x72
  dispatch.TestDeliverAll()
      /home/you/dispatch/stats_test.go:6 +0x2e
==================
--- FAIL: TestDeliverAll (0.00s)
    stats_test.go:7: DeliverAll(10) = 9, want 10
    testing.go:1865: race detected during execution of test
FAIL
```

A report has three parts:

1. **The two conflicting accesses**: a read and a previous write of the same address, each with a stack trace. Here both are line 9, `s.delivered++`, in different goroutines.
2. **Where each goroutine was created**, which shows how the two got running concurrently.
3. **The verdict**: under `go test` the test fails with "race detected during execution of test". A program built with `-race` exits with status 66 when the run ends, if any race was found.

Notice that this run also *lost an update*: 9 instead of 10. That time the bug was visible. Often it isn't, which is exactly why you need the tool.

## What it can and can't do

- **No false positives.** Every report is a real data race. Fix it. Don't argue with it.
- **Only what runs.** It checks the executions it sees. A race in a code path your tests never hit, or between goroutines that never happened to overlap, goes unreported. Tests that exercise real concurrency (many goroutines, `-count=N` to repeat) give it more to see.
- **Data races only.** Race conditions (check-then-act) and deadlocks are invisible to it.
- **It's expensive.** Expect memory use to grow 5-10x and run time 2-20x. That's why it's for tests, CI and staging, not normal production builds.

## Make it a habit

- Run `go test -race ./...` in CI on every change. Race bugs creep in quietly, and CI is where you'll catch them.
- Write tests that actually run things concurrently: call the method from 50 goroutines at once, not just once.
- When a report looks "harmless" (a stats counter that's only a bit off), fix it anyway. A data race is undefined behaviour, and harmless-looking races have a way of corrupting something important later.

(When you press **Submit**, this site runs `go test -v` without `-race`, so the exercise tests catch the *effects* of races instead. On your own machine, always add `-race`.)

## Further reading

- [Data Race Detector](https://go.dev/doc/articles/race_detector), the official guide
