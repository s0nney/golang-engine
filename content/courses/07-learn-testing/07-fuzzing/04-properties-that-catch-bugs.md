---
title: 'Your Turn: Properties That Catch Bugs'
quiz:
  - question: |
      Why does property 1 only say "if `s` doesn't match the grammar,
      `parse` must fail", and not also "if it matches, `parse` must
      succeed"?
    options:
      - text: Because regular expressions can't be trusted
      - text: Because some grammatically valid amounts, like a hundred quadrillion dollars, don't fit in an int64 and must be rejected
        correct: true
      - text: Because the fuzzer only generates invalid strings
      - text: No reason; both directions should be checked
    explanation: |
      The grammar says nothing about size. `"100000000000000000"` matches
      it, but can't be represented in cents, so the correct behaviour is an
      error. A property has to be true of *correct* code for every input,
      or the fuzz test fails on the fix.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"regexp"
    	"strconv"
    	"strings"
    	"testing"
    )

    // checkParseAmount is the body of a fuzz target. It calls parse(s) and
    // reports a failure on t if any of these properties is broken:
    //
    //  1. If s doesn't match validAmount, parse must return an error.
    //  2. If parse succeeds, the result must be negative exactly when s
    //     starts with "-" and the result isn't zero.
    //  3. If parse succeeds with c, parse(Format(c)) must return c, nil.
    func checkParseAmount(t testing.TB, parse func(string) (Cents, error), s string) {
    	t.Helper()
    	// ?
    }

    // ParseAmount parses an amount like "12", "12.3", "12.34" or "-0.05".
    // Amounts beyond ±92233720368547758.07 don't fit and are errors.
    func ParseAmount(s string) (Cents, error) {
    	digits, neg := strings.CutPrefix(s, "-")
    	whole, frac, hasDot := strings.Cut(digits, ".")
    	if whole == "" || (hasDot && (frac == "" || len(frac) > 2)) {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	for len(frac) < 2 {
    		frac += "0"
    	}
    	d, err1 := strconv.ParseInt(whole, 10, 64)
    	c, err2 := strconv.ParseInt(frac, 10, 64)
    	if err1 != nil || err2 != nil {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	n := Cents(d*100 + c)
    	if neg {
    		n = -n
    	}
    	return n, nil
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    var ErrBadAmount = errors.New("ledgerly: bad amount")

    // validAmount is the documented grammar for amounts: an optional "-",
    // one or more digits, and optionally "." and one or two digits. It's slow
    // but obviously right, which makes it a good oracle.
    var validAmount = regexp.MustCompile(`^-?[0-9]+(\.[0-9]{1,2})?$`)

    // Format formats c as a plain decimal like "-12.34", for CSV files.
    // It's only used on amounts that ParseAmount accepted.
    func Format(c Cents) string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s%d.%02d", sign, n/100, n%100)
    }

    // spyTB is a pretend testing.TB so Run can show what checkParseAmount
    // reports without a real test. Unlike the real thing, its Fatal methods
    // don't stop the function.
    type spyTB struct {
    	testing.TB
    	failures []string
    }

    func (s *spyTB) Helper()                   {}
    func (s *spyTB) Log(a ...any)              {}
    func (s *spyTB) Logf(f string, a ...any)   {}
    func (s *spyTB) Fail()                     { s.failures = append(s.failures, "(Fail)") }
    func (s *spyTB) FailNow()                  { s.failures = append(s.failures, "(FailNow)") }
    func (s *spyTB) Error(a ...any)            { s.failures = append(s.failures, fmt.Sprint(a...)) }
    func (s *spyTB) Fatal(a ...any)            { s.failures = append(s.failures, fmt.Sprint(a...)) }
    func (s *spyTB) Errorf(f string, a ...any) { s.failures = append(s.failures, fmt.Sprintf(f, a...)) }
    func (s *spyTB) Fatalf(f string, a ...any) { s.failures = append(s.failures, fmt.Sprintf(f, a...)) }

    func main() {
    	for _, s := range []string{"12.34", "-0.05", "+0", "1.+5", "100000000000000000", "92233720368547758.07", "92233720368547758.08"} {
    		c, err := ParseAmount(s)
    		spy := &spyTB{}
    		checkParseAmount(spy, ParseAmount, s)
    		fmt.Printf("ParseAmount(%q) = %d, %v\n", s, c, err)
    		for _, f := range spy.failures {
    			fmt.Println("    check:", f)
    		}
    	}
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"math"
    	"regexp"
    	"strconv"
    	"strings"
    	"testing"
    )

    // checkParseAmount is the body of a fuzz target. It calls parse(s) and
    // reports a failure on t if any of these properties is broken:
    //
    //  1. If s doesn't match validAmount, parse must return an error.
    //  2. If parse succeeds, the result must be negative exactly when s
    //     starts with "-" and the result isn't zero.
    //  3. If parse succeeds with c, parse(Format(c)) must return c, nil.
    func checkParseAmount(t testing.TB, parse func(string) (Cents, error), s string) {
    	t.Helper()
    	c, err := parse(s)
    	if err != nil {
    		return
    	}
    	if !validAmount.MatchString(s) {
    		t.Fatalf("parse(%q) = %d, nil; want an error, it doesn't match the grammar", s, c)
    	}
    	if neg := strings.HasPrefix(s, "-"); (c < 0) != (neg && c != 0) {
    		t.Fatalf("parse(%q) = %d: wrong sign", s, c)
    	}
    	again, err := parse(Format(c))
    	if err != nil || again != c {
    		t.Fatalf("round trip: parse(%q) = %d, but parse(Format(%d)) = %d, %v", s, c, c, again, err)
    	}
    }

    // ParseAmount parses an amount like "12", "12.3", "12.34" or "-0.05".
    // Amounts beyond ±92233720368547758.07 don't fit and are errors.
    func ParseAmount(s string) (Cents, error) {
    	digits, neg := strings.CutPrefix(s, "-")
    	whole, frac, hasDot := strings.Cut(digits, ".")
    	if !allDigits(whole) || (hasDot && (len(frac) > 2 || !allDigits(frac))) {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	for len(frac) < 2 {
    		frac += "0"
    	}
    	d, err1 := strconv.ParseInt(whole, 10, 64)
    	c, err2 := strconv.ParseInt(frac, 10, 64)
    	if err1 != nil || err2 != nil || d > (math.MaxInt64-c)/100 {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	n := Cents(d*100 + c)
    	if neg {
    		n = -n
    	}
    	return n, nil
    }

    // allDigits reports whether s is non-empty and only contains 0-9.
    func allDigits(s string) bool {
    	if s == "" {
    		return false
    	}
    	for _, r := range s {
    		if r < '0' || r > '9' {
    			return false
    		}
    	}
    	return true
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    var ErrBadAmount = errors.New("ledgerly: bad amount")

    // validAmount is the documented grammar for amounts: an optional "-",
    // one or more digits, and optionally "." and one or two digits. It's slow
    // but obviously right, which makes it a good oracle.
    var validAmount = regexp.MustCompile(`^-?[0-9]+(\.[0-9]{1,2})?$`)

    // Format formats c as a plain decimal like "-12.34", for CSV files.
    // It's only used on amounts that ParseAmount accepted.
    func Format(c Cents) string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s%d.%02d", sign, n/100, n%100)
    }

    // spyTB is a pretend testing.TB so Run can show what checkParseAmount
    // reports without a real test. Unlike the real thing, its Fatal methods
    // don't stop the function.
    type spyTB struct {
    	testing.TB
    	failures []string
    }

    func (s *spyTB) Helper()                   {}
    func (s *spyTB) Log(a ...any)              {}
    func (s *spyTB) Logf(f string, a ...any)   {}
    func (s *spyTB) Fail()                     { s.failures = append(s.failures, "(Fail)") }
    func (s *spyTB) FailNow()                  { s.failures = append(s.failures, "(FailNow)") }
    func (s *spyTB) Error(a ...any)            { s.failures = append(s.failures, fmt.Sprint(a...)) }
    func (s *spyTB) Fatal(a ...any)            { s.failures = append(s.failures, fmt.Sprint(a...)) }
    func (s *spyTB) Errorf(f string, a ...any) { s.failures = append(s.failures, fmt.Sprintf(f, a...)) }
    func (s *spyTB) Fatalf(f string, a ...any) { s.failures = append(s.failures, fmt.Sprintf(f, a...)) }

    func main() {
    	for _, s := range []string{"12.34", "-0.05", "+0", "1.+5", "100000000000000000", "92233720368547758.07", "92233720368547758.08"} {
    		c, err := ParseAmount(s)
    		spy := &spyTB{}
    		checkParseAmount(spy, ParseAmount, s)
    		fmt.Printf("ParseAmount(%q) = %d, %v\n", s, c, err)
    		for _, f := range spy.failures {
    			fmt.Println("    check:", f)
    		}
    	}
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"runtime"
    	"strings"
    	"testing"
    )

    // FuzzParseAmount runs your checkParseAmount against the real ParseAmount.
    // The grader only runs the seeds; locally you'd fuzz it with
    // go test -fuzz FuzzParseAmount.
    func FuzzParseAmount(f *testing.F) {
    	for _, s := range []string{
    		"12.34", "12", "12.3", "-0.05", "0", "-0", "", ".50", "12.", "12.345",
    		"+0", "+12.50", "1.+5", "-+5", "--5", " 5", "1_000",
    		"100000000000000000", "92233720368547758.07", "-92233720368547758.07",
    		"92233720368547758.08", "99999999999999999999", "184467440737095516.16",
    	} {
    		f.Add(s)
    	}
    	f.Fuzz(func(t *testing.T, s string) {
    		checkParseAmount(t, ParseAmount, s)
    	})
    }

    func TestParseAmountFixed(t *testing.T) {
    	for _, s := range []string{"+0", "+12.50", "1.+5", "-+5", "100000000000000000", "92233720368547758.08", "-92233720368547758.08", "184467440737095516.16"} {
    		if c, err := ParseAmount(s); !errors.Is(err, ErrBadAmount) {
    			t.Errorf("ParseAmount(%q) = %d, %v; want an error wrapping ErrBadAmount", s, c, err)
    		}
    	}
    	for s, want := range map[string]Cents{
    		"12.34": 1234, "12.3": 1230, "-0.05": -5, "0": 0,
    		"92233720368547758.07": 9223372036854775807, "-92233720368547758.07": -9223372036854775807,
    	} {
    		if c, err := ParseAmount(s); err != nil || c != want {
    			t.Errorf("ParseAmount(%q) = %d, %v; want %d, nil", s, c, err, want)
    		}
    	}
    }

    // recTB records failures. Fatal stops the goroutine like the real thing.
    type recTB struct {
    	testing.TB
    	failed   bool
    	panicked string
    }

    func (r *recTB) Helper()               {}
    func (r *recTB) Log(...any)            {}
    func (r *recTB) Logf(string, ...any)   {}
    func (r *recTB) Fail()                 { r.failed = true }
    func (r *recTB) Error(...any)          { r.failed = true }
    func (r *recTB) Errorf(string, ...any) { r.failed = true }
    func (r *recTB) FailNow()              { r.failed = true; runtime.Goexit() }
    func (r *recTB) Fatal(...any)          { r.failed = true; runtime.Goexit() }
    func (r *recTB) Fatalf(string, ...any) { r.failed = true; runtime.Goexit() }
    func (r *recTB) Skip(...any)           { runtime.Goexit() }
    func (r *recTB) Skipf(string, ...any)  { runtime.Goexit() }
    func (r *recTB) SkipNow()              { runtime.Goexit() }
    func (r *recTB) Failed() bool          { return r.failed }

    func runCheck(parse func(string) (Cents, error), s string) *recTB {
    	r := &recTB{}
    	done := make(chan struct{})
    	go func() {
    		defer close(done)
    		defer func() {
    			if p := recover(); p != nil {
    				r.panicked = fmt.Sprint(p)
    			}
    		}()
    		checkParseAmount(r, parse, s)
    	}()
    	<-done
    	return r
    }

    // A correct parser for the tests below: the solution's rules.
    func goodParse(s string) (Cents, error) {
    	if !validAmount.MatchString(s) || len(s) > 18 {
    		return 0, ErrBadAmount
    	}
    	neg := strings.HasPrefix(s, "-")
    	var n int64
    	digits := 0
    	for _, r := range strings.TrimPrefix(s, "-") {
    		if r == '.' {
    			continue
    		}
    		n = n*10 + int64(r-'0')
    		digits++
    	}
    	if dot := strings.IndexByte(s, '.'); dot < 0 {
    		n *= 100
    	} else if len(s)-dot-1 == 1 {
    		n *= 10
    	}
    	if neg {
    		n = -n
    	}
    	return Cents(n), nil
    }

    func TestCheckAcceptsCorrectParser(t *testing.T) {
    	for _, s := range []string{"12.34", "12", "12.3", "-0.05", "0", "-0", "", "+0", "abc", "12.345", "-1.50", "999999.99"} {
    		if r := runCheck(goodParse, s); r.panicked != "" {
    			t.Errorf("checkParseAmount(t, correctParser, %q) panicked: %s", s, r.panicked)
    		} else if r.failed {
    			t.Errorf("checkParseAmount(t, correctParser, %q) reported a failure, but the parser is correct", s)
    		}
    	}
    }

    func TestCheckCatchesBugs(t *testing.T) {
    	bugs := []struct {
    		name  string
    		input string
    		parse func(string) (Cents, error)
    	}{
    		{"accepts a leading +", "+5", func(s string) (Cents, error) {
    			if rest, ok := strings.CutPrefix(s, "+"); ok {
    				return goodParse(rest)
    			}
    			return goodParse(s)
    		}},
    		{"accepts letters (property 1)", "12x", func(s string) (Cents, error) {
    			return goodParse(strings.TrimSuffix(s, "x"))
    		}},
    		{"drops the minus sign (property 2)", "-0.05", func(s string) (Cents, error) {
    			c, err := goodParse(s)
    			if c < 0 {
    				c = -c
    			}
    			return c, err
    		}},
    		{"returns a negative for a positive amount (property 2)", "12.34", func(s string) (Cents, error) {
    			c, err := goodParse(s)
    			return -c, err
    		}},
    		{"swaps the meanings of \"12.3\" and \"12.03\" (property 3)", "12.3", func(s string) (Cents, error) {
    			switch s {
    			case "12.3":
    				return 1203, nil
    			case "12.03":
    				return 1230, nil
    			}
    			return goodParse(s)
    		}},
    		{"fails to re-parse its own formatted output (property 3)", "7", func(s string) (Cents, error) {
    			if strings.Contains(s, ".") {
    				return 0, ErrBadAmount
    			}
    			return goodParse(s)
    		}},
    	}
    	for _, b := range bugs {
    		r := runCheck(b.parse, b.input)
    		if r.panicked != "" {
    			t.Errorf("checkParseAmount with a parser that %s, on %q, panicked: %s", b.name, b.input, r.panicked)
    		} else if !r.failed {
    			t.Errorf("checkParseAmount(t, parse, %q) didn't report a failure for a parser that %s", b.input, b.name)
    		}
    	}
    }
---

This chapter's fuzzing found two bugs in `ParseAmount`. Now you'll do the whole job yourself: write the properties, then fix the parser so it satisfies them.

The grader can't run a real fuzzing session (it's open-ended, and your Submit has a few seconds). So it does what a plain `go test` does with a fuzz test: it runs `FuzzParseAmount` over a seed corpus that includes every input the fuzzer found. The fuzz target is a one-liner that calls *your* function:

```go
f.Fuzz(func(t *testing.T, s string) {
	checkParseAmount(t, ParseAmount, s)
})
```

## Part 1: the properties

Write `checkParseAmount(t, parse, s)`. It calls `parse(s)` and uses `t.Fatalf` (or `t.Errorf`) to report any broken property:

1. **Grammar oracle.** If `s` doesn't match `validAmount`, `parse` must return an error.
2. **Sign.** If `parse` succeeds with `c`, then `c < 0` must be true exactly when `s` starts with `"-"` and `c != 0`.
3. **Round trip.** If `parse` succeeds with `c`, then `parse(Format(c))` must return `c` and no error.

Use the `parse` argument, not `ParseAmount` directly. That's how the grader checks your properties are strong enough: it passes in a correct parser, which must produce no failures, and six broken ones, which must each produce one.

## Part 2: fix ParseAmount

Once your properties work, the seed corpus will show that the real `ParseAmount` breaks them:

- **Signs in the wrong place.** `strconv.ParseInt` accepts a leading `+`, so `"+0"`, `"+12.50"`, `"1.+5"` and `"-+5"` are all accepted. Check that the dollars and the cents contain only the digits 0 to 9 before parsing them.
- **Overflow.** `d*100 + c` silently wraps round for huge amounts. Return an error wrapping `ErrBadAmount` if the result wouldn't fit. Ledgerly allows amounts from `-92233720368547758.07` to `92233720368547758.07` (`math.MaxInt64` cents), so every valid amount can be negated safely; `-92233720368547758.08` must be rejected too, even though `math.MinInt64` would hold it. The check `d > (math.MaxInt64-c)/100` (on the magnitude, before applying the sign) tells you it's out of range, and you'll need to import `math`.

**Run** parses a few interesting inputs and shows what your `checkParseAmount` reports for each against the current `ParseAmount`. Before you fix anything, your check should complain about `"+0"`, `"1.+5"` and the overflows. After the fix, it should be silent.
