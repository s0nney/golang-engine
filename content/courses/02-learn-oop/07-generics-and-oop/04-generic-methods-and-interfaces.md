---
title: Generic Methods and Interfaces
quiz:
  - question: |
      Does this compile?

      ```go
      type Looter interface {
      	Loot[T any]() T
      }
      ```
    options:
      - text: Yes, since Go 1.27
      - text: No; interface methods can't have type parameters
        correct: true
      - text: Only if `T` is constrained by `comparable`
      - text: Yes, but only inside package `main`
    explanation: |
      Go 1.27 allows type parameters on concrete methods only. The compiler
      rejects this with "interface method must have no type parameters".
  - question: |
      What happens here?

      ```go
      type Doubler interface{ Double(n int) int }

      type Rune struct{}

      func (Rune) Double[N ~int](n N) N { return n * 2 }

      var d Doubler = Rune{}
      ```
    options:
      - text: It compiles, because `Double[int]` matches `Double(int) int`
      - text: It fails to compile; a generic method never satisfies an interface method
        correct: true
      - text: It compiles but panics at runtime
      - text: It compiles only with an explicit conversion `Doubler(Rune{})`
    explanation: |
      Even though one instantiation would match, a generic method isn't in the
      type's method set for interface purposes. The compiler says `Rune` does not
      implement `Doubler` (wrong type for method Double).
---

Go 1.27's generic methods come with two firm rules, and both are about interfaces. They matter a lot for OOP-style design, because interfaces are how Go does polymorphism.

## Rule 1: interface methods can't have type parameters

You might want to write:

```go
type Container interface {
	Map[U any](f func(Item) U) []U // not allowed
}
```

The compiler rejects it:

```
interface method must have no type parameters
```

The reason is dynamic dispatch. When you call a method through an interface, Go looks up the concrete code at runtime. A generic method has potentially infinitely many versions (`Map[int]`, `Map[string]`, `Map[Hero]`...), and Go compiles generic code ahead of time. There's no reasonable way to guarantee the right version exists for whatever type happens to be inside the interface.

## Rule 2: a generic method can't satisfy an interface method

Even when one instantiation would fit perfectly, a generic method doesn't count:

```go
type Looter interface {
	Loot(n int) int
}

type Chest struct{}

func (Chest) Loot[N ~int](n N) N { return n * 2 }

var l Looter = Chest{} // compile error
```

```
Chest does not implement Looter (wrong type for method Loot)
		have Loot[N ~int](N) N
		want Loot(int) int
```

A generic method is effectively invisible to interfaces. It exists for callers who hold the **concrete type**.

## Designing around the rules

In practice this leads to a clean division of labour:

**Interfaces describe the fixed behaviour.** Keep them small and non-generic, as always:

```go
type Item interface{ Name() string }
```

**Generic methods add convenience to concrete types.** `Inventory.First[T]()` from the last lesson lives on the concrete `*Inventory`. Code that has an `*Inventory` gets the nice typed lookups; nothing about the `Item` interface had to change.

**If you need the behaviour behind an interface, make the interface's type parameter part of the interface itself** instead of the method:

```go
package main

import "fmt"

// The interface is generic; its method is not.
type Source[T any] interface {
	Next() (T, bool)
}

type LootTable struct {
	drops []string
	i     int
}

func (lt *LootTable) Next() (string, bool) {
	if lt.i >= len(lt.drops) {
		return "", false
	}
	d := lt.drops[lt.i]
	lt.i++
	return d, true
}

func drain[T any](s Source[T]) []T {
	var out []T
	for v, ok := s.Next(); ok; v, ok = s.Next() {
		out = append(out, v)
	}
	return out
}

func main() {
	table := &LootTable{drops: []string{"gold", "gem", "bone"}}
	fmt.Println(drain[string](table))
}
```

```
[gold gem bone]
```

`Source[string]` has an ordinary, non-generic method `Next() (string, bool)`, so `*LootTable` satisfies it in the usual implicit way. The type parameter lives on the *interface*, which is fine, because each instantiation of the interface is a normal interface.

## Choosing between them

| You want...                                              | Reach for...                  |
|----------------------------------------------------------|-------------------------------|
| different types behaving differently, chosen at runtime   | an interface                  |
| one algorithm or container that works for any type        | a generic function or type    |
| a type-safe helper on a concrete type, like `inv.First[T]` | a generic method (Go 1.27+) |
| polymorphism *and* a type parameter                       | a generic interface, like `Source[T]` |

And, as always, start with the simplest thing. Most Go code needs none of these, just concrete types and functions.
