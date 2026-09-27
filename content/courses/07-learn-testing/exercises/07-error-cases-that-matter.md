---
title: Error Cases That Matter
difficulty: medium
after: table-driven-tests
hints:
  - 'Go through the contract rule by rule. For each rule, write one input that breaks **only** that rule, with `wantErr` set to that rule''s sentinel. A case that checks the *which* error, not just *an* error, catches a parser that reports the wrong problem.'
  - 'Some bugs only show on valid input: a parser that loses the minus sign still succeeds. Include a negative amount with the exact `want` Txn.'
  - 'The sneaky ones: `"2026-02-30"` looks like a date but isn''t one, `" rent"` has a space the contract forbids, a zero amount is rejected, and a fourth field must be an error even if the first three are fine.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strconv"
    	"strings"
    	"time"
    )

    type txnCase struct {
    	name    string
    	line    string
    	want    Txn   // for valid lines
    	wantErr error // for invalid lines: the sentinel ParseTxnLine must wrap
    }

    var txnCases = []txnCase{}

    // ---- Ledgerly code (already done, and correct) ----

    type Cents int64

    type Txn struct {
    	Date    time.Time
    	Account string
    	Amount  Cents
    }

    var (
    	ErrFieldCount = errors.New("ledgerly: want 3 fields")
    	ErrBadDate    = errors.New("ledgerly: bad date")
    	ErrBadAccount = errors.New("ledgerly: bad account")
    	ErrBadAmount  = errors.New("ledgerly: bad amount")
    )

    // ParseTxnLine parses one CSV line "date,account,amount":
    //   - exactly three comma-separated fields, or ErrFieldCount;
    //   - a real calendar date in YYYY-MM-DD form, or ErrBadDate;
    //   - a non-empty account with no leading or trailing spaces, or ErrBadAccount;
    //   - a non-zero whole number of cents (may be negative), or ErrBadAmount.
    func ParseTxnLine(line string) (Txn, error) {
    	fields := strings.Split(line, ",")
    	if len(fields) != 3 {
    		return Txn{}, fmt.Errorf("%w, got %d in %q", ErrFieldCount, len(fields), line)
    	}
    	date, err := time.Parse(time.DateOnly, fields[0])
    	if err != nil {
    		return Txn{}, fmt.Errorf("%w: %q", ErrBadDate, fields[0])
    	}
    	account := fields[1]
    	if account == "" || strings.TrimSpace(account) != account {
    		return Txn{}, fmt.Errorf("%w: %q", ErrBadAccount, account)
    	}
    	n, err := strconv.ParseInt(fields[2], 10, 64)
    	if err != nil || n == 0 {
    		return Txn{}, fmt.Errorf("%w: %q", ErrBadAmount, fields[2])
    	}
    	return Txn{Date: date, Account: account, Amount: Cents(n)}, nil
    }

    // day is a shortcut for writing dates in your cases: day(2026, 3, 1).
    func day(y int, m time.Month, d int) time.Time {
    	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
    }

    func main() {
    	fmt.Println(len(txnCases), "cases")
    	for _, tc := range txnCases {
    		got, err := ParseTxnLine(tc.line)
    		fmt.Printf("%-20s ParseTxnLine(%q)\n    = %s %q %d, %v\n", tc.name, tc.line, got.Date.Format(time.DateOnly), got.Account, got.Amount, err)
    	}
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strconv"
    	"strings"
    	"time"
    )

    type txnCase struct {
    	name    string
    	line    string
    	want    Txn   // for valid lines
    	wantErr error // for invalid lines: the sentinel ParseTxnLine must wrap
    }

    var txnCases = []txnCase{
    	{name: "valid debit", line: "2026-03-01,rent,-120000", want: Txn{day(2026, 3, 1), "rent", -120000}},
    	{name: "valid credit", line: "2026-12-31,cash,5000", want: Txn{day(2026, 12, 31), "cash", 5000}},
    	{name: "too few fields", line: "2026-03-01,rent", wantErr: ErrFieldCount},
    	{name: "too many fields", line: "2026-03-01,rent,-500,extra", wantErr: ErrFieldCount},
    	{name: "empty line", line: "", wantErr: ErrFieldCount},
    	{name: "not a date", line: "March 1,rent,-500", wantErr: ErrBadDate},
    	{name: "impossible date", line: "2026-02-30,rent,-500", wantErr: ErrBadDate},
    	{name: "empty account", line: "2026-03-01,,-500", wantErr: ErrBadAccount},
    	{name: "padded account", line: "2026-03-01, rent,-500", wantErr: ErrBadAccount},
    	{name: "zero amount", line: "2026-03-01,rent,0", wantErr: ErrBadAmount},
    	{name: "dollar amount", line: "2026-03-01,rent,-12.50", wantErr: ErrBadAmount},
    }

    // ---- Ledgerly code (already done, and correct) ----

    type Cents int64

    type Txn struct {
    	Date    time.Time
    	Account string
    	Amount  Cents
    }

    var (
    	ErrFieldCount = errors.New("ledgerly: want 3 fields")
    	ErrBadDate    = errors.New("ledgerly: bad date")
    	ErrBadAccount = errors.New("ledgerly: bad account")
    	ErrBadAmount  = errors.New("ledgerly: bad amount")
    )

    // ParseTxnLine parses one CSV line "date,account,amount":
    //   - exactly three comma-separated fields, or ErrFieldCount;
    //   - a real calendar date in YYYY-MM-DD form, or ErrBadDate;
    //   - a non-empty account with no leading or trailing spaces, or ErrBadAccount;
    //   - a non-zero whole number of cents (may be negative), or ErrBadAmount.
    func ParseTxnLine(line string) (Txn, error) {
    	fields := strings.Split(line, ",")
    	if len(fields) != 3 {
    		return Txn{}, fmt.Errorf("%w, got %d in %q", ErrFieldCount, len(fields), line)
    	}
    	date, err := time.Parse(time.DateOnly, fields[0])
    	if err != nil {
    		return Txn{}, fmt.Errorf("%w: %q", ErrBadDate, fields[0])
    	}
    	account := fields[1]
    	if account == "" || strings.TrimSpace(account) != account {
    		return Txn{}, fmt.Errorf("%w: %q", ErrBadAccount, account)
    	}
    	n, err := strconv.ParseInt(fields[2], 10, 64)
    	if err != nil || n == 0 {
    		return Txn{}, fmt.Errorf("%w: %q", ErrBadAmount, fields[2])
    	}
    	return Txn{Date: date, Account: account, Amount: Cents(n)}, nil
    }

    // day is a shortcut for writing dates in your cases: day(2026, 3, 1).
    func day(y int, m time.Month, d int) time.Time {
    	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
    }

    func main() {
    	fmt.Println(len(txnCases), "cases")
    	for _, tc := range txnCases {
    		got, err := ParseTxnLine(tc.line)
    		fmt.Printf("%-20s ParseTxnLine(%q)\n    = %s %q %d, %v\n", tc.name, tc.line, got.Date.Format(time.DateOnly), got.Account, got.Amount, err)
    	}
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strings"
    	"testing"
    	"time"
    )

    type parseFunc func(string) (Txn, error)

    func check(parse parseFunc, tc txnCase) string {
    	got, err := parse(tc.line)
    	if tc.wantErr != nil {
    		if !errors.Is(err, tc.wantErr) {
    			return fmt.Sprintf("ParseTxnLine(%q) error = %v, want one wrapping %q", tc.line, err, tc.wantErr)
    		}
    		return ""
    	}
    	if err != nil {
    		return fmt.Sprintf("ParseTxnLine(%q) unexpected error: %v", tc.line, err)
    	}
    	if !got.Date.Equal(tc.want.Date) || got.Account != tc.want.Account || got.Amount != tc.want.Amount {
    		return fmt.Sprintf("ParseTxnLine(%q) = {%s %q %d}, want {%s %q %d}", tc.line,
    			got.Date.Format(time.DateOnly), got.Account, got.Amount,
    			tc.want.Date.Format(time.DateOnly), tc.want.Account, tc.want.Amount)
    	}
    	return ""
    }

    func TestTableIsWellFormed(t *testing.T) {
    	seen := map[string]bool{}
    	for i, tc := range txnCases {
    		if tc.name == "" {
    			t.Errorf("case %d (line %q) has no name", i, tc.line)
    		} else if seen[tc.name] {
    			t.Errorf("two cases are named %q; names must be unique", tc.name)
    		}
    		seen[tc.name] = true
    		if tc.wantErr != nil && tc.want != (Txn{}) {
    			t.Errorf("case %q sets both want and wantErr; set only one", tc.name)
    		}
    	}
    }

    func TestTableAgainstRealParser(t *testing.T) {
    	for _, tc := range txnCases {
    		if msg := check(ParseTxnLine, tc); msg != "" {
    			t.Errorf("case %q is wrong (the real ParseTxnLine is correct): %s", tc.name, msg)
    		}
    	}
    }

    // relabel returns the result of the real parser, but with err's sentinel
    // replaced when it matches from.
    func relabel(line string, from, to error) (Txn, error) {
    	txn, err := ParseTxnLine(line)
    	if errors.Is(err, from) {
    		return Txn{}, to
    	}
    	return txn, err
    }

    var mutants = []struct {
    	bug   string
    	parse parseFunc
    }{
    	{"ignores fields after the third", func(line string) (Txn, error) {
    		if f := strings.Split(line, ","); len(f) > 3 {
    			return ParseTxnLine(strings.Join(f[:3], ","))
    		}
    		return ParseTxnLine(line)
    	}},
    	{"reports a bad date as ErrBadAmount", func(line string) (Txn, error) {
    		return relabel(line, ErrBadDate, ErrBadAmount)
    	}},
    	{"reports a bad account as ErrFieldCount", func(line string) (Txn, error) {
    		return relabel(line, ErrBadAccount, ErrFieldCount)
    	}},
    	{"accepts impossible dates like 2026-02-30 (it only checks the digits)", func(line string) (Txn, error) {
    		f := strings.Split(line, ",")
    		if len(f) == 3 && len(f[0]) == 10 && f[0][4] == '-' && f[0][7] == '-' {
    			if _, err := time.Parse(time.DateOnly, f[0]); err != nil {
    				f[0] = f[0][:8] + "01"
    				txn, err := ParseTxnLine(strings.Join(f, ","))
    				return txn, err
    			}
    		}
    		return ParseTxnLine(line)
    	}},
    	{"trims spaces around the account instead of rejecting them", func(line string) (Txn, error) {
    		f := strings.Split(line, ",")
    		if len(f) == 3 {
    			f[1] = strings.TrimSpace(f[1])
    		}
    		return ParseTxnLine(strings.Join(f, ","))
    	}},
    	{"accepts a zero amount", func(line string) (Txn, error) {
    		f := strings.Split(line, ",")
    		if len(f) == 3 && f[2] == "0" {
    			txn, err := ParseTxnLine(strings.Join(f[:2], ",") + ",1")
    			txn.Amount = 0
    			return txn, err
    		}
    		return ParseTxnLine(line)
    	}},
    	{"drops the minus sign from amounts", func(line string) (Txn, error) {
    		txn, err := ParseTxnLine(line)
    		txn.Amount = max(txn.Amount, -txn.Amount)
    		return txn, err
    	}},
    }

    func TestTableCatchesBugs(t *testing.T) {
    	for _, m := range mutants {
    		caught := false
    		for _, tc := range txnCases {
    			if check(m.parse, tc) != "" {
    				caught = true
    				break
    			}
    		}
    		if !caught {
    			t.Errorf("no case catches a ParseTxnLine that %s", m.bug)
    		}
    	}
    }
---

Ledgerly's CSV importer calls `ParseTxnLine` for every line of a bank
statement. The function is correct, and it returns a **different sentinel
error for each kind of problem**, so the importer can tell the user exactly
what's wrong with line 212. Your job is to write the table that pins all of
that down.

Fill in `txnCases`. Each case has a unique, non-empty `name` and a `line`, and
sets either `want` (the exact `Txn` for a valid line) or `wantErr` (the
sentinel the error must wrap), never both. The grader checks each case like
this:

```go
got, err := parse(tc.line)
if tc.wantErr != nil {
	// pass only if errors.Is(err, tc.wantErr)
} else {
	// pass only if err == nil and got matches tc.want
}
```

Every case must pass against the real `ParseTxnLine`, and each of **seven
broken parsers** must fail at least one case.

## The contract

A line is `date,account,amount`:

| Rule | Error |
| --- | --- |
| exactly three comma-separated fields | `ErrFieldCount` |
| a real calendar date, `YYYY-MM-DD` | `ErrBadDate` |
| a non-empty account with no leading or trailing spaces | `ErrBadAccount` |
| a non-zero whole number of cents, possibly negative | `ErrBadAmount` |

## Examples

```go
{name: "valid debit", line: "2026-03-01,rent,-120000",
	want: Txn{day(2026, 3, 1), "rent", -120000}},
{name: "too few fields", line: "2026-03-01,rent", wantErr: ErrFieldCount},
```

`day(y, m, d)` builds a midnight-UTC date, which is what `time.Parse` returns
for `YYYY-MM-DD`.

**Run** prints what the real parser returns for each of your cases.

## Constraints

- An error case that only proves "some error happened" isn't enough. Some of
  the broken parsers fail on the right lines but return the **wrong
  sentinel**.
