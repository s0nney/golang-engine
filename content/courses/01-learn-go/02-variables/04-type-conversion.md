---
title: Type Conversion
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	price := 2.99
      	fmt.Println(int(price))
      }
      ```
    options:
      - text: '`3`'
      - text: '`2`'
        correct: true
      - text: '`2.99`'
      - text: It doesn't compile
    explanation: |
      Converting a `float64` to an `int` doesn't round. It chops off
      (truncates) everything after the decimal point, so `2.99` becomes `2`.
  - question: How do you turn the number `42` into the string `"42"`?
    options:
      - text: '`string(42)`'
      - text: '`strconv.Itoa(42)`'
        correct: true
      - text: '`int.String(42)`'
    explanation: |
      `strconv.Itoa` ("integer to ASCII") returns the digits as text.
      `string(42)` is a trap: it gives you the character with code 42, which
      is `"*"`. `go vet` even warns you about it.
---

Go is strict about types. It will **never** silently convert a value from one type to another for you. If you want to mix types, you convert explicitly.

## Mixing types doesn't compile

Say Textio charges per message, and you want the total cost:

```go
messages := 150       // int
costPerMessage := 0.01 // float64

total := messages * costPerMessage // error: mismatched types int and float64
```

Many languages would quietly turn `150` into `150.0`. Go makes you say what you mean.

## Converting between number types

To convert a value, write the type name like a function call: `Type(value)`.

```go
package main

import "fmt"

func main() {
	messages := 150
	costPerMessage := 0.01

	total := float64(messages) * costPerMessage
	fmt.Println(total)
}
```

```text
1.5
```

Converting the other way, from `float64` to `int`, **truncates**. It chops off the fractional part rather than rounding:

```go
package main

import "fmt"

func main() {
	average := 7.9
	change := -7.9
	fmt.Println(int(average))
	fmt.Println(int(change))
}
```

```text
7
-7
```

## Untyped constants are flexible

You may wonder why `float64(messages) * 0.01` works while the variables didn't. Plain number literals like `0.01` and `160` are **untyped constants**. They adapt to whatever type they're used with. The strictness applies to *typed* values like variables.

```go
var cost float64 = 2 // fine: the constant 2 becomes a float64
```

## Numbers and strings

Converting between numbers and text is a different job, handled by the `strconv` ("string conversion") package:

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	code := 4821
	text := "Your code is " + strconv.Itoa(code)
	fmt.Println(text)

	n, err := strconv.Atoi("160")
	fmt.Println(n+1, err)
}
```

```text
Your code is 4821
161 <nil>
```

- `strconv.Itoa` turns an `int` into a `string`.
- `strconv.Atoi` turns a `string` into an `int`. Parsing text can fail (what's the number in `"hello"`?), so it also returns an **error**. `<nil>` means "no error". You'll learn all about errors later.

Notice `+` also works on strings: it **joins** (concatenates) them.

## The `string(int)` trap

This looks like it should work, but doesn't do what you think:

```go
fmt.Println(string(65)) // prints "A", not "65"!
```

Converting an integer to a `string` treats it as a character code, and 65 is the code for `A`. This catches so many people that `go vet` (a tool that checks for suspicious code) warns about it. Use `strconv.Itoa` or `fmt.Sprint` instead.

## Further reading

- [A Tour of Go: Type conversions](https://go.dev/tour/basics/13)
- [Go by Example: Number Parsing](https://gobyexample.com/number-parsing)
