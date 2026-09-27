---
title: Starts With
quiz:
  - question: |
      The trie holds only `"dragon"`. What do these return?

      ```go
      t.HasPrefix("drag")
      t.HasPrefix("")
      t.HasPrefix("dragons")
      ```
    options:
      - text: '`true`, `true`, `false`'
        correct: true
      - text: '`true`, `false`, `false`'
      - text: '`false`, `true`, `false`'
      - text: '`true`, `true`, `true`'
    explanation: |
      `"drag"` follows existing nodes. The empty prefix doesn't move at all, so `find`
      returns the root, which isn't nil, and every string starts with `""`.
      `"dragons"` falls off the path at `s`.
  - question: |
      With the `count` field from this lesson, you insert `"drake"`, `"dread"` and
      `"drake"` again. What does `CountPrefix("dr")` return?
    options:
      - text: '3'
      - text: '2'
        correct: true
      - text: '1'
      - text: '0'
    explanation: |
      Duplicates must not be counted twice. `Insert` checks `Contains` first and
      returns early for `"drake"` the second time, so only two words pass through
      the `r` node.
exercise:
  starter: |
    package main

    import "fmt"

    // trieNode stores lowercase a-z names only: children[0] is 'a', children[25] is 'z'.
    type trieNode struct {
    	children [26]*trieNode
    	end      bool
    }

    type Trie struct {
    	root trieNode
    }

    // Insert adds a lowercase name to the trie.
    func (t *Trie) Insert(name string) {
    	// ?
    }

    // Contains reports whether name was inserted.
    func (t *Trie) Contains(name string) bool {
    	// ?
    	return false
    }

    // StartsWith reports whether any inserted name starts with prefix.
    func (t *Trie) StartsWith(prefix string) bool {
    	// ?
    	return false
    }

    func main() {
    	var names Trie
    	for _, name := range []string{"dragon", "drake", "dread", "mira"} {
    		names.Insert(name)
    	}
    	fmt.Println(names.Contains("drake"), names.Contains("drak"))  // want true false
    	fmt.Println(names.StartsWith("drak"), names.StartsWith("xq")) // want true false
    }
  solution: |
    package main

    import "fmt"

    // trieNode stores lowercase a-z names only: children[0] is 'a', children[25] is 'z'.
    type trieNode struct {
    	children [26]*trieNode
    	end      bool
    }

    type Trie struct {
    	root trieNode
    }

    func (t *Trie) Insert(name string) {
    	n := &t.root
    	for i := range len(name) {
    		c := name[i] - 'a'
    		if n.children[c] == nil {
    			n.children[c] = &trieNode{}
    		}
    		n = n.children[c]
    	}
    	n.end = true
    }

    func (t *Trie) find(prefix string) *trieNode {
    	n := &t.root
    	for i := range len(prefix) {
    		n = n.children[prefix[i]-'a']
    		if n == nil {
    			return nil
    		}
    	}
    	return n
    }

    func (t *Trie) Contains(name string) bool {
    	n := t.find(name)
    	return n != nil && n.end
    }

    func (t *Trie) StartsWith(prefix string) bool {
    	return t.find(prefix) != nil
    }

    func main() {
    	var names Trie
    	for _, name := range []string{"dragon", "drake", "dread", "mira"} {
    		names.Insert(name)
    	}
    	fmt.Println(names.Contains("drake"), names.Contains("drak"))  // want true false
    	fmt.Println(names.StartsWith("drak"), names.StartsWith("xq")) // want true false
    }
  tests: |
    package main

    import "testing"

    func newTrie(names ...string) *Trie {
    	t := &Trie{}
    	for _, n := range names {
    		t.Insert(n)
    	}
    	return t
    }

    func TestContains(t *testing.T) {
    	tr := newTrie("dragon", "drake", "dread", "dr", "mira")
    	for _, tt := range []struct {
    		name string
    		want bool
    	}{
    		{"dragon", true}, {"drake", true}, {"dread", true}, {"dr", true}, {"mira", true},
    		{"drag", false}, {"d", false}, {"dragons", false}, {"zed", false}, {"mir", false},
    	} {
    		if got := tr.Contains(tt.name); got != tt.want {
    			t.Errorf("Contains(%q) = %v, want %v", tt.name, got, tt.want)
    		}
    	}
    }

    func TestStartsWith(t *testing.T) {
    	tr := newTrie("dragon", "drake", "mira")
    	for _, tt := range []struct {
    		prefix string
    		want   bool
    	}{
    		{"d", true}, {"dra", true}, {"drag", true}, {"dragon", true}, {"mi", true}, {"", true},
    		{"dragons", false}, {"dre", false}, {"x", false}, {"mirab", false},
    	} {
    		if got := tr.StartsWith(tt.prefix); got != tt.want {
    			t.Errorf("StartsWith(%q) = %v, want %v", tt.prefix, got, tt.want)
    		}
    	}
    }

    func TestEmptyTrie(t *testing.T) {
    	var tr Trie
    	if tr.Contains("a") || tr.StartsWith("a") {
    		t.Errorf(`an empty trie should not contain or start with "a"`)
    	}
    }

    func TestSharesPrefixNodes(t *testing.T) {
    	tr := newTrie("ash", "ashe", "asher")
    	nodes := 0
    	var count func(n *trieNode)
    	count = func(n *trieNode) {
    		nodes++
    		for _, c := range n.children {
    			if c != nil {
    				count(c)
    			}
    		}
    	}
    	count(&tr.root)
    	if nodes != 6 {
    		t.Errorf(`"ash", "ashe", "asher" should use 6 nodes (root, a, s, h, e, r), got %d`, nodes)
    	}
    }
---

The username box shows a hint while you type: "no player names start with `xq`" or
"42 names start with `drag`". That's a prefix question, and it's where tries leave
hashmaps behind.

## HasPrefix

You already wrote the hard part. `find` walks the path for any string and returns the
node it ends on. For `Contains` we also checked `end`. For a prefix check, we don't
care whether a word *ends* there, only that the path *exists*:

```go
func (t *Trie) HasPrefix(prefix string) bool {
	return t.find(prefix) != nil
}
```

If the path exists, then at least one word goes through it, because we never create
nodes that don't lead to a word. That's O(len prefix), however many names are stored.

Compare that with a map or slice of names, where you'd have to check every single
one with `strings.HasPrefix`:

```go
for _, name := range allNames { // O(n) names...
	if strings.HasPrefix(name, prefix) { // ...times O(len prefix) each
		return true
	}
}
```

## Counting names with a prefix

To say *how many* names start with `drag`, you could visit every node below `find`'s
result, but there's a neat trick: store a counter on each node for how many words pass
through it. `Insert` bumps the counter on every node along the path.

```go
package main

import "fmt"

type trieNode struct {
	children map[rune]*trieNode
	count    int // words that pass through (or end at) this node
	end      bool
}

type Trie struct{ root trieNode }

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

func (t *Trie) HasPrefix(prefix string) bool { return t.find(prefix) != nil }

func (t *Trie) Insert(word string) {
	if t.Contains(word) {
		return // don't count duplicates
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

func (t *Trie) CountPrefix(prefix string) int {
	if n := t.find(prefix); n != nil {
		return n.count
	}
	return 0
}

func main() {
	var names Trie
	for _, name := range []string{"dragon", "dragonfly", "drake", "dread", "mira", "drake"} {
		names.Insert(name)
	}
	fmt.Println(names.HasPrefix("drag"), names.HasPrefix("xq"))
	fmt.Println(names.CountPrefix("dr"), names.CountPrefix("drag"), names.CountPrefix(""))
}
```

Output:

```
true false
4 2 5
```

`CountPrefix("")` returns the root's count, which is the total number of distinct
names. Notice the `if n := t.find(prefix); n != nil` form: the variable `n` is scoped
to the `if` statement, which keeps it from leaking into the rest of the function.

## The price of the counter

Each node grew by one `int`, and `Insert` now does a `Contains` walk first to avoid
double counting. That's still O(L). Deleting a word would need to decrement the
counters along its path, and could remove nodes whose count drops to zero. This is a
common pattern with trees: store a little extra summary data per node and keep it up
to date on every change, and some queries become instant.

## Your turn

Time to build the lowercase-only variant from the first lesson of this chapter, where
each node has `children [26]*trieNode` instead of a map. Index a child with
`name[i] - 'a'`, so `'a'` is slot 0 and `'z'` is slot 25. In the exercise, complete
`Insert`, `Contains` and `StartsWith` (a prefix check, like `HasPrefix` above). Names
that share a prefix must share nodes, and the empty prefix counts as a prefix of every
name.
