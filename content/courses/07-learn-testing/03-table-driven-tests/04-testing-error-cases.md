---
title: Testing the Error Cases
quiz:
  - question: |
      `ImportCSV` returns `fmt.Errorf("line %d: %w", n, ErrBadAmount)`.
      Which check is the most robust?
    options:
      - text: '`err == ErrBadAmount`'
      - text: '`err.Error() == "line 3: ledgerly: bad amount"`'
      - text: '`errors.Is(err, ErrBadAmount)`'
        correct: true
      - text: '`strings.Contains(err.Error(), "bad")`'
    explanation: |
      The error is *wrapped*, so `==` fails: `err` is the wrapper, not the
      sentinel. `errors.Is` walks the chain of wrapped errors. Comparing
      message strings works until someone rewords the message, and then the
      test breaks for no good reason.
  - question: |
      What's wrong with this table case check?

      ```go
      got, err := ParseAmount(tt.in)
      if (err != nil) != tt.wantErr {
          t.Errorf("ParseAmount(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
      }
      if got != tt.want {
          t.Errorf("ParseAmount(%q) = %v, want %v", tt.in, got, tt.want)
      }
      ```
    options:
      - text: Nothing, it's the standard pattern
      - text: It only checks that *some* error happened, not which one, and it still compares `got` when an error was expected
        correct: true
      - text: '`wantErr` can''t be a bool'
      - text: It should use `t.Fatalf` everywhere
    explanation: |
      A `wantErr bool` passes if the function fails for a completely
      different reason than the one the case is about. Checking a specific
      error with `errors.Is` is stronger, and when an error is expected there
      is usually nothing meaningful in `got`, so return after checking it.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strconv"
    	"strings"
    )

    // ErrFieldCount means a line didn't have exactly three fields.
    var ErrFieldCount = errors.New("ledgerly: wrong number of fields")

    // LineError reports which line of an import failed, and why.
    type LineError struct {
    	Line int   // 1-based line number
    	Err  error // what went wrong on that line
    }

    // Error formats e as "line 2: <the error>".
    func (e *LineError) Error() string {
    	// ?
    	return "?"
    }

    // ? Unwrap

    // ParseLine parses one "date,account,amount" line.
    func ParseLine(line string) (Transaction, error) {
    	fields := strings.Split(line, ",")
    	if len(fields) != 3 {
    		return Transaction{}, fmt.Errorf("want 3 fields, got %d", len(fields))
    	}
    	amount, err := ParseAmount(fields[2])
    	if err != nil {
    		return Transaction{}, err
    	}
    	return Transaction{Date: fields[0], Account: fields[1], Amount: amount}, nil
    }

    // ImportLines parses every line. If one fails, it returns a *LineError
    // holding the line number (counting from 1) and ParseLine's error.
    func ImportLines(lines []string) ([]Transaction, error) {
    	var txns []Transaction
    	for i, line := range lines {
    		tx, err := ParseLine(line)
    		if err != nil {
    			return nil, fmt.Errorf("line %d: %v", i+1, err)
    		}
    		txns = append(txns, tx)
    	}
    	return txns, nil
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    type Transaction struct {
    	Date    string
    	Account string
    	Amount  Cents
    }

    var ErrBadAmount = errors.New("ledgerly: bad amount")

    // ParseAmount parses "12", "12.3", "12.34" or "-0.05" into cents.
    func ParseAmount(s string) (Cents, error) {
    	digits, neg := strings.CutPrefix(s, "-")
    	whole, frac, hasDot := strings.Cut(digits, ".")
    	if whole == "" || (hasDot && (frac == "" || len(frac) > 2)) {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	for len(frac) < 2 {
    		frac += "0"
    	}
    	d, err1 := strconv.ParseUint(whole, 10, 40)
    	c, err2 := strconv.ParseUint(frac, 10, 8)
    	if err1 != nil || err2 != nil {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	n := Cents(d*100 + c)
    	if neg {
    		n = -n
    	}
    	return n, nil
    }

    func main() {
    	for _, lines := range [][]string{
    		{"2026-03-01,rent,-1200.00", "2026-03-02,cash,12.345"},
    		{"2026-03-01,rent,-1200.00", "2026-03-02,cash,50", "2026-03-03,oops"},
    	} {
    		_, err := ImportLines(lines)
    		lineErr, isLineErr := errors.AsType[*LineError](err)
    		fmt.Println("error:          ", err)
    		fmt.Println("is a *LineError:", isLineErr, lineErr)
    		fmt.Println("ErrBadAmount:   ", errors.Is(err, ErrBadAmount))
    		fmt.Println("ErrFieldCount:  ", errors.Is(err, ErrFieldCount))
    		fmt.Println()
    	}
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strconv"
    	"strings"
    )

    // ErrFieldCount means a line didn't have exactly three fields.
    var ErrFieldCount = errors.New("ledgerly: wrong number of fields")

    // LineError reports which line of an import failed, and why.
    type LineError struct {
    	Line int   // 1-based line number
    	Err  error // what went wrong on that line
    }

    // Error formats e as "line 2: <the error>".
    func (e *LineError) Error() string {
    	return fmt.Sprintf("line %d: %v", e.Line, e.Err)
    }

    // Unwrap returns the underlying error, so errors.Is and errors.AsType can
    // see through a *LineError.
    func (e *LineError) Unwrap() error { return e.Err }

    // ParseLine parses one "date,account,amount" line.
    func ParseLine(line string) (Transaction, error) {
    	fields := strings.Split(line, ",")
    	if len(fields) != 3 {
    		return Transaction{}, fmt.Errorf("%w: want 3, got %d", ErrFieldCount, len(fields))
    	}
    	amount, err := ParseAmount(fields[2])
    	if err != nil {
    		return Transaction{}, err
    	}
    	return Transaction{Date: fields[0], Account: fields[1], Amount: amount}, nil
    }

    // ImportLines parses every line. If one fails, it returns a *LineError
    // holding the line number (counting from 1) and ParseLine's error.
    func ImportLines(lines []string) ([]Transaction, error) {
    	var txns []Transaction
    	for i, line := range lines {
    		tx, err := ParseLine(line)
    		if err != nil {
    			return nil, &LineError{Line: i + 1, Err: err}
    		}
    		txns = append(txns, tx)
    	}
    	return txns, nil
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    type Transaction struct {
    	Date    string
    	Account string
    	Amount  Cents
    }

    var ErrBadAmount = errors.New("ledgerly: bad amount")

    // ParseAmount parses "12", "12.3", "12.34" or "-0.05" into cents.
    func ParseAmount(s string) (Cents, error) {
    	digits, neg := strings.CutPrefix(s, "-")
    	whole, frac, hasDot := strings.Cut(digits, ".")
    	if whole == "" || (hasDot && (frac == "" || len(frac) > 2)) {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	for len(frac) < 2 {
    		frac += "0"
    	}
    	d, err1 := strconv.ParseUint(whole, 10, 40)
    	c, err2 := strconv.ParseUint(frac, 10, 8)
    	if err1 != nil || err2 != nil {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	n := Cents(d*100 + c)
    	if neg {
    		n = -n
    	}
    	return n, nil
    }

    func main() {
    	for _, lines := range [][]string{
    		{"2026-03-01,rent,-1200.00", "2026-03-02,cash,12.345"},
    		{"2026-03-01,rent,-1200.00", "2026-03-02,cash,50", "2026-03-03,oops"},
    	} {
    		_, err := ImportLines(lines)
    		lineErr, isLineErr := errors.AsType[*LineError](err)
    		fmt.Println("error:          ", err)
    		fmt.Println("is a *LineError:", isLineErr, lineErr)
    		fmt.Println("ErrBadAmount:   ", errors.Is(err, ErrBadAmount))
    		fmt.Println("ErrFieldCount:  ", errors.Is(err, ErrFieldCount))
    		fmt.Println()
    	}
    }
  tests: |
    package main

    import (
    	"errors"
    	"testing"
    )

    var goodLines = []string{"2026-03-01,rent,-1200.00", "2026-03-02,cash,50"}

    func TestImportLinesOK(t *testing.T) {
    	txns, err := ImportLines(goodLines)
    	if err != nil {
    		t.Fatalf("ImportLines(%q) unexpected error: %v", goodLines, err)
    	}
    	if len(txns) != 2 || txns[0].Amount != -120000 || txns[1].Account != "cash" {
    		t.Errorf("ImportLines(%q) = %+v, want the two transactions", goodLines, txns)
    	}
    }

    func TestLineErrorMethods(t *testing.T) {
    	inner := errors.New("disk on fire")
    	e := &LineError{Line: 7, Err: inner}
    	if got, want := e.Error(), "line 7: disk on fire"; got != want {
    		t.Errorf("(&LineError{Line: 7, Err: disk on fire}).Error() = %q, want %q", got, want)
    	}
    	if !errors.Is(e, inner) {
    		t.Error("errors.Is(&LineError{Err: inner}, inner) = false, want true: give *LineError an Unwrap() error method")
    	}
    }

    func TestImportLinesErrors(t *testing.T) {
    	for _, tc := range []struct {
    		name     string
    		lines    []string
    		wantLine int
    		wantErr  error
    		wantMsg  string
    	}{
    		{
    			name:     "bad amount",
    			lines:    []string{"2026-03-01,rent,-1200.00", "2026-03-02,cash,12.345"},
    			wantLine: 2,
    			wantErr:  ErrBadAmount,
    			wantMsg:  `line 2: ledgerly: bad amount: "12.345"`,
    		},
    		{
    			name:     "too few fields",
    			lines:    []string{"2026-03-01,rent,-1200.00", "2026-03-02,cash,50", "2026-03-03,oops"},
    			wantLine: 3,
    			wantErr:  ErrFieldCount,
    			wantMsg:  "line 3: ledgerly: wrong number of fields: want 3, got 2",
    		},
    		{
    			name:     "too many fields on the first line",
    			lines:    []string{"2026-03-01,rent,-1200.00,extra"},
    			wantLine: 1,
    			wantErr:  ErrFieldCount,
    			wantMsg:  "line 1: ledgerly: wrong number of fields: want 3, got 4",
    		},
    	} {
    		t.Run(tc.name, func(t *testing.T) {
    			_, err := ImportLines(tc.lines)
    			if err == nil {
    				t.Fatalf("ImportLines(%q) returned no error", tc.lines)
    			}
    			lineErr, ok := errors.AsType[*LineError](err)
    			if !ok {
    				t.Fatalf("ImportLines(%q) error %q is a %T, want a *LineError", tc.lines, err, err)
    			}
    			if lineErr.Line != tc.wantLine {
    				t.Errorf("LineError.Line = %d, want %d", lineErr.Line, tc.wantLine)
    			}
    			if !errors.Is(err, tc.wantErr) {
    				t.Errorf("ImportLines(%q) error = %q; errors.Is(err, %v) = false, want true", tc.lines, err, tc.wantErr)
    			}
    			if got := err.Error(); got != tc.wantMsg {
    				t.Errorf("ImportLines error message = %q, want %q", got, tc.wantMsg)
    			}
    		})
    	}
    }
---

The happy path is where bugs *aren't*. Bugs live in the inputs nobody thought about: empty strings, three decimal places, a CSV line with a missing column. A good table has at least as many failure cases as success cases, and checks each failure precisely.

## wantErr as an error, not a bool

You'll see a lot of tables with `wantErr bool`. It's better than nothing, but it only proves that *something* went wrong. A case meant to test "three decimal places is rejected" still passes if the function fails because of a typo in the test input.

Store the error you expect instead, and check it with `errors.Is`:

```go
func TestParseAmount(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    Cents
		wantErr error
	}{
		{name: "dollars and cents", in: "12.34", want: 1234},
		{name: "three decimal places", in: "12.345", wantErr: ErrBadAmount},
		{name: "empty", in: "", wantErr: ErrBadAmount},
		{name: "letters", in: "twelve", wantErr: ErrBadAmount},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAmount(tt.in)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ParseAmount(%q) error = %v, want %v", tt.in, err, tt.wantErr)
				}
				return // the error was right; nothing else to check
			}
			if err != nil {
				t.Fatalf("ParseAmount(%q) unexpected error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ParseAmount(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
```

Two details matter:

- **`errors.Is`, not `==`.** Ledgerly wraps its errors with context (`fmt.Errorf("%w: %q", ErrBadAmount, s)`), so the returned error is a wrapper around the sentinel. `errors.Is` unwraps the chain. And `errors.Is(nil, ErrBadAmount)` is simply `false`, so a missing error is caught too.
- **Return after a correct error.** When an error is expected, `got` is usually a zero value that nobody should rely on. Asserting on it just couples the test to an accident.

## Never compare error strings

```go
if err.Error() != `ledgerly: bad amount: "12.345"` { // don't
```

Error messages are for humans and change freely. A test pinned to the exact text breaks when someone improves the wording, which trains people to update tests blindly. Test the *identity* of an error (`errors.Is`) or its *type and fields* (next section). If the message really is part of the contract, checking that it `strings.Contains` the key detail (like the bad input) is the most you should do.

## Typed errors: errors.AsType

The CSV importer needs to say *where* things went wrong, so it returns a struct error:

```go
type LineError struct {
	Line int
	Err  error
}

func (e *LineError) Error() string { return fmt.Sprintf("line %d: %v", e.Line, e.Err) }
func (e *LineError) Unwrap() error { return e.Err }
```

A test wants to check the line number. `errors.AsType` (Go 1.26) finds the first error of a given type in the chain and returns it, typed:

```go
_, err := ImportCSV(strings.NewReader("date,account,amount\n2026-03-01,rent,12.345\n"))

lineErr, ok := errors.AsType[*LineError](err)
if !ok {
	t.Fatalf("ImportCSV error = %v, want a *LineError", err)
}
if lineErr.Line != 2 {
	t.Errorf("LineError.Line = %d, want 2", lineErr.Line)
}
if !errors.Is(err, ErrBadAmount) {
	t.Errorf("ImportCSV error = %v, want it to wrap ErrBadAmount", err)
}
```

Because `LineError` has an `Unwrap` method, `errors.Is` sees through it to `ErrBadAmount`, so one error answers both "where?" and "what?". Before Go 1.26 you'd write `var lineErr *LineError; if errors.As(err, &lineErr)`. `AsType` does the same job without the extra variable, and `go fix` will rewrite the old form for you (chapter 9).

## Your turn: errors a test can inspect

Ledgerly's line importer has errors that only a human can read. Run it: the messages look fine, but `errors.AsType` finds no `*LineError`, and `errors.Is` can't find `ErrBadAmount` or `ErrFieldCount` in the chain. A caller (or a test) can't tell *which* line failed or *why* without parsing strings.

Fix it in three places:

1. **`LineError`.** Implement `Error` so it returns `line 2: <the wrapped error>`, and add an `Unwrap() error` method that returns `e.Err`.
2. **`ParseLine`.** When a line doesn't have three fields, wrap the `ErrFieldCount` sentinel with `%w`, keeping the detail: `ledgerly: wrong number of fields: want 3, got 2`.
3. **`ImportLines`.** Instead of flattening the error into a string with `%v`, return `&LineError{Line: ..., Err: err}`, with lines counted from 1.

The grader uses exactly the checks from this lesson: `errors.AsType[*LineError]` for the line number, and `errors.Is` for the cause.

## Which failure cases?

For any parser or validator, run down this list and add a case for each that applies:

- the **empty** input, and input that's only whitespace
- each **boundary**: one decimal digit, two, three; zero; the largest value
- **signs**: negative, and negative values smaller than one unit (like `-0.05`)
- **missing parts**: `".50"`, `"12."`, a CSV line with too few fields
- **junk**: letters, two dots, a stray `+`

That list is exactly what the next exercise asks you to turn into a table.
