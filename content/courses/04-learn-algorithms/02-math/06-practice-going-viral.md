---
title: 'Practice: Going Viral'
exercise:
  starter: |
    package main

    import "fmt"

    // reach returns how many new viewers see a post at the given share level,
    // when every viewer shares it with `shares` more people: shares^level.
    // Level 0 is just the original poster, so reach(3, 0) is 1.
    func reach(shares, level int) int {
    	// ?
    	return 0
    }

    // levelsToReach returns the smallest level at which reach(shares, level)
    // is at least target. shares is always 2 or more. A target of 1 or less
    // is reached at level 0.
    func levelsToReach(shares, target int) int {
    	// ?
    	return 0
    }

    func main() {
    	fmt.Println(reach(3, 5))                  // want 243
    	fmt.Println(levelsToReach(3, 1_000_000))  // want 13
    	fmt.Println(levelsToReach(10, 1_000_000)) // want 6
    }
  solution: |
    package main

    import (
    	"fmt"
    	"math"
    )

    func reach(shares, level int) int {
    	total := 1
    	for range level {
    		total *= shares
    	}
    	return total
    }

    func levelsToReach(shares, target int) int {
    	level, current := 0, 1
    	for current < target {
    		if current > math.MaxInt/shares {
    			// The next multiplication would overflow, so it would certainly
    			// pass target (which is at most math.MaxInt).
    			return level + 1
    		}
    		current *= shares
    		level++
    	}
    	return level
    }

    func main() {
    	fmt.Println(reach(3, 5))
    	fmt.Println(levelsToReach(3, 1_000_000))
    	fmt.Println(levelsToReach(10, 1_000_000))
    }
  tests: |
    package main

    import (
    	"math"
    	"testing"
    	"time"
    )

    func TestReach(t *testing.T) {
    	tests := []struct{ shares, level, want int }{
    		{3, 0, 1},
    		{3, 1, 3},
    		{3, 5, 243},
    		{2, 10, 1024},
    		{10, 6, 1_000_000},
    		{2, 62, 1 << 62},
    	}
    	for _, tt := range tests {
    		if got := reach(tt.shares, tt.level); got != tt.want {
    			t.Errorf("reach(%d, %d) = %d, want %d", tt.shares, tt.level, got, tt.want)
    		}
    	}
    }

    func TestLevelsToReach(t *testing.T) {
    	tests := []struct{ shares, target, want int }{
    		{3, 1, 0},
    		{3, 0, 0},
    		{3, -5, 0},
    		{3, 3, 1},
    		{3, 4, 2},
    		{3, 243, 5},
    		{3, 244, 6},
    		{2, 1024, 10},
    		{2, 1025, 11},
    		{3, 1_000_000, 13},
    		{10, 1_000_000, 6},
    	}
    	for _, tt := range tests {
    		if got := levelsToReach(tt.shares, tt.target); got != tt.want {
    			t.Errorf("levelsToReach(%d, %d) = %d, want %d", tt.shares, tt.target, got, tt.want)
    		}
    	}
    }

    func TestLevelsToReachHuge(t *testing.T) {
    	tests := []struct{ shares, target, want int }{
    		{2, 1 << 62, 62},
    		{2, math.MaxInt, 63},
    		{10, math.MaxInt, 19},
    		{1000, math.MaxInt, 7},
    	}
    	for _, tt := range tests {
    		done := make(chan int, 1)
    		go func() { done <- levelsToReach(tt.shares, tt.target) }()
    		select {
    		case got := <-done:
    			if got != tt.want {
    				t.Errorf("levelsToReach(%d, %d) = %d, want %d (does current overflow?)", tt.shares, tt.target, got, tt.want)
    			}
    		case <-time.After(time.Second):
    			t.Fatalf("levelsToReach(%d, %d) didn't finish within a second: did current overflow and wrap around?", tt.shares, tt.target)
    		}
    	}
    }
---

The growth team wants two numbers on every post's analytics page:

> *How many new people see the post at share level N?*
>
> *How many levels of sharing until it reaches a million people?*

The first is an **exponent**. The second is its inverse, a **logarithm**. You
already know both ideas from this chapter; now you'll compute them with nothing
but multiplication and a loop.

## The model

Each viewer shares the post with `shares` more people. The original poster is
level 0, and each level multiplies the audience by `shares`:

| level | new viewers (shares = 3) |
|------:|-------------------------:|
| 0     | 1                        |
| 1     | 3                        |
| 2     | 9                        |
| 5     | 243                      |
| 13    | 1,594,323                |

So `reach(3, 5)` is 3⁵ = 243, and a million-viewer level with 3 shares each is
level 13: level 12 only reaches 531,441.

## Your task

Complete two functions.

**`reach(shares, level)`** returns shares^level, using repeated multiplication.
`reach(anything, 0)` is 1. (You could reach for `math.Pow`, but it works in
`float64`, which can't represent every large `int` exactly. Stick to integers.)

**`levelsToReach(shares, target)`** returns the **smallest** level whose reach is
at least `target`. That's the logarithm base `shares` of `target`, rounded up.
Start at level 0 with a reach of 1 and keep multiplying until you're there.
You can assume `shares` is at least 2. Any `target` of 1 or less is reached at
level 0.

## The overflow trap

Here's the catch: the tests include `levelsToReach(2, math.MaxInt)`. The answer
is 63, since 2⁶³ is just past the biggest `int`. But 2⁶³ itself doesn't fit!
Multiply 2⁶² by 2 and, as you saw in the exponents lesson, Go silently wraps
around to a negative number. That's still less than the target, so the loop
keeps going. Multiply again and you get 0, and from then on the loop spins
forever.

Check **before** you multiply. If `current > math.MaxInt/shares`, then
`current * shares` would overflow, which means it would be bigger than any
`int`, including `target`. So the next level is the answer: return it without
doing the multiplication. Integer division rounds down, which makes this check
safe: `current <= math.MaxInt/shares` guarantees `current * shares <= math.MaxInt`.

## Tips

- `for range level { ... }` runs the loop body `level` times (Go 1.22+).
- Remember to `import "math"` for `math.MaxInt`.
- A test that doesn't finish within a second fails with a hint about overflow,
  so an infinite loop won't leave you waiting.

**Run** prints three numbers with their expected values in comments.
**Submit** checks small and large powers, exact and in-between targets, and
targets right up at `math.MaxInt`.
