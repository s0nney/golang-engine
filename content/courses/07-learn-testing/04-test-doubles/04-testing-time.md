---
title: Testing Time
quiz:
  - question: |
      A test for "invoices become overdue after 30 days" calls
      `time.Sleep(30 * 24 * time.Hour)`. What's the better approach?
    options:
      - text: Sleep for 30 seconds and scale the business rule down in tests
      - text: Inject a clock and have the test advance a fake clock by 30 days
        correct: true
      - text: Run the test with `-timeout 800h`
      - text: Skip it with `testing.Short()`
    explanation: |
      Code that asks an injected clock for the time can be tested at any
      moment in history, instantly. The fake clock's `Advance` moves time
      forward without anyone waiting.
  - question: |
      What does this print?

      ```go
      a := time.Date(2026, time.March, 1, 9, 0, 0, 0, time.UTC)
      b := a.In(time.FixedZone("CET", 60*60))
      fmt.Println(a == b, a.Equal(b))
      ```
    options:
      - text: '`true true`'
      - text: '`false false`'
      - text: '`false true`'
        correct: true
      - text: '`true false`'
    explanation: |
      Both values are the same instant, so `Equal` is `true`. But `==`
      compares the whole struct, including the location, and `b` carries
      a different one. Always compare times with `Equal` in tests.
---

Ledgerly has plenty of time-dependent rules: an invoice is overdue after 30 days, recurring payments run on the 1st, a statement covers "last month". Code that calls `time.Now()` directly behaves differently depending on when you run it. That leads to tests that pass all month and fail on the 31st, or only in January, or only in a time zone the CI server doesn't use.

The fix is the same as for any hidden input: make it an explicit one.

## Level 1: pass the time in

If a function needs "now" once, take it as a parameter:

```go
func IsOverdue(inv Invoice, now time.Time) bool {
	return now.After(inv.Issued.AddDate(0, 0, 30))
}
```

The test builds exact times with `time.Date`, always with an explicit location:

```go
issued := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
inv := Invoice{Issued: issued}
if IsOverdue(inv, issued.AddDate(0, 0, 30)) {
	t.Error("overdue on day 30, want not yet")
}
if !IsOverdue(inv, issued.AddDate(0, 0, 31)) {
	t.Error("not overdue on day 31, want overdue")
}
```

This is the best option whenever it fits: nothing to fake at all.

## Level 2: inject a clock

Long-lived objects ask for the time repeatedly. Give them a clock, either as a `func() time.Time` field (production sets it to `time.Now`) or as a small interface:

```go
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }
```

The test double is a **fake clock** that only moves when told to:

```go
type FakeClock struct {
	now time.Time
}

func NewFakeClock(t time.Time) *FakeClock    { return &FakeClock{now: t} }
func (c *FakeClock) Now() time.Time          { return c.now }
func (c *FakeClock) Advance(d time.Duration) { c.now = c.now.Add(d) }
```

Now a test can walk through a month in microseconds:

```go
clock := NewFakeClock(time.Date(2026, time.March, 1, 9, 0, 0, 0, time.UTC))
r := NewRecurring(clock, "rent", 120000) // due on the 1st of each month

assertDue(t, r, true)
clock.Advance(24 * time.Hour)
assertDue(t, r, false)
clock.Advance(30 * 24 * time.Hour) // April 1st
assertDue(t, r, true)
```

The pointer receivers matter: the test and the code under test must share *one* clock, so that `Advance` in the test is visible to the code. If the code reads the clock from other goroutines, protect `now` with a mutex.

## Level 3: timers, sleeps and synctest

A fake `Now` doesn't help with code that *waits*: `time.Sleep`, `time.After`, tickers, context timeouts. You could fake those too, behind an interface with `After(d) <-chan time.Time`, but that gets complicated fast.

You already have the right tool from the concurrency course: [`testing/synctest`](/courses/learn-concurrency/testing-concurrent-code/synctest). Inside `synctest.Test`, the bubble's fake clock jumps forward whenever every goroutine is blocked, so a retry helper that makes three attempts, sleeping 10 seconds between them, is tested instantly with no changes to its code, and `time.Since(start)` comes out as *exactly* 20 seconds. Revisit that lesson for the details; here's where each tool fits:

- Pure logic that needs "now": pass a `time.Time`.
- Objects that check the time repeatedly: inject a `Clock` and use a fake.
- Code that sleeps, waits, or uses timers and timeouts: `synctest`.

## Time zones and monotonic clocks

Two more traps:

- `time.Now()` uses the machine's local zone. Your laptop and CI may disagree about what day it is. Build test times with `time.UTC` (or a fixed `time.FixedZone`) and decide explicitly which zone business rules like "the 1st of the month" use.
- Compare times with `t1.Equal(t2)`, not `==`. Values from `time.Now()` carry a monotonic clock reading and a location pointer, so two `time.Time` values for the same instant can be `!=`.
