---
title: Walking Trees
quiz:
  - question: |
      Using the `Node` type from this lesson, what does `countFiles` return for this tree?

      ```go
      root := &Node{Name: "docs", Children: []*Node{
          {Name: "a.md"},
          {Name: "guides", Children: []*Node{
              {Name: "b.md"},
              {Name: "old", Children: []*Node{}},
          }},
      }}
      ```
    options:
      - text: '`2`'
      - text: '`3`'
        correct: true
      - text: '`5`'
      - text: '`6`'
    explanation: |
      `countFiles` counts nodes with *no* children: `a.md`, `b.md` and `old`, the
      empty folder. The `Node` type can't tell an empty folder from a file, so the
      answer is 3. A real tool would add an `IsDir` field, which is exactly what
      `fs.DirEntry` has.
  - question: Why is recursion a natural fit for trees?
    options:
      - text: Each child of a tree node is itself a tree, so the same function handles every level
        correct: true
      - text: Trees can't be stored in slices
      - text: Go's compiler optimises recursive tree walks into loops
    explanation: |
      A tree is a recursive data structure. Handle one node, then call the same
      function on each child, and you've handled the whole tree at any depth.
---

Recursion shines when the **data** is recursive. The classic example is a tree,
where every node can have children, and each child is a tree of its own.

Doc2Doc works on whole folders of documents. A folder contains files and other
folders, which contain more files and folders, to any depth. You can't write that as
a fixed number of nested loops, because you don't know how deep it goes.

## A tree type

```go
type Node struct {
	Name     string
	Children []*Node
}
```

A leaf (a file) has no children. A folder has some. The type refers to itself, which
is a sure sign that recursion is coming.

## Walking it

```go
package main

import (
	"fmt"
	"strings"
)

type Node struct {
	Name     string
	Children []*Node
}

func printTree(n *Node, depth int) {
	fmt.Println(strings.Repeat("  ", depth) + n.Name)
	for _, child := range n.Children {
		printTree(child, depth+1)
	}
}

func countFiles(n *Node) int {
	if len(n.Children) == 0 { // base case: a leaf
		return 1
	}
	total := 0
	for _, child := range n.Children {
		total += countFiles(child)
	}
	return total
}

func main() {
	root := &Node{Name: "docs/", Children: []*Node{
		{Name: "README.md"},
		{Name: "guides/", Children: []*Node{
			{Name: "install.md"},
			{Name: "usage.md"},
		}},
		{Name: "api/", Children: []*Node{
			{Name: "v1/", Children: []*Node{{Name: "endpoints.md"}}},
		}},
	}}
	printTree(root, 0)
	fmt.Println("files:", countFiles(root))
}
```

```text
docs/
  README.md
  guides/
    install.md
    usage.md
  api/
    v1/
      endpoints.md
files: 4
```

Notice that `printTree` has no explicit `if` for its base case. When a node has no
children, the `for` loop simply runs zero times and the function returns. The base
case is hidden in the loop.

## Searching a tree

Returning early from recursion takes a little care: you have to pass the result back
up through every level.

```go
func find(n *Node, name string) *Node {
	if n.Name == name {
		return n
	}
	for _, child := range n.Children {
		if found := find(child, name); found != nil {
			return found
		}
	}
	return nil
}
```

If a child's search succeeds, return straight away. Otherwise try the next child.
Only after every child fails do you return `nil`.

## The standard library does this too

For real folders on disk, you'd use `filepath.WalkDir` or `fs.WalkDir`. They walk
the tree for you and call *your* function for each entry. That's recursion wrapped
up in a higher-order function:

```go
err := filepath.WalkDir("docs", func(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if !d.IsDir() && strings.HasSuffix(path, ".md") {
		fmt.Println(path)
	}
	return nil
})
```

Returning `fs.SkipDir` from the callback skips a whole folder, and returning any
other error stops the walk.
