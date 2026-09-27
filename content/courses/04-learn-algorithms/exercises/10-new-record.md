---
title: Days to a New Record
difficulty: medium
after: stacks
hints:
  - 'The nested-loop version scans forward from every day: O(n²). On a steadily falling account that''s 20 billion steps for 200,000 days.'
  - 'Keep a **stack of day indexes** that are still waiting for a higher count. Their counts only go down from bottom to top (never up), so the top is always the easiest one to beat.'
  - 'For each day `i`: while the stack isn''t empty and `counts[i]` beats the count of the day on top, pop that day `j` and set `wait[j] = i - j`. Then push `i`. Each day is pushed and popped at most once, so it''s O(n).'
exercise:
  starter: |
    package main

    import "fmt"

    func daysToRecord(counts []int) []int {
    	return nil
    }

    func main() {
    	fmt.Println(daysToRecord([]int{70, 71, 69, 68, 72, 72, 65})) // want [1 3 2 1 0 0 0]
    }
  solution: |
    package main

    import "fmt"

    func daysToRecord(counts []int) []int {
    	wait := make([]int, len(counts))
    	var waiting []int // stack of day indexes
    	for i, c := range counts {
    		for len(waiting) > 0 && c > counts[waiting[len(waiting)-1]] {
    			j := waiting[len(waiting)-1]
    			waiting = waiting[:len(waiting)-1]
    			wait[j] = i - j
    		}
    		waiting = append(waiting, i)
    	}
    	return wait
    }

    func main() {
    	fmt.Println(daysToRecord([]int{70, 71, 69, 68, 72, 72, 65}))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    	"time"
    )

    func TestDaysToRecord(t *testing.T) {
    	tests := []struct {
    		counts, want []int
    	}{
    		{[]int{70, 71, 69, 68, 72, 72, 65}, []int{1, 3, 2, 1, 0, 0, 0}},
    		{[]int{}, []int{}},
    		{[]int{42}, []int{0}},
    		{[]int{1, 2, 3, 4}, []int{1, 1, 1, 0}},
    		{[]int{4, 3, 2, 1}, []int{0, 0, 0, 0}},
    		{[]int{5, 5, 5, 6}, []int{3, 2, 1, 0}},
    		{[]int{-3, -5, -1, -2, 0}, []int{2, 1, 2, 1, 0}},
    		{[]int{10, 1, 1, 1, 11}, []int{4, 3, 2, 1, 0}},
    	}
    	for _, tt := range tests {
    		got := daysToRecord(tt.counts)
    		if len(got) != len(tt.counts) || !slices.Equal(got, tt.want) {
    			t.Errorf("daysToRecord(%v) = %v, want %v", tt.counts, got, tt.want)
    		}
    	}
    }

    func TestDaysToRecordLarge(t *testing.T) {
    	n := 200_000
    	counts := make([]int, n)
    	for i := range counts {
    		counts[i] = n - i // falls every day...
    	}
    	counts[n-1] = n + 1 // ...until a huge jump on the last day
    	done := make(chan []int, 1)
    	go func() { done <- daysToRecord(counts) }()
    	select {
    	case got := <-done:
    		if len(got) != n || got[0] != n-1 || got[n/2] != n-1-n/2 || got[n-1] != 0 {
    			t.Errorf("daysToRecord(%d falling days, then a jump) gave the wrong answer", n)
    		}
    	case <-time.After(time.Second):
    		t.Fatalf("daysToRecord(%d days) took over a second: it needs to be O(n), not O(n²)", n)
    	}
    }
---

Clout sends a push notification when an account sets a **new follower record**.
For its analytics, the team wants to know, for every day, how long the account
had to wait until its follower count went **strictly higher** than that day's.

Write `daysToRecord(counts)`. `counts[i]` is the follower count on day `i`.
Return a slice `wait` of the same length, where `wait[i]` is the number of days
from day `i` to the next day with a strictly higher count, or `0` if that never
happens.

## Example

```
counts: [70 71 69 68 72 72 65]
wait:   [ 1  3  2  1  0  0  0]
```

Day 1 (71) waits until day 4 (72), which is 3 days. Days 4 and 5 (72) are never
beaten, and an equal count doesn't count as higher.

## Constraints

- Up to 200,000 days. Counts can be any `int`, including negative (the tests
  sometimes use follower *changes*).
- The large test runs under a one-second limit, so O(n²) won't make it.
