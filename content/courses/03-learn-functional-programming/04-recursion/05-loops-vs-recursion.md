---
title: Loops vs Recursion
quiz:
  - question: What is *tail-call optimisation* (TCO)?
    options:
      - text: Reusing the current stack frame when a function's very last action is to call another function, so tail recursion runs in constant stack space
        correct: true
      - text: Removing unused functions from the binary
      - text: Running recursive calls in parallel goroutines
      - text: Caching the results of recursive calls
    explanation: |
      With TCO, a call in tail position becomes a jump, so the stack doesn't grow.
      Scheme and some other languages guarantee it. The Go compiler doesn't do it,
      so every recursive call in Go adds a frame, even a tail call.
  - question: Which task is the best fit for recursion in Go?
    options:
      - text: Walking a nested section tree that is at most a few dozen levels deep
        correct: true
      - text: Summing a slice of ten million word counts
      - text: Reading every line of a large log file
      - text: Counting down from one billion
    explanation: |
      Recursion is clearest for recursive data, and a shallow tree costs only a
      few stack frames. Long linear jobs like the others would make a million or
      more frames for no benefit, so a loop is better.
---

In languages like Haskell or Scheme, recursion is *the* way to loop. Those languages
have **tail-call optimisation** (TCO): if the last thing a function does is call
itself, the compiler reuses the current stack frame instead of pushing a new one.
Tail recursion then costs no more than a loop.

**Go doesn't do TCO.** The compiler never turns a tail call into a jump, so every
recursive call adds a frame, even when it's the very last thing the function does.

## A tail-recursive sum

```go
package main

import "fmt"

// Tail recursive: the recursive call is the last action.
func sumTo(n, acc int) int {
	if n == 0 {
		return acc
	}
	return sumTo(n-1, acc+n)
}

// The loop version.
func sumToLoop(n int) int {
	acc := 0
	for i := 1; i <= n; i++ {
		acc += i
	}
	return acc
}

func main() {
	fmt.Println(sumTo(1_000_000, 0))
	fmt.Println(sumToLoop(1_000_000))
}
```

```text
500000500000
500000500000
```

Both give the same answer, but the recursive version pushes a million stack frames
first. Go's growable stacks survive this, yet it's slower and uses tens of megabytes
of stack. Push `n` into the hundreds of millions and it crashes with a stack overflow,
while the loop keeps working.

## Rule of thumb

Use **recursion** when:

- The data is recursive (trees, nested sections, JSON), and
- The depth is modest, such as the depth of a folder tree or nested headings.

Use a **loop** when:

- You're walking a flat sequence (lines, words, files in a list), or
- The depth could be huge or is controlled by untrusted input. Someone could send
  Doc2Doc a document with ten million nested lists to crash it.

## Recursion without recursion

Any recursive algorithm can be rewritten with a loop and an explicit **stack**
(a slice you push to and pop from). Here's the folder walk from earlier, done
iteratively:

```go
func countFiles(root *Node) int {
	count := 0
	stack := []*Node{root}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1] // pop
		if len(n.Children) == 0 {
			count++
		}
		stack = append(stack, n.Children...) // push
	}
	return count
}
```

The stack now lives on the heap, where it can grow as big as memory allows, and a
malicious input can't overflow the goroutine's stack. The price is readability: the
recursive version said what it meant in fewer lines.

## The Go verdict

Recursion is a tool, not a religion. Go programmers happily write recursive tree
walkers, parsers and `filepath.WalkDir` callbacks. They rarely write recursive list
processing, because `for` does that job better in Go. Knowing both styles, and when
to switch, is the real skill.
