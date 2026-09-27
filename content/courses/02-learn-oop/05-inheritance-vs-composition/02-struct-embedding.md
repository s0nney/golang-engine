---
title: Struct Embedding
quiz:
  - question: |
      What does this print?

      ```go
      type Stats struct {
      	HP, Attack int
      }

      type Goblin struct {
      	Stats
      	Name string
      }

      func main() {
      	g := Goblin{Stats: Stats{HP: 30, Attack: 4}, Name: "Snik"}
      	g.HP -= 10
      	fmt.Println(g.HP, g.Stats.HP)
      }
      ```
    options:
      - text: '`20 30`'
      - text: '`20 20`'
        correct: true
      - text: '`30 20`'
      - text: It doesn't compile; you must write `g.Stats.HP`
    explanation: |
      `g.HP` is just a shortcut for `g.Stats.HP`. They are the same field, so
      both print 20.
  - question: |
      With `type Goblin struct { Stats; Name string }` in Go 1.27, which composite literal does NOT compile?
    options:
      - text: '`Goblin{HP: 30, Name: "Snik"}`'
      - text: '`Goblin{Stats: Stats{HP: 30}, Name: "Snik"}`'
      - text: '`Goblin{Stats: Stats{Attack: 4}, HP: 30}`'
        correct: true
    explanation: |
      Since Go 1.27, promoted fields can be used as keys, so the first literal is
      fine, and the long form always worked. But you can't set the whole embedded
      `Stats` *and* one of its fields in the same literal: the keys overlap, and
      the compiler reports "cannot specify promoted field HP and enclosing embedded
      field Stats".
---

**Embedding** is Go's tool for reusing fields and methods. You include a type inside a struct **without giving it a field name**:

```go
type Stats struct {
	HP     int
	Attack int
}

type Hero struct {
	Stats // embedded: no field name, just the type
	Name  string
}
```

`Stats` is still a field of `Hero`. Its name is simply the type's name, `Stats`. But the fields *inside* `Stats` are **promoted**: you can reach them as if they belonged to `Hero` directly.

## Promoted fields

```go
package main

import "fmt"

type Stats struct {
	HP     int
	Attack int
}

type Hero struct {
	Stats
	Name string
}

type Dragon struct {
	Stats
	Name    string
	Element string
}

func main() {
	aria := Hero{
		Stats: Stats{HP: 100, Attack: 18},
		Name:  "Aria",
	}
	ember := Dragon{
		Stats:   Stats{HP: 300, Attack: 40},
		Name:    "Ember",
		Element: "fire",
	}

	aria.HP -= ember.Attack    // promoted: same as aria.Stats.HP -= ember.Stats.Attack
	fmt.Println(aria.HP)       // 60
	fmt.Println(aria.Stats.HP) // 60, the very same field
	fmt.Printf("%+v\n", ember)
}
```

```
60
60
{Stats:{HP:300 Attack:40} Name:Ember Element:fire}
```

A few things to notice:

- `aria.HP` and `aria.Stats.HP` are **the same field**. Promotion is purely a naming shortcut.
- In a **composite literal**, the classic way is to initialise the embedded struct by its type name (`Stats: Stats{...}`), as above. New in Go 1.27, you may also use promoted fields as keys directly. More on that below.
- `%+v` reveals the truth: `Stats` is a real, named field inside `Dragon`.

## Go 1.27: promoted fields in literals

Before Go 1.27, `Hero{HP: 100, Name: "Aria"}` was a compile error, and everybody had to spell out `Stats: Stats{HP: 100}`. Go 1.27 relaxed the rule, so both of these now work and build the same value:

```go
a := Hero{Stats: Stats{HP: 100, Attack: 18}, Name: "Aria"}
b := Hero{HP: 100, Attack: 18, Name: "Aria"} // Go 1.27+
```

Two restrictions remain:

- You can't mix the two styles for the same embedded struct. `Hero{Stats: Stats{Attack: 18}, HP: 100}` fails with *cannot specify promoted field HP and enclosing embedded field Stats*.
- You can't reach through an embedded **pointer**. If `Hero` embedded `*Stats`, then `Hero{HP: 100}` would fail with *invalid implicit pointer indirection*, because there's no `Stats` value yet for `HP` to live in.

If your `go.mod` says `go 1.26` or older, the short form is rejected with a message telling you it requires go1.27. You'll see plenty of older code using the long form, and that's still perfectly good Go.

## It's composition, not inheritance

This looks like inheritance, and you'll hear people call it that. It isn't. Under the hood, `Hero` simply **contains** a `Stats` value. There's no parent class and no "is-a" relationship:

```go
var s Stats = aria       // compile error: cannot use aria (variable of struct type Hero) as Stats value
var s Stats = aria.Stats // fine: take the part out explicitly
```

In Python, an `Archer` *is* a `Hero` and can go anywhere a `Hero` can. In Go, a `Hero` *has* `Stats`. If you want to hand the stats to a function, you pass `aria.Stats`.

## Embedding pointers

You can also embed a pointer type:

```go
type Mount struct{ Speed int }

type Knight struct {
	*Mount
	Name string
}
```

Now `k.Speed` goes through the pointer. Several knights could share one `Mount` (a two-seater griffin, perhaps). But if `Mount` is `nil`, then `k.Speed` panics with a nil pointer dereference, so you need to be sure it's set. Embedding a plain value is simpler and safer unless you genuinely need sharing.

## Name clashes

If the outer struct has a field with the same name as a promoted one, the **outer one wins**, and the inner one is still reachable by its full path:

```go
type Dragon struct {
	Stats
	HP int // shadows Stats.HP
}

// d.HP is Dragon.HP; d.Stats.HP is the other one.
```

That's legal, but it's almost always a bug waiting to happen. Two fields both called `HP` is exactly the kind of confusion clean code avoids. The same shadowing rule applies to methods, and there it turns out to be genuinely useful. That's next.

## Further reading

- [Go by Example: Struct Embedding](https://gobyexample.com/struct-embedding)
- [Effective Go: Embedding](https://go.dev/doc/effective_go#embedding)
