---
title: 'Practice: Promo Pack Orders'
exercise:
  starter: |
    package main

    import "fmt"

    // packOrders returns how many different ordered sequences of promo packs
    // (worth 1, 2 or 5 thousand followers each) add up to exactly k thousand.
    // packOrders(0) is 1 (buy nothing) and packOrders of a negative k is 0.
    func packOrders(k int) int {
    	// ?
    	return 0
    }

    func main() {
    	fmt.Println(packOrders(3))  // want 3
    	fmt.Println(packOrders(5))  // want 9
    	fmt.Println(packOrders(40)) // want 1142898376
    }
  solution: |
    package main

    import "fmt"

    func packOrders(k int) int {
    	memo := map[int]int{}
    	var count func(k int) int
    	count = func(k int) int {
    		if k < 0 {
    			return 0
    		}
    		if k == 0 {
    			return 1
    		}
    		if v, ok := memo[k]; ok {
    			return v
    		}
    		v := count(k-1) + count(k-2) + count(k-5)
    		memo[k] = v
    		return v
    	}
    	return count(k)
    }

    func main() {
    	fmt.Println(packOrders(3))
    	fmt.Println(packOrders(5))
    	fmt.Println(packOrders(40))
    }
  tests: |
    package main

    import (
    	"testing"
    	"time"
    )

    func TestPackOrdersSmall(t *testing.T) {
    	tests := []struct{ k, want int }{
    		{-3, 0}, {-1, 0}, {0, 1}, {1, 1}, {2, 2}, {3, 3}, {4, 5}, {5, 9}, {6, 15}, {10, 128}, {20, 26547},
    	}
    	for _, tt := range tests {
    		if got := packOrders(tt.k); got != tt.want {
    			t.Errorf("packOrders(%d) = %d, want %d", tt.k, got, tt.want)
    		}
    	}
    }

    func TestPackOrdersFast(t *testing.T) {
    	tests := []struct{ k, want int }{
    		{40, 1142898376},
    		{60, 49203950608909},
    		{80, 2118323734061751820},
    	}
    	for _, tt := range tests {
    		done := make(chan int, 1)
    		go func() { done <- packOrders(tt.k) }()
    		select {
    		case got := <-done:
    			if got != tt.want {
    				t.Errorf("packOrders(%d) = %d, want %d", tt.k, got, tt.want)
    			}
    		case <-time.After(time.Second):
    			t.Fatalf("packOrders(%d) took over a second: is it memoized?", tt.k)
    		}
    	}
    }

    func TestPackOrdersRepeatable(t *testing.T) {
    	// Calling it again must give the same answer: a memo that leaks
    	// between calls must still be correct.
    	for range 3 {
    		if got := packOrders(30); got != packOrders(30) || got != 5508222 {
    			t.Fatalf("packOrders(30) = %d, want 5508222 every time", got)
    		}
    	}
    }
---

Clout sells **promo packs** that boost an account's followers. There are three
sizes: **1k**, **2k** and **5k** followers. Marketing wants to show a fun stat
on the checkout page:

> *There are 9 different ways to reach 5k new followers!*

Order matters here, because packs are delivered one after another. Buying a 1k
pack then a 2k pack is a different order from 2k then 1k. For 3k there are three
orders:

```
1 + 1 + 1
1 + 2
2 + 1
```

## The recurrence

Think about the **first** pack in an order. It's either a 1k, a 2k or a 5k pack,
and whatever follows it must make up the rest. So:

```
packOrders(k) = packOrders(k-1) + packOrders(k-2) + packOrders(k-5)
```

with two base cases:

- `packOrders(0) = 1`: there's exactly one way to add nothing, which is to buy
  nothing. (This is what makes "a single 5k pack" count as one order.)
- `packOrders(k) = 0` for negative `k`: you overshot, so that order doesn't count.

Check it for 3k: `packOrders(2) + packOrders(1) + packOrders(-2)` = 2 + 1 + 0 = 3.

## The problem

Written straight from the recurrence, this is the Fibonacci trap all over again,
only worse: every call makes **three** recursive calls, and the same `k` values
get recomputed over and over. `packOrders(40)` makes about 2.5 billion calls
and takes several seconds. The checkout page needs `packOrders(80)`.

## Your task

Complete `packOrders(k)` so that it follows the recurrence **and** memoizes it,
so each value of `k` is computed only once. Any of the approaches from the
memoization lesson works:

- a `map[int]int` memo (the closure pattern, with `var count func(int) int`
  declared first so it can call itself, keeps the memo out of sight), or
- bottom-up: fill a slice from 0 up to `k`, where each entry adds up the entries
  1, 2 and 5 places before it (treating positions before 0 as 0).

Either way it becomes O(k).

The answers grow exponentially, roughly 1.7 times per extra thousand. `packOrders(80)`
is about 2.1 × 10¹⁸, which still fits in an `int`. Anything past 83 would
overflow, so the tests stop at 80.

## How it's graded

The tests check small values (including 0 and negative `k`), then call
`packOrders` with 40, 60 and 80 under a one-second limit, so an unmemoized
version fails with a hint. A final test calls it several times in a row, to
make sure a memo shared between calls doesn't give wrong answers.
