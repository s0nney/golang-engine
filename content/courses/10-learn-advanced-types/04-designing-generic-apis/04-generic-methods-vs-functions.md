---
title: Generic Methods vs Functions
quiz:
  - question: |
      You want a `Sum` operation on `List[T any]` that only works when `T` is numeric. What's the right shape?
    options:
      - text: '`func (l *List[T]) Sum() T`'
      - text: '`func (l *List[T]) Sum[T Number]() T`'
      - text: '`func Sum[T Number](l *List[T]) T`, a plain generic function'
        correct: true
      - text: '`func (l *List[int]) Sum() int`'
    explanation: |
      A method can't tighten the receiver's constraint: `T` is `any` in every method of
      `List`. A generic method can only add *new* type parameters. A function can declare
      its own, stricter constraint on the list's element type.
  - question: |
      Why can't `*List[T]` with a generic method `Map[U any](func(T) U) *List[U]` satisfy this interface?

      ```go
      type Mapper interface {
      	Map(func(string) int) *List[int]
      }
      ```
    options:
      - text: Interfaces can't mention generic types
      - text: A generic method is a family of methods, and it can never satisfy an interface method, even one that matches an instantiation
        correct: true
      - text: It can, as long as `T` is `string`
      - text: '`Map` would need a value receiver'
    explanation: |
      Go 1.27 allows generic methods but keeps them out of interfaces entirely: an
      interface can't declare one, and a generic method doesn't count toward
      implementing one. Wrap it in a non-generic method if you need to satisfy an interface.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // List is Stash's ordered collection.
    type List[T any] struct {
    	items []T
    }

    // Push appends values to the list.
    func (l *List[T]) Push(vs ...T) { l.items = append(l.items, vs...) }

    // Items returns the list's values.
    func (l *List[T]) Items() []T { return l.items }

    // Map returns a new list with f applied to every element, in order.
    func (l *List[T]) Map[U any](f func(T) U) *List[U] {
    	// ?
    	return &List[U]{}
    }

    // Fold combines the elements from left to right, starting from init:
    // f(f(f(init, a), b), c) for a list [a b c].
    func (l *List[T]) Fold[A any](init A, f func(A, T) A) A {
    	// ?
    	return init
    }

    // Number is anything Sum can add up.
    type Number interface {
    	~int | ~int64 | ~float64
    }

    // Sum adds up a list of numbers. It can't be a method: List's T is only
    // constrained by any. Implement it using Fold.
    func Sum[T Number](l *List[T]) T {
    	// ?
    	var zero T
    	return zero
    }

    func main() {
    	var words List[string]
    	words.Push("stash", "generic", "types")
    	lengths := words.Map(func(s string) int { return len(s) })
    	fmt.Println(lengths.Items(), Sum(lengths))
    	shout := words.Fold("", func(acc string, s string) string { return acc + strings.ToUpper(s[:1]) })
    	fmt.Println(shout)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // List is Stash's ordered collection.
    type List[T any] struct {
    	items []T
    }

    // Push appends values to the list.
    func (l *List[T]) Push(vs ...T) { l.items = append(l.items, vs...) }

    // Items returns the list's values.
    func (l *List[T]) Items() []T { return l.items }

    // Map returns a new list with f applied to every element, in order.
    func (l *List[T]) Map[U any](f func(T) U) *List[U] {
    	out := &List[U]{items: make([]U, 0, len(l.items))}
    	for _, v := range l.items {
    		out.items = append(out.items, f(v))
    	}
    	return out
    }

    // Fold combines the elements from left to right, starting from init.
    func (l *List[T]) Fold[A any](init A, f func(A, T) A) A {
    	acc := init
    	for _, v := range l.items {
    		acc = f(acc, v)
    	}
    	return acc
    }

    // Number is anything Sum can add up.
    type Number interface {
    	~int | ~int64 | ~float64
    }

    // Sum adds up a list of numbers.
    func Sum[T Number](l *List[T]) T {
    	return l.Fold(0, func(acc, v T) T { return acc + v })
    }

    func main() {
    	var words List[string]
    	words.Push("stash", "generic", "types")
    	lengths := words.Map(func(s string) int { return len(s) })
    	fmt.Println(lengths.Items(), Sum(lengths))
    	shout := words.Fold("", func(acc string, s string) string { return acc + strings.ToUpper(s[:1]) })
    	fmt.Println(shout)
    }
  tests: |
    package main

    import (
    	"slices"
    	"strconv"
    	"testing"
    )

    type Bytes int64

    func TestMap(t *testing.T) {
    	var l List[int]
    	l.Push(1, 2, 3)
    	strs := l.Map(strconv.Itoa)
    	if got := strs.Items(); !slices.Equal(got, []string{"1", "2", "3"}) {
    		t.Errorf("List[int]{1 2 3}.Map(strconv.Itoa) = %q, want [1 2 3]", got)
    	}
    	doubled := l.Map(func(n int) int { return n * 2 })
    	if got := doubled.Items(); !slices.Equal(got, []int{2, 4, 6}) {
    		t.Errorf("Map(double) = %v, want [2 4 6]", got)
    	}
    	if got := l.Items(); !slices.Equal(got, []int{1, 2, 3}) {
    		t.Errorf("Map must not change the original list, now %v", got)
    	}
    	var empty List[string]
    	if got := empty.Map(func(s string) int { return len(s) }).Items(); len(got) != 0 {
    		t.Errorf("Map on an empty list = %v, want empty", got)
    	}
    }

    func TestFold(t *testing.T) {
    	var l List[string]
    	l.Push("a", "b", "c")
    	got := l.Fold(">", func(acc, s string) string { return acc + s })
    	if got != ">abc" {
    		t.Errorf(`Fold(">", concat) over [a b c] = %q, want ">abc" (left to right)`, got)
    	}
    	n := l.Fold(10, func(acc int, s string) int { return acc + len(s) })
    	if n != 13 {
    		t.Errorf("Fold(10, add lengths) = %d, want 13", n)
    	}
    }

    func TestSum(t *testing.T) {
    	var ints List[int]
    	ints.Push(5, 7, 30)
    	if got := Sum(&ints); got != 42 {
    		t.Errorf("Sum([5 7 30]) = %d, want 42", got)
    	}
    	var sizes List[Bytes]
    	sizes.Push(512, 1024)
    	if got := Sum(&sizes); got != 1536 {
    		t.Errorf("Sum([512 1024] Bytes) = %d, want 1536", got)
    	}
    	var none List[float64]
    	if got := Sum(&none); got != 0 {
    		t.Errorf("Sum(empty) = %v, want 0", got)
    	}
    }
---

Go 1.27 lets methods declare their own type parameters, as you saw in [Learn OOP](/courses/learn-oop/generics-and-oop/generic-methods). Now that you *can* write `list.Map(f)`, when *should* you, and when is a plain generic function still the better tool?

## What each can express

A **method** on `List[T]` sees `T` exactly as the type declared it, with the type's constraint. A generic method can add **new** type parameters, but it can't say anything new about `T`.

A **function** declares all of its type parameters itself, so it can put any constraint it likes on the element type:

```go
type List[T any] struct{ items []T }

// Method: T is `any` here, full stop. U is new.
func (l *List[T]) Map[U any](f func(T) U) *List[U]

// Function: free to require more of the element type.
func Sum[T Number](l *List[T]) T
func Sorted[T cmp.Ordered](l *List[T]) *List[T]
```

So the first test is mechanical: **if the operation needs a stronger constraint on the receiver's type parameters, it must be a function.** `Sum`, `Max`, `Sorted`, `Contains` (needs `comparable`) all fall in this bucket for an `any`-constrained container.

## When a generic method is better

When the operation works for *any* `T` and is naturally "something you do to this value", a method reads best and keeps related operations together in the docs:

```go
lengths := words.Map(func(s string) int { return len(s) })
total := words.Fold(0, func(n int, s string) int { return n + len(s) })
```

Methods also chain nicely: `list.Map(f).Map(g)`. With functions you'd write `Map(Map(list, f), g)`, which reads inside out.

## When a function is better anyway

Even when a method is *possible*, prefer a function if:

- **It operates on several values symmetrically.** `Union(a, b)` treats both sets the same; `a.Union(b)` suggests `a` is special. (Either is fine; be consistent.)
- **It works across different container types.** A `Collect[T any](seq iter.Seq[T]) []T` works with every iterator, not just one type's.
- **It must satisfy an interface.** Generic methods can't implement interface methods (see [the rules](/courses/learn-oop/generics-and-oop/generic-methods-and-interfaces)). If `List` needs to be a `fmt.Stringer` or a `sort.Interface`, those methods must be ordinary ones.
- **Compatibility matters.** Code that has to build with Go versions before 1.27 can't use generic methods at all.

## Instantiation gotchas

A generic method must be instantiated before it's used as a value, just like a function: `f := l.Map[int]`, never `f := l.Map`. And since interface types can't declare generic methods, you can't abstract over "anything with a `Map` method" with an interface. If you find yourself wanting that, it's a sign the operation should be a function over a shared representation, such as an `iter.Seq[T]` (chapter 5).

## A Stash rule of thumb

Stash uses:

- **Ordinary methods** for the core operations of a type: `Add`, `Get`, `Len`, `All`.
- **Generic methods** for transformations that work for every element type: `Map`, `Fold`.
- **Functions** for anything that needs more from the element type (`Sum`, `Sorted`), or combines several collections.

## Your turn

Give Stash's `List[T]`:

- `Map[U any](f func(T) U) *List[U]`: a generic method returning a **new** list.
- `Fold[A any](init A, f func(A, T) A) A`: a generic method that combines elements left to right.
- `Sum[T Number](l *List[T]) T`: a **function**, because it needs `T` to be a number. Implement it with `Fold`.
