---
title: Cover Every Branch
difficulty: easy
after: coverage-and-tooling
hints:
  - 'Read `TransferFee` one `if` at a time and ask: what `amount` and `international` would make this condition true? Block 1 needs an amount that isn''t positive, block 2 needs `international: true`, and block 4 needs a fee over 2,000 cents.'
  - 'The fee is 1% of the amount (integer division, so it rounds down), plus 500 cents for international transfers, then clamped to between 25 and 2,000 cents. For $5,000 (`500000`) domestic that''s 5,000, clamped to 2,000.'
exercise:
  starter: |
    package main

    import "fmt"

    // feeCase is one row of the TransferFee test table.
    type feeCase struct {
    	name          string
    	amount        Cents
    	international bool
    	want          Cents
    }

    // feeCases must run every numbered block in TransferFee at least once,
    // and every want must be what the real TransferFee returns.
    var feeCases = []feeCase{
    	{name: "small domestic", amount: 1000, want: 25},
    	// Add cases for the blocks that never run. Press Run to see which.
    }

    // ---- Ledgerly code (already done, and correct) ----

    type Cents int64

    // coverage is what go test -cover adds to your code behind the scenes:
    // one counter per block, bumped every time the block runs.
    var coverage [5]int

    var blockNames = [5]string{
    	"block 0 (function entry)",
    	"block 1 (amount <= 0)",
    	"block 2 (international surcharge)",
    	"block 3 (minimum fee)",
    	"block 4 (maximum fee)",
    }

    // TransferFee returns the fee for moving amount out of an account:
    // 1% of the amount, plus $5 for international transfers, but never
    // less than $0.25 or more than $20. Non-positive amounts are free.
    func TransferFee(amount Cents, international bool) Cents {
    	coverage[0]++
    	if amount <= 0 {
    		coverage[1]++
    		return 0
    	}
    	fee := amount / 100
    	if international {
    		coverage[2]++
    		fee += 500
    	}
    	if fee < 25 {
    		coverage[3]++
    		fee = 25
    	}
    	if fee > 2000 {
    		coverage[4]++
    		fee = 2000
    	}
    	return fee
    }

    func main() {
    	coverage = [5]int{}
    	for _, tc := range feeCases {
    		got := TransferFee(tc.amount, tc.international)
    		status := "ok"
    		if got != tc.want {
    			status = fmt.Sprintf("WRONG, want %d", tc.want)
    		}
    		fmt.Printf("%-20s TransferFee(%d, %v) = %d  %s\n", tc.name, tc.amount, tc.international, got, status)
    	}
    	fmt.Println()
    	for i, n := range coverage {
    		fmt.Printf("%-36s ran %d times\n", blockNames[i], n)
    	}
    }
  solution: |
    package main

    import "fmt"

    // feeCase is one row of the TransferFee test table.
    type feeCase struct {
    	name          string
    	amount        Cents
    	international bool
    	want          Cents
    }

    // feeCases must run every numbered block in TransferFee at least once,
    // and every want must be what the real TransferFee returns.
    var feeCases = []feeCase{
    	{name: "small domestic", amount: 1000, want: 25},
    	{name: "typical domestic", amount: 12345, want: 123},
    	{name: "zero", amount: 0, want: 0},
    	{name: "negative", amount: -500, want: 0},
    	{name: "international", amount: 10000, international: true, want: 600},
    	{name: "huge domestic", amount: 500000, want: 2000},
    }

    // ---- Ledgerly code (already done, and correct) ----

    type Cents int64

    // coverage is what go test -cover adds to your code behind the scenes:
    // one counter per block, bumped every time the block runs.
    var coverage [5]int

    var blockNames = [5]string{
    	"block 0 (function entry)",
    	"block 1 (amount <= 0)",
    	"block 2 (international surcharge)",
    	"block 3 (minimum fee)",
    	"block 4 (maximum fee)",
    }

    // TransferFee returns the fee for moving amount out of an account:
    // 1% of the amount, plus $5 for international transfers, but never
    // less than $0.25 or more than $20. Non-positive amounts are free.
    func TransferFee(amount Cents, international bool) Cents {
    	coverage[0]++
    	if amount <= 0 {
    		coverage[1]++
    		return 0
    	}
    	fee := amount / 100
    	if international {
    		coverage[2]++
    		fee += 500
    	}
    	if fee < 25 {
    		coverage[3]++
    		fee = 25
    	}
    	if fee > 2000 {
    		coverage[4]++
    		fee = 2000
    	}
    	return fee
    }

    func main() {
    	coverage = [5]int{}
    	for _, tc := range feeCases {
    		got := TransferFee(tc.amount, tc.international)
    		status := "ok"
    		if got != tc.want {
    			status = fmt.Sprintf("WRONG, want %d", tc.want)
    		}
    		fmt.Printf("%-20s TransferFee(%d, %v) = %d  %s\n", tc.name, tc.amount, tc.international, got, status)
    	}
    	fmt.Println()
    	for i, n := range coverage {
    		fmt.Printf("%-36s ran %d times\n", blockNames[i], n)
    	}
    }
  tests: |
    package main

    import "testing"

    func TestCasesAreCorrect(t *testing.T) {
    	if len(feeCases) == 0 {
    		t.Fatal("feeCases is empty")
    	}
    	seen := map[string]bool{}
    	for i, tc := range feeCases {
    		if tc.name == "" {
    			t.Errorf("case %d has no name", i)
    		} else if seen[tc.name] {
    			t.Errorf("two cases are named %q; names must be unique", tc.name)
    		}
    		seen[tc.name] = true
    		if got := TransferFee(tc.amount, tc.international); got != tc.want {
    			t.Errorf("case %q: TransferFee(%d, %v) = %d, but the case wants %d (TransferFee is correct, so the case is wrong)",
    				tc.name, tc.amount, tc.international, got, tc.want)
    		}
    	}
    }

    func TestCasesCoverEveryBlock(t *testing.T) {
    	coverage = [5]int{}
    	for _, tc := range feeCases {
    		TransferFee(tc.amount, tc.international)
    	}
    	for i, n := range coverage {
    		if n == 0 {
    			t.Errorf("%s never runs: add a case that reaches it", blockNames[i])
    		}
    	}
    }

    // A fee function with its clamps swapped round must fail some case.
    func TestCasesCheckTheClamps(t *testing.T) {
    	noMax := func(amount Cents, international bool) Cents {
    		if amount <= 0 {
    			return 0
    		}
    		fee := amount / 100
    		if international {
    			fee += 500
    		}
    		return max(fee, 25)
    	}
    	for _, tc := range feeCases {
    		if noMax(tc.amount, tc.international) != tc.want {
    			return
    		}
    	}
    	t.Error("no case notices a TransferFee that forgets the $20 maximum; a case should hit the cap with a fee well over 2000")
    }
---

`go test -cover` works by rewriting your code before compiling it: it adds a
counter to every **block** (a run of statements with no branches in it) and
bumps it each time the block runs. The coverage percentage is simply "blocks
whose counter isn't zero" divided by "all blocks".

Ledgerly's `TransferFee` has been instrumented by hand the same way, with a
`coverage` array and one `coverage[i]++` per block. Your job is to write the
tests that light up every block.

Add cases to `feeCases` so that:

1. every case's `want` is what the real `TransferFee` returns (the function is
   correct, so a failing case means your expectation is wrong);
2. every one of the five blocks runs at least once;
3. case names are non-empty and unique.

**Run** prints each case's result and how many times each block ran.

## The rules TransferFee follows

- An amount of zero or less costs nothing.
- Otherwise the fee is 1% of the amount, rounded down (`amount / 100`).
- International transfers add $5.00 (500 cents).
- The fee is then clamped to at least $0.25 (25) and at most $20.00 (2000).

## Example

```text
small domestic       TransferFee(1000, false) = 25  ok
...
block 2 (international surcharge)    ran 0 times
```

## Constraints

- Coverage tells you what *ran*, not what was *checked*. The grader also
  makes sure a `TransferFee` without the $20 maximum fails one of your cases,
  so the cap case needs a fee that's clearly over the limit.
