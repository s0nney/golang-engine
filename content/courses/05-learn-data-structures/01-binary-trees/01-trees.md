---
title: Trees
quiz:
  - question: In a tree, what is a *leaf*?
    options:
      - text: The node at the very top of the tree
      - text: Any node with exactly one child
      - text: A node with no children
        correct: true
      - text: A node that has been deleted
    explanation: |
      A leaf is a node at the bottom edge of the tree, with no children at all.
      The single node at the top is the *root*.
  - question: |
      Using the skill tree from this lesson, what does this print?

      ```go
      tree := &Node{Name: "Mage", Children: []*Node{
          {Name: "Fireball", Children: []*Node{{Name: "Meteor"}, {Name: "Flame Wall"}}},
          {Name: "Frost Bolt"},
      }}
      fmt.Println(count(tree))
      ```
    options:
      - text: '3'
      - text: '4'
      - text: '5'
        correct: true
      - text: '2'
    explanation: |
      `count` returns 1 for the node itself plus the counts of all its children.
      Mage, Fireball, Meteor, Flame Wall and Frost Bolt make 5 nodes.
  - question: A tree has 10 nodes. How many edges does it have?
    options:
      - text: '9'
        correct: true
      - text: '10'
      - text: '11'
      - text: It depends on the shape of the tree
    explanation: |
      Every node except the root has exactly one edge pointing to it from its parent.
      So a tree with `n` nodes always has `n - 1` edges, whatever its shape.
---

Welcome to the second half of data structures and algorithms! You've already met
*linear* structures: slices, stacks, queues and linked lists. Each item has at most
one "next" item, so the data forms a line.

In this course you're the backend engineer at a small game studio. The game needs a
leaderboard, a username search box with autocomplete, and a world map of zones the
players can walk between. None of those are lines. Let's start with **trees**.

## What's a tree?

A tree is a set of **nodes** joined by **edges**, where each node can point to
*several* children. It's a linked list that branches.

```
            Warrior            <- root (depth 0)
           /       \
   Shield Bash    Cleave       <- depth 1
        |
    Stun Wall                  <- leaf (depth 2)
```

That's the Warrior class's skill tree. You unlock "Shield Bash" before "Stun Wall",
and the tree shape captures that order.

## Vocabulary

- **Root**: the single node at the top. It has no parent.
- **Parent** and **child**: an edge joins a parent to its child. Every node except
  the root has exactly one parent.
- **Leaf**: a node with no children.
- **Siblings**: nodes that share a parent ("Shield Bash" and "Cleave").
- **Subtree**: any node together with everything below it. Every subtree is a tree
  too, which is why recursion fits trees so well.
- **Depth** of a node: the number of edges from the root down to it.
- **Height** of a tree: the depth of its deepest node.

Since every node except the root has exactly one incoming edge, a tree with `n`
nodes always has `n - 1` edges. A tree also has **no cycles**: you can't follow
edges downward and end up back where you started.

## Trees in Go

A general tree node holds a value and a slice of children:

```go
package main

import "fmt"

type Node struct {
	Name     string
	Children []*Node
}

// count returns how many nodes are in the tree rooted at n.
func count(n *Node) int {
	total := 1 // this node
	for _, c := range n.Children {
		total += count(c) // plus every node in each subtree
	}
	return total
}

func main() {
	warrior := &Node{Name: "Warrior", Children: []*Node{
		{Name: "Shield Bash", Children: []*Node{{Name: "Stun Wall"}}},
		{Name: "Cleave"},
	}}
	fmt.Println(count(warrior))
}
```

This prints `4`. Notice the recursion: to count a tree, count the root and then
count each of its subtrees. Almost every tree algorithm in this chapter has that
shape: do something with this node, then recurse into the children.

Inside the slice literal you can write `{Name: "Cleave"}` instead of
`&Node{Name: "Cleave"}`. Go lets you drop the `&Node` from elements of a
`[]*Node` literal because the type is already known.

## Trees are everywhere

You've been using trees all along:

- Your file system (folders containing folders)
- HTML documents (the DOM)
- Go's own compiler, which parses your code into an *abstract syntax tree*
- Org charts, tournament brackets, and game scene graphs

In the next lesson we'll narrow things down to the most useful tree in computer
science: the **binary search tree**.
