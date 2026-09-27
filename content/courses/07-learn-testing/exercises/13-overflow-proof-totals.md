---
title: Overflow-Proof Totals
difficulty: hard
after: fuzzing
hints:
  - 'You need an oracle that can''t overflow. `math/big` is perfect: add every amount into a `big.Int` with `exact.Add(exact, big.NewInt(int64(a)))`, then `exact.IsInt64()` tells you whether the true total fits, and `exact.Int64()` gives it.'
  - 'Take a copy of `amounts` (`slices.Clone`) **before** calling `sum`, and compare afterwards with `slices.Equal`: a `sum` that sorts or rewrites its input is broken even if its answer is right.'
  - 'Then there are just two cases. The total fits: `sum` must return exactly it and a `nil` error, even if a running total along the way wouldn''t fit. It doesn''t fit: `sum` must return `0` and an error for which `errors.Is(err, ErrOverflow)` holds.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"math"
    	"math/bits"
    	"testing"
    )

    // checkSum is the body of a fuzz target. It calls sum(amounts) and
    // reports on t any way the result breaks Sum's contract.
    func checkSum(t testing.TB, sum func([]Cents) (Cents, error), amounts []Cents) {
    	t.Helper()
    }

    // ---- Ledgerly code (already done, and correct) ----

    type Cents int64

    var ErrOverflow = errors.New("ledgerly: total out of range")

    // Sum returns the total of amounts. If the total doesn't fit in an
    // int64 it returns 0 and an error wrapping ErrOverflow. Only the final
    // total has to fit: {MaxInt64, 1, -1} sums to MaxInt64 just fine.
    // Sum doesn't modify amounts.
    func Sum(amounts []Cents) (Cents, error) {
    	// A 128-bit running total: hi:lo, in two's complement.
    	var hi int64
    	var lo uint64
    	for _, a := range amounts {
    		var carry uint64
    		lo, carry = bits.Add64(lo, uint64(a), 0)
    		hi += int64(carry)
    		if a < 0 {
    			hi-- // the high word of a negative int64 is all ones
    		}
    	}
    	if (hi == 0 && int64(lo) >= 0) || (hi == -1 && int64(lo) < 0) {
    		return Cents(lo), nil
    	}
    	return 0, fmt.Errorf("%w: %d amounts", ErrOverflow, len(amounts))
    }

    // wrappingSum is what most people write first.
    func wrappingSum(amounts []Cents) (Cents, error) {
    	var total Cents
    	for _, a := range amounts {
    		total += a
    	}
    	return total, nil
    }

    // spyTB is a pretend testing.TB so Run can show what checkSum reports.
    // Unlike the real thing, its Fatal methods don't stop the function.
    type spyTB struct {
    	testing.TB
    	failures []string
    }

    func (s *spyTB) Helper()                   {}
    func (s *spyTB) Fail()                     { s.failures = append(s.failures, "(Fail)") }
    func (s *spyTB) FailNow()                  { s.failures = append(s.failures, "(FailNow)") }
    func (s *spyTB) Error(a ...any)            { s.failures = append(s.failures, fmt.Sprint(a...)) }
    func (s *spyTB) Fatal(a ...any)            { s.failures = append(s.failures, fmt.Sprint(a...)) }
    func (s *spyTB) Errorf(f string, a ...any) { s.failures = append(s.failures, fmt.Sprintf(f, a...)) }
    func (s *spyTB) Fatalf(f string, a ...any) { s.failures = append(s.failures, fmt.Sprintf(f, a...)) }

    func main() {
    	for _, amounts := range [][]Cents{
    		{1200, -350, 99},
    		{math.MaxInt64, 1},
    		{math.MaxInt64, 1, -1},
    	} {
    		total, err := Sum(amounts)
    		fmt.Printf("Sum(%v) = %d, %v\n", amounts, total, err)
    		for name, sum := range map[string]func([]Cents) (Cents, error){"Sum": Sum, "wrappingSum": wrappingSum} {
    			spy := &spyTB{}
    			checkSum(spy, sum, amounts)
    			for _, f := range spy.failures {
    				fmt.Printf("    checkSum(%s): %s\n", name, f)
    			}
    		}
    	}
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"math"
    	"math/big"
    	"math/bits"
    	"slices"
    	"testing"
    )

    // checkSum is the body of a fuzz target. It calls sum(amounts) and
    // reports on t any way the result breaks Sum's contract.
    func checkSum(t testing.TB, sum func([]Cents) (Cents, error), amounts []Cents) {
    	t.Helper()
    	exact := new(big.Int)
    	for _, a := range amounts {
    		exact.Add(exact, big.NewInt(int64(a)))
    	}
    	before := slices.Clone(amounts)
    	got, err := sum(amounts)
    	if !slices.Equal(amounts, before) {
    		t.Fatalf("sum(%v) changed its input to %v", before, amounts)
    	}
    	if exact.IsInt64() {
    		if err != nil || int64(got) != exact.Int64() {
    			t.Fatalf("sum(%v) = %d, %v; want %d, nil", amounts, got, err, exact)
    		}
    		return
    	}
    	if !errors.Is(err, ErrOverflow) || got != 0 {
    		t.Fatalf("sum(%v) = %d, %v; the total %d doesn't fit, want 0 and an error wrapping ErrOverflow", amounts, got, err, exact)
    	}
    }

    // ---- Ledgerly code (already done, and correct) ----

    type Cents int64

    var ErrOverflow = errors.New("ledgerly: total out of range")

    // Sum returns the total of amounts. If the total doesn't fit in an
    // int64 it returns 0 and an error wrapping ErrOverflow. Only the final
    // total has to fit: {MaxInt64, 1, -1} sums to MaxInt64 just fine.
    // Sum doesn't modify amounts.
    func Sum(amounts []Cents) (Cents, error) {
    	// A 128-bit running total: hi:lo, in two's complement.
    	var hi int64
    	var lo uint64
    	for _, a := range amounts {
    		var carry uint64
    		lo, carry = bits.Add64(lo, uint64(a), 0)
    		hi += int64(carry)
    		if a < 0 {
    			hi-- // the high word of a negative int64 is all ones
    		}
    	}
    	if (hi == 0 && int64(lo) >= 0) || (hi == -1 && int64(lo) < 0) {
    		return Cents(lo), nil
    	}
    	return 0, fmt.Errorf("%w: %d amounts", ErrOverflow, len(amounts))
    }

    // wrappingSum is what most people write first.
    func wrappingSum(amounts []Cents) (Cents, error) {
    	var total Cents
    	for _, a := range amounts {
    		total += a
    	}
    	return total, nil
    }

    // spyTB is a pretend testing.TB so Run can show what checkSum reports.
    // Unlike the real thing, its Fatal methods don't stop the function.
    type spyTB struct {
    	testing.TB
    	failures []string
    }

    func (s *spyTB) Helper()                   {}
    func (s *spyTB) Fail()                     { s.failures = append(s.failures, "(Fail)") }
    func (s *spyTB) FailNow()                  { s.failures = append(s.failures, "(FailNow)") }
    func (s *spyTB) Error(a ...any)            { s.failures = append(s.failures, fmt.Sprint(a...)) }
    func (s *spyTB) Fatal(a ...any)            { s.failures = append(s.failures, fmt.Sprint(a...)) }
    func (s *spyTB) Errorf(f string, a ...any) { s.failures = append(s.failures, fmt.Sprintf(f, a...)) }
    func (s *spyTB) Fatalf(f string, a ...any) { s.failures = append(s.failures, fmt.Sprintf(f, a...)) }

    func main() {
    	for _, amounts := range [][]Cents{
    		{1200, -350, 99},
    		{math.MaxInt64, 1},
    		{math.MaxInt64, 1, -1},
    	} {
    		total, err := Sum(amounts)
    		fmt.Printf("Sum(%v) = %d, %v\n", amounts, total, err)
    		for name, sum := range map[string]func([]Cents) (Cents, error){"Sum": Sum, "wrappingSum": wrappingSum} {
    			spy := &spyTB{}
    			checkSum(spy, sum, amounts)
    			for _, f := range spy.failures {
    				fmt.Printf("    checkSum(%s): %s\n", name, f)
    			}
    		}
    	}
    }
  tests: |
    package main

    import (
    	"encoding/binary"
    	"errors"
    	"fmt"
    	"math"
    	"math/big"
    	"runtime"
    	"slices"
    	"testing"
    )

    // FuzzSum is how checkSum is meant to be used: every 8 bytes of fuzz
    // input become one amount. The grader only runs the seeds.
    func FuzzSum(f *testing.F) {
    	for _, seed := range [][]Cents{
    		nil, {0}, {1200, -350, 99}, {math.MaxInt64}, {math.MinInt64},
    		{math.MaxInt64, 1}, {math.MinInt64, -1}, {math.MaxInt64, 1, -1},
    		{math.MinInt64, -1, 1}, {math.MaxInt64, math.MaxInt64, math.MinInt64},
    	} {
    		var data []byte
    		for _, a := range seed {
    			data = binary.LittleEndian.AppendUint64(data, uint64(a))
    		}
    		f.Add(data)
    	}
    	f.Fuzz(func(t *testing.T, data []byte) {
    		var amounts []Cents
    		for len(data) >= 8 {
    			amounts = append(amounts, Cents(binary.LittleEndian.Uint64(data)))
    			data = data[8:]
    		}
    		checkSum(t, Sum, amounts)
    	})
    }

    // recTB records failures. Fatal stops the goroutine like the real thing.
    type recTB struct {
    	testing.TB
    	failed bool
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
    func (r *recTB) Failed() bool          { return r.failed }

    type sumFunc = func([]Cents) (Cents, error)

    func runCheck(sum sumFunc, amounts []Cents) (failed bool, panicked string) {
    	r := &recTB{}
    	done := make(chan struct{})
    	go func() {
    		defer close(done)
    		defer func() {
    			if p := recover(); p != nil {
    				panicked = fmt.Sprint(p)
    			}
    		}()
    		checkSum(r, sum, amounts)
    	}()
    	<-done
    	return r.failed, panicked
    }

    const maxC, minC = Cents(math.MaxInt64), Cents(math.MinInt64)

    // bigSum is a second correct implementation.
    func bigSum(amounts []Cents) (Cents, error) {
    	total := new(big.Int)
    	for _, a := range amounts {
    		total.Add(total, big.NewInt(int64(a)))
    	}
    	if !total.IsInt64() {
    		return 0, fmt.Errorf("bigSum: %w", ErrOverflow)
    	}
    	return Cents(total.Int64()), nil
    }

    func TestCheckAcceptsCorrectSums(t *testing.T) {
    	inputs := [][]Cents{
    		nil, {}, {0}, {1200, -350, 99}, {-5}, {maxC}, {minC}, {maxC, minC},
    		{maxC - 1, 1}, {minC + 1, -1}, {maxC, 1, -1}, {minC, -1, 1}, {maxC, maxC, minC, minC},
    		{maxC, 1}, {minC, -1}, {maxC, maxC}, {minC, minC}, {maxC, maxC, minC},
    	}
    	for name, sum := range map[string]sumFunc{"Sum": Sum, "another correct sum": bigSum} {
    		for _, in := range inputs {
    			failed, p := runCheck(sum, slices.Clone(in))
    			if p != "" {
    				t.Fatalf("checkSum(t, %s, %v) panicked: %s", name, in, p)
    			}
    			if failed {
    				got, err := sum(slices.Clone(in))
    				t.Fatalf("checkSum(t, %s, %v) reported a failure, but %s is correct (it returns %d, %v)", name, in, name, got, err)
    			}
    		}
    	}
    }

    // stepSum adds one amount at a time with an overflow check on each step.
    func stepSum(amounts []Cents, strictMax, positiveOnly bool) (Cents, error) {
    	var total Cents
    	for _, a := range amounts {
    		limit := maxC - a
    		if strictMax {
    			limit--
    		}
    		if a > 0 && total > limit {
    			return 0, ErrOverflow
    		}
    		if !positiveOnly && a < 0 && total < minC-a {
    			return 0, ErrOverflow
    		}
    		total += a
    	}
    	return total, nil
    }

    func TestCheckCatchesBugs(t *testing.T) {
    	bugs := []struct {
    		bug     string
    		amounts []Cents
    		sum     sumFunc
    	}{
    		{"wraps round silently", []Cents{maxC, 1}, func(a []Cents) (Cents, error) {
    			var total Cents
    			for _, x := range a {
    				total += x
    			}
    			return total, nil
    		}},
    		{"only notices overflow past MaxInt64, not below MinInt64", []Cents{minC, -1}, func(a []Cents) (Cents, error) {
    			return stepSum(a, false, true)
    		}},
    		{"rejects a total of exactly MaxInt64", []Cents{maxC - 1, 1}, func(a []Cents) (Cents, error) {
    			return stepSum(a, true, false)
    		}},
    		{"rejects a running total that overflows even though the final total fits", []Cents{maxC, 1, -1}, func(a []Cents) (Cents, error) {
    			return stepSum(a, false, false)
    		}},
    		{"rejects a total of exactly MinInt64 (it checks -total)", []Cents{minC + 1, -1}, func(a []Cents) (Cents, error) {
    			total, err := Sum(a)
    			if err == nil && total == minC {
    				return 0, fmt.Errorf("%w: -total overflows", ErrOverflow)
    			}
    			return total, err
    		}},
    		{"returns an error that doesn't wrap ErrOverflow", []Cents{maxC, maxC}, func(a []Cents) (Cents, error) {
    			total, err := Sum(a)
    			if err != nil {
    				return 0, errors.New("total out of range")
    			}
    			return total, nil
    		}},
    		{"returns the wrapped-round total along with the error", []Cents{maxC, 10}, func(a []Cents) (Cents, error) {
    			_, err := Sum(a)
    			var total Cents
    			for _, x := range a {
    				total += x
    			}
    			return total, err
    		}},
    		{"sorts its input in place (to add the negatives first)", []Cents{5, -3, 2}, func(a []Cents) (Cents, error) {
    			slices.Sort(a)
    			return Sum(a)
    		}},
    		{"treats an empty list as an error", []Cents{}, func(a []Cents) (Cents, error) {
    			if len(a) == 0 {
    				return 0, fmt.Errorf("%w: nothing to add", ErrOverflow)
    			}
    			return Sum(a)
    		}},
    	}
    	for _, b := range bugs {
    		failed, p := runCheck(b.sum, slices.Clone(b.amounts))
    		if p != "" {
    			t.Errorf("checkSum(t, sum, %v) panicked for a sum that %s: %s", b.amounts, b.bug, p)
    		} else if !failed {
    			got, err := b.sum(slices.Clone(b.amounts))
    			t.Errorf("checkSum(t, sum, %v) reported nothing for a sum that %s (it returns %d, %v)", b.amounts, b.bug, got, err)
    		}
    	}
    }
---

Ledgerly adds up balances across every account a company has, and a total
that silently wraps round from +92 quadrillion dollars to −92 quadrillion is
the worst kind of bug: no crash, just a wrong number. `Sum` is written
carefully with a 128-bit running total. Now it needs a fuzz test that would
catch anyone who "simplifies" it later.

Write `checkSum(t, sum, amounts)`, the body of a fuzz target. It calls
`sum(amounts)` and reports a failure on `t` if the result breaks the
contract:

1. **Exact when it fits.** If the true mathematical total fits in an `int64`,
   `sum` returns exactly that total and a `nil` error, even if a running total
   along the way wouldn't fit (`{MaxInt64, 1, -1}` is fine).
2. **An error when it doesn't.** Otherwise `sum` returns `0` and an error
   for which `errors.Is(err, ErrOverflow)` holds.
3. **Hands off.** `sum` doesn't modify `amounts`.

An empty list sums to `0`.

## Example

```text
sum([1200 -350 99])       → 949, nil
sum([MaxInt64 1])         → 0, ledgerly: total out of range
sum([MaxInt64 1 -1])      → MaxInt64, nil
sum([MinInt64])           → MinInt64, nil
```

## How you're graded

The hidden tests use `checkSum` in a fuzz target that turns every 8 bytes of
fuzz input into an amount. They also pass it two correct implementations on
18 tricky inputs (no failure allowed), and **nine broken ones**, each with
an input that exposes it: wrapping, one-sided overflow checks, off-by-one
errors at `MaxInt64` and `MinInt64`, overly cautious step-by-step checks,
unwrapped errors, a non-zero result alongside the error, an input sorted in
place, and an empty list treated as an error. Each must produce a failure,
and your check must never panic.

**Run** shows `Sum`'s results for three inputs and what your check reports
about `Sum` and about `wrappingSum`.

## Constraints

- Your oracle must not overflow either. Checking the answer with `int64`
  arithmetic brings back exactly the bug you're hunting.
- Use the `sum` argument, not `Sum`.
