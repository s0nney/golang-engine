---
title: Designing for Testability
quiz:
  - question: |
      Why is this function hard to test?

      ```go
      func MonthlyFee() Cents {
          if time.Now().Day() == 1 {
              return 500
          }
          return 0
      }
      ```
    options:
      - text: It returns a named type
      - text: It reaches out to the real clock itself, so a test can't choose what "now" is
        correct: true
      - text: It has an `if` statement
      - text: Functions without parameters can't be tested
    explanation: |
      The result depends on a hidden input, the current date. A test can only
      check one branch, and which one depends on the day you run it. Pass the
      time in (or inject a clock) and both branches become easy to test.
  - question: Which part of a program should hold most of the logic, if you want it to be easy to test?
    options:
      - text: The `main` function
      - text: The code that talks to the network and the disk
      - text: Pure functions that take plain values and return plain values
        correct: true
      - text: Global variables
    explanation: |
      Pure functions need no setup, no doubles and no cleanup: call them and
      compare the result. Keep the I/O at the edges, as thin as you can, and
      push decisions into the pure core.
---

So far every Ledgerly function we've tested was easy: numbers in, numbers out. Real code talks to banks, disks, clocks and email servers. You can't hit a real bank's API in a unit test, so you need a way to swap it for something you control. This chapter is about those swaps, called **test doubles**. First, though, the code has to *allow* a swap.

## A function that fights its tests

Here's a first draft of Ledgerly's bank sync:

```go
func SyncFromBank(l *Ledger) error {
	resp, err := http.Get("https://api.examplebank.com/v1/transactions")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var txns []Transaction
	if err := json.NewDecoder(resp.Body).Decode(&txns); err != nil {
		return err
	}
	for _, tx := range txns {
		if tx.Date.After(time.Now()) {
			continue // skip future-dated transactions
		}
		l.Add(tx)
	}
	return nil
}
```

To test "future-dated transactions are skipped" you'd need the real bank to return one, on a day you can predict. The test would be slow, flaky, need credentials and a network, and break whenever the bank's sandbox is down. The function has **hidden inputs**: the network and the clock. It grabs them itself, so a test can't substitute anything.

## Seams

A **seam** is a place where you can change what code does without editing it. In Go, the seams are:

- **Parameters.** The simplest seam. Pass values in instead of fetching them.
- **Interfaces.** Depend on a small interface, and a test passes its own implementation.
- **Function values.** A field or parameter of type `func() time.Time` is an interface with one method and no ceremony.

Rewriting with seams:

```go
type TransactionSource interface {
	Transactions(ctx context.Context) ([]Transaction, error)
}

func Sync(ctx context.Context, l *Ledger, src TransactionSource, now time.Time) error {
	txns, err := src.Transactions(ctx)
	if err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	for _, tx := range txns {
		if tx.Date.After(now) {
			continue
		}
		l.Add(tx)
	}
	return nil
}
```

`Sync` no longer knows about HTTP, JSON or the wall clock. The HTTP-and-JSON code lives in a separate `BankClient` type that implements `TransactionSource`, and gets its own (fewer, slower) integration tests. `Sync` now has one job, and a test controls every input.

## Functional core, imperative shell

Push this further and you get a common shape for testable programs:

- **A functional core**: pure functions and plain data. Parsing, validating, computing balances, deciding what's due. No I/O, no clock, no globals. Tested with simple tables.
- **An imperative shell**: a thin layer that reads files, calls APIs, reads the clock and passes the results into the core. Tested lightly, often with integration tests.

The core is where the bugs are (it's where the logic is), so that's where you want tests to be cheap.

## Signs code will be hard to test

- It calls `time.Now()`, `rand.IntN`, `os.Getenv` or `http.Get` deep inside.
- It uses package-level variables that tests would have to set and reset.
- A constructor does real work: dials a database, reads a config file.
- It takes a huge concrete type (`*sql.DB`, `*http.Client`) when it only uses one method.

None of these is a crime in `main`. Deep inside your logic, each one is a hidden input.

## Don't overdo it

Designing for testability doesn't mean an interface for every type. If a dependency is fast, deterministic and local (a `strings.Builder`, a `Ledger`, a pure helper), use the real thing in tests. Doubles are for dependencies that are slow, non-deterministic, or have side effects you can't allow in a test. The rest of this chapter shows how to build them by hand, no mocking library required.
