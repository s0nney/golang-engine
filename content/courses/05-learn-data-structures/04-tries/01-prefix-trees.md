---
title: Prefix Trees
quiz:
  - question: In a trie, where is each *character* of a stored word represented?
    options:
      - text: In a node's value field, and the whole word is stored at the leaf
      - text: By the edge you follow from a parent to a child
        correct: true
      - text: In a hash of the whole word stored at the root
      - text: Characters aren't stored; only word lengths are
    explanation: |
      Each edge is labelled with one character. The word is spelled out by the path
      from the root to a node, so a node doesn't need to store the word at all.
  - question: |
      You insert `"drake"` and `"drakes"` into an empty trie. Why do you need an
      `end` flag on nodes?
    options:
      - text: To know where the trie's memory ends
      - text: Because `"drake"` finishes on a node that has a child, so "has no children" can't be how you detect the end of a word
        correct: true
      - text: To mark which nodes are leaves so they can be deleted
      - text: You don't; the node for `e` is enough
    explanation: |
      `"drake"` ends in the middle of the path to `"drakes"`. Without `end`, you couldn't
      tell a stored word from a mere prefix of a longer word, like `"drak"`.
  - question: How many trie nodes (including the root) does it take to store `"ash"`, `"ashe"` and `"asher"`?
    options:
      - text: '6'
        correct: true
      - text: '12'
      - text: '13'
      - text: '4'
    explanation: |
      All three words share the path a → s → h, and each longer word adds one or two
      nodes: root, a, s, h, e, r = 6 nodes. Storing them separately would take
      3 + 4 + 5 = 12 characters.
---

Players pick a username when they sign up, and the name box has two jobs: say whether
a name is taken, and suggest completions as the player types (`dr` → `dragon`,
`drake`, `dread`). A hashmap handles the first job perfectly, but it's useless for the
second: hashing `"dr"` tells you nothing about `"dragon"`. For prefix questions you
want a **trie**.

## What's a trie?

A trie (usually pronounced "try", from re*trie*val) is a tree where each edge is
labelled with a **character**, and a word is spelled out by the path from the root.
Words that share a prefix share the start of their path.

Here's a trie holding `dr`, `dragon`, `drake` and `dread`:

```
(root)
  └─ d
     └─ r ●            <- "dr" is a word
        ├─ a
        │  ├─ g
        │  │  └─ o
        │  │     └─ n ●      "dragon"
        │  └─ k
        │     └─ e ●         "drake"
        └─ e
           └─ a
              └─ d ●         "dread"
```

The `●` marks nodes where a word **ends**. You need that flag because a word can end
partway down another word's path, like `dr` above. "Is this a leaf?" isn't enough.

The name is also why it's called a **prefix tree**: every node represents a prefix,
namely the characters on the path to it, and everything below that node starts with
that prefix.

## A trie node in Go

```go
type trieNode struct {
	children map[rune]*trieNode
	end      bool // a word ends at this node
}

type Trie struct {
	root trieNode
	size int // number of words stored
}
```

Some choices worth explaining:

- **`map[rune]*trieNode`** lets names use any Unicode characters, and only allocates
  space for children that actually exist. A classic alternative for lowercase
  English only is `children [26]*trieNode`, indexed with `r - 'a'`. It's faster but
  uses 26 pointers (208 bytes on 64-bit machines) per node, most of them `nil`.
- **`rune`**, not `byte`: ranging over a string with `for _, r := range s` decodes
  UTF-8 and gives you whole characters. Indexing with `s[i]` gives bytes, which would
  split a name like `"zoë"` into pieces.
- **`root` is a value, not a pointer**, so the zero `Trie` is ready to use. We'll
  create each `children` map lazily, because writing to a nil map panics.

## Costs at a glance

For a word of length L:

| Operation | Trie | Hashmap | Sorted slice |
|---|---|---|---|
| Is `word` stored? | O(L) | O(L) to hash, then O(1) | O(L · log n) |
| Any word with prefix `p`? | O(len p) | O(n · len p), scan everything | O(len p · log n) |
| All words with prefix `p` | O(len p + nodes below it) | O(n · len p) | O(len p · log n + output) |

Notice that a trie's lookup cost doesn't depend on n, the number of words stored.
Finding `"dragon"` takes six steps whether the game has a hundred players or a
hundred million.
