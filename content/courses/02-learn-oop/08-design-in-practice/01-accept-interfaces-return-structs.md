---
title: Accept Interfaces, Return Structs
quiz:
  - question: |
      Which signature best follows "accept interfaces, return structs"?
    options:
      - text: '`func NewArena(log *os.File) Arena` where `Arena` is an interface'
      - text: '`func NewArena(log io.Writer) *Arena` where `Arena` is a struct'
        correct: true
      - text: '`func NewArena(log any) any`'
      - text: '`func NewArena(log *os.File) *Arena`'
    explanation: |
      Accepting `io.Writer` lets callers pass a file, a buffer, a network
      connection or `os.Stdout`. Returning the concrete `*Arena` gives callers
      every method it has, and they can still store it in an interface of their
      own if they want.
  - question: Why is returning an interface from a constructor often a bad idea?
    options:
      - text: Interfaces can't be returned from functions
      - text: It hides the concrete type's other methods and forces the package to decide, up front, which abstraction every caller needs
        correct: true
      - text: It makes the program use more memory
      - text: It breaks `gofmt`
    explanation: |
      Callers are better placed to know which small interface they need. If you
      return the concrete type, each caller can pick its own. Return an interface
      only when there's a real reason, such as several hidden implementations.
---

Here's a Go design guideline you'll hear often:

> **Accept interfaces, return structs.**

It's a short phrase that pulls together everything from the abstraction and polymorphism chapters.

## Accept interfaces

When a function *takes* something, ask for the smallest interface that does the job. Here's an arena that writes a battle log:

```go
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Arena struct {
	log    io.Writer
	rounds int
}

// NewArena accepts an interface...
func NewArena(log io.Writer) *Arena { // ...and returns a concrete struct.
	return &Arena{log: log}
}

func (a *Arena) Clash(attacker, defender string, dmg int) {
	a.rounds++
	fmt.Fprintf(a.log, "round %d: %s hits %s for %d\n", a.rounds, attacker, defender, dmg)
}

func (a *Arena) Rounds() int { return a.rounds }

func main() {
	// Log to the terminal.
	live := NewArena(os.Stdout)
	live.Clash("Aria", "Ember", 25)

	// Log into memory instead, e.g. to show in a replay screen.
	var replay strings.Builder
	quiet := NewArena(&replay)
	quiet.Clash("Ember", "Aria", 40)
	quiet.Clash("Aria", "Ember", 30)
	fmt.Printf("replay has %d rounds:\n%s", quiet.Rounds(), replay.String())
}
```

```
round 1: Aria hits Ember for 25
replay has 2 rounds:
round 1: Ember hits Aria for 40
round 2: Aria hits Ember for 30
```

Because `NewArena` accepts `io.Writer`, it works with the terminal, a `strings.Builder`, a file, a network connection, or a test buffer. `Arena` never needs to change. If it had demanded an `*os.File`, the replay feature would've needed a rewrite.

## Return structs

When a function *gives back* something, return the **concrete type**, usually a pointer to a struct:

- **Callers get everything.** A `*Arena` has `Clash` and `Rounds`, and any method you add later. If `NewArena` returned an interface with only `Clash`, `Rounds` would be hidden and callers would need a type assertion.
- **Callers choose their own abstraction.** Remember that consumers own interfaces. A caller who only needs to clash can declare `type Clasher interface{ Clash(string, string, int) }` and store the `*Arena` in it. Implicit satisfaction makes that free.
- **Adding methods doesn't break anyone.** Adding a method to a struct is backwards compatible. Adding a method to an exported interface breaks every type that implemented it.

## The exceptions

Like every proverb, this is a default, not a law:

- **`error`** is always returned as an interface. That's the convention, and it keeps the nil gotcha away (return literal `nil`).
- **Hiding several implementations.** If a constructor chooses between genuinely different implementations (`NewStorage` returns a disk- or memory-backed store depending on config), returning an interface can make sense.
- **Standard examples.** `io.MultiWriter` returns an `io.Writer`, because the concrete type is an internal detail with nothing extra to offer.

## Putting it in OOP terms

This guideline is Go's version of a classic OOP principle, "program to an interface, not an implementation", applied carefully in both directions. Your **inputs** are programmed to interfaces, so you're flexible about what you accept. Your **outputs** are concrete, so you're generous with what you provide.

## Further reading

- [Learn Go with Tests: Dependency Injection](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/dependency-injection)
