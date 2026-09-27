---
title: Duo Queue
difficulty: easy
after: hashmaps
hints:
  - 'Checking every pair is O(n²). Instead, walk the queue once. For the player at index `j`, the partner you need has rating `target - ratings[j]`. Has anyone **earlier** in the queue got that rating?'
  - 'A `map[int]int` from rating to the **first** index where it appeared answers that in O(1). Look up the partner before you add `ratings[j]` to the map (so nobody pairs with themselves), and only store a rating the first time you see it.'
exercise:
  starter: |
    package main

    import "fmt"

    // duoPair finds two players in the queue whose ratings add up to target.
    // It returns their indexes i < j, choosing the smallest possible j and,
    // for that j, the smallest i. If there's no such pair, ok is false.
    func duoPair(ratings []int, target int) (i, j int, ok bool) {
    	// Walk the queue once. For each player, look up the rating their
    	// partner would need in a map of ratings seen so far.
    	return 0, 0, false
    }

    func main() {
    	queue := []int{1200, 1750, 900, 1500, 1600}
    	fmt.Println(duoPair(queue, 2700)) // want 0 3 true
    	fmt.Println(duoPair(queue, 100))  // want 0 0 false
    }
  solution: |
    package main

    import "fmt"

    func duoPair(ratings []int, target int) (i, j int, ok bool) {
    	first := map[int]int{} // rating -> first index it appeared at
    	for j, r := range ratings {
    		if i, found := first[target-r]; found {
    			return i, j, true
    		}
    		if _, seen := first[r]; !seen {
    			first[r] = j
    		}
    	}
    	return 0, 0, false
    }

    func main() {
    	queue := []int{1200, 1750, 900, 1500, 1600}
    	fmt.Println(duoPair(queue, 2700))
    	fmt.Println(duoPair(queue, 100))
    }
  tests: |
    package main

    import (
    	"testing"
    	"time"
    )

    func TestDuoPair(t *testing.T) {
    	tests := []struct {
    		ratings []int
    		target  int
    		i, j    int
    		ok      bool
    	}{
    		{nil, 1000, 0, 0, false},
    		{[]int{500}, 1000, 0, 0, false}, // a player can't duo with themselves
    		{[]int{500, 500}, 1000, 0, 1, true},
    		{[]int{1200, 1750, 900, 1500, 1600}, 2700, 0, 3, true},
    		{[]int{1200, 1750, 900, 1500, 1600}, 100, 0, 0, false},
    		{[]int{1200, 1750, 900, 1500, 1600}, 3350, 1, 4, true},
    		{[]int{10, 25, 20, 20, 30}, 40, 2, 3, true},  // j = 3 beats 0, 4
    		{[]int{7, 7, 7, 3}, 10, 0, 3, true},          // smallest i for that j
    		{[]int{-50, 400, 250, 150}, 200, 0, 2, true}, // negative ratings are allowed
    		{[]int{0, 3, 0}, 0, 0, 2, true},
    	}
    	for _, tt := range tests {
    		i, j, ok := duoPair(tt.ratings, tt.target)
    		if i != tt.i || j != tt.j || ok != tt.ok {
    			t.Errorf("duoPair(%v, %d) = %d, %d, %v, want %d, %d, %v", tt.ratings, tt.target, i, j, ok, tt.i, tt.j, tt.ok)
    		}
    	}
    }

    func TestDuoPairLarge(t *testing.T) {
    	ratings := make([]int, 100_000)
    	for k := range ratings {
    		ratings[k] = 2 * k // all even
    	}
    	start := time.Now()
    	i, j, ok := duoPair(ratings, 7) // odd target: no pair exists
    	if ok {
    		t.Errorf("duoPair(100,000 even ratings, 7) = %d, %d, true, want 0, 0, false", i, j)
    	}
    	ratings[99_999] = 7 - ratings[31_415]
    	i, j, ok = duoPair(ratings, 7)
    	if i != 31_415 || j != 99_999 || !ok {
    		t.Errorf("duoPair(100,000 ratings, 7) = %d, %d, %v, want 31415, 99999, true", i, j, ok)
    	}
    	if d := time.Since(start); d > time.Second {
    		t.Errorf("duoPair on 100,000 ratings took %v: use a map instead of checking every pair", d)
    	}
    }
---

The duo queue pairs up two players whose ratings **add up** to the target for
the current bracket.

Complete `duoPair(ratings, target)`. `ratings` holds the queue in join order.
Return indexes `i < j` with `ratings[i] + ratings[j] == target`, and `true`.

If several pairs work, pick the one whose **second** player joined earliest (the
smallest `j`), and for that `j` the smallest `i`. If no pair adds up, return
`0, 0, false`. A player can't pair with themselves.

## Examples

```
duoPair([]int{1200, 1750, 900, 1500, 1600}, 2700) // 0, 3, true   (1200 + 1500)
duoPair([]int{10, 25, 20, 20, 30}, 40)            // 2, 3, true   (20 + 20)
duoPair([]int{1200, 1750, 900}, 100)              // 0, 0, false
```

In the second example, `10 + 30` also makes 40, but that pair ends at `j = 4`. The
pair ending at `j = 3` wins.

## Constraints

- Up to 100,000 players; ratings may be negative or repeated.
- Aim for **O(n)** with a map. One test uses 100,000 players with no valid pair,
  where checking every pair (about 5 billion of them) is far too slow.
