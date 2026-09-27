---
title: Structs as Objects
quiz:
  - question: What is the closest thing Go has to a class?
    options:
      - text: The `class` keyword, which works like Python's
      - text: A named type (usually a struct) plus the methods declared on it
        correct: true
      - text: A package containing only functions
      - text: A map from strings to functions
    explanation: |
      Go has no `class` keyword. You declare a named type, most often a struct,
      to hold the data, and then attach methods to it. Together they play the role
      a class plays in other languages.
  - question: |
      What does this print?

      ```go
      type Hero struct {
      	Name string
      	HP   int
      }

      func main() {
      	a := Hero{Name: "Aria", HP: 100}
      	b := a
      	b.HP = 10
      	fmt.Println(a.HP, b.HP)
      }
      ```
    options:
      - text: '`10 10`'
      - text: '`100 100`'
      - text: '`100 10`'
        correct: true
      - text: It doesn't compile
    explanation: |
      Assigning a struct copies all its fields. `b` is a completely separate
      `Hero`, so changing `b.HP` leaves `a.HP` at 100. Python objects behave
      differently: there, `b = a` makes both names point to the same object.
---

In object-oriented programming, an **object** is a bundle of *state* (data) and *behaviour* (functions that act on that data). In Python you'd write a `class Hero:` to describe what every hero looks like, then create hero objects from it.

Go has no classes. Instead, it gives you two separate tools you've already met:

1. a **struct** to describe the data,
2. **methods** to attach behaviour to it (next lesson).

## A hero as a struct

Without structs, a hero is just loose variables:

```go
heroName := "Aria"
heroHP := 100
heroAttack := 18
```

With three heroes that's nine variables, and nothing stops you from pairing Aria's name with Borin's health. A struct groups them:

```go
type Hero struct {
	Name   string
	HP     int
	Attack int
}
```

`Hero` is now a **type**, a blueprint, just like a class. Each value you create from it is an **instance**, or, loosely, an object:

```go
package main

import "fmt"

type Hero struct {
	Name   string
	HP     int
	Attack int
}

func main() {
	aria := Hero{Name: "Aria", HP: 100, Attack: 18}
	borin := Hero{Name: "Borin", HP: 140, Attack: 12}

	party := []Hero{aria, borin}
	for _, h := range party {
		fmt.Printf("%s has %d HP\n", h.Name, h.HP)
	}
}
```

```
Aria has 100 HP
Borin has 140 HP
```

## Go structs are values

Here's the first big difference from Python. In Python, `b = a` makes `b` another name for the same object. In Go, assigning a struct **copies** it:

```go
a := Hero{Name: "Aria", HP: 100}
b := a    // a full copy
b.HP = 10 // only changes the copy
// a.HP is still 100
```

The same applies when you pass a struct to a function: the function gets its own copy. If you want to share one hero between several places, you use a pointer, `*Hero`. We'll see exactly when that matters in the receivers lesson.

## Printing structs

The `%+v` verb prints field names too, which is handy while debugging:

```go
fmt.Printf("%+v\n", Hero{Name: "Aria", HP: 100, Attack: 18})
// {Name:Aria HP:100 Attack:18}
```

## Classes, translated

| Python OOP             | Go                              |
|------------------------|---------------------------------|
| `class Hero:`          | `type Hero struct { ... }`      |
| instance attributes    | struct fields                   |
| `Hero("Aria", 100)`    | `Hero{Name: "Aria", HP: 100}`   |
| `def attack(self):`    | a method, `func (h Hero) ...`   |
| `__init__`             | a constructor function `NewHero`|

A struct on its own is just data. It's the methods, coming up next, that turn it into something that feels like an object.
