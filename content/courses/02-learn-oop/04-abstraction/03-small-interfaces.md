---
title: Small Interfaces
quiz:
  - question: |
      What does this print?

      ```go
      type Dragon struct {
      	Name string
      	HP   int
      }

      func (d Dragon) String() string {
      	return fmt.Sprintf("Dragon %s [%d HP]", d.Name, d.HP)
      }

      func main() {
      	fmt.Println(Dragon{Name: "Ember", HP: 90})
      }
      ```
    options:
      - text: '`{Ember 90}`'
      - text: '`Dragon Ember [90 HP]`'
        correct: true
      - text: '`Dragon{Name:"Ember", HP:90}`'
      - text: It doesn't compile unless you call `.String()` yourself
    explanation: |
      `Dragon` has a `String() string` method, so it satisfies `fmt.Stringer`.
      The `fmt` printing functions check for that interface and use your method.
  - question: How many methods does `io.Reader` have?
    options:
      - text: One, `Read(p []byte) (n int, err error)`
        correct: true
      - text: Two, `Open` and `Read`
      - text: Three, `Open`, `Read` and `Close`
      - text: None; it's an empty interface
    explanation: |
      `io.Reader` has a single method. That tiny contract is why files, network
      connections, strings, gzip streams and HTTP bodies can all be read by the same
      code.
exercise:
  starter: |
    package main

    import "fmt"

    type Element int

    const (
    	Fire Element = iota
    	Ice
    	Storm
    )

    // TODO: give Element a String method:
    // Fire -> "fire", Ice -> "ice", Storm -> "storm", anything else -> "Element(<n>)".

    type Dragon struct {
    	Name    string
    	Element Element
    	HP      int
    }

    // TODO: give Dragon a String method, e.g. "Ember the fire dragon [300 HP]".

    func main() {
    	fmt.Println(Ice, Element(7))
    	fmt.Println(Dragon{Name: "Ember", Element: Fire, HP: 300})
    	fmt.Printf("%v vs %v\n", Dragon{Name: "Frost", Element: Ice, HP: 250}, Dragon{Name: "Volt", Element: Storm, HP: 90})
    }
  solution: |
    package main

    import "fmt"

    type Element int

    const (
    	Fire Element = iota
    	Ice
    	Storm
    )

    func (e Element) String() string {
    	switch e {
    	case Fire:
    		return "fire"
    	case Ice:
    		return "ice"
    	case Storm:
    		return "storm"
    	}
    	return fmt.Sprintf("Element(%d)", int(e))
    }

    type Dragon struct {
    	Name    string
    	Element Element
    	HP      int
    }

    func (d Dragon) String() string {
    	return fmt.Sprintf("%s the %v dragon [%d HP]", d.Name, d.Element, d.HP)
    }

    func main() {
    	fmt.Println(Ice, Element(7))
    	fmt.Println(Dragon{Name: "Ember", Element: Fire, HP: 300})
    	fmt.Printf("%v vs %v\n", Dragon{Name: "Frost", Element: Ice, HP: 250}, Dragon{Name: "Volt", Element: Storm, HP: 90})
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    )

    func TestElementString(t *testing.T) {
    	if _, ok := any(Fire).(fmt.Stringer); !ok {
    		t.Fatal("Element does not implement fmt.Stringer: add a String() string method")
    	}
    	for _, tt := range []struct {
    		e    Element
    		want string
    	}{{Fire, "fire"}, {Ice, "ice"}, {Storm, "storm"}, {Element(7), "Element(7)"}, {Element(-1), "Element(-1)"}} {
    		if got := fmt.Sprint(tt.e); got != tt.want {
    			t.Errorf("fmt.Sprint(Element(%d)) = %q, want %q", int(tt.e), got, tt.want)
    		}
    	}
    }

    func TestDragonString(t *testing.T) {
    	d := Dragon{Name: "Ember", Element: Fire, HP: 300}
    	if _, ok := any(d).(fmt.Stringer); !ok {
    		t.Fatal("Dragon does not implement fmt.Stringer: add a String() string method (with a value receiver, so Dragon values print nicely)")
    	}
    	for _, tt := range []struct {
    		d    Dragon
    		want string
    	}{
    		{d, "Ember the fire dragon [300 HP]"},
    		{Dragon{Name: "Frost", Element: Ice, HP: 250}, "Frost the ice dragon [250 HP]"},
    		{Dragon{Name: "Odd", Element: Element(9), HP: 1}, "Odd the Element(9) dragon [1 HP]"},
    	} {
    		if got := fmt.Sprint(tt.d); got != tt.want {
    			t.Errorf("fmt.Sprint(%#v) = %q, want %q", tt.d, got, tt.want)
    		}
    	}
    }
---

Some of the most powerful abstractions in Go are **tiny**: one method, sometimes two. Let's meet the two most famous ones and see why small wins.

## `fmt.Stringer`

The `fmt` package declares:

```go
type Stringer interface {
	String() string
}
```

Any type with a `String() string` method satisfies it, and `fmt.Println`, `fmt.Printf("%v")` and friends will use that method to print your value:

```go
package main

import "fmt"

type Element int

const (
	Fire Element = iota
	Ice
	Storm
)

func (e Element) String() string {
	switch e {
	case Fire:
		return "fire"
	case Ice:
		return "ice"
	case Storm:
		return "storm"
	}
	return fmt.Sprintf("Element(%d)", int(e))
}

type Dragon struct {
	Name    string
	Element Element
}

func (d Dragon) String() string {
	return fmt.Sprintf("%s the %v dragon", d.Name, d.Element)
}

func main() {
	fmt.Println(Ice)
	fmt.Println(Dragon{Name: "Ember", Element: Fire})
	fmt.Printf("%v and %v\n", Dragon{Name: "Frost", Element: Ice}, Element(9))
}
```

```
ice
Ember the fire dragon
Frost the ice dragon and Element(9)
```

`fmt` never heard of our dragons, yet prints them nicely. Notice `Element(9)`: an unknown element falls through the `switch` to a sensible default instead of printing nothing. One method, and our types plug into the whole formatting system.

## `io.Reader`

The `io` package declares:

```go
type Reader interface {
	Read(p []byte) (n int, err error)
}
```

That's it. "Fill this byte slice with some data and tell me how much you wrote." Because the contract is so small, *everything* implements it: `*os.File`, `*strings.Reader`, `*bytes.Buffer`, network connections, HTTP response bodies, gzip decompressors...

So a function that loads a saved game can accept an `io.Reader`:

```go
func LoadSave(r io.Reader) (*Game, error)
```

In production you pass it a file. In tests you pass `strings.NewReader("...")`. Later, you pass it a network connection for cloud saves. `LoadSave` never changes.

## Composing small interfaces

Small interfaces combine nicely. The `io` package builds bigger ones by embedding smaller ones:

```go
type ReadWriter interface {
	Reader
	Writer
}
```

A function that only reads asks for a `Reader`. One that needs both asks for a `ReadWriter`. Each function asks for **exactly** what it needs and no more.

## Designing our own small interfaces

For the RPG, instead of one giant `Character` interface, think in single behaviours:

```go
type Attacker interface{ Attack() int }
type Healer interface{ Heal(target *Hero) }
type Flyer interface{ Fly(to Location) }
```

A cleric is an `Attacker` and a `Healer`. A dragon is an `Attacker` and a `Flyer`. A shop's healing fountain is only a `Healer`. The naming convention is the method name plus `-er`, just like `Reader`, `Writer` and `Stringer`.

The next lesson makes the case for small interfaces explicit, with a Go proverb and a cautionary tale.

## Assignment

Right now `main` prints dragons as `{Ember 0 300}`, which tells the player nothing. Plug the types into `fmt` by satisfying `fmt.Stringer`:

- Give `Element` a `String() string` method: `Fire` is `"fire"`, `Ice` is `"ice"`, `Storm` is `"storm"`, and any other value is `"Element(<n>)"`, such as `"Element(7)"`.
- Give `Dragon` a `String() string` method that returns, for example, `"Ember the fire dragon [300 HP]"`.

Watch out for one trap. Inside `Element.String`, formatting `e` itself with `%v` or `%s` would call `String` again, forever. Convert it to a plain `int(e)` first, just like the lesson does. Inside `Dragon.String`, formatting `d.Element` with `%v` is fine, and it reuses the method you just wrote.

## Further reading

- [A Tour of Go: Stringers](https://go.dev/tour/methods/17) and [Readers](https://go.dev/tour/methods/21)
