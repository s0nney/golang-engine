---
title: Embedding Is Not Subclassing
quiz:
  - question: |
      What does this print?

      ```go
      type Monster struct{}

      func (m Monster) Name() string { return "monster" }
      func (m Monster) Roar() string { return "The " + m.Name() + " roars!" }

      type Dragon struct{ Monster }

      func (d Dragon) Name() string { return "dragon" }

      func main() {
      	fmt.Println(Dragon{}.Roar())
      }
      ```
    options:
      - text: '`The dragon roars!`'
      - text: '`The monster roars!`'
        correct: true
      - text: It doesn't compile because `Name` is declared twice
      - text: It panics
    explanation: |
      `Roar` is promoted from `Monster`, so it runs with a `Monster` receiver.
      Inside it, `m.Name()` calls `Monster.Name`. The embedded value has no idea
      it's inside a `Dragon`, so there's no virtual dispatch back to `Dragon.Name`.
  - question: How do you get "the base code calls the specialised behaviour" in Go?
    options:
      - text: Mark the method `virtual`
      - text: Use `super` inside the embedded type
      - text: Have the shared code take an interface parameter, and pass the outer value in
        correct: true
      - text: It's impossible in Go
    explanation: |
      Dynamic dispatch in Go goes through interfaces. Write `func Roar(n Namer) string`
      (or store an interface field) and pass the `Dragon`; then `n.Name()` calls
      `Dragon.Name`.
---

This lesson covers the single most common mistake people make when they arrive from Python or Java: treating embedding as inheritance. It *looks* similar. It behaves differently in one crucial way.

## The template method trap

In Python, this works the way you'd hope:

```python
class Monster:
    def name(self):
        return "monster"

    def roar(self):
        return f"The {self.name()} roars!"

class Dragon(Monster):
    def name(self):
        return "dragon"

print(Dragon().roar())  # The dragon roars!
```

`roar` is written once in `Monster`, but `self.name()` calls `Dragon.name`, because `self` really is the dragon. That's called **virtual dispatch**: the base class calls back into the subclass.

Now the same thing in Go:

```go
package main

import "fmt"

type Monster struct {
	HP int
}

func (m Monster) Name() string { return "monster" }

func (m Monster) Roar() string {
	return "The " + m.Name() + " roars!"
}

type Dragon struct {
	Monster
}

func (d Dragon) Name() string { return "dragon" }

func main() {
	d := Dragon{Monster{HP: 300}}
	fmt.Println(d.Name())
	fmt.Println(d.Roar())
}
```

```
dragon
The monster roars!
```

`d.Name()` gives `dragon`, as expected: the outer method shadows the inner one. But `d.Roar()` says *monster*!

## Why

Remember what promotion really does. `d.Roar()` is shorthand for `d.Monster.Roar()`. Its receiver is the embedded `Monster` value, a plain `Monster` that knows nothing about the `Dragon` around it. Inside `Roar`, `m.Name()` can only mean `Monster.Name`.

There is no `self` that points to the "real" object. The inner value never gets a reference to the outer one. Embedding is **composition with some syntactic sugar**, and the sugar only works outside-in, never inside-out.

This is deliberate. It kills the fragile base class problem: nothing a `Dragon` does can change how `Monster`'s own methods behave, so `Monster` can be understood entirely on its own.

## Getting the behaviour you wanted

In Go, dynamic dispatch goes through **interfaces**. When shared code needs to call behaviour that varies, make that explicit by taking an interface:

```go
package main

import "fmt"

type Namer interface {
	Name() string
}

func Roar(n Namer) string {
	return "The " + n.Name() + " roars!"
}

type Monster struct{ HP int }

func (m Monster) Name() string { return "monster" }

type Dragon struct{ Monster }

func (d Dragon) Name() string { return "dragon" }

type Goblin struct{ Monster } // no Name override

func main() {
	fmt.Println(Roar(Dragon{}))
	fmt.Println(Roar(Goblin{}))
}
```

```
The dragon roars!
The monster roars!
```

`Roar` now receives the whole `Dragon` as a `Namer`, so `n.Name()` dispatches to `Dragon.Name`. `Goblin` didn't override `Name`, so it uses the promoted `Monster.Name`, which is a sensible default. You get "override what you like, inherit the rest" without any hidden callbacks.

The key difference from inheritance: **the dependency is visible**. The signature `Roar(n Namer)` tells you precisely which behaviour can vary. In the Python version, any method `roar` calls on `self` might secretly be overridden somewhere.

## Other things embedding doesn't do

- **No "is-a".** A `Dragon` can't be passed where a `Monster` is expected. Pass `d.Monster`, or better, accept an interface.
- **No protected members.** Unexported fields follow the package rule, not a family rule.
- **No constructor chaining.** There's no automatic `super().__init__()`. A `NewDragon` function builds its `Monster` part explicitly.

## The takeaway

Use embedding for what it's good at: **reusing** methods and forwarding calls. Use **interfaces** whenever code needs to work with "anything that behaves like X". Keep those two jobs separate and Go's design clicks into place, which is exactly where the next chapter, on polymorphism, picks up.
