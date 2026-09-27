---
title: Type Assertions and Type Switches
quiz:
  - question: |
      What happens here?

      ```go
      var c Combatant = &Hero{Name: "Aria"}
      d := c.(*Dragon)
      fmt.Println(d)
      ```
    options:
      - text: It prints `<nil>`
      - text: It panics, because the single-result form of a type assertion panics when the type doesn't match
        correct: true
      - text: It fails to compile
      - text: It converts the hero into a dragon
    explanation: |
      `c.(*Dragon)` with one result is a promise: "I'm sure this is a `*Dragon`".
      If you're wrong, the program panics. Use the comma-ok form, `d, ok := c.(*Dragon)`,
      to check safely.
  - question: |
      What does this print?

      ```go
      func describe(v any) string {
      	switch x := v.(type) {
      	case int:
      		return fmt.Sprint("gold: ", x)
      	case string, fmt.Stringer:
      		return "named"
      	case nil:
      		return "nothing"
      	default:
      		return "unknown"
      	}
      }

      fmt.Println(describe(42), describe("Ember"), describe(nil), describe(3.5))
      ```
    options:
      - text: '`gold: 42 named nothing unknown`'
        correct: true
      - text: '`gold: 42 named unknown unknown`'
      - text: '`unknown named nothing unknown`'
      - text: It doesn't compile because a case lists two types
    explanation: |
      Cases can list several types, and `case nil` matches a nil interface. A
      `float64` matches nothing else, so it hits `default`.
exercise:
  starter: |
    package main

    import "fmt"

    type Hero struct{ Name string }

    type Dragon struct {
    	Name     string
    	Treasure int
    }

    type Chest struct{ Gold int }

    // inspect describes whatever the player is looking at:
    //
    //	Hero      -> "a hero named <Name>"
    //	Dragon    -> "a dragon guarding <Treasure> gold"
    //	Chest     -> "a chest with <Gold> gold", or "an empty chest" if Gold is 0
    //	nil       -> "nothing at all"
    //	otherwise -> "something strange"
    func inspect(thing any) string {
    	// ?
    	return "?"
    }

    func main() {
    	things := []any{Hero{Name: "Aria"}, Dragon{Name: "Ember", Treasure: 5000}, Chest{}, Chest{Gold: 12}, nil, 42}
    	for _, th := range things {
    		fmt.Println(inspect(th))
    	}
    }
  solution: |
    package main

    import "fmt"

    type Hero struct{ Name string }

    type Dragon struct {
    	Name     string
    	Treasure int
    }

    type Chest struct{ Gold int }

    func inspect(thing any) string {
    	switch t := thing.(type) {
    	case Hero:
    		return "a hero named " + t.Name
    	case Dragon:
    		return fmt.Sprintf("a dragon guarding %d gold", t.Treasure)
    	case Chest:
    		if t.Gold == 0 {
    			return "an empty chest"
    		}
    		return fmt.Sprintf("a chest with %d gold", t.Gold)
    	case nil:
    		return "nothing at all"
    	default:
    		return "something strange"
    	}
    }

    func main() {
    	things := []any{Hero{Name: "Aria"}, Dragon{Name: "Ember", Treasure: 5000}, Chest{}, Chest{Gold: 12}, nil, 42}
    	for _, th := range things {
    		fmt.Println(inspect(th))
    	}
    }
  tests: |
    package main

    import "testing"

    func TestInspect(t *testing.T) {
    	for _, tt := range []struct {
    		thing any
    		want  string
    	}{
    		{Hero{Name: "Aria"}, "a hero named Aria"},
    		{Hero{Name: "Borin"}, "a hero named Borin"},
    		{Dragon{Name: "Ember", Treasure: 5000}, "a dragon guarding 5000 gold"},
    		{Chest{}, "an empty chest"},
    		{Chest{Gold: 12}, "a chest with 12 gold"},
    		{nil, "nothing at all"},
    		{42, "something strange"},
    		{&Hero{Name: "Aria"}, "something strange"},
    	} {
    		if got := inspect(tt.thing); got != tt.want {
    			t.Errorf("inspect(%#v) = %q, want %q", tt.thing, got, tt.want)
    		}
    	}
    }
---

Interfaces hide the concrete type, which is usually exactly what you want. But now and then you need to ask: "what's *actually* in here?" Go gives you two tools for that.

## Type assertions

A **type assertion** extracts the concrete value from an interface:

```go
var c Combatant = &Dragon{name: "Ember", hp: 300}

d := c.(*Dragon) // "I'm sure c holds a *Dragon"
fmt.Println(d.hp)
```

If you're wrong, this **panics**. So most of the time you use the **comma-ok** form, which never panics:

```go
if d, ok := c.(*Dragon); ok {
	fmt.Println("It's a dragon with", d.hp, "HP")
} else {
	fmt.Println("Not a dragon")
}
```

You can also assert to another **interface**, which asks "does the value inside *also* have these methods?". This is a powerful way to support optional behaviour:

```go
type Flyer interface{ Fly() string }

if f, ok := c.(Flyer); ok {
	fmt.Println(f.Fly())
}
```

The standard library does this all the time. `fmt` checks whether your value is a `fmt.Stringer`; `io.Copy` checks whether the source is an `io.WriterTo`, so it can take a faster path.

## Type switches

When there are several possibilities, a **type switch** is cleaner than a chain of assertions:

```go
package main

import "fmt"

type Hero struct{ Name string }
type Dragon struct {
	Name     string
	Treasure int
}
type Chest struct{ Gold int }

func (d Dragon) Fly() string { return d.Name + " takes to the skies" }

type Flyer interface{ Fly() string }

func inspect(thing any) string {
	switch t := thing.(type) {
	case Hero:
		return "a hero named " + t.Name
	case Dragon:
		return fmt.Sprintf("a dragon guarding %d gold", t.Treasure)
	case Chest:
		if t.Gold == 0 {
			return "an empty chest"
		}
		return fmt.Sprintf("a chest with %d gold", t.Gold)
	case nil:
		return "nothing at all"
	default:
		return fmt.Sprintf("something strange (%T)", t)
	}
}

func main() {
	things := []any{
		Hero{Name: "Aria"},
		Dragon{Name: "Ember", Treasure: 5000},
		Chest{},
		nil,
		42,
	}
	for _, th := range things {
		fmt.Println(inspect(th))
	}

	var f Flyer = Dragon{Name: "Ember"}
	fmt.Println(f.Fly())
}
```

```
a hero named Aria
a dragon guarding 5000 gold
an empty chest
nothing at all
something strange (int)
Ember takes to the skies
```

Inside each `case`, `t` already has that case's concrete type, so `t.Treasure` just works in the `Dragon` branch. In `default` (or a case listing several types), `t` keeps the original interface type.

## Don't let switches replace polymorphism

Here's the trap. It's tempting to write:

```go
func attack(c Combatant) int {
	switch x := c.(type) {
	case *Hero:
		return x.strength * 2
	case *Dragon:
		return 50 + x.age
	}
	return 0
}
```

This throws away everything the last lesson gave you. Every new creature means editing `attack`, and forgetting one silently returns 0. Put an `Attack()` method on each type instead and let dynamic dispatch pick.

Good uses of type assertions and switches:

- **Optional behaviour**: "if this also happens to be a `Flyer`, let it fly".
- **Handling `any`** from things like JSON decoding, where you truly don't know the type.
- **Inspecting errors**, which is so common Go has a dedicated helper for it. That's next.

## Assignment

Complete `inspect(thing any) string` using a type switch, so it describes each thing exactly as its comment says. Watch out for two details: a chest with 0 gold is `"an empty chest"`, and a `*Hero` (a pointer) isn't a `Hero`, so it counts as `"something strange"`.

## Further reading

- [A Tour of Go: Type assertions](https://go.dev/tour/methods/15) and [Type switches](https://go.dev/tour/methods/16)
- [Effective Go: Type switch](https://go.dev/doc/effective_go#type_switch)
