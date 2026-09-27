---
title: The Package Is the Boundary
quiz:
  - question: |
      In package `combat`, you declare `type Hero struct{ hp int }` and, in the same package, a function `func cheat(h *Hero) { h.hp = 9999 }`. What happens?
    options:
      - text: Compile error, because `hp` is private to `Hero`'s methods
      - text: It compiles, because unexported names are visible to all code in the same package
        correct: true
      - text: It compiles but panics at runtime
      - text: It only compiles if `cheat` is a method on `Hero`
    explanation: |
      Go's privacy boundary is the package, not the type. Any file and any
      function in package `combat` can read and write `hp`. Privacy only kicks in
      when code in a *different* package tries.
  - question: What does an `internal/` directory do in a Go module?
    options:
      - text: It hides the package from `go doc`
      - text: It makes every name in it unexported
      - text: Packages under it can only be imported by code rooted at the parent of `internal`
        correct: true
      - text: It marks the package as test-only
    explanation: |
      `rpg/internal/save` can be imported by `rpg/...` packages but not by
      outside modules. It's encapsulation at the package level: exported names that
      are still private to your project.
---

Here's a Go surprise for anyone coming from Java or Python: **Go has no per-type privacy.** Unexported names are private to the *package*, and every piece of code in that package can see them, including functions that have nothing to do with the type.

## Same package, full access

```go
package main

import "fmt"

type Dragon struct {
	name string
	hp   int
}

type Hero struct {
	name string
	hp   int
}

// A plain function, not a method, poking at both types' unexported fields.
func duel(h *Hero, d *Dragon) {
	d.hp -= 40
	h.hp -= 25
}

func main() {
	h := &Hero{name: "Aria", hp: 100}
	d := &Dragon{name: "Smolder", hp: 200}
	duel(h, d)
	fmt.Println(h.hp, d.hp)
}
```

```
75 160
```

`duel` isn't a method on either type, yet it freely changes both `hp` fields. In Go that's completely normal. Everything in `package main` is on the same team.

## Why Go draws the line there

A package is a unit of code written and maintained together, usually by the same people. Go's reasoning is:

- **Inside a package**, you understand all the code, so you're trusted to maintain the invariants. Forcing a getter between two tightly related types in the same file would just be ceremony.
- **Between packages** is where different teams, different release schedules and strangers' code meet. That's where you need hard walls.

So when you design with encapsulation in mind, the question becomes: **which package does this type live in, and what does that package export?**

## Designing packages for encapsulation

A typical layout for our game might look like this:

```
rpg/
├── go.mod
├── main.go            (package main)
├── hero/
│   └── hero.go        (package hero)
├── combat/
│   └── combat.go      (package combat)
└── internal/
    └── save/
        └── save.go    (package save)
```

- `hero` owns the `Hero` type and its rules. Its fields are unexported.
- `combat` can only affect heroes through the exported methods of `hero.Hero`. It *cannot* set `hp` directly, so it can't break the rules.
- `main` wires everything together.

If you notice two packages constantly needing each other's unexported bits, that's a hint they belong in one package.

## `internal/`: private packages

Sometimes a package needs exported names so your *other* packages can use them, but you don't want the rest of the world importing it. Go has a rule for this built into the toolchain:

> A package inside a directory named `internal` can only be imported by code rooted at the parent of that `internal` directory.

So `rpg/internal/save` can be imported by `rpg/main.go`, `rpg/hero` and `rpg/combat`, but if someone else's module tries `import "example.com/rpg/internal/save"`, the build fails. It's encapsulation one level up: exported within your project, hidden from everyone else.

## Small packages, not tiny ones

Don't create one package per type to get Java-style privacy. You'd spend your life writing getters between packages and fighting import cycles (Go forbids package A importing B while B imports A). Group types that work closely together, and put walls where genuinely separate responsibilities meet.
