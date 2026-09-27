---
title: Test Helpers
quiz:
  - question: |
      `assertCents` is called on line 16 of `ledger_test.go` and its
      `t.Errorf` is on line 10. It does **not** call `t.Helper()`. Which line
      does the failure report?
    options:
      - text: '`ledger_test.go:16`'
      - text: '`ledger_test.go:10`'
        correct: true
      - text: Both lines
      - text: No line number is shown for helpers
    explanation: |
      Without `t.Helper()`, the testing package reports the line that called
      `Errorf`, which is inside the helper. Every failure then points at the
      same line and you can't tell which call failed. `t.Helper()` makes it
      skip the helper's frame and report line 16.
  - question: Why do helpers usually take `testing.TB` instead of `*testing.T`?
    options:
      - text: '`testing.TB` is faster'
      - text: So the same helper works in tests, benchmarks and fuzz tests, which get `*testing.T`, `*testing.B` and `*testing.F`
        correct: true
      - text: '`*testing.T` can''t be passed to functions'
      - text: '`testing.TB` doesn''t have `Helper`'
    explanation: |
      `testing.TB` is the interface that `*T`, `*B` and `*F` all satisfy. A
      helper that takes it can be shared across every kind of test.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strconv"
    	"strings"
    	"testing"
    )

    // assertBalance fails the test (without stopping it) if l's balance for
    // account isn't want. The message must look like:
    //
    //	Balance("cash") = $12.50, want $20.00
    func assertBalance(t testing.TB, l *Ledger, account string, want Cents) {
    	// ?
    }

    // mustParse parses s with ParseAmount and stops the test with t.Fatalf
    // if that fails. The message must include s, quoted with %q.
    func mustParse(t testing.TB, s string) Cents {
    	// ?
    	return 0
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    var ErrBadAmount = errors.New("ledgerly: bad amount")

    // ParseAmount parses "12.34" or "12" into cents.
    func ParseAmount(s string) (Cents, error) {
    	whole, frac, hasDot := strings.Cut(s, ".")
    	d, err := strconv.ParseInt(whole, 10, 64)
    	if err != nil || (hasDot && len(frac) != 2) {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	var c int64
    	if hasDot {
    		if c, err = strconv.ParseInt(frac, 10, 64); err != nil {
    			return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    		}
    	}
    	return Cents(d*100 + c), nil
    }

    type Transaction struct {
    	Account string
    	Amount  Cents
    }

    type Ledger struct{ txns []Transaction }

    func (l *Ledger) Add(t Transaction) { l.txns = append(l.txns, t) }

    func (l *Ledger) Balance(account string) Cents {
    	var total Cents
    	for _, tx := range l.txns {
    		if tx.Account == account {
    			total += tx.Amount
    		}
    	}
    	return total
    }

    // ---- A pretend testing.TB so Run can show what your helpers report ----

    type printTB struct{ testing.TB }

    func (printTB) Helper()                        { fmt.Println("  (Helper called)") }
    func (printTB) Errorf(format string, a ...any) { fmt.Printf("  Errorf: "+format+"\n", a...) }
    func (printTB) Fatalf(format string, a ...any) { fmt.Printf("  Fatalf: "+format+"\n", a...) }

    func main() {
    	var t printTB
    	var l Ledger
    	l.Add(Transaction{Account: "cash", Amount: 1250})

    	fmt.Println("assertBalance with the right balance:")
    	assertBalance(t, &l, "cash", 1250)
    	fmt.Println("assertBalance with the wrong balance:")
    	assertBalance(t, &l, "cash", 2000)
    	fmt.Println("mustParse(\"12.50\"):")
    	fmt.Println("  returned", mustParse(t, "12.50"))
    	fmt.Println("mustParse(\"12.x\"):")
    	mustParse(t, "12.x")
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strconv"
    	"strings"
    	"testing"
    )

    // assertBalance fails the test (without stopping it) if l's balance for
    // account isn't want.
    func assertBalance(t testing.TB, l *Ledger, account string, want Cents) {
    	t.Helper()
    	if got := l.Balance(account); got != want {
    		t.Errorf("Balance(%q) = %v, want %v", account, got, want)
    	}
    }

    // mustParse parses s with ParseAmount and stops the test if that fails.
    func mustParse(t testing.TB, s string) Cents {
    	t.Helper()
    	c, err := ParseAmount(s)
    	if err != nil {
    		t.Fatalf("ParseAmount(%q): %v", s, err)
    	}
    	return c
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    var ErrBadAmount = errors.New("ledgerly: bad amount")

    // ParseAmount parses "12.34" or "12" into cents.
    func ParseAmount(s string) (Cents, error) {
    	whole, frac, hasDot := strings.Cut(s, ".")
    	d, err := strconv.ParseInt(whole, 10, 64)
    	if err != nil || (hasDot && len(frac) != 2) {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	var c int64
    	if hasDot {
    		if c, err = strconv.ParseInt(frac, 10, 64); err != nil {
    			return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    		}
    	}
    	return Cents(d*100 + c), nil
    }

    type Transaction struct {
    	Account string
    	Amount  Cents
    }

    type Ledger struct{ txns []Transaction }

    func (l *Ledger) Add(t Transaction) { l.txns = append(l.txns, t) }

    func (l *Ledger) Balance(account string) Cents {
    	var total Cents
    	for _, tx := range l.txns {
    		if tx.Account == account {
    			total += tx.Amount
    		}
    	}
    	return total
    }

    // ---- A pretend testing.TB so Run can show what your helpers report ----

    type printTB struct{ testing.TB }

    func (printTB) Helper()                        { fmt.Println("  (Helper called)") }
    func (printTB) Errorf(format string, a ...any) { fmt.Printf("  Errorf: "+format+"\n", a...) }
    func (printTB) Fatalf(format string, a ...any) { fmt.Printf("  Fatalf: "+format+"\n", a...) }

    func main() {
    	var t printTB
    	var l Ledger
    	l.Add(Transaction{Account: "cash", Amount: 1250})

    	fmt.Println("assertBalance with the right balance:")
    	assertBalance(t, &l, "cash", 1250)
    	fmt.Println("assertBalance with the wrong balance:")
    	assertBalance(t, &l, "cash", 2000)
    	fmt.Println("mustParse(\"12.50\"):")
    	fmt.Println("  returned", mustParse(t, "12.50"))
    	fmt.Println("mustParse(\"12.x\"):")
    	mustParse(t, "12.x")
    }
  tests: |
    package main

    import (
    	"fmt"
    	"runtime"
    	"strings"
    	"testing"
    )

    // spyTB records what a helper does. Methods it doesn't define panic,
    // because the embedded testing.TB is nil.
    type spyTB struct {
    	testing.TB
    	helper   bool
    	errors   []string
    	fatals   []string
    	panicked string
    }

    func (s *spyTB) Helper()                 { s.helper = true }
    func (s *spyTB) Log(a ...any)            {}
    func (s *spyTB) Logf(f string, a ...any) {}
    func (s *spyTB) Fail()                   { s.errors = append(s.errors, "(Fail)") }
    func (s *spyTB) Error(a ...any)          { s.errors = append(s.errors, fmt.Sprint(a...)) }
    func (s *spyTB) Errorf(f string, a ...any) {
    	s.errors = append(s.errors, fmt.Sprintf(f, a...))
    }
    func (s *spyTB) FailNow() {
    	s.fatals = append(s.fatals, "(FailNow)")
    	runtime.Goexit()
    }
    func (s *spyTB) Fatal(a ...any) {
    	s.fatals = append(s.fatals, fmt.Sprint(a...))
    	runtime.Goexit()
    }
    func (s *spyTB) Fatalf(f string, a ...any) {
    	s.fatals = append(s.fatals, fmt.Sprintf(f, a...))
    	runtime.Goexit()
    }

    // run calls f with a fresh spy on its own goroutine, so Fatal can stop it
    // the way the real testing package does.
    func run(f func(tb testing.TB)) *spyTB {
    	s := &spyTB{}
    	done := make(chan struct{})
    	go func() {
    		defer close(done)
    		defer func() {
    			if r := recover(); r != nil {
    				s.panicked = fmt.Sprint(r)
    			}
    		}()
    		f(s)
    	}()
    	<-done
    	return s
    }

    func ledger() *Ledger {
    	var l Ledger
    	l.Add(Transaction{Account: "cash", Amount: 1000})
    	l.Add(Transaction{Account: "cash", Amount: 250})
    	l.Add(Transaction{Account: "savings", Amount: 9900})
    	return &l
    }

    func checkNoPanic(t *testing.T, s *spyTB, call string) {
    	t.Helper()
    	if s.panicked != "" {
    		t.Fatalf("%s panicked: %s (only use Helper, Errorf and Fatalf on t)", call, s.panicked)
    	}
    }

    func TestAssertBalancePasses(t *testing.T) {
    	s := run(func(tb testing.TB) { assertBalance(tb, ledger(), "cash", 1250) })
    	checkNoPanic(t, s, "assertBalance")
    	if !s.helper {
    		t.Error("assertBalance didn't call t.Helper()")
    	}
    	if len(s.errors)+len(s.fatals) != 0 {
    		t.Errorf("assertBalance(cash, $12.50) reported %q%q, want no failure (the balance is right)", s.errors, s.fatals)
    	}
    }

    func TestAssertBalanceFails(t *testing.T) {
    	s := run(func(tb testing.TB) { assertBalance(tb, ledger(), "cash", 2000) })
    	checkNoPanic(t, s, "assertBalance")
    	if !s.helper {
    		t.Error("assertBalance didn't call t.Helper()")
    	}
    	if len(s.fatals) != 0 {
    		t.Fatalf("assertBalance called Fatal/FailNow (%q); use Errorf so the test keeps going", s.fatals)
    	}
    	if len(s.errors) != 1 {
    		t.Fatalf("assertBalance with a wrong balance reported %d errors, want exactly 1", len(s.errors))
    	}
    	msg := s.errors[0]
    	for _, want := range []string{`"cash"`, "$12.50", "$20.00"} {
    		if !strings.Contains(msg, want) {
    			t.Errorf("assertBalance message %q doesn't contain %s", msg, want)
    		}
    	}
    	if strings.Index(msg, "$12.50") > strings.Index(msg, "$20.00") {
    		t.Errorf("assertBalance message %q: put got ($12.50) before want ($20.00)", msg)
    	}
    }

    func TestAssertBalanceOtherAccount(t *testing.T) {
    	s := run(func(tb testing.TB) { assertBalance(tb, ledger(), "savings", 9900) })
    	checkNoPanic(t, s, "assertBalance")
    	if len(s.errors)+len(s.fatals) != 0 {
    		t.Errorf("assertBalance(savings, $99.00) reported %q%q, want no failure", s.errors, s.fatals)
    	}
    }

    func TestMustParseOK(t *testing.T) {
    	var got Cents
    	s := run(func(tb testing.TB) { got = mustParse(tb, "12.50") })
    	checkNoPanic(t, s, "mustParse")
    	if !s.helper {
    		t.Error("mustParse didn't call t.Helper()")
    	}
    	if len(s.errors)+len(s.fatals) != 0 {
    		t.Errorf("mustParse(%q) reported %q%q, want no failure", "12.50", s.errors, s.fatals)
    	}
    	if got != 1250 {
    		t.Errorf("mustParse(%q) = %d, want 1250", "12.50", got)
    	}
    }

    func TestMustParseBad(t *testing.T) {
    	s := run(func(tb testing.TB) { mustParse(tb, "12.x") })
    	checkNoPanic(t, s, "mustParse")
    	if !s.helper {
    		t.Error("mustParse didn't call t.Helper()")
    	}
    	if len(s.fatals) != 1 {
    		t.Fatalf("mustParse(%q) called Fatal %d times (errors: %q), want 1 call to t.Fatalf", "12.x", len(s.fatals), s.errors)
    	}
    	if !strings.Contains(s.fatals[0], `"12.x"`) {
    		t.Errorf("mustParse message %q doesn't include the input quoted with %%q", s.fatals[0])
    	}
    }
---

Tests repeat themselves. Checking a balance, parsing an amount, building a ledger: after the third copy-paste it's time for a **helper**, an ordinary function that takes the test's `t` and does some checking or setup for it.

## The problem with naive helpers

```go
func assertCents(t testing.TB, got, want Cents) {
	if got != want {
		t.Errorf("got %d, want %d", got, want) // line 10
	}
}

func TestBalances(t *testing.T) {
	assertCents(t, 100, 100)
	assertCents(t, 250, 300) // line 16
}
```

```text
--- FAIL: TestBalances (0.00s)
    ledger_test.go:10: got 250, want 300
```

Line 10 is inside the helper. With twenty calls to `assertCents`, which one failed? No idea.

## t.Helper

Add one line at the top of the helper:

```go
func assertCents(t testing.TB, got, want Cents) {
	t.Helper()
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}
```

```text
--- FAIL: TestBalances (0.00s)
    ledger_test.go:16: got 250, want 300
```

`t.Helper()` marks the calling function as a helper, so when it reports a failure the testing package skips its frame and blames the line that *called* it. It stacks, too: a helper calling a helper works as long as both call `t.Helper()`.

## Two kinds of helpers

- **Assertion helpers** check something and report with `t.Errorf`, so the test keeps going and you see every failed check.
- **Setup helpers** build something the test needs (a ledger, a parsed amount, a temp file). If setup fails there's nothing to test, so they call `t.Fatalf`. They're often named `mustXxx` or `newTestXxx` and return the built value, which keeps the test body clean:

```go
func TestTransfer(t *testing.T) {
	amount := mustParse(t, "25.00") // no err to check here
	// ...
}
```

Take `testing.TB` rather than `*testing.T` so benchmarks and fuzz tests can share the helper. And keep helpers small: a helper that hides *what* is being tested makes failures harder to read, not easier.

## Your task

Write two Ledgerly test helpers in `main.go`:

1. `assertBalance(t, l, account, want)` checks `l.Balance(account)`. On a mismatch it reports with `t.Errorf` (not `Fatalf`) using the message `Balance("cash") = $12.50, want $20.00`: the account quoted with `%q`, and got before want, both printed with `%v` so `Cents.String` kicks in.
2. `mustParse(t, s)` returns `ParseAmount(s)`, or stops the test with `t.Fatalf` if it fails. The message must include `s` quoted with `%q`.

Both must call `t.Helper()`. The grader calls your helpers with a *spy* `testing.TB` that records what they do. (Yes, the grader uses a test double to test your test helpers. You'll write your own doubles in chapter 4.) **Run** uses a similar pretend `TB` that prints each call.
