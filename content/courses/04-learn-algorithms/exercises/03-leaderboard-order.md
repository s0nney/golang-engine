---
title: Leaderboard Order
difficulty: easy
after: sorting-algorithms
hints:
  - '`slices.SortFunc(s, func(a, b Influencer) int { ... })` sorts in place. Return a negative number when `a` should come first.'
  - '`cmp.Compare(b.Followers, a.Followers)` (note: `b` first) sorts biggest first. If that gives 0, fall back to `cmp.Compare(a.Handle, b.Handle)`.'
exercise:
  starter: |
    package main

    import "fmt"

    type Influencer struct {
    	Handle    string
    	Followers int
    }

    // sortLeaderboard sorts infs in place: most followers first, and
    // accounts with the same follower count in alphabetical order of Handle.
    func sortLeaderboard(infs []Influencer) {
    	// Use slices.SortFunc with a comparison built from cmp.Compare.
    	// (Add "cmp" and "slices" to the imports.)
    }

    func main() {
    	infs := []Influencer{{"cy", 300}, {"ava", 900}, {"bo", 300}, {"dee", 1200}}
    	sortLeaderboard(infs)
    	fmt.Println(infs) // want: [{dee 1200} {ava 900} {bo 300} {cy 300}]
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"slices"
    )

    type Influencer struct {
    	Handle    string
    	Followers int
    }

    func sortLeaderboard(infs []Influencer) {
    	slices.SortFunc(infs, func(a, b Influencer) int {
    		if c := cmp.Compare(b.Followers, a.Followers); c != 0 {
    			return c
    		}
    		return cmp.Compare(a.Handle, b.Handle)
    	})
    }

    func main() {
    	infs := []Influencer{{"cy", 300}, {"ava", 900}, {"bo", 300}, {"dee", 1200}}
    	sortLeaderboard(infs)
    	fmt.Println(infs)
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestSortLeaderboard(t *testing.T) {
    	tests := []struct {
    		in, want []Influencer
    	}{
    		{nil, nil},
    		{[]Influencer{{"solo", 5}}, []Influencer{{"solo", 5}}},
    		{
    			[]Influencer{{"cy", 300}, {"ava", 900}, {"bo", 300}, {"dee", 1200}},
    			[]Influencer{{"dee", 1200}, {"ava", 900}, {"bo", 300}, {"cy", 300}},
    		},
    		{
    			[]Influencer{{"zed", 7}, {"amy", 7}, {"max", 7}},
    			[]Influencer{{"amy", 7}, {"max", 7}, {"zed", 7}},
    		},
    		{
    			[]Influencer{{"a", 1}, {"b", 2}, {"c", 3}, {"d", 4}},
    			[]Influencer{{"d", 4}, {"c", 3}, {"b", 2}, {"a", 1}},
    		},
    		{
    			[]Influencer{{"new", 0}, {"top", 1_000_000}, {"mid", 50}, {"alt", 50}, {"old", 0}},
    			[]Influencer{{"top", 1_000_000}, {"alt", 50}, {"mid", 50}, {"new", 0}, {"old", 0}},
    		},
    	}
    	for _, tt := range tests {
    		got := slices.Clone(tt.in)
    		sortLeaderboard(got)
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("sortLeaderboard(%v) gave %v, want %v", tt.in, got, tt.want)
    		}
    	}
    }
---

The Clout leaderboard lists accounts with the **most followers first**. When two
accounts have the same number of followers, the one whose `Handle` comes first
alphabetically is listed first, so the order never flickers between refreshes.

Complete `sortLeaderboard(infs)`, which sorts the slice **in place**.

## Example

```go
infs := []Influencer{{"cy", 300}, {"ava", 900}, {"bo", 300}, {"dee", 1200}}
sortLeaderboard(infs)
// [{dee 1200} {ava 900} {bo 300} {cy 300}]
```

`bo` and `cy` both have 300 followers, so they're ordered by handle.

## Constraints

- Up to 100,000 accounts, so use the standard library's O(n log n) sort rather
  than a hand-written bubble sort.
- Handles are unique.
