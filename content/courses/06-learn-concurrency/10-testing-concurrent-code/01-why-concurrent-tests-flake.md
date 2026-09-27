---
title: Why Concurrent Tests Flake
quiz:
  - question: |
      What's wrong with this test?

      ```go
      func TestDispatcherStarts(t *testing.T) {
      	d := NewDispatcher()
      	go d.Run()
      	time.Sleep(10 * time.Millisecond) // give Run time to start
      	if !d.Running() {
      		t.Fatal("dispatcher not running")
      	}
      }
      ```
    options:
      - text: Nothing, 10ms is plenty of time for a goroutine to start
      - text: It relies on timing, so on a loaded CI machine `Run` may not have started after 10ms, and the test fails at random
        correct: true
      - text: '`go d.Run()` must be called with a WaitGroup'
      - text: '`time.Sleep` is not allowed in tests'
    explanation: |
      A sleep is a guess about how long something takes. Nothing guarantees
      the goroutine has even been scheduled after 10ms. Raise it to 100ms
      and the test gets slower *and* can still fail. Wait for the actual
      event instead.
  - question: A concurrent test fails once in every 200 runs. What's the most useful first step?
    options:
      - text: Add a retry to the test
      - text: 'Run it with `go test -race -count=1000 -run TestName` to reproduce it and check for data races'
        correct: true
      - text: Increase all the sleeps in the test
      - text: Mark it as skipped
    explanation: |
      `-count` repeats the test to reproduce the failure, and `-race`
      catches data races that might be behind it. A rare failure is a real
      bug, either in the test or, worse, in the code.
---

You now know how to write concurrent code. Testing it is the next challenge, and it's where many teams give up and live with **flaky tests**: tests that usually pass and occasionally fail for no visible reason. A flaky test is worse than no test. People learn to ignore red builds, and the real failures hide among the random ones.

## The two sources of flakiness

**Scheduling.** The order in which goroutines run isn't fixed. Two goroutines racing to a shared variable, or a test checking a result before the goroutine that produces it has run, will behave differently from run to run and from machine to machine. Your laptop and a busy CI runner schedule very differently.

**Real time.** Code with timeouts, tickers, retries and backoff is naturally tested by waiting, and waiting in tests is both slow and unreliable:

```go
func TestPickupTimeout(t *testing.T) {
	go watchPickup(order, 100*time.Millisecond) // the real timeout is 20 minutes!
	time.Sleep(150 * time.Millisecond)          // hope it has fired by now
	// check that the order was reassigned...
}
```

This test is slow (every timeout costs real time), needs the production timeout shrunk just for testing, and still fails whenever the machine is too busy to run the goroutine within 50ms.

## Rules that help

**Never sleep to synchronize.** A sleep is a guess. Wait on the actual event: receive from a channel the goroutine closes when it's done, `wg.Wait()` for it to finish, or read a result it sends you.

```go
done := make(chan struct{})
go func() {
	defer close(done)
	d.Run(ctx)
}()
cancel()
<-done // Run has definitely returned
```

**Bound every wait.** A test that deadlocks hangs until `go test`'s 10-minute timeout. Where a wait might never complete, use a `select` with a timeout so you get a clear failure message instead.

**Use `t.Context()`** for the code under test. It's cancelled automatically just before the test's cleanup functions run, so background goroutines are told to stop.

**Run with `-race`, and repeat.** `go test -race -count=100 -run TestName` reproduces rare failures and flags data races behind them. `-shuffle=on` randomizes test order too.

**Make time injectable, or fake it.** The traditional fix for time-dependent code is to pass in a fake clock interface. That works, but it means threading a clock through all your code, and it doesn't help with library code that calls `time.Now` or `time.After` directly.

## A better way

Since Go 1.25 the standard library has `testing/synctest`, which fixes both problems at once:

- It runs your test in a **bubble** with a **fake clock**. Time only moves forward when every goroutine in the bubble is blocked, so a 20-minute timeout takes zero real time, and it fires at *exactly* 20 minutes, every time.
- `synctest.Wait` lets a test wait until every other goroutine in the bubble has finished whatever it can do, with no guessing.

Your real code keeps calling `time.After`, `time.NewTicker` and `context.WithTimeout` as normal. No clock interface needed. The next two lessons show how, and you've actually been relying on it already: most of this course's exercises are graded by `synctest` tests.

## Further reading

- [Go blog: Testing concurrent code with testing/synctest](https://go.dev/blog/synctest)
