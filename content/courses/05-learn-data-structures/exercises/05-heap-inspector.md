---
title: Heap Inspector
difficulty: easy
after: heaps-and-priority-queues
hints:
  - 'In a slice-backed heap, the parent of index `i` is at `(i - 1) / 2`. The root (index 0) has no parent, so start checking at index 1.'
  - 'Loop `i` from 1 to `len(h) - 1` and return the first `i` where `h[i] < h[(i-1)/2]`. Equal values are fine in a min-heap. If the loop finishes, return -1.'
exercise:
  starter: |
    package main

    import "fmt"

    // firstViolation checks whether h is a valid min-heap. It returns the
    // smallest index i whose value is less than its parent's value, or -1
    // if every element is >= its parent.
    func firstViolation(h []int) int {
    	// For each index except the root, compare h[i] with its parent.
    	return 0
    }

    func main() {
    	fmt.Println(firstViolation([]int{3, 5, 4, 9, 6, 8})) // want -1
    	fmt.Println(firstViolation([]int{3, 5, 4, 9, 2, 8})) // want 4
    }
  solution: |
    package main

    import "fmt"

    func firstViolation(h []int) int {
    	for i := 1; i < len(h); i++ {
    		if h[i] < h[(i-1)/2] {
    			return i
    		}
    	}
    	return -1
    }

    func main() {
    	fmt.Println(firstViolation([]int{3, 5, 4, 9, 6, 8}))
    	fmt.Println(firstViolation([]int{3, 5, 4, 9, 2, 8}))
    }
  tests: |
    package main

    import "testing"

    func TestFirstViolation(t *testing.T) {
    	tests := []struct {
    		h    []int
    		want int
    	}{
    		{nil, -1},
    		{[]int{42}, -1},
    		{[]int{3, 5, 4, 9, 6, 8}, -1},
    		{[]int{3, 5, 4, 9, 2, 8}, 4},     // 2 is under 5
    		{[]int{5, 3}, 1},                 // the root must be the smallest
    		{[]int{1, 1, 1, 1}, -1},          // ties are allowed
    		{[]int{1, 2, 3, 4, 5, 6, 7}, -1}, // a sorted slice is always a min-heap
    		{[]int{1, 5, 2, 6, 7, 3, 1}, 6},  // 1 is under 2 (index 2), not under the root
    		{[]int{1, 5, 2, 4, 7, 3, 8}, 3},  // 4 is under 5
    		{[]int{10, 20, 30, 25, 15}, 4},
    		{[]int{2, 8, 3, 9, 9, 3, 4, 7}, 7}, // 7 is under 9 at index 3
    		{[]int{-5, -3, -4, -3}, -1},
    	}
    	for _, tt := range tests {
    		if got := firstViolation(tt.h); got != tt.want {
    			t.Errorf("firstViolation(%v) = %d, want %d", tt.h, got, tt.want)
    		}
    	}
    }
---

The respawn queue is a **min-heap** of respawn times stored in a slice, so the
next player to respawn is always at index 0. After a bad deploy, some servers'
queues came back scrambled. Before the queue is used, a health check should
find the first spot where the heap rule is broken.

Complete `firstViolation(h)`. In a valid min-heap every element is **greater
than or equal to** its parent. Return the smallest index whose value is
**smaller** than its parent's, or `-1` if `h` is a valid min-heap. An empty
slice is a valid heap.

## Examples

```
       3                          3
     /   \                      /   \
    5     4                    5     4
   / \   /                    / \   /
  9   6 8                    9  (2) 8

firstViolation([]int{3, 5, 4, 9, 6, 8}) // -1
firstViolation([]int{3, 5, 4, 9, 2, 8}) // 4: h[4] = 2 is smaller than its parent h[1] = 5
```

## Constraints

- O(n): compare each element with its parent once.
- Don't modify `h`.
