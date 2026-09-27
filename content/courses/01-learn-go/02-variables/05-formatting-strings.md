---
title: Formatting Strings
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	cost := 1.5
      	fmt.Printf("Total: $%.2f for %d messages\n", cost, 150)
      }
      ```
    options:
      - text: '`Total: $1.5 for 150 messages`'
      - text: '`Total: $1.50 for 150 messages`'
        correct: true
      - text: '`Total: $%.2f for %d messages`'
      - text: '`Total: $1.500000 for 150 messages`'
    explanation: |
      `%.2f` prints a float with exactly two digits after the decimal point, so
      `1.5` becomes `1.50`. `%d` prints an integer.
  - question: What's the difference between `fmt.Printf` and `fmt.Sprintf`?
    options:
      - text: '`Sprintf` is faster'
      - text: '`Sprintf` only works with strings'
      - text: '`Sprintf` returns the formatted string instead of printing it'
        correct: true
      - text: '`Printf` adds a newline at the end and `Sprintf` doesn''t'
    explanation: |
      Both take a format string and values. `Printf` prints the result;
      `Sprintf` (the "S" is for "string") hands it back so you can store it
      in a variable, send it as an SMS, and so on. Neither adds a newline for
      you.
exercise:
  starter: |
    package main

    import "fmt"

    func main() {
    	name := "Alice"
    	sent := 42
    	costPerMessage := 0.01

    	total := costPerMessage // ? multiply by the number of messages sent

    	fmt.Println(name, sent, total) // ? use fmt.Printf instead
    }
  solution: |
    package main

    import "fmt"

    func main() {
    	name := "Alice"
    	sent := 42
    	costPerMessage := 0.01

    	total := float64(sent) * costPerMessage

    	fmt.Printf("Hi %s, you sent %d messages this month.\n", name, sent)
    	fmt.Printf("Your bill is $%.2f\n", total)
    }
  expected_output: |
    Hi Alice, you sent 42 messages this month.
    Your bill is $0.42
---

`fmt.Println` is great for quick output, but real messages need precise wording. "Hi Alice, your balance is $4.50" needs the name and number dropped into exactly the right spots. That's what `fmt.Printf` is for.

## `fmt.Printf`

`Printf` ("print formatted") takes a **format string** with placeholders called **verbs**, followed by the values to fill them with:

```go
package main

import "fmt"

func main() {
	name := "Alice"
	balance := 4.5
	fmt.Printf("Hi %s, your balance is $%.2f\n", name, balance)
}
```

```text
Hi Alice, your balance is $4.50
```

The values fill the verbs in order: `name` goes into `%s`, and `balance` goes into `%.2f`.

`\n` is the **newline** character. Unlike `Println`, `Printf` does *not* add a newline at the end, so you add one yourself.

## The verbs you'll use most

| Verb | Meaning | Example output |
|------|---------|----------------|
| `%v` | Any value, in its default format | `42`, `true`, `Alice` |
| `%s` | A string | `Alice` |
| `%q` | A string in double quotes | `"Alice"` |
| `%d` | An integer (d is for decimal) | `42` |
| `%f` | A float | `4.500000` |
| `%.2f` | A float with 2 decimal places | `4.50` |
| `%t` | A bool | `true` |
| `%T` | The **type** of the value | `float64` |
| `%%` | A literal percent sign | `%` |

When in doubt, `%v` prints almost anything sensibly. `%T` is great for debugging when you're not sure what type something is:

```go
package main

import "fmt"

func main() {
	fmt.Printf("%v %v %v\n", 160, 0.0075, "SMS")
	fmt.Printf("%T %T %T\n", 160, 0.0075, "SMS")
	fmt.Printf("%q\n", "hello")
	fmt.Printf("%d%% delivered\n", 98)
}
```

```text
160 0.0075 SMS
int float64 string
"hello"
98% delivered
```

## `fmt.Sprintf`: format without printing

Often you don't want to print a string. You want to *build* it and use it, for example as the body of an SMS. `fmt.Sprintf` works exactly like `Printf` but **returns** the string instead:

```go
package main

import "fmt"

func main() {
	code := 4821
	minutes := 10
	sms := fmt.Sprintf("Your Textio code is %d. It expires in %d minutes.", code, minutes)

	fmt.Println(sms)
	fmt.Println(len(sms), "characters")
}
```

```text
Your Textio code is 4821. It expires in 10 minutes.
51 characters
```

`len` is a built-in function that returns the length of a string (in bytes, which for plain English text is the same as characters).

## Width and padding

You can give a verb a **width** to line things up in columns. `%-8s` pads a string to 8 characters, left-aligned; `%5d` pads a number to 5 characters, right-aligned:

```go
fmt.Printf("%-8s|%5d\n", "Alice", 42)
fmt.Printf("%-8s|%5d\n", "Bob", 1337)
```

```text
Alice   |   42
Bob     | 1337
```

## Mismatched verbs

If you use the wrong verb, Go doesn't crash. It prints a warning into your output instead:

```go
fmt.Printf("%d\n", "Alice") // %!d(string=Alice)
```

The `go vet` tool catches these mistakes before you run the code.

## Your turn

Time to send Alice her monthly bill. Complete the program so that it:

1. Calculates `total`, the number of messages sent multiplied by `costPerMessage`. (You'll need a type conversion!)
2. Prints exactly:

```text
Hi Alice, you sent 42 messages this month.
Your bill is $0.42
```

Use `fmt.Printf` with the `%s`, `%d` and `%.2f` verbs, and use the variables rather than typing the values into the string.

## Further reading

- [Go by Example: String Formatting](https://gobyexample.com/string-formatting)
