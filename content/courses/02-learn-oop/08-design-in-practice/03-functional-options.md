---
title: Functional Options
quiz:
  - question: |
      What does this print?

      ```go
      type Dragon struct {
      	Name string
      	HP   int
      	Fire bool
      }

      type Option func(*Dragon)

      func WithHP(hp int) Option { return func(d *Dragon) { d.HP = hp } }
      func Breathing() Option    { return func(d *Dragon) { d.Fire = true } }

      func NewDragon(name string, opts ...Option) *Dragon {
      	d := &Dragon{Name: name, HP: 100}
      	for _, opt := range opts {
      		opt(d)
      	}
      	return d
      }

      func main() {
      	d := NewDragon("Ember", WithHP(300), Breathing(), WithHP(250))
      	fmt.Println(d.HP, d.Fire)
      }
      ```
    options:
      - text: '`100 true`'
      - text: '`300 true`'
      - text: '`250 true`'
        correct: true
      - text: '`250 false`'
    explanation: |
      Options run in order, so the later `WithHP(250)` overwrites the earlier 300.
      `Breathing()` sets `Fire`. Defaults are applied first, then options.
  - question: What's the main advantage of functional options over a long list of constructor parameters?
    options:
      - text: They make the program run faster
      - text: Callers only mention the settings they care about, defaults stay sensible, and new options can be added without breaking existing calls
        correct: true
      - text: They remove the need for a constructor
      - text: They let you skip validation
    explanation: |
      `NewDragon("Ember")` still works after you add ten new options. With
      positional parameters, every new setting would break every existing call site.
exercise:
  starter: |
    package main

    import "fmt"

    type Dragon struct {
    	name     string
    	hp       int
    	element  string
    	canFly   bool
    	treasure int
    }

    type Option func(*Dragon)

    // WithHP sets the dragon's HP.
    func WithHP(hp int) Option {
    	return func(d *Dragon) {} // ?
    }

    // WithElement sets the dragon's element.
    func WithElement(e string) Option {
    	return func(d *Dragon) {} // ?
    }

    // Grounded makes the dragon unable to fly.
    func Grounded() Option {
    	return func(d *Dragon) {} // ?
    }

    // WithHoard ADDS gold to the dragon's treasure, so it can be used more than once.
    func WithHoard(gold int) Option {
    	return func(d *Dragon) {} // ?
    }

    // NewDragon creates a dragon with defaults (200 HP, fire, can fly, no treasure),
    // then applies the options in order.
    func NewDragon(name string, opts ...Option) *Dragon {
    	d := &Dragon{name: name}
    	// ?
    	return d
    }

    func main() {
    	fmt.Printf("%+v\n", *NewDragon("Whelp"))
    	fmt.Printf("%+v\n", *NewDragon("Frost", WithElement("ice"), WithHP(450)))
    	fmt.Printf("%+v\n", *NewDragon("Wyrm", Grounded(), WithHoard(1000), WithHoard(500)))
    }
  solution: |
    package main

    import "fmt"

    type Dragon struct {
    	name     string
    	hp       int
    	element  string
    	canFly   bool
    	treasure int
    }

    type Option func(*Dragon)

    func WithHP(hp int) Option {
    	return func(d *Dragon) { d.hp = hp }
    }

    func WithElement(e string) Option {
    	return func(d *Dragon) { d.element = e }
    }

    func Grounded() Option {
    	return func(d *Dragon) { d.canFly = false }
    }

    func WithHoard(gold int) Option {
    	return func(d *Dragon) { d.treasure += gold }
    }

    func NewDragon(name string, opts ...Option) *Dragon {
    	d := &Dragon{name: name, hp: 200, element: "fire", canFly: true}
    	for _, opt := range opts {
    		opt(d)
    	}
    	return d
    }

    func main() {
    	fmt.Printf("%+v\n", *NewDragon("Whelp"))
    	fmt.Printf("%+v\n", *NewDragon("Frost", WithElement("ice"), WithHP(450)))
    	fmt.Printf("%+v\n", *NewDragon("Wyrm", Grounded(), WithHoard(1000), WithHoard(500)))
    }
  tests: |
    package main

    import "testing"

    func TestDefaults(t *testing.T) {
    	got := *NewDragon("Whelp")
    	want := Dragon{name: "Whelp", hp: 200, element: "fire", canFly: true}
    	if got != want {
    		t.Errorf("NewDragon(\"Whelp\") = %+v, want %+v", got, want)
    	}
    }

    func TestOptions(t *testing.T) {
    	for _, tt := range []struct {
    		desc string
    		opts []Option
    		want Dragon
    	}{
    		{"WithHP(450)", []Option{WithHP(450)}, Dragon{name: "D", hp: 450, element: "fire", canFly: true}},
    		{"WithElement(\"ice\")", []Option{WithElement("ice")}, Dragon{name: "D", hp: 200, element: "ice", canFly: true}},
    		{"Grounded()", []Option{Grounded()}, Dragon{name: "D", hp: 200, element: "fire", canFly: false}},
    		{"WithHoard(1000), WithHoard(500)", []Option{WithHoard(1000), WithHoard(500)}, Dragon{name: "D", hp: 200, element: "fire", canFly: true, treasure: 1500}},
    		{"WithHP(300), WithHP(250)", []Option{WithHP(300), WithHP(250)}, Dragon{name: "D", hp: 250, element: "fire", canFly: true}},
    	} {
    		if got := *NewDragon("D", tt.opts...); got != tt.want {
    			t.Errorf("NewDragon(\"D\", %s) = %+v, want %+v", tt.desc, got, tt.want)
    		}
    	}
    }
---

Dragons in our game are getting complicated. They have a name, HP, an element, wings or not, a treasure hoard, an aggression level... Most dragons use sensible defaults, and each call site only wants to tweak one or two things. How should `NewDragon` look?

## The options that don't scale

**Lots of parameters:**

```go
NewDragon("Ember", 300, "fire", true, 5000, 7)
```

What's `true`? What's `7`? And adding a new setting breaks every call in the codebase.

**A config struct** is better, and a perfectly good choice in many cases:

```go
NewDragon(DragonConfig{Name: "Ember", HP: 300, Element: "fire"})
```

But the zero value of each field is ambiguous. Did the caller mean "0 HP", or "use the default"?

## Functional options

This pattern, popularised by Dave Cheney and Rob Pike, uses a **function type** as the option:

```go
type Option func(*Dragon)
```

Each option is a function that tweaks a dragon under construction. The constructor sets defaults, then applies whatever options it's given, in order:

```go
package main

import "fmt"

type Dragon struct {
	name     string
	hp       int
	element  string
	canFly   bool
	treasure int
}

type Option func(*Dragon)

func WithHP(hp int) Option {
	return func(d *Dragon) { d.hp = hp }
}

func WithElement(e string) Option {
	return func(d *Dragon) { d.element = e }
}

func Grounded() Option {
	return func(d *Dragon) { d.canFly = false }
}

func WithHoard(gold int) Option {
	return func(d *Dragon) { d.treasure += gold }
}

func NewDragon(name string, opts ...Option) *Dragon {
	d := &Dragon{ // sensible defaults
		name:    name,
		hp:      200,
		element: "fire",
		canFly:  true,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

func (d *Dragon) String() string {
	wings := "flies"
	if !d.canFly {
		wings = "walks"
	}
	return fmt.Sprintf("%s: %d HP, %s, %s, hoard %d", d.name, d.hp, d.element, wings, d.treasure)
}

func main() {
	fmt.Println(NewDragon("Whelp"))
	fmt.Println(NewDragon("Frost", WithElement("ice"), WithHP(450)))
	fmt.Println(NewDragon("Wyrm", Grounded(), WithHoard(1000), WithHoard(500)))
}
```

```
Whelp: 200 HP, fire, flies, hoard 0
Frost: 450 HP, ice, flies, hoard 0
Wyrm: 200 HP, fire, walks, hoard 1500
```

Look at the call sites. Each reads like a sentence and mentions only what's different from the default. `NewDragon("Whelp")` needs nothing at all.

## Why it's nice

- **Self-documenting.** `WithHP(450)` says what 450 means.
- **Good defaults.** Anything not mentioned keeps its default. No zero-value ambiguity.
- **Backwards compatible.** Add `WithAggression(n)` next month and every existing call still compiles.
- **Encapsulation intact.** The fields stay unexported. The *only* ways to configure a dragon are the options you choose to export.
- **Options can do more than set a field.** `WithHoard` *adds*, and an option could validate, or set several fields at once.

## Options that can fail

If some options need validation, make the option return an error, and have the constructor collect it:

```go
type Option func(*Dragon) error

func WithHP(hp int) Option {
	return func(d *Dragon) error {
		if hp <= 0 {
			return fmt.Errorf("dragon HP must be positive, got %d", hp)
		}
		d.hp = hp
		return nil
	}
}

func NewDragon(name string, opts ...Option) (*Dragon, error) {
	d := &Dragon{name: name, hp: 200, element: "fire", canFly: true}
	for _, opt := range opts {
		if err := opt(d); err != nil {
			return nil, err
		}
	}
	return d, nil
}
```

## When not to bother

Functional options add a function per setting. For a type with two or three settings, plain parameters or a small config struct are simpler. Reach for this pattern when a constructor has many optional settings, most callers use few of them, and the type is part of a package others depend on.

## Assignment

Finish the dragon builder:

- Make `NewDragon` start from the defaults (200 HP, `"fire"`, can fly, no treasure) and then apply every option in order.
- Make each option do what its comment says. `WithHoard` **adds** gold, so two hoards stack.

## Further reading

- [Functional options for friendly APIs](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis) by Dave Cheney
