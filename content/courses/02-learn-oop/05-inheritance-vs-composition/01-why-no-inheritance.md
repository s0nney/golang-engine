---
title: Why Go Has No Inheritance
quiz:
  - question: What does "favour composition over inheritance" mean?
    options:
      - text: Never write more than one struct per file
      - text: Build types by combining smaller parts that each do one job ("has-a"), rather than by extending a parent class ("is-a")
        correct: true
      - text: Write functions instead of methods
      - text: Use interfaces for everything, including data
    explanation: |
      Composition builds a `Dragon` that *has* wings, a fire breath and a hoard,
      each a separate piece. Inheritance says a `Dragon` *is a* `FlyingMonster`
      which *is a* `Monster`. Go only offers the first.
  - question: Which of these is a well-known problem with deep inheritance hierarchies?
    options:
      - text: They make programs use too little memory
      - text: A change to a base class can silently break subclasses far away (the fragile base class problem)
        correct: true
      - text: They prevent methods from having parameters
      - text: They make it impossible to use interfaces
    explanation: |
      Subclasses depend on the base class's internal behaviour, not just its
      public contract. Tweak the base and distant subclasses can break without any
      compiler warning.
---

The third pillar of classic OOP is **inheritance**: a class can *extend* another class and get all its fields and methods for free.

In Python, boot.dev-style:

```python
class Hero:
    def __init__(self, name, hp):
        self.name = name
        self.hp = hp

    def take_damage(self, n):
        self.hp -= n

class Archer(Hero):
    def __init__(self, name, hp, arrows):
        super().__init__(name, hp)
        self.arrows = arrows

    def shoot(self):
        self.arrows -= 1
        return 8
```

An `Archer` *is a* `Hero`. It inherits `name`, `hp` and `take_damage` and adds arrows and shooting. Neat!

Go has **no inheritance**. No `extends`, no `super()`, no subclasses. That was a deliberate choice, not an oversight.

## Where inheritance goes wrong

Inheritance works beautifully in small examples. Problems show up as a game grows.

**Rigid hierarchies.** Let's add more:

```
Character
├── Hero
│   ├── Archer
│   ├── Warrior
│   └── Mage
└── Monster
    ├── Dragon
    └── Goblin
```

Now design wants a *Ranger*: a hero who shoots like an archer and has an animal companion like a druid. Or a *Dragon Knight*: a hero who can breathe fire like a dragon. Inheritance forces every type into exactly one place in the tree, and real designs refuse to be trees.

**The fragile base class.** Subclasses depend on *how* the parent works, not just *what* it promises. Change `Hero.take_damage` to call `self.on_hit()` and suddenly every subclass that defined an `on_hit` for some other reason starts misbehaving.

**The gorilla and the banana.** Joe Armstrong (creator of Erlang) put it memorably: *"You wanted a banana but what you got was a gorilla holding the banana and the entire jungle."* Inherit from `Hero` to get `take_damage`, and you also get its levelling system, inventory, dialogue and save format, whether you want them or not.

## Composition: has-a instead of is-a

The alternative is **composition**: build types out of smaller parts, each with one job.

A dragon isn't a `FlyingMonster` subclass. A dragon **has** health, **has** wings, **has** a breath attack:

```go
package main

import "fmt"

type Health struct {
	HP, MaxHP int
}

func (h *Health) TakeDamage(n int) { h.HP = max(0, h.HP-n) }

type Quiver struct {
	Arrows int
}

func (q *Quiver) Shoot() int {
	if q.Arrows == 0 {
		return 0
	}
	q.Arrows--
	return 8
}

type Archer struct {
	Name   string
	Health Health
	Quiver Quiver
}

func main() {
	lyra := Archer{Name: "Lyra", Health: Health{HP: 80, MaxHP: 80}, Quiver: Quiver{Arrows: 2}}
	fmt.Println(lyra.Quiver.Shoot(), lyra.Quiver.Shoot(), lyra.Quiver.Shoot())
	lyra.Health.TakeDamage(30)
	fmt.Println(lyra.Name, lyra.Health.HP, lyra.Quiver.Arrows)
}
```

```
8 8 0
Lyra 50 0
```

A Ranger? Give it a `Quiver` and a `Companion`. A Dragon Knight? A `Health`, a `Sword` and a `FireBreath`. Parts snap together in any combination, and no part knows or cares what it's inside.

## So how does Go share behaviour?

Go splits the jobs that inheritance does into two separate tools:

| Inheritance gives you...                   | Go gives you...                         |
|--------------------------------------------|-----------------------------------------|
| reuse of fields and methods                | **struct embedding** (composition)      |
| "an Archer can be used where a Hero is expected" | **interfaces** (polymorphism)     |

Keeping them separate is the key insight. You can reuse code without claiming an "is-a" relationship, and you can be used polymorphically without sharing any code.

Writing `lyra.Quiver.Shoot()` is a bit wordy, though. Next lesson, embedding lets you write `lyra.Shoot()` instead.

## Further reading

- [Go FAQ: Why is there no type inheritance?](https://go.dev/doc/faq#inheritance)
