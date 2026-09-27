---
title: Arrays in Memory
quiz:
  - question: |
      What does this print?

      ```go
      a := [3]int{1, 2, 3}
      b := a
      b[0] = 99
      fmt.Println(a[0], b[0])
      ```
    options:
      - text: '`99 99`'
      - text: '`1 99`'
        correct: true
      - text: '`1 1`'
    explanation: |
      Arrays are values in Go. `b := a` copies all three elements, so changing
      `b` leaves `a` alone. (Slices behave differently, as the next lesson shows.)
  - question: Why is indexing into an array O(1)?
    options:
      - text: Go caches recently used indexes
      - text: The address of element `i` is computed directly as start + i × element size
        correct: true
      - text: Arrays are stored as hash maps
    explanation: |
      Elements sit next to each other in memory, each the same size. One
      multiplication and one addition give the address of any element, no matter
      how big the array is.
---

To understand why some operations are fast and others slow, you need a mental
picture of **memory**. Go's arrays are the simplest picture there is.

## Memory is a long row of boxes

Think of your computer's memory as one enormous row of numbered byte-sized
boxes. Each number is an **address**. An `int` on a 64-bit machine takes 8
boxes (8 bytes).

An **array** stores its elements **contiguously**: side by side, with no gaps,
all the same size. A `[4]int` of weekly follower gains is 32 bytes in a row:

```
address:  1000     1008     1016     1024
         +--------+--------+--------+--------+
         |  120   |  340   |  95    |  410   |
         +--------+--------+--------+--------+
index:       0        1        2        3
```

To find element `i`, the CPU computes `start + i × 8`. One multiply, one add,
done. That's why **array indexing is O(1)**: element 3 of a four-item array and
element 3,000,000 of a huge one take the same time to reach.

We can check the spacing in Go:

```go
package main

import (
	"fmt"
	"unsafe"
)

func main() {
	gains := [4]int{120, 340, 95, 410}
	fmt.Println(unsafe.Sizeof(gains))
	fmt.Println(uintptr(unsafe.Pointer(&gains[1])) - uintptr(unsafe.Pointer(&gains[0])))
}
```

```
32
8
```

The whole array is 32 bytes and neighbours are exactly 8 bytes apart. (The
`unsafe` package is for peeking under the hood like this. Don't use it in
Clout's production code!)

Contiguous memory has another huge benefit: **CPU caches**. When the CPU
fetches one element, it grabs a whole chunk of neighbours (a *cache line*,
typically 64 bytes) at the same time. Looping through an array uses every byte
it fetched, which is why iterating over arrays and slices is so fast in
practice, even compared with other O(n) structures.

## Go arrays are fixed-size values

In Go, an array's length is **part of its type**. `[4]int` and `[5]int` are
different types, and the length must be a constant known at compile time:

```go
var weekly [7]int         // seven zeros
monthly := [...]int{1, 2} // length inferred: [2]int
```

Arrays are also **values**. Assigning or passing one copies every element:

```go
func reset(a [7]int) { a[0] = 0 } // changes a copy only
```

For a `[7]int` the copy is cheap. For a `[1_000_000]int` it's 8 MB every call.

## The trouble with arrays

The fixed size is the real limitation. Clout doesn't know how many
influencers will sign up tomorrow. To "grow" an array you'd have to:

1. Allocate a new, bigger array somewhere else in memory.
2. Copy every element across: O(n).
3. Start using the new one.

And inserting at the front means shifting every element one place right,
which is also O(n).

That's why you rarely use arrays directly in Go. Instead you use **slices**,
which are built on top of arrays and automate the "allocate a bigger one and
copy" dance. Let's see how.
