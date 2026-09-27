---
title: Maps of Slices
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	inbox := map[string][]string{}
      	inbox["alice"] = append(inbox["alice"], "hi")
      	inbox["alice"] = append(inbox["alice"], "you there?")
      	inbox["bob"] = append(inbox["bob"], "yo")
      	fmt.Println(len(inbox), len(inbox["alice"]), len(inbox["carol"]))
      }
      ```
    options:
      - text: '`3 2 0`'
      - text: It panics, because `inbox["alice"]` doesn't exist the first time
      - text: '`2 2 0`'
        correct: true
      - text: '`2 3 0`'
    explanation: |
      A missing key gives a nil slice, and appending to a nil slice works, so
      no special setup is needed. There are two keys (`alice` and `bob`), and
      alice has 2 messages. Looking up `carol` doesn't add her to the map; it
      just returns a nil slice with length 0.
  - question: 'Why can''t you use `map[[]string]int`?'
    options:
      - text: Because slices can't be compared with `==`, so they can't be map keys
        correct: true
      - text: Because maps can only have string keys
      - text: Because slices are always nil
    explanation: |
      Map keys must be comparable with `==`. Slices, maps and functions
      aren't, so they can't be keys. They're perfectly fine as *values*.
---

Map values can be any type, including slices. A **map of slices** groups many items under one key, and it's one of the most useful data structures you'll build. Textio uses one to keep each user's message history.

## Grouping with append

The trick is that a missing key returns a **nil slice**, and you can `append` to a nil slice. So you don't need to check whether a key exists before adding to it:

```go
package main

import "fmt"

func main() {
	senders := []string{"alice", "bob", "alice"}
	texts := []string{"Your code is 4821", "Your order shipped", "Your order arrived"}

	byUser := make(map[string][]string)
	for i, user := range senders {
		byUser[user] = append(byUser[user], texts[i])
	}

	fmt.Println(len(byUser["alice"]), byUser["alice"])
	fmt.Println(len(byUser["bob"]), byUser["bob"])
}
```

```text
2 [Your code is 4821 Your order arrived]
1 [Your order shipped]
```

The key line is:

```go
byUser[user] = append(byUser[user], texts[i])
```

Read it as: "take this user's slice (or nil if they're new), append the message, and store the result back under their name". Storing it back is essential, because `append` may return a new slice.

## A friendlier example

Let's say Textio tracks which phone numbers each customer has registered:

```go
package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	numbers := map[string][]string{}

	register := func(user, phone string) {
		numbers[user] = append(numbers[user], phone)
	}

	register("alice", "+1-555-0100")
	register("bob", "+1-555-0199")
	register("alice", "+44-20-7946-0000")

	for _, user := range slices.Sorted(maps.Keys(numbers)) {
		fmt.Printf("%s has %d number(s): %v\n", user, len(numbers[user]), numbers[user])
	}
}
```

```text
alice has 2 number(s): [+1-555-0100 +44-20-7946-0000]
bob has 1 number(s): [+1-555-0199]
```

Two new things here:

- `register := func(...) { ... }` creates a function and stores it in a variable. Functions are values in Go, and this one can use the `numbers` map from the surrounding code. A function like this is called a **closure**.
- We sorted the keys before printing so the output is predictable (remember: map order is random).

## Removing an item from a user's slice

Changing one user's slice works just like any slice, as long as you store the result back:

```go
numbers["alice"] = slices.DeleteFunc(numbers["alice"], func(p string) bool {
	return p == "+1-555-0100"
})
```

`slices.DeleteFunc` removes every element for which the function returns `true`.

## Maps of maps

Values can be maps too. For example, counts of messages per user per day:

```go
counts := map[string]map[string]int{}
```

But careful: the inner maps start as `nil`, and writing to a nil map panics. You have to create each inner map before using it:

```go
if counts["alice"] == nil {
	counts["alice"] = make(map[string]int)
}
counts["alice"]["monday"]++
```

That extra bookkeeping is one reason maps of slices are more common than maps of maps. When you need several fields per key, a map of **structs** (coming up next chapter) is usually cleaner still.
