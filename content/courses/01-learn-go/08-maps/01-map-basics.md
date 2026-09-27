---
title: Map Basics
quiz:
  - question: |
      What happens when this program runs?

      ```go
      package main

      import "fmt"

      func main() {
      	var credits map[string]int
      	fmt.Println(credits["alice"])
      	credits["alice"] = 10
      	fmt.Println(credits["alice"])
      }
      ```
    options:
      - text: It prints `0` and then `10`
      - text: It prints `0` and then panics
        correct: true
      - text: It panics on the first `Println`
      - text: It doesn't compile
    explanation: |
      `credits` is a nil map, because it was declared but never created with
      `make` or a literal. *Reading* from a nil map is fine and returns the
      zero value, `0`. *Writing* to a nil map panics with
      `assignment to entry in nil map`.
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	sent := map[string]int{"alice": 3}
      	sent["bob"]++
      	sent["alice"]++
      	fmt.Println(sent["alice"], sent["bob"], len(sent))
      }
      ```
    options:
      - text: '`4 1 2`'
        correct: true
      - text: '`4 0 1`'
      - text: It panics, because `"bob"` isn't in the map
      - text: '`3 1 2`'
    explanation: |
      Looking up a missing key gives the zero value, so `sent["bob"]++` starts
      from `0` and stores `1`. `alice` goes from 3 to 4. There are now two
      keys, so `len` is 2.
---

Slices are great when you want items in order and look them up by position. But what if you want to look up a customer's credit balance by their **username**? Searching a slice every time would be slow and clumsy. You want a **map**.

## What's a map?

A map stores **key/value pairs**. You look up a value by its key, like finding a word in a dictionary. Other languages call them dictionaries, hashes or hash maps.

The type is written `map[KeyType]ValueType`:

```go
package main

import "fmt"

func main() {
	credits := map[string]int{
		"alice": 50,
		"bob":   0,
	}

	fmt.Println(credits["alice"])

	credits["carol"] = 25 // add a new key
	credits["alice"] = 45 // update an existing key

	fmt.Println(credits)
	fmt.Println(len(credits))
}
```

```text
50
map[alice:45 bob:0 carol:25]
3
```

- A **map literal** lists `key: value` pairs in braces. Note the trailing comma after the last pair when it's on its own line: Go requires it.
- `m[key]` reads a value; `m[key] = value` adds or updates one.
- `len` returns the number of key/value pairs.
- Each key appears **at most once**. Setting an existing key replaces its value.

When you print a map, `fmt` sorts the keys so the output is easy to read. (That's a printing nicety. As you'll see soon, the map itself has no order.)

## Creating an empty map with `make`

To start with an empty map, use `make`:

```go
phoneNumbers := make(map[string]string)
phoneNumbers["alice"] = "+1-555-0100"
```

An empty literal, `map[string]string{}`, does the same thing.

## Missing keys return the zero value

Looking up a key that isn't in the map doesn't crash. It returns the zero value of the value type:

```go
package main

import "fmt"

func main() {
	credits := map[string]int{"alice": 50}
	fmt.Println(credits["nobody"])
}
```

```text
0
```

That's often convenient. Counting things is a breeze:

```go
package main

import "fmt"

func main() {
	messages := []string{"alice", "bob", "alice", "alice"}
	counts := make(map[string]int)
	for _, sender := range messages {
		counts[sender]++
	}
	fmt.Println(counts)
}
```

```text
map[alice:3 bob:1]
```

But it can also be ambiguous: does Bob have `0` credits, or is Bob not a customer at all? The next lesson shows how to tell the difference.

## The nil map trap

The zero value of a map type is `nil`. If you declare a map with `var` and never create it, you have a **nil map**:

```go
var credits map[string]int // nil!
fmt.Println(credits["alice"]) // fine: prints 0
credits["alice"] = 10         // panic: assignment to entry in nil map
```

Reading from a nil map works (you get zero values), but **writing to one panics**. It's one of the most common beginner crashes in Go. The fix is always to create the map first, with `make` or a literal.

## What can be a key?

Keys must be a type that can be compared with `==`: strings, numbers, bools, arrays and structs made of those. **Slices, maps and functions can't be keys**, because they can't be compared with `==`. Values, on the other hand, can be any type at all, including slices and other maps.

## Further reading

- [Go by Example: Maps](https://gobyexample.com/maps)
- [A Tour of Go: Maps](https://go.dev/tour/moretypes/19)
