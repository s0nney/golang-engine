---
title: 'Practice: Flatten the Loops'
exercise:
  starter: |
    package main

    import "fmt"

    // firstRepeat scans handles from left to right and returns the first handle
    // that has already been seen earlier in the slice, and true.
    // If every handle is unique it returns "" and false.
    // It must not modify handles, and it must run in O(n).
    func firstRepeat(handles []string) (string, bool) {
    	// ?
    	return "", false
    }

    // hasPairSum reports whether two different positions in sorted (ascending)
    // add up to exactly target. It must not modify sorted, and it must run in O(n).
    func hasPairSum(sorted []int, target int) bool {
    	// ?
    	return false
    }

    func main() {
    	signups := []string{"ava", "bo", "cy", "bo", "ava"}
    	fmt.Println(firstRepeat(signups)) // want: bo true

    	followers := []int{120, 450, 800, 1200, 3000}
    	fmt.Println(hasPairSum(followers, 2000)) // want: true (800 + 1200)
    	fmt.Println(hasPairSum(followers, 900))  // want: false
    }
  solution: |
    package main

    import "fmt"

    func firstRepeat(handles []string) (string, bool) {
    	seen := make(map[string]bool, len(handles))
    	for _, h := range handles {
    		if seen[h] {
    			return h, true
    		}
    		seen[h] = true
    	}
    	return "", false
    }

    func hasPairSum(sorted []int, target int) bool {
    	lo, hi := 0, len(sorted)-1
    	for lo < hi {
    		switch sum := sorted[lo] + sorted[hi]; {
    		case sum == target:
    			return true
    		case sum < target:
    			lo++ // need a bigger sum: move the small end up
    		default:
    			hi-- // need a smaller sum: move the big end down
    		}
    	}
    	return false
    }

    func main() {
    	signups := []string{"ava", "bo", "cy", "bo", "ava"}
    	fmt.Println(firstRepeat(signups))

    	followers := []int{120, 450, 800, 1200, 3000}
    	fmt.Println(hasPairSum(followers, 2000))
    	fmt.Println(hasPairSum(followers, 900))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"slices"
    	"testing"
    	"time"
    )

    func TestFirstRepeat(t *testing.T) {
    	tests := []struct {
    		in     []string
    		want   string
    		wantOK bool
    	}{
    		{nil, "", false},
    		{[]string{"ava"}, "", false},
    		{[]string{"ava", "bo", "cy"}, "", false},
    		{[]string{"ava", "ava"}, "ava", true},
    		{[]string{"ava", "bo", "cy", "bo", "ava"}, "bo", true},
    		{[]string{"cy", "ava", "bo", "ava", "cy"}, "ava", true},
    		{[]string{"zed", "amy", "zed", "amy"}, "zed", true},
    		{[]string{"", "x", ""}, "", true},
    	}
    	for _, tt := range tests {
    		in := slices.Clone(tt.in)
    		got, ok := firstRepeat(in)
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("firstRepeat(%q) = %q, %v; want %q, %v", tt.in, got, ok, tt.want, tt.wantOK)
    		}
    		if !slices.Equal(in, tt.in) {
    			t.Errorf("firstRepeat(%q) modified its input to %q", tt.in, in)
    		}
    	}
    }

    func TestHasPairSum(t *testing.T) {
    	tests := []struct {
    		in     []int
    		target int
    		want   bool
    	}{
    		{nil, 0, false},
    		{[]int{500}, 1000, false},
    		{[]int{500, 500}, 1000, true},
    		{[]int{120, 450, 800, 1200, 3000}, 2000, true},
    		{[]int{120, 450, 800, 1200, 3000}, 900, false},
    		{[]int{120, 450, 800, 1200, 3000}, 570, true},
    		{[]int{120, 450, 800, 1200, 3000}, 4200, true},
    		{[]int{120, 450, 800, 1200, 3000}, 240, false},
    		{[]int{-50, 10, 60}, 10, true},
    		{[]int{1, 2, 3, 4}, 8, false},
    	}
    	for _, tt := range tests {
    		in := slices.Clone(tt.in)
    		if got := hasPairSum(in, tt.target); got != tt.want {
    			t.Errorf("hasPairSum(%v, %d) = %v, want %v", tt.in, tt.target, got, tt.want)
    		}
    		if !slices.Equal(in, tt.in) {
    			t.Errorf("hasPairSum(%v, %d) modified its input to %v", tt.in, tt.target, in)
    		}
    	}
    }

    // within fails the test if f takes longer than a second.
    func within(t *testing.T, what string, f func()) {
    	t.Helper()
    	done := make(chan struct{})
    	go func() { f(); close(done) }()
    	select {
    	case <-done:
    	case <-time.After(time.Second):
    		t.Fatalf("%s took over a second on 200,000 items: is it O(n²)?", what)
    	}
    }

    func TestLinearTime(t *testing.T) {
    	const n = 200_000
    	handles := make([]string, n)
    	for i := range handles {
    		handles[i] = fmt.Sprintf("creator%d", i)
    	}
    	handles[n-1] = handles[n-2]
    	within(t, "firstRepeat", func() {
    		if got, ok := firstRepeat(handles); got != handles[n-2] || !ok {
    			t.Errorf("firstRepeat(200,000 handles) = %q, %v; want %q, true", got, ok, handles[n-2])
    		}
    	})

    	counts := make([]int, n)
    	for i := range counts {
    		counts[i] = 2 * i // all even
    	}
    	within(t, "hasPairSum", func() {
    		if hasPairSum(counts, 3) { // odd target: no pair of evens works
    			t.Errorf("hasPairSum(200,000 even numbers, 3) = true, want false")
    		}
    	})
    }
---

Clout's trust-and-safety team has two jobs that run over *every* new signup,
and there are hundreds of thousands of those a day. Someone wrote both with
nested loops, and the nightly batch now takes hours. Your job is to make them
**O(n)**.

## Job 1: the first repeated handle

Sign-up spam often reuses a handle. Given the handles in the order they arrived,
find the **first handle that shows up a second time**: scan left to right, and
stop at the first handle you've already seen.

```
ava  bo  cy  bo  ava
             ^^
             "bo" is the first one we've seen before
```

The O(n²) version compares each handle with every handle before it. The O(n)
version remembers what it's seen in a **map**, whose lookups and inserts are
O(1) on average:

```go
seen := make(map[string]bool)
```

Complete `firstRepeat(handles)` so it returns that handle and `true`, or `""` and
`false` if every handle is unique. Don't sort `handles` to find duplicates: that
would scramble the caller's slice, and sorting is O(n log n) anyway.

## Job 2: a pair that hits the target

A brand wants two creators whose follower counts add up to **exactly** its
target. The counts arrive **sorted ascending**. The nested-loop version tries
every pair. The O(n) version uses **two pointers**, one at each end:

- If `sorted[lo] + sorted[hi]` equals the target, you're done.
- If the sum is **too small**, the only way to grow it is a bigger small number,
  so move `lo` right.
- If the sum is **too big**, move `hi` left.
- Stop when the pointers meet. The two positions must be **different**, but the
  values may be equal (`[500 500]` makes 1000).

Each step moves one pointer inward, so there are at most n steps. This works
only *because* the slice is sorted, the same bargain binary search makes.

Complete `hasPairSum(sorted, target)` this way. It must not modify `sorted`.

## How it's graded

The tests check the behaviour on small tables (empty input, single items, ties,
negatives) and that neither function changes its input. Then they run each
function on **200,000** items with a one-second limit. That's around 20 billion
comparisons for a quadratic version, which fails the limit, but only a few
milliseconds for a linear one.

**Run** prints three results with the expected values in comments.
