---
title: 'Practice: Merge and Quick Sort'
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    // mergeSort returns a NEW sorted slice containing the elements of s,
    // in ascending order. It must not modify s.
    func mergeSort[T cmp.Ordered](s []T) []T {
    	// ? split in half, sort each half, then merge them
    	return s
    }

    // merge combines two sorted slices into one sorted slice.
    func merge[T cmp.Ordered](left, right []T) []T {
    	// ?
    	return nil
    }

    // quickSort sorts s in place, in ascending order.
    func quickSort[T cmp.Ordered](s []T) {
    	if len(s) <= 1 {
    		return
    	}
    	p := partition(s)
    	quickSort(s[:p])
    	quickSort(s[p+1:])
    }

    // partition uses the last element as the pivot. It moves every element
    // smaller than the pivot to its left, puts the pivot straight after them,
    // and returns the pivot's final index.
    func partition[T cmp.Ordered](s []T) int {
    	// ?
    	return 0
    }

    func main() {
    	followers := []int{38, 27, 43, 3, 9, 82, 10}
    	fmt.Println(mergeSort(followers)) // want: [3 9 10 27 38 43 82]
    	fmt.Println(followers)            // want: unchanged, [38 27 43 3 9 82 10]

    	quickSort(followers)
    	fmt.Println(followers) // want: [3 9 10 27 38 43 82]
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    func mergeSort[T cmp.Ordered](s []T) []T {
    	if len(s) <= 1 {
    		return append([]T(nil), s...)
    	}
    	mid := len(s) / 2
    	return merge(mergeSort(s[:mid]), mergeSort(s[mid:]))
    }

    func merge[T cmp.Ordered](left, right []T) []T {
    	result := make([]T, 0, len(left)+len(right))
    	i, j := 0, 0
    	for i < len(left) && j < len(right) {
    		if left[i] <= right[j] {
    			result = append(result, left[i])
    			i++
    		} else {
    			result = append(result, right[j])
    			j++
    		}
    	}
    	result = append(result, left[i:]...)
    	return append(result, right[j:]...)
    }

    func quickSort[T cmp.Ordered](s []T) {
    	if len(s) <= 1 {
    		return
    	}
    	p := partition(s)
    	quickSort(s[:p])
    	quickSort(s[p+1:])
    }

    func partition[T cmp.Ordered](s []T) int {
    	last := len(s) - 1
    	pivot := s[last]
    	i := 0
    	for j := range last {
    		if s[j] < pivot {
    			s[i], s[j] = s[j], s[i]
    			i++
    		}
    	}
    	s[i], s[last] = s[last], s[i]
    	return i
    }

    func main() {
    	followers := []int{38, 27, 43, 3, 9, 82, 10}
    	fmt.Println(mergeSort(followers))
    	fmt.Println(followers)

    	quickSort(followers)
    	fmt.Println(followers)
    }
  tests: |
    package main

    import (
    	"math/rand/v2"
    	"slices"
    	"testing"
    )

    func randomInputs() [][]int {
    	r := rand.New(rand.NewPCG(3, 4))
    	inputs := [][]int{nil, {}, {7}, {2, 1}, {1, 2, 3, 4}, {4, 3, 2, 1}, {5, 5, 1, 5}, {-3, 10, 0, -7}}
    	for range 100 {
    		s := make([]int, r.IntN(300))
    		for i := range s {
    			s[i] = r.IntN(1000) - 200
    		}
    		inputs = append(inputs, s)
    	}
    	return inputs
    }

    func TestMergeSort(t *testing.T) {
    	for _, in := range randomInputs() {
    		orig := slices.Clone(in)
    		got := mergeSort(in)
    		want := slices.Clone(in)
    		slices.Sort(want)
    		if !slices.Equal(got, want) {
    			t.Fatalf("mergeSort(%v) = %v, want %v", orig, got, want)
    		}
    		if !slices.Equal(in, orig) {
    			t.Fatalf("mergeSort changed its input: it was %v, now %v", orig, in)
    		}
    	}
    }

    func TestMerge(t *testing.T) {
    	got := merge([]int{3, 27, 38}, []int{9, 10, 82})
    	if want := []int{3, 9, 10, 27, 38, 82}; !slices.Equal(got, want) {
    		t.Errorf("merge([3 27 38], [9 10 82]) = %v, want %v", got, want)
    	}
    }

    func TestPartition(t *testing.T) {
    	s := []int{9, 2, 7, 4, 5}
    	p := partition(s)
    	if p != 2 || s[2] != 5 {
    		t.Fatalf("partition([9 2 7 4 5]) returned %d leaving %v; want index 2 with the pivot 5 there", p, s)
    	}
    	for i, v := range s {
    		if (i < p && v >= 5) || (i > p && v < 5) {
    			t.Fatalf("partition([9 2 7 4 5]) left %v: everything before index %d must be < 5 and everything after >= 5", s, p)
    		}
    	}
    }

    func TestQuickSort(t *testing.T) {
    	for _, in := range randomInputs() {
    		got := slices.Clone(in)
    		quickSort(got)
    		want := slices.Clone(in)
    		slices.Sort(want)
    		if !slices.Equal(got, want) {
    			t.Fatalf("quickSort(%v) gave %v, want %v", in, got, want)
    		}
    	}
    	h := []string{"zoe", "ava", "mo", "bo"}
    	quickSort(h)
    	if !slices.Equal(h, []string{"ava", "bo", "mo", "zoe"}) {
    		t.Errorf("quickSort of handles gave %v, want [ava bo mo zoe]", h)
    	}
    }
---

Now the fast ones. Both merge sort and quick sort are O(n log n) on average,
fast enough for Clout's full 10-million-influencer leaderboard.

## Part 1: merge sort

Complete `mergeSort` and its helper `merge`:

- `merge(left, right)` takes two **sorted** slices and returns one sorted slice
  with all their elements. Walk both with two indexes, repeatedly taking the
  smaller front element, then append whatever is left over. Use `<=` so ties
  come from `left` first, which keeps the sort stable.
- `mergeSort(s)` returns a **new** sorted slice. Slices of length 0 or 1 are
  already sorted. Otherwise, split at the middle, `mergeSort` each half, and
  `merge` the results.

`mergeSort` must **not modify its input**. The tests check! Be careful with the
base case: returning `s` itself hands the caller a slice that shares memory
with their input. Returning a copy is safer:

```go
if len(s) <= 1 {
	return append([]T(nil), s...)
}
```

## Part 2: quick sort

`quickSort` is written for you. Complete `partition` using the Lomuto scheme:
the last element is the pivot, a boundary `i` starts at 0, and every element
smaller than the pivot is swapped to position `i` (then `i++`). Finally, swap
the pivot into position `i` and return `i`.

For `[9 2 7 4 5]`, `partition` must return `2` and leave `5` at index 2, with
only smaller values before it.

## How it's graded

Both sorts run on 100 random slices of up to 300 elements (fixed seed) and are
compared against `slices.Sort`.
