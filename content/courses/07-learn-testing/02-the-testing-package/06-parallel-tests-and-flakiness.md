---
title: Parallel Tests and Flaky Tests
quiz:
  - question: |
      What does this print with `go test -v`, in order?

      ```go
      func TestA(t *testing.T) {
          t.Parallel()
          t.Log("A")
      }

      func TestB(t *testing.T) {
          t.Log("B")
      }
      ```
    options:
      - text: A then B, always
      - text: B then A, always
        correct: true
      - text: A and B in a random order
      - text: Only B; parallel tests need `-parallel`
    explanation: |
      `t.Parallel()` pauses `TestA` (you'll see `=== PAUSE TestA`). Sequential
      top-level tests run first, then the paused parallel ones continue
      (`=== CONT TestA`). So B is always logged before A.
  - question: A test passes on your laptop and fails about one run in fifty on CI. Which command is the best first step to reproduce it?
    options:
      - text: '`go test -count=200 -run TestThatFlakes ./ledgerly`'
        correct: true
      - text: '`go test -short ./...`'
      - text: '`go vet ./...`'
      - text: '`go test -v ./...` once more'
    explanation: |
      `-count=N` runs the test N times (and bypasses the test cache), turning
      a 1-in-50 failure into a near-certain one. Add `-race` and
      `-shuffle=on` to shake out data races and order dependencies.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    )

    type Cents int64

    type Transaction struct {
    	ID      string
    	Account string
    	Amount  Cents
    }

    type Ledger struct {
    	txns []Transaction
    }

    // nextID numbers transactions: T0001, T0002, ...
    var nextID int

    // Add records tx, giving it the ledger's next ID. The first transaction
    // in every ledger must be T0001.
    func (l *Ledger) Add(account string, amount Cents) Transaction {
    	nextID++
    	tx := Transaction{ID: fmt.Sprintf("T%04d", nextID), Account: account, Amount: amount}
    	l.txns = append(l.txns, tx)
    	return tx
    }

    // Accounts returns each account name used in the ledger once, sorted.
    func (l *Ledger) Accounts() []string {
    	seen := map[string]bool{}
    	for _, tx := range l.txns {
    		seen[tx.Account] = true
    	}
    	var names []string
    	for name := range seen {
    		names = append(names, name)
    	}
    	return names
    }

    func main() {
    	var march, april Ledger
    	fmt.Println(march.Add("cash", 500).ID, march.Add("rent", -120000).ID)
    	fmt.Println(april.Add("cash", 700).ID, "<- should be T0001")

    	for _, a := range []string{"savings", "food", "travel", "bills"} {
    		march.Add(a, 100)
    	}
    	for range 3 {
    		fmt.Println(march.Accounts())
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"maps"
    	"slices"
    )

    type Cents int64

    type Transaction struct {
    	ID      string
    	Account string
    	Amount  Cents
    }

    type Ledger struct {
    	txns   []Transaction
    	nextID int // per ledger, so ledgers (and tests) don't affect each other
    }

    // Add records tx, giving it the ledger's next ID. The first transaction
    // in every ledger must be T0001.
    func (l *Ledger) Add(account string, amount Cents) Transaction {
    	l.nextID++
    	tx := Transaction{ID: fmt.Sprintf("T%04d", l.nextID), Account: account, Amount: amount}
    	l.txns = append(l.txns, tx)
    	return tx
    }

    // Accounts returns each account name used in the ledger once, sorted.
    func (l *Ledger) Accounts() []string {
    	seen := map[string]bool{}
    	for _, tx := range l.txns {
    		seen[tx.Account] = true
    	}
    	return slices.Sorted(maps.Keys(seen))
    }

    func main() {
    	var march, april Ledger
    	fmt.Println(march.Add("cash", 500).ID, march.Add("rent", -120000).ID)
    	fmt.Println(april.Add("cash", 700).ID, "<- should be T0001")

    	for _, a := range []string{"savings", "food", "travel", "bills"} {
    		march.Add(a, 100)
    	}
    	for range 3 {
    		fmt.Println(march.Accounts())
    	}
    }
  tests: |
    package main

    import (
    	"fmt"
    	"slices"
    	"testing"
    )

    func TestIDsStartAtOnePerLedger(t *testing.T) {
    	for i := range 3 {
    		var l Ledger
    		for n := 1; n <= 3; n++ {
    			want := fmt.Sprintf("T%04d", n)
    			if got := l.Add("cash", 100).ID; got != want {
    				t.Fatalf("ledger #%d: transaction %d got ID %q, want %q (each new Ledger starts at T0001)", i+1, n, got, want)
    			}
    		}
    	}
    }

    func TestIDsInParallel(t *testing.T) {
    	for i := range 4 {
    		t.Run(fmt.Sprint("ledger", i), func(t *testing.T) {
    			t.Parallel()
    			var l Ledger
    			for n := 1; n <= 200; n++ {
    				want := fmt.Sprintf("T%04d", n)
    				if got := l.Add("cash", 1).ID; got != want {
    					t.Fatalf("transaction %d got ID %q, want %q; ledgers must not share a counter", n, got, want)
    				}
    			}
    		})
    	}
    }

    func TestAccountsSortedAndUnique(t *testing.T) {
    	var l Ledger
    	for _, a := range []string{"rent", "cash", "savings", "food", "cash", "travel", "bills", "rent", "gym", "tax", "cash"} {
    		l.Add(a, 100)
    	}
    	want := []string{"bills", "cash", "food", "gym", "rent", "savings", "tax", "travel"}
    	for range 30 {
    		if got := l.Accounts(); !slices.Equal(got, want) {
    			t.Fatalf("Accounts() = %q, want %q", got, want)
    		}
    	}
    }

    func TestAccountsEmpty(t *testing.T) {
    	var l Ledger
    	if got := l.Accounts(); len(got) != 0 {
    		t.Errorf("Accounts() on an empty ledger = %q, want no accounts", got)
    	}
    }
---

Go can run tests in parallel, which speeds up slow suites and, as a bonus, flushes out code that isn't safe to run concurrently. It also exposes tests that only passed by luck. Let's look at both sides.

## t.Parallel

Call `t.Parallel()` at the top of a test to say "I can run alongside other parallel tests":

```go
func TestImportMarch(t *testing.T) {
	t.Parallel()
	// ...
}
```

How it schedules:

- A top-level parallel test **pauses** as soon as it calls `t.Parallel()`. All the sequential tests run first, then the paused parallel tests resume together.
- For subtests, parallel subtests run **after the parent's function body returns**, and the parent waits for them before its cleanups run.
- At most `-parallel n` tests run at once (default: `GOMAXPROCS`).

With `-v` you can watch it happen:

```text
=== RUN   TestImportMarch
=== PAUSE TestImportMarch
=== RUN   TestBalance
--- PASS: TestBalance (0.00s)
=== CONT  TestImportMarch
--- PASS: TestImportMarch (0.01s)
```

A parallel table test looks like this. Since Go 1.22 each loop iteration has its own `tt`, so the old `tt := tt` line is gone (`go fix`'s `forvar` fixer removes it):

```go
for _, tt := range cases {
	t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
		if got := tt.in.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	})
}
```

Only make a test parallel if it touches no shared mutable state. Remember that `t.Setenv` and `t.Chdir` refuse to run in parallel tests.

## Flaky tests

A **flaky** test passes sometimes and fails sometimes with no code change. The usual suspects:

1. **Shared global state.** A package-level variable that one test changes and another depends on. The result depends on which runs first, or on timing if they're parallel.
2. **Map iteration order.** Go deliberately randomizes it. Code that builds a slice by ranging over a map returns a different order each call, and a test comparing that slice to a fixed list fails at random.
3. **Time.** `time.Now()`, `time.Sleep(10 * time.Millisecond)` "to let the goroutine finish", timeouts that are fine on your laptop and too short on a busy CI box.
4. **Data races.** Unsynchronized access that usually works.
5. **The outside world.** Network, real files in shared places, other processes.

## Hunting them down

```text
go test -count=100 -run TestAccounts ./ledgerly   # repeat until it fails
go test -race ./...                               # detect data races
go test -shuffle=on ./...                         # randomize test order
```

`-shuffle=on` prints the seed it used (`-test.shuffle 1727312345`); pass that number back with `-shuffle=1727312345` to replay the same order.

The fix is almost never "retry" or "sleep longer". It's making the code or test **deterministic**: sort what needs an order, pass state explicitly instead of sharing globals, and inject time (chapter 4).

## Your task

A teammate's Ledgerly tests are flaky: they pass alone, fail together, and sometimes fail on their own. Two bugs in the *code* are to blame:

1. Transaction IDs come from a **package-level** counter, so the second ledger in a test run starts at `T0003` instead of `T0001`. Move the counter into `Ledger` so every ledger numbers its transactions from `T0001`, independently of every other ledger (including ones in parallel tests).
2. `Accounts` returns names in **map order**. Make it return each account once, **sorted**. (`slices.Sorted(maps.Keys(m))` does it in one line.)

**Run** shows both problems: the second ledger's ID, and the account order changing between calls.
