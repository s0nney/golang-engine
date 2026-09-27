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
  - question: In Go 1.27, which of these `sender` literals does **not** compile?
    options:
      - text: '`sender{name: "bot", rateLimit: 10}`'
      - text: '`sender{user: user{name: "bot"}, rateLimit: 10}`'
      - text: '`sender{user: user{phone: "+1-555-0100"}, name: "bot"}`'
        correct: true
    explanation: |
      Since Go 1.27 you can set promoted fields like `name` directly in a
      literal, and you can still set the whole embedded struct with
      `user: user{...}`. What you can't do is mix the two in one literal.
exercise:
  starter: |
    package main

    import "fmt"

    type account struct {
    	name  string
    	phone string
    }

    // bulkSender should embed account.
    type bulkSender struct {
    	dailyLimit int
    }

    func newBulkSender(name, phone string, limit int) bulkSender {
    	// ?
    	return bulkSender{dailyLimit: limit}
    }

    func label(s bulkSender) string {
    	// ?
    	return ""
    }

    func main() {
    	s := newBulkSender("Acme Deliveries", "+1-555-0142", 500)
    	fmt.Println(label(s))
    }
  solution: |
    package main

    import "fmt"

    type account struct {
    	name  string
    	phone string
    }

    type bulkSender struct {
    	account
    	dailyLimit int
    }

    func newBulkSender(name, phone string, limit int) bulkSender {
    	return bulkSender{name: name, phone: phone, dailyLimit: limit}
    }

    func label(s bulkSender) string {
    	return fmt.Sprintf("%s (%s), %d/day", s.name, s.phone, s.dailyLimit)
    }

    func main() {
    	s := newBulkSender("Acme Deliveries", "+1-555-0142", 500)
    	fmt.Println(label(s))
    }
  tests: |
    package main

    import "testing"

    func TestNewBulkSender(t *testing.T) {
    	s := newBulkSender("Acme Deliveries", "+1-555-0142", 500)
    	want := account{name: "Acme Deliveries", phone: "+1-555-0142"}
    	if s.account != want {
    		t.Errorf("newBulkSender(...).account = %+v, want %+v", s.account, want)
    	}
    	if s.name != want.name {
    		t.Errorf("s.name = %q, want the promoted field to be %q", s.name, want.name)
    	}
    	if s.dailyLimit != 500 {
    		t.Errorf("s.dailyLimit = %d, want 500", s.dailyLimit)
    	}
    }

    func TestLabel(t *testing.T) {
    	for _, tc := range []struct {
    		s    bulkSender
    		want string
    	}{
    		{bulkSender{account{"Acme Deliveries", "+1-555-0142"}, 500}, "Acme Deliveries (+1-555-0142), 500/day"},
    		{bulkSender{account{"Pizza Hub", "+44-20-7946-0000"}, 20}, "Pizza Hub (+44-20-7946-0000), 20/day"},
    	} {
    		if got := label(tc.s); got != tc.want {
    			t.Errorf("label(%+v) = %q, want %q", tc.s, got, tc.want)
    		}
    	}
    }
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
- Since Go 1.27, a literal may also set promoted fields directly: `sender{name: "Acme", rateLimit: 100}`. You can't mix that with the `user:` key in the same literal, though, and it doesn't work through an embedded *pointer* (`*user`).
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

## Your turn

1. **Embed** `account` in `bulkSender`, so a bulk sender has a promoted `name` and
   `phone`.
2. Complete `newBulkSender` so it returns a `bulkSender` with all three values set.
   Either literal style from this lesson works.
3. Complete `label` so it returns a string like
   `Acme Deliveries (+1-555-0142), 500/day`. Use the promoted fields;
   `fmt.Sprintf` with `%s` and `%d` does the formatting.

## Further reading

- [Go by Example: Struct Embedding](https://gobyexample.com/struct-embedding)
