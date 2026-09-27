---
title: Insert and Search
quiz:
  - question: |
      What does this print, using the `Trie` from this lesson?

      ```go
      var t Trie
      t.Insert("dragon")
      fmt.Println(t.Contains("drag"), t.Contains("dragon"), t.Contains("dragons"))
      ```
    options:
      - text: '`true true true`'
      - text: '`true true false`'
      - text: '`false true false`'
        correct: true
      - text: '`false true true`'
    explanation: |
      `"drag"` follows existing nodes, but the `g` node's `end` flag is false, so
      it's only a prefix. `"dragons"` runs out of path at `s` and hits a nil child.
      Only `"dragon"` ends on a node marked `end`.
  - question: In `Insert`, why is `n.children` checked for `nil` before adding a child?
    options:
      - text: Reading from a nil map panics
      - text: Writing to a nil map panics, and new nodes start with a nil `children` map
        correct: true
      - text: To avoid inserting duplicate words
      - text: It isn't necessary; Go creates the map automatically
    explanation: |
      A new `trieNode{}` has a nil map. *Reading* a nil map is fine (that's why
      `find` doesn't need the check), but *writing* one panics. So we `make` it lazily,
      the first time a node gets a child. Leaves never pay for a map.
  - question: |
      Why does `find` work without checking whether `n.children` is nil?

      ```go
      for _, r := range word {
          n = n.children[r]
          if n == nil {
              return nil
          }
      }
      ```
    options:
      - text: Indexing a nil map returns the zero value, which for `*trieNode` is `nil`
        correct: true
      - text: The `children` map is never nil
      - text: Go skips the loop if the map is nil
      - text: It doesn't work; it panics on leaves
    explanation: |
      Reading any key from a nil map returns the value type's zero value. For a
      pointer type that's `nil`, which is exactly the "no such child" signal we want.
---

Time to make the trie do something. Both operations walk down from the root one
character at a time.

## Insert

For each character, follow the matching child, creating it if it doesn't exist.
After the last character, mark the node as the end of a word.

```go
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
	if !n.end {
		n.end = true
		t.size++
	}
}
```

The `if !n.end` check keeps `size` honest: inserting the same username twice
doesn't count it twice.

## Search

Walk the same path. If a child is missing, the word isn't there. If you reach the end
of the word, it's only a *stored word* if `end` is set.

```go
// find returns the node at the end of prefix's path, or nil.
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

func (t *Trie) Contains(word string) bool {
	n := t.find(word)
	return n != nil && n.end
}
```

`find` leans on a handy Go rule: indexing a nil map is allowed and returns the zero
value. For `map[rune]*trieNode` the zero value is `nil`, so leaves (whose map was
never created) naturally report "no child".

We split out `find` because the next two lessons reuse it.

## Try it

```go
package main

import "fmt"

type trieNode struct {
	children map[rune]*trieNode
	end      bool
}

type Trie struct {
	root trieNode
	size int
}

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
	if !n.end {
		n.end = true
		t.size++
	}
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

func (t *Trie) Contains(word string) bool {
	n := t.find(word)
	return n != nil && n.end
}

func main() {
	var names Trie
	for _, name := range []string{"dragon", "drake", "dread", "zoë", "drake"} {
		names.Insert(name)
	}
	fmt.Println(names.size)
	for _, try := range []string{"drake", "drak", "zoë", "zoe"} {
		fmt.Printf("%-5s taken? %v\n", try, names.Contains(try))
	}
}
```

Output:

```
4
drake taken? true
drak  taken? false
zoë   taken? true
zoe   taken? false
```

`"drake"` was inserted twice but counted once. `"zoë"` works because we range over
runes: the `ë` is one edge, even though it's two bytes in UTF-8. (`%-5s` pads by
runes, which is why the columns line up.)

## Cost

Both operations do O(1) work per character, so they're **O(L)** for a word of length
L, and completely independent of how many names are stored.

## Case sensitivity

Should `"Drake"` be the same player as `"drake"`? For usernames, almost certainly.
Normalise before you touch the trie, for example with `strings.ToLower(name)`, and do
it in *both* `Insert` and `Contains`, or lookups will mysteriously fail.
