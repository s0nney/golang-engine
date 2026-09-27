---
title: Basic Types and Zero Values
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	var count int
      	var name string
      	var active bool
      	fmt.Println(count, name == "", active)
      }
      ```
    options:
      - text: '`nil nil nil`'
      - text: '`0 true false`'
        correct: true
      - text: '`0 false false`'
      - text: It doesn't compile, because the variables were never given values
    explanation: |
      Variables declared without a value get their type's **zero value**: `0`
      for `int`, `""` for `string` and `false` for `bool`. The name is the
      empty string, so `name == ""` is `true`.
  - question: Which type should Textio use to store the price of a single message, such as `0.0075` dollars?
    options:
      - text: '`int`'
      - text: '`bool`'
      - text: '`string`'
      - text: '`float64`'
        correct: true
    explanation: |
      `int` can only hold whole numbers, so `0.0075` won't fit. `float64`
      holds numbers with a fractional part. (Real billing systems often count
      in whole fractions of a cent to avoid rounding errors, but `float64` is
      the type for decimals.)
---

Every value in Go has a **type**. The type decides what the value can hold and what you can do with it. You can add two numbers, but you can't add a number to `true`.

## The basic types

Here are the types you'll use constantly:

| Type | Holds | Example |
|------|-------|---------|
| `string` | Text | `"Your code is 4821"` |
| `int` | Whole numbers, positive or negative | `42`, `-7` |
| `float64` | Numbers with a decimal point | `0.0075`, `3.14` |
| `bool` | `true` or `false` | `true` |

```go
package main

import "fmt"

func main() {
	message := "Your code is 4821"
	length := 17
	costPerMessage := 0.0075
	delivered := true

	fmt.Println(message, length, costPerMessage, delivered)
}
```

```text
Your code is 4821 17 0.0075 true
```

## More number types

Go has a whole family of number types that differ in size and sign:

- **Signed integers:** `int`, `int8`, `int16`, `int32`, `int64`
- **Unsigned integers** (zero or more, never negative): `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- **Floats:** `float32`, `float64`

The number is how many bits the type uses. More bits means a bigger range. Unless you have a specific reason, stick with **`int`** for whole numbers and **`float64`** for decimals.

Two aliases you'll see around:

- `byte` is another name for `uint8`, and is used for raw data.
- `rune` is another name for `int32`, and holds a single Unicode character, written in single quotes: `'A'`, `'é'`, `'🚀'`.

## Integer division

Maths with integers stays in integers. Dividing two `int`s throws away the fractional part:

```go
package main

import "fmt"

func main() {
	fmt.Println(7 / 2)
	fmt.Println(7.0 / 2.0)
	fmt.Println(7 % 2)
}
```

```text
3
3.5
1
```

`%` is the **remainder** (or "modulo") operator: 7 divided by 2 is 3, remainder 1. It's surprisingly handy, as you'll see in the loops chapter.

## Zero values

In some languages, a variable you haven't set yet holds garbage or "undefined". Not in Go. Every variable declared without a value gets its type's **zero value**:

| Type | Zero value |
|------|-----------|
| `int`, `float64` and other numbers | `0` |
| `string` | `""` (the empty string) |
| `bool` | `false` |

```go
package main

import "fmt"

func main() {
	var smsCount int
	var sender string
	var isPremium bool

	fmt.Println(smsCount, isPremium)
	fmt.Println("sender:", sender, "(empty)")
}
```

```text
0 false
sender:  (empty)
```

Notice the two spaces after `sender:`. `Println` printed the empty string between them.

Zero values mean you can often declare a variable and start using it right away, for example `var total int` and then add to it. Good Go code leans on this: design your types so their zero value is useful.

## Further reading

- [Go by Example: Values](https://gobyexample.com/values)
- [A Tour of Go: Basic types](https://go.dev/tour/basics/11)
- [A Tour of Go: Zero values](https://go.dev/tour/basics/12)
