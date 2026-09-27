---
title: 'Practice: Binary Search'
exercise:
  starter: |
    package main

    import "fmt"

    // binarySearch looks for target in sorted, which is in ascending order.
    // If found, it returns target's index and true. If not, it returns the
    // index where target would be inserted to keep the slice sorted, and false.
    // It must do O(log n) comparisons, not scan the whole slice.
    func binarySearch(sorted []int, target int) (int, bool) {
    	// ?
    	return 0, false
    }

    func main() {
    	counts := []int{45, 312, 7100, 8200, 48210, 99000, 250000}
    	fmt.Println(binarySearch(counts, 48210)) // want: 4 true
    	fmt.Println(binarySearch(counts, 9000))  // want: 4 false
    	fmt.Println(binarySearch(counts, 1))     // want: 0 false
    }
  solution: |
    package main

    import "fmt"

    func binarySearch(sorted []int, target int) (int, bool) {
    	lo, hi := 0, len(sorted)
    	for lo < hi {
    		mid := lo + (hi-lo)/2
    		switch {
    		case sorted[mid] == target:
    			return mid, true
    		case sorted[mid] < target:
    			lo = mid + 1
    		default:
    			hi = mid
    		}
    	}
    	return lo, false
    }

    func main() {
    	counts := []int{45, 312, 7100, 8200, 48210, 99000, 250000}
    	fmt.Println(binarySearch(counts, 48210))
    	fmt.Println(binarySearch(counts, 9000))
    	fmt.Println(binarySearch(counts, 1))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"math/rand/v2"
    	"slices"
    	"testing"
    	"time"
    )

    func TestBinarySearchSmall(t *testing.T) {
    	counts := []int{45, 312, 7100, 8200, 48210, 99000, 250000}
    	tests := []struct {
    		target, want int
    		wantOK       bool
    	}{
    		{45, 0, true}, {250000, 6, true}, {48210, 4, true},
    		{1, 0, false}, {9000, 4, false}, {999999, 7, false},
    	}
    	for _, tt := range tests {
    		got, ok := binarySearch(counts, tt.target)
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("binarySearch(%v, %d) = %d, %v; want %d, %v", counts, tt.target, got, ok, tt.want, tt.wantOK)
    		}
    	}
    	if got, ok := binarySearch(nil, 5); got != 0 || ok {
    		t.Errorf("binarySearch(nil, 5) = %d, %v; want 0, false", got, ok)
    	}
    }

    func TestBinarySearchRandom(t *testing.T) {
    	r := rand.New(rand.NewPCG(4, 2))
    	for range 200 {
    		n := r.IntN(50)
    		seen := map[int]bool{}
    		var sorted []int
    		for len(sorted) < n {
    			v := r.IntN(1000)
    			if !seen[v] {
    				seen[v] = true
    				sorted = append(sorted, v)
    			}
    		}
    		slices.Sort(sorted)
    		target := r.IntN(1000)
    		want, wantOK := slices.BinarySearch(sorted, target)
    		got, ok := binarySearch(sorted, target)
    		if got != want || ok != wantOK {
    			t.Fatalf("binarySearch(%v, %d) = %d, %v; want %d, %v", sorted, target, got, ok, want, wantOK)
    		}
    	}
    }

    func TestBinarySearchIsLogarithmic(t *testing.T) {
    	sorted := make([]int, 1<<20)
    	for i := range sorted {
    		sorted[i] = i * 2
    	}
    	// Binary search needs about 20 steps per call, so 100,000 calls take
    	// milliseconds. A linear scan needs up to a million steps per call.
    	done := make(chan string, 1)
    	go func() {
    		for i := range 100_000 {
    			target := (i * 7919) % (1 << 21)
    			if _, ok := binarySearch(sorted, target); ok != (target%2 == 0) {
    				done <- fmt.Sprintf("binarySearch(evens up to 2^21, %d) found = %v, want %v", target, ok, target%2 == 0)
    				return
    			}
    		}
    		done <- ""
    	}()
    	select {
    	case msg := <-done:
    		if msg != "" {
    			t.Fatal(msg)
    		}
    	case <-time.After(time.Second):
    		t.Fatal("100,000 searches over 1,048,576 counts took over a second: are you scanning instead of halving?")
    	}
    }
---

A brand asks: "Do you have an influencer with *exactly* 48,210 followers?"
Clout keeps its follower counts sorted, so you can answer with **binary
search** in O(log n) instead of scanning all of them.

## Your task

Complete `binarySearch(sorted, target)`:

- If `target` is in `sorted`, return its index and `true`.
- If not, return the index where it **would be inserted** to keep the slice
  sorted, and `false`. For `[45 312 7100 8200 48210]`, searching for `9000`
  returns `4, false`, and searching for `1` returns `0, false`.
- It must be O(log n). The tests search a slice of over a million counts
  100,000 times, and a linear scan won't finish within the time limit.

The shape from the lesson is a good starting point:

```go
lo, hi := 0, len(sorted) // the half-open range [lo, hi)
for lo < hi {
	mid := lo + (hi-lo)/2
	// compare sorted[mid] with target and shrink the range
}
```

Hints:

- Use the half-open range `[lo, hi)`. When `sorted[mid] < target`, the answer
  is to the right, so `lo = mid + 1`. Otherwise it's at `mid` or to the left.
- When the loop ends, `lo` is the insertion point.
- Write `lo + (hi-lo)/2`, not `(lo+hi)/2`, to avoid overflow.

The tests compare your function with `slices.BinarySearch` on random sorted
slices. (The random inputs have no duplicates, so there's only one correct
index.)
