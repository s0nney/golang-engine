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
exercise:
  starter: |
    package main

    import "fmt"

    // Stack is a last-in, first-out pile of values.
    type Stack[T any] struct {
    	items []T
    }

    // Push puts v on top of the stack.
    func (s *Stack[T]) Push(v T) {
    	// ?
    }

    // Pop removes and returns the top value. On an empty stack it returns
    // the zero value of T and false.
    func (s *Stack[T]) Pop() (T, bool) {
    	var zero T
    	// ?
    	return zero, false
    }

    // Peek returns the top value without removing it, or (zero, false)
    // on an empty stack.
    func (s *Stack[T]) Peek() (T, bool) {
    	var zero T
    	// ?
    	return zero, false
    }

    // Len reports how many values are on the stack.
    func (s *Stack[T]) Len() int {
    	// ?
    	return 0
    }

    type Spell struct {
    	Name string
    	Cost int
    }

    func main() {
    	var chain Stack[Spell]
    	chain.Push(Spell{Name: "fireball", Cost: 30})
    	chain.Push(Spell{Name: "counterspell", Cost: 20})

    	top, _ := chain.Peek()
    	fmt.Println("on top:", top.Name)
    	for chain.Len() > 0 {
    		s, _ := chain.Pop()
    		fmt.Println("resolve", s.Name)
    	}
    	_, ok := chain.Pop()
    	fmt.Println("anything left?", ok)
    }
  solution: |
    package main

    import "fmt"

    // Stack is a last-in, first-out pile of values.
    type Stack[T any] struct {
    	items []T
    }

    func (s *Stack[T]) Push(v T) {
    	s.items = append(s.items, v)
    }

    func (s *Stack[T]) Pop() (T, bool) {
    	var zero T
    	if len(s.items) == 0 {
    		return zero, false
    	}
    	v := s.items[len(s.items)-1]
    	s.items = s.items[:len(s.items)-1]
    	return v, true
    }

    func (s *Stack[T]) Peek() (T, bool) {
    	var zero T
    	if len(s.items) == 0 {
    		return zero, false
    	}
    	return s.items[len(s.items)-1], true
    }

    func (s *Stack[T]) Len() int { return len(s.items) }

    type Spell struct {
    	Name string
    	Cost int
    }

    func main() {
    	var chain Stack[Spell]
    	chain.Push(Spell{Name: "fireball", Cost: 30})
    	chain.Push(Spell{Name: "counterspell", Cost: 20})

    	top, _ := chain.Peek()
    	fmt.Println("on top:", top.Name)
    	for chain.Len() > 0 {
    		s, _ := chain.Pop()
    		fmt.Println("resolve", s.Name)
    	}
    	_, ok := chain.Pop()
    	fmt.Println("anything left?", ok)
    }
  tests: |
    package main

    import "testing"

    func TestStackOfStrings(t *testing.T) {
    	var s Stack[string]
    	if s.Len() != 0 {
    		t.Fatalf("new stack: Len() = %d, want 0", s.Len())
    	}
    	if v, ok := s.Pop(); ok || v != "" {
    		t.Errorf("Pop() on empty stack = (%q, %v), want (\"\", false)", v, ok)
    	}
    	if v, ok := s.Peek(); ok || v != "" {
    		t.Errorf("Peek() on empty stack = (%q, %v), want (\"\", false)", v, ok)
    	}
    	s.Push("fireball")
    	s.Push("counterspell")
    	s.Push("shield")
    	if s.Len() != 3 {
    		t.Errorf("after 3 pushes: Len() = %d, want 3", s.Len())
    	}
    	if v, ok := s.Peek(); !ok || v != "shield" {
    		t.Errorf("Peek() = (%q, %v), want (\"shield\", true)", v, ok)
    	}
    	if s.Len() != 3 {
    		t.Errorf("Peek must not remove anything: Len() = %d, want 3", s.Len())
    	}
    	for _, want := range []string{"shield", "counterspell", "fireball"} {
    		if v, ok := s.Pop(); !ok || v != want {
    			t.Errorf("Pop() = (%q, %v), want (%q, true)", v, ok, want)
    		}
    	}
    	if v, ok := s.Pop(); ok {
    		t.Errorf("Pop() after emptying = (%q, true), want ok=false", v)
    	}
    	if s.Len() != 0 {
    		t.Errorf("after popping everything: Len() = %d, want 0", s.Len())
    	}
    }

    func TestStackOfInts(t *testing.T) {
    	var s Stack[int]
    	for i := range 5 {
    		s.Push(i * 10)
    	}
    	if v, ok := s.Pop(); !ok || v != 40 {
    		t.Errorf("Stack[int]: Pop() = (%d, %v), want (40, true)", v, ok)
    	}
    	s.Push(99)
    	if v, _ := s.Peek(); v != 99 {
    		t.Errorf("Stack[int]: Peek() after Push(99) = %d, want 99", v)
    	}
    	if s.Len() != 5 {
    		t.Errorf("Stack[int]: Len() = %d, want 5", s.Len())
    	}
    }
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

## Your turn

Some spells can be answered by other spells: a counterspell cast in response to a fireball resolves *first*. That's a **stack**, last in, first out.

Complete the generic `Stack[T]`: `Push` adds to the top, `Pop` removes and returns the top value (or the zero value of `T` and `false` when empty), `Peek` returns the top without removing it, and `Len` reports the size. The tests use a `Stack[string]` and a `Stack[int]`, so don't assume anything about `T`.

## Further reading

- [Go by Example: Generics](https://gobyexample.com/generics)
- [An Introduction To Generics](https://go.dev/blog/intro-generics)
