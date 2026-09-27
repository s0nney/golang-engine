---
title: 'Your Turn: A Table That Catches Bugs'
quiz:
  - question: A test table passes against your code. What does that tell you about how good the table is?
    options:
      - text: It's complete, since every case passes
      - text: Very little on its own; a table that passes against broken code too would pass just the same
        correct: true
      - text: The code has no bugs
      - text: It has 100% coverage
    explanation: |
      Passing only shows that the code and the table agree. The real question
      is whether the table would *fail* if the code were wrong. Running it
      against deliberately broken versions (mutants) answers that, and it's
      exactly what this exercise's grader does.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strconv"
    	"strings"
    )

    // parseCase is one row of the ParseAmount test table. For valid input set
    // want; for invalid input set wantErr to the error ParseAmount must return
    // (checked with errors.Is) and leave want alone.
    type parseCase struct {
    	name    string
    	in      string
    	want    Cents
    	wantErr error
    }

    // parseCases is the table. The grader runs it against ParseAmount and
    // against seven broken versions of it. Add cases until every broken
    // version fails at least one of them.
    var parseCases = []parseCase{
    	{name: "dollars and cents", in: "12.34", want: 1234},
    	{name: "whole dollars", in: "12", want: 1200},
    	// ?
    }

    // ---- Ledgerly code (already done, and correct) ----

    type Cents int64

    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    var ErrBadAmount = errors.New("ledgerly: bad amount")

    // ParseAmount parses an optional "-", one or more digits, and optionally
    // a "." followed by one or two digits: "12", "12.3", "12.34", "-0.05".
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
    	if err1 != nil || err2 != nil {
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

    func main() {
    	for _, tc := range parseCases {
    		got, err := ParseAmount(tc.in)
    		fmt.Printf("%-24s ParseAmount(%q) = %v, %v\n", tc.name, tc.in, got, err)
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

    // parseCase is one row of the ParseAmount test table. For valid input set
    // want; for invalid input set wantErr to the error ParseAmount must return
    // (checked with errors.Is) and leave want alone.
    type parseCase struct {
    	name    string
    	in      string
    	want    Cents
    	wantErr error
    }

    // parseCases is the table. The grader runs it against ParseAmount and
    // against seven broken versions of it.
    var parseCases = []parseCase{
    	{name: "dollars and cents", in: "12.34", want: 1234},
    	{name: "whole dollars", in: "12", want: 1200},
    	{name: "one decimal digit", in: "12.3", want: 1230},
    	{name: "negative dollars", in: "-1.50", want: -150},
    	{name: "negative under a dollar", in: "-0.05", want: -5},
    	{name: "three decimal places", in: "12.345", wantErr: ErrBadAmount},
    	{name: "empty", in: "", wantErr: ErrBadAmount},
    	{name: "no dollars", in: ".50", wantErr: ErrBadAmount},
    	{name: "trailing dot", in: "12.", wantErr: ErrBadAmount},
    	{name: "letters", in: "twelve", wantErr: ErrBadAmount},
    }

    // ---- Ledgerly code (already done, and correct) ----

    type Cents int64

    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    var ErrBadAmount = errors.New("ledgerly: bad amount")

    // ParseAmount parses an optional "-", one or more digits, and optionally
    // a "." followed by one or two digits: "12", "12.3", "12.34", "-0.05".
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
    	if err1 != nil || err2 != nil {
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

    func main() {
    	for _, tc := range parseCases {
    		got, err := ParseAmount(tc.in)
    		fmt.Printf("%-24s ParseAmount(%q) = %v, %v\n", tc.name, tc.in, got, err)
    	}
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strings"
    	"testing"
    )

    type parseFunc func(string) (Cents, error)

    // check runs one case against parse and describes the failure, or returns "".
    func check(parse parseFunc, tc parseCase) string {
    	got, err := parse(tc.in)
    	if tc.wantErr != nil {
    		if !errors.Is(err, tc.wantErr) {
    			return fmt.Sprintf("ParseAmount(%q) = %v, %v; want error %v", tc.in, got, err, tc.wantErr)
    		}
    		return ""
    	}
    	if err != nil {
    		return fmt.Sprintf("ParseAmount(%q) unexpected error: %v", tc.in, err)
    	}
    	if got != tc.want {
    		return fmt.Sprintf("ParseAmount(%q) = %v, want %v", tc.in, got, tc.want)
    	}
    	return ""
    }

    func TestTableIsWellFormed(t *testing.T) {
    	seen := map[string]bool{}
    	for i, tc := range parseCases {
    		if tc.name == "" {
    			t.Errorf("case %d (in: %q) has no name", i, tc.in)
    		}
    		if seen[tc.name] {
    			t.Errorf("two cases are named %q; names must be unique", tc.name)
    		}
    		seen[tc.name] = true
    		if tc.wantErr != nil && tc.want != 0 {
    			t.Errorf("case %q sets both want and wantErr; set only one", tc.name)
    		}
    	}
    }

    // TestTableAgainstRealParser: every case must pass against the correct
    // ParseAmount, or the table's expectations are wrong.
    func TestTableAgainstRealParser(t *testing.T) {
    	for _, tc := range parseCases {
    		t.Run(tc.name, func(t *testing.T) {
    			if msg := check(ParseAmount, tc); msg != "" {
    				t.Errorf("your case is wrong (the real ParseAmount is correct): %s", msg)
    			}
    		})
    	}
    }

    // Seven broken parsers, each with one bug.
    var mutants = []struct {
    	bug   string
    	parse parseFunc
    }{
    	{"accepts three decimal places, \"12.345\" gives $12.34", func(s string) (Cents, error) {
    		if w, f, ok := strings.Cut(s, "."); ok && len(f) > 2 {
    			return ParseAmount(w + "." + f[:2])
    		}
    		return ParseAmount(s)
    	}},
    	{"reads one decimal digit as cents, \"12.3\" gives $12.03", func(s string) (Cents, error) {
    		if w, f, ok := strings.Cut(s, "."); ok && len(f) == 1 {
    			return ParseAmount(w + ".0" + f)
    		}
    		return ParseAmount(s)
    	}},
    	{"loses the minus sign under a dollar, \"-0.05\" gives $0.05", func(s string) (Cents, error) {
    		c, err := ParseAmount(s)
    		if err == nil && c < 0 && c > -100 {
    			c = -c
    		}
    		return c, err
    	}},
    	{"only negates the dollars, \"-1.50\" gives -$0.50", func(s string) (Cents, error) {
    		c, err := ParseAmount(s)
    		if err == nil && c < 0 {
    			d := -c
    			c = -(d/100)*100 + d%100
    		}
    		return c, err
    	}},
    	{"accepts the empty string as $0.00", func(s string) (Cents, error) {
    		if s == "" {
    			return 0, nil
    		}
    		return ParseAmount(s)
    	}},
    	{"accepts a missing dollar part, \".50\" gives $0.50", func(s string) (Cents, error) {
    		if rest, ok := strings.CutPrefix(s, "."); ok {
    			return ParseAmount("0." + rest)
    		}
    		return ParseAmount(s)
    	}},
    	{"accepts a trailing dot, \"12.\" gives $12.00", func(s string) (Cents, error) {
    		if rest, ok := strings.CutSuffix(s, "."); ok && rest != "" {
    			return ParseAmount(rest)
    		}
    		return ParseAmount(s)
    	}},
    }

    func TestTableCatchesBugs(t *testing.T) {
    	for _, m := range mutants {
    		caught := false
    		for _, tc := range parseCases {
    			if check(m.parse, tc) != "" {
    				caught = true
    				break
    			}
    		}
    		if !caught {
    			t.Errorf("no case catches a ParseAmount that %s", m.bug)
    		}
    	}
    }

    // A parser that returns the right results but the wrong kind of error.
    func TestTableChecksWhichError(t *testing.T) {
    	plainErrors := func(s string) (Cents, error) {
    		c, err := ParseAmount(s)
    		if err != nil {
    			return 0, errors.New("bad amount")
    		}
    		return c, nil
    	}
    	for _, tc := range parseCases {
    		if check(plainErrors, tc) != "" {
    			return
    		}
    	}
    	t.Error("no case catches a ParseAmount that returns errors.New(\"bad amount\") instead of wrapping ErrBadAmount (add a case with wantErr: ErrBadAmount)")
    }
---

This time you're not writing the code, you're writing the **tests**. `ParseAmount` is finished and correct. Your job is to write the table that proves it.

## How you're graded

How do you grade a test table? You run it against code you *know* is broken, and check that it fails. That idea is called **mutation testing**: make small deliberate bugs ("mutants") and see whether your tests kill them. A mutant that survives points at a missing case.

The grader runs your `parseCases` with this loop, the same shape as the last lesson's `wantErr` table:

```go
got, err := parse(tc.in)
if tc.wantErr != nil {
	// pass only if errors.Is(err, tc.wantErr)
} else {
	// pass only if err == nil && got == tc.want
}
```

It checks three things:

1. **Every case passes against the real `ParseAmount`.** If one doesn't, your expectation is wrong, not the parser.
2. **Every mutant fails at least one case.** There are seven broken parsers, each with one classic bug from the checklist at the end of the last lesson.
3. **At least one case uses `wantErr: ErrBadAmount`**, so a parser that returns the wrong kind of error gets caught.

Case names must be non-empty and unique, and a case sets either `want` or `wantErr`, never both.

## The rules ParseAmount follows

- An optional leading `-`.
- One or more digits for the dollars. Not zero digits: `".50"` is an error.
- Optionally a `.` followed by **one or two** digits. `"12.3"` means $12.30. A trailing dot (`"12."`) or three digits (`"12.345"`) is an error.
- Nothing else. The empty string is an error.
- Every error wraps `ErrBadAmount`.

## Your task

Add cases to `parseCases` until the tests pass. Think about each rule above and write the input that would break a sloppy implementation of it. Don't forget the negative amount that's *less than one dollar*: the obvious implementation gets that one wrong.

**Run** prints what the real `ParseAmount` returns for each of your cases, which helps you get the expectations right. When you Submit, a surviving mutant is reported by the bug it contains.
