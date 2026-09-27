---
title: Constants and iota
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      const (
      	Queued = iota
      	Sending
      	Delivered
      	Failed
      )

      func main() {
      	fmt.Println(Sending, Failed)
      }
      ```
    options:
      - text: '`1 4`'
      - text: '`0 3`'
      - text: '`1 3`'
        correct: true
      - text: '`Sending Failed`'
    explanation: |
      `iota` starts at `0` in a `const` block and goes up by one on each line.
      So `Queued` is 0, `Sending` is 1, `Delivered` is 2 and `Failed` is 3.
      They're plain numbers, so `Println` prints the numbers, not the names.
  - question: What happens if you try to change a constant after declaring it, like `maxLength = 200`?
    options:
      - text: The constant is updated to `200`
      - text: The program compiles but crashes when that line runs
      - text: The program doesn't compile
        correct: true
    explanation: |
      Constants can never change. The compiler rejects any attempt to assign
      to one, so the mistake is caught before the program ever runs.
---

Some values should never change. An SMS message holds at most 160 characters. There are 60 seconds in a minute. For values like these, Go has **constants**.

## Declaring constants

Use `const` instead of `var`:

```go
package main

import "fmt"

const maxSMSLength = 160

func main() {
	const greeting = "Welcome to Textio!"
	fmt.Println(greeting, maxSMSLength)
}
```

```text
Welcome to Textio! 160
```

Constants can be declared at the top level (outside any function) or inside a function. Unlike variables, **you can't use `:=`** with constants, and you can't change them once they exist:

```go
const maxSMSLength = 160
maxSMSLength = 200 // error: cannot assign to maxSMSLength
```

Constants must be known **when the program compiles**. You can build them from other constants:

```go
const secondsPerMinute = 60
const secondsPerHour = secondsPerMinute * 60 // fine: computed at compile time
```

But you can't set a constant to something only known while the program runs, like the current time or a value typed in by a user.

## Why bother?

Using a named constant instead of a bare `160` scattered through your code has two big wins:

1. **Readability.** `if length > maxSMSLength` explains itself. `if length > 160` makes the reader guess.
2. **One place to change.** If carriers raise the limit, you update one line.

## Grouping constants

You can group related constants in parentheses:

```go
const (
	maxSMSLength  = 160
	maxRecipients = 1000
	companyName   = "Textio"
)
```

## Counting with `iota`

Often you need a set of related constants that are just *different from each other*, like the status of a message. Go has a special helper for this: **`iota`**.

Inside a `const` block, `iota` starts at `0` and goes up by one on each line. And if you leave a line's value off, Go repeats the previous expression:

```go
package main

import "fmt"

const (
	Queued    = iota // 0
	Sending          // 1
	Delivered        // 2
	Failed           // 3
)

func main() {
	fmt.Println(Queued, Sending, Delivered, Failed)
}
```

```text
0 1 2 3
```

This style of constant is called an **enumeration** (or "enum"). The actual numbers don't really matter; what matters is that each status is distinct and has a readable name.

`iota` can be used in expressions too. A common trick is to start at 1, so that the zero value means "not set":

```go
const (
	Low    = iota + 1 // 1
	Normal            // 2
	Urgent            // 3
)
```

You'll see `iota` all over real Go code. Whenever you need a small, fixed set of options, reach for a `const` block.

## Further reading

- [Go by Example: Constants](https://gobyexample.com/constants)
- [Go by Example: Enums](https://gobyexample.com/enums)
