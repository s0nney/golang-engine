---
title: 'Practice: Bubble and Insertion Sort'
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    // bubbleSort sorts s in place, in ascending order, by repeatedly
    // swapping out-of-order neighbours. Stop early if a pass makes no swaps.
    func bubbleSort[T cmp.Ordered](s []T) {
    	// ?
    }

    // insertionSort sorts s in place, in ascending order, by taking each
    // element and shifting it left into the sorted part of the slice.
    func insertionSort[T cmp.Ordered](s []T) {
    	// ?
    }

    func main() {
    	followers := []int{5000, 120, 88000, 950, 12}
    	bubbleSort(followers)
    	fmt.Println(followers) // want: [12 120 950 5000 88000]

    	handles := []string{"zoe", "ava", "mo", "bo"}
    	insertionSort(handles)
    	fmt.Println(handles) // want: [ava bo mo zoe]
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    func bubbleSort[T cmp.Ordered](s []T) {
    	for end := len(s); end > 1; end-- {
    		swapped := false
    		for i := 1; i < end; i++ {
    			if s[i-1] > s[i] {
    				s[i-1], s[i] = s[i], s[i-1]
    				swapped = true
    			}
    		}
    		if !swapped {
    			return
    		}
    	}
    }

    func insertionSort[T cmp.Ordered](s []T) {
    	for i := 1; i < len(s); i++ {
    		current := s[i]
    		j := i - 1
    		for j >= 0 && s[j] > current {
    			s[j+1] = s[j]
    			j--
    		}
    		s[j+1] = current
    	}
    }

    func main() {
    	followers := []int{5000, 120, 88000, 950, 12}
    	bubbleSort(followers)
    	fmt.Println(followers)

    	handles := []string{"zoe", "ava", "mo", "bo"}
    	insertionSort(handles)
    	fmt.Println(handles)
    }
  tests: |
    package main

    import (
    	"math/rand/v2"
    	"slices"
    	"testing"
    )

    func randomInputs() [][]int {
    	r := rand.New(rand.NewPCG(1, 2))
    	inputs := [][]int{nil, {}, {7}, {2, 1}, {1, 2, 3, 4}, {4, 3, 2, 1}, {5, 5, 1, 5}, {-3, 10, 0, -7}}
    	for range 100 {
    		s := make([]int, r.IntN(60))
    		for i := range s {
    			s[i] = r.IntN(200) - 50
    		}
    		inputs = append(inputs, s)
    	}
    	return inputs
    }

    func check(t *testing.T, name string, sort func([]int)) {
    	t.Helper()
    	for _, in := range randomInputs() {
    		got := slices.Clone(in)
    		sort(got)
    		want := slices.Clone(in)
    		slices.Sort(want)
    		if !slices.Equal(got, want) {
    			t.Fatalf("%s(%v) gave %v, want %v", name, in, got, want)
    		}
    	}
    }

    func TestBubbleSort(t *testing.T) {
    	check(t, "bubbleSort", bubbleSort[int])
    	s := []string{"zoe", "ava", "mo", "bo"}
    	bubbleSort(s)
    	if !slices.Equal(s, []string{"ava", "bo", "mo", "zoe"}) {
    		t.Errorf("bubbleSort of handles gave %v, want [ava bo mo zoe]", s)
    	}
    }

    func TestInsertionSort(t *testing.T) {
    	check(t, "insertionSort", insertionSort[int])
    	f := []float64{0.031, 0.007, 0.12, 0.045}
    	insertionSort(f)
    	if !slices.Equal(f, []float64{0.007, 0.031, 0.045, 0.12}) {
    		t.Errorf("insertionSort of rates gave %v, want [0.007 0.031 0.045 0.12]", f)
    	}
    }
---

Clout's leaderboard needs sorting, and before you trust `slices.Sort` you're
going to build two sorts yourself.

## Your task

Complete two generic, **in-place**, ascending sorts:

- `bubbleSort`: repeatedly walk the slice, swapping neighbours that are out of
  order. After each pass the largest remaining value has bubbled to the end, so
  the next pass can stop one earlier. If a pass makes **no swaps**, return
  early.
- `insertionSort`: for each element from index 1 onward, remember it, shift
  bigger elements in the sorted part one place right, then drop it into the
  gap.

Both take `[]T` where `T` is `cmp.Ordered`, so the same code sorts follower
counts, engagement rates and handles. Neither returns anything: the slice
shares its backing array with the caller, so sorting in place is visible
outside.

Watch the classic insertion-sort crash. The inner loop condition must check
the index *before* reading the slice:

```go
for j >= 0 && s[j] > current { // j >= 0 first!
```

Go's `&&` short-circuits, so `s[j]` is never evaluated once `j` is `-1`.

## How it's graded

The tests sort 100 random slices (fixed seed, including negatives and
duplicates) plus edge cases like `nil`, one element and reverse order, and
compare every result against `slices.Sort`.
