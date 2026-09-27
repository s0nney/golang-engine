---
title: Correctness
quiz:
  - question: |
      This function should report whether a handle appears in a list. What's wrong
      with it?

      ```go
      func contains(handles []string, target string) bool {
      	for i := 1; i < len(handles); i++ {
      		if handles[i] == target {
      			return true
      		}
      	}
      	return false
      }
      ```
    options:
      - text: It returns `true` too early
      - text: It never checks the first handle
        correct: true
      - text: It panics on an empty slice
      - text: Nothing, it's correct
    explanation: |
      The loop starts at `1`, so `handles[0]` is never compared. `contains([]string{"ava"}, "ava")`
      returns `false`. Off-by-one bugs like this are why you test the edges:
      first element, last element, empty input.
  - question: Why can't a handful of passing tests *prove* an algorithm correct?
    options:
      - text: Tests only check the inputs you thought of, not every possible input
        correct: true
      - text: Go tests don't run the real code
      - text: Tests only measure speed
    explanation: |
      Tests show the algorithm works for the cases you tried. A bug can hide in an
      input you never tested. That's why you also reason about the algorithm
      itself: what's true before, during and after the loop.
---

An algorithm is **correct** if it produces the right output for *every* valid
input. Not most inputs. Not the ones in the demo. Every one. A fast algorithm
that's occasionally wrong is worse than useless at Clout, because a brand that
gets told a 45-follower account is its biggest creator will stop paying us.

## Edge cases

Most bugs live at the edges. For any algorithm over a slice, ask about:

- **The empty slice.** Does it panic? Return something sensible?
- **One element.** Is it handled, or does the loop skip it?
- **The first and last elements.** Off-by-one errors love these.
- **Duplicates.** If two influencers tie, does it still work?
- **Extremes.** Negative numbers, zero, huge values that could overflow.

Here's a sneaky one. This averages engagement across posts:

```go
func average(nums []int) int {
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return sum / len(nums)
}
```

Looks fine, until someone calls `average(nil)`. Integer division by zero
panics: `runtime error: integer divide by zero`. The fix is a guard clause:

```go
func average(nums []int) (int, bool) {
	if len(nums) == 0 {
		return 0, false
	}
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return sum / len(nums), true
}
```

## Reasoning with invariants

Tests can't try every input, so you also need to *reason* about correctness.
The main tool is a **loop invariant**: something that's true before the loop
starts and stays true after every iteration.

For `findMin`, the invariant is:

> After looking at the first `k` elements, `lowest` holds the smallest of them.

- **Before the loop**, `k = 1` and `lowest = nums[0]`. True.
- **Each iteration** compares one more element and keeps the smaller. Still true.
- **When the loop ends**, `k = len(nums)`, so `lowest` is the smallest of all of
  them. That's exactly what we wanted.

You won't write formal proofs day to day, but thinking "what stays true each
time round the loop?" catches a surprising number of bugs before they ship.

## Test the edges too

Reasoning and tests complement each other. A table-driven test makes the edge
cases explicit:

```go
func TestFindMin(t *testing.T) {
	tests := []struct {
		in     []int
		want   int
		wantOK bool
	}{
		{nil, 0, false},
		{[]int{7}, 7, true},
		{[]int{3, 1, 2}, 1, true},
		{[]int{5, 5, 5}, 5, true},
		{[]int{-4, 0, 9}, -4, true},
	}
	for _, tt := range tests {
		got, ok := findMin(tt.in)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("findMin(%v) = %d, %v; want %d, %v", tt.in, got, ok, tt.want, tt.wantOK)
		}
	}
}
```

Empty, single, normal, duplicates and negatives: five lines that cover most of
the ways a minimum-finder can break.
