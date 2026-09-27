---
title: Inside an Interface Value
quiz:
  - question: |
      What does this print?

      ```go
      e := Entry{Key: "logo.png", Size: 2048}
      var v any = e
      e.Size = 0
      fmt.Println(v.(Entry).Size)
      ```
    options:
      - text: '`0`'
      - text: '`2048`'
        correct: true
      - text: It panics
      - text: It doesn't compile
    explanation: |
      Assigning `e` to an interface stores a *copy* of the struct. Changing `e`
      afterwards doesn't affect the copy inside `v`. Had you stored `&e`, the interface
      would hold a copy of the pointer, and you'd see the change.
  - question: |
      What does `fmt.Println(a == b)` print for `var a, b any = 1, int64(1)`?
    options:
      - text: '`true`, because both hold the number 1'
      - text: '`false`, because the dynamic types differ (`int` vs `int64`)'
        correct: true
      - text: It doesn't compile
      - text: It panics
    explanation: |
      Two interface values are equal only if their dynamic types are identical *and*
      their dynamic values are equal. An `int` 1 and an `int64` 1 have different types, so
      the values are never compared at all.
---

In [Learn OOP](/courses/learn-oop/polymorphism/dynamic-dispatch) you saw that an interface value is a pair: a dynamic type and a dynamic value. This chapter goes one level down, because that pair explains copying, comparison, allocation, nil gotchas and performance.

## Two words

On a 64-bit machine, every interface value is **two machine words** (16 bytes):

1. A **type word**. For an empty interface (`any`), it points at the runtime's description of the dynamic type. For an interface with methods, like `io.Writer`, it points at an **itab**: a small table holding the dynamic type *and* pointers to the methods that satisfy this interface. The runtime builds each itab once per (interface, type) pair and caches it.
2. A **data word**: a pointer to the value, or, for pointer-shaped types, the pointer itself.

A method call through an interface loads the function pointer from the itab and calls it with the data word as the receiver. That's dynamic dispatch: one extra indirection, and the compiler usually can't inline the call.

## Storing a value means copying it

```go
package main

import (
	"fmt"
	"testing"
	"unsafe"
)

type Entry struct {
	Key  string
	Size int
}

var sink any

func main() {
	e := Entry{"logo.png", 2048}
	var v any = e // the interface gets its own copy of e
	e.Size = 0
	fmt.Println(v.(Entry).Size)

	var w any = &e // the interface holds a copy of the pointer
	e.Size = 99
	fmt.Println(w.(*Entry).Size)

	fmt.Println(unsafe.Sizeof(v), unsafe.Sizeof(w))

	var a, b any = 1, int64(1)
	fmt.Println(a == b, a == 1)

	// Storing a struct in an interface copies it to the heap; a pointer doesn't need to.
	fmt.Println(testing.AllocsPerRun(100, func() { sink = e }))
	fmt.Println(testing.AllocsPerRun(100, func() { sink = &e }))
}
```

```
2048
99
16 16
false true
1
0
```

Line by line:

- **`var v any = e`** copies the `Entry`. Later changes to `e` don't reach the copy, which is why values stored in interfaces aren't addressable (next lesson).
- **`var w any = &e`** copies only the *pointer*, so `w` sees later changes.
- **Both are 16 bytes**, whatever they hold: two words.
- **`a == b` is false**: equal interface values need identical dynamic types first. `a == 1` is true because the constant `1` is converted to an `any` holding an `int`.
- **Allocation**: a struct doesn't fit in the data word, so storing one in an interface usually copies it to the heap (1 allocation). A pointer *is* a word, so it goes straight in (0 allocations).

The compiler avoids some of those allocations: small integers, values known at compile time, and values that provably don't outlive the function often skip the heap. But as a rule of thumb, **converting a non-pointer value to an interface may allocate**. That's the hidden cost behind `fmt.Println(x)`, `[]any`, and `container/heap`'s `Push(x any)`.

## Comparison rules

Interface comparison follows from the two words:

1. Compare the dynamic types. Different types: not equal. Done.
2. Same type: compare the dynamic values with that type's `==`.
3. If that type isn't comparable (a slice, map or function), **panic**.

A nil interface has neither a type nor a value, so it equals only other nil interfaces. That's the root of the famous gotcha you met in Learn OOP, which we'll revisit in lesson 5.

## Why generics can be different

With an interface parameter, the function receives two words and makes indirect calls. With a type parameter, the compiler knows more about `T` and can often work with values directly, with no boxing at all. "Often" is carrying some weight there: chapter 8 explains exactly how Go compiles generic code, and when it does and doesn't beat interfaces.

## Further reading

- [Go Data Structures: Interfaces](https://research.swtch.com/interfaces) by Russ Cox (from 2009, but the model still holds)
