---
title: Red, Green, Refactor
quiz:
  - question: In TDD, why do you run the new test and watch it fail *before* writing the code?
    options:
      - text: To make sure the test runner is installed
      - text: To prove the test can fail; a test that passes before the feature exists isn't testing the feature
        correct: true
      - text: Because Go refuses to compile code before a test exists
      - text: To measure how slow the test is
    explanation: |
      A test that can't fail is worthless, and it's surprisingly easy to write
      one (a typo in the comparison, checking the wrong variable). Seeing red
      first, with a sensible failure message, proves the test is wired up.
  - question: |
      During the green step you wrote `return 3750` to make the test pass.
      What should you do next?
    options:
      - text: Ship it, the test is green
      - text: Delete the test since it's now passing
      - text: Write another test that the hard-coded value can't pass, forcing a real implementation
        correct: true
      - text: Add a comment saying the code is temporary
    explanation: |
      Faking it is a legitimate move: it keeps the step tiny. The next test
      (a different account, an empty ledger) makes the fake impossible, and
      you're pushed to the general solution.
  - question: What is allowed during the refactor step?
    options:
      - text: Adding new features while the tests are green
      - text: Changing the code's structure without changing its behaviour, running the tests after each change
        correct: true
      - text: Changing the tests so they match the new code
      - text: Nothing; refactoring is optional and usually skipped
    explanation: |
      Refactoring means improving the design with behaviour held constant.
      The tests are your safety net. New behaviour starts a new red step.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // balanceCase is one row of the Balance test table: build a ledger from
    // txns, then Balance(account) must return want.
    type balanceCase struct {
    	name    string
    	txns    []Transaction
    	account string
    	want    Cents
    }

    // balanceCases is the Balance test table. Both cases pass, yet a user has
    // reported a bug. Add a case that reproduces it (red), then fix Balance
    // (green).
    var balanceCases = []balanceCase{
    	{
    		name: "sums one account",
    		txns: []Transaction{
    			{Account: "cash", Amount: 5000},
    			{Account: "cash", Amount: -1250},
    			{Account: "savings", Amount: 999},
    		},
    		account: "cash",
    		want:    3750,
    	},
    	{
    		name:    "empty ledger",
    		account: "cash",
    		want:    0,
    	},
    	// ?
    }

    // ---- Ledgerly code ----

    // Balance returns the sum of every transaction for account, including its
    // sub-accounts: "savings" includes "savings:holiday".
    func (l *Ledger) Balance(account string) Cents {
    	var total Cents
    	for _, tx := range l.txns {
    		if strings.HasPrefix(tx.Account, account) {
    			total += tx.Amount
    		}
    	}
    	return total
    }

    type Cents int64

    type Transaction struct {
    	Account string
    	Amount  Cents
    }

    type Ledger struct{ txns []Transaction }

    func (l *Ledger) Add(tx Transaction) { l.txns = append(l.txns, tx) }

    func main() {
    	for _, tc := range balanceCases {
    		var l Ledger
    		for _, tx := range tc.txns {
    			l.Add(tx)
    		}
    		got := l.Balance(tc.account)
    		status := "ok  "
    		if got != tc.want {
    			status = "FAIL"
    		}
    		fmt.Printf("%s %-24s Balance(%q) = %d, want %d\n", status, tc.name, tc.account, got, tc.want)
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // balanceCase is one row of the Balance test table: build a ledger from
    // txns, then Balance(account) must return want.
    type balanceCase struct {
    	name    string
    	txns    []Transaction
    	account string
    	want    Cents
    }

    // balanceCases is the Balance test table.
    var balanceCases = []balanceCase{
    	{
    		name: "sums one account",
    		txns: []Transaction{
    			{Account: "cash", Amount: 5000},
    			{Account: "cash", Amount: -1250},
    			{Account: "savings", Amount: 999},
    		},
    		account: "cash",
    		want:    3750,
    	},
    	{
    		name:    "empty ledger",
    		account: "cash",
    		want:    0,
    	},
    	{
    		name: "cash is not cashback",
    		txns: []Transaction{
    			{Account: "cash", Amount: 2000},
    			{Account: "cashback", Amount: 150},
    		},
    		account: "cash",
    		want:    2000,
    	},
    	{
    		name: "sub-accounts roll up",
    		txns: []Transaction{
    			{Account: "savings", Amount: 10000},
    			{Account: "savings:holiday", Amount: 2500},
    			{Account: "savings:car", Amount: 500},
    		},
    		account: "savings",
    		want:    13000,
    	},
    }

    // ---- Ledgerly code ----

    // Balance returns the sum of every transaction for account, including its
    // sub-accounts: "savings" includes "savings:holiday".
    func (l *Ledger) Balance(account string) Cents {
    	var total Cents
    	for _, tx := range l.txns {
    		if tx.Account == account || strings.HasPrefix(tx.Account, account+":") {
    			total += tx.Amount
    		}
    	}
    	return total
    }

    type Cents int64

    type Transaction struct {
    	Account string
    	Amount  Cents
    }

    type Ledger struct{ txns []Transaction }

    func (l *Ledger) Add(tx Transaction) { l.txns = append(l.txns, tx) }

    func main() {
    	for _, tc := range balanceCases {
    		var l Ledger
    		for _, tx := range tc.txns {
    			l.Add(tx)
    		}
    		got := l.Balance(tc.account)
    		status := "ok  "
    		if got != tc.want {
    			status = "FAIL"
    		}
    		fmt.Printf("%s %-24s Balance(%q) = %d, want %d\n", status, tc.name, tc.account, got, tc.want)
    	}
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    )

    type balanceFunc func(txns []Transaction, account string) Cents

    func sum(txns []Transaction, match func(string) bool) Cents {
    	var total Cents
    	for _, tx := range txns {
    		if match(tx.Account) {
    			total += tx.Amount
    		}
    	}
    	return total
    }

    // correct is how Balance should behave.
    func correct(txns []Transaction, account string) Cents {
    	return sum(txns, func(a string) bool { return a == account || strings.HasPrefix(a, account+":") })
    }

    // reported is the Balance from the bug report.
    func reported(txns []Transaction, account string) Cents {
    	return sum(txns, func(a string) bool { return strings.HasPrefix(a, account) })
    }

    // exactOnly is an over-eager fix that forgets about sub-accounts.
    func exactOnly(txns []Transaction, account string) Cents {
    	return sum(txns, func(a string) bool { return a == account })
    }

    // caught reports whether some case in the table fails against f.
    func caught(f balanceFunc) bool {
    	for _, tc := range balanceCases {
    		if f(tc.txns, tc.account) != tc.want {
    			return true
    		}
    	}
    	return false
    }

    func TestTableIsCorrect(t *testing.T) {
    	seen := map[string]bool{}
    	for _, tc := range balanceCases {
    		if tc.name == "" || seen[tc.name] {
    			t.Errorf("case names must be non-empty and unique; found %q", tc.name)
    		}
    		seen[tc.name] = true
    		if got := correct(tc.txns, tc.account); got != tc.want {
    			t.Errorf("case %q: a correct Balance(%q) returns %d, but the case wants %d; fix the case's want", tc.name, tc.account, got, tc.want)
    		}
    	}
    }

    func TestTableReproducesBug(t *testing.T) {
    	if !caught(reported) {
    		t.Error("no case fails against the reported Balance: add a case where the account's name is the start of another account's name (like \"cash\" and \"cashback\")")
    	}
    }

    func TestTableCoversSubAccounts(t *testing.T) {
    	if !caught(exactOnly) {
    		t.Error("no case fails against a Balance that only counts exact matches: add a case where sub-accounts (like \"savings:holiday\") must roll up into their parent")
    	}
    }

    func TestBalance(t *testing.T) {
    	var l Ledger
    	for _, tx := range []Transaction{
    		{Account: "cash", Amount: 2000},
    		{Account: "cashback", Amount: 150},
    		{Account: "savings", Amount: 10000},
    		{Account: "savings:holiday", Amount: 2500},
    		{Account: "savings:holidays-old", Amount: 1},
    		{Account: "savings2", Amount: 700},
    	} {
    		l.Add(tx)
    	}
    	for _, tc := range []struct {
    		account string
    		want    Cents
    	}{
    		{"cash", 2000},
    		{"cashback", 150},
    		{"savings", 12501},
    		{"savings:holiday", 2500},
    		{"sav", 0},
    		{"rent", 0},
    	} {
    		if got := l.Balance(tc.account); got != tc.want {
    			t.Errorf("Balance(%q) = %d, want %d", tc.account, got, tc.want)
    		}
    	}
    }
---

You met **test-driven development** (TDD) briefly in course 01. It's a rhythm, not a religion, and it's worth seeing one full loop in slow motion, including the parts people skip. The three tiny steps:

1. **Red**: write a test for behaviour that doesn't exist yet. Run it. Watch it fail.
2. **Green**: write the *simplest* code that makes it pass. Hacks allowed.
3. **Refactor**: clean up code and tests while everything stays green.

Let's do one full loop on Ledgerly's first feature: working out an account's balance.

## Red: start with the test

We have transactions and a ledger that stores them, but no way to ask for a balance. Before writing `Balance`, we write the test we *wish* would pass:

```go
func TestBalanceSumsOneAccount(t *testing.T) {
	var l Ledger
	l.Add(Transaction{Account: "cash", Amount: 5000})
	l.Add(Transaction{Account: "cash", Amount: -1250})
	l.Add(Transaction{Account: "savings", Amount: 999})

	if got, want := l.Balance("cash"), Cents(3750); got != want {
		t.Errorf("Balance(%q) = %d, want %d", "cash", got, want)
	}
}
```

Notice we're designing the API from the caller's side: a method on `*Ledger`, taking an account name, returning `Cents`. Run it:

```text
$ go test ./...
# example.com/ledgerly [example.com/ledgerly.test]
./ledger_test.go:11:20: l.Balance undefined (type Ledger has no field or method Balance)
FAIL	example.com/ledgerly [build failed]
```

A compile error counts as red. Go's first nudge: define the method. Do the *minimum* to compile, so we can see the test fail properly:

```go
func (l *Ledger) Balance(account string) Cents {
	return 0
}
```

```text
--- FAIL: TestBalanceSumsOneAccount (0.00s)
    ledger_test.go:12: Balance("cash") = 0, want 3750
FAIL
```

That's a *good* red: the test runs, fails for the right reason, and the message says what went wrong.

## Green: make it pass

The simplest code that passes? Honestly, `return 3750`. That feels silly, but it's a real TDD move called *fake it till you make it*. It proves the test can go green, and it tells you that you need **another test** to force the real logic. So add one that a constant can't satisfy:

```go
func TestBalanceEmptyLedger(t *testing.T) {
	var l Ledger
	if got := l.Balance("cash"); got != 0 {
		t.Errorf("Balance(%q) on empty ledger = %d, want 0", "cash", got)
	}
}
```

Now `return 3750` fails one test and `return 0` fails the other. Time for the real thing:

```go
// Balance returns the sum of every transaction for account.
func (l *Ledger) Balance(account string) Cents {
	var total Cents
	for _, t := range l.txns {
		if t.Account == account {
			total += t.Amount
		}
	}
	return total
}
```

```text
$ go test -v ./...
=== RUN   TestBalanceSumsOneAccount
--- PASS: TestBalanceSumsOneAccount (0.00s)
=== RUN   TestBalanceEmptyLedger
--- PASS: TestBalanceEmptyLedger (0.00s)
PASS
```

Green.

## Refactor: tidy with a safety net

Now look at what we have, code *and* tests. Is anything duplicated or unclear? The two tests follow the same shape (build a ledger, check a balance), which hints they could become a table later. The loop variable `t` in `Balance` shadows nothing here, but in a test file `t` usually means `*testing.T`, so renaming it `tx` avoids confusion down the line. Make each change, rerun the tests, stay green.

The rule during refactoring: **no new behaviour**. If you catch yourself thinking "and while I'm here, let me handle currencies", stop. That's the next red.

## Why bother?

- You never write code that no test asked for, so everything is covered.
- Each step is tiny, so when something breaks you know exactly which change did it.
- You design the API as a *user* first, which tends to produce nicer APIs.

You don't have to TDD everything. Exploratory code, UI tweaks and one-off scripts often don't benefit. But for logic with clear rules, like money, it's a superb habit, and for **bug fixes** it's nearly always right: first write a test that reproduces the bug (red), then fix it (green).

## Your turn: a bug report

Ledgerly recently gained **sub-accounts**: `Balance("savings")` includes `savings:holiday` and `savings:car`, so you can see the whole pot at once. The code, and both existing test cases, look fine. Then a user writes in:

> My `cash` balance is $1.50 too high. It seems to include my `cashback` rewards!

Do it the TDD way:

1. **Red.** Add a case to `balanceCases` that reproduces the report. Press **Run**: your new case should print `FAIL` against the current `Balance`. That proves the test catches the bug.
2. **Green.** Fix `Balance` so an account's total includes the account itself and anything starting with the account name *followed by a colon*, and nothing else.
3. **Guard the feature.** An easy "fix" is to count exact matches only, which quietly breaks sub-accounts. Add a case that would catch that too.

The grader checks that every case in your table is correct, that some case fails against the buggy `Balance` from the report, that some case fails against the exact-match-only "fix", and then tests your `Balance` directly.

## Further reading

- [Learn Go with Tests: Integers](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/integers) walks through the same cycle.
