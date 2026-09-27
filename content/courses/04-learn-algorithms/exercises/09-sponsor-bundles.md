---
title: Sponsor Bundles
difficulty: medium
after: exponential-time
hints:
  - 'Think recursively about deal `i`: either you **skip** it, or you **take** it and have `budget - prices[i]` left for deals `i+1` onwards. The answer is the sum of both.'
  - 'Base cases: a remaining budget of exactly 0 is one bundle (take nothing more). A negative budget, or running out of deals with money left over, is 0.'
  - 'The plain recursion is O(2ⁿ). The only things that change between calls are `i` and the remaining budget, so memoize on that pair: a `map[[2]int]int` works as a key.'
exercise:
  starter: |
    package main

    import "fmt"

    func countBundles(prices []int, budget int) int {
    	return 0
    }

    func main() {
    	fmt.Println(countBundles([]int{200, 350, 150, 500, 50}, 500)) // want 2
    	fmt.Println(countBundles([]int{5, 5, 5}, 10))                 // want 3
    }
  solution: |
    package main

    import "fmt"

    func countBundles(prices []int, budget int) int {
    	memo := map[[2]int]int{}
    	var count func(i, left int) int
    	count = func(i, left int) int {
    		if left == 0 {
    			return 1
    		}
    		if left < 0 || i == len(prices) {
    			return 0
    		}
    		key := [2]int{i, left}
    		if v, ok := memo[key]; ok {
    			return v
    		}
    		v := count(i+1, left) + count(i+1, left-prices[i])
    		memo[key] = v
    		return v
    	}
    	return count(0, budget)
    }

    func main() {
    	fmt.Println(countBundles([]int{200, 350, 150, 500, 50}, 500))
    	fmt.Println(countBundles([]int{5, 5, 5}, 10))
    }
  tests: |
    package main

    import (
    	"testing"
    	"time"
    )

    func TestCountBundles(t *testing.T) {
    	tests := []struct {
    		prices []int
    		budget int
    		want   int
    	}{
    		{[]int{200, 350, 150, 500, 50}, 500, 2},
    		{[]int{5, 5, 5}, 10, 3},
    		{[]int{5, 5, 5}, 15, 1},
    		{[]int{5, 5, 5}, 0, 1},
    		{[]int{}, 0, 1},
    		{nil, 30, 0},
    		{[]int{10, 20}, -10, 0},
    		{[]int{7}, 7, 1},
    		{[]int{7}, 8, 0},
    		{[]int{1, 2, 3, 4, 5}, 5, 3},
    		{[]int{3, 3, 3, 3}, 6, 6},
    	}
    	for _, tt := range tests {
    		if got := countBundles(tt.prices, tt.budget); got != tt.want {
    			t.Errorf("countBundles(%v, %d) = %d, want %d", tt.prices, tt.budget, got, tt.want)
    		}
    	}
    }

    func TestCountBundlesManyDeals(t *testing.T) {
    	oneToSixty := make([]int, 60)
    	for i := range oneToSixty {
    		oneToSixty[i] = i + 1
    	}
    	mixed := make([]int, 80)
    	for i := range mixed {
    		mixed[i] = (i*37)%97 + 3
    	}
    	tests := []struct {
    		name   string
    		prices []int
    		budget int
    		want   int
    	}{
    		{"prices 1..60", oneToSixty, 100, 437209},
    		{"prices 1..60", oneToSixty, 915, 3360682669655028},
    		{"prices 1..60", oneToSixty, 1830, 1},
    		{"80 mixed prices", mixed, 500, 4073754741128},
    	}
    	for _, tt := range tests {
    		done := make(chan int, 1)
    		go func() { done <- countBundles(tt.prices, tt.budget) }()
    		select {
    		case got := <-done:
    			if got != tt.want {
    				t.Errorf("countBundles(%s, %d) = %d, want %d", tt.name, tt.budget, got, tt.want)
    			}
    		case <-time.After(2 * time.Second):
    			t.Fatalf("countBundles(%s, %d) took over 2 seconds: memoize on (deal index, budget left)", tt.name, tt.budget)
    		}
    	}
    }
---

Brands post **sponsorship deals** on Clout, each with a price in dollars. A
creator with an exact `budget` of ad slots to fill wants to know how many
different **bundles** of deals add up to exactly that budget.

Write `countBundles(prices, budget)`. A bundle is a set of deals: each deal is
used at most once, and order doesn't matter. Two deals with the same price are
still different deals (from different brands).

## Examples

```
countBundles([]int{200, 350, 150, 500, 50}, 500)  // 2: {500} and {350, 150}
countBundles([]int{5, 5, 5}, 10)                  // 3: any two of the three deals
countBundles([]int{5, 5, 5}, 0)                   // 1: the empty bundle
countBundles([]int{10, 20}, -10)                  // 0
```

## Constraints

- Up to 80 deals, each priced from 1 to 100; `budget` is at most 2,000 (and may
  be 0 or negative).
- The answer always fits in an `int`.
- Trying every subset is O(2ⁿ): for 60 deals that's 10¹⁸ subsets, far beyond
  the time limit. The tests give each call two seconds.
