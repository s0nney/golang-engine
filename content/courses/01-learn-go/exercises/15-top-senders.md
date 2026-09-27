---
title: Top Senders
difficulty: hard
after: generics-basics
hints:
  - 'For `TopN`: copy the input with `slices.Clone` so you never reorder the caller''s slice, sort the copy by key **descending**, then keep the first `n`. Handle `n <= 0` and `n > len(items)` with `min` and an early return.'
  - '`slices.SortStableFunc` works like `slices.SortFunc` but keeps equal elements in their original order, which is exactly the tie rule. For descending order, compare `b` with `a`: `cmp.Compare(key(b), key(a))`.'
  - 'For `TopPerGroup`: first build a `map[G][]T` by appending each item to its group''s slice (input order is preserved), then replace every group''s slice with `TopN` of it.'
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    func TopN[T any, K cmp.Ordered](items []T, n int, key func(T) K) []T {
    	return nil
    }

    func TopPerGroup[T any, G comparable, K cmp.Ordered](items []T, n int, group func(T) G, key func(T) K) map[G][]T {
    	return nil
    }

    type Sender struct {
    	Name    string
    	Country string
    	Sent    int
    }

    func main() {
    	senders := []Sender{
    		{"Mia", "UK", 120}, {"Sam", "US", 340}, {"Ana", "UK", 340},
    		{"Bo", "US", 90}, {"Kim", "UK", 15},
    	}
    	bySent := func(s Sender) int { return s.Sent }
    	byCountry := func(s Sender) string { return s.Country }

    	fmt.Println(TopN(senders, 2, bySent))
    	fmt.Println(TopPerGroup(senders, 1, byCountry, bySent))
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"slices"
    )

    func TopN[T any, K cmp.Ordered](items []T, n int, key func(T) K) []T {
    	if n <= 0 {
    		return nil
    	}
    	sorted := slices.Clone(items)
    	slices.SortStableFunc(sorted, func(a, b T) int {
    		return cmp.Compare(key(b), key(a))
    	})
    	return sorted[:min(n, len(sorted))]
    }

    func TopPerGroup[T any, G comparable, K cmp.Ordered](items []T, n int, group func(T) G, key func(T) K) map[G][]T {
    	groups := map[G][]T{}
    	for _, item := range items {
    		g := group(item)
    		groups[g] = append(groups[g], item)
    	}
    	for g, members := range groups {
    		groups[g] = TopN(members, n, key)
    	}
    	return groups
    }

    type Sender struct {
    	Name    string
    	Country string
    	Sent    int
    }

    func main() {
    	senders := []Sender{
    		{"Mia", "UK", 120}, {"Sam", "US", 340}, {"Ana", "UK", 340},
    		{"Bo", "US", 90}, {"Kim", "UK", 15},
    	}
    	bySent := func(s Sender) int { return s.Sent }
    	byCountry := func(s Sender) string { return s.Country }

    	fmt.Println(TopN(senders, 2, bySent))
    	fmt.Println(TopPerGroup(senders, 1, byCountry, bySent))
    }
  tests: |
    package main

    import (
    	"maps"
    	"slices"
    	"strings"
    	"testing"
    	"time"
    )

    var senders = []Sender{
    	{"Mia", "UK", 120}, {"Sam", "US", 340}, {"Ana", "UK", 340},
    	{"Bo", "US", 90}, {"Kim", "UK", 15},
    }

    func bySent(s Sender) int       { return s.Sent }
    func byCountry(s Sender) string { return s.Country }

    func TestTopNSenders(t *testing.T) {
    	tests := []struct {
    		n    int
    		want []Sender
    	}{
    		{2, []Sender{{"Sam", "US", 340}, {"Ana", "UK", 340}}},
    		{3, []Sender{{"Sam", "US", 340}, {"Ana", "UK", 340}, {"Mia", "UK", 120}}},
    		{5, []Sender{{"Sam", "US", 340}, {"Ana", "UK", 340}, {"Mia", "UK", 120}, {"Bo", "US", 90}, {"Kim", "UK", 15}}},
    		{99, []Sender{{"Sam", "US", 340}, {"Ana", "UK", 340}, {"Mia", "UK", 120}, {"Bo", "US", 90}, {"Kim", "UK", 15}}},
    		{0, nil},
    		{-1, nil},
    	}
    	for _, tt := range tests {
    		before := slices.Clone(senders)
    		got := TopN(senders, tt.n, bySent)
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("TopN(senders, %d, bySent) = %v, want %v", tt.n, got, tt.want)
    		}
    		if !slices.Equal(senders, before) {
    			t.Fatalf("TopN(senders, %d, bySent) reordered its input to %v: sort a copy instead", tt.n, senders)
    		}
    	}
    }

    func TestTopNOtherTypes(t *testing.T) {
    	words := []string{"hi", "hello", "hey", "yo", "howdy"}
    	byLen := func(s string) int { return len(s) }
    	if got, want := TopN(words, 3, byLen), []string{"hello", "howdy", "hey"}; !slices.Equal(got, want) {
    		t.Errorf("TopN(%q, 3, byLen) = %q, want %q", words, got, want)
    	}
    	alpha := func(s string) string { return s }
    	if got, want := TopN(words, 2, alpha), []string{"yo", "howdy"}; !slices.Equal(got, want) {
    		t.Errorf("TopN(%q, 2, by the word itself) = %q, want %q", words, got, want)
    	}
    	prices := []float64{0.5, -1.25, 3.75, 0}
    	self := func(f float64) float64 { return f }
    	if got, want := TopN(prices, 3, self), []float64{3.75, 0.5, 0}; !slices.Equal(got, want) {
    		t.Errorf("TopN(%v, 3, self) = %v, want %v", prices, got, want)
    	}
    	if got := TopN([]int(nil), 3, func(i int) int { return i }); len(got) != 0 {
    		t.Errorf("TopN(nil, 3, ...) = %v, want an empty result", got)
    	}
    	emoji := []string{"👋", "🎉🎉🎉", "🐝🐝"}
    	runes := func(s string) int { return len([]rune(s)) }
    	if got, want := TopN(emoji, 1, runes), []string{"🎉🎉🎉"}; !slices.Equal(got, want) {
    		t.Errorf("TopN(%q, 1, rune count) = %q, want %q", emoji, got, want)
    	}
    }

    func TestTopNResultIsIndependent(t *testing.T) {
    	items := []int{5, 1, 9, 3}
    	self := func(i int) int { return i }
    	got := TopN(items, 2, self)
    	if len(got) == 2 {
    		got[0] = -100
    		_ = append(got, -200)
    	}
    	if !slices.Equal(items, []int{5, 1, 9, 3}) {
    		t.Errorf("changing TopN's result changed the input to %v: return a copy", items)
    	}
    }

    func TestTopPerGroup(t *testing.T) {
    	got := TopPerGroup(senders, 1, byCountry, bySent)
    	want := map[string][]Sender{
    		"UK": {{"Ana", "UK", 340}},
    		"US": {{"Sam", "US", 340}},
    	}
    	if !maps.EqualFunc(got, want, slices.Equal) {
    		t.Errorf("TopPerGroup(senders, 1, byCountry, bySent) = %v, want %v", got, want)
    	}

    	got = TopPerGroup(senders, 2, byCountry, bySent)
    	want = map[string][]Sender{
    		"UK": {{"Ana", "UK", 340}, {"Mia", "UK", 120}},
    		"US": {{"Sam", "US", 340}, {"Bo", "US", 90}},
    	}
    	if !maps.EqualFunc(got, want, slices.Equal) {
    		t.Errorf("TopPerGroup(senders, 2, byCountry, bySent) = %v, want %v", got, want)
    	}

    	words := []string{"Hey", "hi", "Hello", "yo", "Yes", "howdy"}
    	firstLetter := func(s string) string { return strings.ToLower(s[:1]) }
    	byLen := func(s string) int { return len(s) }
    	gotWords := TopPerGroup(words, 2, firstLetter, byLen)
    	wantWords := map[string][]string{"h": {"Hello", "howdy"}, "y": {"Yes", "yo"}}
    	if !maps.EqualFunc(gotWords, wantWords, slices.Equal) {
    		t.Errorf("TopPerGroup(%q, 2, first letter, length) = %q, want %q", words, gotWords, wantWords)
    	}

    	if got := TopPerGroup([]Sender(nil), 3, byCountry, bySent); len(got) != 0 {
    		t.Errorf("TopPerGroup(nil, ...) = %v, want an empty map", got)
    	}
    }

    func TestTopNLarge(t *testing.T) {
    	const size = 200_000
    	items := make([]int, size)
    	for i := range items {
    		items[i] = (i * 7919) % size
    	}
    	self := func(i int) int { return i }
    	done := make(chan []int, 1)
    	go func() { done <- TopN(items, size, self) }()
    	select {
    	case got := <-done:
    		if len(got) != size || got[0] != size-1 || got[size-1] != 0 {
    			t.Errorf("TopN(200,000 ints, all of them) should list them from %d down to 0", size-1)
    		}
    	case <-time.After(time.Second):
    		t.Fatalf("TopN(200,000 ints) took over a second: use slices.SortStableFunc rather than a hand-written O(n²) sort")
    	}
    }
---

Textio's admin dashboard has leaderboards everywhere: top senders, the most
expensive campaigns, the busiest countries. Instead of writing a new function
for each, you'll write two generic ones.

**`TopN(items, n, key)`** returns the `n` items with the **largest** `key`,
largest first.

- Items with equal keys keep their original order.
- If `n` is 0 or negative, return an empty (or `nil`) slice. If `n` is bigger
  than `len(items)`, return all of them, sorted.
- Don't change `items`, and return a new slice: changing the result must
  never change the input.

**`TopPerGroup(items, n, group, key)`** splits `items` into groups by
`group(item)` and returns a map from each group to its own `TopN`.

## Example

```go
senders := []Sender{
	{"Mia", "UK", 120}, {"Sam", "US", 340}, {"Ana", "UK", 340},
	{"Bo", "US", 90}, {"Kim", "UK", 15},
}
TopN(senders, 2, bySent)
// [{Sam US 340} {Ana UK 340}]   (a tie: Sam came first)

TopPerGroup(senders, 1, byCountry, bySent)
// map[UK:[{Ana UK 340}] US:[{Sam US 340}]]
```

## Constraints

- `key` can return any ordered type (`int`, `float64`, `string`, ...).
- Up to 200,000 items. One test ranks that many under a one-second limit, so
  use the `slices` package to sort rather than a hand-written nested-loop sort.
