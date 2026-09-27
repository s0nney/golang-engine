---
title: Properties of a Fair Split
difficulty: medium
after: fuzzing
hints:
  - 'A property is a rule that must hold for **every** input, so you never need to know the "right answer" for a particular total. Check the length first and stop with `t.Fatalf` if it''s wrong, because the other checks index into `parts`.'
  - 'Then walk the parts once, keeping a running sum. The sum must equal `total`; each part must be less than or equal to the part before it; and `parts[0] - parts[n-1]` (the biggest minus the smallest, once you know they''re in order) must be at most 1.'
  - 'Use the `split` argument, never `Split` directly: the grader passes in broken versions to see if your properties notice. And remember negative totals: Ledgerly splits refunds too.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"testing"
    )

    // checkSplit is the body of a fuzz target. It calls split(total, n) for
    // n >= 1 and reports on t any broken property of a fair split.
    func checkSplit(t testing.TB, split func(Cents, int) []Cents, total Cents, n int) {
    	t.Helper()
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    // Split divides total into n parts (n >= 1) that differ by at most one
    // cent, with the larger parts first, so the parts always add up to total.
    func Split(total Cents, n int) []Cents {
    	parts := make([]Cents, n)
    	base := total / Cents(n)
    	if total%Cents(n) < 0 {
    		base-- // round down, not toward zero, for negative totals
    	}
    	extra := int(total - base*Cents(n))
    	for i := range parts {
    		parts[i] = base
    		if i < extra {
    			parts[i]++
    		}
    	}
    	return parts
    }

    // naiveSplit ignores the remainder.
    func naiveSplit(total Cents, n int) []Cents {
    	parts := make([]Cents, n)
    	for i := range parts {
    		parts[i] = total / Cents(n)
    	}
    	return parts
    }

    // spyTB is a pretend testing.TB so Run can show what checkSplit reports.
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
    	for _, in := range []struct {
    		total Cents
    		n     int
    	}{{1000, 3}, {-1000, 3}, {7, 1}} {
    		spy := &spyTB{}
    		checkSplit(spy, Split, in.total, in.n)
    		fmt.Printf("Split(%d, %d) = %v, failures: %q\n", in.total, in.n, Split(in.total, in.n), spy.failures)
    		spy = &spyTB{}
    		checkSplit(spy, naiveSplit, in.total, in.n)
    		fmt.Printf("naiveSplit(%d, %d) = %v, failures: %q\n", in.total, in.n, naiveSplit(in.total, in.n), spy.failures)
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"testing"
    )

    // checkSplit is the body of a fuzz target. It calls split(total, n) for
    // n >= 1 and reports on t any broken property of a fair split.
    func checkSplit(t testing.TB, split func(Cents, int) []Cents, total Cents, n int) {
    	t.Helper()
    	parts := split(total, n)
    	if len(parts) != n {
    		t.Fatalf("split(%d, %d) returned %d parts %v, want %d", total, n, len(parts), parts, n)
    	}
    	var sum Cents
    	for i, p := range parts {
    		sum += p
    		if i > 0 && p > parts[i-1] {
    			t.Fatalf("split(%d, %d) = %v: part %d is bigger than part %d; larger parts go first", total, n, parts, i, i-1)
    		}
    	}
    	if sum != total {
    		t.Fatalf("split(%d, %d) = %v, which adds up to %d", total, n, parts, sum)
    	}
    	if parts[0]-parts[n-1] > 1 {
    		t.Fatalf("split(%d, %d) = %v: parts differ by more than one cent", total, n, parts)
    	}
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    // Split divides total into n parts (n >= 1) that differ by at most one
    // cent, with the larger parts first, so the parts always add up to total.
    func Split(total Cents, n int) []Cents {
    	parts := make([]Cents, n)
    	base := total / Cents(n)
    	if total%Cents(n) < 0 {
    		base-- // round down, not toward zero, for negative totals
    	}
    	extra := int(total - base*Cents(n))
    	for i := range parts {
    		parts[i] = base
    		if i < extra {
    			parts[i]++
    		}
    	}
    	return parts
    }

    // naiveSplit ignores the remainder.
    func naiveSplit(total Cents, n int) []Cents {
    	parts := make([]Cents, n)
    	for i := range parts {
    		parts[i] = total / Cents(n)
    	}
    	return parts
    }

    // spyTB is a pretend testing.TB so Run can show what checkSplit reports.
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
    	for _, in := range []struct {
    		total Cents
    		n     int
    	}{{1000, 3}, {-1000, 3}, {7, 1}} {
    		spy := &spyTB{}
    		checkSplit(spy, Split, in.total, in.n)
    		fmt.Printf("Split(%d, %d) = %v, failures: %q\n", in.total, in.n, Split(in.total, in.n), spy.failures)
    		spy = &spyTB{}
    		checkSplit(spy, naiveSplit, in.total, in.n)
    		fmt.Printf("naiveSplit(%d, %d) = %v, failures: %q\n", in.total, in.n, naiveSplit(in.total, in.n), spy.failures)
    	}
    }
  tests: |
    package main

    import (
    	"fmt"
    	"runtime"
    	"testing"
    )

    // FuzzSplit is how checkSplit is meant to be used. The grader only runs
    // the seeds; locally you'd run go test -fuzz FuzzSplit.
    func FuzzSplit(f *testing.F) {
    	for _, s := range []struct {
    		total int64
    		n     int
    	}{{1000, 3}, {-1000, 3}, {0, 5}, {2, 5}, {-2, 5}, {7, 1}, {99, 100}, {-1, 2}, {123456789, 7}} {
    		f.Add(s.total, s.n)
    	}
    	f.Fuzz(func(t *testing.T, total int64, n int) {
    		n = 1 + (n%100+100)%100 // 1..100
    		total %= 1_000_000_000_000
    		checkSplit(t, Split, Cents(total), n)
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

    func runCheck(split func(Cents, int) []Cents, total Cents, n int) (failed bool, panicked string) {
    	r := &recTB{}
    	done := make(chan struct{})
    	go func() {
    		defer close(done)
    		defer func() {
    			if p := recover(); p != nil {
    				panicked = fmt.Sprint(p)
    			}
    		}()
    		checkSplit(r, split, total, n)
    	}()
    	<-done
    	return r.failed, panicked
    }

    // spread is a correct split built a different way: each part gets its
    // share of whatever is still left, rounded up.
    func spread(total Cents, n int) []Cents {
    	parts := make([]Cents, n)
    	left := total
    	for i := range parts {
    		k := Cents(n - i)
    		share := left / k
    		if left%k > 0 {
    			share++
    		}
    		parts[i] = share
    		left -= share
    	}
    	return parts
    }

    func TestCheckAcceptsCorrectSplit(t *testing.T) {
    	for _, total := range []Cents{0, 1, -1, 2, -2, 99, 100, 101, 1000, -1000, 123456789, -123456789} {
    		for _, n := range []int{1, 2, 3, 7, 100} {
    			for name, split := range map[string]func(Cents, int) []Cents{"Split": Split, "another correct split": spread} {
    				failed, p := runCheck(split, total, n)
    				if p != "" {
    					t.Fatalf("checkSplit(t, %s, %d, %d) panicked: %s", name, total, n, p)
    				}
    				if failed {
    					t.Fatalf("checkSplit(t, %s, %d, %d) reported a failure, but %s is correct (it returns %v)", name, total, n, name, split(total, n))
    				}
    			}
    		}
    	}
    }

    func TestCheckCatchesBugs(t *testing.T) {
    	bugs := []struct {
    		bug   string
    		total Cents
    		n     int
    		split func(Cents, int) []Cents
    	}{
    		{"drops the remainder", 1000, 3, func(total Cents, n int) []Cents {
    			parts := make([]Cents, n)
    			for i := range parts {
    				parts[i] = total / Cents(n)
    			}
    			return parts
    		}},
    		{"gives the whole remainder to the first part", 1003, 4, func(total Cents, n int) []Cents {
    			parts := Split(total-total%Cents(n), n)
    			parts[0] += total % Cents(n)
    			return parts
    		}},
    		{"puts the larger parts last", 1000, 3, func(total Cents, n int) []Cents {
    			parts := Split(total, n)
    			for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
    				parts[i], parts[j] = parts[j], parts[i]
    			}
    			return parts
    		}},
    		{"leaves out zero parts", 2, 5, func(total Cents, n int) []Cents {
    			var parts []Cents
    			for _, p := range Split(total, n) {
    				if p != 0 {
    					parts = append(parts, p)
    				}
    			}
    			return parts
    		}},
    		{"rounds toward zero, so negative totals lose cents", -1000, 3, func(total Cents, n int) []Cents {
    			parts := make([]Cents, n)
    			base, extra := total/Cents(n), int(total%Cents(n))
    			for i := range parts {
    				parts[i] = base
    				if i < extra {
    					parts[i]++
    				}
    			}
    			return parts
    		}},
    		{"returns one part too many", 10, 2, func(total Cents, n int) []Cents {
    			return append(Split(total, n), 0)
    		}},
    	}
    	for _, b := range bugs {
    		failed, p := runCheck(b.split, b.total, b.n)
    		if p != "" {
    			t.Errorf("checkSplit(t, split, %d, %d) panicked for a split that %s: %s (check the length before indexing, and stop with t.Fatalf)", b.total, b.n, b.bug, p)
    		} else if !failed {
    			t.Errorf("checkSplit(t, split, %d, %d) reported nothing for a split that %s (it returns %v)", b.total, b.n, b.bug, b.split(b.total, b.n))
    		}
    	}
    }
---

Ledgerly splits shared bills, and `Split` has passed every hand-picked test.
But splitting money has a lot of corners (remainders, tiny totals, negative
refunds), which makes it a perfect job for a fuzzer. A fuzzer needs
**properties**: rules that hold for every input, so it can check inputs
nobody worked out the answer for.

Write `checkSplit(t, split, total, n)`. It calls `split(total, n)` (with
`n ≥ 1`) and reports a failure on `t` if any property of a fair split is
broken:

1. **Length.** There are exactly `n` parts.
2. **Nothing lost.** The parts add up to exactly `total`.
3. **Fair.** No two parts differ by more than one cent.
4. **Ordered.** Larger parts come first: each part is less than or equal to
   the one before it.

## Example

```text
split(1000, 3)  = [334 333 333]     ✓
split(-1000, 3) = [-333 -333 -334]  ✓  (larger means closer to +∞)
split(1000, 3)  = [333 333 333]     ✗  adds up to 999
split(1003, 4)  = [253 250 250 250] ✗  parts differ by 3
```

## How you're graded

The hidden tests use your function as a fuzz target body:

```go
f.Fuzz(func(t *testing.T, total int64, n int) {
	checkSplit(t, Split, Cents(total), clamp(n))
})
```

They also pass in a correct split for many inputs (no failure allowed), and
**six broken splits**, each with an input that exposes it. Each must produce
a failure, and your check must never panic, even when `split` returns the
wrong number of parts.

**Run** shows what your check reports for `Split` and for `naiveSplit`,
which drops the remainder.

## Constraints

- Totals are within ±1,000,000,000,000 cents, and `n` is between 1 and 100, so
  the sum can't overflow.
