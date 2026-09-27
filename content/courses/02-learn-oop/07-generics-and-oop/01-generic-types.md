---
title: Generic Types
quiz:
  - question: |
      Given `type Stack[T any] struct{ items []T }`, which declaration is correct for its `Push` method?
    options:
      - text: '`func (s *Stack) Push(v T)`'
      - text: '`func (s *Stack[T]) Push(v T)`'
        correct: true
      - text: '`func [T any](s *Stack[T]) Push(v T)`'
      - text: '`func (s *Stack[any]) Push(v any)`'
    explanation: |
      The receiver must mention the type parameter, `Stack[T]`. The `T` there
      introduces a name for the type parameter that the method's body and
      signature can use. You don't repeat the constraint.
  - question: |
      What does this print?

      ```go
      func (s *Stack[T]) Pop() (T, bool) {
      	var zero T
      	if len(s.items) == 0 {
      		return zero, false
      	}
      	v := s.items[len(s.items)-1]
      	s.items = s.items[:len(s.items)-1]
      	return v, true
      }

      func main() {
      	var s Stack[string]
      	v, ok := s.Pop()
      	fmt.Printf("%q %v\n", v, ok)
      }
      ```
    options:
      - text: '`nil false`'
      - text: '`"" false`'
        correct: true
      - text: It panics on the empty slice
      - text: '`"" true`'
    explanation: |
      For `Stack[string]`, `T` is `string`, so `var zero T` is the empty string.
      `%q` prints it as `""`. The length check stops the out-of-range slice access.
---

In the Go basics course you wrote generic *functions*, like a `Map` that works on any slice. Types can be generic too, and that matters for OOP, because it lets you write a reusable **container object** once and use it for heroes, items and dragons alike.

## The problem

Our game needs several queues: heroes waiting for their turn, spells waiting to resolve, loot waiting to be picked up. Without generics you'd write `HeroQueue`, `SpellQueue` and `LootQueue`, identical apart from the element type. Or you'd use `[]any` and lose type safety.

## A generic type

Put a **type parameter** list after the type name:

```go
type Queue[T any] struct {
	items []T
}
```

`Queue` on its own isn't a type yet. `Queue[*Hero]` and `Queue[Spell]` are. Supplying the type argument is called **instantiation**.

Methods are declared on the generic type by naming the parameter in the receiver:

```go
package main

import "fmt"

type Queue[T any] struct {
	items []T
}

func (q *Queue[T]) Push(v T) {
	q.items = append(q.items, v)
}

func (q *Queue[T]) Pop() (T, bool) {
	var zero T
	if len(q.items) == 0 {
		return zero, false
	}
	v := q.items[0]
	q.items = q.items[1:]
	return v, true
}

func (q *Queue[T]) Len() int { return len(q.items) }

type Hero struct{ Name string }

type Spell struct {
	Name string
	Cost int
}

func main() {
	var turns Queue[Hero]
	turns.Push(Hero{Name: "Aria"})
	turns.Push(Hero{Name: "Borin"})

	var spells Queue[Spell]
	spells.Push(Spell{Name: "fireball", Cost: 30})

	for turns.Len() > 0 {
		h, _ := turns.Pop()
		fmt.Println(h.Name, "takes a turn")
	}

	s, ok := spells.Pop()
	fmt.Println(s.Name, s.Cost, ok)

	_, ok = spells.Pop()
	fmt.Println(ok)
}
```

```
Aria takes a turn
Borin takes a turn
fireball 30 true
false
```

Some things to notice:

- **`var zero T`** is how you get a zero value of an unknown type. For `Hero` it's `Hero{}`, for `int` it's `0`, for a pointer it's `nil`.
- **The zero value is useful.** `var turns Queue[Hero]` needs no constructor, thanks to `append` handling nil slices, just like in chapter 2.
- **Encapsulation still works.** `items` is unexported, so outside the package the only way in or out is `Push` and `Pop`.
- **Type safety.** `turns.Push(Spell{})` is a compile error. With `[]any` it would have compiled and blown up later.

## Generic types vs interfaces

Both give you "one piece of code, many types", but in different ways:

- An **interface** says "I don't care what type this is, as long as it has these *methods*". Values of different types can sit in the same `[]Combatant`.
- A **generic type** says "this works for any type `T`, but once chosen, it's *one* type". A `Queue[Hero]` holds only heroes, and `Pop` returns a `Hero`, not an interface you need to assert.

Containers (queues, stacks, sets, caches, inventories) are the classic place for generic types. Behaviour that varies by type (attacking, healing, rendering) is the classic place for interfaces.

## Multiple type parameters

Types can take several parameters:

```go
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

loot := Pair[string, int]{Key: "gold", Value: 250}
```

The `comparable` constraint on `K` says "must support `==`". Constraints are the subject of the next lesson.

## Further reading

- [Go by Example: Generics](https://gobyexample.com/generics)
- [An Introduction To Generics](https://go.dev/blog/intro-generics)
