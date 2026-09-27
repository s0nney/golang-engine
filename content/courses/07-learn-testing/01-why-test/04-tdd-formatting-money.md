---
title: 'Your Turn: Formatting Money'
quiz:
  - question: |
      You're TDD-ing `Cents.String`. The only test so far is
      `Cents(1234).String() == "$12.34"`. Which implementation is the best
      *green* step?
    options:
      - text: A full implementation with thousands separators, negatives and currency codes
      - text: '`return fmt.Sprintf("$%d.%02d", c/100, c%100)`'
        correct: true
      - text: Delete the test until the implementation is done
    explanation: |
      Green means the *simplest* thing that passes the tests you have. Commas
      and negative numbers get added when a test demands them. That keeps
      each step small and every line of code justified by a test.
  - question: |
      What does this print?

      ```go
      n := int64(-5)
      fmt.Println(n/100, n%100)
      ```
    options:
      - text: '`-1 95`'
      - text: '`0 -5`'
        correct: true
      - text: '`0 5`'
      - text: '`-0 -5`'
    explanation: |
      Go's integer division truncates towards zero, so `-5/100` is `0`, and
      `%` takes the sign of the dividend, so `-5%100` is `-5`. That's why a
      naive `"$%d.%02d"` prints `$0.-5` for negative amounts. Handle the sign
      first, then format the absolute value.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    )

    // Cents is an amount of money in cents. $12.34 is Cents(1234).
    type Cents int64

    // String formats c as dollars, e.g. "$1,234.56" or "-$0.05".
    func (c Cents) String() string {
    	// ?
    	return fmt.Sprintf("%d", int64(c))
    }

    func main() {
    	for _, c := range []Cents{1234, 5, -5, 100000, 123456789} {
    		fmt.Println(c)
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strconv"
    )

    // Cents is an amount of money in cents. $12.34 is Cents(1234).
    type Cents int64

    // String formats c as dollars, e.g. "$1,234.56" or "-$0.05".
    func (c Cents) String() string {
    	sign := ""
    	n := int64(c)
    	if n < 0 {
    		sign = "-"
    		n = -n
    	}
    	return fmt.Sprintf("%s$%s.%02d", sign, groupThousands(n/100), n%100)
    }

    // groupThousands formats n with a comma between each group of three digits.
    func groupThousands(n int64) string {
    	s := strconv.FormatInt(n, 10)
    	for i := len(s) - 3; i > 0; i -= 3 {
    		s = s[:i] + "," + s[i:]
    	}
    	return s
    }

    func main() {
    	for _, c := range []Cents{1234, 5, -5, 100000, 123456789} {
    		fmt.Println(c)
    	}
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    )

    func TestStringBasics(t *testing.T) {
    	for _, tt := range []struct {
    		c    Cents
    		want string
    	}{
    		{1234, "$12.34"},
    		{100, "$1.00"},
    		{0, "$0.00"},
    		{5, "$0.05"},
    		{99, "$0.99"},
    		{70, "$0.70"},
    	} {
    		if got := tt.c.String(); got != tt.want {
    			t.Errorf("Cents(%d).String() = %q, want %q", int64(tt.c), got, tt.want)
    		}
    	}
    }

    func TestStringNegative(t *testing.T) {
    	for _, tt := range []struct {
    		c    Cents
    		want string
    	}{
    		{-5, "-$0.05"},
    		{-1234, "-$12.34"},
    		{-100, "-$1.00"},
    	} {
    		if got := tt.c.String(); got != tt.want {
    			t.Errorf("Cents(%d).String() = %q, want %q", int64(tt.c), got, tt.want)
    		}
    	}
    }

    func TestStringThousands(t *testing.T) {
    	for _, tt := range []struct {
    		c    Cents
    		want string
    	}{
    		{99999, "$999.99"},
    		{100000, "$1,000.00"},
    		{123456789, "$1,234,567.89"},
    		{12345678, "$123,456.78"},
    		{-100000000, "-$1,000,000.00"},
    	} {
    		if got := tt.c.String(); got != tt.want {
    			t.Errorf("Cents(%d).String() = %q, want %q", int64(tt.c), got, tt.want)
    		}
    	}
    }

    func TestStringUsedByFmt(t *testing.T) {
    	if got := fmt.Sprintf("balance: %v", Cents(4250)); got != "balance: $42.50" {
    		t.Errorf(`fmt.Sprintf("balance: %%v", Cents(4250)) = %q, want "balance: $42.50"`, got)
    	}
    }
---

Time to drive a feature with tests yourself. Ledgerly needs to show money to humans, so `Cents` needs a `String` method. Because it's a `String() string` method, `Cents` satisfies `fmt.Stringer` and `fmt.Println(c)` will use it automatically.

## The tests are already written

Here's the red. The grader runs these checks, grouped the way you'd grow them in TDD:

| Test | Input | Want |
| --- | --- | --- |
| basics | `1234`, `100`, `0`, `5`, `99`, `70` | `$12.34`, `$1.00`, `$0.00`, `$0.05`, `$0.99`, `$0.70` |
| negatives | `-5`, `-1234`, `-100` | `-$0.05`, `-$12.34`, `-$1.00` |
| thousands | `99999`, `100000`, `123456789` | `$999.99`, `$1,000.00`, `$1,234,567.89` |
| via fmt | `fmt.Sprintf("balance: %v", Cents(4250))` | `balance: $42.50` |

Don't try to solve them all at once. Work like the TDD loop:

1. Make the **basics** pass with the simplest format string you can think of. Remember `%02d` pads with zeros.
2. Submit and watch the **negatives** fail. (The quiz above shows why.) Deal with the sign up front, then format the positive number.
3. Add **thousands** separators to the dollars part. One approach: format the dollars with `strconv.FormatInt`, then walk backwards from the end inserting a `,` every three digits.
4. Refactor. Could the comma logic be its own small function?

## Your task

Complete `Cents.String` so it formats cents as dollars with a leading `$`, a `-` *before* the `$` for negative amounts, exactly two decimal places, and commas between groups of three digits in the dollar part.

Press **Run** to see what your `main` prints, and **Submit** to run the tests.

## A bug we're leaving in

There's one input that the obvious solution gets wrong, and none of these tests check it. Can you guess what it is? Hint: think about the smallest possible `int64`. We'll let a fuzzer find it for us in the fuzzing chapter.
