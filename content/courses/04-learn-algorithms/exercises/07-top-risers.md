---
title: Top Risers
difficulty: medium
after: sorting-algorithms
hints:
  - 'Sorting a slice sorts it **in place**, and `infs[:]` shares memory with the caller''s slice. Work on `slices.Clone(infs)` so the caller''s order survives.'
  - 'Sort the copy with `slices.SortFunc`: bigger `Growth` first, and `cmp.Compare` on `Handle` to break ties. Then return the first `k` elements.'
  - '`k` can be negative or bigger than `len(infs)`. Clamp it with the `min` and `max` built-ins before slicing.'
exercise:
  starter: |
    package main

    import "fmt"

    type Influencer struct {
    	Handle string
    	Growth int // followers gained this week (negative if they lost followers)
    }

    func topRisers(infs []Influencer, k int) []Influencer {
    	return nil
    }

    func main() {
    	infs := []Influencer{{"ava", 120}, {"bo", -40}, {"cy", 900}, {"dee", 120}, {"eve", 15}}
    	fmt.Println(topRisers(infs, 3)) // want [{cy 900} {ava 120} {dee 120}]
    	fmt.Println(infs)               // unchanged
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"slices"
    )

    type Influencer struct {
    	Handle string
    	Growth int // followers gained this week (negative if they lost followers)
    }

    func topRisers(infs []Influencer, k int) []Influencer {
    	k = max(0, min(k, len(infs)))
    	sorted := slices.Clone(infs)
    	slices.SortFunc(sorted, func(a, b Influencer) int {
    		if c := cmp.Compare(b.Growth, a.Growth); c != 0 {
    			return c
    		}
    		return cmp.Compare(a.Handle, b.Handle)
    	})
    	return sorted[:k:k]
    }

    func main() {
    	infs := []Influencer{{"ava", 120}, {"bo", -40}, {"cy", 900}, {"dee", 120}, {"eve", 15}}
    	fmt.Println(topRisers(infs, 3))
    	fmt.Println(infs)
    }
  tests: |
    package main

    import (
    	"fmt"
    	"slices"
    	"testing"
    	"time"
    )

    func TestTopRisers(t *testing.T) {
    	base := []Influencer{{"ava", 120}, {"bo", -40}, {"cy", 900}, {"dee", 120}, {"eve", 15}}
    	tests := []struct {
    		name string
    		in   []Influencer
    		k    int
    		want []Influencer
    	}{
    		{"top 3 with a tie", base, 3, []Influencer{{"cy", 900}, {"ava", 120}, {"dee", 120}}},
    		{"top 1", base, 1, []Influencer{{"cy", 900}}},
    		{"k is len", base, 5, []Influencer{{"cy", 900}, {"ava", 120}, {"dee", 120}, {"eve", 15}, {"bo", -40}}},
    		{"k too big", base, 99, []Influencer{{"cy", 900}, {"ava", 120}, {"dee", 120}, {"eve", 15}, {"bo", -40}}},
    		{"k zero", base, 0, []Influencer{}},
    		{"k negative", base, -2, []Influencer{}},
    		{"empty", nil, 3, []Influencer{}},
    		{"single", []Influencer{{"solo", -5}}, 1, []Influencer{{"solo", -5}}},
    		{"all negative", []Influencer{{"a", -10}, {"b", -1}, {"c", -300}}, 2, []Influencer{{"b", -1}, {"a", -10}}},
    		{"all tied", []Influencer{{"zed", 0}, {"amy", 0}, {"max", 0}}, 2, []Influencer{{"amy", 0}, {"max", 0}}},
    	}
    	for _, tt := range tests {
    		in := slices.Clone(tt.in)
    		got := topRisers(in, tt.k)
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("%s: topRisers(%v, %d) = %v, want %v", tt.name, tt.in, tt.k, got, tt.want)
    		}
    		if !slices.Equal(in, tt.in) {
    			t.Errorf("%s: topRisers changed its input to %v; it must leave the caller's slice alone", tt.name, in)
    		}
    	}
    }

    func TestTopRisersResultIsIndependent(t *testing.T) {
    	in := []Influencer{{"ava", 3}, {"bo", 2}, {"cy", 1}}
    	got := topRisers(in, 2)
    	if len(got) != 2 {
    		t.Fatalf("topRisers(%v, 2) returned %d influencers, want 2", in, len(got))
    	}
    	got = append(got, Influencer{"zz", 0})
    	got[0].Growth = 999
    	if in[0].Growth != 3 || in[2] != (Influencer{"cy", 1}) {
    		t.Errorf("changing topRisers' result changed the input to %v: return a copy", in)
    	}
    }

    func TestTopRisersLarge(t *testing.T) {
    	n := 100_000
    	in := make([]Influencer, n)
    	for i := range in {
    		in[i] = Influencer{fmt.Sprintf("user%06d", i), (i * 7919) % 100_003}
    	}
    	start := time.Now()
    	got := topRisers(in, 10)
    	if d := time.Since(start); d > time.Second {
    		t.Errorf("topRisers(%d influencers, 10) took %v: use an O(n log n) sort", n, d)
    	}
    	if len(got) != 10 || got[0].Growth < got[9].Growth {
    		t.Errorf("topRisers(%d influencers, 10) = %v, want the 10 biggest, biggest first", n, got)
    	}
    }
---

Every Monday, Clout emails brands the week's **top risers**: the accounts that
gained the most followers.

Write `topRisers(infs, k)`. It returns the `k` influencers with the highest
`Growth`, highest first. Influencers with equal `Growth` are ordered by `Handle`
alphabetically.

It must **not** change the caller's slice (the leaderboard page still needs it in
its original order), and the result must not share memory with it either.

- If `k` is 0 or negative, return an empty slice.
- If `k` is bigger than `len(infs)`, return all of them, sorted.
- `Growth` can be negative: some accounts lose followers.

## Example

```go
infs := []Influencer{{"ava", 120}, {"bo", -40}, {"cy", 900}, {"dee", 120}, {"eve", 15}}
topRisers(infs, 3) // [{cy 900} {ava 120} {dee 120}]
topRisers(infs, 0) // []
```

## Constraints

- Up to 100,000 influencers; handles are unique.
- O(n log n) is plenty. Bubble sort's O(n²) is not.
