---
title: Autocomplete
quiz:
  - question: |
      The trie holds 50,000 names starting with `"d"`. What's the main benefit of
      `Autocomplete` ranging over `WithPrefix` and breaking out, rather than
      collecting every match and slicing off the first 3?
    options:
      - text: The results come out in a different order
      - text: The DFS stops as soon as it has 3 names, instead of visiting all 50,000 matches
        correct: true
      - text: It uses a hashmap internally
      - text: There's no benefit; both visit every node
    explanation: |
      `break` makes `yield` return `false`, which unwinds the whole DFS at once. Only
      the nodes on the way to the first 3 names are visited. With `slices.Collect`
      followed by `[:3]`, you'd build all 50,000 strings and throw most away.
  - question: |
      Suppose you delete the `if limit <= 0` guard from `Autocomplete`. What does
      `Autocomplete("dr", 0)` return now?

      ```go
      var out []string
      for name := range t.WithPrefix(prefix) {
          out = append(out, name)
          if len(out) == limit {
              break
          }
      }
      return out
      ```
    options:
      - text: An empty slice
      - text: Exactly one name
      - text: Every name that starts with `dr`
        correct: true
      - text: It panics
    explanation: |
      The check runs *after* each append, so `len(out)` is already 1 the first time
      it's compared with 0, and it never equals 0 again. The loop runs to the end and
      returns every match. Edge cases like 0 and negative limits deserve a test.
  - question: Players want suggestions ranked by **popularity** rather than alphabetically. What changes?
    options:
      - text: Nothing; the DFS already finds the most popular names first
      - text: 'You can no longer stop after the first `limit` matches: you must consider every match to know which are the most popular'
        correct: true
      - text: Tries can't store popularity, so you must switch to a hashmap
      - text: You need to sort the whole trie on every keystroke
    explanation: |
      The DFS order is alphabetical, so the most popular name could be the very last
      match. You have to look at all of them and keep the best `limit`, which is the
      "top k" problem that a heap solves efficiently (next chapter). Real systems
      also cache the top suggestions on each node.
---

Let's finish the username box. As the player types each letter, show at most three
suggestions. With `WithPrefix` from the last lesson, that's a few lines:

```go
func (t *Trie) Autocomplete(prefix string, limit int) []string {
	if limit <= 0 {
		return nil
	}
	var out []string
	for name := range t.WithPrefix(prefix) {
		out = append(out, name)
		if len(out) == limit {
			break // stops the DFS right here
		}
	}
	return out
}
```

The `break` is doing real work. It makes `yield` return `false`, and our `walk`
returns `false` all the way up, so the DFS stops the moment it has enough names. If
50,000 names start with `d`, typing `d` still only walks down to the first three.

## Simulating a player typing

```go
package main

import (
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"
)

type trieNode struct {
	children map[rune]*trieNode
	end      bool
}

type Trie struct{ root trieNode }

func (t *Trie) Insert(word string) {
	n := &t.root
	for _, r := range strings.ToLower(word) {
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

func (t *Trie) WithPrefix(prefix string) iter.Seq[string] {
	return func(yield func(string) bool) {
		prefix = strings.ToLower(prefix)
		if n := t.find(prefix); n != nil {
			n.walk([]rune(prefix), yield)
		}
	}
}

func (t *Trie) Autocomplete(prefix string, limit int) []string {
	if limit <= 0 {
		return nil
	}
	var out []string
	for name := range t.WithPrefix(prefix) {
		out = append(out, name)
		if len(out) == limit {
			break
		}
	}
	return out
}

func main() {
	var names Trie
	for _, name := range []string{
		"Dragon", "dragonfly", "drake", "dread", "druid", "dragoon", "mira", "dr",
	} {
		names.Insert(name)
	}

	typed := "Drago"
	for i := 1; i <= len(typed); i++ {
		fmt.Printf("%-8q -> %v\n", typed[:i], names.Autocomplete(typed[:i], 3))
	}
	fmt.Println(names.Autocomplete("x", 3) == nil)
}
```

Output:

```
"D"      -> [dr dragon dragonfly]
"Dr"     -> [dr dragon dragonfly]
"Dra"    -> [dragon dragonfly dragoon]
"Drag"   -> [dragon dragonfly dragoon]
"Drago"  -> [dragon dragonfly dragoon]
true
```

The last line is `true` because no name starts with `x`, so `Autocomplete` returned a
nil slice. This version lowercases in both `Insert` and `WithPrefix`, so `"Dragon"` and `"Dr"`
match regardless of case. The loop slices `typed[:i]` by *bytes*, which is fine here
because every character is ASCII. For names with characters like `ë`, loop over runes
instead.

## Making it production-ready

A real game would add a few things:

- **Ranking.** Alphabetical isn't what players want; popular or recently active names
  are. Store a score on each end node, visit all matches, and keep the best three. The
  next chapter's heaps do "keep the best k" in O(n log k).
- **Caching.** Popular prefixes like `d` get typed constantly. Many systems store the
  top suggestions directly on each trie node, updated on insert.
- **Debouncing.** Don't query on every keystroke; wait until the player pauses for
  ~100 ms.
- **Concurrency.** If signups insert names while other goroutines autocomplete, guard
  the trie with a `sync.RWMutex`, since its maps aren't safe for concurrent writes.

## Further reading

- [Go by Example: Range over Iterators](https://gobyexample.com/range-over-iterators)
