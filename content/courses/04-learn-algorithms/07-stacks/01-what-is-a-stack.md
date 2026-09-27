---
title: What Is a Stack?
quiz:
  - question: You push `"ava"`, then `"bo"`, then `"cy"` onto an empty stack, then pop once. What comes off?
    options:
      - text: '`"ava"`'
      - text: '`"cy"`'
        correct: true
      - text: '`"bo"`'
    explanation: |
      A stack is last in, first out. `"cy"` was pushed last, so it's on top and
      comes off first.
  - question: Which of these is naturally a stack?
    options:
      - text: A line of customers waiting at a checkout
      - text: The undo history in an editor
        correct: true
      - text: A list of influencers sorted by followers
    explanation: |
      Undo reverses your *most recent* action first, then the one before that:
      last in, first out. A checkout line is first in, first out, which is a
      queue.
---

Clout's campaign editor lets brand managers tweak a campaign: add an
influencer, change the budget, remove a hashtag. They make mistakes, so they
want an **Undo** button. Pressing it should reverse the *most recent* change.
Pressing it again should reverse the one before that.

The data structure for that is a **stack**.

## Last in, first out

Picture a stack of plates. You can only add a plate to the top, and you can
only take the top plate off. The last plate you put on is the first one you
take off. That rule is called **LIFO**: last in, first out.

A stack has a tiny set of operations:

| operation | what it does | cost |
|-----------|--------------|------|
| **push**  | put an item on top | O(1) |
| **pop**   | remove and return the top item | O(1) |
| **peek**  | look at the top item without removing it | O(1) |
| **size** / **is empty** | how many items there are | O(1) |

That's it. No indexing into the middle, no searching. The restriction is the
point: because you can only touch the top, every operation is O(1).

## Undo, step by step

```
action              undo stack (top on the right)
add "ava"           [add ava]
set budget 500      [add ava, budget 500]
add "bo"            [add ava, budget 500, add bo]
UNDO → pops "add bo"  [add ava, budget 500]
UNDO → pops "budget 500"  [add ava]
```

Each undo reverses exactly the right change, in exactly the right order,
without any searching.

## Stacks are everywhere

- **The call stack.** When `main` calls `findMin`, which calls `min`, Go pushes
  a stack frame for each call. When `min` returns, its frame is popped and
  execution resumes in `findMin`. Recursion is just a function pushing frames
  of itself. That's also why deep recursion uses O(depth) memory, as you saw in
  the space-complexity lesson.
- **`defer`.** Deferred calls run in LIFO order when the function returns:

```go
package main

import "fmt"

func main() {
	for _, h := range []string{"ava", "bo", "cy"} {
		defer fmt.Println("closing", h)
	}
	fmt.Println("done")
}
```

```
done
closing cy
closing bo
closing ava
```

- **Browser back buttons**, **matching brackets** in code, **evaluating
  expressions** like `3 * (4 + 5)`, and **depth-first search** through a
  graph of who-follows-whom all use stacks.

## A stack is an idea, not an implementation

"Stack" describes *behaviour*: push, pop, peek, LIFO. It doesn't say how the
items are stored. You could build one on a slice, on a linked list, or on a
fixed-size array. In Go, a slice is by far the most natural choice, because
appending to and removing from the *end* of a slice are both cheap. Let's build
one.
