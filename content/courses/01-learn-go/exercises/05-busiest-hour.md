---
title: Busiest Hour
difficulty: easy
after: arrays-and-slices
hints:
  - 'Deal with the empty slice first: there''s no busiest hour, so return `-1`.'
  - 'Remember the index of the best hour so far, starting with `0`. Range over the slice and switch to a new index only when its count is **strictly** bigger, so earlier hours win ties.'
  - 'Or let the `slices` package do it: `slices.Max` finds the biggest count and `slices.Index` finds where it first appears.'
exercise:
  starter: |
    package main

    import "fmt"

    // busiestHour returns the index of the largest count in counts.
    // If several hours tie, it returns the earliest one.
    // If counts is empty, it returns -1.
    func busiestHour(counts []int) int {
    	// ?
    	return 0
    }

    func main() {
    	sent := []int{12, 40, 95, 95, 30, 8}
    	fmt.Println(busiestHour(sent)) // want: 2
    }
  solution: |
    package main

    import "fmt"

    func busiestHour(counts []int) int {
    	if len(counts) == 0 {
    		return -1
    	}
    	best := 0
    	for i, c := range counts {
    		if c > counts[best] {
    			best = i
    		}
    	}
    	return best
    }

    func main() {
    	sent := []int{12, 40, 95, 95, 30, 8}
    	fmt.Println(busiestHour(sent))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestBusiestHour(t *testing.T) {
    	tests := []struct {
    		counts []int
    		want   int
    	}{
    		{[]int{12, 40, 95, 95, 30, 8}, 2},
    		{nil, -1},
    		{[]int{}, -1},
    		{[]int{7}, 0},
    		{[]int{1, 2, 3, 4}, 3},
    		{[]int{9, 2, 3}, 0},
    		{[]int{0, 0, 0}, 0},
    		{[]int{5, 1, 5}, 0},
    		{[]int{-4, -2, -9}, 1},
    	}
    	for _, tt := range tests {
    		before := slices.Clone(tt.counts)
    		if got := busiestHour(tt.counts); got != tt.want {
    			t.Errorf("busiestHour(%v) = %d, want %d", before, got, tt.want)
    		}
    		if !slices.Equal(tt.counts, before) {
    			t.Errorf("busiestHour changed its input from %v to %v; leave the slice as it is", before, tt.counts)
    		}
    	}
    }
---

Textio's dashboard highlights the busiest hour of the day so customers know
when their audience is most active. `counts[i]` is the number of messages sent
during hour `i`.

Complete `busiestHour(counts)`. It returns the **index** of the largest count.
If several hours share the largest count, return the **earliest** one. If
`counts` is empty, return `-1`.

## Examples

```
busiestHour([]int{12, 40, 95, 95, 30, 8})  // 2   (hours 2 and 3 tie, 2 is earlier)
busiestHour([]int{-4, -2, -9})             // 1
busiestHour(nil)                           // -1
```

## Constraints

- `counts` can be empty or `nil`, and may hold zeros or (after a data glitch)
  negative numbers.
- Don't sort or change `counts`: the index matters.
