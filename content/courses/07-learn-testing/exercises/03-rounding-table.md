---
title: A Rounding Table
difficulty: easy
after: table-driven-tests
hints:
  - 'Bugs hide at the edges. For rounding, the edges are the amounts just below, exactly on, and just above the halfway point: `149`, `150` and `151`.'
  - 'Now do the same for negative amounts (`-149`, `-150`, `-151`). "Away from zero" means `-150` becomes `-200`, and a sloppy implementation that just adds 50 gets the negative side wrong.'
exercise:
  starter: |
    package main

    import "fmt"

    // roundCase is one row of the RoundToDollar test table.
    type roundCase struct {
    	name string
    	in   Cents
    	want Cents
    }

    // roundCases is run against RoundToDollar (must all pass) and against
    // four broken versions of it (each must fail at least one case).
    var roundCases = []roundCase{
    	{name: "exact dollar", in: 1200, want: 1200},
    	// Add cases. Think about halfway points and negative amounts.
    }

    // ---- Ledgerly code (already done, and correct) ----

    type Cents int64

    // RoundToDollar rounds c to the nearest whole dollar. Exact halves
    // round away from zero: $1.50 becomes $2.00 and -$1.50 becomes -$2.00.
    func RoundToDollar(c Cents) Cents {
    	rem := c % 100
    	switch {
    	case rem >= 50:
    		return c - rem + 100
    	case rem <= -50:
    		return c - rem - 100
    	default:
    		return c - rem
    	}
    }

    func main() {
    	for _, tc := range roundCases {
    		fmt.Printf("%-20s RoundToDollar(%d) = %d, want %d\n", tc.name, tc.in, RoundToDollar(tc.in), tc.want)
    	}
    }
  solution: |
    package main

    import "fmt"

    // roundCase is one row of the RoundToDollar test table.
    type roundCase struct {
    	name string
    	in   Cents
    	want Cents
    }

    // roundCases is run against RoundToDollar (must all pass) and against
    // four broken versions of it (each must fail at least one case).
    var roundCases = []roundCase{
    	{name: "exact dollar", in: 1200, want: 1200},
    	{name: "zero", in: 0, want: 0},
    	{name: "just under half", in: 149, want: 100},
    	{name: "exactly half", in: 150, want: 200},
    	{name: "just over half", in: 151, want: 200},
    	{name: "negative under half", in: -149, want: -100},
    	{name: "negative half", in: -150, want: -200},
    	{name: "negative over half", in: -151, want: -200},
    }

    // ---- Ledgerly code (already done, and correct) ----

    type Cents int64

    // RoundToDollar rounds c to the nearest whole dollar. Exact halves
    // round away from zero: $1.50 becomes $2.00 and -$1.50 becomes -$2.00.
    func RoundToDollar(c Cents) Cents {
    	rem := c % 100
    	switch {
    	case rem >= 50:
    		return c - rem + 100
    	case rem <= -50:
    		return c - rem - 100
    	default:
    		return c - rem
    	}
    }

    func main() {
    	for _, tc := range roundCases {
    		fmt.Printf("%-20s RoundToDollar(%d) = %d, want %d\n", tc.name, tc.in, RoundToDollar(tc.in), tc.want)
    	}
    }
  tests: |
    package main

    import "testing"

    func TestTableIsWellFormed(t *testing.T) {
    	seen := map[string]bool{}
    	for i, tc := range roundCases {
    		if tc.name == "" {
    			t.Errorf("case %d (in: %d) has no name", i, tc.in)
    		} else if seen[tc.name] {
    			t.Errorf("two cases are named %q; names must be unique", tc.name)
    		}
    		seen[tc.name] = true
    	}
    }

    func TestTableAgainstRealRounding(t *testing.T) {
    	for _, tc := range roundCases {
    		if got := RoundToDollar(tc.in); got != tc.want {
    			t.Errorf("case %q is wrong: RoundToDollar(%d) = %d, but the case wants %d (RoundToDollar is correct)", tc.name, tc.in, got, tc.want)
    		}
    	}
    }

    var mutants = []struct {
    	bug   string
    	round func(Cents) Cents
    }{
    	{"always rounds down (truncates), so 150 gives 100", func(c Cents) Cents {
    		return c - c%100
    	}},
    	{"rounds halves up instead of away from zero, so -150 gives -100", func(c Cents) Cents {
    		if c%100 == -50 {
    			return c + 50
    		}
    		return RoundToDollar(c)
    	}},
    	{"adds 50 and truncates, which breaks negative amounts (-151 gives -100)", func(c Cents) Cents {
    		return (c + 50) / 100 * 100
    	}},
    	{"treats 49 cents as half, so 149 gives 200", func(c Cents) Cents {
    		if c%100 == 49 {
    			return c + 51
    		}
    		if c%100 == -49 {
    			return c - 51
    		}
    		return RoundToDollar(c)
    	}},
    }

    func TestTableCatchesBugs(t *testing.T) {
    	for _, m := range mutants {
    		caught := false
    		for _, tc := range roundCases {
    			if m.round(tc.in) != tc.want {
    				caught = true
    				break
    			}
    		}
    		if !caught {
    			t.Errorf("no case catches a RoundToDollar that %s", m.bug)
    		}
    	}
    }
---

Ledgerly's yearly summary shows whole dollars, so it rounds every total with
`RoundToDollar`. The function is finished and correct. Your job is to write
the **test table** that would catch anyone who breaks it later.

Add cases to `roundCases`. The grader checks that:

1. every case passes against the real `RoundToDollar`;
2. each of **four broken versions** fails at least one of your cases (a
   surviving one is reported by the bug it contains);
3. case names are non-empty and unique.

## The rules RoundToDollar follows

- It rounds to the nearest multiple of 100 cents.
- Amounts exactly halfway round **away from zero**.

## Examples

```text
RoundToDollar(1249)  = 1200
RoundToDollar(1250)  = 1300
RoundToDollar(-1250) = -1300
RoundToDollar(-1249) = -1200
```

**Run** prints what the real function returns for each of your cases, next to
what the case expects.

## Constraints

- Don't change `RoundToDollar`. The grader uses its own copy of the bugs, not
  your code.
