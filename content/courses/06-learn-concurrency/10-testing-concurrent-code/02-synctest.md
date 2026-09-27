---
title: Fake Time with testing/synctest
quiz:
  - question: Inside a `synctest.Test` bubble, the code under test calls `time.After(20 * time.Minute)` and nothing else is going on. How long does the test take in real time?
    options:
      - text: 20 minutes
      - text: It hangs forever, since fake time never moves
      - text: Practically no time; once every goroutine in the bubble is blocked, the fake clock jumps straight to the next timer
        correct: true
      - text: It panics, because timers aren't allowed in a bubble
    explanation: |
      The bubble's clock advances only when every goroutine in it is
      durably blocked, and then it jumps straight to the moment the next
      timer fires. Waiting 20 fake minutes costs microseconds.
  - question: Which of these does **not** count as "durably blocked" inside a bubble?
    options:
      - text: A receive on a channel created in the bubble
      - text: '`time.Sleep`'
      - text: Waiting to lock a `sync.Mutex`
        correct: true
      - text: '`sync.WaitGroup.Wait`'
    explanation: |
      A mutex could be unlocked by a goroutine outside the bubble, so
      synctest can't treat waiting on one as durable. While a goroutine is
      stuck on a mutex, the fake clock won't advance. Don't hold a lock
      while sleeping or waiting on a timer in code you test this way.
---

`testing/synctest` has two functions you'll use in almost every test: `synctest.Test` to run a test in a bubble, and `synctest.Wait`, covered in the next lesson.

## synctest.Test

```go
func Test(t *testing.T, f func(*testing.T))
```

`Test` runs `f` in a new goroutine inside an isolated **bubble**. Every goroutine started from inside the bubble joins it. Inside the bubble:

- The `time` package uses a **fake clock** that starts at midnight UTC on 2000-01-01.
- The clock **only moves when every goroutine in the bubble is durably blocked**. Then it jumps straight to the next moment that would unblock something, such as a timer firing or a `Sleep` ending.
- If every goroutine is blocked and **no** timer can ever unblock them, that's a deadlock, and `Test` panics instead of hanging.
- When `f` returns, `Test` waits for every goroutine in the bubble to exit. If some are blocked forever, that's a **leak**, and `Test` panics.

## Testing a timeout

Here's a Dispatchly function that waits for a courier to pick up an order, with a 20-minute timeout:

```go
var ErrPickupTimeout = errors.New("courier did not pick up in time")

func WaitForPickup(ctx context.Context, picked <-chan string, timeout time.Duration) (string, error) {
	select {
	case courier := <-picked:
		return courier, nil
	case <-time.After(timeout):
		return "", ErrPickupTimeout
	case <-ctx.Done():
		return "", context.Cause(ctx)
	}
}
```

The test uses the **real** 20-minute timeout, and even checks it to the nanosecond:

```go
func TestWaitForPickupTimesOut(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		picked := make(chan string) // no courier ever picks up
		start := time.Now()

		_, err := WaitForPickup(t.Context(), picked, 20*time.Minute)

		if err != ErrPickupTimeout {
			t.Errorf("err = %v, want ErrPickupTimeout", err)
		}
		if took := time.Since(start); took != 20*time.Minute {
			t.Errorf("gave up after %v, want exactly 20m0s", took)
		}
	})
}

func TestWaitForPickupSucceeds(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		picked := make(chan string)
		go func() {
			time.Sleep(7 * time.Minute)
			picked <- "ana"
		}()

		courier, err := WaitForPickup(t.Context(), picked, 20*time.Minute)
		if courier != "ana" || err != nil {
			t.Errorf("WaitForPickup = %q, %v; want ana, <nil>", courier, err)
		}
	})
}
```

```text
=== RUN   TestWaitForPickupTimesOut
--- PASS: TestWaitForPickupTimesOut (0.00s)
=== RUN   TestWaitForPickupSucceeds
--- PASS: TestWaitForPickupSucceeds (0.00s)
PASS
```

Twenty minutes of waiting, tested in 0.00s. In the first test the only goroutine is blocked in `select`, so the clock jumps 20 minutes and the timer fires. In the second, both goroutines block: the helper in `Sleep`, the test in `select`. The clock jumps 7 minutes, the helper sends, and the `select` takes that case. It's the same every run, on every machine.

Notice the use of `t.Context()` inside the bubble. The `t` passed to `f` belongs to the bubble, and its context is cancelled when `f` finishes.

## What "durably blocked" means

The clock can only advance when nothing in the bubble could possibly make progress. synctest counts a goroutine as **durably blocked** only if nothing *outside* the bubble could wake it:

| Durably blocked | Not durably blocked |
| --- | --- |
| send/receive on a channel created in the bubble | locking a `sync.Mutex` or `RWMutex` |
| `select` where every case is a bubble channel | network or file I/O |
| `time.Sleep` | system calls |
| `sync.Cond.Wait` | channels created outside the bubble |
| `sync.WaitGroup.Wait` (if `Add`/`Go` was called in the bubble) | |

The mutex entry is the one that surprises people. If code sleeps or waits on a timer while **holding a lock**, and another goroutine is waiting for that lock, time can't advance and the test hangs. That's usually a sign of a real design problem anyway (don't hold locks during slow operations), so synctest nudges you towards better code. For networking, use an in-memory fake such as `net.Pipe` instead of real sockets.

## Rules of the bubble

- Channels, timers and tickers created in a bubble belong to it. Using them from outside panics.
- Don't use goroutines or channels created *outside* the bubble from inside it.
- Keep each test self-contained: start what you need inside `f`, and make sure it all shuts down.

## Further reading

- [`testing/synctest` documentation](https://pkg.go.dev/testing/synctest)
