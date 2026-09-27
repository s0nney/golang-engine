---
title: Trending Hashtag
difficulty: easy
after: big-o-analysis
hints:
  - Comparing every tag with every other tag is O(n²). A map from tag to count lets you count everything in **one** pass.
  - 'For the tie rule, make a second pass over `tags` (not over the map, whose order is random) and keep a tag only if its count is **strictly** bigger than the best so far.'
exercise:
  starter: |
    package main

    import "fmt"

    // trending returns the hashtag that appears most often in tags.
    // If several tags tie, return the one that appears first in tags.
    // If tags is empty, return "" and false.
    func trending(tags []string) (string, bool) {
    	// 1. Count how often each tag appears (a map is perfect for this).
    	// 2. Walk tags in order and remember the tag with the highest count.
    	return "", false
    }

    func main() {
    	tags := []string{"#ootd", "#gym", "#ootd", "#food", "#gym", "#ootd"}
    	fmt.Println(trending(tags)) // want: #ootd true
    }
  solution: |
    package main

    import "fmt"

    func trending(tags []string) (string, bool) {
    	counts := map[string]int{}
    	for _, t := range tags {
    		counts[t]++
    	}
    	best, bestCount := "", 0
    	for _, t := range tags {
    		if counts[t] > bestCount {
    			best, bestCount = t, counts[t]
    		}
    	}
    	return best, bestCount > 0
    }

    func main() {
    	tags := []string{"#ootd", "#gym", "#ootd", "#food", "#gym", "#ootd"}
    	fmt.Println(trending(tags))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    	"time"
    )

    func TestTrending(t *testing.T) {
    	tests := []struct {
    		tags   []string
    		want   string
    		wantOK bool
    	}{
    		{nil, "", false},
    		{[]string{"#solo"}, "#solo", true},
    		{[]string{"#ootd", "#gym", "#ootd", "#food", "#gym", "#ootd"}, "#ootd", true},
    		{[]string{"#a", "#b", "#b", "#a"}, "#a", true},
    		{[]string{"#x", "#y", "#z"}, "#x", true},
    		{[]string{"#cat", "#dog", "#dog", "#cat", "#dog"}, "#dog", true},
    		{[]string{"#late", "#early", "#early", "#late", "#late"}, "#late", true},
    	}
    	for _, tt := range tests {
    		got, ok := trending(tt.tags)
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("trending(%q) = %q, %v, want %q, %v", tt.tags, got, ok, tt.want, tt.wantOK)
    		}
    	}
    }

    func TestTrendingLarge(t *testing.T) {
    	tags := make([]string, 100_000)
    	for i := range tags {
    		tags[i] = fmt.Sprintf("#tag%d", i%50_000)
    	}
    	tags[len(tags)-1] = "#tag7"
    	start := time.Now()
    	got, ok := trending(tags)
    	if got != "#tag7" || !ok {
    		t.Errorf("trending(100,000 tags) = %q, %v, want %q, true", got, ok, "#tag7")
    	}
    	if d := time.Since(start); d > time.Second {
    		t.Errorf("trending(100,000 tags) took %v: count with a map instead of comparing every pair", d)
    	}
    }
---

Clout's Explore tab shows one **trending hashtag**: the tag used most often in
the last hour of posts.

Complete `trending(tags)`. It returns the tag that appears most often in `tags`,
and `true`. If several tags share the top count, return the one that appears
**first** in `tags`. If `tags` is empty, return `""` and `false`.

## Examples

```
trending([]string{"#ootd", "#gym", "#ootd", "#food", "#gym", "#ootd"})  // "#ootd", true
trending([]string{"#a", "#b", "#b", "#a"})                              // "#a", true (tie, "#a" came first)
trending(nil)                                                           // "", false
```

## Constraints

- `tags` holds up to 100,000 tags.
- Aim for **O(n)** time. One of the tests uses 100,000 tags, where comparing
  every tag with every other one is far too slow.
