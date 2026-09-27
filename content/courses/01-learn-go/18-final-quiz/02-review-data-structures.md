---
title: 'Review: Data Structures and Pointers'
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	queue := []string{"a", "b", "c", "d"}
      	batch := queue[:2]
      	batch = append(batch, "X")
      	fmt.Println(queue)
      }
      ```
    options:
      - text: '`[a b c d]`'
      - text: '`[a b X d]`'
        correct: true
      - text: '`[a b X c d]`'
      - text: '`[a b c d X]`'
    explanation: |
      `batch` shares `queue`'s array and has spare capacity (4), so `append`
      writes `X` into the shared array at index 2, overwriting `c`. This is
      the slice aliasing trap. Use `slices.Clone` when you need an
      independent copy.
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      type account struct {
      	credits int
      }

      func (a account) refund(n int) { a.credits += n }

      func main() {
      	accounts := map[string]*account{"alice": {credits: 5}}
      	accounts["alice"].credits += 10
      	accounts["alice"].refund(100)
      	fmt.Println(accounts["alice"].credits)
      }
      ```
    options:
      - text: '`115`'
      - text: '`5`'
      - text: '`15`'
        correct: true
      - text: It panics, because you can't change a struct through a map
    explanation: |
      The map holds a pointer, so `accounts["alice"].credits += 10` changes
      the real account: 15. But `refund` has a value receiver, so it adds 100
      to a copy that is thrown away. Result: 15.
  - question: Which of these operations panics?
    options:
      - text: Reading a missing key from a nil map
      - text: Appending to a nil slice
      - text: Writing a key to a nil map
        correct: true
      - text: Calling `len` on a nil slice
    explanation: |
      Nil slices and nil maps are fine to read, `len` and (for slices)
      `append` to. Only *writing* to a nil map panics, with
      `assignment to entry in nil map`. Create maps with `make` or a literal
      first.
---

This review covers arrays and slices, maps, structs and pointers: the building blocks you'll use to model data in every Go program.

## Quick recap

Here's a sketch of how Textio might track a customer's message history, combining a struct, a pointer receiver, a map of slices and the `slices` package:

```go
package main

import (
	"fmt"
	"slices"
)

type inbox struct {
	byUser map[string][]string
}

func newInbox() *inbox {
	return &inbox{byUser: make(map[string][]string)}
}

func (in *inbox) add(user, body string) {
	in.byUser[user] = append(in.byUser[user], body)
}

func main() {
	in := newInbox()
	in.add("alice", "Your code is 4821")
	in.add("bob", "Your order shipped")
	in.add("alice", "Your order arrived")

	history := in.byUser["alice"]
	fmt.Println(len(history), slices.Contains(history, "Your code is 4821"))

	if _, ok := in.byUser["carol"]; !ok {
		fmt.Println("carol has no messages")
	}
}
```

```text
2 true
carol has no messages
```

The gotchas worth remembering from these chapters:

- Slicing and passing slices share the underlying array. `append` may or may not reallocate, so always write `s = append(s, ...)`.
- A missing map key returns the zero value. Use comma-ok to tell "missing" apart from "zero".
- Writing to a nil map panics. Always `make` your maps.
- Map iteration order is random. Sort the keys when order matters.
- Go passes everything by value. Methods that modify their receiver need a pointer receiver.
- Every pointer might be nil. Check before you dereference.

Now take the quiz.
