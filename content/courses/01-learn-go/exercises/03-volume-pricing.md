---
title: Volume Pricing
difficulty: easy
after: conditionals
hints:
  - 'Handle the "nothing sent" case first with an early return: if `messages <= 0`, return `0`.'
  - 'A `switch` with no value (or an `if` / `else if` chain) picks the price per message. Check the boundaries carefully: `1000` is still 3 cents, `1001` is 2 cents.'
  - 'Compute `total := messages * price`, then, if `nonprofit` is true, halve it with integer division (`total / 2` rounds down).'
exercise:
  starter: |
    package main

    import "fmt"

    // monthlyCharge returns a customer's bill in cents.
    //
    //	messages <= 0        -> 0
    //	1 to 1,000           -> 3 cents per message
    //	1,001 to 10,000      -> 2 cents per message
    //	more than 10,000     -> 1 cent per message
    //
    // Nonprofits pay half, rounded down to a whole cent.
    func monthlyCharge(messages int, nonprofit bool) int {
    	// ?
    	return 0
    }

    func main() {
    	fmt.Println(monthlyCharge(500, false))   // want: 1500
    	fmt.Println(monthlyCharge(5000, true))   // want: 5000
    	fmt.Println(monthlyCharge(20000, false)) // want: 20000
    }
  solution: |
    package main

    import "fmt"

    func monthlyCharge(messages int, nonprofit bool) int {
    	if messages <= 0 {
    		return 0
    	}
    	var price int
    	switch {
    	case messages <= 1000:
    		price = 3
    	case messages <= 10000:
    		price = 2
    	default:
    		price = 1
    	}
    	total := messages * price
    	if nonprofit {
    		total /= 2
    	}
    	return total
    }

    func main() {
    	fmt.Println(monthlyCharge(500, false))
    	fmt.Println(monthlyCharge(5000, true))
    	fmt.Println(monthlyCharge(20000, false))
    }
  tests: |
    package main

    import "testing"

    func TestMonthlyCharge(t *testing.T) {
    	tests := []struct {
    		messages  int
    		nonprofit bool
    		want      int
    	}{
    		{500, false, 1500},
    		{5000, true, 5000},
    		{20000, false, 20000},
    		{0, false, 0},
    		{-3, false, 0},
    		{-3, true, 0},
    		{1, false, 3},
    		{1, true, 1},
    		{1000, false, 3000},
    		{1001, false, 2002},
    		{10000, false, 20000},
    		{10001, false, 10001},
    		{10001, true, 5000},
    		{333, true, 499},
    	}
    	for _, tt := range tests {
    		if got := monthlyCharge(tt.messages, tt.nonprofit); got != tt.want {
    			t.Errorf("monthlyCharge(%d, %v) = %d, want %d", tt.messages, tt.nonprofit, got, tt.want)
    		}
    	}
    }
---

Textio rewards big senders: the more messages a customer sends in a month,
the cheaper **every** message gets. Nonprofits also get a 50% discount.

Complete `monthlyCharge(messages, nonprofit)`. It returns the monthly bill in
**cents**:

| Messages this month | Price per message |
| --- | --- |
| 0 or fewer | nothing to pay |
| 1 to 1,000 | 3 cents |
| 1,001 to 10,000 | 2 cents |
| more than 10,000 | 1 cent |

The price applies to all of the month's messages, not just the ones above a
boundary. If `nonprofit` is `true`, the bill is halved, rounding down to a
whole cent.

## Examples

```
monthlyCharge(500, false)    // 1500  (500 × 3)
monthlyCharge(1001, false)   // 2002  (1001 × 2)
monthlyCharge(333, true)     // 499   (999 / 2, rounded down)
monthlyCharge(-3, false)     // 0
```

## Constraints

- `messages` can be any `int`, including 0 and negative numbers (a billing
  glitch), which cost nothing.
