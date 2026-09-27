---
title: Good Properties, and What Fuzzing Can't Find
quiz:
  - question: You have `Format(c Cents) string` and `ParseAmount(s string) (Cents, error)`. Which property makes a good fuzz target for them?
    options:
      - text: '`Format(c)` is not empty'
      - text: For every `c`, `ParseAmount(Format(c))` returns `c` and no error
        correct: true
      - text: '`Format(c)` equals `"12.34"`'
      - text: '`ParseAmount` never returns an error'
    explanation: |
      A round trip checks that two functions agree with each other for
      every input, without you knowing any expected output. It catches
      bugs in either function.
  - question: |
      `Cents(math.MinInt64).String()` returns garbage. Fuzzing an `int64`
      for 90 seconds (33 million inputs) didn't find it. Why not?
    options:
      - text: The fuzzer can't generate negative numbers
      - text: One bad value out of 2⁶⁴, with no branch leading towards it, is effectively impossible to hit at random
        correct: true
      - text: Fuzzing only works on strings
      - text: The fuzzer skips values that would fail
    explanation: |
      Coverage guidance helps the fuzzer get past `if` statements, but this
      bug has no special branch; it's an arithmetic edge. Put known boundary
      values like `math.MinInt64` and `math.MaxInt64` into the seed corpus.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"math"
    	"regexp"
    	"strconv"
    	"strings"
    	"testing"
    )

    // formatted matches a well-formed Ledgerly amount: "$0.05", "-$1,234.56".
    var formatted = regexp.MustCompile(`^-?\$[0-9]{1,3}(,[0-9]{3})*\.[0-9]{2}$`)

    // digitsOnly strips the decoration from a formatted amount, so
    // "-$1,234.56" becomes "-123456".
    var digitsOnly = strings.NewReplacer("$", "", ",", "", ".", "")

    // edgeSeeds are extra seeds for FuzzCentsString: the boundary values a
    // fuzzer is unlikely to stumble on by itself.
    var edgeSeeds = []int64{
    	// ?
    }

    // checkString reports a failure on t unless str(Cents(n)) is well-formed
    // (it matches formatted) and its digits parse back to n.
    func checkString(t testing.TB, str func(Cents) string, n int64) {
    	// ?
    }

    // String formats c like "$1,234.56" or "-$0.05".
    func (c Cents) String() string {
    	sign := ""
    	n := int64(c)
    	if n < 0 {
    		sign = "-"
    		n = -n
    	}
    	return fmt.Sprintf("%s$%s.%02d", sign, groupThousands(uint64(n/100)), n%100)
    }

    // ---- Ledgerly code ----

    // Cents is an amount of money in cents. $12.34 is Cents(1234).
    type Cents int64

    // groupThousands formats n with a comma between each group of three digits.
    func groupThousands(n uint64) string {
    	s := strconv.FormatUint(n, 10)
    	for i := len(s) - 3; i > 0; i -= 3 {
    		s = s[:i] + "," + s[i:]
    	}
    	return s
    }

    func main() {
    	for _, n := range []int64{0, -5, 123456, math.MaxInt64, math.MinInt64} {
    		fmt.Printf("Cents(%d).String() = %q\n", n, Cents(n).String())
    	}
    	fmt.Println("edge seeds:", edgeSeeds)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"math"
    	"regexp"
    	"strconv"
    	"strings"
    	"testing"
    )

    // formatted matches a well-formed Ledgerly amount: "$0.05", "-$1,234.56".
    var formatted = regexp.MustCompile(`^-?\$[0-9]{1,3}(,[0-9]{3})*\.[0-9]{2}$`)

    // digitsOnly strips the decoration from a formatted amount, so
    // "-$1,234.56" becomes "-123456".
    var digitsOnly = strings.NewReplacer("$", "", ",", "", ".", "")

    // edgeSeeds are extra seeds for FuzzCentsString: the boundary values a
    // fuzzer is unlikely to stumble on by itself.
    var edgeSeeds = []int64{math.MinInt64, math.MinInt64 + 1, -1, 0, 1, math.MaxInt64}

    // checkString reports a failure on t unless str(Cents(n)) is well-formed
    // (it matches formatted) and its digits parse back to n.
    func checkString(t testing.TB, str func(Cents) string, n int64) {
    	t.Helper()
    	s := str(Cents(n))
    	if !formatted.MatchString(s) {
    		t.Fatalf("Cents(%d).String() = %q, which isn't a well-formed amount", n, s)
    	}
    	got, err := strconv.ParseInt(digitsOnly.Replace(s), 10, 64)
    	if err != nil || got != n {
    		t.Fatalf("Cents(%d).String() = %q, whose digits parse to %d, %v; want %d", n, s, got, err, n)
    	}
    }

    // String formats c like "$1,234.56" or "-$0.05".
    func (c Cents) String() string {
    	sign := ""
    	u := uint64(c)
    	if c < 0 {
    		sign = "-"
    		u = -u // unsigned negation is exact, even for math.MinInt64
    	}
    	return fmt.Sprintf("%s$%s.%02d", sign, groupThousands(u/100), u%100)
    }

    // ---- Ledgerly code ----

    // Cents is an amount of money in cents. $12.34 is Cents(1234).
    type Cents int64

    // groupThousands formats n with a comma between each group of three digits.
    func groupThousands(n uint64) string {
    	s := strconv.FormatUint(n, 10)
    	for i := len(s) - 3; i > 0; i -= 3 {
    		s = s[:i] + "," + s[i:]
    	}
    	return s
    }

    func main() {
    	for _, n := range []int64{0, -5, 123456, math.MaxInt64, math.MinInt64} {
    		fmt.Printf("Cents(%d).String() = %q\n", n, Cents(n).String())
    	}
    	fmt.Println("edge seeds:", edgeSeeds)
    }
  tests: |
    package main

    import (
    	"fmt"
    	"math"
    	"runtime"
    	"slices"
    	"strconv"
    	"testing"
    )

    func FuzzCentsString(f *testing.F) {
    	for _, n := range []int64{0, 5, 123456, -123456789} {
    		f.Add(n)
    	}
    	for _, n := range edgeSeeds {
    		f.Add(n)
    	}
    	f.Fuzz(func(t *testing.T, n int64) {
    		checkString(t, Cents.String, n)
    	})
    }

    func TestEdgeSeeds(t *testing.T) {
    	for _, want := range []int64{math.MinInt64, math.MaxInt64, 0} {
    		if !slices.Contains(edgeSeeds, want) {
    			t.Errorf("edgeSeeds = %v; it should include the boundary value %d", edgeSeeds, want)
    		}
    	}
    }

    func TestStringBoundaries(t *testing.T) {
    	for _, tc := range []struct {
    		n    int64
    		want string
    	}{
    		{0, "$0.00"},
    		{-5, "-$0.05"},
    		{123456, "$1,234.56"},
    		{math.MaxInt64, "$92,233,720,368,547,758.07"},
    		{math.MinInt64, "-$92,233,720,368,547,758.08"},
    		{math.MinInt64 + 1, "-$92,233,720,368,547,758.07"},
    	} {
    		if got := Cents(tc.n).String(); got != tc.want {
    			t.Errorf("Cents(%d).String() = %q, want %q", tc.n, got, tc.want)
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

    func runCheck(str func(Cents) string, n int64) *recTB {
    	r := &recTB{}
    	done := make(chan struct{})
    	go func() {
    		defer close(done)
    		defer func() {
    			if p := recover(); p != nil {
    				r.panicked = fmt.Sprint(p)
    			}
    		}()
    		checkString(r, str, n)
    	}()
    	<-done
    	return r
    }

    // goodString is a correct formatter, written differently from String.
    func goodString(c Cents) string {
    	s := strconv.FormatInt(int64(c), 10)
    	sign := ""
    	if s[0] == '-' {
    		sign, s = "-", s[1:]
    	}
    	for len(s) < 3 {
    		s = "0" + s
    	}
    	dollars, cents := s[:len(s)-2], s[len(s)-2:]
    	for i := len(dollars) - 3; i > 0; i -= 3 {
    		dollars = dollars[:i] + "," + dollars[i:]
    	}
    	return sign + "$" + dollars + "." + cents
    }

    func TestCheckAcceptsCorrectFormatter(t *testing.T) {
    	for _, n := range []int64{0, 5, -5, 99, 100, -100, 123456, -123456789, 100000000, math.MaxInt64, math.MinInt64} {
    		if r := runCheck(goodString, n); r.panicked != "" {
    			t.Errorf("checkString(t, correctFormatter, %d) panicked: %s", n, r.panicked)
    		} else if r.failed {
    			t.Errorf("checkString(t, correctFormatter, %d) reported a failure (%q), but the formatter is correct", n, goodString(Cents(n)))
    		}
    	}
    }

    func TestCheckCatchesBugs(t *testing.T) {
    	bugs := []struct {
    		name string
    		n    int64
    		str  func(Cents) string
    	}{
    		{"doesn't pad the cents", 1205, func(c Cents) string {
    			return fmt.Sprintf("$%d.%d", int64(c)/100, int64(c)%100)
    		}},
    		{"leaves out the thousands separator", 123456, func(c Cents) string {
    			return fmt.Sprintf("$%d.%02d", int64(c)/100, int64(c)%100)
    		}},
    		{"groups digits in twos", 123456, func(Cents) string { return "$12,34.56" }},
    		{"drops the minus sign (the output looks fine!)", -5, func(Cents) string { return "$0.05" }},
    		{"rounds the dollars up (the output looks fine!)", 123456, func(Cents) string { return "$1,235.56" }},
    		{"negates math.MinInt64 naively", math.MinInt64, func(c Cents) string {
    			n := int64(c)
    			if n < 0 {
    				n = -n
    			}
    			return fmt.Sprintf("-$%d.%02d", n/100, n%100)
    		}},
    	}
    	for _, b := range bugs {
    		r := runCheck(b.str, b.n)
    		if r.panicked != "" {
    			t.Errorf("checkString with a formatter that %s, on %d, panicked: %s", b.name, b.n, r.panicked)
    		} else if !r.failed {
    			t.Errorf("checkString(t, str, %d) didn't report a failure for a formatter that %s (it returned %q)", b.n, b.name, b.str(Cents(b.n)))
    		}
    	}
    }
---

A fuzz test is only as good as its properties. "Doesn't panic" is a start, but most bugs return a wrong answer rather than crashing. Here are the patterns that catch them.

## Pattern 1: round trips

If you have an encoder and a decoder, decoding what you encoded must give back the original:

```go
func FuzzFormatRoundTrip(f *testing.F) {
	f.Add(int64(1234))
	f.Add(int64(-5))
	f.Fuzz(func(t *testing.T, n int64) {
		c := Cents(n)
		got, err := ParseAmount(Format(c))
		if err != nil {
			t.Fatalf("ParseAmount(Format(%d)) = error %v; Format gave %q", n, err, Format(c))
		}
		if got != c {
			t.Fatalf("ParseAmount(Format(%d)) = %d", n, got)
		}
	})
}
```

Round trips apply everywhere: JSON marshal/unmarshal, CSV write/read, compress/decompress, `String`/`Parse`.

## Pattern 2: an oracle

Compare against a second implementation that's slow or limited but obviously right. In the first lesson, the oracle was a regular expression for the amount grammar. Other oracles:

- the old implementation, when you rewrite something for speed (**differential** fuzzing)
- a brute-force version of a clever algorithm
- the standard library (`AppendCents` from the benchmarks chapter must produce exactly the bytes of `String`, and a hand-rolled integer parser should agree with `strconv.ParseInt`)

## Pattern 3: invariants

Things that must be true of any output, whatever the input:

- A statement's `TOTAL` line equals the sum of its rows.
- Sorting returns a permutation of the input, in order.
- `ImportCSV` never returns more transactions than the input has lines.
- A parsed amount's sign matches the input's sign (this is what caught the overflow).

## Pattern 4: no crash on garbage

For code that handles untrusted input (file imports, network protocols), just calling the function with arbitrary bytes is valuable. `ImportCSV(bytes.NewReader(data))` must return an error for junk, never panic, never hang.

## What coverage guidance does

Go's fuzzer is **coverage-guided**. It instruments your code, watches which branches each input reaches, and keeps inputs that reach new ones. It then mutates those. That's how it finds `"+0"` in 64 tries: it doesn't guess randomly, it climbs through the `if` statements of `ParseAmount` and `strconv.ParseInt` one branch at a time.

## What it can't find

Remember the bug left in `Cents.String` back in chapter 1? For negative amounts it computes `n = -n`, and for the smallest `int64`, `-9223372036854775808`, negation overflows and gives back the same negative number. The result is `-$-92,233,720,368,547,758.-8`.

Here's a fuzz test that checks the output looks like a well-formed amount:

```go
var formatted = regexp.MustCompile(`^-?\$[0-9]{1,3}(,[0-9]{3})*\.[0-9]{2}$`)

func FuzzCentsString(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(123456))
	f.Add(int64(-5))
	f.Fuzz(func(t *testing.T, n int64) {
		if s := Cents(n).String(); !formatted.MatchString(s) {
			t.Fatalf("Cents(%d).String() = %q, which isn't a well-formed amount", n, s)
		}
	})
}
```

Run it for a minute and a half:

```text
fuzz: elapsed: 1m30s, execs: 32964539 (0/sec), new interesting: 0 (total: 3)
PASS
```

Thirty-three million inputs, no failure. The property is right and the bug is real, but there's exactly one bad value in 2⁶⁴, and no branch in `String` points the fuzzer towards it. "New interesting: 0" is the tell: the fuzzer found nothing new to explore.

The fix is to **seed the boundaries yourself**:

```go
f.Add(int64(math.MinInt64))
f.Add(int64(math.MaxInt64))
```

Now a plain `go test` fails immediately, no fuzzing needed. Fuzzing and hand-picked cases complement each other: you know where the cliffs are (zero, one, max, min, empty, huge), and the fuzzer explores everything in between.

## So what should String do?

Once found, you still have to decide. For Ledgerly, `math.MinInt64` cents is about -92 quadrillion dollars, so no real ledger will contain it. Reasonable fixes are handling it specially, or working with `uint64` for the magnitude (`uint64(-n)` is correct even for the minimum value). What matters is that the behaviour is now a *decision*, pinned by a test, rather than an accident.

## Your turn: seed the cliff, then fix it

The grader has a fuzz test that calls your code, and a plain `go test` runs it over its seed corpus:

```go
f.Fuzz(func(t *testing.T, n int64) {
	checkString(t, Cents.String, n)
})
```

1. **Write `checkString(t, str, n)`.** Call `str(Cents(n))` and report a failure with `t.Fatalf` unless the result matches `formatted` **and** its digits parse back to `n`: strip the decoration with `digitsOnly.Replace(s)` and parse it with `strconv.ParseInt(..., 10, 64)`. You need both halves. The regex alone accepts `"$0.05"` for `-5`, and the round trip alone accepts `"$1234.56"`. Use the `str` argument: the grader passes in six broken formatters that your check must catch, and a correct one it must accept.
2. **Fill in `edgeSeeds`** with the boundary values from this lesson, including `math.MinInt64`, `math.MaxInt64` and `0`. The grader adds them to the seed corpus.
3. Your seeds now make the fuzz test fail. **Fix `String`** so it's correct for every `int64`, using a `uint64` magnitude as described above. `groupThousands` already takes a `uint64`.

**Run** prints `String` for a few amounts, including both extremes.
