---
title: Exported vs Unexported
quiz:
  - question: Which of these identifiers is visible outside its package?
    options:
      - text: '`maxHP`'
      - text: '`_Name`'
      - text: '`TakeDamage`'
        correct: true
      - text: '`hero`'
    explanation: |
      A name is exported if it starts with an upper-case letter. `_Name` starts
      with an underscore, which isn't upper-case, so it's unexported.
  - question: |
      Package `hero` declares:

      ```go
      type Hero struct {
      	Name string
      	hp   int
      }
      ```

      In package `main`, what happens with `h := hero.Hero{Name: "Aria", hp: 100}`?
    options:
      - text: It works, since `main` can set any field in a composite literal
      - text: It compiles but `hp` is silently ignored
      - text: It fails to compile, because `hp` is unexported and can't be referenced from `main`
        correct: true
      - text: It panics at runtime
    explanation: |
      Unexported fields can't be named outside their package, not even in a
      struct literal. `hero.Hero{Name: "Aria"}` would compile (leaving `hp` at 0),
      which is why you often pair unexported fields with a constructor.
---

Go's access control fits in one sentence:

> **A name that starts with an upper-case letter is exported (public). Anything else is unexported (private to its package).**

No `public`, `private` or `protected` keywords. The capital letter *is* the keyword.

## It applies to everything

The rule works the same way for every kind of name declared at package level, and for fields and methods:

```go
package hero

const MaxLevel = 50 // exported
const baseHP = 100  // unexported

type Hero struct { // exported type
	Name  string // exported field
	hp    int    // unexported field
	level int    // unexported field
}

func New(name string) *Hero { // exported function
	return &Hero{Name: name, hp: baseHP, level: 1}
}

func (h *Hero) TakeDamage(n int) { // exported method
	h.hp = max(0, h.hp-n)
}

func (h *Hero) levelUp() { // unexported method
	h.level = min(h.level+1, MaxLevel)
}
```

## Seeing it from outside

Now imagine a `main` package that imports this `hero` package. Here's what it can and can't do, with the compiler's real error messages:

```go
// In package main, which imports "example.com/rpg/hero":
aria := hero.New("Aria") // OK: New is exported
aria.TakeDamage(30)      // OK: TakeDamage is exported
fmt.Println(aria.Name)   // OK: Name is exported

aria.hp = 9999  // error: aria.hp undefined (cannot refer to unexported field hp)
aria.levelUp()  // error: aria.levelUp undefined (cannot refer to unexported method levelUp)
_ = hero.baseHP // error: undefined: hero.baseHP
```

The compiler enforces it. Unlike Python's underscore convention, there's no way to "just reach in", short of the `unsafe` or `reflect` packages, which you should treat as a sign you're doing something wrong.

## Why capital letters?

It seems odd at first, but it has a lovely property: **you can tell whether something is public by reading its name at the call site**. When you see `h.hp` you know you're inside the `hero` package. When you see `h.Name` you know any package could be touching it. No need to scroll up to find a declaration.

## A few gotchas

- **The first character decides.** `_Name`, `name` and `ñame` are unexported. Only upper-case letters export.
- **Exported type, unexported fields is common and good.** `hero.Hero` is visible, but its insides are guarded.
- **An unexported type with exported methods also works.** A package can return an unexported type through an exported function. Callers can use its exported methods but can't name the type. This is rare; prefer exporting the type and hiding the fields.
- **JSON and friends need exported fields.** Packages like `encoding/json` can only see exported fields. If your `Hero` saves to JSON, the fields it saves must be capitalised (or you write custom marshalling).

## Default to unexported

A good habit: start everything unexported, and export only what other packages genuinely need. Anything you export is a promise, because other code will start depending on it. Anything unexported you can rename, remove or rework tomorrow without breaking anybody.
