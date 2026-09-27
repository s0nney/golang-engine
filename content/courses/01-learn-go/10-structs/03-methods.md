---
title: Methods
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      type message struct {
      	body string
      }

      func (m message) segments() int {
      	return (len(m.body) + 159) / 160
      }

      func main() {
      	m := message{body: "hi"}
      	fmt.Println(m.segments(), message{}.segments())
      }
      ```
    options:
      - text: '`1 1`'
      - text: '`2 0`'
      - text: '`1 0`'
        correct: true
      - text: It doesn't compile, because `message{}` has no body
    explanation: |
      For `"hi"`, `(2 + 159) / 160` is `161 / 160`, which is `1` in integer
      division. The empty message has length 0, and `159 / 160` is `0`.
  - question: In `func (m message) segments() int`, what is `(m message)` called?
    options:
      - text: The receiver
        correct: true
      - text: The return type
      - text: The package name
      - text: A type parameter
    explanation: |
      The receiver goes between `func` and the method name. It says which
      type the method belongs to, and gives a name (`m`) for the value the
      method was called on.
exercise:
  starter: |
    package main

    import "fmt"

    type message struct {
    	to   string
    	body string
    }

    // segments returns how many 160-character SMS segments the body needs.
    // An empty body needs 0 segments.
    func (m message) segments() int {
    	// ?
    	return 0
    }

    // cost returns the price in cents: 2 cents per segment, doubled
    // for international numbers (those that don't start with "+1").
    func (m message) cost() int {
    	// ?
    	return 0
    }

    func main() {
    	m := message{to: "+44-20-7946-0000", body: "Your code is 4821"}
    	fmt.Println(m.segments(), m.cost()) // want 1 4
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    type message struct {
    	to   string
    	body string
    }

    func (m message) segments() int {
    	return (len(m.body) + 159) / 160
    }

    func (m message) cost() int {
    	c := m.segments() * 2
    	if !strings.HasPrefix(m.to, "+1") {
    		c *= 2
    	}
    	return c
    }

    func main() {
    	m := message{to: "+44-20-7946-0000", body: "Your code is 4821"}
    	fmt.Println(m.segments(), m.cost())
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    )

    func TestMessage(t *testing.T) {
    	tests := []struct {
    		m            message
    		wantSegments int
    		wantCost     int
    	}{
    		{message{"+1-555-0100", "hi"}, 1, 2},
    		{message{"+44-20-7946-0000", "Your code is 4821"}, 1, 4},
    		{message{"+1-555-0100", ""}, 0, 0},
    		{message{"+1-555-0100", strings.Repeat("a", 160)}, 1, 2},
    		{message{"+1-555-0100", strings.Repeat("a", 161)}, 2, 4},
    		{message{"+33-1-23-45-67-89", strings.Repeat("a", 400)}, 3, 12},
    	}
    	for _, tt := range tests {
    		if got := tt.m.segments(); got != tt.wantSegments {
    			t.Errorf("message{to: %q, body: %d chars}.segments() = %d, want %d", tt.m.to, len(tt.m.body), got, tt.wantSegments)
    		}
    		if got := tt.m.cost(); got != tt.wantCost {
    			t.Errorf("message{to: %q, body: %d chars}.cost() = %d, want %d", tt.m.to, len(tt.m.body), got, tt.wantCost)
    		}
    	}
    }
---

Structs hold data. **Methods** let you attach behaviour to that data, so a `message` can answer questions about itself, like "how much do I cost?"

## Declaring a method

A method is a function with a **receiver**: an extra parameter, written in parentheses **before** the function name, that says which type the method belongs to:

```go
package main

import "fmt"

type message struct {
	to   string
	body string
}

func (m message) cost() float64 {
	if len(m.body) > 160 {
		return 0.02
	}
	return 0.01
}

func main() {
	m := message{to: "Alice", body: "Your code is 4821"}
	fmt.Println(m.cost())
}
```

```text
0.01
```

- `(m message)` is the receiver. Inside the method, `m` is the message the method was called on.
- You call a method with a dot, just like reading a field: `m.cost()`.
- By convention the receiver name is short, often the first letter of the type: `m` for `message`, `u` for `user`. Don't use `this` or `self`.

A method is really just a function with a special first argument. Writing `m.cost()` is much like writing `cost(m)`, but it groups the behaviour with the type and reads more naturally.

## Methods can take parameters and return anything

```go
package main

import (
	"fmt"
	"strings"
)

type message struct {
	to   string
	body string
}

func (m message) preview(n int) string {
	if len(m.body) <= n {
		return m.body
	}
	return m.body[:n] + "..."
}

func (m message) isMarketing() bool {
	return strings.Contains(m.body, "SALE")
}

func main() {
	m := message{to: "Bob", body: "Big SALE this weekend, everything 50% off!"}
	fmt.Println(m.preview(10))
	fmt.Println(m.isMarketing())
}
```

```text
Big SALE t...
true
```

Slicing works on strings too: `m.body[:n]` is the first `n` bytes of the body. `strings.Contains` is from the `strings` package, which has loads of handy functions for working with text.

## Value receivers get a copy

With a receiver like `(m message)`, the method gets a **copy** of the struct, just like any function parameter. So a method like this does nothing useful:

```go
func (m message) markSent() {
	m.body = "[sent] " + m.body // changes the copy only!
}
```

To modify the original, you need a **pointer receiver**, which is the next chapter's main event.

## Methods on embedded structs are promoted

Remember embedding? Methods are promoted just like fields:

```go
package main

import "fmt"

type user struct {
	name string
}

func (u user) greeting() string {
	return "Hi " + u.name + "!"
}

type sender struct {
	user
	rateLimit int
}

func main() {
	s := sender{user: user{name: "Acme"}, rateLimit: 100}
	fmt.Println(s.greeting())
}
```

```text
Hi Acme!
```

`s.greeting()` is a shortcut for `s.user.greeting()`.

## Methods on any named type

Methods aren't only for structs. You can define methods on any type **you declare in your own package**:

```go
type credits int

func (c credits) canAfford(cost credits) bool {
	return c >= cost
}
```

`type credits int` creates a new type whose underlying type is `int`. You can't add methods to `int` itself, or to types from other packages, but you can wrap them in your own type and add methods to that.

## Your turn

Give Textio's `message` type two methods:

- `segments` returns how many 160-character segments the body needs, rounding **up** (the quiz above shows a trick for that). An empty body needs 0.
- `cost` returns the price in cents: 2 cents per segment, doubled if the recipient isn't a US number. US numbers start with `"+1"`.

`cost` can call `m.segments()`. To check the prefix, use `strings.HasPrefix(m.to, "+1")` from the `strings` package (you'll need to import it).

## Further reading

- [Go by Example: Methods](https://gobyexample.com/methods)
- [A Tour of Go: Methods](https://go.dev/tour/methods/1)
- [Learn Go with Tests: Structs, methods & interfaces](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/structs-methods-and-interfaces)
