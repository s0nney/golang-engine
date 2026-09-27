---
title: The Call Stack
quiz:
  - question: What happens when a recursive Go function has no base case?
    options:
      - text: The goroutine's stack keeps growing until it hits the limit, and the program crashes with a stack overflow
        correct: true
      - text: The compiler refuses to build it
      - text: It runs forever using constant memory
      - text: Go automatically converts it into a loop
    explanation: |
      Every call pushes a new frame onto the stack. Go grows goroutine stacks on
      demand, but only up to a maximum (1 GB on 64-bit systems by default). Past
      that, the runtime aborts with `fatal error: stack overflow`.
  - question: |
      What does this print?

      ```go
      func depth(n int) {
          if n == 0 {
              return
          }
          depth(n - 1)
          fmt.Print(n, " ")
      }

      func main() { depth(3) }
      ```
    options:
      - text: '`3 2 1 `'
      - text: '`0 1 2 3 `'
      - text: '`1 2 3 `'
        correct: true
      - text: Nothing
    explanation: |
      The print happens *after* the recursive call returns. `depth(1)` finishes
      first and prints 1, then control unwinds to `depth(2)`, then `depth(3)`.
---

To understand recursion you need to understand the **call stack**.

Every time a function is called, Go creates a **stack frame** for it, a small chunk
of memory holding its parameters, local variables and the spot to return to. Frames
are stacked on top of each other: the newest call is on top, and when it returns,
its frame is popped off and the caller continues.

## Watching the stack

This program prints as it goes down *and* as it comes back up:

```go
package main

import (
	"fmt"
	"strings"
)

func countSections(level, max int) int {
	pad := strings.Repeat("  ", level)
	fmt.Printf("%senter level %d\n", pad, level)
	if level == max {
		fmt.Printf("%sbase case\n", pad)
		return 1
	}
	n := 1 + countSections(level+1, max)
	fmt.Printf("%sreturn %d from level %d\n", pad, n, level)
	return n
}

func main() {
	fmt.Println("total:", countSections(0, 2))
}
```

```text
enter level 0
  enter level 1
    enter level 2
    base case
  return 2 from level 1
return 3 from level 0
total: 3
```

At the deepest point three frames are on the stack at once, each with its own `level`,
`pad` and `n`. They don't interfere, because each call has its own copies of its
local variables.

## Stack overflow

If there's no base case, or the input never shrinks, frames pile up forever:

```go
func forever(n int) int {
	return forever(n + 1) // no base case!
}
```

Go stacks start small (a few kilobytes) and grow automatically, which is why Go
handles much deeper recursion than many languages. But there's a ceiling: by default
1 GB on 64-bit systems. Hit it and the program dies:

```text
runtime: goroutine stack exceeds 1000000000-byte limit
...
fatal error: stack overflow
```

That's a `fatal error`, not a normal panic, so `recover` can't catch it. The whole
program stops.

## Stack traces are call stacks

When your program panics, Go prints the call stack, one frame per line, newest first.
A panic deep inside recursion shows the same function over and over, which is a
strong hint that your base case is wrong.

## The cost of a frame

Each call costs a little time (setting up the frame) and a little memory. For a few
hundred or even a few thousand levels, that's nothing. For a million levels, a loop is
far cheaper. Keep that in mind; it's the main reason Go programmers prefer loops for
long linear problems.
