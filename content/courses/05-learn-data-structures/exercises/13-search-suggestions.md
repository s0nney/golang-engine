---
title: Search Suggestions
difficulty: hard
after: tries
hints:
  - 'Walking the whole subtree under the prefix (as `Suggest` did in the tries chapter) and sorting every match is correct, but a one-letter prefix matches thousands of names, 200,000 times. The answer has to be **ready** at the prefix''s node: store, in every trie node, the best (up to) 3 names in its subtree.'
  - 'Counts only ever go **up**. So when `name` is searched, the only lists that can change are those on the path from the root to `name`''s last node, and in each of them only `name`''s position can change. Walk that path and, at every node (the root too), put `name` into the list if it isn''t there, move it up past anything it now beats, and cut the list back to 3.'
  - 'Keep one record per name, e.g. `type entry struct{ name string; count int }`, in a `map[string]*entry`, and store `*entry` pointers in the nodes'' lists. Bumping `e.count` then updates every list that holds it at once, and `slices.Index(list, e)` finds it. Build a fresh `[]string` in `Suggest` so callers can''t modify your lists.'
exercise:
  starter: |
    package main

    import "fmt"

    type SearchBox struct {
    	// your fields here
    }

    func NewSearchBox() *SearchBox {
    	return &SearchBox{}
    }

    // Searched records one more search for name.
    func (b *SearchBox) Searched(name string) {
    }

    // Suggest returns up to 3 searched names that start with prefix: most
    // searched first, equal counts in alphabetical order.
    func (b *SearchBox) Suggest(prefix string) []string {
    	return nil
    }

    func main() {
    	box := NewSearchBox()
    	for _, name := range []string{"mira", "milo", "kai", "mira", "max", "milo", "mirabel", "mira"} {
    		box.Searched(name)
    	}
    	fmt.Println(box.Suggest("mi")) // want [mira milo mirabel]
    	fmt.Println(box.Suggest("m"))  // want [mira milo max]
    	fmt.Println(box.Suggest("x"))  // want []
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    )

    type entry struct {
    	name  string
    	count int
    }

    // better reports whether a should be suggested before b.
    func better(a, b *entry) bool {
    	if a.count != b.count {
    		return a.count > b.count
    	}
    	return a.name < b.name
    }

    type sbNode struct {
    	children map[rune]*sbNode
    	top      []*entry // the best (up to 3) names in this subtree, best first
    }

    // bump moves e into place in n.top after e's count went up.
    func (n *sbNode) bump(e *entry) {
    	i := slices.Index(n.top, e)
    	if i < 0 {
    		n.top = append(n.top, e)
    		i = len(n.top) - 1
    	}
    	for i > 0 && better(n.top[i], n.top[i-1]) {
    		n.top[i], n.top[i-1] = n.top[i-1], n.top[i]
    		i--
    	}
    	if len(n.top) > 3 {
    		n.top = n.top[:3]
    	}
    }

    // SearchBox suggests the most searched names for a prefix.
    type SearchBox struct {
    	root  sbNode
    	names map[string]*entry
    }

    func NewSearchBox() *SearchBox {
    	return &SearchBox{names: map[string]*entry{}}
    }

    func (b *SearchBox) Searched(name string) {
    	e, ok := b.names[name]
    	if !ok {
    		e = &entry{name: name}
    		b.names[name] = e
    	}
    	e.count++
    	n := &b.root
    	n.bump(e)
    	for _, r := range name {
    		if n.children == nil {
    			n.children = map[rune]*sbNode{}
    		}
    		next, ok := n.children[r]
    		if !ok {
    			next = &sbNode{}
    			n.children[r] = next
    		}
    		n = next
    		n.bump(e)
    	}
    }

    func (b *SearchBox) Suggest(prefix string) []string {
    	n := &b.root
    	for _, r := range prefix {
    		n = n.children[r]
    		if n == nil {
    			return nil
    		}
    	}
    	out := make([]string, len(n.top))
    	for i, e := range n.top {
    		out[i] = e.name
    	}
    	return out
    }

    func main() {
    	box := NewSearchBox()
    	for _, name := range []string{"mira", "milo", "kai", "mira", "max", "milo", "mirabel", "mira"} {
    		box.Searched(name)
    	}
    	fmt.Println(box.Suggest("mi"))
    	fmt.Println(box.Suggest("m"))
    	fmt.Println(box.Suggest("x"))
    }
  tests: |
    package main

    import (
    	"cmp"
    	"slices"
    	"strings"
    	"testing"
    	"time"
    )

    func search(b *SearchBox, names ...string) {
    	for _, n := range names {
    		b.Searched(n)
    	}
    }

    func expect(t *testing.T, step string, b *SearchBox, prefix string, want ...string) {
    	t.Helper()
    	if got := b.Suggest(prefix); !slices.Equal(got, want) {
    		t.Errorf("after %s: Suggest(%q) = %q, want %q", step, prefix, got, want)
    	}
    }

    func TestSearchBox(t *testing.T) {
    	b := NewSearchBox()
    	expect(t, "no searches", b, "")
    	expect(t, "no searches", b, "mi")
    	search(b, "mira", "milo", "kai", "mira", "max", "milo", "mirabel", "mira")
    	step := "mira x3, milo x2, kai, max, mirabel"
    	expect(t, step, b, "mi", "mira", "milo", "mirabel")
    	expect(t, step, b, "m", "mira", "milo", "max")
    	expect(t, step, b, "", "mira", "milo", "kai")
    	expect(t, step, b, "mira", "mira", "mirabel")
    	expect(t, step, b, "mirab", "mirabel")
    	expect(t, step, b, "k", "kai")
    	expect(t, step, b, "x")
    	expect(t, step, b, "miraa")
    	expect(t, step, b, "mirabella") // longer than any name
    	search(b, "max", "max")
    	step += ", then max x2"
    	expect(t, step, b, "m", "max", "mira", "milo") // max (3) ties mira (3): alphabetical
    	search(b, "mirabel", "mirabel", "mirabel", "mirabel")
    	step += ", then mirabel x4"
    	expect(t, step, b, "mi", "mirabel", "mira", "milo")
    	expect(t, step, b, "m", "mirabel", "max", "mira")
    	expect(t, step, b, "", "mirabel", "max", "mira")
    }

    func TestSearchBoxOutsiderClimbs(t *testing.T) {
    	b := NewSearchBox()
    	search(b, "ana", "ana", "ana", "abe", "abe", "ada", "ada", "amy")
    	expect(t, "ana x3, abe x2, ada x2, amy", b, "a", "ana", "abe", "ada")
    	search(b, "amy", "amy", "amy")
    	expect(t, "then amy x3", b, "a", "amy", "ana", "abe")
    	search(b, "zoë", "zoë", "zoë", "zoë", "zoë")
    	expect(t, "then zoë x5", b, "zo", "zoë")
    	expect(t, "then zoë x5", b, "zoë", "zoë")
    	expect(t, "then zoë x5", b, "", "zoë", "amy", "ana")
    }

    func TestSuggestReturnsACopy(t *testing.T) {
    	b := NewSearchBox()
    	search(b, "kai", "kai", "kira")
    	got := b.Suggest("k")
    	if len(got) != 2 {
    		t.Fatalf("Suggest(%q) = %q, want [kai kira]", "k", got)
    	}
    	got[0] = "hacked"
    	expect(t, "the caller changed Suggest's result", b, "k", "kai", "kira")
    }

    func TestTwoSearchBoxes(t *testing.T) {
    	a, b := NewSearchBox(), NewSearchBox()
    	a.Searched("mira")
    	if got := b.Suggest(""); len(got) != 0 {
    		t.Errorf("two search boxes share state: a fresh box suggests %q", got)
    	}
    }

    // TestSearchBoxAgainstBruteForce compares every answer with a simple scan
    // over all names.
    func TestSearchBoxAgainstBruteForce(t *testing.T) {
    	b := NewSearchBox()
    	counts := map[string]int{}
    	var x uint64 = 5
    	next := func(n int) int {
    		x = x*6364136223846793005 + 1442695040888963407
    		return int(x>>33) % n
    	}
    	randName := func() string {
    		buf := make([]byte, 1+next(4))
    		for i := range buf {
    			buf[i] = "abc"[next(3)]
    		}
    		return string(buf)
    	}
    	for i := range 4000 {
    		name := randName()
    		b.Searched(name)
    		counts[name]++
    		prefix := randName()
    		prefix = prefix[:min(len(prefix), next(3))]
    		var want []string
    		for n := range counts {
    			if strings.HasPrefix(n, prefix) {
    				want = append(want, n)
    			}
    		}
    		slices.SortFunc(want, func(p, q string) int {
    			return cmp.Or(cmp.Compare(counts[q], counts[p]), cmp.Compare(p, q))
    		})
    		want = want[:min(3, len(want))]
    		if got := b.Suggest(prefix); !slices.Equal(got, want) {
    			t.Fatalf("after %d searches (last: %q): Suggest(%q) = %q, want %q", i+1, name, prefix, got, want)
    		}
    	}
    }

    func TestSearchBoxLarge(t *testing.T) {
    	var x uint64 = 2024
    	next := func(n int) int {
    		x = x*6364136223846793005 + 1442695040888963407
    		return int(x>>33) % n
    	}
    	const letters = "aeiklmnorst"
    	names := make([]string, 100_000)
    	for i := range names {
    		buf := make([]byte, 3+next(6))
    		for j := range buf {
    			buf[j] = letters[next(len(letters))]
    		}
    		names[i] = string(buf)
    	}
    	done := make(chan uint64, 1)
    	go func() {
    		b := NewSearchBox()
    		var h uint64 = 14695981039346656037 // FNV-1a hash of every suggestion
    		for i := range 300_000 {
    			j := next(len(names))
    			j = j * next(len(names)) / len(names) // low indexes are popular
    			b.Searched(names[j])
    			if i%3 != 0 {
    				q := names[next(len(names))]
    				q = q[:1+next(3)]
    				if i%101 == 0 {
    					q = "z" + q // no matches
    				}
    				for _, s := range b.Suggest(q) {
    					for k := range len(s) {
    						h = (h ^ uint64(s[k])) * 1099511628211
    					}
    					h = (h ^ ',') * 1099511628211
    				}
    				h = (h ^ ';') * 1099511628211
    			}
    		}
    		done <- h
    	}()
    	select {
    	case got := <-done:
    		if want := uint64(3299478197483395515); got != want {
    			t.Errorf("300,000 searches and 200,000 suggestions: hash of all suggestions = %d, want %d (some Suggest answer was wrong)", got, want)
    		}
    	case <-time.After(time.Second):
    		t.Fatalf("300,000 searches and 200,000 short-prefix suggestions took over a second: keep each node's top 3 ready instead of walking its subtree")
    	}
    }
---

The player-search box suggests names as you type. Showing **every** player whose
name starts with `mi` isn't helpful, so it shows the **3 most searched** ones.

Implement `SearchBox`:

- `NewSearchBox()` returns an empty search box.
- `Searched(name)` records one more search for `name`. The first search adds the
  name to the box.
- `Suggest(prefix)` returns up to 3 names that start with `prefix`, **most
  searched first**; names with equal counts are in **alphabetical** order. With no
  match it returns an empty or nil slice. `Suggest("")` considers every name. The
  caller may modify the returned slice, so it must not share memory with the box.

## Example

```go
box := NewSearchBox()
for _, name := range []string{"mira", "milo", "kai", "mira", "max", "milo", "mirabel", "mira"} {
	box.Searched(name)
}
// counts: mira 3, milo 2, kai 1, max 1, mirabel 1
box.Suggest("mi")   // [mira milo mirabel]
box.Suggest("m")    // [mira milo max]   (max and mirabel tie at 1: max is first alphabetically)
box.Suggest("mira") // [mira mirabel]
box.Suggest("x")    // []
```

## Constraints

- Up to 100,000 distinct names, 300,000 `Searched` calls and 200,000 `Suggest`
  calls, many with one- or two-letter prefixes that match thousands of names.
- `Searched` should be O(L) and `Suggest` O(len(prefix)), where L is the name's
  length (the list size, 3, is a constant). Collecting and sorting every match on
  each keystroke takes minutes on the performance test; the limit is one second.
- Names may contain any Unicode letters.
