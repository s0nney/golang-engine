---
title: Login Streak
difficulty: medium
after: hashmaps
hints:
  - 'Put every day into a set (`map[int]bool` or `map[int]struct{}`). Duplicates disappear, and "did they log in on day d?" becomes an O(1) question.'
  - 'A day `d` **starts** a streak only if `d - 1` is not in the set. From each start, count upwards while `d + 1`, `d + 2`, ... are in the set. Days that aren''t starts are skipped, so every day is counted at most twice in total: O(n).'
  - 'Map iteration order is random, so apply the tie rule explicitly: replace the best streak if the new one is longer, **or** equally long and starts on an earlier day.'
exercise:
  starter: |
    package main

    import "fmt"

    func longestStreak(days []int) (start, length int) {
    	return 0, 0
    }

    func main() {
    	days := []int{12, 3, 10, 4, 11, 2, 13, 3, 20}
    	fmt.Println(longestStreak(days)) // want 10 4
    }
  solution: |
    package main

    import "fmt"

    func longestStreak(days []int) (start, length int) {
    	seen := make(map[int]bool, len(days))
    	for _, d := range days {
    		seen[d] = true
    	}
    	for d := range seen {
    		if seen[d-1] {
    			continue // not the first day of a streak
    		}
    		n := 1
    		for seen[d+n] {
    			n++
    		}
    		if n > length || n == length && d < start {
    			start, length = d, n
    		}
    	}
    	return start, length
    }

    func main() {
    	days := []int{12, 3, 10, 4, 11, 2, 13, 3, 20}
    	fmt.Println(longestStreak(days))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    	"time"
    )

    func TestLongestStreak(t *testing.T) {
    	tests := []struct {
    		name          string
    		days          []int
    		start, length int
    	}{
    		{"no logins", nil, 0, 0},
    		{"one login", []int{42}, 42, 1},
    		{"same day twice", []int{7, 7, 7}, 7, 1},
    		{"example", []int{12, 3, 10, 4, 11, 2, 13, 3, 20}, 10, 4},
    		{"already sorted", []int{1, 2, 3, 5, 6}, 1, 3},
    		{"reverse order", []int{9, 8, 7, 6, 1}, 6, 4},
    		{"tie: earlier start wins", []int{30, 31, 5, 6, 100}, 5, 2},
    		{"tie of singles", []int{40, 20, 60}, 20, 1},
    		{"duplicates inside a streak", []int{4, 5, 5, 6, 4, 7, 6}, 4, 4},
    		{"negative days", []int{-3, -1, -2, 0, 5}, -3, 4},
    		{"gap of one", []int{1, 3, 5, 7}, 1, 1},
    	}
    	for _, tt := range tests {
    		in := slices.Clone(tt.days)
    		// Repeat: map iteration order changes between runs.
    		for range 20 {
    			start, length := longestStreak(in)
    			if start != tt.start || length != tt.length {
    				t.Errorf("%s: longestStreak(%v) = %d, %d, want %d, %d", tt.name, tt.days, start, length, tt.start, tt.length)
    				break
    			}
    		}
    		if !slices.Equal(in, tt.days) {
    			t.Errorf("%s: longestStreak changed its input to %v", tt.name, in)
    		}
    	}
    }

    func TestLongestStreakLarge(t *testing.T) {
    	// 200,000 logins: days 0..99,999 except every 10,000th day, shuffled
    	// deterministically, then every day again. Plus one long run far away.
    	var days []int
    	for d := range 100_000 {
    		if d%10_000 != 0 {
    			days = append(days, (d*7919)%100_000)
    		}
    	}
    	days = append(days, days...)
    	for d := 500_000; d < 512_000; d++ {
    		days = append(days, d)
    	}
    	start := time.Now()
    	s, l := longestStreak(days)
    	if s != 500_000 || l != 12_000 {
    		t.Errorf("longestStreak(%d logins) = %d, %d, want 500000, 12000", len(days), s, l)
    	}
    	if d := time.Since(start); d > time.Second {
    		t.Errorf("longestStreak(%d logins) took %v: only count upward from days that start a streak", len(days), d)
    	}
    }
---

The rewards team wants to celebrate each player's **longest login streak**: the
most consecutive days they logged in.

Write `longestStreak(days)`. `days` holds the day numbers of every login, in no
particular order, and a player can log in several times on the same day. Return
the first day and the length of the longest run of consecutive days. If several
runs are equally long, return the one that starts **earliest**. With no logins,
return `0, 0`.

Don't modify `days`.

## Example

```go
longestStreak([]int{12, 3, 10, 4, 11, 2, 13, 3, 20}) // 10, 4
```

The distinct days are 2, 3, 4, 10, 11, 12, 13 and 20. That's a 3-day streak from
day 2, a 4-day streak from day 10, and a 1-day streak on day 20.

```go
longestStreak([]int{30, 31, 5, 6, 100}) // 5, 2 (ties with 30-31; 5 is earlier)
```

## Constraints

- Up to 200,000 logins. Day numbers can be negative (days before launch).
- Aim for **O(n)** with a hashmap. Sorting a copy (O(n log n)) also works.
  Counting a full streak from **every** login does not: that's O(n²) on one long streak.
