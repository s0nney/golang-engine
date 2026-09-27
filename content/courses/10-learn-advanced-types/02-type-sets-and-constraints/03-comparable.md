---
title: comparable and the Go 1.20 Change
quiz:
  - question: |
      What happens here?

      ```go
      func Count[T comparable](xs []T, v T) int {
      	n := 0
      	for _, x := range xs {
      		if x == v {
      			n++
      		}
      	}
      	return n
      }

      func main() {
      	fmt.Println(Count([]any{1, "a", 1}, 1))
      	fmt.Println(Count([]any{[]int{1}}, any([]int{1})))
      }
      ```
    options:
      - text: It doesn't compile, because `any` doesn't satisfy `comparable`
      - text: It prints `2`, then `1`
      - text: It prints `2`, then panics comparing two `[]int` values
        correct: true
      - text: It prints `2`, then `0`
    explanation: |
      Since Go 1.20, `any` satisfies `comparable`, because interface values support `==`.
      The first call is fine. In the second, both interfaces hold slices, and comparing
      them panics with `runtime error: comparing uncomparable type []int`.
  - question: Which of these types does **not** satisfy `comparable`?
    options:
      - text: '`struct{ ID int; Name string }`'
      - text: '`[3]string`'
      - text: '`*[]int`'
      - text: '`struct{ Tags []string }`'
        correct: true
    explanation: |
      A struct is comparable only if all its fields are, and slices aren't. Arrays of
      comparable elements are fine, and *every* pointer is comparable (it compares
      addresses), including a pointer to a slice.
---

Stash's `Set[T]` will be a map underneath, and map keys must support `==`. That's what the predeclared constraint **`comparable`** is for. It has a subtle history.

## Which types are comparable?

The spec calls a type *comparable* if `==` and `!=` work on it:

- booleans, numbers, strings, pointers, channels: yes;
- arrays: if their element type is comparable;
- structs: if every field is comparable;
- **interfaces: yes**, by comparing dynamic type and value;
- slices, maps and functions: **no** (they can only be compared to `nil`).

Interfaces are the tricky one. Two interface values are equal if they hold identical dynamic types and equal values. But if both hold, say, a `[]int`, there's no way to compare them, and `==` **panics at runtime**.

So there are really two levels: types that are *strictly comparable* (`==` can never panic), and types that are comparable but might panic (interfaces, and arrays or structs containing them).

## Go 1.18 and 1.19: strict

At first, `comparable` meant *strictly* comparable. That was safe, but annoying: `any` didn't satisfy it, so you couldn't make a `Set[any]` or call `Count` on a `[]any`, even though a plain `map[any]bool` has worked forever.

## Go 1.20: interfaces allowed

Go 1.20 relaxed the rule. Now **any comparable type satisfies `comparable`**, including interfaces and structs containing them. The price is the same as for maps: `==` can panic at runtime if two interfaces hold the same uncomparable dynamic type.

```go
package main

import "fmt"

func Count[T comparable](xs []T, v T) int {
	n := 0
	for _, x := range xs {
		if x == v {
			n++
		}
	}
	return n
}

type Tagged struct {
	Name string
	Meta any // an interface field: comparable, but not strictly
}

func main() {
	fmt.Println(Count([]any{1, "a", 1}, 1))
	fmt.Println(Count([]Tagged{{"a", 1}, {"a", 2}}, Tagged{"a", 1}))

	defer func() { fmt.Println("panic:", recover()) }()
	fmt.Println(Count([]any{[]int{1}}, any([]int{1})))
}
```

```
2
1
panic: runtime error: comparing uncomparable type []int
```

Before Go 1.20, both `Count([]any...)` and `Count([]Tagged...)` would have been compile errors.

## What still doesn't compile

Types that are *never* comparable are still rejected, at compile time:

```go
Count([][]int{{1}}, []int{1}) // []int does not satisfy comparable
```

And `comparable` is a general interface, so it can only be a constraint: `var c comparable` fails with `cannot use type comparable outside a type constraint`.

## `comparable` in your own constraints

You can embed it to combine "usable as a map key" with other requirements:

```go
type Key interface {
	comparable
	String() string
}
```

But you can't put it in a union (`comparable | ~[]byte` is an error).

## The practical rules

1. Use `comparable` for anything that ends up as a **map key** or is compared with `==`. Stash's `Set[T comparable]` and `OrderedMap[K comparable, V any]` do.
2. If your element type is (or contains) an **interface**, remember `==` can panic, just like a map with `any` keys.
3. If you need "equal" to mean something custom (case-insensitive keys, comparing slices), don't fight `comparable`. Take an `eq func(a, b T) bool` or a key function instead, like `slices.EqualFunc` does.

## Further reading

- [All your comparable types](https://go.dev/blog/comparable), the Go blog on the 1.20 change
