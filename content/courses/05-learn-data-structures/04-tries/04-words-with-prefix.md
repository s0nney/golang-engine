---
title: Words With a Prefix
quiz:
  - question: The trie holds `dr`, `dragon`, `drake` and `dread`. In what order does `WithPrefix("dr")` yield them?
    options:
      - text: '`dragon`, `drake`, `dread`, `dr`'
      - text: '`dr`, `dragon`, `drake`, `dread`'
        correct: true
      - text: '`dr`, `dread`, `drake`, `dragon`'
      - text: In a random order
    explanation: |
      `walk` yields a node's own word *before* visiting its children (a pre-order
      DFS), so `dr` comes first. Children are visited in sorted rune order, so the
      `a` branch (`dragon`, `drake`) comes before the `e` branch (`dread`).
  - question: |
      `walk` calls `append(prefix, r)` for each child, and all the children share the
      same `prefix` slice. Why doesn't that corrupt the results?
    options:
      - text: '`append` always copies the slice'
      - text: 'Each call converts to `string(prefix)` before yielding, which copies the runes, and siblings run one after another, so a sibling only overwrites that slot once the previous sibling is completely finished'
        correct: true
      - text: It does corrupt them; that's a bug
      - text: Runes are immutable, so slices of them can't be overwritten
    explanation: |
      `append` may reuse the backing array, so siblings do write into the same slot.
      That's safe here only because the DFS finishes one child entirely before the next
      starts, and `string(...)` makes a copy. If you stored the `[]rune` slices
      themselves, you'd see this aliasing bug.
exercise:
  starter: |
    package main

    import "fmt"

    type trieNode struct {
    	children map[rune]*trieNode
    	end      bool
    	count    int // how many stored words pass through (or end at) this node
    }

    type Trie struct{ root trieNode }

    // find returns the node reached by spelling prefix, or nil.
    func (t *Trie) find(prefix string) *trieNode {
    	n := &t.root
    	for _, r := range prefix {
    		n = n.children[r]
    		if n == nil {
    			return nil
    		}
    	}
    	return n
    }

    // Insert adds word. Every node on the word's path, from the root to the
    // word's last node, must count it once. Inserting a word that's already
    // stored changes nothing.
    func (t *Trie) Insert(word string) {
    	n := &t.root
    	// ? skip words that are already stored, and keep count up to date
    	for _, r := range word {
    		if n.children == nil {
    			n.children = make(map[rune]*trieNode)
    		}
    		child, ok := n.children[r]
    		if !ok {
    			child = &trieNode{}
    			n.children[r] = child
    		}
    		n = child
    	}
    	n.end = true
    }

    // CountWithPrefix returns how many stored words start with prefix,
    // in O(len(prefix)) time.
    func (t *Trie) CountWithPrefix(prefix string) int {
    	// ?
    	return 0
    }

    // Suggest returns up to limit stored words starting with prefix, in
    // alphabetical order. It stops searching as soon as it has limit words.
    func (t *Trie) Suggest(prefix string, limit int) []string {
    	var out []string
    	// ? depth-first, children in sorted order, stop at limit
    	return out
    }

    func main() {
    	var names Trie
    	for _, name := range []string{"dread", "mira", "dragon", "dr", "drake", "dragonfly", "drake"} {
    		names.Insert(name)
    	}
    	fmt.Println(names.CountWithPrefix("dr"), names.CountWithPrefix(""), names.CountWithPrefix("x")) // want: 5 6 0
    	fmt.Println(names.Suggest("dr", 3))                                                             // want: [dr dragon dragonfly]
    }
  solution: |
    package main

    import (
    	"fmt"
    	"maps"
    	"slices"
    )

    type trieNode struct {
    	children map[rune]*trieNode
    	end      bool
    	count    int // how many stored words pass through (or end at) this node
    }

    type Trie struct{ root trieNode }

    // find returns the node reached by spelling prefix, or nil.
    func (t *Trie) find(prefix string) *trieNode {
    	n := &t.root
    	for _, r := range prefix {
    		n = n.children[r]
    		if n == nil {
    			return nil
    		}
    	}
    	return n
    }

    // Insert adds word. Every node on the word's path, from the root to the
    // word's last node, must count it once. Inserting a word that's already
    // stored changes nothing.
    func (t *Trie) Insert(word string) {
    	if n := t.find(word); n != nil && n.end {
    		return
    	}
    	n := &t.root
    	n.count++
    	for _, r := range word {
    		if n.children == nil {
    			n.children = make(map[rune]*trieNode)
    		}
    		child, ok := n.children[r]
    		if !ok {
    			child = &trieNode{}
    			n.children[r] = child
    		}
    		child.count++
    		n = child
    	}
    	n.end = true
    }

    // CountWithPrefix returns how many stored words start with prefix,
    // in O(len(prefix)) time.
    func (t *Trie) CountWithPrefix(prefix string) int {
    	if n := t.find(prefix); n != nil {
    		return n.count
    	}
    	return 0
    }

    // Suggest returns up to limit stored words starting with prefix, in
    // alphabetical order. It stops searching as soon as it has limit words.
    func (t *Trie) Suggest(prefix string, limit int) []string {
    	var out []string
    	n := t.find(prefix)
    	if n == nil || limit <= 0 {
    		return out
    	}
    	var walk func(n *trieNode, spelled []rune) bool
    	walk = func(n *trieNode, spelled []rune) bool {
    		if n.end {
    			out = append(out, string(spelled))
    			if len(out) == limit {
    				return false
    			}
    		}
    		for _, r := range slices.Sorted(maps.Keys(n.children)) {
    			if !walk(n.children[r], append(spelled, r)) {
    				return false
    			}
    		}
    		return true
    	}
    	walk(n, []rune(prefix))
    	return out
    }

    func main() {
    	var names Trie
    	for _, name := range []string{"dread", "mira", "dragon", "dr", "drake", "dragonfly", "drake"} {
    		names.Insert(name)
    	}
    	fmt.Println(names.CountWithPrefix("dr"), names.CountWithPrefix(""), names.CountWithPrefix("x"))
    	fmt.Println(names.Suggest("dr", 3))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func newTrie(words ...string) *Trie {
    	t := &Trie{}
    	for _, w := range words {
    		t.Insert(w)
    	}
    	return t
    }

    var players = []string{"dread", "mira", "dragon", "dr", "drake", "dragonfly", "miro", "émile", "drake", "mira"}

    func TestCountWithPrefix(t *testing.T) {
    	tr := newTrie(players...)
    	for _, tt := range []struct {
    		prefix string
    		want   int
    	}{{"", 8}, {"dr", 5}, {"drag", 2}, {"dragonfly", 1}, {"dragonflyz", 0}, {"mir", 2}, {"é", 1}, {"zz", 0}} {
    		if got := tr.CountWithPrefix(tt.prefix); got != tt.want {
    			t.Errorf("CountWithPrefix(%q) = %d, want %d (words: %q; duplicates count once)", tt.prefix, got, tt.want, players)
    		}
    	}
    	var empty Trie
    	if got := empty.CountWithPrefix(""); got != 0 {
    		t.Errorf("empty trie: CountWithPrefix(\"\") = %d, want 0", got)
    	}
    }

    func TestSuggest(t *testing.T) {
    	tr := newTrie(players...)
    	for _, tt := range []struct {
    		prefix string
    		limit  int
    		want   []string
    	}{
    		{"dr", 3, []string{"dr", "dragon", "dragonfly"}},
    		{"dr", 10, []string{"dr", "dragon", "dragonfly", "drake", "dread"}},
    		{"", 4, []string{"dr", "dragon", "dragonfly", "drake"}},
    		{"mi", 1, []string{"mira"}},
    		{"drake", 5, []string{"drake"}},
    		{"q", 5, nil},
    		{"dr", 0, nil},
    	} {
    		got := tr.Suggest(tt.prefix, tt.limit)
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("Suggest(%q, %d) = %q, want %q", tt.prefix, tt.limit, got, tt.want)
    		}
    	}
    }

    func TestSuggestStopsEarly(t *testing.T) {
    	tr := &Trie{}
    	for a := 'a'; a <= 'z'; a++ {
    		for b := 'a'; b <= 'z'; b++ {
    			for c := 'a'; c <= 'z'; c++ {
    				tr.Insert(string([]rune{'x', a, b, c}))
    			}
    		}
    	}
    	// Booby-trap the far end of the trie with a nil child. A walk that
    	// stops once it has 3 words never gets anywhere near it.
    	tr.find("xzzz").children = map[rune]*trieNode{'!': nil}
    	defer func() {
    		if r := recover(); r != nil {
    			t.Fatalf("Suggest(\"x\", 3) walked the whole trie (it hit a trap at \"xzzz\"): stop as soon as you have limit words")
    		}
    	}()
    	got := tr.Suggest("x", 3)
    	if want := []string{"xaaa", "xaab", "xaac"}; !slices.Equal(got, want) {
    		t.Errorf("Suggest(\"x\", 3) = %q, want %q", got, want)
    	}
    }
---

The player types `dr` and the box should list every name that starts with it.
`find` gets us to the `r` node in two steps. Every word below that node starts
with `dr`, so all that's left is to visit the whole subtree and collect the words.
That's a **depth-first search** (DFS), exactly like the tree traversals from chapter 1.

## The DFS

At each node, carry along the characters spelled so far. If the node ends a word,
report it. Then recurse into each child with that child's character appended.

```go
package main

import (
	"fmt"
	"iter"
	"maps"
	"slices"
)

type trieNode struct {
	children map[rune]*trieNode
	end      bool
}

type Trie struct{ root trieNode }

func (t *Trie) Insert(word string) {
	n := &t.root
	for _, r := range word {
		if n.children == nil {
			n.children = make(map[rune]*trieNode)
		}
		child, ok := n.children[r]
		if !ok {
			child = &trieNode{}
			n.children[r] = child
		}
		n = child
	}
	n.end = true
}

func (t *Trie) find(prefix string) *trieNode {
	n := &t.root
	for _, r := range prefix {
		n = n.children[r]
		if n == nil {
			return nil
		}
	}
	return n
}

// walk yields every word in n's subtree and reports whether to keep going.
func (n *trieNode) walk(prefix []rune, yield func(string) bool) bool {
	if n.end && !yield(string(prefix)) {
		return false
	}
	for _, r := range slices.Sorted(maps.Keys(n.children)) {
		if !n.children[r].walk(append(prefix, r), yield) {
			return false
		}
	}
	return true
}

// WithPrefix yields every stored word starting with prefix, in sorted order.
func (t *Trie) WithPrefix(prefix string) iter.Seq[string] {
	return func(yield func(string) bool) {
		if n := t.find(prefix); n != nil {
			n.walk([]rune(prefix), yield)
		}
	}
}

func main() {
	var names Trie
	for _, name := range []string{"dread", "mira", "dragon", "dr", "drake", "dragonfly"} {
		names.Insert(name)
	}
	for name := range names.WithPrefix("drag") {
		fmt.Println(name)
	}
	fmt.Println(slices.Collect(names.WithPrefix("dr")))
	fmt.Println(len(slices.Collect(names.WithPrefix("zz"))))
}
```

Output:

```
dragon
dragonfly
[dr dragon dragonfly drake dread]
0
```

## How it works

- **Pre-order.** `walk` handles the node's own word before its children. Combined
  with sorted children, that yields words in alphabetical order, because a word always
  sorts before its own extensions (`dragon` < `dragonfly`).
- **Sorted children.** A `map`'s iteration order is random. `slices.Sorted(maps.Keys(...))`
  collects the keys (an `iter.Seq[rune]`) into a sorted slice. If you don't care about
  order, drop it and range over the map directly, which is faster.
- **An iterator, not a slice.** `WithPrefix` returns an `iter.Seq[string]`, so callers
  can `range` over it, stop early with `break`, or collect it with `slices.Collect`.
  The early-exit plumbing is the same as the BST's `All`: when `yield` returns
  `false`, every level returns `false` straight away.
- **No matches, no problem.** If `find` returns nil, the iterator yields nothing.

## A slice aliasing trap, safely avoided

All children share the same `prefix` slice, and `append(prefix, r)` may write into the
same backing array each time. So sibling calls overwrite each other's last rune! It
works anyway, for two reasons: each child's subtree is completely finished before the
next sibling starts, and `string(prefix)` *copies* the runes into a fresh string before
yielding. If you changed `walk` to yield the `[]rune` itself and stored those slices,
you'd end up with a list of words all mangled into the same few spellings. When in
doubt, copy with `slices.Clone`.

## Your turn: counts and capped suggestions

Two upgrades for the name box:

1. **`CountWithPrefix(prefix)`** should say "37 players start with *dr*" in
   O(len(prefix)), without walking the subtree. Give every node a `count` of
   how many stored words pass through it (the root counts every word), and keep
   it up to date in `Insert`. Careful: inserting a name that's already stored
   must not count it twice, so check for that first with `find`.
2. **`Suggest(prefix, limit)`** returns at most `limit` words in alphabetical
   order, as a slice. It must **stop walking** as soon as it has `limit` words,
   because a popular prefix can have millions of matches. A recursive helper
   that returns `false` to mean "stop" works well, just like `walk` above.
   `Suggest` with a `limit` of 0 or less returns nothing.

## Cost

Finding the start node is O(len prefix). The DFS then visits every node below it once,
plus the cost of building each output string and sorting child keys. In short: you pay
for the prefix and for the results, never for the rest of the trie.
