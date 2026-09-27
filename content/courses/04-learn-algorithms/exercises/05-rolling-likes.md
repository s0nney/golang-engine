---
title: Rolling Likes
difficulty: easy
after: queues
hints:
  - 'Keep the last `k` values in a queue and a running `sum` next to it. Each new value is enqueued and added to `sum`.'
  - 'When the queue grows past `k` items, dequeue the oldest and subtract it from `sum`. Now each step is O(1) instead of re-adding `k` numbers.'
exercise:
  starter: |
    package main

    import "fmt"

    // rollingLikes returns, for each hour i, the total likes over the last k
    // hours up to and including hour i. Near the start, when fewer than k
    // hours have passed, it adds up all the hours so far. k is at least 1.
    func rollingLikes(likes []int, k int) []int {
    	sums := make([]int, 0, len(likes))
    	// Enqueue each hour's likes and keep a running sum.
    	// Once the queue holds more than k hours, dequeue the oldest.
    	return sums
    }

    func main() {
    	fmt.Println(rollingLikes([]int{5, 1, 3, 10, 0, 2}, 3)) // want [5 6 9 14 13 12]
    }
  solution: |
    package main

    import "fmt"

    func rollingLikes(likes []int, k int) []int {
    	sums := make([]int, 0, len(likes))
    	var window []int
    	sum := 0
    	for _, l := range likes {
    		window = append(window, l)
    		sum += l
    		if len(window) > k {
    			sum -= window[0]
    			window = window[1:]
    		}
    		sums = append(sums, sum)
    	}
    	return sums
    }

    func main() {
    	fmt.Println(rollingLikes([]int{5, 1, 3, 10, 0, 2}, 3))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    	"time"
    )

    func TestRollingLikes(t *testing.T) {
    	tests := []struct {
    		likes []int
    		k     int
    		want  []int
    	}{
    		{[]int{5, 1, 3, 10, 0, 2}, 3, []int{5, 6, 9, 14, 13, 12}},
    		{[]int{5, 1, 3}, 1, []int{5, 1, 3}},
    		{[]int{5, 1, 3}, 10, []int{5, 6, 9}},
    		{[]int{7}, 2, []int{7}},
    		{[]int{}, 3, []int{}},
    		{[]int{4, 4, 4, 4}, 2, []int{4, 8, 8, 8}},
    		{[]int{1, 2, 3, 4, 5}, 5, []int{1, 3, 6, 10, 15}},
    	}
    	for _, tt := range tests {
    		got := rollingLikes(tt.likes, tt.k)
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("rollingLikes(%v, %d) = %v, want %v", tt.likes, tt.k, got, tt.want)
    		}
    	}
    }

    func TestRollingLikesLarge(t *testing.T) {
    	n, k := 200_000, 50_000
    	likes := make([]int, n)
    	for i := range likes {
    		likes[i] = i % 7
    	}
    	start := time.Now()
    	got := rollingLikes(likes, k)
    	d := time.Since(start)
    	if len(got) != n {
    		t.Fatalf("rollingLikes(%d hours, %d) returned %d sums, want %d", n, k, len(got), n)
    	}
    	want := 0
    	for _, l := range likes[n-k:] {
    		want += l
    	}
    	if got[n-1] != want {
    		t.Errorf("rollingLikes(%d hours, %d): last sum = %d, want %d", n, k, got[n-1], want)
    	}
    	if d > time.Second {
    		t.Errorf("rollingLikes(%d hours, %d) took %v: keep a running sum instead of re-adding the window", n, k, d)
    	}
    }
---

A creator's dashboard shows a **rolling total**: for each hour, the likes they
got over the last `k` hours.

Complete `rollingLikes(likes, k)`. `likes[i]` is the number of likes in hour
`i`. Return a slice of the same length where entry `i` is the sum of
`likes[i-k+1] … likes[i]`. For the first few hours, when fewer than `k` hours
exist yet, add up everything so far.

## Example

```
rollingLikes([]int{5, 1, 3, 10, 0, 2}, 3)  // [5 6 9 14 13 12]
```

Hour 3's total is `1 + 3 + 10 = 14`; hour 0's is just `5`.

## Constraints

- `len(likes)` is up to 200,000 and `1 ≤ k`.
- Re-adding all `k` values every hour is O(n·k), which is 10 billion additions
  for the large test. A **queue** of the values in the window, plus a running
  sum, makes it O(n).
- An empty `likes` gives an empty (non-nil or nil, either is fine) slice.
