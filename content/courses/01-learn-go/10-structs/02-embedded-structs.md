---
title: Embedded Structs
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      type user struct {
      	name string
      }

      type sender struct {
      	user
      	rateLimit int
      }

      func main() {
      	s := sender{user: user{name: "textio-bot"}, rateLimit: 10}
      	fmt.Println(s.name, s.user.name == s.name)
      }
      ```
    options:
      - text: It doesn't compile, because `sender` has no `name` field
      - text: '`textio-bot true`'
        correct: true
      - text: '`textio-bot false`'
      - text: An empty name, then `true`
    explanation: |
      `sender` embeds `user`, so `user`'s fields are promoted: `s.name` is a
      shortcut for `s.user.name`. They're the same field, so the comparison
      is `true`.
  - question: How do you write the literal for a struct with an embedded `user` field?
    options:
      - text: '`sender{name: "bot", rateLimit: 10}`'
      - text: '`sender{user: user{name: "bot"}, rateLimit: 10}`'
        correct: true
      - text: '`sender{user.name: "bot", rateLimit: 10}`'
    explanation: |
      Promotion only works for *reading and writing* fields. In a literal,
      the embedded struct is a field whose name is its type name, so you set
      it as `user: user{...}`.
---

Textio has several kinds of accounts: regular users, senders that are allowed to send bulk messages, and admins. They all have a name and a phone number. How do you share those fields without copying them into every struct?

## Nesting, the long way

You could nest a named field, as in the last lesson:

```go
type user struct {
	name  string
	phone string
}

type sender struct {
	account   user
	rateLimit int
}
```

That works, but you have to write `s.account.name` every time.

## Embedding

Go has a shortcut: **embed** the struct by writing just its type, with no field name:

```go
package main

import "fmt"

type user struct {
	name  string
	phone string
}

type sender struct {
	user
	rateLimit int
}

func main() {
	s := sender{
		user:      user{name: "Acme Deliveries", phone: "+1-555-0142"},
		rateLimit: 100,
	}

	fmt.Println(s.name)
	fmt.Println(s.phone)
	fmt.Println(s.user.name)
	fmt.Println(s.rateLimit)
}
```

```text
Acme Deliveries
+1-555-0142
Acme Deliveries
100
```

The embedded struct's fields are **promoted** to the outer struct. `s.name` is a shortcut for `s.user.name`, and both work.

A few details:

- The embedded field is still a real field, and its name is its type name: `user`. That's why the literal says `user: user{...}`.
- In a literal you **can't** set promoted fields directly. `sender{name: "Acme"}` doesn't compile.
- Methods of the embedded type are promoted too, which you'll see in the next lesson.

## Name clashes

If the outer struct has a field with the same name as one in the embedded struct, the outer one wins. You can still reach the inner one explicitly:

```go
package main

import "fmt"

type user struct {
	name string
}

type admin struct {
	user
	name string // shadows user.name
}

func main() {
	a := admin{user: user{name: "alice"}, name: "Alice (Admin)"}
	fmt.Println(a.name)
	fmt.Println(a.user.name)
}
```

```text
Alice (Admin)
alice
```

Yes, that's shadowing again. It's legal, but confusing, so avoid it when you can.

## Embedding is not inheritance

If you've heard of object-oriented languages like Java, embedding might look like inheritance. It isn't. A `sender` is **not** a `user`. You can't pass a `sender` to a function that expects a `user`:

```go
func greet(u user) { ... }

greet(s)      // error: cannot use s (variable of struct type sender) as user value
greet(s.user) // fine
```

Embedding is really just **composition** with some convenient shortcuts: a `sender` *has* a `user` inside it, and Go lets you reach its fields directly. Go deliberately doesn't have class inheritance. Instead, you build bigger types out of smaller ones.

## When to embed

Embed when the outer type really should expose the inner type's fields and methods as its own, like `sender` exposing `name` and `phone`. When a struct just *uses* another struct internally, a regular named field is usually clearer.

## Further reading

- [Go by Example: Struct Embedding](https://gobyexample.com/struct-embedding)
