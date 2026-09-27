---
title: Why Paradigms Exist
quiz:
  - question: Who is the main audience for the way you organise your code?
    options:
      - text: The compiler, because it needs a particular structure to run fast
      - text: The CPU, because objects map directly onto hardware
      - text: Humans, including future you, who have to read and change the code
        correct: true
      - text: Nobody. Structure is purely a matter of taste
    explanation: |
      The computer will happily run a 10,000-line `main` function. Paradigms like
      object-oriented and functional programming exist to help *people* understand
      and safely change code as it grows.
  - question: How does Go relate to object-oriented programming?
    options:
      - text: Go is a pure OOP language with classes and inheritance, just like Java
      - text: Go supports OOP ideas like methods, encapsulation and polymorphism, but without classes or inheritance
        correct: true
      - text: Go is purely functional and has no way to attach behaviour to data
      - text: Go bans OOP entirely, so you must write everything as free functions
    explanation: |
      Go borrows the useful parts of OOP (methods on types, package-level privacy,
      interfaces for polymorphism) and leaves out classes and inheritance on purpose,
      preferring composition.
---

Welcome to the Object-Oriented Programming course! Grab your sword, because we'll be building a small fantasy RPG as we go: heroes, archers, dragons and a combat system to make them fight.

Before any of that, a question: **why do programming paradigms exist at all?**

## The computer doesn't care

Here's a working "game":

```go
package main

import "fmt"

func main() {
	h := 100
	d := 250
	h -= 30
	d -= 45
	d -= 45
	fmt.Println(h, d)
}
```

It prints:

```
70 160
```

The CPU is perfectly happy with this. It doesn't know that `h` is a hero's health or that `d` is a dragon's. Now imagine this file after two years of work: 40 heroes, 12 spells, a save system and a shop. Every change means reading thousands of lines of `h`, `d`, `x2` and `tmp`.

The problem is never "can the computer run it?" It's **"can a human understand it and change it without breaking something?"**

## Paradigms are ways of organising

A *programming paradigm* is a style of organising code. The big three you'll hear about:

- **Procedural**: a list of steps, grouped into functions. This is what you've been writing so far.
- **Functional**: build programs from pure functions that take inputs and return outputs, without changing shared state.
- **Object-oriented (OOP)**: bundle data together with the behaviour that operates on it, into *objects*.

None of these is "correct". They're tools. Good programmers mix them.

## The four pillars

OOP is usually described with four big ideas, and each gets its own chapter in this course:

1. **Encapsulation**: hide an object's internals so nobody can put it into a broken state.
2. **Abstraction**: hide complexity behind a simple interface.
3. **Inheritance**: share behaviour between related types.
4. **Polymorphism**: treat different types the same way through a common interface.

## Where Go stands

Languages like Python and Java build OOP around **classes**. Go has no classes. It has:

- **structs** for data,
- **methods** you can attach to any named type,
- **interfaces** for polymorphism,
- **packages** for encapsulation,
- **embedding** for composition, instead of inheritance.

So this course works in two steps. For each OOP idea, we'll learn what problem it solves, then see how Go solves the same problem, sometimes in a very different (and often simpler) way. When Go deliberately *doesn't* copy a feature, we'll look at why.

Let's start with the thing every paradigm is ultimately chasing: clean code.
