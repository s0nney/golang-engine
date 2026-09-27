---
title: Methods
quiz:
  - question: |
      In `func (h Hero) IsAlive() bool`, what is `h`?
    options:
      - text: A type parameter
      - text: The receiver, the value the method was called on
        correct: true
      - text: A global variable
      - text: The return value
    explanation: |
      The part in parentheses before the method name is the receiver. It works
      like Python's `self`, except that you choose its name, and Go convention is
      a short one or two letter abbreviation of the type.
  - question: Can you declare a method on a type that isn't a struct?
    options:
      - text: No, only structs can have methods
      - text: Yes, on any named type declared in the same package, like `type Gold int`
        correct: true
      - text: Yes, on any type at all, including built-in `int` directly
      - text: Only on interfaces
    explanation: |
      Any named type you declare in your package can have methods, whether it's
      based on a struct, an int, a slice or a func. You can't add methods to types
      from other packages, including built-ins like `int`, so you declare your own
      named type first.
---

A **method** is a function attached to a type. It's how Go gives structs behaviour.

## Declaring a method

A method looks like a function with an extra parameter, called the **receiver**, written *before* the name:

```go
type Hero struct {
	Name string
	HP   int
}

func (h Hero) IsAlive() bool {
	return h.HP > 0
}
```

Read it as "`IsAlive` is a method on `Hero`; inside it, the hero is called `h`". You call it with a dot, just like accessing a field:

```go
package main

import "fmt"

type Hero struct {
	Name string
	HP   int
}

func (h Hero) IsAlive() bool {
	return h.HP > 0
}

func (h Hero) Describe() string {
	status := "standing"
	if !h.IsAlive() {
		status = "fallen"
	}
	return fmt.Sprintf("%s (%d HP) is %s", h.Name, h.HP, status)
}

func main() {
	aria := Hero{Name: "Aria", HP: 100}
	borin := Hero{Name: "Borin", HP: 0}
	fmt.Println(aria.Describe())
	fmt.Println(borin.Describe())
}
```

```
Aria (100 HP) is standing
Borin (0 HP) is fallen
```

## The receiver is Go's `self`

In Python you'd write:

```python
class Hero:
    def is_alive(self):
        return self.hp > 0
```

Go's receiver plays the role of `self`, but with a few differences:

- **You name it.** Convention is a short abbreviation of the type: `h` for `Hero`, `d` for `Dragon`. Don't call it `self` or `this`; linters will complain.
- **It's explicit about copying.** `(h Hero)` gets a *copy* of the hero; `(h *Hero)` gets a pointer to the original. More on that soon.
- **Methods live outside the type.** There's no class body. You declare methods anywhere in the same package, usually right after the type.

## Methods are just functions

Under the hood, a method is a function whose first argument is the receiver. Go even lets you write it that way, as a *method expression*:

```go
alive := Hero.IsAlive(aria) // same as aria.IsAlive()
```

You'll rarely need that, but it demystifies things. There's no hidden object machinery. `aria.IsAlive()` is simply a nicer way to call a function with `aria` as the first argument.

## Methods on any named type

Methods aren't only for structs. Any named type declared in your package can have them:

```go
type Gold int

func (g Gold) String() string {
	return fmt.Sprintf("%d gp", int(g))
}
```

Now `Gold(250).String()` returns `"250 gp"`. You can't add methods to `int` itself, or to types from other packages. That rule keeps packages from monkey-patching each other, which is a small but real win for readability.

## When to use a method vs a function

Use a method when the behaviour clearly *belongs* to the type: a hero's `IsAlive`, a dragon's `Breathe`. Use a plain function when it involves several types equally, such as `resolveCombat(hero, dragon)`. Go doesn't force everything into an object, and that's a feature.

## Further reading

- [A Tour of Go: Methods](https://go.dev/tour/methods/1)
- [Go by Example: Methods](https://gobyexample.com/methods)
