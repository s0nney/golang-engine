---
title: Pointer Basics
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	credits := 10
      	p := &credits
      	*p = *p + 5
      	credits++
      	fmt.Println(credits, *p)
      }
      ```
    options:
      - text: '`11 15`'
      - text: '`16 16`'
        correct: true
      - text: '`15 15`'
      - text: '`11 16`'
    explanation: |
      `p` points at `credits`, so `*p` and `credits` are two names for the
      same variable. `*p = *p + 5` makes it 15, and `credits++` makes it 16.
      Both print 16.
  - question: What does `p := new(3.5)` give you in Go 1.26 and later?
    options:
      - text: A `float64` with the value `3.5`
      - text: A `*float64` pointing at a new variable holding `3.5`
        correct: true
      - text: A slice of 3.5 elements
      - text: A compile error, because `new` only accepts a type
    explanation: |
      Since Go 1.26, `new` accepts an expression as well as a type. It
      allocates a variable, sets it to the value, and returns a pointer to
      it, so `p` is a `*float64` and `*p` is `3.5`.
---

Every variable lives somewhere in the computer's memory, and every spot in memory has an **address**, a number that says where it is. A **pointer** is a value that holds the address of a variable. Pointers let you share one variable between different parts of your program instead of copying it.

## `&` gets an address, `*` follows it

```go
package main

import "fmt"

func main() {
	credits := 50
	p := &credits // p holds the address of credits

	fmt.Println(*p) // read credits through the pointer

	*p = 45 // write credits through the pointer
	fmt.Println(credits)
}
```

```text
50
45
```

- `&credits` means "the address of `credits`". Its type is `*int`, read as "pointer to int".
- `*p` means "the variable `p` points at". Following a pointer like this is called **dereferencing**.
- Writing `*p = 45` changes `credits` itself, because `p` points at it.

If you print a pointer itself (`fmt.Println(p)`), you get the address, something like `0xc000012345`. The exact number is different every run and rarely useful.

## The `*` has two jobs

This confuses everyone at first, so let's be explicit:

- In a **type**, `*` means "pointer to": `var p *int` declares a pointer to an int.
- In an **expression**, `*` means "follow the pointer": `*p` is the int that `p` points at.

## The zero value is `nil`

A pointer that doesn't point at anything is `nil`, and that's the zero value for every pointer type:

```go
var p *int
fmt.Println(p == nil) // true
```

Following a nil pointer crashes your program. That gets its own lesson soon.

## Pointers to structs

Pointers to structs are the most common kind you'll use. Go gives you a shortcut: you can use `.` on a struct pointer directly, without writing `(*p).field`:

```go
package main

import "fmt"

type user struct {
	name    string
	credits int
}

func main() {
	u := &user{name: "alice", credits: 10}
	u.credits -= 3 // same as (*u).credits -= 3
	fmt.Println(u.name, u.credits)
}
```

```text
alice 7
```

`&user{...}` creates a struct and gives you a pointer to it in one step. You'll write this all the time.

## `new`

The built-in `new` also allocates a variable and returns a pointer to it. Given a **type**, it starts the variable at the zero value:

```go
count := new(int) // *int pointing at 0
*count += 2
```

Since **Go 1.26**, `new` also accepts an **expression**, and sets the new variable to that value:

```go
package main

import "fmt"

type config struct {
	retries *int // nil means "use the default"
}

func main() {
	limit := new(160) // *int pointing at 160
	fmt.Println(*limit)

	c := config{retries: new(3)}
	fmt.Println(*c.retries)
}
```

```text
160
3
```

Before Go 1.26, you'd need a temporary variable (`r := 3` then `&r`), because you can't take the address of a plain number: `&3` doesn't compile. `new(3)` is especially handy for optional fields, where `nil` means "not set" and a pointer means "set to this value".

## Why pointers?

Go copies values when you assign them or pass them to functions. Pointers let you avoid that copy when you need to:

1. **Share and change** one variable from several places, such as a function that updates a user's balance.
2. **Avoid copying** large structs.
3. **Represent "no value"** with `nil`.

The next lesson shows the first one in action.

## Further reading

- [Go by Example: Pointers](https://gobyexample.com/pointers)
- [A Tour of Go: Pointers](https://go.dev/tour/moretypes/1)
