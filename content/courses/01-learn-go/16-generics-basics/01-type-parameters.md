---
title: Type Parameters
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func last[T any](items []T) T {
      	return items[len(items)-1]
      }

      func main() {
      	fmt.Println(last([]string{"hi", "yo"}), last([]int{4, 8, 15}))
      }
      ```
    options:
      - text: '`yo 15`'
        correct: true
      - text: '`hi 4`'
      - text: It doesn't compile, because `T` is used with two different types
      - text: '`yo 8`'
    explanation: |
      Each call picks its own type for `T`. The first call infers
      `T = string` and returns `"yo"`; the second infers `T = int` and
      returns `15`.
  - question: 'In `func first[T any](items []T) T`, what is `any`?'
    options:
      - text: The constraint on `T`, meaning `T` can be any type at all
        correct: true
      - text: The name of the type parameter
      - text: The return type
      - text: A keyword that disables type checking
    explanation: |
      Type parameters are declared as `[Name Constraint]`. `T` is the name
      and `any` is the constraint. `any` allows every type, and the compiler
      still checks each call fully.
---

Here's a function that returns the first recipient in a list:

```go
func firstString(items []string) string {
	return items[0]
}
```

Now you need the same thing for a list of message IDs (`[]int`). And for a list of costs (`[]float64`). Copy and paste three times? The logic is identical; only the **type** changes. **Generics** let you write it once.

## Type parameters

A generic function declares one or more **type parameters** in square brackets, between the name and the regular parameters:

```go
package main

import "fmt"

func first[T any](items []T) T {
	return items[0]
}

func main() {
	recipients := []string{"alice", "bob"}
	ids := []int{1042, 1043}
	costs := []float64{0.01, 0.02}

	fmt.Println(first(recipients))
	fmt.Println(first(ids))
	fmt.Println(first(costs))
}
```

```text
alice
1042
0.01
```

- `[T any]` declares a type parameter named `T`. You can use `T` anywhere in the signature and body where you'd write a type.
- `any` is the **constraint**: it says which types `T` is allowed to be. `any` means every type.
- When you call `first(ids)`, Go works out that `T` is `int` from the argument. That's called **type inference**.

It's still fully type-checked. `first(ids)` returns an `int`, not some unknown "anything", so you can't accidentally use it as a string.

You can also pass the type explicitly, which is occasionally needed when Go can't infer it:

```go
id := first[int](ids)
```

You already did this with `errors.AsType[*carrierError](err)`, where there's no argument to infer the type from.

## Several type parameters

Functions can have more than one type parameter:

```go
package main

import "fmt"

func mapSlice[T, U any](items []T, f func(T) U) []U {
	result := make([]U, 0, len(items))
	for _, item := range items {
		result = append(result, f(item))
	}
	return result
}

func main() {
	bodies := []string{"hi", "Your code is 4821"}
	lengths := mapSlice(bodies, func(s string) int { return len(s) })
	fmt.Println(lengths)
}
```

```text
[2 17]
```

`mapSlice` turns a `[]T` into a `[]U` by calling `f` on each element. Here `T` is `string` and `U` is `int`. The parameter `f func(T) U` is a function type: "a function that takes a `T` and returns a `U`".

## Generic types

Types can have type parameters too. Here's a simple queue that works for any element type:

```go
package main

import "fmt"

type Queue[T any] struct {
	items []T
}

func (q *Queue[T]) Push(item T) {
	q.items = append(q.items, item)
}

func (q *Queue[T]) Len() int {
	return len(q.items)
}

func main() {
	var outbox Queue[string]
	outbox.Push("reminder for alice")
	outbox.Push("reminder for bob")
	fmt.Println(outbox.Len())
}
```

```text
2
```

When you use a generic type, you supply the type argument: `Queue[string]`. Methods refer to the type as `Queue[T]`.

## What can you do with a `T`?

With the `any` constraint, Go only lets you do things that work for **every** type: assign it, pass it around, store it in slices and maps. You can't add two `T`s with `+` or compare them with `<`, because not every type supports that. To do more, you need a tighter constraint, which is the next lesson.

## Further reading

- [Go by Example: Generics](https://gobyexample.com/generics)
- [A Tour of Go: Type parameters](https://go.dev/tour/generics/1)
- [Learn Go with Tests: Generics](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/generics)
