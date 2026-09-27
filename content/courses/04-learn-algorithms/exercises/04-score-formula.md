---
title: Score Formula
difficulty: easy
after: stacks
hints:
  - 'Use a `[]int` as a stack. A number is pushed with `append`; an operator pops the top two values, combines them and pushes the result.'
  - 'Order matters for `-`: the value popped **first** is the right-hand side. For `"10 4 -"` you pop 4, then 10, and push `10 - 4`.'
  - 'It''s malformed if an operator finds fewer than two values, if a token is neither an operator nor a number (`strconv.Atoi` returns an error), or if the stack doesn''t hold exactly one value at the end.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // evalScore evaluates a score formula written in postfix notation
    // (operators come after their two operands), e.g. "3 4 + 2 *" is (3+4)*2.
    // Tokens are separated by single spaces. The operators are +, - and *.
    // It returns the result and true, or 0 and false if the formula is malformed.
    func evalScore(formula string) (int, bool) {
    	tokens := strings.Fields(formula)
    	// For each token:
    	//   a number   -> push it
    	//   an operator -> pop two values, push the result
    	// At the end exactly one value must be left.
    	_ = tokens
    	return 0, false
    }

    func main() {
    	fmt.Println(evalScore("3 4 + 2 *")) // want: 14 true
    	fmt.Println(evalScore("10 4 -"))    // want: 6 true
    	fmt.Println(evalScore("1 +"))       // want: 0 false
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strconv"
    	"strings"
    )

    func evalScore(formula string) (int, bool) {
    	var stack []int
    	for _, tok := range strings.Fields(formula) {
    		switch tok {
    		case "+", "-", "*":
    			if len(stack) < 2 {
    				return 0, false
    			}
    			a, b := stack[len(stack)-2], stack[len(stack)-1]
    			stack = stack[:len(stack)-2]
    			switch tok {
    			case "+":
    				stack = append(stack, a+b)
    			case "-":
    				stack = append(stack, a-b)
    			case "*":
    				stack = append(stack, a*b)
    			}
    		default:
    			n, err := strconv.Atoi(tok)
    			if err != nil {
    				return 0, false
    			}
    			stack = append(stack, n)
    		}
    	}
    	if len(stack) != 1 {
    		return 0, false
    	}
    	return stack[0], true
    }

    func main() {
    	fmt.Println(evalScore("3 4 + 2 *"))
    	fmt.Println(evalScore("10 4 -"))
    	fmt.Println(evalScore("1 +"))
    }
  tests: |
    package main

    import "testing"

    func TestEvalScore(t *testing.T) {
    	tests := []struct {
    		formula string
    		want    int
    		wantOK  bool
    	}{
    		{"3 4 + 2 *", 14, true},
    		{"10 4 -", 6, true},
    		{"4 10 -", -6, true},
    		{"42", 42, true},
    		{"-5 3 *", -15, true},
    		{"2 3 4 * +", 14, true},
    		{"100 2 * 30 3 * + 5 -", 285, true},
    		{"1 +", 0, false},
    		{"+", 0, false},
    		{"", 0, false},
    		{"1 2", 0, false},
    		{"1 2 /", 0, false},
    		{"likes 2 *", 0, false},
    	}
    	for _, tt := range tests {
    		got, ok := evalScore(tt.formula)
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("evalScore(%q) = %d, %v, want %d, %v", tt.formula, got, ok, tt.want, tt.wantOK)
    		}
    	}
    }
---

Clout lets brands define their own **engagement score** for a campaign. To keep
the parser simple, formulas are written in *postfix* notation: each operator
comes **after** its two operands, so no brackets are ever needed.

```
3 4 + 2 *     means (3 + 4) * 2 = 14
2 3 4 * +     means 2 + (3 * 4) = 14
10 4 -        means 10 - 4 = 6
```

A stack evaluates this in one left-to-right pass: push numbers, and when you
meet an operator, pop two values, apply it, and push the result.

Complete `evalScore(formula)`. Tokens are separated by spaces; each is an
integer (possibly negative, like `-5`) or one of `+`, `-`, `*`. Return the
result and `true`, or `0` and `false` if the formula is malformed:

- an operator doesn't have two values to work on (`"1 +"`),
- a token isn't a number or a known operator (`"1 2 /"`),
- or anything other than exactly one value is left at the end (`"1 2"`, `""`).

## Constraints

- Formulas have up to 10,000 tokens and all intermediate values fit in an `int`.
