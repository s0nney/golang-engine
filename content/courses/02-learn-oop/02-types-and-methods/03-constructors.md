---
title: Constructor Functions
quiz:
  - question: What is `NewHero` in Go?
    options:
      - text: A special constructor the compiler calls automatically when you write `Hero{}`
      - text: An ordinary function that, by convention, builds and returns a ready-to-use `Hero`
        correct: true
      - text: A reserved keyword like `new`
      - text: A method that must be declared on every struct
    explanation: |
      Go has no constructors built into the language. `NewX` is a naming
      convention for a normal function. Nothing forces callers to use it, and
      `Hero{}` never calls it.
  - question: |
      What does this print?

      ```go
      func NewHero(name string) (*Hero, error) {
      	if name == "" {
      		return nil, errors.New("hero needs a name")
      	}
      	return &Hero{Name: name, HP: 100}, nil
      }

      func main() {
      	h, err := NewHero("")
      	fmt.Println(h == nil, err)
      }
      ```
    options:
      - text: '`false <nil>`'
      - text: '`true hero needs a name`'
        correct: true
      - text: It panics with a nil pointer dereference
      - text: '`true <nil>`'
    explanation: |
      The constructor validates its input and returns a nil pointer plus an error.
      Printing a nil pointer comparison is fine; only dereferencing it (e.g. `h.Name`)
      would panic.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    type Hero struct {
    	Name  string
    	Class string
    	HP    int
    	MaxHP int
    }

    var classHP = map[string]int{"warrior": 150, "archer": 100, "mage": 80}

    // NewHero builds a hero at full health for the given class.
    // It returns an error if the name is empty or the class is unknown.
    func NewHero(name, class string) (*Hero, error) {
    	// ?
    	return nil, errors.New("not implemented")
    }

    func main() {
    	for _, args := range [][2]string{{"Aria", "archer"}, {"", "mage"}, {"Borin", "bard"}} {
    		h, err := NewHero(args[0], args[1])
    		if err != nil {
    			fmt.Println("error:", err)
    			continue
    		}
    		fmt.Printf("%+v\n", *h)
    	}
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    type Hero struct {
    	Name  string
    	Class string
    	HP    int
    	MaxHP int
    }

    var classHP = map[string]int{"warrior": 150, "archer": 100, "mage": 80}

    func NewHero(name, class string) (*Hero, error) {
    	if name == "" {
    		return nil, errors.New("hero needs a name")
    	}
    	hp, ok := classHP[class]
    	if !ok {
    		return nil, fmt.Errorf("unknown class %q", class)
    	}
    	return &Hero{Name: name, Class: class, HP: hp, MaxHP: hp}, nil
    }

    func main() {
    	for _, args := range [][2]string{{"Aria", "archer"}, {"", "mage"}, {"Borin", "bard"}} {
    		h, err := NewHero(args[0], args[1])
    		if err != nil {
    			fmt.Println("error:", err)
    			continue
    		}
    		fmt.Printf("%+v\n", *h)
    	}
    }
  tests: |
    package main

    import "testing"

    func TestNewHeroValid(t *testing.T) {
    	for class, hp := range map[string]int{"warrior": 150, "archer": 100, "mage": 80} {
    		h, err := NewHero("Aria", class)
    		if err != nil {
    			t.Fatalf("NewHero(%q, %q) returned error %v, want nil", "Aria", class, err)
    		}
    		if h == nil {
    			t.Fatalf("NewHero(%q, %q) returned a nil hero", "Aria", class)
    		}
    		want := Hero{Name: "Aria", Class: class, HP: hp, MaxHP: hp}
    		if *h != want {
    			t.Errorf("NewHero(%q, %q) = %+v, want %+v", "Aria", class, *h, want)
    		}
    	}
    }

    func TestNewHeroInvalid(t *testing.T) {
    	for _, args := range [][2]string{{"", "mage"}, {"Borin", "bard"}, {"Cyra", ""}} {
    		h, err := NewHero(args[0], args[1])
    		if err == nil {
    			t.Errorf("NewHero(%q, %q) returned no error, want one", args[0], args[1])
    		}
    		if h != nil {
    			t.Errorf("NewHero(%q, %q) returned hero %+v, want nil on error", args[0], args[1], *h)
    		}
    	}
    }
---

In Python, a class's `__init__` method runs automatically every time you create an object. It sets defaults and checks arguments. Go has nothing like that. `Hero{}` always just gives you a struct with the fields you listed and zero values for the rest.

So how do Go programmers make sure a hero starts life correctly? With a plain function, named `New` something.

## The `NewX` convention

```go
func NewHero(name string) *Hero {
	return &Hero{
		Name:   name,
		HP:     100,
		MaxHP:  100,
		Attack: 10,
	}
}
```

It's just a function. There's nothing magic about the name, but every Go programmer who sees `NewHero` knows what it does. It returns a pointer here because heroes will be changed by methods (taking damage, levelling up), and we want everyone to share the same hero rather than copies.

## Validating input

Because it's an ordinary function, a constructor can return an error when the input doesn't make sense:

```go
package main

import (
	"errors"
	"fmt"
)

type Hero struct {
	Name   string
	HP     int
	MaxHP  int
	Attack int
}

func NewHero(name string, maxHP int) (*Hero, error) {
	if name == "" {
		return nil, errors.New("hero needs a name")
	}
	if maxHP <= 0 {
		return nil, fmt.Errorf("max HP must be positive, got %d", maxHP)
	}
	return &Hero{Name: name, HP: maxHP, MaxHP: maxHP, Attack: 10}, nil
}

func main() {
	aria, err := NewHero("Aria", 120)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%+v\n", *aria)

	_, err = NewHero("Ghost", 0)
	fmt.Println("error:", err)
}
```

```
{Name:Aria HP:120 MaxHP:120 Attack:10}
error: max HP must be positive, got 0
```

Python would `raise ValueError` inside `__init__`. Go returns an error value, the same pattern you've used everywhere else.

## Naming

- In a package that holds several types, use `NewHero`, `NewDragon`, `NewArcher`.
- In a package built around one type, just `New`: `hero.New("Aria")` reads nicely and avoids stutter.
- Standard library examples: `errors.New`, `bytes.NewBuffer`, `bufio.NewReader`, `strings.NewReplacer`.

## Constructors are optional

Here's a Go twist: a constructor is a *convenience*, not a requirement. Anyone in the package can still write `Hero{}` directly and skip your validation. The next chapter shows how unexported fields stop code in *other* packages from doing that.

And many Go types don't need a constructor at all, because their zero value is already ready to use. That's a big Go idea, and it's coming up at the end of this chapter.

## Go 1.26's `new(expr)`

You'll sometimes want a pointer to a simple value, say an optional starting level. Since Go 1.26 the built-in `new` accepts an expression:

```go
level := new(5) // *int pointing at 5
```

Before 1.26 you'd have needed a temporary variable (`l := 5; level := &l`). It's a small thing, but handy when filling in optional pointer fields in a struct literal.

## Assignment

Complete `NewHero(name, class string) (*Hero, error)`:

- If `name` is empty, return `nil` and an error.
- Look up the class in the `classHP` map. If it isn't there, return `nil` and an error.
- Otherwise return a hero with that name and class, at full health (`HP` and `MaxHP` both set to the class's HP) and a `nil` error.

## Further reading

- [Effective Go: Constructors and composite literals](https://go.dev/doc/effective_go#composite_literals)
