---
title: Generic Types With Methods
quiz:
  - question: |
      What does this print?

      ```go
      type Set[T comparable] struct{ m map[T]struct{} }

      func (s Set[T]) Add(v T) {
      	if s.m == nil {
      		s.m = make(map[T]struct{})
      	}
      	s.m[v] = struct{}{}
      }

      func (s Set[T]) Len() int { return len(s.m) }

      func main() {
      	var s Set[string]
      	s.Add("go")
      	s.Add("types")
      	fmt.Println(s.Len())
      }
      ```
    options:
      - text: '`2`'
      - text: '`1`'
      - text: '`0`'
        correct: true
      - text: It panics writing to a nil map
    explanation: |
      `Add` has a value receiver, so it creates the map in a *copy* of `s`, and the copy is
      thrown away when `Add` returns. The original's `m` stays nil, and `len` of a nil map
      is 0. Lazily-initialised types need pointer receivers.
  - question: |
      What's wrong with `func (s Set[int]) Sum() int`?
    options:
      - text: Nothing; it's a method only for `Set[int]`
      - text: 'Methods can''t return `int`'
      - text: 'It declares a receiver type parameter *named* `int`, which shadows the real `int`; Go has no methods for just one instantiation'
        correct: true
      - text: Receivers must be pointers on generic types
    explanation: |
      The names in a receiver's brackets always *declare* type parameters. `int` here is a
      new type parameter constrained by `comparable`, and `total += k` then fails with a
      confusing "mismatched types int and int" error. Write a function
      `func SumSet(s Set[int]) int` instead.
---

You met generic types in [Learn OOP](/courses/learn-oop/generics-and-oop/generic-types). Stash is going to be full of them, so this lesson collects the rules and gotchas of giving them methods.

## Receiver type parameters

A method on a generic type must list the type's parameters in the receiver, without constraints:

```go
type Counter[K comparable] struct{ m map[K]int }

func (c *Counter[K]) Inc(k K) { c.m[k]++ }

// Any name works for the receiver's type parameter...
func (c *Counter[Key]) Get(k Key) int { return c.m[k] }

// ...and _ is fine if the method doesn't use it.
func (c *Counter[_]) Len() int { return len(c.m) }
```

The names are **new declarations**, local to the method. They don't have to match the type declaration, and they inherit its constraints automatically. `_` is fine when the method body doesn't mention the parameter.

## No methods for one instantiation

Because the receiver brackets *declare* names, you can't write a method that exists only for `Set[int]`:

```go
func (s Set[int]) Sum() int {
	total := 0
	for k := range s.m {
		total += k // invalid operation: total += k (mismatched types int and int /* with int declared at ... */)
	}
	return total
}
```

`int` here is a type parameter that *shadows* the predeclared `int`. The error message ("mismatched types int and int") is the giveaway. Go has no specialization; write a plain function `func SumSet(s Set[int]) int` or a generic one with a tighter constraint. More on that in the next lesson.

## Pointer receivers and lazy initialisation

A generic container with a useful zero value usually creates its internal map on first use. That **only works with pointer receivers**:

```go
package main

import "fmt"

type Counter[K comparable] struct {
	m map[K]int
}

func (c *Counter[K]) Inc(k K) {
	if c.m == nil {
		c.m = make(map[K]int) // writes to the caller's Counter
	}
	c.m[k]++
}

func (c *Counter[K]) Get(k K) int { return c.m[k] } // reading a nil map is fine

func (c *Counter[_]) Len() int { return len(c.m) }

func main() {
	var hits Counter[string] // no constructor needed
	hits.Inc("/index")
	hits.Inc("/index")
	hits.Inc("/about")
	fmt.Println(hits.Len(), hits.Get("/index"), hits.Get("/missing"))
}
```

```
2 2 0
```

With a value receiver, `Inc` would build the map in a copy and lose it (the quiz shows exactly that). Rule: if **any** method needs a pointer receiver, give them all pointer receivers, just as for non-generic types.

The alternative is a constructor, `func NewCounter[K comparable]() *Counter[K]`, that makes the map up front. Constructors get type inference only from their arguments, so `NewCounter[string]()` needs an explicit type argument, whereas a zero value works with `var hits Counter[string]`. Many Stash types support both.

## What you can't do with a type parameter in a type

- **Embed it.** `type Box[T any] struct{ T }` fails with `embedded field type cannot be a (pointer to a) type parameter`. Name the field instead: `struct{ Value T }`.
- **Use it as the whole underlying type.** `type Wrapper[T any] T` is an error too; the compiler needs to know the structure of the type.
- **Add a constraint later.** A method can't require more of `K` than the type declaration does (no `Sum` method on `Counter[K comparable]` that needs `K` to be a number).

## Embedding generic types is fine

You *can* embed an **instantiated** generic type, and its methods are promoted as usual:

```go
type TaggedCounter struct {
	Counter[string] // promotes Inc, Get and Len
	Tags            []string
}
```

That's composition in the generic world. Stash's LRU cache in chapter 5 will be built this way from smaller pieces.
