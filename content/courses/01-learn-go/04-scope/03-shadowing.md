---
title: Shadowing
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	plan := "free"
      	premium := true
      	if premium {
      		plan := "pro"
      		fmt.Print(plan, " ")
      	}
      	fmt.Println(plan)
      }
      ```
    options:
      - text: '`pro pro`'
      - text: '`free free`'
      - text: '`pro free`'
        correct: true
      - text: It doesn't compile, because `plan` is declared twice
    explanation: |
      Inside the `if`, `plan := "pro"` declares a brand-new variable that
      shadows the outer one. It prints `pro`, then disappears at the `}`. The
      outer `plan` was never changed, so the last line prints `free`.
  - question: How do you fix the bug in the previous question so that `plan` really becomes `"pro"`?
    options:
      - text: Change `plan := "pro"` to `plan = "pro"`
        correct: true
      - text: Move `fmt.Println(plan)` above the `if`
      - text: Change `plan := "free"` to `var plan = "free"`
    explanation: |
      `=` assigns to the existing variable from the outer scope, while `:=`
      creates a new one. Using `=` updates the outer `plan`.
exercise:
  starter: |
    package main

    import "fmt"

    // finalCost returns the cost in cents of a monthly plan.
    // Premium customers get 20% off.
    func finalCost(cost int, isPremium bool) int {
    	total := cost
    	if isPremium {
    		total := cost * 80 / 100
    		fmt.Println("premium discount applied:", total)
    	}
    	return total
    }

    func main() {
    	fmt.Println(finalCost(1000, true))  // want 800
    	fmt.Println(finalCost(1000, false)) // want 1000
    }
  solution: |
    package main

    import "fmt"

    func finalCost(cost int, isPremium bool) int {
    	total := cost
    	if isPremium {
    		total = cost * 80 / 100
    		fmt.Println("premium discount applied:", total)
    	}
    	return total
    }

    func main() {
    	fmt.Println(finalCost(1000, true))
    	fmt.Println(finalCost(1000, false))
    }
  tests: |
    package main

    import "testing"

    func TestFinalCost(t *testing.T) {
    	tests := []struct {
    		cost      int
    		isPremium bool
    		want      int
    	}{
    		{1000, true, 800},
    		{1000, false, 1000},
    		{500, true, 400},
    		{0, true, 0},
    	}
    	for _, tt := range tests {
    		if got := finalCost(tt.cost, tt.isPremium); got != tt.want {
    			t.Errorf("finalCost(%d, %v) = %d, want %d", tt.cost, tt.isPremium, got, tt.want)
    		}
    	}
    }
---

Here's a bug that has bitten nearly every Go programmer at least once. Look carefully:

```go
package main

import "fmt"

func main() {
	status := "queued"
	delivered := true

	if delivered {
		status := "delivered"
		fmt.Println("inside:", status)
	}

	fmt.Println("outside:", status)
}
```

```text
inside: delivered
outside: queued
```

The message was delivered, but `status` still says `"queued"`. What happened?

## A new variable with the same name

Inside the `if` block, `status := "delivered"` doesn't *change* the outer `status`. The `:=` **declares a brand-new variable** that happens to have the same name. Inside that block, the new `status` hides, or **shadows**, the outer one. When the block ends, the inner variable disappears, and the untouched outer one is visible again.

Remember: `:=` in an **inner** block is allowed even if the name exists outside. It's only an error to redeclare a name in the **same** block.

The fix is to *assign* instead of *declare*:

```go
if delivered {
	status = "delivered" // = updates the outer variable
}
```

## The sneaky version: multiple return values

The most common real-world shadowing bug involves `:=` with multiple return values. When at least one variable on the left is new, `:=` is allowed, and it reuses the existing ones **in the same scope**. But in an inner block, it declares *all* of them fresh:

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	var credits int
	input := "25"

	if input != "" {
		credits, err := strconv.Atoi(input) // new credits AND new err!
		fmt.Println("parsed:", credits, err)
	}

	fmt.Println("credits:", credits)
}
```

```text
parsed: 25 <nil>
credits: 0
```

The parsed value went into a new, inner `credits` that vanished at the `}`. The fix is to declare `err` beforehand and use `=`:

```go
var err error
if input != "" {
	credits, err = strconv.Atoi(input)
}
```

## Shadowing built-in names

You can even shadow names from the universe scope, like `len`, `max` or `string`. Nothing stops you, but it's confusing:

```go
func main() {
	len := 3               // shadows the built-in len function
	fmt.Println(len("hi")) // error: cannot call len (variable of type int)
}
```

Avoid naming variables after built-ins such as `len`, `cap`, `new`, `min`, `max`, `string` or `copy`.

## Spotting shadowing

- Be suspicious whenever you see `:=` inside an `if`, `for` or `switch` using a name that already exists outside.
- Linters (automatic code checkers) can flag shadowing for you. The `shadow` analyzer, used by many teams, reports exactly this pattern.
- If a value you "set" seems to vanish, look for a stray `:=`.

Shadowing isn't always a bug. You'll see deliberate shadowing in idiomatic Go, like `err` being redeclared in many small `if` statements. The danger is shadowing *by accident*.

## Your turn

Premium customers are complaining: they're being charged full price! The program even prints "premium discount applied", so what's going on?

Find and fix the bug in `finalCost` so premium customers get their 20% discount. It's a one-character fix.
