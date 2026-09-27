---
title: Defining Structs
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      type message struct {
      	to   string
      	body string
      	sent bool
      }

      func main() {
      	m := message{to: "Alice"}
      	fmt.Printf("%q %q %v\n", m.to, m.body, m.sent)
      }
      ```
    options:
      - text: It doesn't compile, because `body` and `sent` are missing
      - text: '`"Alice" nil false`'
      - text: '`"Alice" "" false`'
        correct: true
      - text: '`"Alice" "" true`'
    explanation: |
      Fields you leave out of a struct literal get their zero values: `""`
      for the string and `false` for the bool. `%q` shows the empty string
      as `""`.
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      type user struct {
      	name    string
      	credits int
      }

      func main() {
      	a := user{"alice", 10}
      	b := a
      	b.credits = 0
      	fmt.Println(a.credits, b.credits)
      }
      ```
    options:
      - text: '`10 0`'
        correct: true
      - text: '`0 0`'
      - text: '`10 10`'
    explanation: |
      Assigning a struct copies all of its fields, just like an array. `b`
      is an independent copy, so changing `b.credits` doesn't touch `a`.
---

A Textio message has a recipient, a body, a send time and a delivery status. You *could* juggle four separate variables for every message, but that gets out of hand fast. A **struct** bundles related values into one.

## Declaring a struct type

```go
type message struct {
	to     string
	body   string
	length int
	sent   bool
}
```

This creates a new **type** called `message`. Each value of type `message` has four **fields**, each with a name and a type. `type` is the keyword for defining your own types, and `struct` says what kind of type it is.

## Creating and using structs

Build a struct value with a **struct literal**, and read or set fields with a dot:

```go
package main

import "fmt"

type message struct {
	to     string
	body   string
	length int
	sent   bool
}

func main() {
	m := message{
		to:   "+1-555-0100",
		body: "Your code is 4821",
	}
	m.length = len(m.body)
	m.sent = true

	fmt.Println(m.to, m.length, m.sent)
	fmt.Printf("%+v\n", m)
}
```

```text
+1-555-0100 17 true
{to:+1-555-0100 body:Your code is 4821 length:17 sent:true}
```

- Name the fields in the literal (`to: ...`). Any fields you leave out get their **zero values**.
- `%+v` prints a struct with its field names, which is great for debugging. Plain `%v` prints only the values.
- You can also write a literal with no names, like `message{"+1-555-0100", "hi", 2, false}`, but then you must give every field, in order. Named fields are clearer and don't break when someone adds a field, so prefer them.

A struct's zero value is a struct whose fields are all zero values. `var m message` gives you a perfectly usable, empty message.

## Structs are values

Like arrays, structs are copied when you assign them or pass them to a function. The function gets its own copy, and changing it doesn't affect the caller's struct. You'll see how to change that with pointers in the next chapter.

## Nested structs

A field can itself be a struct:

```go
package main

import "fmt"

type contact struct {
	name  string
	phone string
}

type message struct {
	sender    contact
	recipient contact
	body      string
}

func main() {
	m := message{
		sender:    contact{name: "Textio", phone: "+1-555-0000"},
		recipient: contact{name: "Alice", phone: "+1-555-0100"},
		body:      "Welcome!",
	}
	fmt.Println(m.recipient.name, "gets", m.body)
}
```

```text
Alice gets Welcome!
```

Chain the dots to reach inner fields: `m.recipient.name`.

## Anonymous structs

If you need a struct shape only once, you can skip naming the type and write the struct type inline:

```go
package main

import "fmt"

func main() {
	config := struct {
		retries int
		carrier string
	}{
		retries: 3,
		carrier: "Alpha Mobile",
	}
	fmt.Println(config.carrier, config.retries)
}
```

```text
Alpha Mobile 3
```

Anonymous structs are handy for one-off groupings, such as test cases (you'll use them a lot in the testing chapter). If you find yourself writing the same anonymous struct twice, give it a name.

## Comparing structs

Structs whose fields are all comparable can be compared with `==`. Two structs are equal when all their fields are equal:

```go
a := contact{name: "Alice", phone: "+1-555-0100"}
b := contact{name: "Alice", phone: "+1-555-0100"}
fmt.Println(a == b) // true
```

That also means such structs can be used as map keys. (A struct containing a slice or map can't be compared, so it can't be a key.)

## Further reading

- [Go by Example: Structs](https://gobyexample.com/structs)
- [A Tour of Go: Structs](https://go.dev/tour/moretypes/2)
