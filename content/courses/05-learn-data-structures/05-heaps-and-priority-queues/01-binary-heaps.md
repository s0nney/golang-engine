---
title: Binary Heaps
quiz:
  - question: In a heap stored in a slice, what's the index of the parent of the node at index `9`?
    options:
      - text: '3'
      - text: '4'
        correct: true
      - text: '5'
      - text: '18'
    explanation: |
      The parent of index `i` is `(i - 1) / 2` with integer division: (9 − 1) / 2 = 4.
      Check it the other way: the children of 4 are 2·4 + 1 = 9 and 2·4 + 2 = 10.
  - question: Which slice is a valid **max**-heap?
    options:
      - text: '`[90, 70, 80, 75, 10]`'
      - text: '`[90, 80, 70, 10, 75]`'
        correct: true
      - text: '`[10, 70, 80, 75, 90]`'
      - text: '`[90, 10, 80, 75, 70]`'
    explanation: |
      Check every parent against its children. In `[90, 80, 70, 10, 75]`: 90 ≥ 80 and
      70; 80 ≥ 10 and 75. Valid. In the first option, 70 at index 1 has child 75 at
      index 3. In the last, 10 at index 1 has children 75 and 70.
  - question: What does the heap property guarantee?
    options:
      - text: The slice is sorted
      - text: Every parent is at least as large as its children (for a max-heap), so the maximum is at the root
        correct: true
      - text: The left child is always smaller than the right child
      - text: Searching for any value is O(log n)
    explanation: |
      A heap is only *partially* ordered: parents beat children, but siblings and
      cousins can be in any order. That's just enough to find the max instantly, and
      cheap enough to maintain in O(log n) per change. Finding an arbitrary value still
      means scanning everything.
---

The leaderboard BST keeps *every* score in order. But the most common question is
simpler: "who's in first place?" Or, for the matchmaking service, "which waiting
player has waited the longest?" When you only ever need the **best** item, a
**heap** does the job with less work and much less memory overhead.

## The heap property

A binary **max-heap** is a binary tree where every parent is **greater than or equal
to** its children. (A **min-heap** flips that: every parent is less than or equal to
its children.)

```
           95
         /    \
       80      90
      /  \    /
    40   75  60
```

The largest value is always at the root. Unlike a BST, there's no left-versus-right
rule: 80 and 90 could swap places and it would still be a heap.

A heap is also a **complete** tree: every level is full except possibly the last,
which fills from left to right. That shape rule is what makes the next trick possible.

## A tree in a slice

Because the tree is complete, you can number the nodes level by level, left to right,
and store them in a plain slice with **no pointers at all**:

```
index:   0   1   2   3   4   5
value: [95, 80, 90, 40, 75, 60]
```

The parent-child links are just arithmetic:

| Want | Formula |
|---|---|
| Left child of `i` | `2*i + 1` |
| Right child of `i` | `2*i + 2` |
| Parent of `i` | `(i - 1) / 2` |

Go's integer division rounds toward zero, which is exactly what the parent formula
needs: the parent of both 3 and 4 is 1.

```go
package main

import "fmt"

func isMaxHeap(h []int) bool {
	for i := 1; i < len(h); i++ {
		if h[(i-1)/2] < h[i] { // parent smaller than child
			return false
		}
	}
	return true
}

func main() {
	scores := []int{95, 80, 90, 40, 75, 60}
	for i, s := range scores {
		var kids []int
		for _, c := range []int{2*i + 1, 2*i + 2} {
			if c < len(scores) {
				kids = append(kids, scores[c])
			}
		}
		if len(kids) > 0 {
			fmt.Println(s, "has children", kids)
		}
	}
	fmt.Println(isMaxHeap(scores), isMaxHeap([]int{95, 80, 90, 85}))
}
```

Output:

```
95 has children [80 90]
80 has children [40 75]
90 has children [60]
true false
```

`[95, 80, 90, 85]` fails because 85 at index 3 is the child of 80 at index 1.

## Why a slice is so good

- **No pointers**: each node is just a value, with no `left`, `right` or `parent` fields.
- **Cache-friendly**: the whole heap sits in one block of memory, so the CPU can
  load it efficiently.
- **Growth is free**: adding a node is `append`, and the tree stays complete by
  construction.

## What heaps are good at

| Operation | Cost |
|---|---|
| Peek at the max | O(1), it's `h[0]` |
| Insert | O(log n) |
| Remove the max | O(log n) |
| Build from n items | O(n) |
| Search for an arbitrary value | O(n) |

A heap is the classic way to implement a **priority queue**: a queue where the item
that comes out next is the one with the highest priority, not the one that arrived
first. The next two lessons implement insert and remove.
