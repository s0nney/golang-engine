---
title: 'Practice: Slices Without Surprises'
exercise:
  starter: |
    package main

    import "fmt"

    type Influencer struct {
    	Handle    string
    	Followers int
    }

    // followerCounts returns the follower count of every influencer, in order.
    // Allocate the result ONCE with exactly the right capacity.
    func followerCounts(infs []Influencer) []int {
    	var counts []int
    	for _, inf := range infs {
    		counts = append(counts, inf.Followers)
    	}
    	return counts
    }

    // withGuest returns a new line-up: base plus guest on the end.
    // Bug: two calls on the same base can overwrite each other's guest.
    // It must never share memory with base (or with another call's result).
    func withGuest(base []string, guest string) []string {
    	return append(base, guest)
    }

    func main() {
    	infs := []Influencer{{"ava", 90_000}, {"bo", 1_200}, {"cy", 450_000}, {"dee", 30_000}, {"eve", 800}}
    	counts := followerCounts(infs)
    	fmt.Println(counts, "len", len(counts), "cap", cap(counts)) // want cap 5

    	base := make([]string, 2, 4)
    	base[0], base[1] = "ava", "bo"
    	a := withGuest(base, "cy")
    	b := withGuest(base, "dee")
    	fmt.Println(a, b) // want: [ava bo cy] [ava bo dee]
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    )

    type Influencer struct {
    	Handle    string
    	Followers int
    }

    func followerCounts(infs []Influencer) []int {
    	counts := make([]int, 0, len(infs))
    	for _, inf := range infs {
    		counts = append(counts, inf.Followers)
    	}
    	return counts
    }

    func withGuest(base []string, guest string) []string {
    	return append(slices.Clone(base), guest)
    }

    func main() {
    	infs := []Influencer{{"ava", 90_000}, {"bo", 1_200}, {"cy", 450_000}, {"dee", 30_000}, {"eve", 800}}
    	counts := followerCounts(infs)
    	fmt.Println(counts, "len", len(counts), "cap", cap(counts))

    	base := make([]string, 2, 4)
    	base[0], base[1] = "ava", "bo"
    	a := withGuest(base, "cy")
    	b := withGuest(base, "dee")
    	fmt.Println(a, b)
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestFollowerCounts(t *testing.T) {
    	for n := range 40 {
    		infs := make([]Influencer, n)
    		for i := range infs {
    			infs[i] = Influencer{Handle: "h", Followers: i * 10}
    		}
    		got := followerCounts(infs)
    		if len(got) != n {
    			t.Fatalf("followerCounts of %d influencers returned %d counts", n, len(got))
    		}
    		for i, c := range got {
    			if c != i*10 {
    				t.Fatalf("followerCounts(...)[%d] = %d, want %d", i, c, i*10)
    			}
    		}
    		if cap(got) != n {
    			t.Fatalf("followerCounts of %d influencers has cap %d, want exactly %d: preallocate with make", n, cap(got), n)
    		}
    	}
    	infs := []Influencer{{"ava", 1}, {"bo", 2}, {"cy", 3}, {"dee", 4}, {"eve", 5}}
    	if allocs := testing.AllocsPerRun(100, func() { followerCounts(infs) }); allocs > 1 {
    		t.Errorf("followerCounts made %.0f allocations per call, want 1", allocs)
    	}
    }

    func TestWithGuest(t *testing.T) {
    	base := make([]string, 2, 4)
    	base[0], base[1] = "ava", "bo"
    	a := withGuest(base, "cy")
    	b := withGuest(base, "dee")
    	if !slices.Equal(a, []string{"ava", "bo", "cy"}) {
    		t.Errorf("first withGuest(base, \"cy\") is now %v, want [ava bo cy]: did the second call overwrite it?", a)
    	}
    	if !slices.Equal(b, []string{"ava", "bo", "dee"}) {
    		t.Errorf("withGuest(base, \"dee\") = %v, want [ava bo dee]", b)
    	}
    	a[0] = "zed"
    	if base[0] != "ava" {
    		t.Errorf("changing the result of withGuest changed base to %v: the result must not share memory with base", base)
    	}
    	if got := withGuest(nil, "solo"); !slices.Equal(got, []string{"solo"}) {
    		t.Errorf("withGuest(nil, \"solo\") = %v, want [solo]", got)
    	}
    }
---

Two small functions from Clout's campaign builder, each with a slice problem
you learned to spot in this chapter.

## followerCounts: preallocate

`followerCounts` works, but it starts from a nil slice and lets `append` grow
it: 1, 2, 4, 8... reallocating and copying along the way. You know the final
length before the loop even starts, so allocate **once**:

```go
counts := make([]int, 0, len(infs)) // length 0, capacity n
```

The tests check that the result's capacity is exactly `len(infs)` and that the
function makes a single allocation. Keep `append` with length 0, or use
`make([]int, len(infs))` and assign by index. Just don't mix the two, or you'll
append *after* a row of zeros.

## withGuest: stop the aliasing bug

`withGuest(base, guest)` should return a new line-up with one extra
influencer. But when `base` has spare capacity, `append(base, guest)` writes
into `base`'s backing array, so two calls overwrite each other:

```
a := withGuest(base, "cy")
b := withGuest(base, "dee")
// a is now [ava bo dee]!
```

Fix it so the result **never shares memory** with `base`: changing the result
must not change `base`, and a second call must not change the first result.
`slices.Clone` is your friend. It must still work when `base` is `nil`.
