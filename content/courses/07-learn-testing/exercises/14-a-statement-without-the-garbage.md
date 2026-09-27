---
title: A Statement Without the Garbage
difficulty: hard
after: benchmarks
hints:
  - 'Every `+=` on a string copies everything so far into a new string, and every `fmt.Sprintf` allocates its result (and boxes its arguments). Build the whole statement in **one** `[]byte` instead, created with `make([]byte, 0, capacity)` where the capacity is a generous estimate from `len(txns)`, and convert it to a string once at the end. That''s two allocations.'
  - 'Most standard types can append themselves without allocating: `t.Date.AppendFormat(buf, time.DateOnly)`, `strconv.AppendInt(buf, n, 10)` and plain `append(buf, s...)` for strings. For padding, append `width - len(s)` spaces (none if it''s negative, since `%-12s` never truncates).'
  - 'For the amounts, format into a small stack array first (`var tmp [32]byte`, then `s := appendCents(tmp[:0], c)`) so you know the length before padding. Inside `appendCents`, take the magnitude as a `uint64` (`n := uint64(c); if c < 0 { n = -n }`), which also works for `math.MinInt64`, then add a comma before each digit whose distance from the end is a multiple of 3.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strconv"
    	"testing"
    	"time"
    )

    // Statement returns exactly the same text as slowStatement, but allocates
    // at most twice per call.
    func Statement(txns []Txn) string {
    	return slowStatement(txns)
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    type Txn struct {
    	Date    time.Time
    	Account string
    	Amount  Cents
    }

    // String formats c like "$1,234.56" or "-$0.05".
    func (c Cents) String() string {
    	n := uint64(c)
    	sign := ""
    	if c < 0 {
    		sign, n = "-", -n
    	}
    	dollars := strconv.FormatUint(n/100, 10)
    	for i := len(dollars) - 3; i > 0; i -= 3 {
    		dollars = dollars[:i] + "," + dollars[i:]
    	}
    	return fmt.Sprintf("%s$%s.%02d", sign, dollars, n%100)
    }

    // slowStatement is the reference: correct, readable and wasteful.
    func slowStatement(txns []Txn) string {
    	out := ""
    	var bal Cents
    	for _, t := range txns {
    		bal += t.Amount
    		out += fmt.Sprintf("%s  %-12s %12s %12s\n", t.Date.Format(time.DateOnly), t.Account, t.Amount, bal)
    	}
    	out += fmt.Sprintf("%d transactions, closing balance %s\n", len(txns), bal)
    	return out
    }

    // sampleMonth returns n transactions for Run and the grader.
    func sampleMonth(n int) []Txn {
    	accounts := []string{"rent", "groceries", "cash", "fuel", "subscriptions"}
    	txns := make([]Txn, n)
    	for i := range txns {
    		txns[i] = Txn{
    			Date:    time.Date(2026, 3, 1+i%31, 0, 0, 0, 0, time.UTC),
    			Account: accounts[i%len(accounts)],
    			Amount:  Cents((i*104_729)%500_000 - 200_000),
    		}
    	}
    	return txns
    }

    func main() {
    	fmt.Print(Statement(sampleMonth(6)))
    	txns := sampleMonth(500)
    	fmt.Println("same as slowStatement:", Statement(txns) == slowStatement(txns))
    	fmt.Println("allocations per call: ", testing.AllocsPerRun(20, func() { Statement(txns) }))
    	fmt.Println("slowStatement allocs: ", testing.AllocsPerRun(20, func() { slowStatement(txns) }))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strconv"
    	"testing"
    	"time"
    )

    // Statement returns the same text as slowStatement, but allocates at most
    // twice per call.
    func Statement(txns []Txn) string {
    	buf := make([]byte, 0, 64*len(txns)+64)
    	var bal Cents
    	for _, t := range txns {
    		bal += t.Amount
    		buf = t.Date.AppendFormat(buf, time.DateOnly)
    		buf = append(buf, "  "...)
    		buf = append(buf, t.Account...)
    		for range 12 - len(t.Account) {
    			buf = append(buf, ' ')
    		}
    		buf = append(buf, ' ')
    		buf = appendPadded(buf, t.Amount, 12)
    		buf = append(buf, ' ')
    		buf = appendPadded(buf, bal, 12)
    		buf = append(buf, '\n')
    	}
    	buf = strconv.AppendInt(buf, int64(len(txns)), 10)
    	buf = append(buf, " transactions, closing balance "...)
    	buf = appendCents(buf, bal)
    	buf = append(buf, '\n')
    	return string(buf)
    }

    // appendPadded appends c right-aligned in a field of width bytes.
    func appendPadded(dst []byte, c Cents, width int) []byte {
    	var tmp [32]byte
    	s := appendCents(tmp[:0], c)
    	for range width - len(s) {
    		dst = append(dst, ' ')
    	}
    	return append(dst, s...)
    }

    // appendCents appends c formatted like c.String().
    func appendCents(dst []byte, c Cents) []byte {
    	n := uint64(c)
    	if c < 0 {
    		dst = append(dst, '-')
    		n = -n
    	}
    	dst = append(dst, '$')
    	var digits [20]byte
    	d := strconv.AppendUint(digits[:0], n/100, 10)
    	for i, ch := range d {
    		if i > 0 && (len(d)-i)%3 == 0 {
    			dst = append(dst, ',')
    		}
    		dst = append(dst, ch)
    	}
    	cents := n % 100
    	return append(dst, '.', byte('0'+cents/10), byte('0'+cents%10))
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    type Txn struct {
    	Date    time.Time
    	Account string
    	Amount  Cents
    }

    // String formats c like "$1,234.56" or "-$0.05".
    func (c Cents) String() string {
    	n := uint64(c)
    	sign := ""
    	if c < 0 {
    		sign, n = "-", -n
    	}
    	dollars := strconv.FormatUint(n/100, 10)
    	for i := len(dollars) - 3; i > 0; i -= 3 {
    		dollars = dollars[:i] + "," + dollars[i:]
    	}
    	return fmt.Sprintf("%s$%s.%02d", sign, dollars, n%100)
    }

    // slowStatement is the reference: correct, readable and wasteful.
    func slowStatement(txns []Txn) string {
    	out := ""
    	var bal Cents
    	for _, t := range txns {
    		bal += t.Amount
    		out += fmt.Sprintf("%s  %-12s %12s %12s\n", t.Date.Format(time.DateOnly), t.Account, t.Amount, bal)
    	}
    	out += fmt.Sprintf("%d transactions, closing balance %s\n", len(txns), bal)
    	return out
    }

    // sampleMonth returns n transactions for Run and the grader.
    func sampleMonth(n int) []Txn {
    	accounts := []string{"rent", "groceries", "cash", "fuel", "subscriptions"}
    	txns := make([]Txn, n)
    	for i := range txns {
    		txns[i] = Txn{
    			Date:    time.Date(2026, 3, 1+i%31, 0, 0, 0, 0, time.UTC),
    			Account: accounts[i%len(accounts)],
    			Amount:  Cents((i*104_729)%500_000 - 200_000),
    		}
    	}
    	return txns
    }

    func main() {
    	fmt.Print(Statement(sampleMonth(6)))
    	txns := sampleMonth(500)
    	fmt.Println("same as slowStatement:", Statement(txns) == slowStatement(txns))
    	fmt.Println("allocations per call: ", testing.AllocsPerRun(20, func() { Statement(txns) }))
    	fmt.Println("slowStatement allocs: ", testing.AllocsPerRun(20, func() { slowStatement(txns) }))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"math"
    	"strings"
    	"testing"
    	"time"
    )

    // firstDiff describes where two statements first differ.
    func firstDiff(got, want string) string {
    	g, w := strings.Split(got, "\n"), strings.Split(want, "\n")
    	for i := range max(len(g), len(w)) {
    		var gl, wl string
    		if i < len(g) {
    			gl = g[i]
    		}
    		if i < len(w) {
    			wl = w[i]
    		}
    		if gl != wl {
    			return fmt.Sprintf("line %d:\n  got:  %q\n  want: %q", i+1, gl, wl)
    		}
    	}
    	return "(no difference found)"
    }

    func day(d int) time.Time { return time.Date(2026, 3, d, 0, 0, 0, 0, time.UTC) }

    func TestStatementMatchesReference(t *testing.T) {
    	inputs := map[string][]Txn{
    		"no transactions": nil,
    		"a month":         sampleMonth(40),
    		"small amounts":   {{day(1), "cash", 5}, {day(2), "cash", -5}, {day(3), "cash", 0}, {day(4), "cash", 100}, {day(5), "cash", -99}},
    		"thousands":       {{day(1), "rent", 99_999}, {day(2), "rent", 100_000}, {day(3), "rent", 123_456_789}, {day(4), "rent", -100_000_000}},
    		"account widths":  {{day(1), "", 1}, {day(2), "exactly12chr", 2}, {day(3), "longer-than-twelve", 3}, {day(4), "a", 4}},
    		"huge amounts":    {{day(1), "vault", math.MaxInt64}, {day(2), "vault", 1}, {day(3), "vault", math.MinInt64}},
    		"most negative":   {{day(31), "debt", math.MinInt64}},
    		"different years": {{time.Date(1999, 12, 31, 0, 0, 0, 0, time.UTC), "y2k", -1}, {time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), "new", 1}},
    	}
    	for name, txns := range inputs {
    		got, want := Statement(txns), slowStatement(txns)
    		if got != want {
    			t.Errorf("%s: Statement differs from slowStatement at %s", name, firstDiff(got, want))
    		}
    	}
    }

    func TestStatementAllocations(t *testing.T) {
    	for _, n := range []int{500, 2000} {
    		txns := sampleMonth(n)
    		if Statement(txns) != slowStatement(txns) {
    			t.Fatalf("Statement(sampleMonth(%d)) differs from slowStatement; fix the output first", n)
    		}
    		allocs := testing.AllocsPerRun(10, func() { Statement(txns) })
    		if allocs > 2 {
    			t.Errorf("Statement(sampleMonth(%d)) allocates %.0f times per call, want at most 2", n, allocs)
    		}
    	}
    }
---

Ledgerly's monthly statement is generated for every customer at the end of
the month, and the profile says `slowStatement` spends most of its time in
the garbage collector. It's correct and readable, but for 500 transactions it
allocates about **7,500 times**: two `fmt.Sprintf` calls per line, a
`String()` call per amount with its own string surgery, and a `+=` that copies
the whole statement so far on every line.

Write `Statement(txns)`. It must return **exactly** the same text as
`slowStatement(txns)` for every input, and allocate **at most twice** per
call, as measured by `testing.AllocsPerRun`.

## The format

```text
2026-03-01  rent           -$2,000.00   -$2,000.00
2026-03-02  groceries        -$952.71   -$2,952.71
2026-03-05  subscriptions    $2,189.16      $472.90
6 transactions, closing balance -$1,290.65
```

Per transaction: the date as `YYYY-MM-DD`, two spaces, the account padded
with spaces to 12 bytes (`%-12s`; a longer account isn't cut, it just pushes
the line along, as `subscriptions` does), a space, the amount right-aligned in
12 bytes (`%12s`), a space, and the running balance right-aligned in 12
bytes. Amounts look like `Cents.String()`: `$1,234.56`, `-$0.05`. The last
line gives the count and the closing balance. Accounts are ASCII.

## How you're graded

- Your output is compared with `slowStatement` on a range of inputs: no
  transactions, a normal month, tiny and huge amounts (including
  `math.MaxInt64` and `math.MinInt64`, where the running balance wraps round
  the same way in both), empty, exact-width and long account names, and
  different years. A mismatch is reported with the first differing line.
- `testing.AllocsPerRun` must report at most 2 allocations per call for
  `sampleMonth(500)` **and** `sampleMonth(2000)`, so the buffer size has to
  scale with the input.

**Run** prints a short statement, checks it against `slowStatement` for 500
transactions, and prints the allocations per call for both versions.

## Constraints

- Standard library only. No `unsafe`, and no caching results between calls.
- Profile, don't guess: in a real project you'd confirm the win with
  `go test -bench Statement -benchmem` before and after.
