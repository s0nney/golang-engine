---
title: Insert and Recolour
quiz:
  - question: Why is a newly inserted node coloured **red**?
    options:
      - text: Red nodes are faster to compare
      - text: Adding a red node can't change any path's black count, so rule 5 stays intact
        correct: true
      - text: The root must always be red
      - text: It doesn't matter; either colour works equally well
    explanation: |
      A new black node would add one black to every path through it and break rule 5
      immediately, which is hard to fix. A red node leaves black counts alone. At
      worst it breaks rule 4 (red parent, red child), and that's what the fix-up repairs.
  - question: |
      The new red node's parent is red, and its **uncle** is also red. What does
      the fix-up do?
    options:
      - text: Rotates the grandparent
      - text: Deletes the uncle
      - text: Makes the parent and uncle black and the grandparent red, then continues from the grandparent
        correct: true
      - text: Nothing; two reds in a row are allowed if the uncle is red
    explanation: |
      That's the recolouring case. Flipping the colours keeps every path's black
      count the same and removes the red-red pair, but the grandparent is now red and
      might have a red parent, so the loop moves two levels up and checks again.
  - question: |
      Using the `fixInsert` from this lesson, what colour does the root end up after inserting `10`, then `20`, then `30` into an empty tree?
    options:
      - text: '`20` is the root and it''s black'
        correct: true
      - text: '`10` is the root and it''s black'
      - text: '`20` is the root and it''s red'
      - text: '`30` is the root and it''s black'
    explanation: |
      30 is red with a red parent (20) and a nil, black uncle. It's the "line" case,
      so 20 turns black, 10 turns red, and the tree rotates left at 10. Now 20 is the
      black root with red children 10 and 30.
---

Inserting into a red-black tree happens in two steps:

1. Insert the value exactly as in a normal BST, and colour the new node **red**.
2. Run a **fix-up** that repairs any broken rule, walking up from the new node.

Why red? A red node adds nothing to any path's black count, so rule 5 stays true.
The only rules that can break are rule 2 (if the new node is the root) and rule 4
(if its parent is also red). Rule 2 is trivial: paint the root black at the end.
Rule 4 is where the work is.

## The family

When the new node `z` has a red parent `p`, look at the grandparent `g` and the
**uncle** `u` (the grandparent's other child). `g` must exist and must be black: `p`
is red, so it isn't the root, and the tree was valid before, so `p`'s parent can't
be red.

### Case 1: the uncle is red → recolour

```
        g(B)                     g(R)   <- might now clash with its own parent
       /    \                   /    \
     p(R)   u(R)     --->     p(B)   u(B)
     /                        /
   z(R)                     z(R)
```

Flip the colours: parent and uncle become black, grandparent becomes red. Every
path through `g` still has the same black count, and `z` no longer has a red parent.
But `g` is now red and *its* parent might be red, so set `z = g` and loop again.
This case can repeat all the way to the root, two levels at a time.

### Case 2: the uncle is black and z is an "inner" child → rotate into case 3

If `p` is a left child but `z` is a *right* child (or the mirror image), the three
nodes form a triangle. Rotate at `p` to straighten them into a line, then carry on
with case 3 treating the old `p` as `z`.

```
      g                 g
     /                 /
    p        --->     z
     \               /
      z             p
```

### Case 3: the uncle is black and z is an "outer" child → rotate the grandparent

Now `g`, `p` and `z` are in a straight line. Recolour `p` black and `g` red, then
rotate at `g` the other way. `p` takes the grandparent's place as a black node with
two red children, and the loop ends because `p` is black.

```
        g(B)               p(B)
       /                  /    \
     p(R)      --->     z(R)   g(R)
     /
   z(R)
```

## The code

Here's the fix-up for the case where `p` is a left child. The `else` branch is the
mirror image, with every `left`/`right` and `rotateLeft`/`rotateRight` swapped.

```go
func (t *RBTree[T]) fixInsert(z *rbNode[T]) {
	for isRed(z.parent) {
		p := z.parent
		g := p.parent // p is red, so it isn't the root, so g exists
		if p == g.left {
			uncle := g.right
			if isRed(uncle) { // case 1: recolour and move up
				p.red, uncle.red, g.red = false, false, true
				z = g
				continue
			}
			if z == p.right { // case 2: triangle, rotate into a line
				t.rotateLeft(p)
				z, p = p, z
			}
			p.red, g.red = false, true // case 3: line, rotate the grandparent
			t.rotateRight(g)
		} else {
			// mirror image of the above
		}
	}
	t.root.red = false // rule 2
}
```

A few Go details worth noticing:

- `isRed(uncle)` is safe when `uncle` is `nil`, because `isRed` checks for `nil`
  first. A missing uncle counts as black, exactly as rule 3 says.
- `p.red, uncle.red, g.red = false, false, true` is a tuple assignment: all three
  colours change in one statement.
- After the case-2 rotation, `z` and `p` have swapped roles in the tree, so we swap
  the variables with `z, p = p, z`.

## How much work is that?

Case 1 only recolours and moves up two levels, so it runs at most O(log n) times.
Cases 2 and 3 do at most **two rotations in total** and then the loop ends. So an
insert is O(log n) for the BST descent plus O(log n) for the fix-up, which is
O(log n) overall.
