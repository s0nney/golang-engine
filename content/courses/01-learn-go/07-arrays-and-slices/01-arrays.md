---
title: Arrays
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	retries := [3]int{5, 10}
      	fmt.Println(len(retries), retries[2])
      }
      ```
    options:
      - text: '`2 10`'
      - text: '`3 10`'
      - text: '`3 0`'
        correct: true
      - text: It doesn't compile, because only two values were given
    explanation: |
      The length is part of the array's type, `[3]int`, so `len` is 3. Only
      the first two elements were given values; the third gets the zero value
      for `int`, which is `0`.
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	a := [2]string{"hi", "yo"}
      	b := a
      	b[0] = "hey"
      	fmt.Println(a[0])
      }
      ```
    options:
      - text: '`hi`'
        correct: true
      - text: '`hey`'
      - text: '`yo`'
    explanation: |
      Assigning an array copies every element. `b` is an independent copy of
      `a`, so changing `b[0]` leaves `a` untouched.
---

So far, each variable has held a single value. But Textio deals in lists: lists of phone numbers, lists of messages, lists of delivery attempts. Go has two ways to store an ordered list of values: **arrays** and **slices**. Arrays come first because slices are built on top of them.

## Declaring an array

An array is a **fixed-size** sequence of values, all of the same type. The size goes in square brackets before the type:

```go
package main

import "fmt"

func main() {
	var backupCarriers [3]string
	backupCarriers[0] = "Alpha Mobile"
	backupCarriers[1] = "Beta Tel"
	backupCarriers[2] = "Gamma Net"

	fmt.Println(backupCarriers)
	fmt.Println(backupCarriers[1])
	fmt.Println(len(backupCarriers))
}
```

```text
[Alpha Mobile Beta Tel Gamma Net]
Beta Tel
3
```

- Each value in the array is an **element**.
- You get or set an element with its **index** in square brackets. Indexes start at **0**, so a 3-element array has indexes 0, 1 and 2.
- `len` returns the number of elements.

## Array literals

You can create an array and fill it in one step with an **array literal**:

```go
primes := [5]int{2, 3, 5, 7, 11}
```

Let the compiler count the elements for you with `...`:

```go
codes := [...]string{"US", "GB", "FR"} // type is [3]string
```

Any elements you don't give a value get the type's zero value:

```go
var attempts [4]int // [0 0 0 0]
```

## Out of range

Using an index that doesn't exist is an error. If Go can spot it while compiling, it refuses to compile:

```go
codes := [3]string{"US", "GB", "FR"}
fmt.Println(codes[3]) // error: invalid argument: index 3 out of bounds [0:3]
```

If the index is only known while the program runs, the program **panics** (crashes) with `index out of range` instead. The last valid index is always `len - 1`.

## The size is part of the type

This is the big thing about arrays: `[3]string` and `[4]string` are **different types**. The size is fixed when you write the code and can never change. You can't add a fourth carrier to `backupCarriers`.

That makes arrays awkward for most real work. How many messages will a customer send? You don't know in advance.

## Arrays are values

When you assign an array or pass it to a function, Go copies **every element**:

```go
package main

import "fmt"

func main() {
	original := [3]int{1, 2, 3}
	copied := original
	copied[0] = 100

	fmt.Println(original, copied)
}
```

```text
[1 2 3] [100 2 3]
```

That's safe and predictable, but copying a big array is expensive.

## So when do you use arrays?

Honestly, rarely. Arrays make sense when the size is truly fixed by nature: the 3 colour channels of a pixel, the 16 bytes of a UUID, the 7 days of the week. For everything else, Go programmers use **slices**, which you'll meet next. But slices store their data in arrays behind the scenes, so understanding arrays will help you understand how slices behave.

## Further reading

- [Go by Example: Arrays](https://gobyexample.com/arrays)
- [A Tour of Go: Arrays](https://go.dev/tour/moretypes/6)
