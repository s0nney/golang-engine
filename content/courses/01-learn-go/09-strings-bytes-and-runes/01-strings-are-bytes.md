---
title: Strings Are Bytes
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	fmt.Println(len("café"))
      }
      ```
    options:
      - text: '`4`'
      - text: '`5`'
        correct: true
      - text: '`8`'
      - text: It doesn't compile, because `é` isn't allowed in a string
    explanation: |
      `len` counts **bytes**, not characters. Go strings are UTF-8, where
      `c`, `a` and `f` take one byte each and `é` takes two. That's 5 bytes.
  - question: 'Given `msg := "Textio"`, what is the type of `msg[0]`?'
    options:
      - text: '`string`'
      - text: '`rune`'
      - text: '`byte`'
        correct: true
      - text: '`int`'
    explanation: |
      Indexing a string gives you a single **byte** (a `uint8`), here `84`,
      the code for `T`. It's not a one-letter string. Use `msg[0:1]` if
      you want a string.
  - question: 'Why doesn''t `msg[0] = ''t''` compile when `msg` is a `string`?'
    options:
      - text: Because strings are immutable, so their bytes can't be changed
        correct: true
      - text: Because `'t'` should be written `"t"`
      - text: Because you need a pointer to change a string
    explanation: |
      Once a string is created, its bytes never change. To "change" a
      string you build a new one, for example by concatenating, or by
      converting to `[]byte`, editing, and converting back.
---

You've been using strings since your very first program. Textio is a *text* message company, so it's time to look inside them. What exactly is a string?

## A string is a sequence of bytes

A Go `string` is a read-only sequence of **bytes**. A byte is a small number from 0 to 255, and Go's `byte` type is just another name for `uint8`.

Text becomes bytes through an **encoding**. Go source code and Go strings use **UTF-8**, the encoding almost the whole internet uses. In UTF-8, plain English letters, digits and punctuation take **one byte** each, so for most of what you've written so far, bytes and characters lined up perfectly.

```go
package main

import "fmt"

func main() {
	msg := "Hi Alice"
	fmt.Println(len(msg))
	fmt.Println(msg[0])
	fmt.Printf("%c\n", msg[0])
	fmt.Println(msg[3:])
}
```

```text
8
72
H
Alice
```

- `len(msg)` is the number of **bytes**: 8.
- `msg[0]` is the first **byte**, and printing a byte prints a number. 72 is the code for `H`. The `%c` verb prints it as a character instead.
- Slicing a string, `msg[3:]`, works just like slicing a slice, and gives you a new string. The indexes are byte positions.

## Beyond English

Textio has customers all over the world. UTF-8 can encode every character in every language (and emoji!), but anything outside plain English takes **more than one byte**:

```go
package main

import "fmt"

func main() {
	fmt.Println(len("hello"))
	fmt.Println(len("héllo"))
	fmt.Println(len("こんにちは"))
	fmt.Println(len("👋"))
}
```

```text
5
6
15
4
```

`é` is 2 bytes, each Japanese character is 3 bytes, and the waving hand emoji is 4. So:

> `len(s)` counts **bytes**, not characters.

That's a classic gotcha. If Textio charged per character using `len`, a customer sending `"👋"` would be billed for 4 characters. You'll fix that properly in the next lesson.

It also means indexing and slicing can land in the *middle* of a character. `"é"[:1]` is half of an `é`, and prints as the replacement character `�`. With English text you're safe; with anything else, think in characters, not bytes.

## Strings are immutable

You can't change the bytes of an existing string:

```go
msg := "hi alice"
msg[0] = 'H' // error: cannot assign to msg[0] (neither addressable nor a map index expression)
```

Strings are **immutable**. Every string operation that looks like it changes a string actually creates a **new** one:

```go
msg := "hi alice"
msg = "H" + msg[1:] // a brand new string; msg now points at it
```

Immutability makes strings safe to share. Passing a string to a function, or slicing it, never copies the text, because nobody can change it underneath you.

## Bytes you can change: `[]byte`

When you really need to edit bytes in place, convert the string to a **byte slice**, change it, and convert back:

```go
package main

import "fmt"

func main() {
	msg := "hi alice"
	b := []byte(msg) // a copy of the bytes
	b[0] = 'H'
	b[3] = 'A'
	fmt.Println(string(b))
	fmt.Println(msg)
}
```

```text
Hi Alice
hi alice
```

Each conversion **copies** the data, so the original string is untouched. `'H'` in single quotes is a character literal. You'll meet it properly in the next lesson.

## Further reading

- [Strings, bytes, runes and characters in Go (Go blog)](https://go.dev/blog/strings)
