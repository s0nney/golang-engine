---
title: Slices, make and append
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	queue := make([]string, 2, 10)
      	queue = append(queue, "hello")
      	fmt.Println(len(queue), cap(queue))
      }
      ```
    options:
      - text: '`1 10`'
      - text: '`3 10`'
        correct: true
      - text: '`3 3`'
      - text: '`2 10`'
    explanation: |
      `make([]string, 2, 10)` creates a slice that already has 2 elements
      (both empty strings) with room for 10. `append` adds a third element
      after them, so the length is 3. The capacity is still 10 because there
      was room.
  - question: What's wrong with the line `append(messages, "hi")` on its own?
    options:
      - text: Nothing, `messages` now contains `"hi"`
      - text: It doesn't compile, because the result of `append` must be used
        correct: true
      - text: It panics if `messages` is `nil`
    explanation: |
      `append` returns the updated slice. You must store the result, almost
      always back into the same variable: `messages = append(messages, "hi")`.
      Go refuses to compile a call to `append` whose result is thrown away.
---

A **slice** is Go's everyday list. Like an array, it holds an ordered sequence of values of one type. Unlike an array, its length can grow.

## Slice literals

A slice type is written like an array type with **no size**: `[]string` instead of `[3]string`.

```go
package main

import "fmt"

func main() {
	recipients := []string{"Alice", "Bob", "Carol"}
	fmt.Println(recipients)
	fmt.Println(recipients[0], len(recipients))

	recipients[1] = "Bobby"
	fmt.Println(recipients)
}
```

```text
[Alice Bob Carol]
Alice 3
[Alice Bobby Carol]
```

Indexing works just as with arrays, and going out of range panics.

## Growing with `append`

The built-in `append` adds elements to the end of a slice and **returns the updated slice**:

```go
package main

import "fmt"

func main() {
	var outbox []string // nil slice: no elements yet
	fmt.Println(len(outbox), outbox == nil)

	outbox = append(outbox, "Your code is 4821")
	outbox = append(outbox, "Your order shipped", "Your order arrived")
	fmt.Println(len(outbox))
	fmt.Println(outbox)
}
```

```text
0 true
3
[Your code is 4821 Your order shipped Your order arrived]
```

A few things to notice:

- The zero value of a slice is **`nil`**, meaning "no slice at all". A nil slice has length 0, and you can `append` to it, `len` it and loop over it without any special care.
- `append` is variadic, so you can add several elements at once.
- **Always assign the result back**: `outbox = append(outbox, ...)`. You'll see why next lesson.

To append one slice to another, spread it with `...`:

```go
more := []string{"Reminder: tomorrow at 9"}
outbox = append(outbox, more...)
```

## Length and capacity

A slice is a small "window" onto an array that Go manages for you. The slice tracks two numbers:

- **Length** (`len`): how many elements are in the slice right now.
- **Capacity** (`cap`): how many elements fit in the underlying array before Go must allocate a bigger one.

When you `append` to a full slice, Go allocates a new, larger array (roughly double the size for small slices), copies the elements across and returns a slice pointing at the new array:

```go
package main

import "fmt"

func main() {
	var ids []int
	for i := range 5 {
		ids = append(ids, i)
		fmt.Println("len", len(ids), "cap", cap(ids))
	}
}
```

```text
len 1 cap 4
len 2 cap 4
len 3 cap 4
len 4 cap 4
len 5 cap 8
```

Go 1.27 starts this small slice with room for 4, then doubles it to 8 when it fills up. The exact growth pattern is an implementation detail that changes between Go versions, so never write code that depends on it.

## `make`

If you know roughly how many elements you'll need, create the slice with `make` and save Go from reallocating over and over:

```go
lengths := make([]int, 3)       // len 3, cap 3: [0 0 0]
batch := make([]string, 0, 100) // len 0, cap 100: empty, with room for 100
```

`make([]T, length, capacity)` creates a slice whose elements are all zero values. The capacity is optional and defaults to the length.

A very common mistake is mixing these up:

```go
package main

import "fmt"

func main() {
	names := make([]string, 3)
	names = append(names, "Alice")
	fmt.Printf("%q\n", names)
}
```

```text
["" "" "" "Alice"]
```

`make([]string, 3)` already has **three** empty strings in it, and `append` adds after them. If you plan to `append`, use `make([]string, 0, 3)` instead.

## Further reading

- [Go by Example: Slices](https://gobyexample.com/slices)
- [Learn Go with Tests: Arrays and slices](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/arrays-and-slices)
