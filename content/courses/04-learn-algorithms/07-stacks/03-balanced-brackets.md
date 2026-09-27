---
title: Balanced Brackets
quiz:
  - question: Which of these strings is balanced?
    options:
      - text: '`([)]`'
      - text: '`{[()()]}`'
        correct: true
      - text: '`(()`'
      - text: '`)(`'
    explanation: |
      `{[()()]}` closes every bracket in the reverse order it was opened.
      `([)]` has the right counts but closes `(` before `[`, `(()` leaves one
      open, and `)(` closes before anything is open.
  - question: |
      `isBalanced` finishes the loop over `"(()"` without returning early. Why
      must it still check the stack before returning `true`?
    options:
      - text: The stack might contain a closing bracket
      - text: A non-empty stack means some brackets were opened and never closed
        correct: true
      - text: It doesn't need to; reaching the end means the string is balanced
    explanation: |
      Every opener is pushed and only popped by its matching closer. After
      `"(()"`, one `(` is still on the stack, so the string is unbalanced. The
      final check is `return stack.Len() == 0`.
---

Clout lets brands write post templates like this:

```
Check out {{handle}}'s new drop! [Shop now](https://shop.example/{{code}})
```

Before sending a template to thousands of influencers, we want to catch typos
like a missing `}` or a `]` in the wrong place. The core question: **are the
brackets balanced?**

## What balanced means

A string is balanced if every opening bracket `(`, `[`, `{` has a matching
closing bracket of the same kind, and they're **properly nested**: whatever was
opened most recently must be closed first.

- `{[()]}`: balanced
- `()[]{}`: balanced
- `([)]`: **not** balanced. The `(` is closed by `]`.
- `((`: **not** balanced. Two are never closed.
- `)(`: **not** balanced. The first `)` has nothing to close.

"The most recently opened must be closed first" is LIFO in disguise. That's
your cue to reach for a stack.

## The algorithm

Walk the string one character at a time:

1. If it's an **opening** bracket, push it.
2. If it's a **closing** bracket, pop the top of the stack. If the stack was
   empty, or the popped bracket isn't the matching opener, the string is
   unbalanced.
3. Ignore every other character.

At the end, the string is balanced only if the stack is **empty**. Anything
left over was opened and never closed.

## In Go

```go
package main

import "fmt"

type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }
func (s *Stack[T]) Len() int { return len(s.items) }
func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

var openerFor = map[rune]rune{')': '(', ']': '[', '}': '{'}

func isBalanced(text string) bool {
	var stack Stack[rune]
	for _, r := range text {
		switch r {
		case '(', '[', '{':
			stack.Push(r)
		case ')', ']', '}':
			top, ok := stack.Pop()
			if !ok || top != openerFor[r] {
				return false
			}
		}
	}
	return stack.Len() == 0
}

func main() {
	templates := []string{
		"Check out {{handle}}'s new drop! [Shop now](https://shop.example/{{code}})",
		"{[()()]}",
		"([)]",
		"Big news {{handle}! (link in bio)",
		")(",
	}
	for _, t := range templates {
		fmt.Println(isBalanced(t), t)
	}
}
```

```
true Check out {{handle}}'s new drop! [Shop now](https://shop.example/{{code}})
true {[()()]}
false ([)]
false Big news {{handle}! (link in bio)
false )(
```

(The stack here skips zeroing popped slots, since a `rune` holds no pointers
for the garbage collector to worry about.)

## Details worth noticing

- **`range` over a string yields runes**, not bytes. Our brackets are all
  single-byte ASCII, so bytes would work too, but ranging over runes means
  emoji-heavy influencer captions don't trip us up.
- **The map `openerFor`** turns three `if` branches into one lookup, and makes
  adding a new pair (say `<` and `>`) a one-line change.
- **Early return.** The moment we find a mismatch, we're done. No need to scan
  the rest.
- **The final check** catches leftovers like `((`. Forgetting it is the most
  common bug in this exercise.

## Complexity

We visit each of the `n` characters once, and each push or pop is O(1)
(amortized for push). So the whole check is **O(n) time**. In the worst case
(a string of only opening brackets) the stack holds all `n` of them, so it's
**O(n) space**.

Could you do it with a counter instead of a stack? For a *single* kind of
bracket, yes: add 1 for `(`, subtract 1 for `)`, and fail if it ever goes
negative. But with several kinds, a counter can't tell that `([)]` is wrong.
You need to remember the *order* things were opened in, and that's exactly
what a stack remembers.
